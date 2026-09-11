# Process Parity — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-08-26
---

## 1. Overview

Four repositories — `atlas`, `home-hub`, `Harbormaster`, and `MyFleet` — share the
same agentic development *workflow*: the four phase commands, the worktree
convention, the artifact-location override, the three reviewer agents, and the two
guideline skills. What they do not share is the *context-discipline* layer that
`atlas` grew afterwards: the enforcement hooks, the owner-document set, the
budget-capped implementer, the isolated verifier, and the terse rule-list
`CLAUDE.md` shape that an owner-doc table makes possible.

`home-hub` sits at the previous generation. Its `CLAUDE.md` is 69 lines of prose
narrative, its only hook is `skill-activation-prompt`, it has no `tools/`
directory at all, and it has no implementer/verifier/reviewer trio — so
`/execute-task` falls back to uncapped generic dispatch. There is also no single
verification entrypoint: `scripts/ci-build.sh`, `scripts/ci-test.sh`, and
`scripts/lint-all.sh` exist as three separate scripts with no composed gate, and
the `CLAUDE.md` rule about verifying Docker builds when shared libraries change
lives only in prose, where it is easy to skip.

This task executes `docs/process-parity.md` §6 step 2 for `home-hub`: full parity,
not a subset. The canonical specification is `docs/process-parity.md` (a verbatim
copy of the atlas original, pinned at atlas commit `e83f59e61`); the home-hub
scoping notes are in `process-parity-brief.md` (this task folder). Both are
committed alongside this PRD so the task folder is self-contained. There is no
sync mechanism between
repositories by design — each ends up self-contained, and consistency is asserted
mechanically at the end (§7 of the specification) rather than maintained
continuously.

## 2. Goals

Primary goals:

- Port the seven byte-identical enforcement hooks plus `commit-boundary.sh` from
  atlas into `.claude/hooks/`, and wire them in `.claude/settings.json` at the
  same events atlas uses.
- Adapt `format-on-write.sh` to home-hub's toolchain, resolving the open Go
  formatter question rather than deferring it or shipping a dangling reference.
- Create the `task-implementer` / `task-verifier` / `task-reviewer` agent trio so
  `/execute-task` dispatches capped, isolated, reviewable units of work.
- Create `tools/` — which does not exist in this repository today — containing
  `task-numbers.sh` (+ its test), `task-brief.sh`, and a new `tools/verify.sh`.
- Make `tools/verify.sh` the single verification entrypoint, composed from the
  existing `scripts/` and honouring the shared-library Docker rule.
- Port the nine owner documents into `docs/`, genericized so that no rule is lost
  and no atlas-specific example survives.
- Add the `/fix-pr-bug` (Phase 5) command.
- Restructure `CLAUDE.md` into the eight-heading rule-list shape, preserving every
  operative rule currently expressed as prose.
- Reconcile the existing `docs/superpowers-integration.md` against atlas's
  185-line version rather than overwriting it.
- Introduce a pinned Go linter (`golangci-lint`) with a checked-in configuration,
  and clear the resulting violation backlog so the flagless `tools/verify.sh`
  exits 0.

Non-goals:

- Any change to service behaviour, API surface, data model, or database schema.
- Porting atlas's `docs/packets/`, `docs/reverse-engineering.md`,
  `docs/adding-a-new-service.md`, or `docs/observability.md`.
- Work in `Harbormaster` or `MyFleet` — those are sibling task cycles.
- Building a cross-repository sync mechanism. Drift is accepted by design.
- Adding `golangci-lint` to `.github/workflows/` CI. This task makes the local
  gate real; wiring it into GitHub Actions is a follow-up decision.

## 3. User Stories

- As the repository owner, I want `/execute-task` to dispatch a budget-capped
  implementer so that a single plan task cannot silently consume an unbounded
  number of tool calls.
- As the repository owner, I want verification to run in a clean context via
  `task-verifier` so that a `PASS` verdict is not contaminated by the
  implementer's own assumptions.
- As a developer, I want one command — `tools/verify.sh` — whose flagless exit 0
  means "this branch may be called done", so I never have to remember which of
  three scripts to run.
- As a developer changing `shared/go/*`, I want the Docker image build to be part
  of the verification gate automatically, so the rule cannot be skipped by
  forgetting it.
- As a developer, I want Go and TypeScript files formatted on write so that
  formatting never appears as review noise.
- As an agent reading `CLAUDE.md`, I want a trigger → owner-document table so that
  I can find the governing procedure without loading 1,500 lines of context.
- As the repository owner, I want collision-safe task numbering so that two
  concurrent `/spec-task` runs cannot claim the same `NNN`.

## 4. Functional Requirements

### 4.1 Hooks — ported verbatim

Copy from `$ATLAS/.claude/hooks/` into `.claude/hooks/` with no edits:

| File | Wired at |
|---|---|
| `wait-loop-guard.sh` | `PreToolUse` / `Bash` |
| `wait-loop-guard_test.sh` | — (ships with its subject) |
| `block-home-paths-in-docs.sh` | `PreToolUse` / `Write\|Edit` |
| `turn-budget.sh` | `PostToolUse` / `*` |
| `turn-budget-guard.sh` | `PreToolUse` / `*` |
| `fork-dispatch-guard.sh` | `PreToolUse` / `Agent` |
| `commit-boundary.sh` | `PostToolUse` / `Bash` |
| `task-num-collision-detector.sh` | `SessionStart` |

Requirements:

- FR-H1. After copying, `grep -l 'atlas-' .claude/hooks/*.sh` must print nothing.
- FR-H2. Each copied file must be byte-identical to its atlas source at
  `e83f59e61`, verifiable by `diff`.
- FR-H3. `wait-loop-guard_test.sh` must pass when run in this repository.
- FR-H4. `commit-boundary.sh` references `tools/task-brief.sh`; that script must
  exist (FR-T2) so the reference is not dangling.
- FR-H5. `task-num-collision-detector.sh` requires `tools/task-numbers.sh`; that
  script must exist (FR-T1).

### 4.2 Hook — `format-on-write.sh` (adapted)

`format-on-write.sh` must **not** be copied verbatim. Two bindings change.

- FR-F1. **Go half.** home-hub gains `tools/toolchain.versions` as the single
  source of truth for toolchain pins, containing at minimum `GO_VERSION` and
  `GOLANGCI_LINT_VERSION`. The hook sources it and uses the cached pinned
  `golangci-lint` binary, exactly as atlas does.
- FR-F2. home-hub gains a checked-in `.golangci.yml`. No such file exists today.
- FR-F3. A bootstrap path must exist that caches the pinned `golangci-lint`
  binary. The hook itself must never bootstrap — atlas's hook deliberately uses
  the binary only if it has already been cached, to avoid a multi-minute stall on
  first Write. `tools/verify.sh` is the natural bootstrapper.
- FR-F4. **Frontend half.** Rebind from `services/atlas-ui` to `frontend/`.
  home-hub's frontend has **no prettier** dependency (eslint 10 only), so the
  formatter is `npx --no-install eslint --fix` run from `frontend/`.
- FR-F5. The hook must remain fail-open in every branch: missing toolchain,
  missing cached binary, unparseable input, tool error, non-absolute path — all
  exit 0 silently. A local convenience hook must never block an edit.

### 4.3 Agents

- FR-A1. Create `.claude/agents/task-implementer.md`, `task-verifier.md`, and
  `task-reviewer.md`, ported from `$ATLAS/.claude/agents/`.
- FR-A2. Names are the generic `task-*` forms. No `atlas-*` agent name may appear
  anywhere in this repository outside `docs/process-parity.md` (see FR-D8).
- FR-A3. Agent bodies must be genericized per specification §5.2 — atlas examples
  (packet work, WZ data, IDA, service opcodes, atlas verify flags) replaced with
  home-hub equivalents or neutral ones. A rule is never deleted because its
  example does not transfer.
- FR-A4. `task-implementer` retains its 120 tool-call budget and `PARTIAL`
  hand-back; `task-verifier` never edits; `task-reviewer` performs no recursive
  fan-out.
- FR-A5. The existing agents — `backend-guidelines-reviewer`,
  `frontend-guidelines-reviewer`, `plan-adherence-reviewer`, `todo-scanner`,
  `service-documentation` — are unchanged and must not be re-created.

### 4.4 Tools

`tools/` does not exist in home-hub and must be created.

- FR-T1. Port `tools/task-numbers.sh` and `tools/task-numbers_test.sh` from
  atlas. The test must pass in this repository.
- FR-T2. Port `tools/task-brief.sh` from atlas.
- FR-T3. Create `tools/toolchain.versions` (see FR-F1).
- FR-T4. Create `tools/verify.sh` — see §4.5.
- FR-T5. Every script in `tools/` must be executable (`chmod +x`) and pass
  `bash -n`.

### 4.5 `tools/verify.sh`

The contract is fixed by the specification and must match atlas's:

- FR-V1. **A flagless run that exits 0 means the branch may be called done.**
- FR-V2. `--quick` and `--no-docker` also exit 0 on success but skip the slow
  gates and explicitly do **not** count as done. The script must say so.
- FR-V3. `--no-docker` skips the Docker leg only. `--quick` implies `--no-docker`
  and additionally skips the slow test/lint legs.
- FR-V4. `-h` / `--help` prints usage and exits 0. An unknown option exits 2.
- FR-V5. The Go build leg covers the eight `shared/go/*` packages and the nine
  services, matching `scripts/ci-build.sh`'s coverage.
- FR-V6. The test leg matches `scripts/ci-test.sh`'s coverage (`go test ./...
  -count=1` per module).
- FR-V7. The lint leg runs the pinned `golangci-lint` over the same module set as
  `scripts/lint-all.sh`, plus `npx eslint .` in `frontend/`. It must bootstrap the
  pinned binary if it is not already cached.
- FR-V8. The frontend leg runs `npm ci` and `npm run build` (`tsc -b && vite
  build`) and `npm test` (`vitest run`).
- FR-V9. **Docker leg.** The flagless run builds the compose images when — and
  only when — the branch's diff against its base touches `shared/`. This is the
  existing `CLAUDE.md` Docker rule made mechanical. Detection is a
  `git diff --name-only <base>..HEAD` filtered on `^shared/`; the base defaults to
  the merge-base with `main` and is overridable with `--base <rev>`.
- FR-V10. `verify.sh` must not duplicate the logic in `scripts/*.sh` by
  copy-paste where composition is possible; the specification calls it "composed
  from" those scripts. Where a script's granularity is wrong for flag handling,
  reimplementing that leg is acceptable but must stay coverage-equivalent —
  divergence is a defect.
- FR-V11. Each leg prints a clearly delimited step banner and, on failure, stops
  at the first failing block with that block's output intact.

### 4.6 Linter introduction and backlog

home-hub has never been linted. There is no `.golangci.yml`, `golangci-lint` is
not installed on the development machine, and no CI workflow invokes it —
`scripts/lint-all.sh` is effectively dead code today.

- FR-L1. The design phase must, as its **first action**, bootstrap the pinned
  `golangci-lint` and produce a violation count per module. The size of this
  backlog is currently unknown and is the largest risk in the task.
- FR-L2. `.golangci.yml` must be checked in. The enabled linter set is a design
  decision informed by FR-L1.
- FR-L3. All violations reported by the chosen configuration must be fixed, so
  that the flagless `tools/verify.sh` exits 0 (specification §7 check 2). Fixes
  must be behaviour-preserving; any fix that changes behaviour is out of scope and
  must be raised rather than applied.
- FR-L4. `scripts/lint-all.sh` must be reconciled with the new pinned linter — it
  may delegate to `tools/verify.sh`, or be updated to use the pinned binary, but
  it must not continue to call an unpinned `golangci-lint` from `PATH`.

### 4.7 Owner documents

Port the nine documents from `$ATLAS/docs/` into `docs/`, genericized per §5.2.

| Document | atlas lines | Owns |
|---|---|---|
| `agent-dispatch.md` | 244 | Model pinning, fan-out vs. fork, handoff decision |
| `verification.md` | 368 | Gate failures, script/CI disagreement |
| `superpowers-integration.md` | 185 | Bare task numbers, skills outside a phase command |
| `review-protocol.md` | 178 | Dispatching a reviewer, writing up a review |
| `post-implementation.md` | 160 | Phase 5, `/fix-pr-bug` |
| `codemod-vs-agents.md` | 138 | Second implementer at the same transformation |
| `slice-first.md` | 107 | Reading a large document, diff, plan, or tool result |
| `tooling-conventions.md` | 95 | Long-running processes, mechanical repo facts, shell conventions |
| `git-workflow.md` | 52 | Committing, pushing, rebasing, stray `main` commits |

- FR-D1. All nine must exist in `docs/` when the task is done.
- FR-D2. Genericization replaces atlas illustrations with home-hub equivalents.
  Concretely: packet/WZ/IDA examples become recipe, meal-plan, tracker, or
  shopping-list examples drawn from this repository's actual services; atlas
  `verify.sh` flag specifics become home-hub's flags from §4.5.
- FR-D3. **No rule may be dropped because its example does not transfer.** If an
  example has no home-hub analogue, write a neutral one.
- FR-D4. `docs/superpowers-integration.md` **already exists here** at 76 lines and
  has diverged from atlas's 185. It must be reconciled by deliberate merge — diff
  the two, keep home-hub-specific content that atlas lacks, adopt atlas content
  home-hub lacks. Overwriting is explicitly forbidden.
- FR-D5. `docs/verification.md` must describe the `tools/verify.sh` built in §4.5,
  not atlas's.
- FR-D6. `docs/agent-dispatch.md` must name the generic `task-*` agents and must
  **not** carry atlas's historical-cutoff note about the rename — home-hub never
  used the `atlas-*` names.
- FR-D7. Do not port `docs/packets/`, `docs/reverse-engineering.md`,
  `docs/adding-a-new-service.md`, or `docs/observability.md`.
- FR-D8. `docs/process-parity.md` is committed in `docs/` and is the sole file
  exempt from the FR-D6 name check. `process-parity-brief.md` (the transient
  home-hub scoping notes for this task) lives in this task's folder, not in
  `docs/` directly.

### 4.8 Commands

- FR-C1. Port `.claude/commands/fix-pr-bug.md` from atlas as-is, adapting only
  repository-specific paths and command names.
- FR-C2. The existing nine commands are unchanged.

### 4.9 Settings

- FR-S1. `.claude/settings.json` gains `"disableBundledSkills": true`.
- FR-S2. It wires `PreToolUse` (`Write|Edit`, `Agent`, `Bash`, `*`), `PostToolUse`
  (`Write|Edit`, `*`, `Bash`), `SessionStart`, and `UserPromptSubmit` to the hooks
  named in §4.1 and §4.2, at the same events as atlas.
- FR-S3. The existing `UserPromptSubmit` → `skill-activation-prompt.sh` wiring and
  the `enabledPlugins` block are preserved.
- FR-S4. The file must remain valid JSON (`jq . .claude/settings.json`).

### 4.10 `CLAUDE.md` restructure

- FR-M1. Rewrite from prose narrative into the rule-list shape with exactly these
  eight headings, in this order: `## Never do this`, `## Evidence & grounding`,
  `## Development workflow`, `## Done means verified`, `## Dispatching agents`,
  `## Handing off context`, `## Repository conventions`,
  `## Where the procedures live`.
- FR-M2. The document ends with the `## Where the procedures live` trigger → owner
  table, and **every target file in that table must exist**.
- FR-M3. Rules currently in prose that must survive the rewrite, per specification
  §5.3:
  - verify Docker builds when shared libraries change (now expressed as "run
    flagless `tools/verify.sh`", which enforces it per FR-V9);
  - `scripts/local-up.sh` for local deployment;
  - the four-phase development workflow and its worktree discipline;
  - the artifact-location override (`docs/tasks/task-NNN-slug/`);
  - the code-review-before-PR rule;
  - the design/plan output style rule (write the full document, do not walk
    through it interactively);
  - the verification-over-memory rule.
- FR-M4. Repo-specific content that stays and is not homogenized: project
  overview, build commands, deployment specifics, domain conventions.
- FR-M5. Every `docs/` link in `CLAUDE.md` must resolve.

## 5. API Surface

Not applicable. This task changes no HTTP surface. No endpoint is added, removed,
or modified, and no service handler is touched except where FR-L3 requires a
behaviour-preserving lint fix.

## 6. Data Model

Not applicable. No entity, field, relationship, constraint, or migration is
introduced. No `tenant_id` scoping question arises because no persisted data is
involved.

## 7. Service Impact

No Go service changes behaviour. The impact is repository-level:

| Area | Change |
|---|---|
| `.claude/hooks/` | 8 files copied verbatim, 1 adapted (`format-on-write.sh`) |
| `.claude/agents/` | 3 new agent definitions |
| `.claude/commands/` | 1 new command (`/fix-pr-bug`) |
| `.claude/settings.json` | Full hook wiring + `disableBundledSkills` |
| `tools/` | New directory: `verify.sh`, `task-numbers.sh` + test, `task-brief.sh`, `toolchain.versions` |
| `docs/` | 8 new owner docs, 1 reconciled (`superpowers-integration.md`) |
| `CLAUDE.md` | Restructured |
| `.golangci.yml` | New |
| `scripts/lint-all.sh` | Reconciled against the pinned linter (FR-L4) |
| `services/*`, `shared/go/*` | Behaviour-preserving lint fixes only (FR-L3), scope unknown until FR-L1 |
| `frontend/` | Possible eslint `--fix` churn only |

## 8. Non-Functional Requirements

- NFR-1. **Fail-open vs. fail-closed.** `format-on-write.sh` is fail-open by
  design. The guard hooks (`wait-loop-guard`, `block-home-paths-in-docs`,
  `turn-budget-guard`, `fork-dispatch-guard`) are fail-closed and must stay so.
- NFR-2. **Verification latency.** A flagless run on a branch that does not touch
  `shared/` must not pay the Docker cost. `--quick` must be fast enough for an
  inner loop.
- NFR-3. **No secrets.** No hook or tool may read `.env` or emit credential
  material into logs.
- NFR-4. **Portability.** The eight verbatim hooks must remain byte-identical to
  atlas so a future re-harmonization is a file copy, not a merge.
- NFR-5. **No network at hook time.** Hooks must not fetch anything; the linter
  bootstrap belongs to `tools/verify.sh`, which may.
- NFR-6. Multi-tenancy, performance, and observability are unaffected — nothing in
  scope runs in production.

## 9. Open Questions

1. **Linter backlog size (blocking the plan, not the design).** Unknown until
   FR-L1 runs. If the count is large enough that fixing it dominates the task, the
   lint remediation may warrant splitting into `task-056`, leaving this task to
   land the tooling with the lint leg behind a flag. That call is the user's and
   must be surfaced with a number, not a guess.
2. **`.golangci.yml` content.** Adopt atlas's configuration wholesale, or start
   from a reduced set sized to the backlog? Depends on (1).
3. **`scripts/*.sh` future.** Once `tools/verify.sh` exists, do `ci-build.sh`,
   `ci-test.sh`, and `lint-all.sh` remain as independently callable entrypoints,
   or become thin delegates? FR-V10 permits either; the design should pick one.
4. **CI alignment.** `.github/workflows/pr.yml` runs Go 1.26 while `go.work`
   declares `go 1.26.1` and the local toolchain is 1.27.0. `tools/toolchain.versions`
   introduces a third place a Go version is written. Whether to add a pin-guard
   (atlas has `tools/toolchain-pin-guard.sh`) is a design decision.
5. **`--no-ui` flag.** atlas's `verify.sh` has one. Does home-hub want the
   equivalent for `frontend/`? Not required by the contract; cheap to add.

## 10. Acceptance Criteria

Mechanically checkable unless noted.

- [ ] AC-1. `.claude/hooks/` contains all nine hooks; `diff` against
      `$ATLAS/.claude/hooks/<f>` is empty for each of the eight verbatim ones.
- [ ] AC-2. `grep -l 'atlas-' .claude/hooks/*.sh` prints nothing.
- [ ] AC-3. `bash .claude/hooks/wait-loop-guard_test.sh` passes.
- [ ] AC-4. `bash tools/task-numbers_test.sh` passes.
- [ ] AC-5. `tools/verify.sh`, `tools/task-numbers.sh`, `tools/task-brief.sh`,
      and `tools/toolchain.versions` all exist; the three scripts are executable.
- [ ] AC-6. `tools/verify.sh --help` exits 0; an unknown flag exits 2.
- [ ] AC-7. `tools/verify.sh --quick` exits 0 and states that it does not count as
      done.
- [ ] AC-8. **Flagless `tools/verify.sh` exits 0.** (Specification §7 check 2.)
- [ ] AC-9. On a branch touching `shared/`, the flagless run performs the compose
      image build; on one that does not, it skips it. Demonstrated, not asserted.
- [ ] AC-10. `.claude/agents/` defines `task-implementer`, `task-verifier`, and
      `task-reviewer`.
- [ ] AC-11. This prints nothing (specification §7 check 3, home-hub carve-out):
      ```sh
      git grep -lE 'atlas-(implementer|verifier|reviewer)' -- . ':!docs/tasks' \
        | grep -vxE 'docs/process-parity\.md'
      ```
- [ ] AC-12. `jq . .claude/settings.json` succeeds; `disableBundledSkills` is
      `true`; every hook path in it resolves to an existing executable file.
- [ ] AC-13. `CLAUDE.md` contains the eight §5.3 headings in order and ends with
      the `## Where the procedures live` table.
- [ ] AC-14. Every `docs/` link in `CLAUDE.md` resolves to an existing file.
- [ ] AC-15. All nine owner documents exist in `docs/`.
- [ ] AC-16. No owner document contains the strings `atlas-ui`, `WZ`, `IDA`, or
      `packet` outside a deliberate, reviewed context.
- [ ] AC-17. `docs/superpowers-integration.md` is a reconciliation, not an
      overwrite — the diff against both prior versions shows retained content from
      each. Reviewed by a human, not mechanical.
- [ ] AC-18. `.claude/commands/fix-pr-bug.md` exists.
- [ ] AC-19. `.golangci.yml` exists and `scripts/lint-all.sh` no longer invokes an
      unpinned `golangci-lint` from `PATH`.
- [ ] AC-20. Specification §7 checks 1, 4, 5, and 6 are cross-repository and
      **cannot be evaluated from home-hub alone.** home-hub's side of each must be
      reported, and the pairwise comparison explicitly declared not evaluable here.
      Asserting them as passed is a defect.
