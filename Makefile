export PGCLIENTENCODING = UTF-8

# Main database connection settings
DB_NAME = cinema
DB_USER = postgres
DB_PASS = postgres
DB_HOST = localhost
DB_PORT = 5432
DB_SSL = disable

# Test database settings
TEST_DB_NAME = cinema_test
TEST_DB_USER = postgres
TEST_DB_PASS = postgres

# Connection parameters
PSQL_CONN = psql "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASS) dbname=$(DB_NAME) sslmode=$(DB_SSL)"
TEST_PSQL_CONN = psql "host=$(DB_HOST) port=$(DB_PORT) user=$(TEST_DB_USER) password=$(TEST_DB_PASS) dbname=$(TEST_DB_NAME) sslmode=$(DB_SSL)"

.PHONY: db-init db-clean test-init test-clean run test test-v swagger cover docker-up docker-down ci-docker api-run api-test gofmt-check scripts-check no-cyrillic-check verify-fast integration perf-degradation-get perf-all analysis-copy analysis-degradation smoke bdd bdd-test

# Initialize main DB
db-init: db-clean
	@echo "Initialize main DB..."
	@$(PSQL_CONN) -q -f sql/prod_db_init/004_create_app_roles.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/005_create_main.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/006_set_app_roles_privileges.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/007_seed_main.sql
	@echo "Main DB is ready!"

# Clean main DB
db-clean:
	@echo "Clean main DB..."
	@$(PSQL_CONN) -q -f sql/prod_db_clean/001_revoke_app_roles_privileges.sql || true
	@$(PSQL_CONN) -q -f sql/prod_db_clean/002_drop_main.sql || true
	@$(PSQL_CONN) -q -f sql/prod_db_clean/003_drop_app_roles.sql || true

# Initialize test DB
test-init: test-clean
	@echo "Initialize test DB..."
	@$(TEST_PSQL_CONN) -q -f sql/003_create_test_roles.sql
	@$(TEST_PSQL_CONN) -q -f sql/002_create_main.sql
	@$(TEST_PSQL_CONN) -q -f sql/005_set_test_roles_privileges.sql
	@echo "Test DB is ready!"

# Clean test DB
test-clean:
	@echo "Clean test DB..."
	@$(TEST_PSQL_CONN) -q -f sql/revoke_test_roles_privileges.sql || true
	@$(TEST_PSQL_CONN) -q -f sql/drop_main.sql || true
	@$(TEST_PSQL_CONN) -q -f sql/drop_test_roles.sql || true

# Run application
run: swagger
	@echo "Run application..."
	@go run ./cmd/api/main.go

# Canonical API run alias
api-run: run

# Update Swagger documentation
swagger:
	@echo "Update Swagger documentation..."
	@go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go

# Build and run in Docker
docker-up:
	@echo "Building and starting containers..."
	@docker compose up --build -d

# Stop Docker environment
docker-down:
	@echo "Stopping containers..."
	@docker compose down -v

# CI: generates Swagger and checks Docker startup
ci-docker: swagger
	@echo "Checking application startup in Docker..."
	@set -eu; \
	tmp_env_created=0; \
	if [ ! -f .env ]; then : > .env; tmp_env_created=1; fi; \
	export COMPOSE_PROJECT_NAME=ci-check; \
	export POSTGRES_CPUSET=0-1 APP_CPUSET=2-3; \
	export HOST_POSTGRES_PORT=15433 HOST_APP_PORT=18080; \
	export POSTGRES_CONTAINER_NAME=ci_postgres_db APP_CONTAINER_NAME=ci_go_app; \
	trap 'docker compose logs --no-color --tail=200 app postgres >/dev/null 2>&1 || true; docker compose down -v --remove-orphans >/dev/null 2>&1 || true; if [ "$$tmp_env_created" -eq 1 ]; then rm -f .env; fi' EXIT; \
	docker compose down -v --remove-orphans >/dev/null 2>&1 || true; \
	docker rm -f ci_postgres_db ci_go_app >/dev/null 2>&1 || true; \
	docker compose up --build -d; \
	docker compose ps; \
	ok=0; \
	for i in $$(seq 1 30); do \
		if curl --fail --silent --show-error http://localhost:18080/metrics >/dev/null; then \
			ok=1; \
			break; \
		fi; \
		sleep 2; \
	done; \
	if [ "$$ok" -ne 1 ]; then \
		echo "Application did not become available at /metrics"; \
		docker compose ps; \
		docker compose logs --no-color --tail=200 app postgres; \
		exit 1; \
	fi; \
	echo "Application is available at /metrics"

# Test coverage report
cover:
	@echo "Test coverage report..."
	@go test -cover -coverprofile=c.out ./...
	@go tool cover -html=c.out

# Run unit tests (default, without integration/bdd/smoke)
test:
	@echo "Running tests..."
	@go test -count=1 ./...

# Canonical API/Go test alias
api-test: test

# Run tests with verbose verification
test-v:
	@echo "Run tests with verbose verification..."
	@go test -cover -count=1 -v ./...

# Build application
build:
	@echo "Build application..."
	@go build -o bin/api ./cmd/api

# Clean build artifacts
clean:
	@echo "Clean build artifacts..."
	@if exist bin rmdir /s /q bin
	@if exist c.out del c.out

# Run linter
lint:
	@echo "Run linter..."
	@go vet ./...
	@go fmt ./...
	@~/go/bin/gocyclo -over 10 . || echo "Cyclomatic complexity check passed"
	@golangci-lint run --config .golangci.yml --timeout=5m || echo "golangci-lint completed with issues"

# Format code
fmt:
	@echo "Format code..."
	@gofmt -w .

# Check formatting without rewriting files
gofmt-check:
	@echo "Checking Go formatting (gofmt -l)..."
	@out=$$(find . -type f -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -l); \
	if [ -n "$$out" ]; then \
		echo "Unformatted files:"; \
		echo "$$out"; \
		exit 1; \
	fi

# Shell/Python script checks
scripts-check:
	@echo "Checking shell scripts (bash -n)..."
	@find scripts -type f -name '*.sh' -print0 | xargs -0 -r bash -n
	@echo "Checking Python scripts (py_compile)..."
	@python3 -m py_compile $$(find scripts -type f -name '*.py' | sort)

# EN-only guardrail: fail on Cyrillic characters in repository files (excluding .git objects)
no-cyrillic-check:
	@echo "Checking repository for Cyrillic characters..."
	@if rg -n --pcre2 --hidden --no-ignore-vcs --glob '!.git/**' "\p{Cyrillic}" .; then \
		echo "Cyrillic text is not allowed. Please use English only."; \
		exit 1; \
	fi

# Run all checks (lint + tests)
check: lint test

# Fast local checks (no Docker/DB/smoke)
verify-fast: gofmt-check test scripts-check no-cyrillic-check

# Integration tests (require DB/environment)
integration:
	@echo "Running integration tests (build tag: integration)..."
	@go test -tags=integration -count=1 ./...

# Run with hot reload (if air is installed)
air:
	@echo "Starting hot reload..."
	@air

# Show dependencies
deps:
	@echo "Inspecting dependencies..."
	@go mod graph

# Load/perf scenarios (canonical commands)
perf-degradation-get:
	@echo "Running perf scenario: degradation GET..."
	@./scripts/perf/k6_scenario.sh "GET degradation test" "degradation_get_test" "scripts/k6/degradation_test_get.js" "copy"

perf-all:
	@echo "Running perf benchmark series..."
	@./scripts/perf/benchmarks_all.sh

# Analyze results (set RESULTS_DIR=<dir>)
analysis-copy:
	@test -n "$(RESULTS_DIR)" || (echo "Usage: make analysis-copy RESULTS_DIR=<dir>" && exit 1)
	@python3 ./scripts/analysis/report.py copy "$(RESULTS_DIR)"

analysis-degradation:
	@test -n "$(RESULTS_DIR)" || (echo "Usage: make analysis-degradation RESULTS_DIR=<dir>" && exit 1)
	@python3 ./scripts/analysis/report.py degradation "$(RESULTS_DIR)"

# Smoke test (requires a running API, e.g. docker compose up)
smoke:
	@echo "Running smoke tests (build tag: smoke)..."
	@go test -tags=smoke -v ./tests/smoke

# BDD tests (godog, separate layer)
bdd: bdd-test

# Run BDD tests
bdd-test: db-clean
	@$(PSQL_CONN) -q -f sql/prod_db_init/004_create_app_roles.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/005_create_main.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/006_set_app_roles_privileges.sql
	@echo "Run BDD tests..."
	@cd tests/bdd && go test -tags=bdd -v

# Help
help:
	@echo "Available commands:"
	@echo "  run     - Run application"
	@echo "  api-run - Canonical API run (alias for run)"
	@echo "  test    - Run tests"
	@echo "  api-test - Canonical test run (alias for test)"
	@echo "  integration - Integration tests (tag integration, DB/environment required)"
	@echo "  smoke   - API smoke tests (tag smoke, expects running service)"
	@echo "  bdd     - BDD/godog tests (separate layer, DB/environment required)"
	@echo "  test-v  - Run tests with verbose output"
	@echo "  bdd-test - Run BDD tests"
	@echo "  swagger - Update Swagger docs"
	@echo "  docker-up - Build and start Docker environment"
	@echo "  docker-down - Stop Docker environment"
	@echo "  ci-docker - Swagger + startup check in Docker (for CI)"
	@echo "  cover   - Test coverage report"
	@echo "  build   - Build application"
	@echo "  clean   - Clean build artifacts"
	@echo "  lint    - Run linter"
	@echo "  fmt     - Format code"
	@echo "  gofmt-check - Formatting check (no file changes)"
	@echo "  scripts-check - Shell/Python script checks"
	@echo "  no-cyrillic-check - Fail if Cyrillic text is found in repository files"
	@echo "  check   - Run all checks"
	@echo "  verify-fast - Fast checks without Docker/DB (gofmt, go test, bash -n, py_compile)"
	@echo "  air     - Run with hot reload"
	@echo "  deps    - Show dependencies"
	@echo "  perf-degradation-get - Canonical perf scenario (GET degradation)"
	@echo "  perf-all - Canonical perf benchmark series"
	@echo "  analysis-copy RESULTS_DIR=<dir> - Analyze k6 results (copy)"
	@echo "  analysis-degradation RESULTS_DIR=<dir> - Detailed degradation report"
	@echo "  help    - Show this help"
