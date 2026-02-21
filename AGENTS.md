# AI Agent Coding Guidelines

This document provides instructions for AI coding assistants (Claude, Cursor, Copilot, etc.) working on Wombat.

**Read this FIRST before writing any code.**

---

## Project Context

Wombat is a Go rewrite of a Java platform for multi-project orchestration. Currently in spike phase.

**Key facts:**

- Modular monolith (not microservices yet)
- Clean architecture (handler → service → store)
- Plugin system for integrations
- PostgreSQL + sqlc for database
- Chi router for HTTP
- OpenAPI-first API design

**Current phase:** Spike (2 weeks)
**Focus:** Projects CRUD + forge-ovh-cli integration

---

## Core Principles

### 1. Clean Architecture

**Always follow this flow:**

```
HTTP Request → Handler → Service → Store → Database
                   ↓
                 Plugin → External System
```

**NEVER:**

- Call database directly from handler
- Put business logic in handlers
- Import handler from service
- Import service from store

**Example (CORRECT):**

```go
// handler.go
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
    var req CreateProjectRequest
    json.NewDecoder(r.Body).Decode(&req)

    project, err := h.service.Create(r.Context(), req) // ✅ Handler calls service
    if err != nil {
        httperr.Handle(w, err)
        return
    }

    json.NewEncoder(w).Encode(project)
}

// service.go
func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    // Business logic here
    return s.store.Create(ctx, req) // ✅ Service calls store
}

// store.go
func (s *Store) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    // SQL query via sqlc
}
```

**Example (WRONG):**

```go
// handler.go - WRONG
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
    var req CreateProjectRequest
    json.NewDecoder(r.Body).Decode(&req)

    project, err := h.store.Create(r.Context(), req) // ❌ Skip service layer
    // ...
}
```

---

### 2. Domain Boundaries

Each domain is self-contained in its own package.

**Structure:**

```
internal/projects/
├── handler.go     # HTTP layer
├── service.go     # Business logic
├── store.go       # Database queries
├── types.go       # Domain types
└── queries.sql    # SQL queries for sqlc
```

**Rules:**

✅ **DO:**

- Keep all domain logic in domain package
- Export only what's needed
- Use domain types everywhere

❌ **DON'T:**

- Import other domains directly (use service interfaces)
- Share types between domains (duplicate if needed)
- Put domain logic in shared packages

**Example (types.go):**

```go
package projects

import "time"

// Project is the core domain entity
type Project struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    UnixName    string    `json:"unix_name"`
    Description string    `json:"description,omitempty"`
    Active      bool      `json:"active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProjectRequest is the input DTO
type CreateProjectRequest struct {
    Name        string `json:"name" validate:"required,min=3,max=255"`
    UnixName    string `json:"unix_name" validate:"required,alphanum,min=3,max=100"`
    Description string `json:"description,omitempty"`
}

// UpdateProjectRequest is the update DTO
type UpdateProjectRequest struct {
    Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
    Description *string `json:"description,omitempty"`
    Active      *bool   `json:"active,omitempty"`
}
```

---

### 3. Error Handling

**Use domain errors, not HTTP errors in business logic.**

✅ **DO:**

```go
// service.go
var (
    ErrProjectNotFound = errors.New("project not found")
    ErrProjectExists   = errors.New("project already exists")
    ErrInvalidUnixName = errors.New("invalid unix name format")
)

func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    if !isValidUnixName(req.UnixName) {
        return nil, ErrInvalidUnixName // ✅ Domain error
    }

    exists, _ := s.store.ExistsByUnixName(ctx, req.UnixName)
    if exists {
        return nil, ErrProjectExists // ✅ Domain error
    }

    return s.store.Create(ctx, req)
}
```

❌ **DON'T:**

```go
// service.go - WRONG
func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    if !isValidUnixName(req.UnixName) {
        return nil, httperr.BadRequest("invalid unix name") // ❌ HTTP in service
    }
}
```

**Handler maps domain errors to HTTP:**

```go
// handler.go
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
    project, err := h.service.Create(r.Context(), req)
    if err != nil {
        switch {
        case errors.Is(err, ErrProjectExists):
            httperr.Conflict(w, "project already exists")
        case errors.Is(err, ErrInvalidUnixName):
            httperr.BadRequest(w, "invalid unix name")
        default:
            httperr.InternalError(w, err)
        }
        return
    }
    json.NewEncoder(w).Encode(project)
}
```

---

### 4. Database (sqlc)

**Always use sqlc for queries. Never write raw SQL in Go code.**

**Workflow:**

1. Write SQL in `queries.sql`
2. Run `sqlc generate`
3. Use generated code in `store.go`

**Example:**

```sql
-- internal/projects/queries.sql

-- name: GetProject :one
SELECT id, name, unix_name, description, active, created_at, updated_at
FROM projects
WHERE id = $1;

-- name: CreateProject :one
INSERT INTO projects (name, unix_name, description)
VALUES ($1, $2, $3)
RETURNING id, name, unix_name, description, active, created_at, updated_at;

-- name: ListProjects :many
SELECT id, name, unix_name, description, active, created_at, updated_at
FROM projects
WHERE active = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
```

```go
// store.go
package projects

import (
    "context"
    "database/sql"
)

type Store struct {
    db      *sql.DB
    queries *Queries // Generated by sqlc
}

func NewStore(db *sql.DB) *Store {
    return &Store{
        db:      db,
        queries: New(db),
    }
}

func (s *Store) GetByID(ctx context.Context, id string) (*Project, error) {
    return s.queries.GetProject(ctx, id) // ✅ Generated method
}

func (s *Store) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    return s.queries.CreateProject(ctx, CreateProjectParams{
        Name:        req.Name,
        UnixName:    req.UnixName,
        Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
    })
}
```

---

### 5. Configuration

**No global state. Pass config explicitly.**

✅ **DO:**

```go
// main.go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    db := setupDB(cfg.Database)
    server := setupServer(cfg.Server, db)
    server.Run()
}

// server.go
func setupServer(cfg config.ServerConfig, db *sql.DB) *Server {
    projectStore := projects.NewStore(db)
    projectService := projects.NewService(projectStore)
    projectHandler := projects.NewHandler(projectService)

    return &Server{
        config:  cfg,
        handler: projectHandler,
    }
}
```

❌ **DON'T:**

```go
// WRONG - global config
var cfg *config.Config

func init() {
    cfg = config.Load() // ❌ Global state
}

func setupServer() {
    db := setupDB(cfg.Database) // ❌ Reading global
}
```

---

### 6. Testing

**Write tests for service layer (business logic).**

```go
// service_test.go
package projects

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type mockStore struct {
    projects map[string]*Project
}

func (m *mockStore) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    p := &Project{
        ID:       "test-id",
        Name:     req.Name,
        UnixName: req.UnixName,
    }
    m.projects[p.ID] = p
    return p, nil
}

func (m *mockStore) ExistsByUnixName(ctx context.Context, unixName string) (bool, error) {
    for _, p := range m.projects {
        if p.UnixName == unixName {
            return true, nil
        }
    }
    return false, nil
}

func TestService_Create(t *testing.T) {
    store := &mockStore{projects: make(map[string]*Project)}
    service := NewService(store)

    req := CreateProjectRequest{
        Name:     "Test Project",
        UnixName: "test-project",
    }

    project, err := service.Create(context.Background(), req)

    require.NoError(t, err)
    assert.Equal(t, "test-project", project.UnixName)
    assert.NotEmpty(t, project.ID)
}

func TestService_Create_DuplicateUnixName(t *testing.T) {
    store := &mockStore{projects: make(map[string]*Project)}
    service := NewService(store)

    req := CreateProjectRequest{
        Name:     "Test",
        UnixName: "test",
    }

    // First create succeeds
    _, err := service.Create(context.Background(), req)
    require.NoError(t, err)

    // Second create fails
    _, err = service.Create(context.Background(), req)
    assert.ErrorIs(t, err, ErrProjectExists)
}
```

---

## Plugin System

**All external integrations are plugins.**

### Plugin Interface

```go
// internal/plugin/plugin.go
package plugin

import "context"

type Plugin interface {
    Name() string
    Health(ctx context.Context) error
    Provision(ctx context.Context, req ProvisionRequest) (*ProvisionResult, error)
    Status(ctx context.Context, resourceID string) (*StatusResult, error)
    Deprovision(ctx context.Context, resourceID string) error
}

type ProvisionRequest struct {
    ProjectID   string
    ProjectName string
    Template    string
    Resources   map[string]interface{}
}

type ProvisionResult struct {
    ResourceID string
    Metadata   map[string]string
    Status     string
}

type StatusResult struct {
    Status   string
    Metadata map[string]string
}
```

### Implementing a Plugin

```go
// internal/plugin/proxmox/proxmox.go
package proxmox

import (
    "context"
    "os/exec"
    "github.com/Searge/wombat/internal/plugin"
)

type Plugin struct {
    cliPath string
}

func New(cliPath string) *Plugin {
    return &Plugin{cliPath: cliPath}
}

func (p *Plugin) Name() string {
    return "proxmox"
}

func (p *Plugin) Health(ctx context.Context) error {
    cmd := exec.CommandContext(ctx, p.cliPath, "health")
    return cmd.Run()
}

func (p *Plugin) Provision(ctx context.Context, req plugin.ProvisionRequest) (*plugin.ProvisionResult, error) {
    cmd := exec.CommandContext(ctx, p.cliPath, "create",
        "--name", req.ProjectName,
        "--template", req.Template,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, err
    }

    // Parse output and extract resource ID
    resourceID := parseResourceID(output)

    return &plugin.ProvisionResult{
        ResourceID: resourceID,
        Status:     "creating",
        Metadata: map[string]string{
            "output": string(output),
        },
    }, nil
}
```

---

## Common Patterns

### HTTP Handler Pattern

```go
package projects

import (
    "encoding/json"
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
        httperr.BadRequest(w, "invalid JSON")
        return
    }

    project, err := h.service.Create(r.Context(), req)
    if err != nil {
        httperr.Handle(w, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteStatus(http.StatusCreated)
    json.NewEncoder(w).Encode(project)
}
```

### Service Pattern

```go
package projects

import (
    "context"
    "errors"
)

type Service struct {
    store  Store
    plugin plugin.Plugin
}

func NewService(store Store, plugin plugin.Plugin) *Service {
    return &Service{
        store:  store,
        plugin: plugin,
    }
}

func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    // Validation
    if err := s.validate(req); err != nil {
        return nil, err
    }

    // Check duplicates
    exists, err := s.store.ExistsByUnixName(ctx, req.UnixName)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrProjectExists
    }

    // Create in database
    project, err := s.store.Create(ctx, req)
    if err != nil {
        return nil, err
    }

    // Provision external resources (async in production, sync in spike)
    if err := s.provisionResources(ctx, project); err != nil {
        // Rollback or mark as failed
        return project, err
    }

    return project, nil
}

func (s *Service) validate(req CreateProjectRequest) error {
    if !isValidUnixName(req.UnixName) {
        return ErrInvalidUnixName
    }
    return nil
}
```

---

## Code Style

### Naming

- **Packages:** `lowercase`, single word when possible (`projects`, `users`, `plugin`)
- **Files:** `snake_case.go` (`service.go`, `handler.go`, `types.go`)
- **Types:** `PascalCase` (`Project`, `CreateProjectRequest`)
- **Functions:** `PascalCase` for exported, `camelCase` for private
- **Variables:** `camelCase` (`projectID`, `userName`)
- **Constants:** `PascalCase` or `UPPER_CASE` for enums

### Comments

```go
// Package projects provides project management functionality.
package projects

// Project represents a client project in the system.
type Project struct {
    ID   string // Unique identifier
    Name string // Display name
}

// Create creates a new project.
// Returns ErrProjectExists if unix_name already exists.
func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    // Implementation
}
```

### Error Messages

- Use lowercase, no punctuation: `"project not found"`
- Be specific: `"invalid unix name format"`, not `"invalid input"`
- Don't include field names in domain errors (handler adds them)

---

## What NOT to Do

❌ **Global variables** (except package-level errors)
❌ **`init()` functions** for side effects
❌ **Panics** in business logic (only for programmer errors)
❌ **Interface pollution** (only when you have 2+ implementations)
❌ **Premature abstraction** (YAGNI - You Ain't Gonna Need It)
❌ **God objects** (services with 20+ methods)
❌ **Circular imports** (fix with interfaces or restructure)
❌ **Mocking database in unit tests** (use real DB for integration tests)

---

## Checklist Before Committing

- [ ] Code follows clean architecture (handler → service → store)
- [ ] No business logic in handlers
- [ ] Domain errors, not HTTP errors in services
- [ ] Tests written for service layer
- [ ] sqlc generated if queries changed
- [ ] No global state
- [ ] Config passed explicitly
- [ ] Error handling is consistent
- [ ] Comments on exported types/functions
- [ ] `task fmt` and `task lint` pass
- [ ] All tests pass (`task test`)

---

## When in Doubt

1. Check `docs/ARCHITECTURE.md` for patterns
2. Look at existing code in similar domain
3. Ask: "Is this the simplest thing that could work?"
4. Follow Go idioms: <https://go.dev/doc/effective_go>
5. Prefer explicit over clever

---

## See Also

- `docs/ARCHITECTURE.md` — System design
- `docs/SPIKE.md` — Current phase
- `docs/ROADMAP.md` — Future plans
- `README.md` — Project overview
