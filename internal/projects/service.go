package projects

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/searge/quokka/internal/plugin"
)

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrProjectExists    = errors.New("project unix name already exists")
	ErrInvalidUnixName  = errors.New("invalid unix name format")
	ErrInvalidProjectID = errors.New("invalid project id format")

	unixNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)
)

// Service houses the central business logic for Projects.
type Service struct {
	store    projectStore
	registry pluginRegistry
	log      *slog.Logger
	validate *validator.Validate
}

type projectStore interface {
	Create(ctx context.Context, req CreateProjectRequest) (*Project, error)
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, limit, offset int32) ([]*Project, error)
	Update(ctx context.Context, id string, req UpdateProjectRequest) (*Project, error)
	Delete(ctx context.Context, id string) error
}

type pluginRegistry interface {
	Get(name string) (plugin.Plugin, error)
}

// NewService creates a new Service.
func NewService(store *Store, registry *plugin.Registry, logger *slog.Logger) *Service {
	return newService(store, registry, logger)
}

func newService(store projectStore, registry pluginRegistry, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	validate := validator.New()
	err := validate.RegisterValidation("unix_name", func(fl validator.FieldLevel) bool {
		return unixNameRegex.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(fmt.Errorf("failed to register unix_name validator: %w", err))
	}

	return &Service{
		store:    store,
		registry: registry,
		log:      logger,
		validate: validate,
	}
}

// Create generates a new project entity and attempts resource provisioning via plugins.
func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	if err := s.validateCreate(req); err != nil {
		return nil, err
	}

	// Persist to database
	project, err := s.store.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// For the Spike, synchronously trigger the Proxmox plugin via registry
	if proxmoxPlugin, err := s.registry.Get("proxmox"); err == nil {
		provCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if _, err := proxmoxPlugin.Provision(provCtx, plugin.ProvisionRequest{
			ProjectID:   project.ID,
			ProjectName: project.Name,
		}); err != nil {
			// GO-004: We swallow the error from the client's perspective to avoid
			// "500 Internal Error" when the DB creation actually succeeded.
			// Future work: Track ProvisionStatus on the Project entity.
			// Currently, we just log the failure.
			s.log.Warn("provisioning failed", "project_id", project.ID, "error", err)
		}
	}

	return project, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Project, error) {
	project, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectID) {
			return nil, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return project, nil
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]*Project, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.store.List(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateProjectRequest) (*Project, error) {
	project, err := s.store.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectID) {
			return nil, err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return project, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	err := s.store.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectID) {
			return err
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProjectNotFound
		}
		return err
	}
	return nil
}

func (s *Service) validateCreate(req CreateProjectRequest) error {
	if err := s.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, fieldErr := range validationErrors {
				if fieldErr.Field() == "UnixName" && fieldErr.Tag() == "unix_name" {
					return ErrInvalidUnixName
				}
			}
		}
		return err
	}
	return nil
}
