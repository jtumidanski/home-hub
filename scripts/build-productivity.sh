#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building productivity-service..."
cd "$ROOT_DIR/services/productivity-service"
go build ./...
echo "productivity-service built."
