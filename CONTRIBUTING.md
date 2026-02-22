# Contributing

This document describes the standard development workflow for Quokka.

## Principles

- Keep business logic in `internal/` packages, not in `cmd/`
- Prefer small, explicit interfaces at usage points
- Keep side effects at boundaries and domain logic deterministic
- Wrap and propagate errors with context

See full style reference in [docs/CODING_GUIDELINES.md](docs/CODING_GUIDELINES.md).

## Development Workflow

1. Create a short-lived branch from `main`
2. Implement changes with tests
3. Run local checks
4. Open a merge request with clear scope and risk notes

## Local Commands

```bash
task fmt
task lint
task test
task build
```

Or run full cycle:

```bash
task dev
```

## Code Organization

- `cmd/` CLI entrypoints only
- `internal/` application-private domains and services
- `pkg/` reusable packages
- `migrations/` database schema migrations
- `docs/` architecture and planning docs

## Pull Request Expectations

- Small focused diff
- Backward compatibility documented if API contracts change
- Migration steps included when schema changes
- Tests added or updated for behavior changes

## Documentation Policy

- Keep `README.md` high-level and external-facing
- Put design/process details in `docs/` or dedicated files
- Update docs in the same change when behavior or contracts change
