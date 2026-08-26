# Superpowers Integration — When to Use What

This document is the quick-reference companion to `CLAUDE.md`. It tells you which command, agent, or skill to reach for in each situation. The full architectural design lives in `docs/tasks/task-044-superpowers-integration/design.md`.

This document owns *which* command, agent, or skill to reach for in a given
situation. [`docs/agent-dispatch.md`](agent-dispatch.md) owns *how* to
dispatch any agent — model, budget, isolation, handoff.

## The Four-Phase Workflow

| Phase | Command | What it does | Output |
|---|---|---|---|
| 1. Requirements | `/spec-task <idea>` | Interactive PRD interview | `docs/tasks/task-NNN-slug/prd.md` |
| 2. Design | `/design-task <task-folder>` | Architecture, alternatives, tradeoffs | `design.md` |
| 3. Plan | `/plan-task <task-folder>` | Bite-sized TDD step-by-step plan | `plan.md` + `context.md` |
| 4. Execute | `/execute-task <task-folder>` | Subagent-driven implementation | code + commits |
| 5. After implementation | `/fix-pr-bug <task> <bug-slug>` | PR validation, live testing, debugging, follow-up fixes — diagnosis to a durable file, fix in a fresh context | `bug-<slug>.md` + fix commits |

Run `/clear` between phases 1–4. Each command consumes only the prior phase's documented artifacts.

Phase 5 is not a `/clear` boundary and does not require one: it hands off by
writing a diagnosis and dispatching against it. See
[`docs/post-implementation.md`](post-implementation.md) for the full loop —
reproduce inline, diagnose into a file, delegate the fix, verify fresh.

### Task resolution

Phase commands accept fuzzy task identifiers: `task-044-slug`, `task-044`,
`044`, and `44` all resolve to the same folder.

- Home-hub has no single resolver/facts script. `/design-task`,
  `/execute-task`, and `/fix-pr-bug` Step 1 each glob `docs/tasks/task-*`
  (main) and `.worktrees/*/docs/tasks/task-*` (sibling worktrees) and match
  the identifier against folder names — exact, number-only, or slug
  fragment.
- `tools/task-numbers.sh next` picks the smallest unused number for a new
  task; `tools/task-numbers.sh check` reports any number with more than one
  distinct task ID; `tools/task-numbers.sh list` prints every assignment
  seen. Check `next` before planning so a number does not collide with an
  in-flight task.
- `tools/task-brief.sh <plan-file> <task-number> [outfile]` extracts one
  task's full text from `plan.md` into a standalone brief an implementer
  reads in one call — the vendored replacement for the plugin's
  `scripts/task-brief`, used by `/execute-task` Step 4b.
- Searching for a task artifact: search across all worktrees
  (`git worktree list`) before concluding a file is missing.

### Artifact location override

`superpowers:brainstorming` and `superpowers:writing-plans` default to
`docs/superpowers/specs/` and `docs/superpowers/plans/`. In this project both
go under `docs/tasks/task-NNN-slug/` instead. When invoking those skills
directly, outside the phase commands, pass the task folder explicitly so
artifacts land in the right place.

### Phase 4 context budget

`task-implementer` replaces `general-purpose` for every Phase 4
implementation dispatch. Its contracts override the plugin's
`implementer-prompt.md` where they disagree — in particular, the plugin
template's "run the full suite once before committing" does not apply here:
implementers run module-local `go build ./... && go test ./...` only, never
`tools/verify.sh`; the repo gate belongs to `task-verifier`, in its own clean
context.

The tool-call cap (120, warned at 100), the `PARTIAL` hand-back, the
verification split, and the front-loaded file inventory are owned by
[`docs/agent-dispatch.md`](agent-dispatch.md).

## Code Review

Invoke `superpowers:requesting-code-review` after completing a logical chunk of work. The skill dispatches the relevant subset of these agents in parallel:

- `plan-adherence-reviewer` — checks every task in `plan.md` was implemented; cites file:line evidence
- `backend-guidelines-reviewer` — adversarial Go audit (DOM-*, SUB-*, SEC-* checks) when Go files changed
- `frontend-guidelines-reviewer` — adversarial TS/React audit (FE-* checks) when TS files changed
- `task-reviewer` — per-unit / ad-hoc correctness review of one commit range against its brief. This is the named home for what used to ride bare `general-purpose`; use it rather than dispatching `general-purpose` with a review prompt.

For ad-hoc one-off checks, invoke any agent directly by name without the orchestration skill.

### Picking the roster

Home-hub has no classifier script for this — pick the roster by hand from
what changed: Go files changed → `backend-guidelines-reviewer`; frontend TS
files changed → `frontend-guidelines-reviewer`; a `plan.md` exists for the
work under review → `plan-adherence-reviewer`. When in doubt about whether a
family applies, include it — a reviewer that is not needed says so cheaply;
a missing reviewer is a silent gap.

### What a reviewer returns

Every reviewer writes its full reasoning to a durable artifact and returns a
compact verdict-first block. The contract, the verdict semantics, and the
controller's read rule are in [`docs/review-protocol.md`](review-protocol.md).
Short version: `verdict` is the first line, blocking findings are enumerated
with `file:line`, everything else is a count, and the controller opens the
artifact only when the verdict is not `APPROVED`.

Each agent writes findings to `docs/tasks/task-NNN-slug/audit.md` (backend
also writes `audit.json`).

Code review is mandatory before opening a PR and is a **different gate**
from verification: a green `tools/verify.sh` does not mean the branch is
correct. Every module can build, vet, test, and bake clean while the branch
carries blocking defects, because each service is self-consistent in
isolation. The gate cannot see a producer emptying a compartment the
consumer still reads, or a class of missing handling that existing tests
actively pin as the old behavior. When a change crosses a service boundary,
trace the change into its consumers by hand and check that a test asserts
the new contract, not the old silent gap.

## Maintenance Commands

| Command | What it does | Underlying agent |
|---|---|---|
| `/review-todos` | Whole-codebase TODO/FIXME scan; generates/updates `docs/TODO.md` | `todo-scanner` |
| `/service-doc <service>` | Generates/updates documentation for one service | `service-documentation` |
| `/recipe-to-cooklang` | Converts recipe text to Cooklang format | (no agent — direct command) |

## Domain Skills

These activate via the project hook (`.claude/hooks/skill-activation-prompt.sh`) when you mention relevant keywords or work on relevant files:

- `backend-dev-guidelines` — Go service patterns
- `frontend-dev-guidelines` — React/TypeScript patterns

The hook produces a visible "🎯 SKILL ACTIVATION CHECK" banner. Heed it before responding.

## Superpowers Skills (Self-Activating)

Reach for these explicitly when relevant; they also self-activate via Claude's native skill matching:

- `using-superpowers` — invoke at the start of any conversation
- `brainstorming` — used inside `/design-task`
- `writing-plans` — used inside `/plan-task`
- `subagent-driven-development` — used inside `/execute-task`
- `executing-plans` — fallback for inline execution
- `systematic-debugging` — for any bug, test failure, or unexpected behavior
- `test-driven-development` — when implementing any feature or bugfix
- `verification-before-completion` — before claiming work is complete
- `using-git-worktrees` — for isolated workspaces
- `finishing-a-development-branch` — when implementation is complete and tests pass
- `requesting-code-review` — used at the end of a chunk of work
- `receiving-code-review` — when processing review feedback
- `dispatching-parallel-agents` — used by code-review orchestration
- `writing-skills` — when authoring new skills

## When NOT to Use Superpowers

- **Trivial fixes** (typo, version bump, one-line change) — no workflow needed; commit directly.
- **Documentation-only updates** that don't need a PRD — go straight to editing.
- **Personal recipe conversion** — use `/recipe-to-cooklang` directly.

## File Locations Cheat Sheet

| Artifact | Location |
|---|---|
| PRD, design, plan, context, audit | `docs/tasks/task-NNN-slug/` |
| Audit JSON output (backend) | `docs/tasks/task-NNN-slug/audit.json` |
| Per-service docs | `services/<service>/docs/` |
| TODO list (generated by `/review-todos`; not present until first run) | `docs/TODO.md` |
| Recipes | `recipes/` |
| Verification, tooling, git-workflow, slice-first, agent-dispatch, review-protocol, post-implementation, codemod-vs-agents, superpowers-integration docs | `docs/*.md` |
| Repo tooling (`verify.sh`, `task-numbers.sh`, `task-brief.sh`) | `tools/` |
