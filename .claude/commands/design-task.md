---
description: Phase 2 — invoke superpowers:brainstorming to produce a design doc inside the task worktree
argument-hint: Task identifier — accepts "task-044-superpowers-integration", "task-044", "044", or "44"
---

You are starting Phase 2 of the Home Hub four-phase development workflow. Argument: **$ARGUMENTS**

## Process

### Step 1 — Resolve the task

1. **Run the resolver** — do NOT glob for the folder yourself:

   ```sh
   tools/task-resolve.sh "$ARGUMENTS"
   ```

   It prints `<task-id>\t<task-dir>\t<worktree>` and accepts exact
   (`task-NNN-slug`), number-only (`44`/`044`/`task-44`/`task-044`), or
   slug-fragment identifiers. Never glob `.worktrees/*/docs/tasks/task-*`:
   every worktree carries a full copy of `docs/tasks/`, so that pattern
   returns (tasks × worktrees) mostly-duplicate paths.

2. Exit 3 (no match): stop and ask the user for a corrected identifier.
3. Exit 4 (ambiguous): list the candidates from stderr and ask the user to pick.
4. If `<worktree>` is the main repo root, the task has no worktree — that's an error state — the four-phase workflow requires a task worktree. Stop and tell the user:
   > Task `<id>` exists on main but has no worktree. The current workflow expects every task to have its own worktree (created by `/spec-task`). Either move the task into a worktree or run `/spec-task` from scratch.
5. Otherwise, the resolved location is `<worktree>/docs/tasks/<id>/`. Record `<worktree>` as the absolute path you'll use for all subsequent operations.

### Step 2 — Ensure we're in the right worktree

Run `pwd`. If it does NOT match `<worktree>`, `cd <worktree>` yourself and continue from there. Do NOT ask the user to re-run the command — per CLAUDE.md's Development workflow rule, cd into the task worktree yourself.

### Step 3 — Validate inputs

1. Confirm `<worktree>/docs/tasks/<id>/prd.md` exists. If not, tell the user to run `/spec-task` first.
2. Confirm `design.md` does NOT already exist. If it does, ask whether to overwrite or open the existing one.

### Step 4 — Load context

Read:
- `<worktree>/docs/tasks/<id>/prd.md`
- `<worktree>/CLAUDE.md`
- `<worktree>/docs/superpowers-integration.md`
- Code areas implied by the PRD's Service Impact section

### Step 5 — Invoke brainstorming

Use the Skill tool to invoke `superpowers:brainstorming`. Pass:

- The PRD is at `<worktree>/docs/tasks/<id>/prd.md` and is approved — SKIP the default what/why questions.
- Focus on architecture, alternatives, tradeoffs.
- Output MUST be saved to `<worktree>/docs/tasks/<id>/design.md` (NOT the skill's default location).
- Do NOT auto-invoke `writing-plans`. The user runs `/clear` then `/plan-task <id>` separately.

### Step 6 — Commit and summarize

Once the design is approved, commit it on the task branch:

```
git add docs/tasks/<id>/design.md
git commit -m "design(<id>): architecture and tradeoffs"
```

Verify post-commit:

```
git rev-parse --show-toplevel  # must end with /.worktrees/<id>
git branch --show-current      # must be <id>
```

If either is wrong, STOP and report BLOCKED. Then tell the user:

> Design saved and committed. Now run `/clear`, then `/plan-task <id>`. (You're already in the right worktree.)

## Important Rules

- All file I/O uses absolute paths under `<worktree>`.
- Never write design artifacts under main's `docs/tasks/`.
- DO NOT begin implementation. This phase produces a design document only.

Write the full design.md in one shot. Commit it. Reply only with the file path and commit SHA — do NOT summarize or walk through sections.
