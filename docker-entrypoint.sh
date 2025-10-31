#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -p 5912 -U postgres; do
  echo "Waiting for database connection..."
  sleep 2
done

echo "PostgreSQL is ready!"

echo "Running database migrations..."
dbmate --url "$DATABASE_URL" --migrations-dir ./db/migrations up

echo "Migrations completed successfully!"

echo "Starting Facturme API..."
exec ./facturme-api
