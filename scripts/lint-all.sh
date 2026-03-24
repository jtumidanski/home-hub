#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Linting all services..."

for service in auth-service account-service productivity-service; do
  echo "Linting $service..."
  cd "$ROOT_DIR/services/$service"
  golangci-lint run ./...
done

for pkg in model tenant logging database server auth http testing; do
  echo "Linting shared/go/$pkg..."
  cd "$ROOT_DIR/shared/go/$pkg"
  golangci-lint run ./...
done

echo "Linting frontend..."
cd "$ROOT_DIR/frontend"
npx eslint .

echo "All linting complete."
