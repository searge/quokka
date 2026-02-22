package projects

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

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
func NewService(store *Store, registry *plugin.Registry) *Service {
	return newService(store, registry)
}

func newService(store projectStore, registry pluginRegistry) *Service {
	return &Service{
		store:    store,
		registry: registry,
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
			fmt.Printf("WARNING: provisioning failed for project %s: %v\n", project.ID, err)
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
		// In a real pgx setup, we should check for pgx.ErrNoRows.
		// For the spike, assuming any error from store that's not parse
		// for a single ID lookup could be not found or an actual db failure.
		// A cleaner strategy is returning ErrProjectNotFound.
		return nil, ErrProjectNotFound
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
	// First check if it exists
	_, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectID) {
			return nil, err
		}
		return nil, ErrProjectNotFound
	}
	return s.store.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// Check identity
	_, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrInvalidProjectID) {
			return err
		}
		return ErrProjectNotFound
	}

	return s.store.Delete(ctx, id)
}

func (s *Service) validateCreate(req CreateProjectRequest) error {
	if len(req.Name) < 3 || len(req.Name) > 255 {
		return errors.New("name must be between 3 and 255 characters")
	}
	if len(req.UnixName) < 3 || len(req.UnixName) > 100 {
		return errors.New("unix_name must be between 3 and 100 characters")
	}
	if !unixNameRegex.MatchString(req.UnixName) {
		return ErrInvalidUnixName
	}
	return nil
}
