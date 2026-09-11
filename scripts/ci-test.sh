#!/bin/bash
# Thin delegate — see scripts/ci-build.sh (task-055).
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/../tools/verify.sh" --only test "$@"
