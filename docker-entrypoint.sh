#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
# Extract host from DATABASE_URL (format: postgres://user:pass@host:port/db)
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')

# Default to postgres if extraction fails
DB_HOST=${DB_HOST:-postgres}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}

echo "Connecting to PostgreSQL at $DB_HOST:$DB_PORT as $DB_USER..."

# Wait for PostgreSQL to be ready
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "Waiting for database connection..."
  sleep 2
done

echo "PostgreSQL is ready!"

echo "Running database migrations..."
dbmate --url "$DATABASE_URL" --migrations-dir ./db/migrations up

echo "Migrations completed successfully!"

echo "Starting FacturMe API..."
exec ./facturme-api
