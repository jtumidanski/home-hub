# Process Parity — Execution Context

Companion to `plan.md`. Everything an implementer needs that the plan's task
bodies assume rather than restate.

---

## 1. Where you are

| | |
|---|---|
| Worktree (`$ROOT`) | `<repo-root>` |
| Branch | `task-055-process-parity` |
| Base branch | `main` |
| Task folder | `docs/tasks/task-055-process-parity/` |
| Atlas source (`$ATLAS`) | `<atlas-repo-root>` |

Every command runs from `$ROOT`. Never `cd` to the main repo at
`<main-repo-root>`. `$ATLAS` is read-only for this task — never
write into it.

The canonical cross-repository specification is committed here at
`docs/process-parity.md` (verbatim copy of atlas's, pinned at atlas commit
`e83f59e61`); home-hub scoping notes are at `process-parity-brief.md` (this
task folder). The design phase confirmed by `diff` that the pinned copy is
current.

---

## 2. Repository shape (verified, not remembered)

**24 Go modules under `go.work`** — 12 services, 12 shared:

```
services/  account auth calendar category dashboard package
           productivity recipe shopping tracker weather workout   (all -service)
shared/go/ auth dashboard database events http kafka logging model
           retention server tenant testing
```

**13 Docker images** — the 12 services (context `.`, dockerfile
`services/<svc>/Dockerfile`) plus `frontend` (context `frontend`, dockerfile
`frontend/Dockerfile`). This mirrors `.github/workflows/pr.yml`'s matrix
exactly.

**Frontend:** `frontend/` — React + TypeScript + Vite, eslint 10, **no
prettier**, vitest (105 files / 695 tests, all green). Build is
`tsc -b && vite build`.

**Existing `.claude/` inventory** (do not re-create any of it):

- Agents: `backend-guidelines-reviewer`, `frontend-guidelines-reviewer`,
  `plan-adherence-reviewer`, `todo-scanner`, `service-documentation`.
- Commands: `audit-plan`, `backend-audit`, `design-task`, `execute-task`,
  `plan-task`, `recipe-to-cooklang`, `review-todos`, `service-doc`, `spec-task`.
- Hooks: `skill-activation-prompt.sh` + `.py` only.
- Skills: `backend-dev-guidelines`, `frontend-dev-guidelines` (+
  `skill-rules.json`).

**`tools/` does not exist.** This task creates it.

**`scripts/`** has 13 scripts. Only `ci-build.sh`, `ci-test.sh`, and
`lint-all.sh` are touched (they become delegates). The `build-*.sh`,
`local-up.sh`, `local-down.sh`, and `test-all.sh` scripts are untouched.

---

## 3. Measured baseline (from the design phase — re-measure, don't trust)

These numbers were measured with `golangci-lint v2.13.1` and atlas's linter set.
They are the plan's sizing input, not its acceptance criteria — always re-run the
gate for current counts, because Task 17's formatter sweep shifts line numbers.

| Layer | Count | Character |
|---|---|---|
| Go formatter (`gofumpt`/`goimports`) | 180 files / 14 modules | One `tools/verify.sh --fix-fmt` invocation |
| `errcheck` | 78 | `resp.Body.Close()`, `w.Write()`, `db.AutoMigrate()` |
| `staticcheck` | 22 | S1016 ×10, ST1005 ×4, QF1008 ×4, QF1012 ×3, QF1003/S1009 ×2 |
| `unused` | 8 | Dead funcs in `provider.go`/`administrator.go`; a `mockClient` |
| `ineffassign` | 1 | `processor.go` |
| `govet` | 1 | `tenant_callbacks.go` — `reflect.Ptr` → `reflect.Pointer` |
| frontend eslint | 10 errors, 5 warnings | See below |

**Already green:** all 24 modules build and test clean; `npm run build` and
`npm test` pass; `tools/task-numbers.sh` copied unmodified returns `056`;
`wait-loop-guard_test.sh` passes here (33/33).

**Frontend eslint errors** (warnings do not fail the leg):

| Rule | × | Where |
|---|---|---|
| `react-hooks/set-state-in-effect` | 3 | `features/tracker/calendar-grid.tsx:250`, `components/ui/sidebar.tsx:43`, `lib/hooks/use-cooklang-preview.ts:35` |
| `react-hooks/rules-of-hooks` | 2 | `pages/DashboardDesigner.tsx:56,57` |
| `react-refresh/only-export-components` | 2 | `features/dashboards/new-dashboard-modal.tsx:31,33` |
| `@typescript-eslint/no-explicit-any` | 2 | `pages/__tests__/WorkoutReviewPage.test.tsx:55,56` |
| `@typescript-eslint/no-unused-vars` | 1 | `lib/calendar/recurrence.ts:7` |

---

## 4. Decisions already made — do not re-litigate

### Settled in the design phase

| Question | Decision | Where |
|---|---|---|
| Split the lint backlog into task-056? | **No.** 180 are one tool invocation, 78 are `errcheck` one-liners; only ~32 Go + ~7 frontend need judgment. Splitting would ship `verify.sh` with its lint leg behind a flag — a flagless run that does not mean "done", breaking the one contract the exercise exists to establish. | design §3.5 |
| `.golangci.yml` content | Adopt atlas's wholesale: `version: "2"`, `linters.default: standard`, formatters `gofumpt` + `goimports` with `local-prefixes: [github.com/jtumidanski/home-hub]`. Drop atlas's single `exclusions.rules` (names an atlas test file). | design §5.2 |
| `--new-from-rev` gating | **No.** atlas rev-gates because its backlog is too large to clear; home-hub's is 110 and can be zero. A clean tree with no rev-gate is strictly stronger and makes the flagless contract literally true. Document the divergence in `docs/verification.md`. | design §5.2, §6.6 |
| Future of `scripts/*.sh` | **Thin delegates.** `verify.sh` gains `--only <legs>`; the three scripts `exec` into it. Their granularity does not match the flag set, so composing *from* them cannot express `--quick`. Inverting leaves one module list in the repo. | design §5.3 |
| Toolchain pins | `GO_VERSION=1.27.0`, `ALPINE_VERSION=3.24`, `GOLANGCI_LINT_VERSION=v2.13.1`. **Do not port `tools/toolchain-pin-guard.sh`** — atlas's reconciles `go.mod`/`go.work`/Dockerfile ARGs/`docker-bake.hcl` across 80 modules; home-hub has neither bake nor that surface. A ~20-line `pins` leg replaces it. Bump `go.work` `1.26.1` → `1.27.0`. | design §5.4 |
| `--no-ui` flag | **Yes.** The frontend legs cost ~1 min even warm; backend-only iteration should not pay it. | design §5.5 |
| Failure handling | **Run every leg**, print a `PASSED/FAILED/SKIPPED` summary, reproduce the first failing leg's output verbatim. FR-V11's literal "stop at first failure" turns a 5–15 min gate into serial round-trips. | design §6.3 |
| Module coverage | **All 24, by discovery.** FR-V5/V6 pin coverage to `scripts/ci-build.sh`'s 9+8, which is stale — `go.work` and CI cover 12+12. Encoding the gap means a flagless exit 0 that never compiled `dashboard-service`, `tracker-service`, or `workout-service`. | design §6.1 |

### Settled by the user during planning (override the design where they differ)

1. **Full `execute-task.md` port**, not a minimal graft. `spec-task.md` Step 2
   rewired to `tools/task-numbers.sh next`. FR-C2 reads as "add no *new*
   command beyond `/fix-pr-bug`". Home-hub's worktree-discipline language must
   survive the replacement, folded into the ported Steps 1–3.
2. **Docker leg mirrors CI's matrix**, not FR-V9's shared-only rule. `shared/**`
   → all 12 service images; a service-only change → that image; `frontend/**` →
   the frontend image.
3. **Clear all 10 frontend eslint errors**, including the two `rules-of-hooks`
   in `DashboardDesigner.tsx`. AC-8 must pass without `--no-ui`.

### Forced by decision 1, recorded in the plan

- `verify.sh` gains **`--facts`**, because the ported `execute-task.md` Step 4c
  instructs the operator to run `tools/verify.sh --facts --quick --base <rev>`
  when a gate behaves unexpectedly.
- Leg 3 is plain **`go vet`**, not `golangci-lint govet`, so `--quick` never
  triggers a cold multi-minute linter bootstrap. The `lint` leg's `standard` set
  includes `govet`, so flagless coverage is unchanged.
- `verify.sh` gains **`--fix-fmt`**, because the `FMT FAIL` message must name a
  command that exists.

---

## 5. `tools/verify.sh` at a glance

Leg names, fixed across the whole plan:

```
pins  build  vet  test  lint  frontend-build  eslint  docker
```

| Flag | Effect | Counts as done? |
|---|---|---|
| *(none)* | all 8 legs | **Yes** — this is the contract |
| `--quick` | `pins build vet`; implies `--no-docker` | No |
| `--no-docker` | drops `docker` | No |
| `--no-ui` | drops `frontend-build eslint` | No |
| `--only <legs>` | exactly what is named | No |
| `--base <rev>` | change-detection base for `docker` | n/a |
| `--facts` | print base, changed services, fan-out reason, image list, legs, module count; exit 0 without building | n/a |
| `--fix-fmt` | apply formatting in place across all modules, exit 0 | n/a |
| `-h`/`--help` | usage, exit 0 | n/a |
| unknown option / unknown leg / missing value | exit **2** | n/a |

Exit codes: `0` all selected legs passed, `1` at least one failed, `2` usage
error.

Two atlas settings carried verbatim, both hard-won and both applicable here:

- `GOLANGCI_LINT_CACHE=$ROOT/.cache/golangci-lint` — the default is shared by
  every worktree and golangci-lint replays cached issues **by package path**,
  surfacing stale findings from a sibling worktree whose files no longer exist.
- `--allow-parallel-runners` on `run` — golangci-lint takes an exclusive flock
  on a machine-global `$TMPDIR/golangci-lint.lock`; two worktrees running the
  gate concurrently make the loser exit 3 with "parallel golangci-lint is
  running" and **no findings** — a spurious failure, not a lint result.

---

## 6. Hard constraints an implementer can violate by accident

- **The eight verbatim hooks must stay byte-identical to atlas.** AC-1 `diff`s
  them. A well-meaning edit — even a comment — breaks NFR-4, which exists so a
  future re-harmonization is a file copy, not a merge. `format-on-write.sh` is
  the only hook that may differ.
- **Guard hooks are fail-closed; `format-on-write.sh` is fail-open in every
  branch.** Missing toolchain, missing cached binary, unparseable input,
  non-absolute path, tool error — all exit 0 silently. A local convenience hook
  must never block an edit.
- **Hooks never bootstrap and never reach the network.** `verify.sh` owns the
  bootstrap (FR-F3, NFR-5) so a first Write never eats a multi-minute download.
- **No hook or tool may read `.env`** (NFR-3). This is why the docker leg uses
  `docker build` per service rather than `docker compose` in `deploy/compose`,
  which requires `--env-file .env`.
- **`.gitignore` `.cache/` before anything bootstraps** — otherwise a ~60 MB
  binary is staged by the first `git add`.
- **Lint fixes are behaviour-preserving.** Default to `_ =` for genuinely
  ignored returns; surface an error only where the enclosing function *already*
  returns one. **Never introduce a new error path.** The single authorised
  behaviour change in this task is the `DashboardDesigner.tsx` conditional-hook
  fix (user decision 3).
- **`docs/superpowers-integration.md` must be reconciled, never overwritten**
  (FR-D4). AC-17 is human-reviewed: the diff against *both* prior versions must
  show retained content from each.
- **No rule may be dropped because its example does not transfer** (FR-D3).
  Where there is no home-hub analogue, write a neutral one.
- **`docs/process-parity.md` is the only file exempt** from the
  `atlas-(implementer|verifier|reviewer)` name check (AC-11).
- **AC-20 is a reporting obligation, not a check.** Spec §7 checks 1, 4, 5, 6
  are pairwise across four repositories. Report home-hub's side; declare the
  comparison not evaluable here. Asserting them passed is a defect.

---

## 7. Genericization mapping (applies to every ported file)

| atlas illustration | home-hub replacement |
|---|---|
| packet encode/decode, opcode tables | JSON:API resource and attribute shapes |
| WZ data as ground truth | service source + GORM migrations as ground truth |
| IDA / client binary reading | *no analogue* → neutral: "the authoritative artifact" |
| `libs/atlas-constants` lookup | `shared/go/model`, `shared/go/tenant` lookup |
| `libs/` (module root) | `shared/go/` |
| `services/atlas-ui` + prettier | `frontend/` + eslint |
| `verify.sh --all` / docker bake | home-hub's flags (§5 above) |
| `tools/lint.sh` | `tools/verify.sh --only lint` |
| cross-service event fan-out | recipe → shopping-list ingredient flow; meal-plan → tracker |
| `atlas-implementer` / `-verifier` / `-reviewer` | `task-implementer` / `task-verifier` / `task-reviewer` |

Two additional rules for `docs/agent-dispatch.md`: name the generic `task-*`
agents, and **delete atlas's historical-cutoff note about the rename** —
home-hub never used the `atlas-*` names, so the note is false here (FR-D6).

Do **not** port `docs/packets/`, `docs/reverse-engineering.md`,
`docs/adding-a-new-service.md`, or `docs/observability.md` (FR-D7). Any
reference to them in a ported file must be removed or repointed at one of the
nine owner documents.

---

## 8. Dependency order — why the tasks are in this sequence

```
  T1   .gitignore .cache/, toolchain.versions, .golangci.yml
   │   (must precede any bootstrap, or a 60 MB binary gets staged)
   ├── T2   task-numbers.sh(+test), task-brief.sh
   │        (hooks in T8 reference both — no dangling references)
   └── T3   verify.sh skeleton: flags, leg selection, summary, pins leg
        │   (+ go.work 1.26.1 → 1.27.0, surfaced by the pins leg)
        ├── T4  build / vet / test legs
        ├── T5  golangci bootstrap + lint leg + frontend legs
        │       (establishes the cache path format-on-write.sh depends on)
        ├── T6  docker leg + --facts
        └── T7  scripts/*.sh → delegates
             │
   T8   eight verbatim hooks ──┐
   T9   format-on-write.sh + settings.json wiring
        (settings LAST: AC-12 checks every wired path is an existing executable)
             │
   T10  agent trio ──┐
   T11  /fix-pr-bug  │
   T12  execute-task.md port + spec-task.md rewiring
             │
   T13  verification / tooling-conventions / git-workflow / slice-first
   T14  agent-dispatch / review-protocol / post-implementation / codemod-vs-agents
   T15  superpowers-integration.md reconciliation
   T16  CLAUDE.md restructure
        (LAST in L4: AC-14 requires every table target already exist)
             │
   T17  formatter sweep — 180 files, ISOLATED commit, first in L5
   T18  errcheck ×78
   T19  staticcheck/unused/ineffassign/govet ×32
   T20  frontend eslint ×10
   T21  acceptance sweep + AC-20 report + code review
```

`verify.sh` must exist before any L5 remediation: remediation is defined by what
the gate reports, and measuring it any other way invites a second, different
number.

---

## 9. Fast commands

```bash
cd <repo-root>

tools/verify.sh --quick                 # inner loop: pins, build, vet
tools/verify.sh --only lint             # current Go lint backlog
tools/verify.sh --only eslint           # current frontend backlog
tools/verify.sh --facts                 # what would the gate do, and why
tools/verify.sh                         # THE gate — exit 0 means done

bash tools/task-numbers_test.sh         # AC-4
bash .claude/hooks/wait-loop-guard_test.sh   # AC-3
bash tools/verify_test.sh               # verify.sh's own contract tests
jq . .claude/settings.json              # AC-12
```

---

## 10. Risks carried into execution

| Risk | Mitigation |
|---|---|
| The 180-file formatter diff is too large to review | Land as **one isolated commit** from a single tool invocation, with the command in the message. Reviewed by re-running the tool, not by reading. |
| An `errcheck` fix changes behaviour | Default `_ =`; surface only where the function already returns an error. Never add an `error` return — `git diff main...HEAD \| grep '^+.*func .*) error'` must be empty. |
| Deleting an `unused` function removes something reflectively referenced | `git grep` the symbol before deleting. GORM/JSON:API code here is not reflective over Go func names; `go build ./...` + the existing suite is sufficient evidence. |
| The `DashboardDesigner.tsx` hook fix regresses the designer | 695 vitest tests are the baseline; check whether the designer is covered before editing. Authorised fallback: scoped `eslint-disable-next-line` + TODO. Never weaken the shared config or drop the leg. |
| The docker leg makes the flagless run unusably slow | Change-gated by default; `--no-docker`/`--quick` for the inner loop. A non-`shared/` change builds only affected images (NFR-2). |
| The eight verbatim hooks drift from atlas via a well-meaning edit | AC-1's `diff` is the guard; `docs/verification.md` must state these are upstream-fixed copies. |
| `.cache/` gets committed | Ignored in T1, before the linter is ever bootstrapped. Verified by `git status --porcelain .cache` printing nothing. |
