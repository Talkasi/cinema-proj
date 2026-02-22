# Cinema Management System

A cinema management system with a Go API, PostgreSQL database, and a TUI client.

## About

The project includes:
- PostgreSQL database with a structured schema
- REST API in Go with CRUD endpoints
- Role-based access (guest, user, admin)
- Booking and movie show management workflows

## Tech Stack

- **Backend**: Go (Golang)
- **Database**: PostgreSQL
- **API Documentation**: Swagger
- **Development Environment**: Visual Studio Code

## Quick Start (Local)

1. Copy `.env.example` to `.env` and fill in the required variables.
2. Start infrastructure:
   - `docker compose up -d`
3. Start API:
   - `go run ./cmd/api`
4. (Optional) Start TUI client:
   - `go run ./cmd/tui`

## Main Environment Variables

- `ADDR` - HTTP API address (for example `:8080`)
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASS`, `DB_SSL` - PostgreSQL connection
- `JWT_SECRET`, `TOKEN_DURATION` - JWT secret and token TTL (for example `24h`)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `FROM_EMAIL` - email settings for 2FA
- `OTEL_*` - observability/tracing settings
- `LOG_*` - logging configuration

Architecture overview and data flow: `docs/architecture.md`.
DTO requests are gradually migrated to centralized runtime validation via `validate` tags.

## Quality Checks

- Static analysis: `./scripts/tools/static_analysis.sh`
- Unit tests (default): `go test ./...`
- Fast checks (no Docker/DB): `make verify-fast`
- Smoke test (requires a running API): `make smoke`
- Full validation flow (when needed): `make verify-fast` -> `make integration` -> `make smoke` -> `make bdd`

## Test Layers

- `unit` (default): `go test ./...`
  - No Docker/DB required
  - Integration/BDD/Smoke tests are excluded via build tags
- `integration`: `make integration` or `go test -tags=integration ./...`
  - Requires prepared DB/environment
  - Includes PostgreSQL repository tests
- `smoke`: `make smoke` or `go test -tags=smoke ./tests/smoke`
  - Requires a running API (for example `docker compose up`)
- `bdd`: `make bdd` (godog)
  - Separate scenario layer, requires DB/environment

## Canonical Commands

- API run: `make api-run`
- Go tests: `make api-test`
- Perf (GET degradation): `make perf-degradation-get`
- Perf (full series): `make perf-all`
- Analyze k6 results: `make analysis-copy RESULTS_DIR=<dir>`
- Detailed degradation report: `make analysis-degradation RESULTS_DIR=<dir>`

Canonical directories:
- `scripts/perf/` - perf/load scenario execution
- `scripts/analysis/` - result analysis
- `scripts/tools/` - utility scripts (lint/checks/entrypoint etc.)

## Contributing

- English-only policy: do not add Cyrillic text to code, comments, logs, prompts, scripts, or documentation.
- Run fast checks before pushing: `make verify-fast`
- The Cyrillic guardrail is enforced by `make no-cyrillic-check` and included in `make verify-fast`.
- If environment-dependent layers are relevant, run: `make integration`, `make smoke`, `make bdd`.
