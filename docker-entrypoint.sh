#!/bin/sh
set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-cinema}
DB_USER=${DB_USER:-postgres}
DB_PASS=${DB_PASS:-postgres}

wait_for_db() {
  echo "Waiting for database to be ready..."
  until PGPASSWORD="$DB_PASS" pg_isready -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" >/dev/null 2>&1; do
    sleep 1
  done
  echo "Database is ready!"
}

run_sql_first_existing() {
  for sql_file in "$@"; do
    if [ -f "$sql_file" ]; then
      PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f "$sql_file"
      return 0
    fi
  done
  echo "No SQL file found among: $*" >&2
  return 1
}

wait_for_db

echo "Running database migrations..."

run_sql_first_existing sql/prod_db_clean/001_revoke_app_roles_privileges.sql sql/revoke_app_roles_privileges.sql
run_sql_first_existing sql/prod_db_clean/002_drop_main.sql sql/drop_main.sql
run_sql_first_existing sql/prod_db_clean/003_drop_app_roles.sql sql/drop_app_roles.sql
run_sql_first_existing sql/prod_db_init/004_create_app_roles.sql sql/001_create_app_roles.sql
run_sql_first_existing sql/prod_db_init/005_create_main.sql sql/002_create_main.sql
run_sql_first_existing sql/prod_db_init/006_set_app_roles_privileges.sql sql/004_set_app_roles_privileges.sql
run_sql_first_existing sql/prod_db_init/007_seed_main.sql sql/006_seed_main.sql

echo "Starting application..."
exec ./api
