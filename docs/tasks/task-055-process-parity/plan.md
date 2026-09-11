# Process Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring `home-hub` to full context-discipline parity with `atlas` — enforcement hooks, the implementer/verifier/reviewer agent trio, a `tools/` directory whose `verify.sh` is the single verification entrypoint, nine owner documents, a restructured `CLAUDE.md`, and a green lint baseline — so that a flagless `tools/verify.sh` exiting 0 means the branch may be called done.

**Architecture:** Five dependency-ordered layers. L0 pins the toolchain (`tools/toolchain.versions`, `.golangci.yml`, `.gitignore`). L1 builds `tools/` — `task-numbers.sh`, `task-brief.sh`, and `verify.sh`, the only genuinely new engineering. L2 copies eight byte-identical hooks plus an adapted `format-on-write.sh` and wires `.claude/settings.json`. L3 adds the `task-implementer`/`task-verifier`/`task-reviewer` agents, `/fix-pr-bug`, and the command wiring that makes them live. L4 ports nine owner documents and restructures `CLAUDE.md`. L5 clears the 290 Go findings and 10 frontend eslint errors so the gate built in L1 exits 0.

**Tech Stack:** Bash (all tooling), Go 1.27.0 across 24 modules under `go.work`, `golangci-lint v2.13.1` (pinned, cached in `.cache/tools/bin/`), Node 24 / eslint 10 / vitest in `frontend/`, Docker (per-service image builds mirroring CI), Claude Code hooks + agents + slash commands under `.claude/`.

**Spec:** `docs/tasks/task-055-process-parity/design.md` (PRD at `docs/tasks/task-055-process-parity/prd.md`; canonical cross-repo specification at `docs/process-parity.md`)

## Global Constraints

- `$ATLAS` = `<atlas-repo-root>`. This is the pinned source for every port. Never read atlas from any other path.
- `$ROOT` = `<repo-root>`. Every command runs from here. Never `cd` to the main repo.
- `GO_VERSION=1.27.0`, `ALPINE_VERSION=3.24`, `GOLANGCI_LINT_VERSION=v2.13.1`. These exact values go in `tools/toolchain.versions`.
- Go module set is **24 modules**, discovered by `find services shared/go -name go.mod`, never hardcoded. 12 services, 12 `shared/go/*` modules.
- Docker image set is **13**: the 12 services (context `.`, dockerfile `services/<svc>/Dockerfile`) plus `frontend` (context `frontend`, dockerfile `frontend/Dockerfile`). Image tag `home-hub-<name>:verify`.
- The eight verbatim hooks must stay byte-identical to atlas (`diff` empty). `format-on-write.sh` is the only hook that may differ.
- Guard hooks (`wait-loop-guard`, `block-home-paths-in-docs`, `turn-budget-guard`, `fork-dispatch-guard`) are **fail-closed**. `format-on-write.sh` is **fail-open in every branch**.
- No hook may reach the network. Only `tools/verify.sh` may (the `golangci-lint` bootstrap).
- No hook or tool may read `.env` or emit credential material.
- Lint fixes must be **behaviour-preserving**. Never introduce a new error path. Anything that would change behaviour is raised, not applied — with the one exception explicitly authorised in Task 19.
- No `--new-from-rev` gating anywhere. home-hub's tree is clean from day one.
- `docs/process-parity.md` is the sole file exempt from the `atlas-(implementer|verifier|reviewer)` name check.

### User decisions taken during planning (supersede the design where they differ)

1. **Command wiring (design §6.4):** port atlas's **full** `.claude/commands/execute-task.md`, genericized — not a minimal graft. `spec-task.md` Step 2 is rewired to `tools/task-numbers.sh next`. FR-C2 is read as "add no *new* command beyond `/fix-pr-bug`".
2. **Docker leg (design §6.2):** mirror CI's per-service matrix. `shared/**` fans out to all 12 service images; a service-only change builds that service's image; `frontend/**` builds the frontend image. Not the FR-V9-literal shared-only rule.
3. **Frontend eslint (design §6.5):** clear **all 10** errors in this task, including the two `react-hooks/rules-of-hooks` in `DashboardDesigner.tsx`. AC-8 must pass without `--no-ui`.

### Deviations from the design, forced by decision 1

Porting atlas's full `execute-task.md` imports its references to `tools/verify.sh --facts`. The design's §4.1 flag list omits `--facts`. Rather than strip a genuinely useful affordance out of the ported command, `verify.sh` **gains `--facts`** (Task 6): it prints the change base, the changed services, the fan-out reason, the module count, and every leg that would run, then exits 0 without building.

The design's leg 3 is "go vet (via golangci-lint `govet`)". This plan implements leg 3 as plain `go vet ./...` instead, because `--quick` must not trigger a multi-minute cold `golangci-lint` bootstrap. The `lint` leg's `standard` set includes `govet`, so flagless coverage is unchanged; the redundancy buys a bootstrap-free `--quick`.

---

## File Structure

**Created:**

| Path | Responsibility |
|---|---|
| `tools/toolchain.versions` | Shell-sourceable toolchain pins. Single source of truth. |
| `tools/task-numbers.sh` (+`_test.sh`) | Collision-safe task numbering across main + worktrees + branches. Verbatim port. |
| `tools/task-brief.sh` | Renders a task brief for `commit-boundary.sh`. Verbatim port. |
| `tools/verify.sh` | The single verification entrypoint. 8 legs, flag contract, linter bootstrap. **New engineering.** |
| `.golangci.yml` | One lint config for all 24 modules. |
| `.claude/hooks/*.sh` (9) | 8 verbatim guards + adapted `format-on-write.sh`. |
| `.claude/agents/task-{implementer,verifier,reviewer}.md` | The capped/isolated/reviewing trio. |
| `.claude/commands/fix-pr-bug.md` | Phase 5. |
| `docs/{agent-dispatch,verification,review-protocol,post-implementation,codemod-vs-agents,slice-first,tooling-conventions,git-workflow}.md` | Eight new owner documents. |

**Modified:**

| Path | Change |
|---|---|
| `.gitignore` | Add `.cache/`. |
| `go.work` | `go 1.26.1` → `go 1.27.0`. |
| `.claude/settings.json` | Full hook wiring + `disableBundledSkills`. |
| `.claude/commands/execute-task.md` | Replaced by genericized atlas port. |
| `.claude/commands/spec-task.md` | Step 2 rewired to `tools/task-numbers.sh next`. |
| `scripts/{ci-build,ci-test,lint-all}.sh` | Become thin `exec tools/verify.sh --only …` delegates. |
| `docs/superpowers-integration.md` | Reconciled with atlas's 185-line version. |
| `CLAUDE.md` | Restructured into the eight-heading rule list. |
| `services/**`, `shared/go/**` | Behaviour-preserving lint fixes only. |
| `frontend/src/**` | eslint fixes only. |

---

## Task 1: Toolchain pins and cache ignore (L0)

**Files:**
- Modify: `.gitignore`
- Create: `tools/toolchain.versions`
- Create: `.golangci.yml`

**Interfaces:**
- Consumes: nothing.
- Produces: `tools/toolchain.versions` exporting `GO_VERSION=1.27.0`, `ALPINE_VERSION=3.24`, `GOLANGCI_LINT_VERSION=v2.13.1` when sourced. `.golangci.yml` at repo root, passed as `-c "$ROOT/.golangci.yml"` by every `golangci-lint` invocation in the repo.

This must land before anything bootstraps the linter, or a ~60 MB binary in `.cache/` gets staged by the first `git add` (design §8.1).

- [ ] **Step 1: Add `.cache/` to `.gitignore`**

Insert after the `# Docker` block in `.gitignore`:

```
# Tool cache (pinned golangci-lint binary, golangci-lint cache)
.cache/
```

- [ ] **Step 2: Verify the ignore works before creating anything under it**

```bash
cd <repo-root>
mkdir -p .cache/tools/bin && touch .cache/tools/bin/probe
git status --porcelain .cache
```

Expected: no output (`.cache` is ignored). Then `rm -rf .cache`.

- [ ] **Step 3: Create `tools/toolchain.versions`**

```bash
mkdir -p tools
```

Write `tools/toolchain.versions`:

```sh
# tools/toolchain.versions — the single source of truth for this repo's
# toolchain pins (task-055). Shell-sourceable KEY=value; read by
# tools/verify.sh and .claude/hooks/format-on-write.sh.
#
# Nothing here is read at build time by go.mod, go.work, Dockerfile ARG
# defaults, or .github/workflows/*.yml — none of those formats can read an
# external file. Those pins are duplicated on purpose and machine-checked
# against this file by tools/verify.sh's `pins` leg. Bump here, run
# `tools/verify.sh --only pins`, fix what it names.
#
# gofumpt/goimports versions are embedded in the golangci-lint release.
# Node tooling (eslint, vite, vitest) is pinned by frontend/package.json and
# its lockfile.
GO_VERSION=1.27.0
ALPINE_VERSION=3.24
GOLANGCI_LINT_VERSION=v2.13.1
```

- [ ] **Step 4: Verify it sources cleanly**

```bash
bash -c 'source tools/toolchain.versions && echo "$GO_VERSION $GOLANGCI_LINT_VERSION"'
```

Expected: `1.27.0 v2.13.1`

- [ ] **Step 5: Create `.golangci.yml`**

Adopted from atlas wholesale per design §5.2, with atlas's single `exclusions.rules` entry dropped (it names an atlas test file) and `local-prefixes` rebound to this module path.

```yaml
# Root golangci-lint v2 config — the single config source for every Go module
# in the repo (task-055). tools/verify.sh runs golangci-lint from each module's
# own directory with this file passed via -c; do not add per-module configs.
#
# Deliberate divergence from atlas: no --new-from-rev gating. atlas gates
# findings to new code because its backlog is too large to clear. home-hub's
# backlog was cleared in task-055, so the tree is clean from day one and the
# flagless tools/verify.sh contract is literally true rather than
# true-modulo-a-baseline. Do not add rev-gating without first re-reading
# docs/verification.md.
#
# Any exclusion added here must carry a comment naming the follow-up that
# removes it.
version: "2"

linters:
  # The v2 `standard` default group: errcheck, govet, ineffassign,
  # staticcheck, unused. Membership is fixed by the version pin in
  # tools/toolchain.versions.
  default: standard

formatters:
  enable:
    - gofumpt
    - goimports
  settings:
    goimports:
      # Group intra-repo imports separately. gofumpt's module-path stays UNSET
      # — one shared config serves 24 modules; golangci-lint derives it per
      # module from each go.mod.
      local-prefixes:
        - github.com/jtumidanski/home-hub
```

- [ ] **Step 6: Verify the module path in `local-prefixes` is real**

```bash
head -1 services/recipe-service/go.mod
head -1 shared/go/model/go.mod
```

Expected: both module paths begin `github.com/jtumidanski/home-hub`. If they do not, correct `local-prefixes` to the actual shared prefix before committing — a wrong prefix silently mis-groups every import in the Task 17 formatter sweep.

- [ ] **Step 7: Verify the YAML parses**

```bash
python3 -c "import yaml,sys; yaml.safe_load(open('.golangci.yml')); print('ok')"
```

Expected: `ok`

- [ ] **Step 8: Commit**

```bash
git add .gitignore tools/toolchain.versions .golangci.yml
git commit -m "chore(task-055): pin Go toolchain and golangci-lint, add .golangci.yml"
```

---

## Task 2: Port `tools/task-numbers.sh` and `tools/task-brief.sh` (L1)

**Files:**
- Create: `tools/task-numbers.sh`, `tools/task-numbers_test.sh`, `tools/task-brief.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `tools/task-numbers.sh next` → prints the next free zero-padded three-digit task number on stdout. `tools/task-numbers.sh` also supports the `list`/collision subcommands its `--help` documents. `tools/task-brief.sh` is invoked by `.claude/hooks/commit-boundary.sh` (Task 9) and by `spec-task.md`/`execute-task.md`. Both are consumed by `.claude/hooks/task-num-collision-detector.sh` (Task 9).

Both scan `docs/tasks/` and `.worktrees/*/docs/tasks/` — exactly home-hub's layout — so they port unmodified. The design verified `task-numbers.sh next` already returns `056` here against this repo's real folders.

- [ ] **Step 1: Copy all three files verbatim**

```bash
cd <repo-root>
A=<atlas-repo-root>
cp "$A/tools/task-numbers.sh" tools/task-numbers.sh
cp "$A/tools/task-numbers_test.sh" tools/task-numbers_test.sh
cp "$A/tools/task-brief.sh" tools/task-brief.sh
chmod +x tools/task-numbers.sh tools/task-numbers_test.sh tools/task-brief.sh
```

- [ ] **Step 2: Verify no atlas-specific paths leaked in**

```bash
grep -n 'atlas\|Chronicle20\|libs/' tools/task-numbers.sh tools/task-brief.sh tools/task-numbers_test.sh
```

Expected: no output. If any line matches, it is a repo-specific path that must be rebound to home-hub's layout (`services/`, `shared/go/`) before proceeding — do not commit a dangling path.

- [ ] **Step 3: Verify syntax**

```bash
bash -n tools/task-numbers.sh && bash -n tools/task-brief.sh && bash -n tools/task-numbers_test.sh && echo "syntax ok"
```

Expected: `syntax ok`

- [ ] **Step 4: Run the test suite (AC-4)**

```bash
bash tools/task-numbers_test.sh
```

Expected: exit 0, a passing summary with `failed: 0`.

- [ ] **Step 5: Verify it produces the right answer against the real repo**

```bash
tools/task-numbers.sh next
```

Expected: `056`. `055` would mean it is not seeing this worktree; anything below `055` means the worktree scan is broken.

- [ ] **Step 6: Verify `task-brief.sh` runs without erroring**

```bash
tools/task-brief.sh task-055-process-parity; echo "exit=$?"
```

Expected: exit 0 and a brief naming task-055. If it exits non-zero, read its usage — do not paper over it, `commit-boundary.sh` depends on this contract.

- [ ] **Step 7: Commit**

```bash
git add tools/task-numbers.sh tools/task-numbers_test.sh tools/task-brief.sh
git commit -m "feat(task-055): port task-numbers.sh and task-brief.sh from atlas"
```

---

## Task 3: `tools/verify.sh` — CLI contract, leg selection, summary (L1)

**Files:**
- Create: `tools/verify.sh`
- Create: `tools/verify_test.sh`
- Modify: `go.work:1`

**Interfaces:**
- Consumes: `tools/toolchain.versions` (Task 1).
- Produces: `tools/verify.sh` with flags `--quick`, `--no-docker`, `--no-ui`, `--base <rev>`, `--only <leg>[,<leg>]`, `--facts`, `-h|--help`. Leg names, fixed for the rest of this plan: `pins build vet test lint frontend-build eslint docker`. Shell functions `discover_modules`, `resolve_base`, `run_leg`, `majmin`, and the arrays `SELECTED`, `LEG_STATUS`. Tasks 4–6 add `leg_*` function bodies; this task defines every one as a stub that returns 0 so the harness is testable in isolation.

This task builds the skeleton and its contract tests. The legs are empty. That is deliberate: FR-V1–V4 are contract requirements testable without any build machinery, and getting them right first means Tasks 4–6 only add leg bodies.

- [ ] **Step 1: Write the failing contract test**

Create `tools/verify_test.sh`:

```bash
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

echo "passed: $passed  failed: $failed"
[ "$failed" -eq 0 ]
```

```bash
chmod +x tools/verify_test.sh
```

- [ ] **Step 2: Run it to verify it fails**

```bash
bash tools/verify_test.sh
```

Expected: FAIL — `tools/verify.sh` does not exist yet, so every check fails.

- [ ] **Step 3: Write `tools/verify.sh` — skeleton with stub legs**

```bash
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
majmin() { printf '%s\n' "$1" | awk -F. '{print $1"."$2}'; }

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
leg_pins()           { :; }
leg_build()          { :; }
leg_vet()            { :; }
leg_test()           { :; }
leg_lint()           { :; }
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
```

```bash
chmod +x tools/verify.sh
```

Note on `--quick` and `--facts` interaction: `print_facts` runs before any leg, so `--facts` never builds. The `COMPLETE` flag drives both the facts note and the summary wording, so the two can never disagree.

- [ ] **Step 4: Run the contract tests to verify they pass**

```bash
bash -n tools/verify.sh && bash tools/verify_test.sh
```

Expected: `passed: 14  failed: 0`, exit 0.

- [ ] **Step 5: Write the failing pin-consistency test**

Append to `tools/verify_test.sh`, before the final `echo "passed: …"` line:

```bash
# --- pins leg (real, not dry-run) -------------------------------------------
"$V" --only pins >/dev/null 2>&1
check "pins leg passes against the committed toolchain" 0 $?
```

- [ ] **Step 6: Run it — it fails because `go.work` still says 1.26.1**

```bash
bash tools/verify_test.sh
```

Expected: this new check passes vacuously right now (the leg is a stub returning 0). That is the point of the next step: implement the leg, watch it fail, then fix `go.work`.

- [ ] **Step 7: Implement `leg_pins`**

Replace the `leg_pins() { :; }` stub in `tools/verify.sh` with:

```bash
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
```

Note: `majmin '1.27'` yields `1.` because awk's `$2` is empty. Guard it — change `majmin` to:

```bash
majmin() { printf '%s\n' "$1" | awk -F. '{ if (NF >= 2) print $1"."$2; else print $1 }'; }
```

- [ ] **Step 8: Run the pins leg and watch it fail**

```bash
tools/verify.sh --only pins
```

Expected: FAILED, with `pin mismatch: go.work declares 'go 1.26.1', tools/toolchain.versions says GO_VERSION=1.27.0`.

- [ ] **Step 9: Fix `go.work`**

Change line 1 of `go.work` from `go 1.26.1` to `go 1.27.0`. This is behaviour-neutral: CI already pins `1.27` in all 13 Go jobs and the local toolchain is 1.27.0.

- [ ] **Step 10: Verify the leg now passes and nothing broke**

```bash
tools/verify.sh --only pins
go build ./services/recipe-service/... && echo "build ok"
bash tools/verify_test.sh
```

Expected: pins `PASSED`; `build ok`; `failed: 0`.

- [ ] **Step 11: Commit**

```bash
git add tools/verify.sh tools/verify_test.sh go.work
git commit -m "feat(task-055): add tools/verify.sh CLI contract and pin-consistency leg"
```

---

## Task 4: `tools/verify.sh` — Go build, vet, and test legs (L1)

**Files:**
- Modify: `tools/verify.sh` (replace the `leg_build`, `leg_vet`, `leg_test` stubs)

**Interfaces:**
- Consumes: `discover_modules`, `run_leg` (Task 3).
- Produces: legs `build`, `vet`, `test` covering all 24 modules by discovery.

Design §6.1: `scripts/ci-build.sh` and `ci-test.sh` cover 9 services + 8 shared modules; `go.work` and CI cover 12 + 12. Encoding the stale set would mean a flagless exit 0 that never compiled `dashboard-service`, `tracker-service`, or `workout-service`. Discovery removes the possibility of that drift rather than fixing this instance of it.

- [ ] **Step 1: Write the failing coverage test**

Append to `tools/verify_test.sh`, before the final summary echo:

```bash
# --- module discovery covers everything go.work declares --------------------
want_modules="$(awk '/^\t\.\//{gsub(/^\t\.\//,""); print}' "$ROOT/go.work" | sort)"
got_modules="$(cd "$ROOT" && tools/verify.sh --facts 2>/dev/null | sed -n 's/^modules: //p')"
check "discovery finds every go.work module" "$(printf '%s\n' "$want_modules" | wc -l | tr -d ' ')" "$got_modules"
```

- [ ] **Step 2: Run it to verify it passes**

```bash
bash tools/verify_test.sh
```

Expected: `failed: 0`, with the module count check comparing 24 to 24. If it reports a mismatch, `discover_modules` is missing a module — fix discovery, never the expectation.

- [ ] **Step 3: Implement the three Go legs**

Replace the `leg_build`, `leg_vet`, and `leg_test` stubs in `tools/verify.sh`:

```bash
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
```

- [ ] **Step 4: Run each leg and verify it passes**

```bash
tools/verify.sh --only build
tools/verify.sh --only vet
tools/verify.sh --only test
```

Expected: each prints 24 `--- <label>: <module>` lines and ends `PASSED for the legs that ran — this does not count as done.` The design measured all 24 modules already green, so a failure here is a regression introduced by Task 3's `go.work` bump — investigate, do not skip.

- [ ] **Step 5: Verify `--quick` now does real work**

```bash
time tools/verify.sh --quick
```

Expected: `pins`, `build`, `vet` all `PASSED`; `test`, `lint`, `frontend-build`, `eslint`, `docker` all `SKIPPED`; the "does not count as done" disclaimer present.

- [ ] **Step 6: Re-run the contract tests**

```bash
bash tools/verify_test.sh
```

Expected: `failed: 0`.

- [ ] **Step 7: Commit**

```bash
git add tools/verify.sh tools/verify_test.sh
git commit -m "feat(task-055): add build, vet, and test legs to verify.sh over all 24 modules"
```

---

## Task 5: `tools/verify.sh` — linter bootstrap, lint leg, frontend legs (L1)

**Files:**
- Modify: `tools/verify.sh` (replace `leg_lint`, `leg_frontend_build`, `leg_eslint` stubs; add `ensure_golangci`)

**Interfaces:**
- Consumes: `tools/toolchain.versions` (`GOLANGCI_LINT_VERSION`), `.golangci.yml` (Task 1).
- Produces: `ensure_golangci()` caching the pinned binary at `.cache/tools/bin/golangci-lint-$GOLANGCI_LINT_VERSION`; the `lint`, `frontend-build`, and `eslint` legs. `format-on-write.sh` (Task 8) depends on that exact cache path and never bootstraps it itself (FR-F3).

This leg is the **only** place in the repository that touches the network (NFR-5).

- [ ] **Step 1: Add `ensure_golangci` and the lint leg**

Insert `ensure_golangci` above the leg bodies in `tools/verify.sh`, ported from atlas's `tools/lint.sh`:

```bash
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
```

Then replace the `leg_lint` stub:

```bash
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
```

- [ ] **Step 2: Add the `--fix-fmt` escape hatch the message just promised**

A failure message must name a command that exists. Add to the flag loop in `tools/verify.sh`, next to `--facts`:

```bash
        --fix-fmt)   FIX_FMT=1 ;;
```

initialise `FIX_FMT=0` beside the other flag defaults, add to `usage()` under `--facts`:

```
  --fix-fmt         Apply gofumpt/goimports formatting in place across all
                    modules, then exit 0 without running any other leg.
```

and insert this immediately before the `if [ "$FACTS" -eq 1 ]` block:

```bash
if [ "${FIX_FMT:-0}" -eq 1 ]; then
    ensure_golangci || exit 1
    while IFS= read -r moddir; do
        echo "--- fmt: ${moddir#"$ROOT"/}"
        (cd "$moddir" && "$GOLANGCI" fmt -c "$ROOT/.golangci.yml" ./...) || exit 1
    done < <(discover_modules)
    exit 0
fi
```

- [ ] **Step 3: Bootstrap the linter and confirm it lands in the ignored cache**

```bash
tools/verify.sh --fix-fmt >/dev/null 2>&1 || true
ls -l .cache/tools/bin/
git status --porcelain .cache
```

Expected: `golangci-lint-v2.13.1` present and executable; `git status` prints nothing. Then discard the formatting it just applied — Task 17 owns that sweep as an isolated commit:

```bash
git checkout -- services shared
git status --porcelain | head
```

Expected: no modified Go files.

- [ ] **Step 4: Confirm the lint leg reports the expected backlog**

```bash
tools/verify.sh --only lint 2>&1 | tail -40
```

Expected: FAILED. This is correct — the 290-finding backlog is real and Tasks 17–19 clear it. Record the module-level failure list; it is the input to those tasks.

- [ ] **Step 5: Implement the two frontend legs**

Replace the `leg_frontend_build` and `leg_eslint` stubs:

```bash
leg_frontend_build() {
    (cd "$ROOT/frontend" && npm ci && npm run build && npm test)
}

leg_eslint() {
    (cd "$ROOT/frontend" && npx eslint .)
}
```

- [ ] **Step 6: Run them**

```bash
tools/verify.sh --only frontend-build
tools/verify.sh --only eslint
```

Expected: `frontend-build` PASSED (design §3.4 measured 105 files / 695 tests green). `eslint` FAILED with 10 errors and 5 warnings — the backlog Task 19 clears.

- [ ] **Step 7: Re-run the contract tests**

```bash
bash tools/verify_test.sh
```

Expected: `failed: 0`. If `--fix-fmt` broke the unknown-option check, the flag was added in the wrong place — it must be a recognised case, not a fallthrough.

- [ ] **Step 8: Commit**

```bash
git add tools/verify.sh
git commit -m "feat(task-055): add pinned golangci-lint bootstrap, lint leg, and frontend legs to verify.sh"
```

---

## Task 6: `tools/verify.sh` — docker leg and `--facts` (L1)

**Files:**
- Modify: `tools/verify.sh` (replace `leg_docker` stub; extend `print_facts`)

**Interfaces:**
- Consumes: `resolve_base` (Task 3).
- Produces: leg `docker`; `--facts` output lines `base:`, `changed-shared:`, `changed-services:`, `fan-out:`, `docker-images:`, `legs:`, `modules:`. `execute-task.md` (Task 12) instructs the operator to run `tools/verify.sh --facts --quick --base <rev>` when a gate behaves unexpectedly; that contract is fixed here.

Per user decision 2 and design §6.2, this mirrors CI's matrix rather than FR-V9's literal shared-only rule. CI builds with `docker build -f <dockerfile> -t home-hub-<svc>:pr <context>` — no `.env`, satisfying NFR-3, unlike `docker compose` in `deploy/compose` which requires `--env-file .env`.

- [ ] **Step 1: Add the image table and target selection**

Insert above the leg bodies in `tools/verify.sh`:

```bash
# Mirrors .github/workflows/pr.yml's docker matrix exactly: 12 services built
# from the repo root, plus the frontend built from frontend/. Keep in sync with
# that workflow — a service added there must be added here.
# Format: <name>|<dockerfile>|<context>
DOCKER_IMAGES=(
    "auth-service|services/auth-service/Dockerfile|."
    "account-service|services/account-service/Dockerfile|."
    "calendar-service|services/calendar-service/Dockerfile|."
    "category-service|services/category-service/Dockerfile|."
    "dashboard-service|services/dashboard-service/Dockerfile|."
    "package-service|services/package-service/Dockerfile|."
    "productivity-service|services/productivity-service/Dockerfile|."
    "recipe-service|services/recipe-service/Dockerfile|."
    "shopping-service|services/shopping-service/Dockerfile|."
    "tracker-service|services/tracker-service/Dockerfile|."
    "weather-service|services/weather-service/Dockerfile|."
    "workout-service|services/workout-service/Dockerfile|."
    "frontend|frontend/Dockerfile|frontend"
)

# Reason the current selection came out the way it did. Set by docker_targets,
# read by print_facts — so --facts can never disagree with a real run.
DOCKER_REASON=""

docker_targets() {
    local base changed entry name
    if ! base="$(resolve_base)"; then
        DOCKER_REASON="no merge base with origin/main or main resolvable; building everything (never fewer)"
        for entry in "${DOCKER_IMAGES[@]}"; do printf '%s\n' "$entry"; done
        return 0
    fi

    changed="$(git -C "$ROOT" diff --name-only "$base"..HEAD)"

    if printf '%s\n' "$changed" | grep -q '^shared/'; then
        DOCKER_REASON="shared/ changed; fanning out to all 12 service images"
        for entry in "${DOCKER_IMAGES[@]}"; do
            name="${entry%%|*}"
            [ "$name" = "frontend" ] && continue
            printf '%s\n' "$entry"
        done
        if printf '%s\n' "$changed" | grep -q '^frontend/'; then
            DOCKER_REASON="$DOCKER_REASON; frontend/ changed too"
            printf '%s\n' "frontend|frontend/Dockerfile|frontend"
        fi
        return 0
    fi

    DOCKER_REASON="per-service change detection against $base"
    for entry in "${DOCKER_IMAGES[@]}"; do
        name="${entry%%|*}"
        if [ "$name" = "frontend" ]; then
            printf '%s\n' "$changed" | grep -q '^frontend/' && printf '%s\n' "$entry"
        else
            printf '%s\n' "$changed" | grep -q "^services/$name/" && printf '%s\n' "$entry"
        fi
    done
    return 0
}
```

Note `^shared/` rather than CI's `shared/go/**`: a superset, and it matches FR-V9's literal wording, so nothing that CI would build is ever missed.

- [ ] **Step 2: Implement the docker leg**

Replace the `leg_docker` stub:

```bash
leg_docker() {
    local rc=0 targets entry name dockerfile context
    targets="$(docker_targets)"
    if [ -z "$targets" ]; then
        echo "docker: no service, shared, or frontend change since the base — nothing to build"
        return 0
    fi
    if ! command -v docker >/dev/null 2>&1; then
        echo "docker: ERROR — docker not found, but the diff selects images to build."
        echo "docker: install docker or re-run with --no-docker (which does not count as done)."
        return 1
    fi
    echo "docker: $DOCKER_REASON"
    while IFS='|' read -r name dockerfile context; do
        [ -z "$name" ] && continue
        echo "--- docker: $name"
        if ! (cd "$ROOT" && docker build -f "$dockerfile" -t "home-hub-$name:verify" "$context"); then
            echo "DOCKER FAIL — $name"
            rc=1
        fi
    done <<< "$targets"
    return "$rc"
}
```

- [ ] **Step 3: Extend `print_facts`**

Replace the `print_facts` body:

```bash
print_facts() {
    local base targets changed
    if base="$(resolve_base)"; then
        printf 'base: %s\n' "$base"
    else
        printf 'base: <unresolvable>\n'
    fi
    changed="$(git -C "$ROOT" diff --name-only "$base"..HEAD 2>/dev/null)"
    if printf '%s\n' "$changed" | grep -q '^shared/'; then
        printf 'changed-shared: yes\n'
    else
        printf 'changed-shared: no\n'
    fi
    printf 'changed-services: %s\n' \
        "$(printf '%s\n' "$changed" | sed -n 's|^services/\([^/]*\)/.*|\1|p' | sort -u | paste -sd, - )"
    targets="$(docker_targets)"
    printf 'fan-out: %s\n' "$DOCKER_REASON"
    printf 'docker-images: %s\n' "$(printf '%s\n' "$targets" | cut -d'|' -f1 | paste -sd, -)"
    printf 'legs: %s\n' "${SELECTED[*]}"
    printf 'modules: %s\n' "$(discover_modules | wc -l | tr -d ' ')"
}
```

`print_facts` calls the same `docker_targets` a real run calls, so it is the real code path with the work removed — it cannot disagree with a run.

- [ ] **Step 4: Demonstrate the docker trigger both ways (AC-9)**

This branch touches `shared/`? Not yet — so start with the negative case:

```bash
tools/verify.sh --facts | grep -E '^(base|changed-shared|fan-out|docker-images):'
```

Expected: `changed-shared: no`, `docker-images:` listing nothing (or only genuinely changed services).

Now the positive case, without committing anything:

```bash
touch shared/go/model/verify_probe.txt
git add shared/go/model/verify_probe.txt
git commit -q -m "tmp: docker fan-out probe"
tools/verify.sh --facts | grep -E '^(changed-shared|fan-out|docker-images):'
```

Expected: `changed-shared: yes`, fan-out reason naming shared, and all 12 service images listed. Then undo:

```bash
git reset --hard HEAD~1
tools/verify.sh --facts | grep '^changed-shared:'
```

Expected: back to `changed-shared: no`. Record both outputs — they are AC-9's evidence, which the acceptance criterion requires be demonstrated, not asserted.

- [ ] **Step 5: Verify one real image actually builds**

```bash
tools/verify.sh --only docker --base HEAD~1 2>&1 | tail -20
```

If the diff selects nothing, force one image directly to prove the command line is right:

```bash
docker build -f services/weather-service/Dockerfile -t home-hub-weather-service:verify .
```

Expected: successful build.

- [ ] **Step 6: Re-run the contract tests**

```bash
bash tools/verify_test.sh
```

Expected: `failed: 0`.

- [ ] **Step 7: Commit**

```bash
git add tools/verify.sh
git commit -m "feat(task-055): add docker leg mirroring CI's matrix and --facts to verify.sh"
```

---

## Task 7: Reduce `scripts/*.sh` to thin delegates (L1)

**Files:**
- Modify: `scripts/ci-build.sh`, `scripts/ci-test.sh`, `scripts/lint-all.sh`

**Interfaces:**
- Consumes: `tools/verify.sh --only <legs>` (Tasks 3–6).
- Produces: nothing new. Removes three hardcoded module lists from the repository.

Design §5.3, answering Open Question 3: invert the dependency. The three scripts' granularity does not match the flag set — `ci-build.sh` bundles the frontend build into the Go build, `lint-all.sh` bundles eslint into the Go lint — so composing *from* them cannot express `--quick`. Delegating leaves exactly one module list in the repo, which permanently removes the drift class §6.1 documents. It also discharges FR-L4 by construction: `lint-all.sh` stops calling an unpinned `golangci-lint` because it stops calling `golangci-lint` at all.

- [ ] **Step 1: Record what the old scripts covered, for the equivalence check**

```bash
grep -o 'go build \./[a-z/-]*' scripts/ci-build.sh | sort > /tmp/old-build-cov.txt
wc -l /tmp/old-build-cov.txt
```

Expected: 9 service lines (the shared loop is separate). Keep this — Step 4 asserts the new coverage is a strict superset.

- [ ] **Step 2: Rewrite `scripts/ci-build.sh`**

```bash
#!/bin/bash
# Thin delegate. tools/verify.sh owns all leg logic and the one module list
# (task-055). This script exists so existing muscle memory and any external
# caller keep working; it is not a second implementation.
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/../tools/verify.sh" --only build,frontend-build "$@"
```

- [ ] **Step 3: Rewrite `scripts/ci-test.sh` and `scripts/lint-all.sh`**

`scripts/ci-test.sh`:

```bash
#!/bin/bash
# Thin delegate — see scripts/ci-build.sh (task-055).
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/../tools/verify.sh" --only test "$@"
```

`scripts/lint-all.sh`:

```bash
#!/bin/bash
# Thin delegate — see scripts/ci-build.sh (task-055). This script previously
# invoked an unpinned `golangci-lint` from PATH against a hardcoded 17-module
# list. It now delegates to the pinned binary and the discovered module set
# (FR-L4).
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/../tools/verify.sh" --only lint,eslint "$@"
```

- [ ] **Step 4: Verify coverage is a superset, not a rewrite that lost something**

```bash
bash -n scripts/ci-build.sh && bash -n scripts/ci-test.sh && bash -n scripts/lint-all.sh && echo "syntax ok"
grep -c golangci-lint scripts/lint-all.sh
tools/verify.sh --facts --only build | sed -n 's/^modules: //p'
```

Expected: `syntax ok`; `0` (AC-19 — no `golangci-lint` invocation left in `lint-all.sh`); `24` modules, versus the old script's 9 services + 8 shared = 17.

- [ ] **Step 5: Verify the delegates run**

```bash
scripts/ci-test.sh 2>&1 | tail -12
```

Expected: the `test` leg runs over 24 modules and the summary prints, including the "does not count as done" line (it is an `--only` run).

- [ ] **Step 6: Commit**

```bash
git add scripts/ci-build.sh scripts/ci-test.sh scripts/lint-all.sh
git commit -m "refactor(task-055): reduce scripts/ci-*.sh and lint-all.sh to verify.sh delegates"
```

---

## Task 8: Port the eight verbatim hooks (L2)

**Files:**
- Create: `.claude/hooks/wait-loop-guard.sh`, `wait-loop-guard_test.sh`, `block-home-paths-in-docs.sh`, `turn-budget.sh`, `turn-budget-guard.sh`, `fork-dispatch-guard.sh`, `commit-boundary.sh`, `task-num-collision-detector.sh`

**Interfaces:**
- Consumes: `tools/task-brief.sh` and `tools/task-numbers.sh` (Task 2) — `commit-boundary.sh` and `task-num-collision-detector.sh` invoke them respectively. Both exist, so no reference is dangling (FR-H4, FR-H5).
- Produces: eight executable hooks, byte-identical to atlas, wired by Task 9's `settings.json`.

These are copied, never edited. NFR-4 exists so a future re-harmonization is a file copy rather than a merge. `turn-budget.sh` and `turn-budget-guard.sh` mention `task-implementer` in operator-facing text — that is the name this repository will use (Task 11), so nothing needs rewriting.

- [ ] **Step 1: Copy all eight**

```bash
cd <repo-root>
A=<atlas-repo-root>
for f in wait-loop-guard.sh wait-loop-guard_test.sh block-home-paths-in-docs.sh \
         turn-budget.sh turn-budget-guard.sh fork-dispatch-guard.sh \
         commit-boundary.sh task-num-collision-detector.sh; do
  cp "$A/.claude/hooks/$f" ".claude/hooks/$f"
  chmod +x ".claude/hooks/$f"
done
ls -l .claude/hooks/
```

Expected: eight new files plus the two pre-existing `skill-activation-prompt.*`.

- [ ] **Step 2: Verify byte-identity (AC-1)**

```bash
A=<atlas-repo-root>
for f in wait-loop-guard.sh wait-loop-guard_test.sh block-home-paths-in-docs.sh \
         turn-budget.sh turn-budget-guard.sh fork-dispatch-guard.sh \
         commit-boundary.sh task-num-collision-detector.sh; do
  if diff -q "$A/.claude/hooks/$f" ".claude/hooks/$f" >/dev/null; then
    echo "identical: $f"
  else
    echo "DIFFERS: $f"
  fi
done
```

Expected: eight `identical:` lines, no `DIFFERS:`.

- [ ] **Step 3: Verify no atlas-specific agent names leaked (AC-2)**

```bash
grep -l 'atlas-' .claude/hooks/*.sh; echo "exit=$?"
```

Expected: no filenames printed (`exit=1` from grep finding nothing is correct).

- [ ] **Step 4: Verify syntax**

```bash
for f in .claude/hooks/*.sh; do bash -n "$f" || echo "SYNTAX FAIL: $f"; done; echo "done"
```

Expected: `done` with no failures.

- [ ] **Step 5: Run the guard's own test suite (AC-3)**

```bash
bash .claude/hooks/wait-loop-guard_test.sh
```

Expected: `passed: 33  failed: 0`, exit 0.

- [ ] **Step 6: Verify the two external dependencies resolve**

```bash
grep -n 'task-brief.sh' .claude/hooks/commit-boundary.sh | head -3
grep -n 'task-numbers.sh' .claude/hooks/task-num-collision-detector.sh | head -3
test -x tools/task-brief.sh && test -x tools/task-numbers.sh && echo "both dependencies exist and are executable"
```

Expected: the references print, and the final line confirms both targets exist.

- [ ] **Step 7: Smoke-test the collision detector against the real repo**

```bash
bash .claude/hooks/task-num-collision-detector.sh </dev/null; echo "exit=$?"
```

Expected: exit 0 with no collision reported. If it reports a collision, that is a genuine finding about the repo's task folders — surface it, do not suppress the hook.

- [ ] **Step 8: Commit**

```bash
git add .claude/hooks/
git commit -m "feat(task-055): port eight enforcement hooks verbatim from atlas"
```

---

## Task 9: Adapt `format-on-write.sh` and wire `.claude/settings.json` (L2)

**Files:**
- Create: `.claude/hooks/format-on-write.sh`
- Modify: `.claude/settings.json`

**Interfaces:**
- Consumes: `tools/toolchain.versions` and `.golangci.yml` (Task 1); the cached binary path `.cache/tools/bin/golangci-lint-$GOLANGCI_LINT_VERSION` established by Task 5's `ensure_golangci`.
- Produces: the ninth hook, and the settings wiring that makes all nine live.

Two bindings change from atlas; structure and every fail-open branch stay byte-for-byte. Frontend rebinds from `services/atlas-ui` + prettier to `frontend/` + eslint, because `frontend/package.json` has **no prettier dependency** (eslint 10 only). The hook still never bootstraps — `verify.sh` does (FR-F3) — so a first Write never eats a multi-minute download.

Settings comes last (design §8.5): AC-12 checks every wired path resolves to an existing executable, so wiring before the hooks exist means a broken session for whoever `/clear`s next.

- [ ] **Step 1: Confirm the FR-F4 premise still holds**

```bash
grep -c prettier frontend/package.json; echo "---"; grep -n '"eslint"' frontend/package.json
```

Expected: `0` prettier matches; an eslint 10 entry. If prettier has since appeared, use it instead of `eslint --fix` — a formatter is the right tool for a formatting hook.

- [ ] **Step 2: Write `.claude/hooks/format-on-write.sh`**

```bash
#!/usr/bin/env bash
# PostToolUse hook — format the file a Write/Edit just touched (task-055).
#
# DELIBERATELY FAIL-OPEN: a local convenience hook must never block an edit.
# Missing toolchain, missing cached binary, unparseable input, tool error — all
# exit 0 silently. tools/verify.sh is the enforcement point. To avoid a
# multi-minute stall on first Write, the hook never bootstraps golangci-lint
# itself; it uses the binary only if tools/verify.sh has already cached it.
set -u

[ -t 0 ] && exit 0

input="$(cat)"
fp="$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty' 2>/dev/null)" || exit 0
[ -z "$fp" ] && exit 0
[ -f "$fp" ] || exit 0

# Fail-open on a non-absolute path: the hook resolves nothing relative to the
# repo, and dirname-walk on a relative path can spin. First-party Write/Edit
# always pass an absolute file_path.
case "$fp" in /*) ;; *) exit 0 ;; esac

ROOT="${CLAUDE_PROJECT_DIR:-$(pwd)}"

case "$fp" in
    *.go)
        # shellcheck source=../../tools/toolchain.versions
        source "$ROOT/tools/toolchain.versions" 2>/dev/null || exit 0
        GOLANGCI="$ROOT/.cache/tools/bin/golangci-lint-${GOLANGCI_LINT_VERSION:-}"
        [ -x "$GOLANGCI" ] || exit 0
        # Format from the file's own module dir so gofumpt sees its go.mod.
        moddir="$(dirname "$fp")"
        while [ "$moddir" != "/" ] && [ ! -f "$moddir/go.mod" ]; do
            moddir="$(dirname "$moddir")"
        done
        [ -f "$moddir/go.mod" ] || exit 0
        (cd "$moddir" && "$GOLANGCI" fmt -c "$ROOT/.golangci.yml" "$fp") >/dev/null 2>&1 || true
        ;;
    */frontend/*.ts | */frontend/*.tsx)
        # --no-install is what keeps this fail-open: no node_modules, no
        # formatter, silent exit 0. frontend/ has no prettier dependency
        # (eslint 10 only), so --fix is the formatter here.
        (cd "$ROOT/frontend" && npx --no-install eslint --fix "$fp") >/dev/null 2>&1 || true
        ;;
esac

exit 0
```

```bash
chmod +x .claude/hooks/format-on-write.sh
```

- [ ] **Step 3: Verify every fail-open branch (FR-F5, NFR-1)**

```bash
H=.claude/hooks/format-on-write.sh
bash -n "$H" && echo "syntax ok"

echo '' | bash "$H"; echo "empty input   -> $?"
echo 'not json'  | bash "$H"; echo "bad json      -> $?"
echo '{}'        | bash "$H"; echo "no file_path  -> $?"
echo '{"tool_input":{"file_path":"relative/path.go"}}' | bash "$H"; echo "relative path -> $?"
echo '{"tool_input":{"file_path":"/nonexistent/x.go"}}' | bash "$H"; echo "missing file  -> $?"
```

Expected: `syntax ok` and every case `-> 0`.

- [ ] **Step 4: Verify it actually formats a Go file**

```bash
cat > /tmp/fmt_probe.go <<'EOF'
package model
import "fmt"
func   Probe( )  { fmt.Println( "x" ) }
EOF
cp /tmp/fmt_probe.go shared/go/model/fmt_probe.go
echo "{\"tool_input\":{\"file_path\":\"$PWD/shared/go/model/fmt_probe.go\"}}" \
  | CLAUDE_PROJECT_DIR="$PWD" bash .claude/hooks/format-on-write.sh
cat shared/go/model/fmt_probe.go
rm -f shared/go/model/fmt_probe.go
```

Expected: the printed file is gofumpt-formatted (`func Probe() {`, import block normalised). If unchanged, the cached binary is missing — run `tools/verify.sh --fix-fmt` once to bootstrap, then retry.

- [ ] **Step 5: Write the new `.claude/settings.json`**

Preserves the existing `UserPromptSubmit` → `skill-activation-prompt.sh` wiring and the `enabledPlugins` block (FR-S3), adds `disableBundledSkills` (FR-S1), and wires every hook at the same events atlas uses (FR-S2).

```json
{
  "disableBundledSkills": true,
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/task-num-collision-detector.sh"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/skill-activation-prompt.sh"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/block-home-paths-in-docs.sh"
          }
        ]
      },
      {
        "matcher": "Agent",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/fork-dispatch-guard.sh"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/wait-loop-guard.sh"
          }
        ]
      },
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/turn-budget-guard.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/format-on-write.sh"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/commit-boundary.sh"
          }
        ]
      },
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/turn-budget.sh"
          }
        ]
      }
    ]
  },
  "enabledPlugins": {
    "superpowers@claude-plugins-official": true
  }
}
```

- [ ] **Step 6: Verify the settings file (AC-12)**

```bash
jq . .claude/settings.json >/dev/null && echo "valid json"
jq -r '.disableBundledSkills' .claude/settings.json
jq -r '.enabledPlugins["superpowers@claude-plugins-official"]' .claude/settings.json
jq -r '.hooks | to_entries[] | .value[] | .hooks[] | .command' .claude/settings.json \
  | sed "s|\$CLAUDE_PROJECT_DIR|$PWD|" \
  | while read -r p; do
      if [ -x "$p" ]; then echo "ok   $p"; else echo "MISSING/NOT-EXEC $p"; fi
    done
```

Expected: `valid json`; `true`; `true`; nine `ok` lines and zero `MISSING/NOT-EXEC`.

- [ ] **Step 7: Commit**

```bash
git add .claude/hooks/format-on-write.sh .claude/settings.json
git commit -m "feat(task-055): adapt format-on-write.sh to home-hub and wire all nine hooks"
```

---

## Task 10: Port the `task-implementer` / `task-verifier` / `task-reviewer` agents (L3)

**Files:**
- Create: `.claude/agents/task-implementer.md`, `.claude/agents/task-verifier.md`, `.claude/agents/task-reviewer.md`

**Interfaces:**
- Consumes: `tools/verify.sh` flags (Tasks 3–6) — `task-verifier` runs `tools/verify.sh --quick --base <rev>` and reads its summary block.
- Produces: three agent definitions dispatched by `execute-task.md` (Task 12) as `subagent_type: task-implementer` / `task-verifier` / `task-reviewer`.

The rules port unchanged; only illustrations move. FR-D3 is the governing constraint: **a rule is never deleted because its example does not transfer.** Where atlas's example has no home-hub analogue, write a neutral one.

- [ ] **Step 1: Copy the three agents**

```bash
cd <repo-root>
A=<atlas-repo-root>
cp "$A/.claude/agents/task-implementer.md" .claude/agents/task-implementer.md
cp "$A/.claude/agents/task-verifier.md"    .claude/agents/task-verifier.md
cp "$A/.claude/agents/task-reviewer.md"    .claude/agents/task-reviewer.md
```

- [ ] **Step 2: Inventory every atlas-specific reference that must be rebound**

```bash
grep -nE 'atlas|WZ|IDA|packet|opcode|libs/|atlas-ui|--facts|--all|--no-ui|bake|Chronicle20' \
  .claude/agents/task-implementer.md .claude/agents/task-verifier.md .claude/agents/task-reviewer.md
```

Work through every hit. Apply this mapping (design §4.4):

| atlas illustration | home-hub replacement |
|---|---|
| packet encode/decode, opcode tables | JSON:API resource and attribute shapes |
| WZ data as ground truth | service source + GORM migrations as ground truth |
| IDA / client binary reading | *no analogue* → neutral: "the authoritative artifact" |
| `libs/atlas-constants` lookup | `shared/go/model`, `shared/go/tenant` lookup |
| `libs/` (module root) | `shared/go/` |
| `services/atlas-ui` + prettier | `frontend/` + eslint |
| `verify.sh --all` / `--facts` / `--no-ui` / docker bake | home-hub's flags: `--quick`, `--no-docker`, `--no-ui`, `--base`, `--only`, `--facts` (all real — see Task 6) |
| cross-service event fan-out | recipe → shopping-list ingredient flow; meal-plan → tracker |
| `tools/lint.sh` | `tools/verify.sh --only lint` |

- [ ] **Step 3: Verify the invariants survived the edit (FR-A4)**

```bash
grep -n '120' .claude/agents/task-implementer.md
grep -n 'PARTIAL' .claude/agents/task-implementer.md
grep -niE 'never edit|does not edit|read-only' .claude/agents/task-verifier.md
grep -niE 'no recursive|does not dispatch|no fan-out|never dispatch' .claude/agents/task-reviewer.md
```

Expected: the 120 tool-call budget and `PARTIAL` hand-back both present in `task-implementer`; the never-edits rule present in `task-verifier`; the no-recursive-fan-out rule present in `task-reviewer`. The 120 cap is enforced by `turn-budget-guard.sh`, one of the verbatim files — changing the number here desynchronizes the hook.

- [ ] **Step 4: Verify no rule was dropped along with its example (FR-D3)**

```bash
A=<atlas-repo-root>
for f in task-implementer task-verifier task-reviewer; do
  echo "=== $f: atlas $(grep -c '^' "$A/.claude/agents/$f.md") lines, home-hub $(grep -c '^' ".claude/agents/$f.md") lines"
  diff <(grep -oE '^\s*[-*] \*\*[^*]+\*\*' "$A/.claude/agents/$f.md") \
       <(grep -oE '^\s*[-*] \*\*[^*]+\*\*' ".claude/agents/$f.md")
done
```

Expected: line counts within ~10% of atlas's; the bolded-rule diff empty or showing only renames that the Step 2 mapping explains. A rule that vanished with no replacement is a defect — restore it with a neutral example.

- [ ] **Step 5: Verify the frontmatter is well-formed and names are generic (AC-10)**

```bash
for f in .claude/agents/task-*.md; do
  echo "--- $f"; sed -n '1,6p' "$f"
done
grep -rlE 'atlas-(implementer|verifier|reviewer)' .claude/agents/; echo "exit=$?"
```

Expected: each file opens with `---`, a `name:` matching its filename, and a `description:`. No files match the `atlas-*` grep.

- [ ] **Step 6: Verify every `docs/` link the agents reference is either created by Task 13/14 or already present**

```bash
grep -ohE 'docs/[a-z0-9-]+\.md' .claude/agents/task-*.md | sort -u
```

Expected: only names from the nine owner documents. Any other `docs/` target is an atlas file this task deliberately does not port (FR-D7) — remove the reference or point it at the home-hub equivalent. Links to files Tasks 13–14 create are fine; Task 21 re-checks them all.

- [ ] **Step 7: Commit**

```bash
git add .claude/agents/task-implementer.md .claude/agents/task-verifier.md .claude/agents/task-reviewer.md
git commit -m "feat(task-055): add task-implementer, task-verifier, and task-reviewer agents"
```

---

## Task 11: Port `/fix-pr-bug` (L3)

**Files:**
- Create: `.claude/commands/fix-pr-bug.md`

**Interfaces:**
- Consumes: `tools/verify.sh` (Tasks 3–6), the agent trio (Task 10).
- Produces: the `/fix-pr-bug` slash command. `docs/post-implementation.md` (Task 14) is its owner document and must describe it.

- [ ] **Step 1: Copy and inventory**

```bash
cd <repo-root>
A=<atlas-repo-root>
cp "$A/.claude/commands/fix-pr-bug.md" .claude/commands/fix-pr-bug.md
grep -nE 'atlas|libs/|packet|WZ|IDA|bake|--facts|--all|--no-ui|tools/lint.sh|Chronicle20' .claude/commands/fix-pr-bug.md
```

- [ ] **Step 2: Rebind every hit**

Apply the Task 10 Step 2 mapping. Additionally:

- Repository paths: `libs/` → `shared/go/`, `services/atlas-ui` → `frontend/`.
- Verification commands: any `tools/lint.sh …` → `tools/verify.sh --only lint`; any docker bake reference → `tools/verify.sh --only docker`.
- Task-folder paths stay `docs/tasks/task-NNN-slug/` — identical in both repos.

- [ ] **Step 3: Verify frontmatter and that referenced commands exist**

```bash
sed -n '1,6p' .claude/commands/fix-pr-bug.md
grep -ohE '/[a-z-]+-task|/fix-pr-bug|tools/[a-z-]+\.sh' .claude/commands/fix-pr-bug.md | sort -u
ls .claude/commands/
```

Expected: valid `---` frontmatter with `description:` and `argument-hint:`; every `/xxx-task` referenced exists in `.claude/commands/`; every `tools/*.sh` referenced exists.

- [ ] **Step 4: Verify no atlas leakage (AC-11 precursor)**

```bash
grep -nE 'atlas' .claude/commands/fix-pr-bug.md; echo "exit=$?"
```

Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add .claude/commands/fix-pr-bug.md
git commit -m "feat(task-055): add /fix-pr-bug Phase 5 command"
```

---

## Task 12: Port atlas's `execute-task.md` and rewire `spec-task.md` (L3)

**Files:**
- Modify: `.claude/commands/execute-task.md` (replaced wholesale)
- Modify: `.claude/commands/spec-task.md:16-23` (Step 2)

**Interfaces:**
- Consumes: the agent trio (Task 10); `tools/verify.sh --quick --base <rev>` and `--facts` (Tasks 3–6); `tools/task-numbers.sh next` (Task 2).
- Produces: `/execute-task` dispatching `subagent_type: task-implementer` per unit, `task-reviewer` per unit, and `task-verifier` for the gate, with the `PARTIAL` continuation protocol. `/spec-task` taking its task number from one source.

Per user decision 1: the **full** atlas `execute-task.md` (≈376 lines, Steps 1, 2, 3, 4, 4a–4f, 5), not a graft. Home-hub's current 60-line version carries worktree-discipline language atlas lacks — that language must survive the replacement, folded into the ported Steps 1–3.

Without this task the trio and `task-numbers.sh` ship inert, directly contradicting two of the PRD's own user stories.

- [ ] **Step 1: Preserve home-hub's worktree-discipline language before overwriting**

```bash
cd <repo-root>
cp .claude/commands/execute-task.md /tmp/hh-execute-task-original.md
grep -nE 'worktree|absolute path|git add -A|destructive|post-commit branch' /tmp/hh-execute-task-original.md
```

Keep this output open. Every rule it names must appear in the new file.

- [ ] **Step 2: Copy atlas's version and inventory what must be rebound**

```bash
A=<atlas-repo-root>
cp "$A/.claude/commands/execute-task.md" .claude/commands/execute-task.md
grep -nE 'atlas|libs/|packet|WZ|IDA|bake|service-wiring-recipe|query-scope-audit|tools/lint.sh|--all\b|Chronicle20' .claude/commands/execute-task.md
```

- [ ] **Step 3: Rebind every hit**

Apply the Task 10 Step 2 mapping, plus these specifics:

- `tools/verify.sh --quick --base <last-gated-commit>` — keep verbatim; both flags exist here (Tasks 3, 6).
- `tools/verify.sh --facts --quick --base <last-gated-commit>` — keep verbatim; `--facts` exists here (Task 6).
- `tools/lint.sh` → `tools/verify.sh --only lint`.
- docker bake → `tools/verify.sh --only docker`.
- atlas doc references that this task does not port (FR-D7: `docs/packets/`, `docs/reverse-engineering.md`, `docs/adding-a-new-service.md`, `docs/observability.md`) → delete the reference, or repoint at the nearest owner document from the nine.
- `service-wiring-recipe.md` / `query-scope-audit.md` examples in Step 4b → replace with a home-hub file-inventory example, e.g. *"say `the JSON:API resource shape in shared/go/model/resource.go` and `the tenant scoping in shared/go/tenant/context.go`, and the implementer reaches for exactly those two files"*.
- Guideline reviewers named in Step 4c → home-hub's real ones: `backend-guidelines-reviewer`, `frontend-guidelines-reviewer`, `plan-adherence-reviewer`.

- [ ] **Step 4: Fold home-hub's worktree discipline back in**

Steps 1–3 of the ported file handle task resolution, worktree check, and input validation. Ensure these home-hub rules from `/tmp/hh-execute-task-original.md` appear verbatim in that region:

- fuzzy task identifiers resolve across `docs/tasks/` and `.worktrees/*/docs/tasks/`;
- if `pwd` is not the task worktree, tell the user to `cd` — do not auto-`cd`, do not create a new worktree;
- every dispatched subagent gets the **worktree absolute path** and must prefix every Bash call with `cd <worktree> && …`;
- post-commit branch verification; no destructive git ops; no `git add -A` / `git add .`;
- `plan.md` **and** `context.md` must both exist before dispatching.

- [ ] **Step 5: Verify the ported command is coherent**

```bash
sed -n '1,6p' .claude/commands/execute-task.md
grep -c '^' .claude/commands/execute-task.md
grep -n 'task-implementer\|task-verifier\|task-reviewer' .claude/commands/execute-task.md | head
grep -nE 'atlas' .claude/commands/execute-task.md; echo "atlas grep exit=$?"
grep -ohE 'docs/[a-z0-9/-]+\.md' .claude/commands/execute-task.md | sort -u
grep -ohE 'tools/[a-z-]+\.sh( --[a-z-]+)*' .claude/commands/execute-task.md | sort -u
```

Expected: valid frontmatter; roughly 350–400 lines; all three agent names present; no `atlas` matches; every `docs/` target among the nine owner documents; every `tools/` invocation using a flag `verify.sh` actually accepts.

- [ ] **Step 6: Verify every flag the command tells the operator to run**

```bash
for f in "--quick" "--facts --quick" "--only lint" "--only docker" "--no-docker" "--no-ui"; do
  # shellcheck disable=SC2086
  tools/verify.sh --facts $f >/dev/null 2>&1
  echo "$f -> exit=$?"
done
```

Expected: every line `exit=0`. A `2` means the command references a flag that does not exist — fix the command or add the flag, never leave the reference dangling.

- [ ] **Step 7: Rewire `spec-task.md` Step 2**

Replace items 1–2 of `### Step 2 — Determine task number and working slug` with:

```markdown
1. Get the next free task number from the one source that knows about main,
   every worktree, and every local branch:

   ```sh
   tools/task-numbers.sh next
   ```

   Do not scan folders by hand. Hand-scanning is what lets two concurrent
   `/spec-task` runs claim the same `NNN`; `task-numbers.sh` also counts
   numbers reserved by an in-flight branch whose `docs/tasks/` folder does not
   exist yet, which a folder scan cannot see.
```

Renumber the surviving items (slug derivation, task identifier composition) accordingly.

- [ ] **Step 8: Verify the rewiring**

```bash
grep -n 'task-numbers.sh' .claude/commands/spec-task.md
grep -n 'find .worktrees -maxdepth 4' .claude/commands/spec-task.md; echo "old hand-scan grep exit=$?"
tools/task-numbers.sh next
```

Expected: the new invocation present; the old hand-scan instruction gone (`exit=1`); `056` printed.

- [ ] **Step 9: Commit**

```bash
git add .claude/commands/execute-task.md .claude/commands/spec-task.md
git commit -m "feat(task-055): port atlas execute-task.md and wire spec-task.md to task-numbers.sh"
```

---

## Task 13: Port four owner documents — verification, tooling, git, slice-first (L4)

**Files:**
- Create: `docs/verification.md`, `docs/tooling-conventions.md`, `docs/git-workflow.md`, `docs/slice-first.md`

**Interfaces:**
- Consumes: `tools/verify.sh`'s full flag and leg contract (Tasks 3–7).
- Produces: four owner documents. `CLAUDE.md`'s trigger table (Task 16) links to all four, and AC-14 requires every link resolve.

Ownership, per PRD §4.7:

| Document | Owns |
|---|---|
| `verification.md` | Gate failures, script/CI disagreement |
| `tooling-conventions.md` | Long-running processes, mechanical repo facts, shell conventions |
| `git-workflow.md` | Committing, pushing, rebasing, stray `main` commits |
| `slice-first.md` | Reading a large document, diff, plan, or tool result |

- [ ] **Step 1: Copy all four and inventory**

```bash
cd <repo-root>
A=<atlas-repo-root>
for f in verification tooling-conventions git-workflow slice-first; do
  cp "$A/docs/$f.md" "docs/$f.md"
done
grep -cnE 'atlas|libs/|packet|WZ|IDA|atlas-ui|bake' docs/verification.md docs/tooling-conventions.md docs/git-workflow.md docs/slice-first.md
```

- [ ] **Step 2: Rewrite `docs/verification.md` around home-hub's gate (FR-D5)**

This document must describe **the `tools/verify.sh` built in Tasks 3–7**, not atlas's. Rewrite, do not merely find-and-replace. It must state, accurately:

- The contract: flagless exit 0 means the branch may be called done; `--quick`, `--no-docker`, `--no-ui`, and `--only` exit 0 on success and explicitly do not.
- The eight legs in order — `pins build vet test lint frontend-build eslint docker` — and which each preset skips.
- That the module set is **discovered**, never listed, and why: `scripts/ci-build.sh` drifted three services behind `go.work` precisely because it hardcoded one (design §6.1).
- That the docker leg mirrors `.github/workflows/pr.yml`'s matrix: `shared/**` fans out to all 12 service images, a service-only change builds that image, `frontend/**` builds the frontend image; and that a service added to that workflow must be added to `DOCKER_IMAGES` in `verify.sh`.
- That every selected leg runs even after an earlier one fails, and the summary reproduces the first failing leg's output verbatim (design §6.3).
- **The no-`--new-from-rev` divergence from atlas, explicitly** (design §5.2, §6.6): home-hub's backlog was cleared in task-055 so the gate is absolute. Do not add rev-gating without re-reading this section.
- That `.claude/hooks/`'s eight verbatim files are **copies of atlas, fixed upstream** — a well-meaning local edit breaks AC-1's `diff`. Re-harmonization is a file copy, not a merge (design §7).
- That `scripts/ci-build.sh`, `ci-test.sh`, and `lint-all.sh` are thin delegates with no logic of their own (Task 7).
- What to do when `verify.sh` and CI disagree: **CI is the authority**; a disagreement is a `verify.sh` coverage defect to fix, not a result to argue with.
- That `--facts` is how you ask the gate what it selected — never reverse-engineer the selection from the source.

- [ ] **Step 3: Genericize the other three**

For `tooling-conventions.md`, `git-workflow.md`, and `slice-first.md`, apply the Task 10 Step 2 mapping. Specifically:

- `libs/` → `shared/go/`; `services/atlas-ui` → `frontend/`.
- Long-running-process examples → home-hub's real slow commands: `tools/verify.sh` flagless, `npm ci && npm run build` in `frontend/`, a 13-image docker leg.
- Mechanical-repo-fact examples → home-hub facts: "which services exist" is `go.work`; "which images CI builds" is `.github/workflows/pr.yml`; "the toolchain pin" is `tools/toolchain.versions`.
- Large-document examples in `slice-first.md` → home-hub's real large artifacts: `docs/architecture.md` (22 K), `docs/process-parity.md` (14.6 K), a 180-file formatter diff.
- `git-workflow.md`'s branch/worktree conventions → home-hub's: task branches named `task-NNN-slug`, worktrees at `.worktrees/task-NNN-slug/`, PRs to `main`.

FR-D3 governs: no rule may be dropped because its example does not transfer. Where there is no home-hub analogue, write a neutral one.

- [ ] **Step 4: Verify no atlas leakage (AC-16)**

```bash
grep -nE 'atlas-ui|\bWZ\b|\bIDA\b|packet|Chronicle20|libs/' \
  docs/verification.md docs/tooling-conventions.md docs/git-workflow.md docs/slice-first.md
echo "exit=$?"
```

Expected: no matches. Any surviving hit must be a deliberate, reviewed cross-repository reference — and if it is, note why in the document itself.

- [ ] **Step 5: Verify every command these documents name actually exists**

```bash
grep -ohE 'tools/[a-z-]+\.sh( --[a-z-]+)*' docs/verification.md docs/tooling-conventions.md docs/git-workflow.md docs/slice-first.md | sort -u
grep -ohE 'scripts/[a-z-]+\.sh' docs/*.md | sort -u | while read -r s; do
  test -f "$s" && echo "ok   $s" || echo "MISSING $s"
done
```

Expected: every `tools/` invocation uses a real flag; every `scripts/` path `ok`.

- [ ] **Step 6: Verify `verification.md` covers the eight legs and the divergence**

```bash
for leg in pins build vet test lint frontend-build eslint docker; do
  grep -q "$leg" docs/verification.md && echo "ok   $leg" || echo "MISSING $leg"
done
grep -qi 'new-from-rev' docs/verification.md && echo "ok   rev-gate divergence documented" || echo "MISSING rev-gate divergence"
grep -qi 'byte-identical\|copies of atlas\|fixed upstream' docs/verification.md && echo "ok   verbatim-hook note" || echo "MISSING verbatim-hook note"
```

Expected: all `ok`.

- [ ] **Step 7: Commit**

```bash
git add docs/verification.md docs/tooling-conventions.md docs/git-workflow.md docs/slice-first.md
git commit -m "docs(task-055): port verification, tooling-conventions, git-workflow, and slice-first owner docs"
```

---

## Task 14: Port four owner documents — dispatch, review, post-implementation, codemod (L4)

**Files:**
- Create: `docs/agent-dispatch.md`, `docs/review-protocol.md`, `docs/post-implementation.md`, `docs/codemod-vs-agents.md`

**Interfaces:**
- Consumes: the agent trio (Task 10), `/fix-pr-bug` (Task 11), `execute-task.md` (Task 12).
- Produces: four owner documents, linked from `CLAUDE.md`'s table (Task 16).

Ownership, per PRD §4.7:

| Document | Owns |
|---|---|
| `agent-dispatch.md` | Model pinning, fan-out vs. fork, handoff decision |
| `review-protocol.md` | Dispatching a reviewer, writing up a review |
| `post-implementation.md` | Phase 5, `/fix-pr-bug` |
| `codemod-vs-agents.md` | Second implementer at the same transformation |

- [ ] **Step 1: Copy all four and inventory**

```bash
cd <repo-root>
A=<atlas-repo-root>
for f in agent-dispatch review-protocol post-implementation codemod-vs-agents; do
  cp "$A/docs/$f.md" "docs/$f.md"
done
grep -cnE 'atlas|libs/|packet|WZ|IDA|atlas-ui|bake' docs/agent-dispatch.md docs/review-protocol.md docs/post-implementation.md docs/codemod-vs-agents.md
```

- [ ] **Step 2: Genericize `docs/agent-dispatch.md` (FR-D6)**

Apply the Task 10 Step 2 mapping, and additionally:

- Name the **generic** `task-implementer` / `task-verifier` / `task-reviewer` throughout.
- **Delete atlas's historical-cutoff note about the `atlas-*` → `task-*` rename.** home-hub never used the `atlas-*` names, so the note is false here. This is FR-D6 and is checked by AC-11.
- Home-hub's real agent roster in any fan-out example: `backend-guidelines-reviewer`, `frontend-guidelines-reviewer`, `plan-adherence-reviewer`, `todo-scanner`, `service-documentation`, plus the new trio.
- The Verification-split section must point at `tools/verify.sh --quick --base <rev>` as built in Tasks 3–6.

- [ ] **Step 3: Genericize the other three**

- `review-protocol.md` — rebind the reviewer roster to home-hub's five existing agents plus `task-reviewer`; audit output path stays `docs/tasks/task-NNN-slug/audit.md` (identical in both repos); backend checklist is DOM-*/SUB-*/SEC-*, frontend is FE-*.
- `post-implementation.md` — describe `/fix-pr-bug` as ported in Task 11; PR target is `main`; the gate is `tools/verify.sh`.
- `codemod-vs-agents.md` — replace atlas's transformation examples with home-hub ones drawn from this very task: the 180-file `golangci-lint fmt` sweep (Task 17) is the canonical codemod, the 78 `errcheck` sites (Task 18) the canonical borderline case, and the 32 hand-judged `staticcheck`/`unused` findings (Task 19) the canonical agent case.

- [ ] **Step 4: Verify no atlas leakage and no `atlas-*` agent names (AC-11, AC-16)**

```bash
grep -nE 'atlas-ui|\bWZ\b|\bIDA\b|packet|Chronicle20|libs/' \
  docs/agent-dispatch.md docs/review-protocol.md docs/post-implementation.md docs/codemod-vs-agents.md
echo "leakage grep exit=$?"
grep -nE 'atlas-(implementer|verifier|reviewer)' docs/*.md | grep -v '^docs/process-parity.md:'
echo "agent-name grep exit=$?"
```

Expected: both `exit=1` (no matches). Only `docs/process-parity.md` may carry the `atlas-*` names.

- [ ] **Step 5: Verify every agent and command these documents name exists**

```bash
grep -ohE '`(task|backend|frontend|plan)-[a-z-]+`' docs/agent-dispatch.md docs/review-protocol.md \
  | tr -d '`' | sort -u | while read -r a; do
      test -f ".claude/agents/$a.md" && echo "ok   $a" || echo "MISSING $a"
    done
grep -ohE '/[a-z-]+-task|/fix-pr-bug|/backend-audit|/audit-plan|/service-doc|/review-todos' docs/*.md \
  | sort -u | while read -r c; do
      test -f ".claude/commands/${c#/}.md" && echo "ok   $c" || echo "MISSING $c"
    done
```

Expected: every line `ok`. A `MISSING` means the document names something this repo does not have — remove the reference or create the thing.

- [ ] **Step 6: Commit**

```bash
git add docs/agent-dispatch.md docs/review-protocol.md docs/post-implementation.md docs/codemod-vs-agents.md
git commit -m "docs(task-055): port agent-dispatch, review-protocol, post-implementation, and codemod-vs-agents owner docs"
```

---

## Task 15: Reconcile `docs/superpowers-integration.md` (L4)

**Files:**
- Modify: `docs/superpowers-integration.md` (76 lines → reconciled)

**Interfaces:**
- Consumes: everything Tasks 10–14 created — the reconciled document is the "which command, agent, or skill do I reach for" index.
- Produces: the ninth owner document. Owns: bare task numbers, skills used outside a phase command.

**Overwriting is explicitly forbidden (FR-D4).** AC-17 is reviewed by a human, not mechanically: the diff against *both* prior versions must show retained content from each.

Design §4.5 established the merge is mechanical and safe: the two share an identical eight-heading skeleton, and atlas's 185 lines are a superset of home-hub's 76 plus `### Task resolution`, `### Artifact location override`, `### Phase 4 context budget`, `### Picking the roster`, `### What a reviewer returns`, and `## Packet Work`.

- [ ] **Step 1: Snapshot both inputs so the reconciliation is auditable**

```bash
cd <repo-root>
A=<atlas-repo-root>
cp docs/superpowers-integration.md /tmp/hh-si-before.md
cp "$A/docs/superpowers-integration.md" /tmp/atlas-si.md
diff <(grep '^#' /tmp/hh-si-before.md) <(grep '^#' /tmp/atlas-si.md)
```

Expected: the shared skeleton, with atlas's six extra headings shown as additions. Confirm `## Packet Work` is among them.

- [ ] **Step 2: Reconcile**

Adopt atlas's structure. For each heading:

- **Shared headings** — keep home-hub's existing body text, then add atlas's subsections around it. Home-hub content survives by construction, which is what AC-17 asks a human to confirm.
- **`### Task resolution`, `### Artifact location override`, `### Phase 4 context budget`, `### Picking the roster`, `### What a reviewer returns`** — adopt from atlas, genericized per the Task 10 Step 2 mapping.
- **`## Packet Work`** — drop entirely. It has no home-hub analogue and FR-D7 excludes `docs/packets/`.

Home-hub content that must survive verbatim or near-verbatim (it has no atlas counterpart):

- the four-phase table with home-hub's exact command names and outputs;
- the `/recipe-to-cooklang` row and the "Personal recipe conversion" entry under **When NOT to Use Superpowers**;
- the **Domain Skills** section naming `backend-dev-guidelines` and `frontend-dev-guidelines` and the `skill-activation-prompt.py` "🎯 SKILL ACTIVATION CHECK" banner;
- the **File Locations Cheat Sheet**, extended with `tools/` and the nine owner docs;
- the pointer to `docs/tasks/task-044-superpowers-integration/design.md`.

Content to add or update from this task: the `task-implementer`/`task-verifier`/`task-reviewer` rows, `/fix-pr-bug` as Phase 5, and `tools/verify.sh` as the gate.

- [ ] **Step 3: Verify both parents are represented (AC-17 evidence)**

```bash
echo "=== retained from home-hub's version ==="
grep -c 'recipe-to-cooklang\|skill-activation-prompt\|task-044-superpowers-integration' docs/superpowers-integration.md
echo "=== adopted from atlas's version ==="
for h in 'Task resolution' 'Artifact location override' 'Phase 4 context budget' 'Picking the roster' 'What a reviewer returns'; do
  grep -q "$h" docs/superpowers-integration.md && echo "ok   $h" || echo "MISSING $h"
done
echo "=== dropped ==="
grep -q 'Packet Work' docs/superpowers-integration.md && echo "STILL PRESENT: Packet Work" || echo "ok   Packet Work dropped"
```

Expected: a non-zero retained count; five `ok` adopted headings; `Packet Work dropped`.

- [ ] **Step 4: Produce the two diffs a human reviews for AC-17**

```bash
echo "=== vs home-hub's prior version ==="; diff /tmp/hh-si-before.md docs/superpowers-integration.md | head -60
echo "=== vs atlas's version ==="; diff /tmp/atlas-si.md docs/superpowers-integration.md | head -60
```

Both diffs must be non-empty. A near-empty diff against atlas means this was an overwrite — FR-D4 forbids that; redo Step 2.

- [ ] **Step 5: Verify no atlas leakage and that every reference resolves**

```bash
grep -nE 'atlas-ui|\bWZ\b|\bIDA\b|packet|Chronicle20|libs/' docs/superpowers-integration.md; echo "exit=$?"
grep -ohE '`/[a-z-]+`' docs/superpowers-integration.md | tr -d '`/' | sort -u | while read -r c; do
  test -f ".claude/commands/$c.md" && echo "ok   /$c" || echo "MISSING /$c"
done
```

Expected: no leakage; every command `ok`.

- [ ] **Step 6: Commit**

```bash
git add docs/superpowers-integration.md
git commit -m "docs(task-055): reconcile superpowers-integration.md with atlas's version"
```

---

## Task 16: Restructure `CLAUDE.md` (L4)

**Files:**
- Modify: `CLAUDE.md` (full rewrite, 69 lines of prose → rule list)

**Interfaces:**
- Consumes: all nine owner documents (Tasks 13–15) — every target in the trigger table must already exist.
- Produces: the eight-heading rule-list `CLAUDE.md`, ending with the trigger → owner table.

Last within L4 by design §8.4: FR-M2 and AC-14 require every table target exist, and writing the table before the nine docs land produces broken links that nothing catches until the end.

- [ ] **Step 1: Confirm all nine owner documents exist before writing the table**

```bash
cd <repo-root>
for d in agent-dispatch verification superpowers-integration review-protocol post-implementation \
         codemod-vs-agents slice-first tooling-conventions git-workflow; do
  test -f "docs/$d.md" && echo "ok   docs/$d.md" || echo "MISSING docs/$d.md"
done
```

Expected: nine `ok` lines. Any `MISSING` blocks this task — go finish Task 13, 14, or 15.

- [ ] **Step 2: Snapshot the current file and inventory the rules that must survive**

```bash
cp CLAUDE.md /tmp/hh-claude-before.md
```

Nine operative rules are currently expressed as prose. Every one has a destination (design §4.6, FR-M3):

| Current prose rule | Lands under |
|---|---|
| Verify Docker builds when shared libraries change | `## Done means verified` — as "run flagless `tools/verify.sh`", which enforces it |
| `scripts/local-up.sh` for local deployment | `## Repository conventions` |
| Plan ≠ implement; wait for approval | `## Development workflow` |
| Four-phase flow + worktree discipline | `## Development workflow` |
| Artifact-location override (`docs/tasks/task-NNN-slug/`) | `## Development workflow` |
| Code review before PR | `## Done means verified` |
| Design/plan output style (write the full doc, don't walk through it) | `## Development workflow` |
| Verification over memory | `## Evidence & grounding` |
| Shared-type refactor / no cross-layer internals | `## Repository conventions` |

The "Code Review Pattern" list becomes one `## Dispatching agents` bullet plus a table row pointing at `docs/review-protocol.md` — that is precisely the compression the owner-doc table buys.

- [ ] **Step 3: Write the new `CLAUDE.md`**

Exactly these eight `##` headings, in this order (FR-M1): `## Never do this`, `## Evidence & grounding`, `## Development workflow`, `## Done means verified`, `## Dispatching agents`, `## Handing off context`, `## Repository conventions`, `## Where the procedures live`.

Repo-specific content that stays and is **not** homogenized (FR-M4): the project overview (Go microservices; TypeScript only for the frontend), build commands, deployment specifics (`scripts/local-up.sh`), and domain conventions (JSON:API, GORM, `tenant_id` scoping, the four-phase flow's exact command names).

The file ends with, and only with, the trigger table:

```markdown
## Where the procedures live

Each row names a trigger and the one document that owns it. Read the owner
before acting — do not reconstruct the procedure from this file.

| When you are about to… | Read |
|---|---|
| Dispatch an agent, pick a model, choose fan-out vs. fork, or hand off | [`docs/agent-dispatch.md`](docs/agent-dispatch.md) |
| Run the gate, interpret a gate failure, or reconcile `verify.sh` with CI | [`docs/verification.md`](docs/verification.md) |
| Use a bare task number, or reach for a skill outside a phase command | [`docs/superpowers-integration.md`](docs/superpowers-integration.md) |
| Dispatch a reviewer, or write up a review | [`docs/review-protocol.md`](docs/review-protocol.md) |
| Fix a bug found after the PR opened (Phase 5, `/fix-pr-bug`) | [`docs/post-implementation.md`](docs/post-implementation.md) |
| Point a second implementer at the same mechanical transformation | [`docs/codemod-vs-agents.md`](docs/codemod-vs-agents.md) |
| Read a large document, diff, plan, or tool result | [`docs/slice-first.md`](docs/slice-first.md) |
| Start a long-running process, or look up a mechanical repo fact | [`docs/tooling-conventions.md`](docs/tooling-conventions.md) |
| Commit, push, rebase, or clean up a stray `main` commit | [`docs/git-workflow.md`](docs/git-workflow.md) |
```

Under `## Done means verified`, the Docker rule must appear as an enforced rule, not a reminder:

```markdown
- A branch is done when flagless `tools/verify.sh` exits 0. Nothing else is
  "done" — not `--quick`, not `--no-docker`, not "the tests passed".
- The old "verify Docker builds when shared libraries change" rule is now
  mechanical: the flagless run's docker leg fires on any `shared/` change and
  fans out to all 12 service images. You cannot forget it by forgetting it.
- Always run the code-review step before opening a PR, even when the plan
  looks complete.
```

- [ ] **Step 4: Verify the eight headings appear in order (AC-13)**

```bash
grep -n '^## ' CLAUDE.md
```

Expected, in exactly this order: `Never do this`, `Evidence & grounding`, `Development workflow`, `Done means verified`, `Dispatching agents`, `Handing off context`, `Repository conventions`, `Where the procedures live`.

```bash
diff <(grep '^## ' CLAUDE.md) <(printf '## %s\n' \
  'Never do this' 'Evidence & grounding' 'Development workflow' 'Done means verified' \
  'Dispatching agents' 'Handing off context' 'Repository conventions' 'Where the procedures live') \
  && echo "headings exact and in order"
```

Expected: `headings exact and in order`.

- [ ] **Step 5: Verify the file ends with the trigger table (FR-M2)**

```bash
tail -3 CLAUDE.md
grep -n '^## Where the procedures live' CLAUDE.md
grep -c '^## ' CLAUDE.md
```

Expected: the last non-blank line is a table row; `Where the procedures live` is the last `##`; exactly 8 headings.

- [ ] **Step 6: Verify every `docs/` link resolves (AC-14, FR-M5)**

```bash
grep -ohE '\(docs/[a-z0-9/-]+\.md\)' CLAUDE.md | tr -d '()' | sort -u | while read -r p; do
  test -f "$p" && echo "ok   $p" || echo "BROKEN $p"
done
```

Expected: nine `ok` lines, zero `BROKEN`.

- [ ] **Step 7: Verify no operative rule was lost (FR-M3)**

```bash
for r in 'local-up.sh' 'tools/verify.sh' 'code-review' 'docs/tasks/task-NNN-slug' \
         'worktree' 'verify against local source' 'write the full document'; do
  grep -qi "$r" CLAUDE.md && echo "ok   $r" || echo "MISSING $r"
done
diff <(grep -oiE 'local-up|worktree|json:api|tenant' /tmp/hh-claude-before.md | sort -u) \
     <(grep -oiE 'local-up|worktree|json:api|tenant' CLAUDE.md | sort -u)
```

Expected: seven `ok` lines; the diff empty. Any `MISSING` means a rule was dropped in the rewrite — restore it under its mapped heading.

- [ ] **Step 8: Commit**

```bash
git add CLAUDE.md
git commit -m "docs(task-055): restructure CLAUDE.md into the eight-heading rule list"
```

---

## Task 17: Go formatter sweep — 180 files, isolated (L5)

**Files:**
- Modify: ~180 `.go` files across 14 modules (measured in design §3.1)

**Interfaces:**
- Consumes: `tools/verify.sh --fix-fmt` (Task 5), `.golangci.yml` (Task 1).
- Produces: a tree where `golangci-lint fmt --diff` is empty, so the `lint` leg's formatter layer passes.

**This commit touches nothing but formatting and is generated by a single tool invocation.** Design §7 and §8.3: it must land isolated and first within L5, so every later lint fix is readable *against* it rather than buried inside it. It is reviewable by re-running the tool, not by reading 180 files.

Per module, the measured counts: recipe 33, workout 30, productivity 21, account 19, calendar 18, shopping 13, tracker 12, package 11, auth 8, dashboard 6, category 5, weather 2, `shared/go/auth` 1, `shared/go/kafka` 1. Ten of the twelve `shared/go/*` modules are already clean.

- [ ] **Step 1: Confirm the working tree is clean before the sweep**

```bash
cd <repo-root>
git status --porcelain
```

Expected: no output. A dirty tree here means unrelated changes would be swept into the formatter commit — commit or set them aside first.

- [ ] **Step 2: Record the baseline the sweep is supposed to fix**

```bash
tools/verify.sh --only lint 2>&1 | grep -c 'FMT FAIL'
```

Expected: a non-zero count of modules failing the formatter layer. Record it.

- [ ] **Step 3: Run the sweep — one invocation, nothing else**

```bash
tools/verify.sh --fix-fmt
```

- [ ] **Step 4: Verify the diff is formatting only**

```bash
git status --porcelain | wc -l
git diff --stat | tail -3
git diff --numstat | awk '$1 != $2 && ($1 > 3 || $2 > 3) {print}' | head -20
```

Expected: roughly 180 changed files, all `.go`. The third command lists files whose added and deleted line counts differ substantially — a pure reformat is usually balanced. Spot-check any it prints; a formatter should not be deleting logic.

```bash
git diff -- '*.md' '*.yml' '*.json' '*.ts' '*.tsx' | wc -l
```

Expected: `0`. The sweep must not have touched anything but Go.

- [ ] **Step 5: Verify the tree still builds and tests green**

```bash
tools/verify.sh --only build,test 2>&1 | tail -20
```

Expected: both `PASSED`. Formatting is behaviour-preserving by construction; a failure here means something other than the formatter ran.

- [ ] **Step 6: Verify the formatter layer is now clean**

```bash
tools/verify.sh --only lint 2>&1 | grep -c 'FMT FAIL'
```

Expected: `0`. The `LINT FAIL` lines remain — Tasks 18 and 19 clear those.

- [ ] **Step 7: Commit, naming the command in the message**

```bash
git add services shared
git commit -m "$(cat <<'EOF'
style(task-055): apply gofumpt and goimports across all Go modules

Generated by a single invocation, no hand edits:

    tools/verify.sh --fix-fmt

Formatting only. Review by re-running the command on the parent commit and
diffing, not by reading 180 files.
EOF
)"
```

- [ ] **Step 8: Verify the commit is formatting-only**

```bash
git show --stat HEAD | tail -3
git show --name-only HEAD | grep -v '\.go$' | grep -v '^$' | head
```

Expected: only `.go` files (plus the commit header lines). Anything else must be removed from this commit.

---

## Task 18: Clear the 78 `errcheck` findings (L5)

**Files:**
- Modify: Go files across the 24 modules, wherever `errcheck` fires

**Interfaces:**
- Consumes: the formatted tree (Task 17), `tools/verify.sh --only lint` (Task 5).
- Produces: zero `errcheck` findings.

Measured in design §3.2: 78 findings — `resp.Body.Close()`, `w.Write()`, `db.AutoMigrate()`, test-helper calls. Mechanical.

**The governing constraint (FR-L3, design §7):** default to `_ =` for genuinely-ignored returns; only surface a real error where the surrounding function **already** returns one. **Never introduce a new error path.** A fix that changes behaviour is out of scope and must be raised, not applied.

- [ ] **Step 1: Get the full, current finding list**

```bash
cd <repo-root>
tools/verify.sh --only lint 2>&1 | grep 'errcheck' | tee /tmp/errcheck.txt | wc -l
```

Expected: ~78 lines. This list, not the design's counts, is the authority — Task 17 may have shifted line numbers.

- [ ] **Step 2: Bucket them before fixing any**

```bash
awk -F'errcheck' '{print $1}' /tmp/errcheck.txt | sed 's/:[0-9]*:[0-9]*:.*//' | sort | uniq -c | sort -rn
grep -oE '\b[A-Za-z_.]+\([^)]*\)$' /tmp/errcheck.txt | sort | uniq -c | sort -rn | head -20
```

This gives per-module counts and the most common call shapes. Work module by module, not finding by finding, so each commit is coherent.

- [ ] **Step 3: Apply the three fix shapes**

**Shape A — deferred cleanup whose failure is genuinely uninteresting.** The overwhelmingly common case:

```go
// before
defer resp.Body.Close()
// after
defer func() { _ = resp.Body.Close() }()
```

**Shape B — a write whose error the caller cannot act on** (response already committed, header already sent):

```go
// before
w.Write(payload)
// after
_, _ = w.Write(payload)
```

**Shape C — a call inside a function that already returns an error.** Only here may the error be surfaced, because doing so adds no new error path:

```go
// before
func migrate(db *gorm.DB) error {
    db.AutoMigrate(&Entity{})
    return nil
}
// after
func migrate(db *gorm.DB) error {
    return db.AutoMigrate(&Entity{})
}
```

If a call site fits none of the three — in particular, if surfacing the error would require changing a function's signature — **stop and raise it**. That is a behaviour change and out of scope.

- [ ] **Step 4: After each module, verify it builds, tests, and lints clean**

```bash
MOD=services/recipe-service   # repeat per module
(cd "$MOD" && go build ./... && go test ./... -count=1) && echo "module green"
tools/verify.sh --only lint 2>&1 | grep "errcheck" | grep -c "$MOD"
```

Expected: `module green`, then `0` remaining errcheck findings for that module.

- [ ] **Step 5: Commit per module**

```bash
git add "$MOD"
git commit -m "fix(task-055): check or explicitly ignore errors in ${MOD##*/} (errcheck)"
```

- [ ] **Step 6: Verify the whole class is cleared**

```bash
tools/verify.sh --only lint 2>&1 | grep -c 'errcheck'
tools/verify.sh --only build,test 2>&1 | tail -5
```

Expected: `0` errcheck findings; `build` and `test` both `PASSED`.

- [ ] **Step 7: Confirm no new error path was introduced**

```bash
git diff main...HEAD -- 'services/**/*.go' 'shared/**/*.go' \
  | grep -E '^\+.*func .*\) error' | head -20
```

Expected: no output. A function that gained an `error` return is a behaviour change — revert it and raise the case instead.

---

## Task 19: Clear the remaining 32 Go findings (L5)

**Files:**
- Modify: Go files flagged by `staticcheck` (22), `unused` (8), `ineffassign` (1), `govet` (1)

**Interfaces:**
- Consumes: the errcheck-clean tree (Task 18).
- Produces: zero Go lint findings, so the `lint` leg passes.

Measured in design §3.2. These need judgment, not a sweep — this is the ~32 genuinely hand-written findings.

- [ ] **Step 1: Get the current list, bucketed by linter**

```bash
cd <repo-root>
tools/verify.sh --only lint 2>&1 | grep -oE '\((staticcheck|unused|ineffassign|govet)\)' | sort | uniq -c
tools/verify.sh --only lint 2>&1 | grep -E '\((staticcheck|unused|ineffassign|govet)\)' > /tmp/go-lint-rest.txt
cat /tmp/go-lint-rest.txt
```

Expected: roughly staticcheck 22, unused 8, ineffassign 1, govet 1.

- [ ] **Step 2: Fix the staticcheck findings**

The measured breakdown and the fix for each:

- **`S1016` struct-literal → conversion (10).** Replace a field-by-field literal with a type conversion where the two structs have identical underlying types:
  ```go
  // before
  return Entity{ID: m.ID, Name: m.Name, TenantID: m.TenantID}
  // after
  return Entity(m)
  ```
  Verify the underlying types really are identical before converting — if the compiler rejects the conversion, the finding was a false positive for that site and the literal stays.
- **`QF1012` `Fprintf` (3).** `fmt.Fprintf(w, "%s", s)` → `io.WriteString(w, s)`, or `w.Write([]byte(s))`, matching whatever the file already does.
- **`ST1005` error-string casing (4).** Lowercase the first word and drop trailing punctuation: `errors.New("Invalid tenant.")` → `errors.New("invalid tenant")`. **Check first that no test asserts on the exact string** — if one does, update the test in the same commit.
- **`QF1008` embedded-field selector (4).** Drop the redundant embedded type name: `m.Model.ID` → `m.ID`.
- **`QF1003`/`S1009` (2).** `if/else if` chain on one value → `switch`; redundant `nil` check before `len()` → drop the check.

- [ ] **Step 3: Fix the unused findings — but confirm they are truly dead first**

Eight dead functions in `provider.go`/`administrator.go` across dashboard and calendar, plus a `mockClient` in `rest_test.go`.

```bash
grep -n 'unused' /tmp/go-lint-rest.txt
```

For each name, before deleting:

```bash
SYM=<the function name>
git grep -n "\b$SYM\b" -- '*.go' | grep -v '_test.go:.*func ' | head
```

Expected: only the declaration. If anything else references it, it is not dead — the linter is scoped per module, so a cross-module reference is possible. Design §7 assessed the reflective-reference risk as low: GORM and JSON:API code here is not reflective over Go function names. `go build ./...` plus the existing suite is sufficient evidence.

Delete the declaration and any now-unused imports it pulled in.

- [ ] **Step 4: Fix `ineffassign` and `govet`**

- **`ineffassign` in `processor.go`** — an ineffectual assignment to `m`. Remove the dead assignment. If the assignment looks intentional (a value meant to be used and silently is not), that is a latent bug: fix it so the value is used, and say so in the commit message.
- **`govet` in `tenant_callbacks.go`** — `reflect.Ptr` should be inlined. Replace `reflect.Ptr` with `reflect.Pointer` per the current API.

- [ ] **Step 5: Verify after each linter class**

```bash
tools/verify.sh --only build,test 2>&1 | tail -8
tools/verify.sh --only lint 2>&1 | grep -cE '\((staticcheck|unused|ineffassign|govet)\)'
```

Expected: `build` and `test` `PASSED`; the finding count strictly decreasing to `0`.

- [ ] **Step 6: Commit per linter class**

```bash
git add services shared
git commit -m "fix(task-055): resolve staticcheck findings (behaviour-preserving)"
# then, separately:
git commit -m "fix(task-055): remove dead functions flagged by unused"
git commit -m "fix(task-055): fix ineffectual assignment and inline reflect.Pointer"
```

- [ ] **Step 7: Verify the Go lint leg is fully green**

```bash
tools/verify.sh --only lint 2>&1 | tail -20
```

Expected: `lint  PASSED` — no `FMT FAIL`, no `LINT FAIL`. This is the first time the Go half of the gate has ever been green in this repository.

---

## Task 20: Clear the 10 frontend eslint errors (L5)

**Files:**
- Modify: `frontend/src/lib/calendar/recurrence.ts:7`, `frontend/src/pages/__tests__/WorkoutReviewPage.test.tsx:55-56`, `frontend/src/features/dashboards/new-dashboard-modal.tsx:31,33`, `frontend/src/features/tracker/calendar-grid.tsx:250`, `frontend/src/components/ui/sidebar.tsx:43`, `frontend/src/lib/hooks/use-cooklang-preview.ts:35`, `frontend/src/pages/DashboardDesigner.tsx:56-57`

**Interfaces:**
- Consumes: `tools/verify.sh --only eslint` (Task 5).
- Produces: `npx eslint .` exiting 0 in `frontend/`, so the `eslint` leg passes and AC-8 can be met without `--no-ui`.

Per user decision 3: clear all 10, including the two `rules-of-hooks` errors. This is scope the PRD never disclosed — `scripts/lint-all.sh` has evidently never completed and `.github/workflows/pr.yml` has no lint job, so nothing has ever enforced this.

Line numbers come from the design's measurement and may have shifted; always re-run eslint for the current positions.

- [ ] **Step 1: Get the current error list**

```bash
cd <repo-root>/frontend
npx eslint . 2>&1 | tail -30
```

Expected: 10 errors, 5 warnings. Warnings do not fail the leg; errors do. Fix the errors.

- [ ] **Step 2: Record the passing test baseline before touching anything**

```bash
npm test 2>&1 | tail -5
```

Expected: 105 files / 695 tests passing. This is the regression baseline for Steps 4–6.

- [ ] **Step 3: Fix the four trivially mechanical errors**

- **`@typescript-eslint/no-unused-vars` — `src/lib/calendar/recurrence.ts:7`.** Delete the unused binding. If it is an intentionally-unused parameter, prefix it with `_`.
- **`@typescript-eslint/no-explicit-any` ×2 — `src/pages/__tests__/WorkoutReviewPage.test.tsx:55,56`.** Replace `any` with the real type. In a test this is usually the component's prop type or a `Partial<T>` of a service response — read the surrounding assertions and use what they actually rely on. `unknown` plus a narrowing cast is acceptable where the shape is genuinely opaque; `any` is not.

- [ ] **Step 4: Fix `react-refresh/only-export-components` ×2 — `src/features/dashboards/new-dashboard-modal.tsx:31,33`**

The rule fires because the module exports both a component and non-component values, which breaks fast refresh. Move the two non-component exports into a sibling module (e.g. `new-dashboard-modal.types.ts` or `.constants.ts`) and re-import them.

```bash
grep -n '^export' src/features/dashboards/new-dashboard-modal.tsx
grep -rn "from '.*new-dashboard-modal'" src/ | head
```

Update every importer the second command finds.

- [ ] **Step 5: Fix `react-hooks/set-state-in-effect` ×3**

`src/features/tracker/calendar-grid.tsx:250`, `src/components/ui/sidebar.tsx:43`, `src/lib/hooks/use-cooklang-preview.ts:35`.

Each is a `useEffect` that only calls `setState` from values already available during render. Preferred fix, in order:

1. **Derive during render.** If the state is a pure function of props/other state, delete the state and the effect and compute it inline (or in `useMemo` if it is expensive).
2. **`useSyncExternalStore`.** If the effect is subscribing to something outside React (a media query, a resize observer — `sidebar.tsx:43` is likely this shape), use `useSyncExternalStore` instead.

Only if neither applies is the effect genuinely necessary; in that case the rule is telling you about a real render loop and the fix is to restructure, not to disable.

- [ ] **Step 6: Verify tests after each file**

```bash
npm test 2>&1 | tail -5
```

Expected: still 695 passing. A drop means the derivation changed behaviour — revert and reconsider before continuing.

- [ ] **Step 7: Fix `react-hooks/rules-of-hooks` ×2 — `src/pages/DashboardDesigner.tsx:56,57`**

A conditional hook call is a latent correctness bug: on a render where the early return fires, React's hook order shifts and every subsequent hook reads the wrong slot.

```bash
sed -n '35,75p' src/pages/DashboardDesigner.tsx
grep -rn 'DashboardDesigner' src/pages/__tests__/ src/**/__tests__/ 2>/dev/null | head
```

Fix: hoist both hooks above the early return, then keep the early return for rendering only.

```tsx
// before
if (!dashboard) return <Spinner />;
const [layout, setLayout] = useState(dashboard.layout);
const widgets = useWidgets(dashboard.id);

// after
const [layout, setLayout] = useState(dashboard?.layout ?? []);
const widgets = useWidgets(dashboard?.id);
if (!dashboard) return <Spinner />;
```

`useWidgets` must tolerate an `undefined` id — check its implementation and, if it is a React Query hook, gate it with `enabled: !!dashboard?.id` rather than calling it conditionally.

This **changes behaviour** (the hooks now run on the loading render), which is why user decision 3 authorises it explicitly: it is a real bug and the correct fix. If it proves non-trivial, the authorised fallback is a scoped `// eslint-disable-next-line react-hooks/rules-of-hooks` carrying a TODO that names the follow-up — never weaken the shared config, never drop the leg.

- [ ] **Step 8: Verify the designer still works and the leg is green**

```bash
npm test 2>&1 | tail -5
npx eslint . 2>&1 | tail -10
npm run build 2>&1 | tail -5
```

Expected: 695 tests passing; eslint reporting **0 errors** (the 5 warnings may remain); a successful build.

- [ ] **Step 9: Commit**

```bash
cd <repo-root>
git add frontend/src
git commit -m "fix(task-055): clear all frontend eslint errors, including conditional hooks in DashboardDesigner"
```

- [ ] **Step 10: Verify through the gate, not just directly**

```bash
tools/verify.sh --only eslint,frontend-build 2>&1 | tail -12
```

Expected: both `PASSED`.

---

## Task 21: Full acceptance sweep and AC-20 cross-repository report

**Files:**
- Create: `docs/tasks/task-055-process-parity/acceptance.md`

**Interfaces:**
- Consumes: everything.
- Produces: recorded evidence for AC-1 through AC-20.

AC-20 is a **reporting obligation, not a check** (design §8.6): specification §7 checks 1, 4, 5, and 6 are pairwise across four repositories and cannot be evaluated from home-hub alone. Report home-hub's side and state plainly that the comparison is not evaluable here. **Asserting them as passed is a defect**, per the PRD's own wording.

- [ ] **Step 1: Run the flagless gate — the one that decides AC-8**

```bash
cd <repo-root>
tools/verify.sh 2>&1 | tee /tmp/verify-flagless.log | tail -30
echo "exit=${PIPESTATUS[0]}"
```

Expected: every leg `PASSED` and `verify.sh: PASSED — this branch may be called done.` This branch touches `shared/`? Only if a lint fix landed there — the `--facts` output from Step 3 will say. Either way the docker leg must have run or correctly reported nothing to build.

- [ ] **Step 2: Run the mechanical acceptance checks**

```bash
A=<atlas-repo-root>

echo "--- AC-1: nine hooks, eight byte-identical"
ls .claude/hooks/*.sh | wc -l
for f in wait-loop-guard.sh wait-loop-guard_test.sh block-home-paths-in-docs.sh \
         turn-budget.sh turn-budget-guard.sh fork-dispatch-guard.sh \
         commit-boundary.sh task-num-collision-detector.sh; do
  diff -q "$A/.claude/hooks/$f" ".claude/hooks/$f" >/dev/null && echo "identical: $f" || echo "DIFFERS: $f"
done

echo "--- AC-2: no atlas- in hooks"
grep -l 'atlas-' .claude/hooks/*.sh; echo "exit=$?"

echo "--- AC-3: wait-loop-guard test"
bash .claude/hooks/wait-loop-guard_test.sh | tail -1

echo "--- AC-4: task-numbers test"
bash tools/task-numbers_test.sh | tail -1

echo "--- AC-5: tools/ present and executable"
ls -l tools/verify.sh tools/task-numbers.sh tools/task-brief.sh tools/toolchain.versions

echo "--- AC-6: help exits 0, unknown flag exits 2"
tools/verify.sh --help >/dev/null; echo "help=$?"
tools/verify.sh --bogus >/dev/null 2>&1; echo "bogus=$?"

echo "--- AC-7: --quick exits 0 and disclaims"
tools/verify.sh --quick 2>&1 | tail -3

echo "--- AC-10: agent trio"
ls .claude/agents/task-implementer.md .claude/agents/task-verifier.md .claude/agents/task-reviewer.md

echo "--- AC-11: no atlas-* agent names outside process-parity.md"
git grep -lE 'atlas-(implementer|verifier|reviewer)' -- . ':!docs/tasks' \
  | grep -vxE 'docs/process-parity\.md'; echo "exit=$?"

echo "--- AC-12: settings.json"
jq . .claude/settings.json >/dev/null && echo "valid json"
jq -r '.disableBundledSkills' .claude/settings.json
jq -r '.hooks | to_entries[] | .value[] | .hooks[] | .command' .claude/settings.json \
  | sed "s|\$CLAUDE_PROJECT_DIR|$PWD|" \
  | while read -r p; do [ -x "$p" ] && echo "ok $p" || echo "MISSING $p"; done

echo "--- AC-13: eight CLAUDE.md headings in order"
grep -n '^## ' CLAUDE.md

echo "--- AC-14: every docs/ link in CLAUDE.md resolves"
grep -ohE '\(docs/[a-z0-9/-]+\.md\)' CLAUDE.md | tr -d '()' | sort -u \
  | while read -r p; do [ -f "$p" ] && echo "ok $p" || echo "BROKEN $p"; done

echo "--- AC-15: nine owner docs"
for d in agent-dispatch verification superpowers-integration review-protocol post-implementation \
         codemod-vs-agents slice-first tooling-conventions git-workflow; do
  [ -f "docs/$d.md" ] && echo "ok docs/$d.md" || echo "MISSING docs/$d.md"
done

echo "--- AC-16: no atlas illustrations in owner docs"
grep -nE 'atlas-ui|\bWZ\b|\bIDA\b|packet' \
  docs/agent-dispatch.md docs/verification.md docs/superpowers-integration.md \
  docs/review-protocol.md docs/post-implementation.md docs/codemod-vs-agents.md \
  docs/slice-first.md docs/tooling-conventions.md docs/git-workflow.md; echo "exit=$?"

echo "--- AC-18: fix-pr-bug command"
ls .claude/commands/fix-pr-bug.md

echo "--- AC-19: .golangci.yml present, lint-all.sh no longer calls golangci-lint"
ls .golangci.yml
grep -c golangci-lint scripts/lint-all.sh
```

Expected: `AC-1` nine hooks and eight `identical:`; `AC-2`/`AC-11`/`AC-16` all `exit=1` with no output; `AC-3`/`AC-4` `failed: 0`; `AC-6` `help=0` and `bogus=2`; `AC-12` `valid json`, `true`, nine `ok`; `AC-13` the eight headings in order; `AC-14`/`AC-15` all `ok`; `AC-19` `0`.

- [ ] **Step 3: Demonstrate AC-9 both ways and capture the evidence**

AC-9 requires demonstration, not assertion.

```bash
echo "=== negative case: base with no shared/ change ==="
tools/verify.sh --facts --base HEAD | grep -E '^(base|changed-shared|fan-out|docker-images):'

echo "=== positive case: base spanning this branch's shared/ changes ==="
tools/verify.sh --facts --base "$(git merge-base HEAD main)" | grep -E '^(base|changed-shared|fan-out|docker-images):'
```

Expected: the negative case shows `changed-shared: no` and an empty or short image list; the positive case, if this branch touched `shared/`, shows `changed-shared: yes` and all 12 service images. If this branch happens not to have touched `shared/`, reproduce the probe from Task 6 Step 4 and record that output instead.

- [ ] **Step 4: Write `acceptance.md`**

Create `docs/tasks/task-055-process-parity/acceptance.md` with one row per AC-1…AC-20: the criterion, the exact command run, its output (trimmed), and PASS / FAIL / NOT-EVALUABLE.

AC-17 is marked **"reviewed by a human"** and carries the two diffs from Task 15 Step 4 as its evidence — it is not machine-decidable.

AC-20 gets this section verbatim in substance:

```markdown
## AC-20 — cross-repository checks (NOT EVALUABLE from home-hub)

Specification §7 checks 1, 4, 5, and 6 are pairwise comparisons across
`atlas`, `home-hub`, `Harbormaster`, and `MyFleet`. Only two of those four
repositories are present on this machine, and this task's scope is home-hub
alone. **home-hub's side of each is reported below. The pairwise comparison is
explicitly declared not evaluable here. Asserting these as passed would be a
defect.**

| Spec §7 check | home-hub's side | Pairwise verdict |
|---|---|---|
| 1 — the §3.1 hook files are byte-identical across all four repositories | 8/8 `diff`-identical to atlas @ `e83f59e61` (AC-1 output above) | NOT EVALUABLE — Harbormaster and MyFleet not inspected |
| 4 — each `.claude/settings.json` wires the same hook set at the same events | 9 hooks wired at `SessionStart`, `UserPromptSubmit`, `PreToolUse` (`Write\|Edit`, `Agent`, `Bash`, `*`), `PostToolUse` (`Write\|Edit`, `Bash`, `*`); every path resolves to an existing executable (AC-12 output above) | NOT EVALUABLE — pairwise across four repos |
| 5 — each `CLAUDE.md` carries the same eight headings and ends with a `## Where the procedures live` table whose every target exists | 8 headings present in spec order; file ends with the table; 9/9 targets exist (AC-13, AC-14 output above) | NOT EVALUABLE — pairwise across four repos |
| 6 — each repository has the nine owner documents, and every `docs/` link in every `CLAUDE.md` resolves | 9/9 owner documents present; every `docs/` link in `CLAUDE.md` resolves (AC-15, AC-14 output above) | NOT EVALUABLE — pairwise across four repos |
```

Paste the real command outputs into the "home-hub's side" cells rather than restating them from this plan. Note that spec §7 check 3 **is** evaluable here and is AC-11 — home-hub's carve-out is narrower than atlas's: only `docs/process-parity.md` is exempt, because home-hub never used the `atlas-*` names and so needs no historical-cutoff note in `docs/agent-dispatch.md`.

- [ ] **Step 5: Verify `acceptance.md` has no unfilled rows**

```bash
grep -nE '\[record|TBD|TODO|\.\.\.' docs/tasks/task-055-process-parity/acceptance.md; echo "exit=$?"
grep -c '^| AC-' docs/tasks/task-055-process-parity/acceptance.md
```

Expected: `exit=1` (no placeholders left); 20 AC rows.

- [ ] **Step 6: Final flagless run on the finished tree**

```bash
tools/verify.sh 2>&1 | tail -15; echo "exit=${PIPESTATUS[0]}"
```

Expected: every leg `PASSED`, `exit=0`, and the message `this branch may be called done.` **This is AC-8, and it is the single check the whole task exists to make true.**

- [ ] **Step 7: Commit**

```bash
git add docs/tasks/task-055-process-parity/acceptance.md
git commit -m "docs(task-055): record acceptance evidence for AC-1 through AC-20"
```

- [ ] **Step 8: Run the code review before opening a PR**

Per `CLAUDE.md`'s code-review-before-PR rule, invoke `superpowers:requesting-code-review`. It dispatches `plan-adherence-reviewer`, `backend-guidelines-reviewer` (Go files changed), and `frontend-guidelines-reviewer` (TS files changed) in parallel, each writing to `docs/tasks/task-055-process-parity/audit.md`.

---

## Self-Review

**Spec coverage.** Every PRD section maps to a task: §4.1 hooks → Task 8; §4.2 `format-on-write.sh` → Task 9; §4.3 agents → Task 10; §4.4 tools → Tasks 1, 2, 3–6; §4.5 `verify.sh` → Tasks 3–6; §4.6 linter + backlog → Tasks 1, 17–19; §4.7 owner docs → Tasks 13–15; §4.8 commands → Tasks 11–12; §4.9 settings → Task 9; §4.10 `CLAUDE.md` → Task 16; AC-1…AC-20 → Task 21. Design §§2–8 map to the L0–L5 spine and the six deviations, all of which are carried into the tasks that implement them.

**Known deliberate departures from the design, all recorded above:** `--facts` added (forced by user decision 1's full `execute-task.md` port); leg 3 is plain `go vet` rather than `golangci-lint govet` (so `--quick` needs no bootstrap); `--fix-fmt` added (a failure message must name a command that exists); user decisions 1–3 override design §6.4 alternatives, §6.2's literal-FR-V9 alternative, and §6.5's defer option respectively.

**Ordering constraints from design §8, all honoured:** `.gitignore` `.cache/` lands in Task 1 before any bootstrap (Task 5). `verify.sh` (Tasks 3–6) precedes all L5 remediation (Tasks 17–20). The formatter sweep is isolated and first within L5 (Task 17). `CLAUDE.md` is last within L4 (Task 16, after Tasks 13–15). `settings.json` is wired only after all nine hooks exist (Task 9, after Task 8). AC-20 is treated as a reporting obligation (Task 21).

**Leg names are consistent throughout:** `pins build vet test lint frontend-build eslint docker`, defined in Task 3 and used unchanged in Tasks 4–7, 13, 17–21. Function names `discover_modules`, `resolve_base`, `run_leg`, `majmin`, `for_each_module`, `ensure_golangci`, `docker_targets`, `print_facts`, `print_summary`, and the `leg_*` family are defined once and referenced consistently. `$ATLAS` and `$ROOT` are fixed in Global Constraints and used verbatim in every task.
