#!/usr/bin/env bash
# View logs for the local Home Hub stack.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"

cd "$SCRIPT_DIR"
exec docker compose --env-file "$ENV_FILE" logs "$@"
