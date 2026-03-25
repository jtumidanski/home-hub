#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building recipe-service..."
cd "$ROOT_DIR/services/recipe-service"
go build ./...
echo "recipe-service built."
