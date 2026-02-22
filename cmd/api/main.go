package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/searge/quokka/internal/integration/proxmox"
	"github.com/searge/quokka/internal/platform"
	"github.com/searge/quokka/internal/plugin"
	"github.com/searge/quokka/internal/projects"
)

func main() {
	// Initialize context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Starting Quokka API server...")

	// Setup database connection
	dbpool, err := platform.NewDatabasePool(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbpool.Close()

	// Initialize Plugin Registry
	pluginRegistry := plugin.NewRegistry()

	// Initialize and Register Proxmox Plugin
	proxmoxPlugin := proxmox.New("")
	if err := pluginRegistry.Register(proxmoxPlugin); err != nil {
		log.Fatalf("Failed to register proxmox plugin: %v", err)
	}

	// Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize Projects Domain
	projectStore := projects.NewStore(dbpool)
	projectService := projects.NewService(projectStore, pluginRegistry, logger)
	projectHandler := projects.NewHandler(projectService, logger)

	// Initialize the router
	router := platform.NewRouter()

	// API version 1
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", platform.HealthCheckHandler)
		r.Mount("/projects", projectHandler.Routes())
	})

	// Configure the HTTP server
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Server listening on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server bind error: %v", err)
		}
	}()

	// Wait for interruption signal
	<-ctx.Done()
	log.Println("Shutting down server gracefully...")

	// Graceful shutdown with 5s timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped successfully")
}
