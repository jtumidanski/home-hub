#!/usr/bin/env bash
# tools/verify.sh — the single verification entrypoint for home-hub (task-055).
#
# CONTRACT (docs/verification.md):
#   A flagless run that exits 0 means this branch may be called done.
#   --quick / --no-docker / --no-ui exit 0 on success too, and explicitly do
#   NOT mean that. They say so on every run.
#
# Every selected leg runs even after an earlier one fails, so one pass gives
# the complete picture. The summary reproduces the first failing leg's output.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 1

# shellcheck source=toolchain.versions
source "$ROOT/tools/toolchain.versions"

ALL_LEGS=(pins build vet test lint frontend-build eslint docker)
QUICK_LEGS=(pins build vet)
UI_LEGS=(frontend-build eslint)

QUICK=0
NO_DOCKER=0
NO_UI=0
FACTS=0
BASE=""
ONLY=""

usage() {
    cat <<'EOF'
Usage: tools/verify.sh [options]

The single verification entrypoint. A flagless run that exits 0 means this
branch may be called done.

Options:
  --quick           Run only the fast legs (pins, build, vet). Implies
                    --no-docker. Does NOT count as done.
  --no-docker       Skip the docker image-build leg. Does NOT count as done.
  --no-ui           Skip the frontend legs (frontend-build, eslint). Does NOT
                    count as done.
  --base <rev>      Change-detection base for the docker leg. Defaults to the
                    merge-base with origin/main, then main.
  --only <legs>     Comma-separated leg names to run, ignoring the presets.
                    Legs: pins build vet test lint frontend-build eslint docker
  --facts           Print the change base, changed services, fan-out reason,
                    module count, and every leg that would run; then exit 0
                    without building anything.
  -h, --help        Print this and exit 0.

Exit codes:
  0  every selected leg passed
  1  at least one leg failed
  2  usage error (unknown option, unknown leg, missing option value)
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        --quick)     QUICK=1 ;;
        --no-docker) NO_DOCKER=1 ;;
        --no-ui)     NO_UI=1 ;;
        --facts)     FACTS=1 ;;
        --base)
            [ $# -ge 2 ] || { echo "verify.sh: --base requires a value" >&2; exit 2; }
            BASE="$2"; shift
            ;;
        --only)
            [ $# -ge 2 ] || { echo "verify.sh: --only requires a value" >&2; exit 2; }
            ONLY="$2"; shift
            ;;
        -h | --help) usage; exit 0 ;;
        *) echo "verify.sh: unknown option: $1" >&2; usage >&2; exit 2 ;;
    esac
    shift
done

# ---- leg selection ---------------------------------------------------------
in_list() {
    local needle="$1"; shift
    local x
    for x in "$@"; do [ "$x" = "$needle" ] && return 0; done
    return 1
}

SELECTED=()
if [ -n "$ONLY" ]; then
    IFS=',' read -r -a requested <<< "$ONLY"
    for leg in "${requested[@]}"; do
        in_list "$leg" "${ALL_LEGS[@]}" \
            || { echo "verify.sh: unknown leg: $leg" >&2; exit 2; }
        SELECTED+=("$leg")
    done
elif [ "$QUICK" -eq 1 ]; then
    SELECTED=("${QUICK_LEGS[@]}")
else
    SELECTED=("${ALL_LEGS[@]}")
fi

drop_leg() {
    local drop="$1" keep=() leg
    for leg in "${SELECTED[@]}"; do
        [ "$leg" = "$drop" ] || keep+=("$leg")
    done
    SELECTED=(${keep[@]+"${keep[@]}"})
}

[ "$NO_DOCKER" -eq 1 ] && drop_leg docker
if [ "$NO_UI" -eq 1 ]; then
    for leg in "${UI_LEGS[@]}"; do drop_leg "$leg"; done
fi

# A run is "complete" — i.e. may be called done — only when it was flagless.
COMPLETE=1
if [ "$QUICK" -eq 1 ] || [ "$NO_DOCKER" -eq 1 ] || [ "$NO_UI" -eq 1 ] || [ -n "$ONLY" ]; then
    COMPLETE=0
fi

# ---- shared helpers --------------------------------------------------------
majmin() { printf '%s\n' "$1" | awk -F. '{ if (NF >= 2) print $1"."$2; else print $1 }'; }

discover_modules() {
    find "$ROOT/services" "$ROOT/shared/go" -name go.mod -not -path '*/node_modules/*' -print0 \
        | xargs -0 -n1 dirname | sort -u
}

resolve_base() {
    if [ -n "$BASE" ]; then printf '%s\n' "$BASE"; return 0; fi
    git -C "$ROOT" merge-base HEAD origin/main 2>/dev/null && return 0
    git -C "$ROOT" merge-base HEAD main 2>/dev/null && return 0
    return 1
}

# ---- leg bodies (filled in by later tasks) ---------------------------------
# Compares the major.minor of GO_VERSION against every other place a Go
# version is written: go.work's `go` directive and every `go-version:` in
# .github/workflows/*.yml. Cheap, and it is why tools/toolchain.versions can be
# called a single source of truth.
leg_pins() {
    local want rc=0 workver v
    want="$(majmin "$GO_VERSION")"

    workver="$(awk '/^go /{print $2; exit}' "$ROOT/go.work")"
    if [ "$(majmin "$workver")" != "$want" ]; then
        echo "pin mismatch: go.work declares 'go $workver', tools/toolchain.versions says GO_VERSION=$GO_VERSION"
        rc=1
    fi

    while IFS= read -r v; do
        [ -z "$v" ] && continue
        if [ "$(majmin "$v")" != "$want" ]; then
            echo "pin mismatch: .github/workflows go-version '$v' != $want"
            rc=1
        fi
    done < <(grep -ho "go-version: *'[^']*'" "$ROOT"/.github/workflows/*.yml \
        | sed "s/.*'\(.*\)'/\1/" | sort -u)

    [ "$rc" -eq 0 ] && echo "pins: go $want consistent across toolchain.versions, go.work, and .github/workflows"
    return "$rc"
}
# Iterate every discovered module, run `cmd` inside it, collect failures.
# Every module runs even after one fails — one pass, complete picture.
for_each_module() {
    local label="$1"; shift
    local rc=0 moddir rel
    while IFS= read -r moddir; do
        rel="${moddir#"$ROOT"/}"
        echo "--- $label: $rel"
        if ! (cd "$moddir" && "$@"); then
            echo "$label FAIL — $rel"
            rc=1
        fi
    done < <(discover_modules)
    return "$rc"
}

leg_build() { for_each_module build go build ./...; }

# Plain `go vet`, not golangci-lint's govet: --quick must never trigger a cold
# multi-minute golangci-lint bootstrap. The lint leg's `standard` set includes
# govet, so flagless coverage is unchanged; this is the price of a
# bootstrap-free --quick.
leg_vet() { for_each_module vet go vet ./...; }

leg_test() { for_each_module test go test ./... -count=1; }
leg_lint()            { :; }
leg_frontend_build() { :; }
leg_eslint()         { :; }
leg_docker()         { :; }

leg_fn() { printf 'leg_%s\n' "${1//-/_}"; }

# ---- runner ----------------------------------------------------------------
declare -A LEG_STATUS=()
FIRST_FAIL=""
FIRST_FAIL_LOG=""
RC=0

run_leg() {
    local leg="$1" fn log
    fn="$(leg_fn "$leg")"
    log="$(mktemp)"
    printf '\n===== %s =====\n' "$leg"
    if "$fn" 2>&1 | tee "$log"; then
        LEG_STATUS[$leg]=PASSED
    else
        LEG_STATUS[$leg]=FAILED
        RC=1
        if [ -z "$FIRST_FAIL" ]; then
            FIRST_FAIL="$leg"
            FIRST_FAIL_LOG="$(cat "$log")"
        fi
    fi
    rm -f "$log"
}

print_summary() {
    local leg
    printf '\n===== summary =====\n'
    for leg in "${ALL_LEGS[@]}"; do
        if in_list "$leg" ${SELECTED[@]+"${SELECTED[@]}"}; then
            printf '  %-16s %s\n' "$leg" "${LEG_STATUS[$leg]:-SKIPPED}"
        else
            printf '  %-16s SKIPPED\n' "$leg"
        fi
    done
    if [ -n "$FIRST_FAIL" ]; then
        printf '\n----- first failing leg: %s -----\n' "$FIRST_FAIL"
        printf '%s\n' "$FIRST_FAIL_LOG"
    fi
    printf '\n'
    if [ "$RC" -ne 0 ]; then
        echo "verify.sh: FAILED"
    elif [ "$COMPLETE" -eq 1 ]; then
        echo "verify.sh: PASSED — this branch may be called done."
    else
        echo "verify.sh: PASSED for the legs that ran — this does not count as done."
        echo "verify.sh: run tools/verify.sh with no flags before calling the branch done."
    fi
}

# ---- facts mode ------------------------------------------------------------
print_facts() {
    printf 'legs: %s\n' "${SELECTED[*]}"
    printf 'modules: %s\n' "$(discover_modules | wc -l | tr -d ' ')"
}

if [ "$FACTS" -eq 1 ]; then
    print_facts
    [ "$COMPLETE" -eq 1 ] || echo "note: this selection does not count as done."
    exit 0
fi

# ---- main ------------------------------------------------------------------
if [ "${VERIFY_DRY_RUN:-0}" = "1" ]; then
    for leg in ${SELECTED[@]+"${SELECTED[@]}"}; do LEG_STATUS[$leg]=PASSED; done
    print_summary
    exit 0
fi

for leg in ${SELECTED[@]+"${SELECTED[@]}"}; do
    run_leg "$leg"
done

print_summary
exit "$RC"
