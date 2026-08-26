#!/bin/bash
# Thin delegate. tools/verify.sh owns all leg logic and the one module list
# (task-055). This script exists so existing muscle memory and any external
# caller keep working; it is not a second implementation.
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/../tools/verify.sh" --only build,frontend-build "$@"
