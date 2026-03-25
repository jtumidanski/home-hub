#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building weather-service..."
cd "$ROOT_DIR/services/weather-service"
go build ./...
echo "weather-service built."
