#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building all services..."

"$SCRIPT_DIR/build-auth.sh"
"$SCRIPT_DIR/build-account.sh"
"$SCRIPT_DIR/build-productivity.sh"
"$SCRIPT_DIR/build-weather.sh"
"$SCRIPT_DIR/build-recipe.sh"
"$SCRIPT_DIR/build-frontend.sh"

echo "All builds complete."
