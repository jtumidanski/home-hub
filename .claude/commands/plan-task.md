---
description: Phase 3 — invoke superpowers:writing-plans to produce an implementation plan inside the task worktree
argument-hint: Task identifier — accepts "task-044-superpowers-integration", "task-044", "044", or "44"
---

You are starting Phase 3 of the Home Hub four-phase development workflow. Argument: **$ARGUMENTS**

## Process

### Step 1 — Resolve the task

Same as `/design-task` Step 1 — run `tools/task-resolve.sh "$ARGUMENTS"`
(prints `<task-id>\t<task-dir>\t<worktree>`; exit 3 no match, 4 ambiguous
with candidates on stderr). Never glob `.worktrees/*/docs/tasks/task-*`. If
`<worktree>` is the main repo root, stop and tell the user the task needs a
worktree.

### Step 2 — Ensure we're in the right worktree

Run `pwd`. If it does NOT match `<worktree>`, `cd <worktree>` yourself and continue from there. Do NOT ask the user to re-run the command — per CLAUDE.md's Development workflow rule, cd into the task worktree yourself.

### Step 3 — Validate inputs

1. Confirm both `prd.md` and `design.md` exist. If either is missing, stop and tell the user to complete the prior phase.
2. Confirm `plan.md` does NOT already exist. If it does, ask whether to overwrite.

### Step 4 — Load context

Read:
- `<worktree>/docs/tasks/<id>/prd.md`
- `<worktree>/docs/tasks/<id>/design.md`
- `<worktree>/CLAUDE.md`
- `<worktree>/docs/superpowers-integration.md`
- Code areas the design touches

### Step 5 — Invoke writing-plans

Use the Skill tool to invoke `superpowers:writing-plans`. Pass:

- Spec at `<worktree>/docs/tasks/<id>/design.md` (PRD at `prd.md` for reference).
- Plan output MUST be saved to `<worktree>/docs/tasks/<id>/plan.md`.
- Also produce `<worktree>/docs/tasks/<id>/context.md` summarizing key files, decisions, dependencies.
- Do NOT auto-invoke execution.

Run the `writing-plans` skill's self-review (placeholder scan, type consistency, spec coverage) before saving.

### Step 6 — Commit and summarize

```
git add docs/tasks/<id>/plan.md docs/tasks/<id>/context.md
git commit -m "plan(<id>): implementation plan and context"
```

Verify post-commit:

```
git rev-parse --show-toplevel  # must end with /.worktrees/<id>
git branch --show-current      # must be <id>
```

If either is wrong, STOP and report BLOCKED. Then tell the user:

> Plan and context saved and committed. Now run `/clear`, then `/execute-task <id>`. (You're already in the right worktree.)

## Important Rules

- All file I/O uses absolute paths under `<worktree>`.
- Never write plan artifacts under main's `docs/tasks/`.
- DO NOT begin implementation. This phase produces planning documents only.
