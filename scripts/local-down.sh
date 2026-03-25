#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Stopping local environment..."
cd "$ROOT_DIR/deploy/compose"
docker compose down
echo "Local environment stopped."
