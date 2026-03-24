#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Starting local environment..."
cd "$ROOT_DIR/deploy/compose"
docker compose up -d
echo "Local environment started."
