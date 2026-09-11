# Home Hub

Go microservices project, 12 services plus a React/TypeScript frontend.
Go is the primary language; TypeScript is used only for the frontend UI.

## Never do this

- Never start implementing code changes when asked only to understand or plan
  something. Planning and implementation are separate phases — wait for
  explicit approval before making any edits.
- Never edit files in the main repo when a task worktree exists for that
  work.
- Never open a PR without the code-review step, even when the plan looks
  complete.
- Never claim "done" or "verified" from a flagged or partial `tools/verify.sh`
  run.
- Never break a service boundary by having one layer call another's
  internals directly.

## Evidence & grounding

- For API contracts, schemas, configuration values, and service-to-service
  interactions, verify against local source rather than citing values from
  memory or general knowledge.
- When uncertain about a JSON:API shape, migration state, or which service
  owns a behavior, read the source rather than speculating.

## Development workflow

- Asked to understand or plan? Do not implement. Wait for explicit approval
  before any edit.
- The canonical flow for any non-trivial change is four phases —
  `/spec-task` → `/design-task` → `/plan-task` → `/execute-task` — each a
  separate slash command, invoked from a fresh (`/clear`'d) session so the
  next phase consumes only the prior phase's documented artifacts:
  1. `/spec-task <idea>` — run from the main repo. Creates the worktree at
     `.worktrees/task-NNN-slug/` on branch `task-NNN-slug` and commits the
     PRD to `docs/tasks/task-NNN-slug/prd.md`.
  2. `cd .worktrees/task-NNN-slug`, `/clear`, then `/design-task <task-id>`
     — invokes `superpowers:brainstorming`. Output: `design.md`.
  3. `/clear`, then `/plan-task <task-id>` — invokes
     `superpowers:writing-plans`. Output: `plan.md` + `context.md`.
  4. `/clear`, then `/execute-task <task-id>` — invokes
     `superpowers:subagent-driven-development`. Reuses the existing
     worktree; never creates a new one.
  Phase commands accept fuzzy task identifiers: `task-044-slug`, `task-044`,
  `044`, or `44` all resolve to the same folder. Skip `/spec-task` only for
  trivial fixes that don't warrant a PRD; document those directly via a
  brainstorming session.
- `superpowers:brainstorming` and `superpowers:writing-plans` default to
  `docs/superpowers/specs/` and `docs/superpowers/plans/` — in this project
  both go under `docs/tasks/task-NNN-slug/` instead. When invoking those
  skills directly (outside the phase commands), pass the task folder
  explicitly so artifacts land in the right place.
- Verify cwd is the correct worktree before planning, designing, or
  executing a task; if it is not, tell the user to `cd` into it rather than
  proceeding from the wrong directory. When searching for a task's
  PRD/design/plan, search across all worktrees (`git worktree list`) before
  concluding a file is missing.
- When producing `design.md` or `plan.md`, write the full document directly
  to the file; do not walk through sections interactively or ask for
  per-section approval. The user reads the committed file.

## Done means verified

- A branch is done when flagless `tools/verify.sh` exits 0. Nothing else is
  "done" — not `--quick`, not `--no-docker`, not "the tests passed".
- The old "verify Docker builds when shared libraries change" rule is now
  mechanical: the flagless run's docker leg fires on any `shared/` change
  and fans out to all 12 service images. You cannot forget it by forgetting
  it. Expect several fix-and-rebuild cycles for large refactors.
- Always run the code-review step before opening a PR, even when the plan
  looks complete.

## Dispatching agents

- Code review is a separate gate from verification and uses named reviewer
  agents (`plan-adherence-reviewer`, `backend-guidelines-reviewer`,
  `frontend-guidelines-reviewer`), dispatched in parallel via
  `superpowers:requesting-code-review`, each writing findings to
  `docs/tasks/task-NNN-slug/audit.md`. See
  [`docs/review-protocol.md`](docs/review-protocol.md) for what a reviewer
  returns, and [`docs/superpowers-integration.md`](docs/superpowers-integration.md)
  for the full when-to-use-what reference.

## Handing off context

- At every durable boundary — a commit landing, a gate returning, a
  fan-out reporting — ask whether the next unit of work depends materially
  on this conversation's history, or only on repository state. If it can be
  resumed from repo state and a short written diagnosis, hand off rather
  than continuing to accumulate context.

## Repository conventions

- Use `scripts/local-up.sh` to build and start all services locally via
  Docker Compose. It handles the `.env` file and build context
  automatically.
- When refactoring shared types or creating common libraries, prefer
  straightforward moves over re-exporting type aliases. Keep abstractions
  clean — don't break service boundaries by having one layer call another's
  internals directly.
- JSON:API is the transport contract between frontend and backend; GORM
  entities are tenant-scoped via `tenant_id`. Verify these against source
  rather than memory (see Evidence & grounding).

## Context & cost guards

This repo opts in to the guards in `~/.claude/hooks/` via
`.claude/guards.json`. They are enforced by hooks, not by honor system —
several will refuse a tool call outright. Full reference:
`~/.claude/hooks/README.md`.

- **Hand off past ~150k context.** The controller never carries a session
  past ~60 tool calls (≈150k tokens) into a *new* unit of work —
  unconditionally, no carve-out for tasks remaining. Write the diagnosis
  into `docs/tasks/task-NNN-slug/`, then have the user `/clear` and re-run
  the phase command; it resumes from the committed artifacts. Finishing the
  unit in flight (reviewers, verifiers, doc agents) is still allowed.
- **Subagents stop at 120 tool calls.** Commit what works and report PARTIAL
  with what is done file by file, what remains, and the exact next step.
  PARTIAL at the cap is the contracted outcome, not a failure — the
  controller dispatches a continuation with fresh context.
- **Brief a fresh agent; do not fork.** A fork re-reads this entire
  conversation on every turn. Dispatch a named agent type with an explicit
  brief instead; when sharding a review, give each child the artifact path
  plus its own scope.
- **Never spend a turn waiting.** No `sleep`, no `ps aux`/`pgrep` polling,
  no re-reading a log until it changes. Use `run_in_background: true` and
  let the completion notify you, or `Monitor` with an `until` loop and an
  explicit timeout.
- **No absolute home paths under `docs/`.** Write repo-relative paths.

Each of these has a one-line escape hatch when the exception is real:
`CONTEXT-JUSTIFIED:`, `FORK-JUSTIFIED:`, or `POLL-JUSTIFIED:` followed by
the reason, anywhere in the prompt or command.

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
