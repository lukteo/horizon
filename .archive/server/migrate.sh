#!/bin/bash

# Script to run database migrations
# Usage: ./migrate.sh up or ./migrate.sh down

set -e

if [ -z "$DATABASE_URL" ]; then
    echo "DATABASE_URL environment variable is required"
    exit 1
fi

if [ $# -eq 0 ]; then
    echo "Usage: $0 {up|down|status}"
    exit 1
fi

case "$1" in
    up)
        echo "Running database migrations up..."
        go run -tags 'postgres' vendor/github.com/pressly/goose/v3/cmd/goose/main.go -dir migrations postgres "$DATABASE_URL" up
        ;;
    down)
        echo "Running database migrations down..."
        go run -tags 'postgres' vendor/github.com/pressly/goose/v3/cmd/goose/main.go -dir migrations postgres "$DATABASE_URL" down
        ;;
    status)
        echo "Checking migration status..."
        go run -tags 'postgres' vendor/github.com/pressly/goose/v3/cmd/goose/main.go -dir migrations postgres "$DATABASE_URL" status
        ;;
    *)
        echo "Invalid command. Use: up, down, or status"
        exit 1
        ;;
esac