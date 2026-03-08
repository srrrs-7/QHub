#!/bin/bash
set -euo pipefail

echo "=== Database Migration ==="
echo "DB_URI is ${DB_URI:+set}"

# Wait for database to be ready
echo "Waiting for database to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0
# Helper function to check database availability
check_db() {
    if [ -n "${DB_URI:-}" ]; then
        pg_isready -d "${DB_URI}" > /dev/null 2>&1
    else
        pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USERNAME}" > /dev/null 2>&1
    fi
}

until check_db; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: Database not ready after ${MAX_RETRIES} attempts"
        exit 1
    fi
    echo "Database not ready, waiting... (${RETRY_COUNT}/${MAX_RETRIES})"
    sleep 2
done
echo "Database is ready!"

# Run Atlas migration
echo "Running Atlas migrations..."
if [ -n "${DB_URI:-}" ]; then
    atlas migrate apply \
        --env ci \
        --url "${DB_URI}"
else
    atlas migrate apply \
        --env ci \
        --url "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DBNAME}?sslmode=require"
fi

RESULT=$?

if [ $RESULT -eq 0 ]; then
    echo "=== Migration completed successfully ==="
else
    echo "=== Migration failed with exit code: ${RESULT} ==="
    exit $RESULT
fi
