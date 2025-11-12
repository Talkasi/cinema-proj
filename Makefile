export PGCLIENTENCODING = UTF-8

# Параметры подключения
PSQL_CONN = psql "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASS) dbname=$(DB_NAME) sslmode=$(DB_SSL)"
TEST_PSQL_CONN = psql "host=$(DB_HOST) port=$(DB_PORT) user=$(TEST_DB_USER) password=$(TEST_DB_PASS) dbname=$(TEST_DB_NAME) sslmode=$(DB_SSL)"

.PHONY: db-init db-clean test-init test-clean run test test-v swagger cover

# Инициализация основной БД
db-init: db-clean
	@echo "Инициализация основной БД..."
	@$(PSQL_CONN) -q -f sql/prod_db_init/004_create_app_roles.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/005_create_main.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/006_set_app_roles_privileges.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/007_seed_main.sql
	@echo "Основная БД готова!"

# Очистка основной БД
db-clean:
	@echo "Очистка основной БД..."
	@$(PSQL_CONN) -q -f sql/prod_db_clean/001_revoke_app_roles_privileges.sql || true
	@$(PSQL_CONN) -q -f sql/prod_db_clean/002_drop_main.sql || true
	@$(PSQL_CONN) -q -f sql/prod_db_clean/003_drop_app_roles.sql || true

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
run: swagger
	@echo "Запуск приложения..."
	@go run ./cmd/api/main.go

docker-up: docker-down
	@docker compose up

docker-down:
	@docker compose down -v
	@docker image rm cinema-proj-app || true

# Обновление документации Swagger
swagger:
	@echo "Обновление документации Swagger..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	$$(go env GOPATH)/bin/swag init -g cmd/api/main.go --output docs

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

# Запуск BDD тестов
bdd-test: db-clean
	@$(PSQL_CONN) -q -f sql/prod_db_init/004_create_app_roles.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/005_create_main.sql
	@$(PSQL_CONN) -q -f sql/prod_db_init/006_set_app_roles_privileges.sql
	@echo "Запуск BDD тестов..."
	@cd tests/bdd && go test -v

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  run     - Запуск приложения"
	@echo "  test    - Запуск тестов"
	@echo "  test-v  - Запуск тестов с детальным выводом"
	@echo "  bdd-test - Запуск BDD тестов"
	@echo "  swagger - Обновление Swagger документации"
	@echo "  cover   - Анализ покрытия тестами"
	@echo "  build   - Сборка приложения"
	@echo "  clean   - Очистка билдов"
	@echo "  lint    - Запуск линтера"
	@echo "  fmt     - Форматирование кода"
	@echo "  check   - Запуск всех проверок"
	@echo "  air     - Запуск с горячей перезагрузкой"
	@echo "  deps    - Показать зависимости"
	@echo "  help    - Показать эту справку"