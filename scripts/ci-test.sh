#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "Testing all shared modules..."
for pkg in model tenant logging database server auth http testing; do
  echo "Testing shared/go/$pkg..."
  cd "$ROOT_DIR/shared/go/$pkg"
  go test ./... -count=1
done

echo "Testing all services..."
for service in auth-service account-service productivity-service recipe-service calendar-service weather-service category-service shopping-service package-service; do
  echo "Testing $service..."
  cd "$ROOT_DIR/services/$service"
  go test ./... -count=1
done

echo "CI tests complete."
