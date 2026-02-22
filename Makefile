export PGCLIENTENCODING = UTF-8

# Настройки подключения к основной БД
DB_NAME = cinema
DB_USER = postgres
DB_PASS = postgres
DB_HOST = localhost
DB_PORT = 5432
DB_SSL = disable

# Настройки тестовой БД
TEST_DB_NAME = cinema_test
TEST_DB_USER = postgres
TEST_DB_PASS = postgres

# Параметры подключения
PSQL_CONN = psql "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASS) dbname=$(DB_NAME) sslmode=$(DB_SSL)"
TEST_PSQL_CONN = psql "host=$(DB_HOST) port=$(DB_PORT) user=$(TEST_DB_USER) password=$(TEST_DB_PASS) dbname=$(TEST_DB_NAME) sslmode=$(DB_SSL)"

.PHONY: db-init db-clean test-init test-clean run test test-v swagger cover docker-up docker-down ci-docker

# Инициализация основной БД
db-init: db-clean
	@echo "Инициализация основной БД..."
	@$(PSQL_CONN) -q -f sql/001_create_app_roles.sql
	@$(PSQL_CONN) -q -f sql/002_create_main.sql
	@$(PSQL_CONN) -q -f sql/004_set_app_roles_privileges.sql
	@$(PSQL_CONN) -q -f sql/006_seed_main.sql
	@echo "Основная БД готова!"

# Очистка основной БД
db-clean:
	@echo "Очистка основной БД..."
	@$(PSQL_CONN) -q -f sql/revoke_app_roles_privileges.sql || true
	@$(PSQL_CONN) -q -f sql/drop_main.sql || true
	@$(PSQL_CONN) -q -f sql/drop_app_roles.sql || true

# Инициализация тестовой БД
test-init: test-clean
	@echo "Инициализация тестовой БД..."
	@$(TEST_PSQL_CONN) -q -f sql/003_create_test_roles.sql
	@$(TEST_PSQL_CONN) -q -f sql/002_create_main.sql
	@$(TEST_PSQL_CONN) -q -f sql/005_set_test_roles_privileges.sql
	@echo "Тестовая БД готова!"

# Очистка тестовой БД
test-clean:
	@echo "Очистка тестовой БД..."
	@$(TEST_PSQL_CONN) -q -f sql/revoke_test_roles_privileges.sql || true
	@$(TEST_PSQL_CONN) -q -f sql/drop_main.sql || true
	@$(TEST_PSQL_CONN) -q -f sql/drop_test_roles.sql || true

# Запуск приложения
run: 
	@echo "Запуск приложения..."
	@go run ./cmd/api/main.go

# Обновление документации Swagger
swagger:
	@echo "Обновление документации Swagger..."
	@go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go

# Сборка и запуск в Docker
docker-up:
	@echo "Сборка и запуск контейнеров..."
	@docker compose up --build -d

# Остановка Docker окружения
docker-down:
	@echo "Остановка контейнеров..."
	@docker compose down -v

# CI: генерирует Swagger и проверяет запуск приложения в Docker
ci-docker: swagger
	@echo "Проверка запуска приложения в Docker..."
	@set -eu; \
	export COMPOSE_PROJECT_NAME=ci-check; \
	export POSTGRES_CPUSET=0-1 APP_CPUSET=2-3; \
	export HOST_POSTGRES_PORT=15433 HOST_APP_PORT=18080; \
	export POSTGRES_CONTAINER_NAME=ci_postgres_db APP_CONTAINER_NAME=ci_go_app; \
	trap 'docker compose logs --no-color --tail=200 app postgres >/dev/null 2>&1 || true; docker compose down -v --remove-orphans >/dev/null 2>&1 || true' EXIT; \
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
		echo "Приложение не стало доступно по /metrics"; \
		docker compose ps; \
		docker compose logs --no-color --tail=200 app postgres; \
		exit 1; \
	fi; \
	echo "Приложение доступно по /metrics"

# Анализ покрытия тестами
cover:
	@echo "Анализ покрытия тестами..."
	@go test -cover -coverprofile=c.out ./...
	@go tool cover -html=c.out

# Запуск тестов
test:
	@echo "Запуск тестов..."
	@go test -cover -count=1 ./...

# Запуск тестов с верификацией
test-v:
	@echo "Запуск тестов с верификацией..."
	@go test -cover -count=1 -v ./...

# Сборка приложения
build:
	@echo "Сборка приложения..."
	@go build -o bin/api ./cmd/api

# Очистка билдов
clean:
	@echo "Очистка билдов..."
	@if exist bin rmdir /s /q bin
	@if exist c.out del c.out

# Запуск линтера
lint:
	@echo "Запуск линтера..."
	@golangci-lint run ./...

# Форматирование кода
fmt:
	@echo "Форматирование кода..."
	@gofmt -w .

# Запуск всех проверок (линтер + тесты)
check: lint test

# Запуск с горячей перезагрузкой (если установлен air)
air:
	@echo "Запуск с горячей перезагрузкой..."
	@air

# Показать зависимости
deps:
	@echo "Анализ зависимостей..."
	@go mod graph

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  run     - Запуск приложения"
	@echo "  test    - Запуск тестов"
	@echo "  test-v  - Запуск тестов с детальным выводом"
	@echo "  swagger - Обновление Swagger документации"
	@echo "  docker-up - Сборка и запуск Docker окружения"
	@echo "  docker-down - Остановка Docker окружения"
	@echo "  ci-docker - Swagger + проверка запуска в Docker (для CI)"
	@echo "  cover   - Анализ покрытия тестами"
	@echo "  build   - Сборка приложения"
	@echo "  clean   - Очистка билдов"
	@echo "  lint    - Запуск линтера"
	@echo "  fmt     - Форматирование кода"
	@echo "  check   - Запуск всех проверок"
	@echo "  air     - Запуск с горячей перезагрузкой"
	@echo "  deps    - Показать зависимости"
	@echo "  help    - Показать эту справку"
