#!/bin/bash
set -e

echo "running database migrations..."
goose -dir ${GOOSE_DIR} up

echo "starting server..."
exec "$@"