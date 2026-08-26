#!/usr/bin/env bash
# Contract tests for tools/verify.sh (FR-V1–V4). These exercise flag handling
# and leg selection only — no leg does real work under VERIFY_DRY_RUN=1.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
V="$ROOT/tools/verify.sh"
passed=0
failed=0

check() {
    local desc="$1" expected="$2" actual="$3"
    if [ "$expected" = "$actual" ]; then
        passed=$((passed + 1))
    else
        failed=$((failed + 1))
        echo "FAIL: $desc"
        echo "  expected: $expected"
        echo "  actual:   $actual"
    fi
}

# --- exit codes -------------------------------------------------------------
"$V" --help >/dev/null 2>&1
check "--help exits 0" 0 $?

"$V" -h >/dev/null 2>&1
check "-h exits 0" 0 $?

"$V" --no-such-flag >/dev/null 2>&1
check "unknown option exits 2" 2 $?

"$V" --only no-such-leg >/dev/null 2>&1
check "unknown --only leg exits 2" 2 $?

"$V" --base >/dev/null 2>&1
check "--base with no value exits 2" 2 $?

# --- leg selection (--facts prints the plan and exits without building) -----
legs() { VERIFY_DRY_RUN=1 "$V" --facts "$@" 2>/dev/null | sed -n 's/^legs: *//p'; }

check "flagless runs all eight legs" \
    "pins build vet test lint frontend-build eslint docker" "$(legs)"
check "--no-docker drops only docker" \
    "pins build vet test lint frontend-build eslint" "$(legs --no-docker)"
check "--quick is pins,build,vet" \
    "pins build vet" "$(legs --quick)"
check "--quick implies --no-docker" \
    "pins build vet" "$(legs --quick --no-docker)"
check "--no-ui drops the two frontend legs" \
    "pins build vet test lint docker" "$(legs --no-ui)"
check "--only selects exactly what is named" \
    "test lint" "$(legs --only test,lint)"

# --- "does not count as done" wording (FR-V2) -------------------------------
out="$(VERIFY_DRY_RUN=1 "$V" --quick 2>&1)"
case "$out" in
    *"does not count as done"*) passed=$((passed + 1)) ;;
    *) failed=$((failed + 1)); echo "FAIL: --quick must state it does not count as done" ;;
esac

out="$(VERIFY_DRY_RUN=1 "$V" --no-docker 2>&1)"
case "$out" in
    *"does not count as done"*) passed=$((passed + 1)) ;;
    *) failed=$((failed + 1)); echo "FAIL: --no-docker must state it does not count as done" ;;
esac

out="$(VERIFY_DRY_RUN=1 "$V" 2>&1)"
case "$out" in
    *"does not count as done"*) failed=$((failed + 1)); echo "FAIL: flagless must NOT disclaim" ;;
    *) passed=$((passed + 1)) ;;
esac

# --- pins leg (real, not dry-run) -------------------------------------------
"$V" --only pins >/dev/null 2>&1
check "pins leg passes against the committed toolchain" 0 $?

# --- module discovery covers everything go.work declares --------------------
want_modules="$(awk '/^\t\.\//{gsub(/^\t\.\//,""); print}' "$ROOT/go.work" | sort)"
got_modules="$(cd "$ROOT" && tools/verify.sh --facts 2>/dev/null | sed -n 's/^modules: //p')"
check "discovery finds every go.work module" "$(printf '%s\n' "$want_modules" | wc -l | tr -d ' ')" "$got_modules"

echo "passed: $passed  failed: $failed"
[ "$failed" -eq 0 ]
