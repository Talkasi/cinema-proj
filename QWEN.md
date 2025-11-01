# Cinema Management System

## Project Overview

This is a Go-based cinema management system that provides a REST API for managing cinema operations. The application follows clean architecture principles with separate layers for data access, domain services, and data transfer objects (DTOs).

### Key Features
- User authentication and management
- Movie and genre management
- Hall and seat configuration
- Movie show scheduling
- Ticket sales and management
- Reviews and feedback
- Screen and seat type management
- Comprehensive API documentation via Swagger
- Prometheus metrics for monitoring
- Performance testing with K6

### Architecture
- **Data Access Layer**: PostgreSQL database with pgx driver
- **Domain Layer**: Business logic services
- **DTO Layer**: HTTP handlers using Chi router
- **API Documentation**: Swagger integration
- **Authentication**: JWT-based authentication system

### Technologies Used
- **Backend**: Go 1.24.0
- **Database**: PostgreSQL
- **Web Framework**: Chi router
- **Authentication**: JWT (golang-jwt)
- **Documentation**: Swagger
- **Monitoring**: Prometheus
- **Testing**: Testify
- **Containerization**: Docker and Docker Compose
- **Performance Testing**: K6

## Building and Running

### Prerequisites
- Go 1.24.0+
- Docker and Docker Compose
- PostgreSQL (for local development)

### Setup

1. **Clone the repository** (already done)

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Initialize the database** (for local development):
   ```bash
   make db-init
   ```

4. **Run the application**:
   ```bash
   make run
   ```

### Using Docker

The application can also be run using Docker Compose:

```bash
# Build and start the services
docker-compose up --build

# Run in background
docker-compose up -d --build
```

The API will be available at `http://localhost:8080`.

### Available Make Commands

- `make run` - Run the application
- `make build` - Build the application binary
- `make test` - Run tests with coverage
- `make test-v` - Run tests with verbose output
- `make db-init` - Initialize the main database
- `make db-clean` - Clean the main database
- `make test-init` - Initialize the test database
- `make test-clean` - Clean the test database
- `make swagger` - Update Swagger documentation
- `make cover` - Analyze test coverage
- `make lint` - Run linter
- `make fmt` - Format code
- `make check` - Run all checks (lint + test)
- `make help` - Show available commands

### Environment Variables

The application uses the following environment variables:

- `ADDR` - Server address (default: ":8080")
- `JWT_SECRET` - Secret key for JWT tokens (default: "secret-key")
- `TOKEN_DURATION` - JWT token duration (default: "24h")
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_NAME` - Database name
- `DB_USER` - Database user
- `DB_PASS` - Database password
- `DB_SSL` - SSL mode for database connection

## Development Conventions

### Code Structure
- `/cmd` - Main application entry points
  - `/cmd/api` - API server implementation
  - `/cmd/tui` - Terminal user interface (if available)
- `/internal` - Internal application code
  - `/internal/dataAccess` - Database configuration and repository implementations
  - `/internal/domain` - Business logic services
  - `/internal/dto` - HTTP handlers and data transfer objects
  - `/internal/utils` - Utility functions
  - `/internal/tui` - Terminal user interface code
- `/sql` - Database schema and seed files
- `/docs` - API documentation
- `/scripts` - Deployment and utility scripts
- `/monitoring` - Monitoring configuration
- `/k6_results` - Performance test results

### Database Schema
The database schema is defined in SQL files under the `/sql` directory:
- `001_create_app_roles.sql` - Application roles
- `002_create_main.sql` - Main database tables
- `003_create_test_roles.sql` - Test database roles
- `004_set_app_roles_privileges.sql` - Application role privileges
- `005_set_test_roles_privileges.sql` - Test role privileges
- `006_seed_main.sql` - Sample data for main database
- `007_seed_test.sql` - Sample data for test database

### Performance Testing
The project includes comprehensive performance testing capabilities:
- K6 scripts in `/scripts` directory
- Automated benchmarking with `run_all_benchmarks.sh`
- Single test runner with `run_single_test.sh`
- Performance analysis with `analyze_degradation.py`
- Resource usage monitoring during tests

## API Endpoints

The API is documented using Swagger and available at:
- API documentation: `http://localhost:8080/api/v1/swagger/index.html`
- API base path: `/api/v1`
- Metrics endpoint: `/metrics` (Prometheus)

The API includes endpoints for managing:
- Users
- Genres
- Halls
- Movie shows
- Movies
- Screen types
- Seat types
- Seats
- Tickets
- Reviews

## Testing

The application includes comprehensive testing:
- Unit tests using the Testify framework
- Database testing with a dedicated test database
- Performance testing with K6
- Automated test execution with coverage reports

Run tests using:
```bash
make test      # Run tests with coverage
make test-v    # Run tests with verbose output
```

## Deployment

The application is designed for containerized deployment with Docker:

1. **Build the Docker image**:
   ```bash
   docker build -t cinema-api .
   ```

2. **Run with Docker Compose**:
   ```bash
   docker-compose up -d
   ```