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
	@$(PSQL_CONN) -q -f sql/001_create_app_roles.sql
	@$(PSQL_CONN) -q -f sql/002_create_main.sql
	@$(PSQL_CONN) -q -f sql/004_set_app_roles_privileges.sql
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
run: db-init
	@echo "Запуск приложения..."
	@go run .

# Запуск тестов
test: test-init
	@echo "Запуск тестов..."
	@go test -cover -count=1 ./...

# Запуск тестов с верификацией
test-v: test-init
	@echo "Запуск тестов с верификацией..."
	@go test -cover -count=1 -v ./...