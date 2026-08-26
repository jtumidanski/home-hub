# Verification

`tools/verify.sh` is the pre-PR gate. It must exit 0 before a branch is called
"done," "ready for PR," or handed to `superpowers:finishing-a-development-branch`.

```sh
tools/verify.sh                        # full gate — what you run before opening a PR
tools/verify.sh --quick                # inner loop: pins/build/vet, no docker, no frontend
tools/verify.sh --no-docker            # everything except the docker leg
tools/verify.sh --no-ui                # everything except frontend-build/eslint
tools/verify.sh --only lint,eslint     # exactly these legs, ignoring the presets

tools/verify.sh --quick --base <rev>   # iteration gate: only the increment
```

Only the **flagless** invocation counts as verified. `--quick`, `--no-docker`,
`--no-ui`, and `--only` also exit 0 on success — the script says so on every
run — but none of them means "done." Never claim verified from a subset.

The script mirrors the jobs in `.github/workflows/pr.yml`. **CI is the
authority**: if the two ever disagree, that is a `tools/verify.sh` coverage
defect to fix, not a result to argue with.

This document explains *why* each check exists and what breaks when it is
skipped. The list of checks lives in the script, not here — do not maintain a
second copy of it in `CLAUDE.md` or anywhere else.

---

## The eight legs

In order: `pins`, `build`, `vet`, `test`, `lint`, `frontend-build`, `eslint`,
`docker`. A flagless run selects all eight.

| Leg | What it does | Skipped by |
|---|---|---|
| `pins` | Compares `GO_VERSION` in `tools/toolchain.versions` against `go.work`'s `go` directive and every `go-version:` in `.github/workflows/*.yml` | nothing (always runs when selected) |
| `build` | `go build ./...` in every discovered module | nothing |
| `vet` | `go vet ./...` in every discovered module | nothing |
| `test` | `go test ./... -count=1` in every discovered module | `--quick` |
| `lint` | gofumpt/goimports diff-check plus golangci-lint's `standard` set, per module | `--quick` |
| `frontend-build` | `npm ci && npm run build && npm test` in `frontend/` | `--quick`, `--no-ui` |
| `eslint` | `npx eslint .` in `frontend/` | `--quick`, `--no-ui` |
| `docker` | `docker build` for the images the diff selects | `--quick`, `--no-docker` |

`--quick` runs only `pins build vet` and implies `--no-docker`. `--no-docker`
drops `docker` alone. `--no-ui` drops `frontend-build` and `eslint`. `--only`
ignores the presets entirely and runs exactly the comma-separated legs named.

## Asking the gate what it selected — `--facts`

Before investigating why a run was broad, slow, or skipped something, ask it:

```sh
tools/verify.sh --facts --quick --base <sha>
```

It prints the resolved change base, which services changed, whether
`shared/` changed, the fan-out reason, which docker images would build, which
legs are selected, and the discovered module count — then exits 0 without
building anything. Pass the same flags you would really run; the answer
reflects them. Never reverse-engineer the selection by reading or diffing
`tools/verify.sh`'s source — ask `--facts` instead.

## Iteration gate

The default change base is the merge base with `origin/main` — the **whole
branch**. That is right for the pre-PR gate and wrong for a per-task gate run
repeatedly on the same branch, because a single `shared/` commit fans every
later docker-leg run out to all 12 service images. For a gate you run per
task, scope it to the increment:

```sh
tools/verify.sh --quick --base <last-commit-you-already-gated>
```

Launch the gate in the background and keep working; never idle waiting on it.

The narrowing is safe only because every commit in the range gets gated by
*some* run — the increment's base must be the last commit that actually
passed, not blindly `HEAD~1`. The flagless pre-PR run always uses the merge
base and covers the branch as a whole regardless.

Note on the fan-out warning: the shared-lib fan-out reason is surfaced by the
**docker leg only** (`docker_targets` in `tools/verify.sh`), and only in
`--facts` output or the `docker` leg's own log line. `--quick` never runs the
docker leg, so a `tools/verify.sh --quick --base <rev>` run — the recommended
per-task invocation above — will never print that warning even when
`shared/` changed. If you need to know whether the current diff would fan
out, ask `tools/verify.sh --facts` (without `--quick`) or `--facts --only
docker`.

## The Go layer

Modules are **discovered**, never listed: `discover_modules()` walks
`services/` and `shared/go/` for `go.mod` files with `find`. This is
deliberate — `scripts/ci-build.sh` used to hardcode a module list and drifted
three services behind `go.work` before this task rewrote it as a thin
delegate to `tools/verify.sh`. A hardcoded list silently stops covering new
modules; discovery cannot.

`go vet` runs full-module here on purpose, separately from golangci-lint's
`govet` in the `lint` leg — this repo carries no `--new-from-rev` gating (see
below), so the two are redundant by design rather than a gap.

## The docker layer

The `docker` leg builds `docker build -f <dockerfile> -t home-hub-<name>:verify
<context>` for the images the diff selects, mirroring
`.github/workflows/pr.yml`'s docker matrix: 12 service images plus the
frontend image, driven by the `DOCKER_IMAGES` table in `tools/verify.sh`. A
service added to that CI workflow must also be added to `DOCKER_IMAGES` in
`tools/verify.sh`, or the gate silently stops covering it.

Selection logic (`docker_targets` in `tools/verify.sh`):

- A change under `shared/` fans out to all 12 service images (plus the
  frontend image too, if `frontend/` also changed).
- Otherwise, each service image builds only when its own `services/<name>/`
  changed, and the frontend image builds only when `frontend/` changed.
- No resolvable merge base builds everything, never fewer.

For large refactors expect several fix-and-rebuild cycles. Do not shortcut
the docker leg.

## No `--new-from-rev` (deliberate divergence)

home-hub's `.golangci.yml` carries no `--new-from-rev` baseline gating. That
is a deliberate divergence from repos that gate lint findings to new code
because their backlog is too large to clear all at once: home-hub's backlog
was cleared in task-055, so the tree is clean from day one and the flagless
`tools/verify.sh` contract is literally true rather than true-modulo-a-
baseline. Do not add rev-gating without first re-reading this section and
`.golangci.yml`'s own header comment.

## `.claude/hooks/` are copies, not local originals

Several files under `.claude/hooks/` are ported verbatim from a sibling repo
and kept byte-identical on purpose (see `docs/process-parity.md`). A
well-meaning local edit to one of them breaks the parity `diff` check that
guards them. Re-harmonization is a file copy from the upstream source, not a
manual merge — fix the drift upstream, then re-copy.

## `scripts/*.sh` are thin delegates

`scripts/ci-build.sh`, `scripts/ci-test.sh`, and `scripts/lint-all.sh` have no
logic of their own — each is `exec tools/verify.sh --only <legs> "$@"`. They
exist so existing muscle memory and any external caller keep working. Do not
add logic to them; add it to `tools/verify.sh` and its `--only` leg set
instead, or the two will drift the way `scripts/ci-build.sh`'s hardcoded
module list once did.

## When `verify.sh` and CI disagree

**CI is the authority.** A disagreement is a `tools/verify.sh` coverage
defect to fix, not a result to argue with — if CI fails on something the
local gate passed, or vice versa, treat the gate as under-covering (or
over-covering) reality and correct `tools/verify.sh`, not your understanding
of CI.

---

## Lint & format

`tools/verify.sh`'s `lint` leg runs two layers per module: gofumpt +
goimports formatting, checked with `--diff` so the gate never rewrites the
tree behind you, and golangci-lint's `standard` set (errcheck, govet,
ineffassign, staticcheck, unused). Plus Prettier/ESLint for `frontend/` via
the separate `eslint` leg.

Fix mode is `tools/verify.sh --fix-fmt` — it rewrites files in place across
all modules, then exits without running any other leg. Run it before
committing when the `lint` leg reports an `FMT FAIL`.

Known footguns:

- Cross-worktree golangci-lint lock contention: two worktrees linting at once
  can contend on the machine-global `golangci-lint.lock`. `tools/verify.sh`
  passes `--allow-parallel-runners` and keys the lint cache to the repo root
  to reduce this, but serialize concurrent lint runs across worktrees when
  practical.
- The pinned golangci-lint binary is fetched (or built from source as a
  fallback) into `.cache/tools/bin/golangci-lint-<version>` on first use —
  the first `lint` run after a fresh clone or a version bump takes longer
  while it downloads.
