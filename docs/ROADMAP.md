# Wombat Greenfield Roadmap

**Prerequisite:** Spike completed successfully (see `docs/SPIKE.md`)

This roadmap assumes GO decision after spike validation. Timeline is realistic for production-ready rewrite with gradual migration.

**Total timeline:** 18-24 months to full production cutover

---

## Phase 0: Spike ✅ (Weeks 1-2)

**Completed:**

- Projects CRUD API
- forge-ovh-cli integration
- Plugin architecture proof
- GO decision

**Next:** Phase 1 if spike successful

---

## Phase 1: MVP Foundation + Frontend (Months 1-3)

**Goal:** Working product that Smile team can start using internally for new projects.

### Month 1: Core Backend Modules

**Backend tasks:**

- [ ] Refactor spike code to production quality
- [ ] Add Users module (profiles, SSH keys)
- [ ] Add basic RBAC (hardcoded roles first)
- [ ] Implement API versioning (`/api/v1`, `/api/v2`)
- [ ] Add contract tests for all endpoints
- [ ] Setup CI/CD pipeline (GitHub Actions)
  - Lint (golangci-lint)
  - Test
  - Build
  - Docker image

**Database schema:**

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    login VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- SSH Keys
CREATE TABLE ssh_keys (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR(100),
    public_key TEXT NOT NULL,
    fingerprint VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Project Members
CREATE TABLE project_members (
    project_id UUID REFERENCES projects(id),
    user_id UUID REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    PRIMARY KEY (project_id, user_id)
);
```

**API additions:**

```
# Users
POST   /api/v1/users
GET    /api/v1/users
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}

# SSH Keys
POST   /api/v1/users/{id}/ssh-keys
GET    /api/v1/users/{id}/ssh-keys
DELETE /api/v1/users/{id}/ssh-keys/{keyId}

# Project Members
POST   /api/v1/projects/{id}/members
GET    /api/v1/projects/{id}/members
DELETE /api/v1/projects/{id}/members/{userId}
```

**Deliverable:** Backend with Users + Projects + Members

---

### Month 2: Frontend Bootstrap

**Setup:**

- [ ] Initialize Vue 3 + TypeScript + Vite repository
- [ ] Configure TanStack Query for API calls
- [ ] Setup Pinia for client state
- [ ] Choose UI library (Vuetify or PrimeVue)
- [ ] Generate TypeScript API client from OpenAPI
- [ ] Setup environment configs (local, preprod, prod)

**Core pages:**

- [ ] Login page (stub auth for now)
- [ ] Dashboard (projects list)
- [ ] Project detail page
- [ ] Users list
- [ ] User profile page

**UI patterns:**

- Standard dashboard layout (top bar, side nav, content)
- Data tables with filters/pagination
- Forms with validation
- Status badges
- Loading states

**Deliverable:** Working UI for Projects and Users management

---

### Month 3: Integration + Polish

**Backend:**

- [ ] Improve forge-ovh-cli integration
  - Better error handling
  - Status polling
  - Logs/events
- [ ] Add basic audit log
  - Who created what when
  - Store in `audit_events` table
- [ ] Health check endpoint
- [ ] Metrics endpoint (Prometheus format)

**Frontend:**

- [ ] Container status view
  - List containers per project
  - Start/stop/restart actions
  - Status polling (live updates)
- [ ] Forms for create/edit
- [ ] Error handling and user feedback
- [ ] Loading indicators

**Deployment:**

- [ ] Docker Compose for local dev
- [ ] Deploy to preprod environment
- [ ] Internal dogfooding (Smile team uses it for 5 new projects)

**Deliverable:** Usable product for internal testing

---

## Phase 2: Core Integrations (Months 4-9)

**Goal:** Add remaining critical integrations one by one. Quality over speed.

**Strategy:** One integration per month, properly implemented.

### Month 4: LDAP Integration

**Tasks:**

- [ ] LDAP sync service
  - Import users from LDAP
  - Sync on schedule
  - Map LDAP groups to Wombat roles
- [ ] Update user creation flow
  - Create in Wombat → sync to LDAP
- [ ] LDAP authentication (optional, can use OIDC)

**Deliverable:** User management synced with LDAP

---

### Month 5: GitLab Integration

**Tasks:**

- [ ] GitLab plugin implementation
  - Create repository
  - Add project members with permissions
  - Manage deploy keys
- [ ] API endpoints for GitLab operations
- [ ] UI for GitLab repository view

**Deliverable:** Projects automatically get GitLab repos

---

### Month 6: Rancher/Kubernetes Integration

**Tasks:**

- [ ] Rancher plugin
  - Create namespace
  - Setup RBAC
  - Deploy basic workloads
- [ ] Kubernetes resource views in UI
  - Pods, deployments, services
  - Logs, events
- [ ] kubectl exec proxy (optional)

**Deliverable:** K8s namespace per project

---

### Month 7: Harbor Integration

**Tasks:**

- [ ] Harbor plugin
  - Create project in Harbor
  - Setup webhooks
  - Manage robot accounts
- [ ] UI for container images
- [ ] Integration with GitLab CI (push images)

**Deliverable:** Private registry per project

---

### Month 8: Redmine Integration (if needed)

**Tasks:**

- [ ] Redmine plugin
  - Create project
  - Setup trackers
  - Sync users/permissions
- [ ] UI link to Redmine

**Alternative:** Skip if moving away from Redmine

---

### Month 9: Jenkins Integration (if needed)

**Tasks:**

- [ ] Jenkins plugin
  - Create jobs from templates
  - Trigger builds
  - Get build status
- [ ] CI/CD pipeline view in UI

**Alternative:** Most teams moved to GitLab CI

---

## Phase 3: Authentication & RBAC (Months 10-11)

**Goal:** Replace hardcoded auth with proper OIDC and fine-grained permissions.

### Month 10: OIDC Authentication

**Tasks:**

- [ ] OIDC provider integration (Keycloak, Auth0, or self-hosted)
- [ ] JWT validation middleware
- [ ] Session management
- [ ] Login/logout flows in frontend
- [ ] Refresh token handling

**Deliverable:** Real authentication

---

### Month 11: Advanced RBAC

**Tasks:**

- [ ] Role definitions
  - Admin, Project Manager, Developer, Viewer
- [ ] Permission system
  - Create/Read/Update/Delete per resource
  - Project-level permissions
- [ ] UI permission checks
  - Hide/disable actions based on role
- [ ] Audit all actions

**Deliverable:** Production-ready access control

---

## Phase 4: Task System + Async Operations (Months 12-13)

**Goal:** Handle long-running operations properly.

### Month 12: Task Engine

**Tasks:**

- [ ] Task queue implementation
  - Postgres-based queue or Redis
  - Worker pool
  - Retry logic with backoff
- [ ] Task status API
- [ ] Task history and logs
- [ ] Idempotency keys

**Deliverable:** Reliable async operations

---

### Month 13: Task UI + Notifications

**Tasks:**

- [ ] Task status widgets in UI
  - Progress bars
  - Real-time updates (WebSocket or SSE)
- [ ] Task history page
- [ ] Email notifications (optional)
- [ ] Webhook integration

**Deliverable:** Transparent async operations for users

---

## Phase 5: Data Migration Prep (Months 14-15)

**Goal:** Prepare for migrating existing projects from Java version.

### Month 14: Migration Tooling

**Tasks:**

- [ ] ETL scripts for canonical data
  - Users, Projects, Companies
  - Use old API as source of truth
- [ ] Idempotent migration jobs
  - Can replay safely
  - Track migration state
- [ ] Reconciliation reporting
  - Compare old vs new
  - Checksums, counts
- [ ] Rollback procedures

**Deliverable:** Safe, repeatable migration process

---

### Month 15: Dry Run Migrations

**Tasks:**

- [ ] Migrate preprod data
  - 50 test projects
  - All users
  - Historical data
- [ ] Validate data integrity
- [ ] Test all integrations
- [ ] Performance testing under load
- [ ] Fix issues discovered

**Deliverable:** Proven migration process

---

## Phase 6: Parallel Run + Gradual Rollout (Months 16-20)

**Goal:** Run both systems in parallel, gradually migrate projects.

### Month 16: Dual Operations Setup

**Tasks:**

- [ ] Deploy new system to production
- [ ] Setup monitoring and alerts
- [ ] Runbooks for common issues
- [ ] On-call rotation
- [ ] Keep old system running

**Deliverable:** Both systems operational

---

### Month 17-18: Gradual Migration (Batches)

**Strategy:** Migrate in small batches, learn and improve.

**Batch 1 (10 projects):**

- [ ] Migrate low-risk projects
- [ ] Monitor for 2 weeks
- [ ] Gather user feedback
- [ ] Fix issues

**Batch 2 (25 projects):**

- [ ] Apply lessons learned
- [ ] Migrate medium-risk projects
- [ ] Monitor, feedback, fix

**Batch 3 (50 projects):**

- [ ] Increase batch size
- [ ] Automate more of migration
- [ ] Continue monitoring

---

### Month 19-20: Complete Migration

**Tasks:**

- [ ] Migrate remaining projects
- [ ] Verify all integrations working
- [ ] Final data reconciliation
- [ ] User training/documentation
- [ ] Communication plan

**Deliverable:** 100% of projects on new platform

---

## Phase 7: Decommission Legacy (Months 21-22)

**Goal:** Turn off Java version safely.

### Month 21: Read-Only Mode

**Tasks:**

- [ ] Set old system to read-only
- [ ] Archive database
- [ ] Keep available for reference
- [ ] Monitor for any missed dependencies

---

### Month 22: Final Decommission

**Tasks:**

- [ ] Final backups
- [ ] Turn off services
- [ ] Archive code repository
- [ ] Update documentation
- [ ] Celebrate! 🎉

**Deliverable:** Legacy system retired

---

## Phase 8: Polish + OpenSource (Months 23-24)

**Goal:** Prepare for public release.

### Month 23: Code Cleanup

**Tasks:**

- [ ] Remove Smile-specific code
- [ ] Generalize configurations
- [ ] Improve documentation
- [ ] Add example configurations
- [ ] Security audit

---

### Month 24: OpenSource Launch

**Tasks:**

- [ ] Choose license (Apache 2.0 recommended)
- [ ] Write contribution guidelines
- [ ] Create GitHub org
- [ ] Publish repository
- [ ] Launch announcement
  - HackerNews Show HN
  - r/selfhosted
  - r/devops
  - Platform Engineering communities
- [ ] First community contributors

**Deliverable:** Public open-source project

---

## Success Metrics

Track these throughout:

| Metric | Target |
|--------|--------|
| Projects on new platform | 100% by month 20 |
| System uptime | > 99.5% |
| API response time (p95) | < 200ms |
| User satisfaction (internal) | > 80% satisfied |
| GitHub stars (if OSS) | 100+ in first month |
| Active contributors | 5+ by month 24 |

---

## Risk Management

| Risk | Mitigation |
|------|-----------|
| Rewrite takes longer | Gradual rollout allows parallel run indefinitely |
| Integration breaks | Keep old system as fallback |
| Team capacity | Adjust timeline, reduce scope if needed |
| User adoption low | Internal dogfooding before wider rollout |
| OpenSource interest low | Product works for Smile regardless |

---

## Decision Points

**After Month 3:** Continue to Phase 2?

- Is MVP useful internally?
- Is team confident in tech stack?
- Is timeline realistic?

**After Month 9:** Commit to full migration?

- Are integrations stable?
- Is performance acceptable?
- Can we support this long-term?

**After Month 15:** Proceed with production rollout?

- Did dry run succeed?
- Is data migration safe?
- Are we ready for prime time?

---

## Notes

- This is a **maximum** timeline. Some phases may go faster.
- **Quality over speed** — better to delay than ship broken features.
- **Feedback loops** — internal users test each phase before moving on.
- **Parallel tracks** — backend and frontend work can overlap.
- **Community** — engage early, even before full OpenSource launch.

---

**Last updated:** [Date after spike completion]
**Status:** [Pending spike results | In progress | Complete]
