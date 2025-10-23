#!/bin/sh
set -e

echo "Waiting for database to be ready..."

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-cinema}
DB_USER=${DB_USER:-postgres}
DB_PASS=${DB_PASS:-postgres}

echo "Database is ready!"

echo "Running database migrations..."

PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/revoke_app_roles_privileges.sql
PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/drop_main.sql
PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/drop_app_roles.sql
PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/001_create_app_roles.sql
PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/002_create_main.sql
PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/004_set_app_roles_privileges.sql
PGPASSWORD=$DB_PASS psql -h "$DB_HOST" -U "$DB_USER" -p "$DB_PORT" -d "$DB_NAME" -f sql/006_seed_main.sql

echo "Starting application..."
exec ./api