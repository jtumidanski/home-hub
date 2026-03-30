#!/usr/bin/env bash
# Local deployment script for Home Hub.
# Loads the root .env file so docker compose services get their secrets.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"

if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: $ENV_FILE not found."
  echo "Copy .env.example and fill in the required values:"
  echo "  cp $ROOT_DIR/.env.example $ENV_FILE"
  exit 1
fi

cd "$SCRIPT_DIR"
exec docker compose --env-file "$ENV_FILE" up --build "$@"
