#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "Building all shared modules..."
for pkg in model tenant logging database server auth http testing; do
  echo "Building shared/go/$pkg..."
  go build ./shared/go/$pkg/...
done

echo "Building all services..."
go build ./services/auth-service/...
go build ./services/account-service/...
go build ./services/productivity-service/...
go build ./services/recipe-service/...

echo "Building frontend..."
cd "$ROOT_DIR/frontend"
npm ci
npm run build

echo "CI build complete."
