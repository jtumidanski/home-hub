#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building and starting local environment..."
cd "$ROOT_DIR/deploy/compose"
docker compose --env-file "$ROOT_DIR/.env" up -d --build
echo "Local environment started."
