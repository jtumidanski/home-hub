#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building auth-service..."
cd "$ROOT_DIR/services/auth-service"
go build ./...
echo "auth-service built."
