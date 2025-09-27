#!/bin/sh

# This script waits for the database to be ready, then runs migrations.

# Exit immediately if a command exits with a non-zero status.
set -e

MIGRATE_CMD="migrate -path internal/migration -database postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "Waiting for database to be ready at ${DB_HOST}:${DB_PORT}..."

# Loop until the database is ready to accept connections
# We use nc (netcat) which needs to be installed in our Docker image.
until nc -z -v -w30 ${DB_HOST} ${DB_PORT}
do
  echo "Waiting for database connection..."
  sleep 1
done
echo "Database is ready."

echo "Running database migrations..."
$MIGRATE_CMD up
echo "Migrations finished."