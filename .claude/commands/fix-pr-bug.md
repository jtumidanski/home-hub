---
description: Phase 5 — diagnose a post-implementation bug into a durable file, then fix it in a fresh context
argument-hint: Task identifier plus a short bug slug — e.g. "227 recipe-import-crash"
---

You are running Phase 5 of the Home Hub workflow: a bug found after
implementation — in PR validation, live testing, or regression — fixed
without carrying a debugging conversation into the fix.

Argument: **$ARGUMENTS** (a task identifier, then a short bug slug).

The full rationale and the loop this command mechanizes live in
[`docs/post-implementation.md`](../../docs/post-implementation.md). Read it if
anything below is ambiguous.

## Step 1 — Resolve the task and confirm the workspace

Home-hub has no single "facts" script; assemble the block yourself, cheaply,
rather than re-deriving it by hand later:

1. Resolve the task folder using the same fuzzy-match algorithm as
   `/execute-task` Step 1: glob `docs/tasks/task-*` (main) and
   `.worktrees/*/docs/tasks/task-*` (sibling worktrees), and match the task
   identifier in **$ARGUMENTS** against folder names — exact, number-only
   (`44`/`044`/`task-44`/`task-044`), or slug fragment.
2. `cd` into `<worktree>` yourself if `pwd` does not match. Do not create a
   worktree.
3. `git branch --show-current` and `git rev-parse --show-toplevel` — confirm
   branch and worktree agree with the resolved task.
4. `ls docs/tasks/<task>/` — note which artifacts already exist (`prd.md`,
   `design.md`, `plan.md`, `context.md`, any prior `bug-*.md`).
5. `tools/verify.sh --facts --quick --base <last commit you know is gated>` —
   the changed services/modules and the legs that would run, without
   building anything.

## Step 2 — Reproduce, in your own context

Reproduction is interactive and stays here. Confirm the tenant and the exact
service/build you're reproducing against first. Read service logs before
anything else — `deploy/compose/logs.sh <service>` — and read the logs for
the service(s) you name, never a whole-stack dump.

If a prior `docs/tasks/<task>/bug-*.md` already describes this symptom, read
that instead of reproducing from scratch.

## Step 3 — Write the diagnosis to disk

Write `docs/tasks/<task>/bug-<slug>.md` using the template in
`docs/post-implementation.md`: reproduced / observed / expected / root cause,
then a `## Fix` file inventory, then `## Not yet answered`.

**Write it before dispatching anything.** This file is the boundary — after it,
everything must be resumable from repository state plus this file. If the root
cause is not established, say so and name what is ruled out; do not guess one.

The `## Fix` inventory is what removes the fix agent's discovery phase. You
already know the paths from reproducing; without them the agent pays to
rediscover what you just found, at its own context depth.

## Step 4 — Dispatch a fresh implementer

```text
subagent_type: task-implementer
model: sonnet
```

Brief: the bug file path. Add only what the file cannot carry — the worktree
absolute path, and any ruling you made after writing it. Do not restate the
file, and do not restate the agent's own contracts.

On `PARTIAL`, follow `/execute-task` Step 4d: continuation brief, fresh agent,
same report file.

## Step 5 — Verify in a clean context

Launch the gate for the fix range and **keep going** — do not poll it:

```sh
tools/verify.sh --quick --base <last-gated-commit> > /tmp/gate-<slug>.log 2>&1
```

with `run_in_background: true`, or dispatch `task-verifier` (`model: haiku`)
for a summarized verdict.

If the fix crosses a service boundary or changes a contract, also dispatch
`task-reviewer` (`model: sonnet`) against the fix range with the bug file as
the requirement — the gate cannot see a seam defect.

## Step 6 — Record the outcome

Home-hub has no separate ledger tool; the bug file is the record. Update
`docs/tasks/<task>/bug-<slug>.md` with the outcome — the commit that fixed it,
which agents ran (implementer, and verifier or reviewer if dispatched) with
their verdicts, and whether live testing confirmed it. A bug file that never
records its resolution is the next session's rediscovery.

## Step 7 — Decide whether to continue here

Ask the handoff question: does the next bug depend on this conversation, or only
on repository state? An unrelated bug is a fresh unit — run `/fix-pr-bug` again
against its own bug file rather than accumulating.

Past ~150k context, stop starting new investigations in this session: write the
remaining leads into the task folder and say so as your final output. A handoff
the same context then works past is not a handoff.

## Important rules

- Never reproduce inside a subagent; never implement the fix inside this context.
- Never dispatch without a bug file on disk.
- Never poll a background process — `.claude/hooks/wait-loop-guard.sh` refuses it.
- Never commit to `main`; the fix lands on the task branch.
- Never claim the bug fixed without the gate verdict and, where applicable, a
  live re-test.
