package projects

import (
	"context"
	"errors"
	"testing"

	"github.com/searge/quokka/internal/plugin"
)

type mockStore struct {
	createFn func(context.Context, CreateProjectRequest) (*Project, error)
	getByID  func(context.Context, string) (*Project, error)
	listFn   func(context.Context, int32, int32) ([]*Project, error)
	updateFn func(context.Context, string, UpdateProjectRequest) (*Project, error)
	deleteFn func(context.Context, string) error
}

func (m mockStore) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	if m.createFn == nil {
		return nil, errors.New("createFn is not set")
	}
	return m.createFn(ctx, req)
}

func (m mockStore) GetByID(ctx context.Context, id string) (*Project, error) {
	if m.getByID == nil {
		return nil, errors.New("getByID is not set")
	}
	return m.getByID(ctx, id)
}

func (m mockStore) List(ctx context.Context, limit, offset int32) ([]*Project, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, limit, offset)
}

func (m mockStore) Update(ctx context.Context, id string, req UpdateProjectRequest) (*Project, error) {
	if m.updateFn == nil {
		return nil, errors.New("updateFn is not set")
	}
	return m.updateFn(ctx, id, req)
}

func (m mockStore) Delete(ctx context.Context, id string) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

type mockRegistry struct {
	getFn func(string) (plugin.Plugin, error)
}

func (m mockRegistry) Get(name string) (plugin.Plugin, error) {
	if m.getFn == nil {
		return nil, errors.New("getFn is not set")
	}
	return m.getFn(name)
}

type mockPlugin struct {
	provisionFn func(context.Context, plugin.ProvisionRequest) (*plugin.ProvisionResult, error)
}

func (m mockPlugin) Name() string { return "proxmox" }
func (m mockPlugin) Health(context.Context) error {
	return nil
}

func (m mockPlugin) Provision(ctx context.Context, req plugin.ProvisionRequest) (*plugin.ProvisionResult, error) {
	if m.provisionFn == nil {
		return &plugin.ProvisionResult{ResourceID: "res-1", Status: "ok"}, nil
	}
	return m.provisionFn(ctx, req)
}

func (m mockPlugin) Status(context.Context, string) (*plugin.StatusResult, error) {
	return &plugin.StatusResult{Status: "running"}, nil
}

func (m mockPlugin) Deprovision(context.Context, string) error { return nil }

func TestServiceCreateRejectsInvalidUnixName(t *testing.T) {
	s := newService(mockStore{}, mockRegistry{}, nil)

	_, err := s.Create(context.Background(), CreateProjectRequest{
		Name:     "Valid Name",
		UnixName: "bad_name",
	})
	if !errors.Is(err, ErrInvalidUnixName) {
		t.Fatalf("expected ErrInvalidUnixName, got %v", err)
	}
}

func TestServiceCreatePropagatesErrProjectExists(t *testing.T) {
	s := newService(
		mockStore{
			createFn: func(context.Context, CreateProjectRequest) (*Project, error) {
				return nil, ErrProjectExists
			},
		},
		mockRegistry{
			getFn: func(string) (plugin.Plugin, error) {
				return nil, plugin.ErrPluginNotFound
			},
		},
		nil,
	)

	_, err := s.Create(context.Background(), CreateProjectRequest{
		Name:     "Valid Name",
		UnixName: "valid-name",
	})
	if !errors.Is(err, ErrProjectExists) {
		t.Fatalf("expected ErrProjectExists, got %v", err)
	}
}

func TestServiceCreateCallsProvisionWithProjectData(t *testing.T) {
	var gotReq plugin.ProvisionRequest

	s := newService(
		mockStore{
			createFn: func(context.Context, CreateProjectRequest) (*Project, error) {
				return &Project{
					ID:   "p-123",
					Name: "Alpha",
				}, nil
			},
		},
		mockRegistry{
			getFn: func(string) (plugin.Plugin, error) {
				return mockPlugin{
					provisionFn: func(_ context.Context, req plugin.ProvisionRequest) (*plugin.ProvisionResult, error) {
						gotReq = req
						return &plugin.ProvisionResult{ResourceID: "r-1", Status: "ok"}, nil
					},
				}, nil
			},
		},
		nil,
	)

	_, err := s.Create(context.Background(), CreateProjectRequest{
		Name:     "Alpha",
		UnixName: "alpha",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotReq.ProjectID != "p-123" || gotReq.ProjectName != "Alpha" {
		t.Fatalf("unexpected provision request: %+v", gotReq)
	}
}

func TestServiceGetPropagatesInvalidProjectID(t *testing.T) {
	s := newService(
		mockStore{
			getByID: func(context.Context, string) (*Project, error) {
				return nil, ErrInvalidProjectID
			},
		},
		mockRegistry{},
		nil,
	)

	_, err := s.Get(context.Background(), "not-a-uuid")
	if !errors.Is(err, ErrInvalidProjectID) {
		t.Fatalf("expected ErrInvalidProjectID, got %v", err)
	}
}
