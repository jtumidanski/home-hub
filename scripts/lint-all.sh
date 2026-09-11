#!/bin/bash
# Thin delegate — see scripts/ci-build.sh (task-055). This script previously
# invoked an unpinned `golangci-lint` from PATH against a hardcoded 17-module
# list. It now delegates to the pinned binary and the discovered module set
# (FR-L4).
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/../tools/verify.sh" --only lint,eslint "$@"
