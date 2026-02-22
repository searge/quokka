package projects

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	project, err := h.service.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectExists):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, ErrInvalidUnixName):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			log.Printf("internal error in projects handler: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("failed to encode create project response: %v", err)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.List(r.Context(), 100, 0)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		log.Printf("failed to encode list projects response: %v", err)
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	project, err := h.service.Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			http.Error(w, "project not found", http.StatusNotFound)
		case errors.Is(err, ErrInvalidProjectID):
			http.Error(w, "invalid project id", http.StatusBadRequest)
		default:
			log.Printf("internal error in projects handler: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("failed to encode get project response: %v", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	project, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			http.Error(w, "project not found", http.StatusNotFound)
		case errors.Is(err, ErrInvalidProjectID):
			http.Error(w, "invalid project id", http.StatusBadRequest)
		default:
			log.Printf("internal error in projects handler: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("failed to encode update project response: %v", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.service.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			http.Error(w, "project not found", http.StatusNotFound)
		case errors.Is(err, ErrInvalidProjectID):
			http.Error(w, "invalid project id", http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
