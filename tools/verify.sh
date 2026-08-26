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
FIX_FMT=0
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
  --fix-fmt         Apply gofumpt/goimports formatting in place across all
                    modules, then exit 0 without running any other leg.
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
        --fix-fmt)   FIX_FMT=1 ;;
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

TOOLS_BIN="$ROOT/.cache/tools/bin"
GOLANGCI="$TOOLS_BIN/golangci-lint-$GOLANGCI_LINT_VERSION"

# Per-tree lint cache. The default (~/.cache/golangci-lint) is shared by every
# worktree, and golangci-lint replays cached issues by package path — so
# linting shared/go/model in the main repo surfaces stale findings recorded
# from a sibling worktree whose files no longer exist. Keying the cache to
# $ROOT gives each worktree its own and removes the crosstalk.
#
# The per-tree cache is NOT enough alone: `golangci-lint run` also takes an
# exclusive flock on $TMPDIR/golangci-lint.lock, a machine-global path no cache
# setting isolates. Two concurrent runs sharing a $TMPDIR contend on it and the
# loser exits 3 with "parallel golangci-lint is running" and no findings — a
# spurious failure, not a lint result. `run` is passed --allow-parallel-runners
# below to skip that lock. What the lock protects against is concurrent writers
# to ONE cache, which the per-tree keying already rules out.
export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-$ROOT/.cache/golangci-lint}"

ensure_golangci() {
    [ -x "$GOLANGCI" ] && return 0
    mkdir -p "$TOOLS_BIN" "$GOLANGCI_LINT_CACHE"

    # Fast path: download the pinned prebuilt release and verify it against the
    # release's published SHA256 checksums. ~10s vs the multi-minute source
    # build. Falls back to `go install` when unavailable (no curl/sha256sum,
    # unknown platform, or offline).
    local ver="${GOLANGCI_LINT_VERSION#v}" os="" arch="" asset url tmp
    case "$(uname -s)" in
        Linux) os=linux ;;
        Darwin) os=darwin ;;
    esac
    case "$(uname -m)" in
        x86_64 | amd64) arch=amd64 ;;
        arm64 | aarch64) arch=arm64 ;;
    esac

    if [ -n "$os" ] && [ -n "$arch" ] \
        && command -v curl >/dev/null 2>&1 && command -v sha256sum >/dev/null 2>&1; then
        asset="golangci-lint-${ver}-${os}-${arch}.tar.gz"
        url="https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}"
        echo "verify.sh: downloading golangci-lint $GOLANGCI_LINT_VERSION prebuilt ($os-$arch) into $TOOLS_BIN ..."
        tmp="$(mktemp -d)"
        if curl -sSfL "$url/$asset" -o "$tmp/$asset" \
            && curl -sSfL "$url/golangci-lint-${ver}-checksums.txt" -o "$tmp/checksums.txt" \
            && (cd "$tmp" && grep " ${asset}\$" checksums.txt | sha256sum -c - >/dev/null 2>&1) \
            && tar -xzf "$tmp/$asset" -C "$tmp" \
            && mv "$tmp/golangci-lint-${ver}-${os}-${arch}/golangci-lint" "$GOLANGCI"; then
            chmod +x "$GOLANGCI"
            rm -rf "$tmp"
            return 0
        fi
        echo "verify.sh: WARNING — prebuilt download/verify failed; falling back to 'go install' (slower)." >&2
        rm -rf "$tmp"
    fi

    if ! command -v go >/dev/null 2>&1; then
        echo "verify.sh: ERROR — cannot fetch prebuilt golangci-lint and no go toolchain for the source fallback" >&2
        return 1
    fi
    echo "verify.sh: installing golangci-lint $GOLANGCI_LINT_VERSION from source into $TOOLS_BIN ..."
    tmp="$(mktemp -d)"
    GOBIN="$tmp" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_LINT_VERSION" || return 1
    mv "$tmp/golangci-lint" "$GOLANGCI"
    rm -rf "$tmp"
}

# Two layers per module: the formatter layer (gofumpt + goimports, checked with
# --diff so verify.sh never rewrites the tree behind you) and the linter layer
# (the `standard` set). No --new-from-rev: home-hub's tree is clean, so the
# gate is absolute rather than relative to a baseline.
leg_lint() {
    ensure_golangci || return 1
    local rc=0 moddir rel fmt_out
    while IFS= read -r moddir; do
        rel="${moddir#"$ROOT"/}"
        echo "--- lint: $rel"
        if fmt_out="$(cd "$moddir" && "$GOLANGCI" fmt --diff -c "$ROOT/.golangci.yml" ./... 2>&1)" \
            && [ -z "$fmt_out" ]; then
            :
        else
            echo "FMT FAIL — $rel (run: tools/verify.sh --fix-fmt)"
            printf '%s\n' "$fmt_out" | head -40
            rc=1
        fi
        if ! (cd "$moddir" && "$GOLANGCI" run --allow-parallel-runners -c "$ROOT/.golangci.yml" ./...); then
            echo "LINT FAIL — $rel"
            rc=1
        fi
    done < <(discover_modules)
    return "$rc"
}

leg_frontend_build() {
    (cd "$ROOT/frontend" && npm ci && npm run build && npm test)
}

leg_eslint() {
    (cd "$ROOT/frontend" && npx eslint .)
}

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

if [ "${FIX_FMT:-0}" -eq 1 ]; then
    ensure_golangci || exit 1
    while IFS= read -r moddir; do
        echo "--- fmt: ${moddir#"$ROOT"/}"
        (cd "$moddir" && "$GOLANGCI" fmt -c "$ROOT/.golangci.yml" ./...) || exit 1
    done < <(discover_modules)
    exit 0
fi

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
