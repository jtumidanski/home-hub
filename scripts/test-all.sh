#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Running all tests..."

for service in auth-service account-service productivity-service weather-service recipe-service calendar-service category-service shopping-service package-service; do
  echo "Testing $service..."
  cd "$ROOT_DIR/services/$service"
  go test ./... -count=1
done

for pkg in model tenant logging database server auth http testing; do
  echo "Testing shared/go/$pkg..."
  cd "$ROOT_DIR/shared/go/$pkg"
  go test ./... -count=1
done

echo "Running frontend tests..."
cd "$ROOT_DIR/frontend"
npm test

echo "All tests complete."
