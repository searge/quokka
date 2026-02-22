package projects

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/searge/quokka/internal/plugin"
)

func newGetRequestWithID(id string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/projects/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}

func TestHandlerGetByIDReturns400ForInvalidProjectID(t *testing.T) {
	svc := newService(
		mockStore{
			getByID: func(context.Context, string) (*Project, error) {
				return nil, ErrInvalidProjectID
			},
		},
		mockRegistry{
			getFn: func(string) (plugin.Plugin, error) {
				return nil, plugin.ErrPluginNotFound
			},
		},
		nil,
	)
	h := NewHandler(svc, nil)

	rr := httptest.NewRecorder()
	h.GetByID(rr, newGetRequestWithID("bad-id"))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"]["code"] != "INVALID_PROJECT_ID" {
		t.Fatalf("expected code INVALID_PROJECT_ID, got %q", body["error"]["code"])
	}
}

func TestHandlerGetByIDReturns404ForNotFound(t *testing.T) {
	svc := newService(
		mockStore{
			getByID: func(context.Context, string) (*Project, error) {
				return nil, pgx.ErrNoRows
			},
		},
		mockRegistry{
			getFn: func(string) (plugin.Plugin, error) {
				return nil, plugin.ErrPluginNotFound
			},
		},
		nil,
	)
	h := NewHandler(svc, nil)

	rr := httptest.NewRecorder()
	h.GetByID(rr, newGetRequestWithID("2a4e6b16-8a62-4d57-a05b-9f59248dbdb2"))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandlerGetByIDReturns500ForInternalError(t *testing.T) {
	svc := newService(
		mockStore{
			getByID: func(context.Context, string) (*Project, error) {
				return nil, errors.New("db connection failed")
			},
		},
		mockRegistry{
			getFn: func(string) (plugin.Plugin, error) {
				return nil, plugin.ErrPluginNotFound
			},
		},
		nil,
	)
	h := NewHandler(svc, nil)

	rr := httptest.NewRecorder()
	h.GetByID(rr, newGetRequestWithID("2a4e6b16-8a62-4d57-a05b-9f59248dbdb2"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
