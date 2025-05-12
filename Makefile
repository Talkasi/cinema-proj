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

.PHONY: db-init db-clean test-init test-clean run test test-v

# Инициализация основной БД
db-init: db-clean
	@echo "Инициализация основной БД..."
	@$(PSQL_CONN) -q -f sql/_prepare_db.sql
	@echo "Основная БД готова!"

# Очистка основной БД
db-clean:
	@echo "Очистка основной БД..."
	@$(PSQL_CONN) -q -f _clear_all_db.sql || true

# Инициализация тестовой БД
test-init: test-clean
	@echo "Инициализация тестовой БД..."
	@$(TEST_PSQL_CONN) -q -f sql/_prepare_test_db.sql
	@echo "Тестовая БД готова!"

# Очистка тестовой БД
test-clean:
	@echo "Очистка тестовой БД..."
	@$(TEST_PSQL_CONN) -q -f sql/_clear_all_test_db.sql || true

# Запуск приложения
run: db-init
	@echo "Запуск приложения..."
	@go run main.go

# Запуск тестов
test: test-init
	@echo "Запуск тестов..."
	@go test -cover ./...

# Запуск тестов с верификацией
test-v: test-init
	@echo "Запуск тестов с верификацией..."
	@go test -cover -v ./...