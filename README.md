# Quokka

Quokka is a Go-based backend for provisioning and lifecycle management of infrastructure services in Forge environments.

It is designed as a clean replacement path for legacy backend logic: predictable APIs, explicit domain boundaries, and safer automation around external systems (Proxmox, GitLab, and other integrations).

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Development roadmap](docs/ROADMAP.md)
- [Spike notes](docs/SPIKE.md)
- [Coding guidelines](docs/CODING_GUIDELINES.md)
- [Contributing](CONTRIBUTING.md)

## What Quokka Solves

- Standardizes infrastructure workflows behind one API
- Reduces manual operational steps in project/service provisioning
- Isolates external provider logic via plugin-style integrations
- Enables incremental migration away from legacy backend dependencies

## Architecture Highlights

- Modular monolith with clear domain modules
- Layered flow: handler -> service -> store
- PostgreSQL + migrations + typed queries
- Integration boundaries suitable for adapters/plugins
- API-first development approach for frontend and automation clients

## Quick Start

```bash
task setup
task dev
```

Run locally:

```bash
task run -- server
```

## Project Status

Active greenfield development. Scope and sequencing are tracked in `docs/` to keep this README concise.
