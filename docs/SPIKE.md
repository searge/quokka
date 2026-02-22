# Quokka Spike: 2-Week Proof of Concept

**Goal:** Validate Go tech stack and plugin architecture. Prove that Projects CRUD + forge-ovh-cli integration works. Decide GO/NO-GO for full rewrite.

**Deliverables:**

- Working REST API for Projects
- forge-ovh-cli as first plugin
- OpenAPI spec
- curl/Postman examples
- Decision document

**NOT in scope:**

- Frontend
- Authentication (hardcoded user context)
- Data migration
- Full error handling
- Production deployment

---

## Week 1: Foundation + Projects CRUD

### Day 1-2: Repository Setup

**Tasks:**

- [x] Initialize Go module
- [x] Setup Cobra CLI structure
- [x] Configure router (chi or gin)
- [x] Setup sqlc + pgx
- [x] Database migrations (golang-migrate)
- [x] Docker Compose for local Postgres
- [x] Basic logging (structured logs)

**Deliverable:** `go run . server` starts HTTP server on :8080

---

### Day 3-5: Projects Domain

**Database schema:**

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    unix_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    visibility VARCHAR(50) DEFAULT 'private',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE project_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    tag VARCHAR(100) NOT NULL
);
```

**API endpoints:**

```
POST   /api/v1/projects          # Create project
GET    /api/v1/projects          # List projects
GET    /api/v1/projects/{id}     # Get project
PUT    /api/v1/projects/{id}     # Update project
DELETE /api/v1/projects/{id}     # Delete project
```

**Implementation:**

- `internal/store/project.go` — sqlc queries
- `internal/service/project.go` — business logic
- `internal/handler/project.go` — HTTP handlers
- `internal/domain/project.go` — types

**Tests:**

- Unit tests for service layer
- Integration tests for handlers

**Deliverable:** Full CRUD working, tested with curl

---

## Week 2: forge-ovh-cli Integration + Plugin Architecture

### Day 6-7: Plugin Interface Design

**Plugin contract:**

```go
package plugin

// Plugin defines the interface all provider plugins must implement
type Plugin interface {
    // Name returns the plugin identifier
    Name() string

    // Provision creates resources for a project
    Provision(ctx context.Context, req ProvisionRequest) (*ProvisionResult, error)

    // Status checks current resource status
    Status(ctx context.Context, projectID string) (*StatusResult, error)

    // Deprovision removes project resources
    Deprovision(ctx context.Context, projectID string) error
}

type ProvisionRequest struct {
    ProjectID   string
    ProjectName string
    Template    string  // e.g., "lxc-debian-12"
    Resources   map[string]interface{}
}

type ProvisionResult struct {
    ResourceID string
    Metadata   map[string]string
    Status     string
}
```

**Tasks:**

- Design plugin interface
- Implement plugin registry
- Add plugin lifecycle hooks (init, health check)

**Deliverable:** Working plugin framework

---

### Day 8-10: forge-ovh-cli Integration

**Approach:** Wrap existing Python CLI via subprocess

```go
package proxmox

type ForgePlugin struct {
    cliPath string
    config  ForgeConfig
}

func (p *ForgePlugin) Provision(ctx context.Context, req plugin.ProvisionRequest) (*plugin.ProvisionResult, error) {
    cmd := exec.CommandContext(ctx, p.cliPath, "create",
        "--name", req.ProjectName,
        "--os", req.Template,
    )

    // Parse JSON output
    // Return result
}
```

**Tasks:**

- Implement ForgePlugin wrapper
- Add configuration (forge-ovh-cli path, credentials)
- Test create/status/delete operations
- Error handling and logging

**New API endpoints:**

```
POST   /api/v1/projects/{id}/containers       # Create container via forge-ovh-cli
GET    /api/v1/projects/{id}/containers       # List containers
DELETE /api/v1/projects/{id}/containers/{cid} # Delete container
```

**Deliverable:** Can create Proxmox LXC container via Quokka API

---

### Day 11: OpenAPI Spec + Documentation

**Tasks:**

- Generate OpenAPI spec from code (swag/oapi-codegen)
- Write README with curl examples
- Document plugin interface
- Architecture diagram

**README examples:**

```bash
# Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Project",
    "unix_name": "test-project",
    "description": "Spike test"
  }'

# Create container
curl -X POST http://localhost:8080/api/v1/projects/{id}/containers \
  -H "Content-Type: application/json" \
  -d '{
    "template": "debian-12",
    "cpu": 2,
    "memory": 2048
  }'
```

**Deliverable:** Complete API documentation

---

### Day 12: Testing + Decision

**Testing checklist:**

- [ ] Projects CRUD works
- [ ] forge-ovh-cli integration creates containers
- [ ] Errors handled gracefully
- [ ] Logs are readable
- [ ] Tests pass
- [ ] Performance is acceptable (< 200ms for CRUD)

**Decision criteria:**

| Question | Answer |
|----------|---------|
| Is Go stack easier than Java? | Yes/No |
| Is plugin architecture feasible? | Yes/No |
| Did we complete spike on time? | Yes/No |
| Do we want to continue? | Yes/No |
| Can we show this to Maxime? | Yes/No |

**Deliverable:** `docs/SPIKE_RESULTS.md` with GO/NO-GO decision

---

## Success Criteria

**GO if:**

- ✅ Projects CRUD working
- ✅ forge-ovh-cli creates containers successfully
- ✅ Code is cleaner than Java version
- ✅ Team wants to continue
- ✅ Can demo to stakeholders

**NO-GO if:**

- ❌ Technical blockers (Go ecosystem issues)
- ❌ Integration too complex
- ❌ No motivation to continue
- ❌ Better to invest time elsewhere

---

## Next Steps (if GO)

1. **Week 3-4:** Add Vue 3 frontend
   - Dashboard with projects list
   - Container status view
   - Basic forms

2. **Week 5-6:** Decision on full roadmap
   - Gather feedback from Maxime
   - Estimate effort for remaining integrations
   - Decide on timeline

3. **Month 2+:** See `docs/ROADMAP.md`

---

## Notes

- Spike code can be "quick and dirty" — focus on learning
- Don't over-engineer — this is proof of concept
- Document blockers and surprises
- Keep decision criteria in mind daily
- If stuck > 4 hours, pivot or ask for help

---

**Start date:** [Fill in]
**Target completion:** [Start date + 10 working days]
**Decision meeting:** [Day 12]
