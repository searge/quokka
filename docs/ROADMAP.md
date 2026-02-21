# Wombat Greenfield Roadmap

This document outlines the strategic roadmap for the Wombat Go Rewrite (Greenfield project). It is based on the [Golang Greenfield Plan](../back/tmp/docs/golang/golang-greenfield-plan.md) and [Frontend Stack Recommendation](../back/tmp/docs/golang/frontend-stack-recommendation.md).

## Phase 1: Foundation & Architecture Validation (Months 1-2)

**Goal:** Establish the core repository, architecture guardrails, and validate the new tech stack.

* [ ] **Backend Bootstrap (Go):**
  * Initialize repository structure (modular monolith).
  * Set up HTTP router (`chi` or `gin`) and OpenAPI contract generation pipeline.
  * Configure database (`sqlc` + `pgx` with `golang-migrate` or `atlas`).
  * Implement foundational Observability (OpenTelemetry, structured logging).
* [ ] **Frontend Bootstrap (Vue 3):**
  * Initialize Vue 3 + TypeScript + Vite repository.
  * Set up state management (Pinia, TanStack Query) and UI component library (Vuetify/PrimeVue).
  * Generate API client from OpenAPI specs.
* [ ] **Identity & Auth Module:**
  * Implement OIDC authentication (replacing legacy CAS).
  * Implement Role-Based Access Control (RBAC) mapping.
* [ ] **Architecture Guardrails:**
  * Establish ADR process for cross-cutting concerns.
  * Setup CI/CD pipelines (lint, test, build).

## Phase 2: MVP Domain Slice & Core Data Migration (Months 3-4)

**Goal:** Implement the essential business logic and perform low-risk data migrations.

* [ ] **Core Domain Setup (Go):**
  * Implement `users` module (profiles, ssh keys, memberships).
  * Implement `projects` module (CRUD, tags, companies, business units).
* [ ] **Basic Frontend Integration:**
  * Develop dashboard UI for Projects and Users.
  * Implement routing and standard table/form patterns.
* [ ] **Low-Risk Data Migration (ETL):**
  * Develop restart-safe, idempotent ETL jobs.
  * Migrate canonical records (users, companies, project metadata) from legacy DB to new Postgres schema.
  * Implement reconciliation reporting for migrated data.

## Phase 3: Core Orchestration & First Integrations (Months 5-6)

**Goal:** Bring the primary value proposition online—provisioning and orchestrating services.

* [ ] **Task Engine v1:**
  * Develop internal worker pool and Postgres queue table for async jobs.
  * Define new task report data model (replacing legacy serialized payloads).
* [ ] **Services Module:**
  * Implement project component lifecycle management.
  * Establish adapter interfaces for external providers.
* [ ] **Priority Integrations:**
  * **Proxmox Plugin:** Container provisioning flow, status checks.
  * **GitLab Plugin:** Repository mapping, user/group permission syncing.
* [ ] **Frontend Updates:**
  * Integrate service status views and job-progress widgets in the UI.

## Phase 4: Extended Integrations & Medium/High-Risk Migration (Months 7-8)

**Goal:** Expand capabilities and migrate complex stateful entities.

* [ ] **Extended Integrations:**
  * Develop plugins for Redmine, Jenkins, Harbor, Rancher.
* [ ] **Medium/High-Risk Data Migration:**
  * Migrate project component states (mapping legacy statuses: `IN_PROGRESS`, `COMPLETED`, etc.).
  * Rebuild Audit trails and legacy statistics generation into canonical event streams.
* [ ] **UI Refinements:**
  * Finalize all admin-controlled operations and diagnostic views.

## Phase 5: Pre-production Soak, Dual Run & Cutover (Months 9-10)

**Goal:** Ensure stability, test migration procedures, and execute the final transition.

* [ ] **Dual Operations:**
  * Run legacy Preprod and new Go Preprod in parallel.
  * Execute full replayable migration jobs continuously.
* [ ] **Testing & Hardening:**
  * Perform failure drills and load testing.
  * Complete API compatibility checks (especially for frontend integrations).
* [ ] **Release & Stabilization:**
  * Execute production cutover.
  * Monitor OpenTelemetry metrics and handle immediate post-launch support.
  * Decommission legacy Tomcat/Java platform.
