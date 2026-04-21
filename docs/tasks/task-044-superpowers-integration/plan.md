# Superpowers Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the superpowers plugin the default development workflow for Home Hub while preserving the project's numbered-task-folder discipline, PRD-style requirements, and project-specific tooling.

**Architecture:** Four phase-specific slash commands (`/spec-task` → `/design-task` → `/plan-task` → `/execute-task`) wrap superpowers skills, producing artifacts under `docs/tasks/task-NNN-slug/`. Three modular reviewer agents replace the existing audit commands. Two maintenance commands are promoted to proper agents. The project hook keeps its narrow domain-skill role; superpowers skills self-activate.

**Tech Stack:** Markdown agent/command/skill files, YAML frontmatter, JSON skill rules, Python skill-activation hook (unchanged).

**Companion documents:**
- `docs/tasks/task-044-superpowers-integration/design.md` — full design
- `docs/tasks/task-044-superpowers-integration/context.md` — implementation context

---

## File Plan

### New files

| Path | Purpose |
|---|---|
| `.claude/commands/design-task.md` | Phase 2 wrapper invoking `superpowers:brainstorming` |
| `.claude/commands/plan-task.md` | Phase 3 wrapper invoking `superpowers:writing-plans` |
| `.claude/commands/execute-task.md` | Phase 4 wrapper invoking `superpowers:subagent-driven-development` |
| `.claude/agents/plan-adherence-reviewer.md` | Replaces `/audit-plan` body |
| `.claude/agents/backend-guidelines-reviewer.md` | Replaces `/backend-audit` body |
| `.claude/agents/frontend-guidelines-reviewer.md` | New — mirrors backend reviewer for TS/React |
| `.claude/agents/todo-scanner.md` | Replaces `/review-todos` body |
| `.claude/agents/service-documentation.md` | Renamed from `documentation.md`, with proper frontmatter; absorbs `/service-doc` body |
| `docs/superpowers-integration.md` | When-to-use-what reference |

### Modified files

| Path | Change |
|---|---|
| `.claude/commands/spec-task.md` | Update Step 5 handoff suggestion (was `/dev-docs` → becomes `/clear` then `/design-task`) |
| `.claude/commands/audit-plan.md` | Shrink to one-line wrapper invoking `plan-adherence-reviewer` agent |
| `.claude/commands/backend-audit.md` | Shrink to one-line wrapper invoking `backend-guidelines-reviewer` agent |
| `.claude/commands/review-todos.md` | Shrink to one-line wrapper invoking `todo-scanner` agent |
| `.claude/commands/service-doc.md` | Shrink to one-line wrapper invoking `service-documentation` agent |
| `.claude/skills/skill-rules.json` | Add `frontend-dev-guidelines` entry |
| `CLAUDE.md` | Add workflow section, location override, code-review pattern |

### Removed files

| Path | Reason |
|---|---|
| `.claude/commands/dev-docs.md` | Role split between `/design-task` and `/plan-task` |
| `.claude/agents/documentation.md` | Renamed to `service-documentation.md` |

---

## Execution Order Rationale

1. **Tasks 1–2 (Documentation foundation)** — write the workflow docs first so subsequent agent/command files can reference them by path.
2. **Tasks 3–7 (Phase commands)** — replace the planning workflow.
3. **Tasks 8–13 (Code review agents + slash wrappers)** — replace the audit commands.
4. **Tasks 14–17 (Maintenance agents + slash wrappers)** — replace `/review-todos` and `/service-doc`.
5. **Task 18 (Hook)** — add `frontend-dev-guidelines` to skill-rules.
6. **Tasks 19–20 (Removals)** — delete `dev-docs.md` and the old `documentation.md` last, after their replacements are in place.
7. **Task 21 (Smoke test)** — manual verification that the new workflow works end-to-end on a tiny throwaway task.

---

## Task 1: Update CLAUDE.md with workflow section

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Read the current CLAUDE.md**

Run: `cat CLAUDE.md`

Confirm the current sections: Project Overview, Workflow Rules, Build & Verification, Local Deployment, Code Patterns.

- [ ] **Step 2: Append the new sections**

Append the following content to the end of `CLAUDE.md`:

```markdown

## Development Workflow

The canonical flow for any non-trivial change is four phases. Each phase is a separate slash command, each invoked from a fresh (`/clear`'d) session so the next phase consumes only the prior phase's documented artifacts:

1. `/spec-task <idea>` — interactive PRD interview. Output: `docs/tasks/task-NNN-slug/prd.md`.
2. `/clear`, then `/design-task <task-folder>` — invokes `superpowers:brainstorming` for architecture/tradeoffs. Output: `design.md` in same folder.
3. `/clear`, then `/plan-task <task-folder>` — invokes `superpowers:writing-plans` for bite-sized TDD steps. Output: `plan.md` + `context.md`.
4. `/clear`, then `/execute-task <task-folder>` — invokes `superpowers:subagent-driven-development` (default) or `superpowers:executing-plans` (fallback).

Skip `/spec-task` only for trivial fixes that don't warrant a PRD; document those directly via a brainstorming session.

### Artifact Location Override

Both `superpowers:brainstorming` and `superpowers:writing-plans` default to `docs/superpowers/specs/` and `docs/superpowers/plans/`. **In this project, both go under `docs/tasks/task-NNN-slug/` instead.** When invoking those skills directly (outside the phase commands), pass the task folder explicitly so artifacts land in the right place.

### Code Review Pattern

Code review uses three modular reviewer agents, dispatched in parallel:

- `plan-adherence-reviewer` — verifies plan tasks were actually implemented
- `backend-guidelines-reviewer` — Go DOM-* checklist (when Go files changed)
- `frontend-guidelines-reviewer` — TS/React FE-* checklist (when TS files changed)

Invoke via `superpowers:requesting-code-review` (it dispatches the appropriate subset), or invoke an individual agent directly for ad-hoc checks. Each agent writes its findings to `docs/tasks/task-NNN-slug/audit.md`.

See `docs/superpowers-integration.md` for a complete when-to-use-what reference.
```

- [ ] **Step 3: Verify the file**

Run: `wc -l CLAUDE.md && tail -40 CLAUDE.md`

Expected: line count increased, the new sections render at the bottom.

- [ ] **Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: document the four-phase workflow and code-review pattern in CLAUDE.md"
```

---

## Task 2: Create docs/superpowers-integration.md

**Files:**
- Create: `docs/superpowers-integration.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls docs/superpowers-integration.md 2>&1 || echo "absent"`

Expected: `absent` (or `No such file`).

- [ ] **Step 2: Create the file**

Write the following content to `docs/superpowers-integration.md`:

```markdown
# Superpowers Integration — When to Use What

This document is the quick-reference companion to `CLAUDE.md`. It tells you which command, agent, or skill to reach for in each situation. The full architectural design lives in `docs/tasks/task-044-superpowers-integration/design.md`.

## The Four-Phase Workflow

| Phase | Command | What it does | Output |
|---|---|---|---|
| 1. Requirements | `/spec-task <idea>` | Interactive PRD interview | `docs/tasks/task-NNN-slug/prd.md` |
| 2. Design | `/design-task <task-folder>` | Architecture, alternatives, tradeoffs | `design.md` |
| 3. Plan | `/plan-task <task-folder>` | Bite-sized TDD step-by-step plan | `plan.md` + `context.md` |
| 4. Execute | `/execute-task <task-folder>` | Subagent-driven implementation | code + commits |

Run `/clear` between phases. Each command consumes only the prior phase's documented artifacts.

## Code Review

Invoke `superpowers:requesting-code-review` after completing a logical chunk of work. The skill dispatches the relevant subset of these agents in parallel:

- `plan-adherence-reviewer` — checks every task in `plan.md` was implemented; cites file:line evidence
- `backend-guidelines-reviewer` — adversarial Go audit (DOM-*, SUB-*, SEC-* checks)
- `frontend-guidelines-reviewer` — adversarial TS/React audit (FE-* checks)

For ad-hoc one-off checks, invoke any agent directly by name without the orchestration skill.

## Maintenance Commands

| Command | What it does | Underlying agent |
|---|---|---|
| `/review-todos` | Whole-codebase TODO/FIXME scan; updates `docs/TODO.md` | `todo-scanner` |
| `/service-doc <service>` | Generates/updates documentation for one service | `service-documentation` |
| `/recipe-to-cooklang` | Converts recipe text to Cooklang format | (no agent — direct command) |

## Domain Skills

These activate via the project hook (`skill-activation-prompt.py`) when you mention relevant keywords or work on relevant files:

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
| TODO list | `docs/TODO.md` |
| Recipes | `recipes/` |
```

- [ ] **Step 3: Verify the file**

Run: `wc -l docs/superpowers-integration.md`

Expected: ~70+ lines.

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers-integration.md
git commit -m "docs: add superpowers-integration when-to-use-what reference"
```

---

## Task 3: Update commands/spec-task.md handoff

**Files:**
- Modify: `.claude/commands/spec-task.md`

- [ ] **Step 1: Read the current file**

Run: `cat .claude/commands/spec-task.md | tail -20`

Locate Step 5's "Suggested next step" line which currently reads: `Suggested next step (e.g., "Run \`/dev-docs task-NNN-slug\` to create an implementation plan")`.

- [ ] **Step 2: Update the handoff line**

Use the Edit tool to change the handoff suggestion in `.claude/commands/spec-task.md`.

Find:
```
2. Suggested next step (e.g., "Run `/dev-docs task-NNN-slug` to create an implementation plan")
```

Replace with:
```
2. Suggested next step: "Now run `/clear` to reset context, then `/design-task task-NNN-slug` to invoke the brainstorming/design phase"
```

- [ ] **Step 3: Verify the change**

Run: `grep -n "design-task\|dev-docs" .claude/commands/spec-task.md`

Expected: matches `design-task` reference, no `dev-docs` reference.

- [ ] **Step 4: Commit**

```bash
git add .claude/commands/spec-task.md
git commit -m "chore: update spec-task handoff to point at /design-task"
```

---

## Task 4: Create commands/design-task.md

**Files:**
- Create: `.claude/commands/design-task.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/commands/design-task.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/commands/design-task.md`:

```markdown
---
description: Phase 2 — invoke superpowers:brainstorming to produce a design doc, using an existing PRD as input context
argument-hint: Task folder name under docs/tasks/ (e.g., "task-044-superpowers-integration")
---

You are starting Phase 2 of the Home Hub four-phase development workflow. The task folder is: **$ARGUMENTS**

## Process

### Step 1 — Validate input

1. Treat `$ARGUMENTS` as a folder name under `docs/tasks/`. Resolve to `docs/tasks/$ARGUMENTS/`.
2. Confirm the folder exists. If not, stop and ask the user for the correct task folder name.
3. Confirm `prd.md` exists in that folder. If not, stop and tell the user to run `/spec-task` first or provide the missing PRD.
4. Confirm `design.md` does NOT already exist. If it does, ask the user whether to overwrite or open the existing one.

### Step 2 — Load context

Read the following:

- `docs/tasks/$ARGUMENTS/prd.md` (the requirements being designed against)
- `CLAUDE.md` (project conventions)
- `docs/superpowers-integration.md` (workflow reference)
- Relevant pieces of the codebase based on the PRD's Service Impact section

### Step 3 — Invoke the brainstorming skill

Use the Skill tool to invoke `superpowers:brainstorming`. Pass context that:

- The PRD has already been written and approved at `docs/tasks/$ARGUMENTS/prd.md`.
- The brainstorming skill should SKIP its default what/why questions (purpose, success criteria) because the PRD answers them.
- Focus questions on architecture, alternatives, and tradeoffs.
- The design output MUST be saved to `docs/tasks/$ARGUMENTS/design.md` (NOT the skill's default `docs/superpowers/specs/...` location).
- After the design is approved by the user, do NOT proceed to invoke `writing-plans`. The user will run `/clear` and then `/plan-task $ARGUMENTS` separately, in a fresh session.

### Step 4 — Save and confirm

Once the brainstorming skill has produced an approved design, ensure it is written to `docs/tasks/$ARGUMENTS/design.md`. Then tell the user:

> Design saved to `docs/tasks/$ARGUMENTS/design.md`. Now run `/clear` to reset context, then `/plan-task $ARGUMENTS` to produce the implementation plan.

## Important Rules

- DO NOT begin implementation. This phase produces a design document only.
- DO NOT auto-invoke `writing-plans`. The `/clear` between phases is intentional.
- Honor the artifact location override — the design lives in the task folder, not the superpowers default location.
```

- [ ] **Step 3: Verify the file**

Run: `head -5 .claude/commands/design-task.md && echo "---" && wc -l .claude/commands/design-task.md`

Expected: frontmatter with `description:` and `argument-hint:`, file is 40+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/commands/design-task.md
git commit -m "feat: add /design-task phase command wrapping superpowers:brainstorming"
```

---

## Task 5: Create commands/plan-task.md

**Files:**
- Create: `.claude/commands/plan-task.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/commands/plan-task.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/commands/plan-task.md`:

```markdown
---
description: Phase 3 — invoke superpowers:writing-plans to produce an implementation plan, using PRD and design as input context
argument-hint: Task folder name under docs/tasks/ (e.g., "task-044-superpowers-integration")
---

You are starting Phase 3 of the Home Hub four-phase development workflow. The task folder is: **$ARGUMENTS**

## Process

### Step 1 — Validate input

1. Resolve `docs/tasks/$ARGUMENTS/`. Confirm the folder exists.
2. Confirm both `prd.md` and `design.md` exist. If either is missing, stop and tell the user to complete the prior phase first.
3. Confirm `plan.md` does NOT already exist. If it does, ask the user whether to overwrite.

### Step 2 — Load context

Read the following:

- `docs/tasks/$ARGUMENTS/prd.md`
- `docs/tasks/$ARGUMENTS/design.md`
- `CLAUDE.md`
- `docs/superpowers-integration.md`
- The relevant code areas the design touches

### Step 3 — Invoke the writing-plans skill

Use the Skill tool to invoke `superpowers:writing-plans`. Pass context that:

- The spec is at `docs/tasks/$ARGUMENTS/design.md` (and the PRD at `prd.md` for additional reference).
- The plan output MUST be saved to `docs/tasks/$ARGUMENTS/plan.md` (NOT `docs/superpowers/plans/...`).
- Also produce a `docs/tasks/$ARGUMENTS/context.md` summarizing key files, decisions, and dependencies, useful as a quick reference for executing agents.
- After the plan is written and self-reviewed, do NOT proceed to invoke execution. The user will run `/clear` and then `/execute-task $ARGUMENTS` separately.

### Step 4 — Save and confirm

Once `plan.md` and `context.md` are written, tell the user:

> Plan and context saved to `docs/tasks/$ARGUMENTS/`. Now run `/clear` to reset context, then `/execute-task $ARGUMENTS` to begin implementation.

## Important Rules

- DO NOT begin implementation. This phase produces planning documents only.
- DO NOT auto-invoke any execution skill.
- Honor the artifact location override — plan and context live in the task folder.
- Run the `writing-plans` skill's self-review before saving (placeholder scan, type consistency, spec coverage).
```

- [ ] **Step 3: Verify the file**

Run: `head -5 .claude/commands/plan-task.md && wc -l .claude/commands/plan-task.md`

Expected: frontmatter present, 40+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/commands/plan-task.md
git commit -m "feat: add /plan-task phase command wrapping superpowers:writing-plans"
```

---

## Task 6: Create commands/execute-task.md

**Files:**
- Create: `.claude/commands/execute-task.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/commands/execute-task.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/commands/execute-task.md`:

```markdown
---
description: Phase 4 — invoke superpowers:subagent-driven-development to implement a planned task, with subagent-per-task isolation
argument-hint: Task folder name under docs/tasks/ (e.g., "task-044-superpowers-integration")
---

You are starting Phase 4 of the Home Hub four-phase development workflow. The task folder is: **$ARGUMENTS**

## Process

### Step 1 — Validate input

1. Resolve `docs/tasks/$ARGUMENTS/`. Confirm `plan.md` exists.
2. Confirm `context.md` exists alongside.
3. If either is missing, stop and tell the user to complete `/plan-task` first.

### Step 2 — Confirm execution mode

Ask the user once: subagent-driven (recommended) or inline?

- **Subagent-driven (default):** fresh subagent per task, two-stage review between tasks. Use `superpowers:subagent-driven-development`.
- **Inline:** batch execution in current session with checkpoints. Use `superpowers:executing-plans`.

If the user does not respond within the same message, default to subagent-driven.

### Step 3 — Recommend a worktree

If the current branch is `main` (or another protected branch), strongly recommend invoking `superpowers:using-git-worktrees` first to create an isolated workspace. Do NOT begin implementation on `main` without explicit user consent.

### Step 4 — Invoke the chosen execution skill

Use the Skill tool to invoke either `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Pass:

- Plan path: `docs/tasks/$ARGUMENTS/plan.md`
- Context path: `docs/tasks/$ARGUMENTS/context.md`
- Project conventions: `CLAUDE.md`

### Step 5 — On completion

After all plan tasks complete and verify, the chosen skill will hand off to `superpowers:finishing-a-development-branch`. Honor that handoff. Then suggest the user run code review:

> All plan tasks complete. Recommend running `superpowers:requesting-code-review` next, which will dispatch the appropriate reviewer agents (plan-adherence, backend-guidelines, frontend-guidelines) in parallel.

## Important Rules

- Never start implementation on `main`/`master` without explicit user consent.
- Follow plan steps exactly; stop and ask when blocked rather than guessing.
- Run the verification commands the plan specifies; don't claim completion based on assumption.
```

- [ ] **Step 3: Verify the file**

Run: `head -5 .claude/commands/execute-task.md && wc -l .claude/commands/execute-task.md`

Expected: frontmatter present, 40+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/commands/execute-task.md
git commit -m "feat: add /execute-task phase command wrapping superpowers:subagent-driven-development"
```

---

## Task 7: Create agents/plan-adherence-reviewer.md

**Files:**
- Create: `.claude/agents/plan-adherence-reviewer.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/agents/plan-adherence-reviewer.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/agents/plan-adherence-reviewer.md`:

```markdown
---
name: plan-adherence-reviewer
description: |
  Use this agent to verify that an implementation plan was faithfully executed — every task was actually implemented, nothing silently skipped or deferred. Cites file:line evidence for each task. Runs builds and tests on affected services. Produces an audit report at docs/tasks/<task-folder>/audit.md.

  <example>
  Context: All tasks in plan.md are checked off.
  user: "I think we're done with task-044. Can you verify?"
  assistant: "Let me dispatch the plan-adherence-reviewer agent to verify every task in plan.md was actually implemented."
  </example>

  <example>
  Context: After running superpowers:requesting-code-review, this agent is invoked in parallel with the guideline reviewers.
  </example>
model: inherit
---

You are an implementation plan auditor for the Home Hub project. Your job is to verify that the implementation described in a plan was faithfully executed, nothing was silently skipped or deferred, and the resulting code adheres to the project's developer guidelines.

## Input

You will be given a task folder path (e.g., `docs/tasks/task-044-superpowers-integration`). The plan to audit is at `<task-folder>/plan.md`.

## Process

### Step 1: Load the Plan

1. Read `<task-folder>/plan.md`. If `design.md` and `context.md` exist alongside, read those too for context.
2. Parse every task item (lines matching `- [ ]` or `- [x]`). Record total count, completed count, and incomplete count.

### Step 2: Determine Scope

1. From the plan and context files, identify which services/libraries were expected to be modified and which files were expected to be created or changed.
2. Use `git log` and `git diff main...HEAD` (or the appropriate base branch) to identify what was actually changed on the current branch.

### Step 3: Task Completion Audit

For each task in the plan:

1. **Check if the task was implemented.** Look for evidence in the git diff, file system, or build artifacts.
2. **Classify each task** as one of:
   - `DONE` — Evidence found that the task was completed.
   - `PARTIAL` — Some but not all acceptance criteria met.
   - `SKIPPED` — No evidence of implementation found and task is unchecked.
   - `DEFERRED` — Explicitly marked as deferred in plan or conversation.
   - `NOT_APPLICABLE` — Task is no longer relevant (explain why).
3. For `PARTIAL` or `SKIPPED` tasks, note what specifically is missing.

### Step 4: Build & Test Verification

For each affected service:

1. Run `go build ./...` (or the appropriate build command) from the service directory.
2. Run `go test ./... -count=1` from the service directory.
3. Record pass/fail status for each.

For frontend changes, run `npm run build` and `npm test` from the relevant app directory.

### Step 5: Produce Audit Report

Write the report to `<task-folder>/audit.md` (overwriting any existing audit). Format:

```markdown
# Plan Audit — <task-folder-name>

**Plan Path:** <task-folder>/plan.md
**Audit Date:** YYYY-MM-DD
**Branch:** <current branch>
**Base Branch:** main

## Executive Summary

[2–4 sentences: overall completion rate, any critical gaps, build/test status]

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|------------------|
| 1 | Description | DONE/PARTIAL/SKIPPED | file:line or commit ref |
| ... | ... | ... | ... |

**Completion Rate:** X/Y tasks (Z%)
**Skipped without approval:** [count]
**Partial implementations:** [count]

## Skipped / Deferred Tasks

[For each SKIPPED or PARTIAL task, explain what is missing and the potential impact.]

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| service-name | PASS/FAIL | PASS/FAIL | error details if any |

## Overall Assessment

- **Plan Adherence:** FULL / MOSTLY_COMPLETE / INCOMPLETE
- **Recommendation:** READY_TO_MERGE / NEEDS_FIXES / NEEDS_REVIEW

## Action Items

[Numbered list of required fixes before the plan can be considered complete.]
```

## Important Rules

- This is a read-only audit. Make NO code changes.
- Every finding must include evidence (file path, line number, git commit, or specific code reference).
- If a task's completion status is ambiguous, mark it `PARTIAL` and explain what you found vs. what was expected.
- Be thorough but fair — if a task was completed in a slightly different way than described but achieves the same goal, mark it `DONE` with a note.
```

- [ ] **Step 3: Verify the file**

Run: `head -10 .claude/agents/plan-adherence-reviewer.md && wc -l .claude/agents/plan-adherence-reviewer.md`

Expected: frontmatter with `name: plan-adherence-reviewer`, `description:`, `model: inherit`. File is 70+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/agents/plan-adherence-reviewer.md
git commit -m "feat: add plan-adherence-reviewer agent (replaces /audit-plan body)"
```

---

## Task 8: Create agents/backend-guidelines-reviewer.md

**Files:**
- Create: `.claude/agents/backend-guidelines-reviewer.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/agents/backend-guidelines-reviewer.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/agents/backend-guidelines-reviewer.md`:

```markdown
---
name: backend-guidelines-reviewer
description: |
  Use this agent to adversarially audit a Go service or changed Go packages against the Home Hub backend developer guidelines. Runs the 20-item DOM-* domain checklist, the SUB-* sub-domain checklist, and SEC-* security checks where applicable. Default mindset is FAIL until file:line evidence proves PASS. Produces audit.md and audit.json.

  <example>
  Context: A feature touched services/recipe-service.
  user: "Audit the recipe-service against backend guidelines."
  assistant: "Dispatching backend-guidelines-reviewer to run the DOM checklist on services/recipe-service."
  </example>

  <example>
  Context: superpowers:requesting-code-review detects Go file changes.
  </example>
model: inherit
---

You are an adversarial backend auditor for the Home Hub microservice platform. Your job is to find every violation. Assume every check FAILS until you find the specific line of code that proves compliance. "Looks correct" is not evidence — cite the file path and line number or it fails.

## Input

You will be given either:

- A service path (e.g., `services/auth-service`) — audit the entire service.
- A list of changed Go packages (e.g., from a `git diff` summary) — audit only those packages.

If invoked with no argument and a `plan.md` exists in the current branch's task folder, derive the audit scope from the plan's `Files:` sections.

## Mindset

- You are a skeptic, not a reviewer. Your default answer is FAIL.
- Never use phrases like "mostly compliant", "generally follows", or "appears correct".
- Every PASS requires a file:line citation. Every FAIL requires a file:line citation showing what's wrong (or noting the file/symbol is absent).
- Do not invent new rules. Only enforce what exists in the guidelines.
- Do not suggest improvements beyond what the guidelines require.

## Phase 0: Setup

1. Derive `service-name` as the last path segment of the service path (e.g., `services/auth-service` → `auth-service`).
2. Read the backend developer guidelines fully:
   - `.claude/skills/backend-dev-guidelines/resources/ai-guidance.md` (includes Commonly Missed Items Checklist)
   - `.claude/skills/backend-dev-guidelines/resources/file-responsibilities.md`
   - `.claude/skills/backend-dev-guidelines/resources/anti-patterns.md`
   - `.claude/skills/backend-dev-guidelines/resources/testing-guide.md`
   - `.claude/skills/backend-dev-guidelines/resources/patterns-provider.md`
   - `.claude/skills/backend-dev-guidelines/resources/patterns-multitenancy-context.md`
   - `.claude/skills/backend-dev-guidelines/resources/patterns-rest-jsonapi.md`
   - `.claude/skills/backend-dev-guidelines/resources/patterns-functional.md`
   - `.claude/skills/backend-dev-guidelines/resources/scaffolding-checklist.md`

## Phase 1: Build & Test (Objective Gate)

```bash
cd <service-path> && go build ./...
cd <service-path> && go test ./... -count=1
```

If either fails, the audit overall status is automatically `fail`. Record the build errors as the audit result and DO NOT proceed to Phase 2.

## Phase 2: Domain Discovery

1. List all packages under `<service-path>/internal/`.
2. For each package, classify it as:
   - **Domain package**: has `model.go` → full DOM checklist applies.
   - **Sub-domain package**: has `resource.go` but no `model.go` (action-event pattern) → SUB checklist applies.
   - **Support package**: neither → skip checklist, note its purpose.

## Phase 3: Per-Domain Mechanical Checks

For EACH domain package identified in Phase 2, run every check below. These are binary — the symbol/pattern either exists or it doesn't. Use grep/read to verify each one.

### Domain Package Checklist (every domain with `model.go`)

| ID | Check | How to Verify | Pass Criteria |
|----|-------|---------------|---------------|
| DOM-01 | `builder.go` exists | File exists in package | File present with `NewBuilder()`, fluent setters, `Build()` with validation |
| DOM-02 | `ToEntity()` method | Grep for `func (m Model) ToEntity()` or `func (m *Model) ToEntity()` in `entity.go` | Method exists on Model type |
| DOM-03 | `Make(Entity)` function | Grep for `func Make(` in `entity.go` | Function exists, returns `(Model, error)` |
| DOM-04 | `Transform` function | Grep for `func Transform(` in `rest.go` | Function exists |
| DOM-05 | `TransformSlice` function | Grep for `func TransformSlice(` in `rest.go` | Function exists, list handlers use it (no inline loops in resource.go) |
| DOM-06 | Processor accepts `FieldLogger` | Read `processor.go` constructor | Parameter type is `logrus.FieldLogger`, NOT `*logrus.Logger` |
| DOM-07 | Handlers pass `d.Logger()` | Grep `resource.go` for `NewProcessor` calls | All pass `d.Logger()`, none pass `logrus.StandardLogger()` |
| DOM-08 | POST/PATCH use `RegisterInputHandler` | Grep `resource.go` for `Methods(http.MethodPost)` and `Methods(http.MethodPatch)` | Each is registered with `RegisterInputHandler[T]`, not `RegisterHandler` |
| DOM-09 | Transform errors handled | Grep `resource.go` for `Transform(` calls | None use `_, _ :=` or `_ =` pattern; all check error |
| DOM-10 | Test DB has tenant callbacks | Read test files, find `setupTestDB` or equivalent | Calls `database.RegisterTenantCallbacks(l, db)` |
| DOM-11 | Providers use lazy evaluation | Read `provider.go` | Uses `database.Query`/`database.SliceQuery`, not eager execution wrapped in `FixedProvider` |
| DOM-12 | No `os.Getenv()` in handlers | Grep `resource.go` for `os.Getenv` | Zero matches |
| DOM-13 | No cross-domain logic in handlers | Read `resource.go` handler functions | Handlers call only their domain's processor; cross-domain orchestration is in processor layer |
| DOM-14 | Handlers don't call providers directly | Grep `resource.go` for provider function calls | Handlers call processor methods only |
| DOM-15 | No direct entity creation in handlers | Grep `resource.go` for `db.Create`, `db.Save`, `db.Delete` | Zero matches — all writes go through processor → administrator |
| DOM-16 | `administrator.go` exists for write operations | File exists if domain has create/update/delete | Write functions defined here, called by processor |
| DOM-17 | Domain error → HTTP status mapping | Read `resource.go` error handling | Validation errors → 400, not-found → 404, conflicts → 409, else → 500 |
| DOM-18 | JSON:API interface on REST models | Read `rest.go` | RestModel implements `GetName()`, `GetID()`, `SetID()` |
| DOM-19 | Request models use flat structure | Read `rest.go` | CreateRequest/UpdateRequest have no nested Data/Type/Attributes structs |
| DOM-20 | Table-driven tests | Read test files | Tests use `tests := []struct{...}` pattern with `t.Run` |

### Sub-Domain Package Checklist (action-event packages without `model.go`)

| ID | Check | How to Verify | Pass Criteria |
|----|-------|---------------|---------------|
| SUB-01 | Has processor or uses parent processor | File exists or parent processor has methods for this action | Business logic not in handler |
| SUB-02 | Has administrator for writes | `administrator.go` exists or parent administrator handles writes | No `db.Create`/`db.Save` in `resource.go` |
| SUB-03 | Uses `RegisterInputHandler[T]` for POST | Grep `resource.go` | POST endpoints use typed input handler |
| SUB-04 | No manual JSON parsing | Grep `resource.go` for `json.NewDecoder`, `json.Unmarshal`, `io.ReadAll` | Zero matches |

## Phase 4: Security Review (auth-related services only)

If the service handles authentication, authorization, or token management:

| ID | Check | How to Verify |
|----|-------|---------------|
| SEC-01 | JWT validation uses verified parsing | Grep for `ParseUnverified`, `Parse(` — ensure tokens are validated with proper key/claims |
| SEC-02 | Token revocation checks validated tokens | Read logout/revocation handlers — ensure they don't extract claims from unvalidated tokens |
| SEC-03 | No open redirect | Read callback/redirect handlers — ensure redirect URLs are validated/sanitized |
| SEC-04 | Secrets not hardcoded | Grep for hardcoded keys, passwords, secrets in source |

## Phase 5: Produce Audit Artifacts

If invoked with a single service path, write to `docs/audits/<service-name>/audit.md` and `audit.json`.

If invoked from a task folder context (i.e., changes from a feature branch), append to `docs/tasks/<task-folder>/audit.md` and `audit.json` (so the combined code review has one location per task).

### audit.md format

```markdown
# Backend Audit — <service-name>

- **Service Path:** ...
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** YYYY-MM-DD
- **Build:** PASS/FAIL
- **Tests:** X passed, Y failed
- **Overall:** PASS / NEEDS-WORK / FAIL

## Build & Test Results

[Verbatim output summary from Phase 1]

## Domain Checklist Results

### <domain-package-name>

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | internal/domain/builder.go:1 |
| DOM-02 | ToEntity() method | FAIL | No ToEntity() found in entity.go |
| ... | ... | ... | ... |

## Sub-Domain Checklist Results
[Same format per sub-domain]

## Security Review
[Same format, if applicable]

## Summary

### Blocking (must fix)
- [Bulleted list of FAIL items with IDs]

### Non-Blocking (should fix)
- [Bulleted list of WARN items with IDs]
```

### audit.json format

```json
{
  "service": "string",
  "path": "string",
  "date": "YYYY-MM-DD",
  "build": "pass | fail",
  "testsPassed": 0,
  "testsFailed": 0,
  "overallStatus": "pass | needs-work | fail",
  "domains": [
    {
      "name": "string",
      "type": "domain | sub-domain",
      "checks": [
        {
          "id": "DOM-01",
          "name": "builder.go exists",
          "status": "pass | fail | warn",
          "evidence": "file:line or absence note"
        }
      ]
    }
  ],
  "blocking": ["DOM-02: domain/entity.go missing ToEntity()"],
  "nonBlocking": []
}
```

## Rules for Status Assignment

- **PASS**: Build passes, tests pass, zero FAIL checks across all domains.
- **NEEDS-WORK**: Build and tests pass, but one or more FAIL checks exist.
- **FAIL**: Build fails, tests fail, or security checks fail.

A single FAIL check in any domain prevents overall PASS. There is no curve.
```

- [ ] **Step 3: Verify the file**

Run: `head -10 .claude/agents/backend-guidelines-reviewer.md && wc -l .claude/agents/backend-guidelines-reviewer.md`

Expected: frontmatter with `name: backend-guidelines-reviewer`, file is 150+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/agents/backend-guidelines-reviewer.md
git commit -m "feat: add backend-guidelines-reviewer agent (replaces /backend-audit body)"
```

---

## Task 9: Create agents/frontend-guidelines-reviewer.md

**Files:**
- Create: `.claude/agents/frontend-guidelines-reviewer.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/agents/frontend-guidelines-reviewer.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/agents/frontend-guidelines-reviewer.md`. The FE-* checklist is derived from the existing `frontend-dev-guidelines/resources/anti-patterns.md` and `SKILL.md` Quick Start Checklist.

```markdown
---
name: frontend-guidelines-reviewer
description: |
  Use this agent to adversarially audit a frontend area or changed TypeScript/React files against the Home Hub frontend developer guidelines. Runs the FE-* checklist covering anti-patterns, JSON:API typing, multi-tenancy, React Query usage, form/Zod validation, styling, and testing. Default mindset is FAIL until file:line evidence proves PASS.

  <example>
  Context: A feature touched frontend/src/pages and frontend/src/services.
  user: "Run frontend audit on this branch."
  assistant: "Dispatching frontend-guidelines-reviewer over the changed TS files."
  </example>

  <example>
  Context: superpowers:requesting-code-review detects TS file changes.
  </example>
model: inherit
---

You are an adversarial frontend auditor for the Home Hub UI. Your job is to find every violation. Assume every check FAILS until you find the specific line of code that proves compliance. "Looks correct" is not evidence — cite the file path and line number or it fails.

## Input

You will be given either:

- A frontend path (e.g., `frontend/src`) — audit the area.
- A list of changed TypeScript/React files (e.g., from a `git diff` summary) — audit only those.

If invoked with no argument and a `plan.md` exists in the current branch's task folder, derive the audit scope from the plan's `Files:` sections (any `.ts` / `.tsx` paths).

## Mindset

- Default answer is FAIL.
- Every PASS requires a file:line citation. Every FAIL requires a file:line citation showing what's wrong (or noting absence).
- Do not invent new rules. Enforce only what exists in the guidelines.

## Phase 0: Setup

Read the frontend developer guidelines:

- `.claude/skills/frontend-dev-guidelines/SKILL.md`
- `.claude/skills/frontend-dev-guidelines/resources/anti-patterns.md`
- `.claude/skills/frontend-dev-guidelines/resources/architecture-overview.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-react-query.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-forms-validation.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-service-layer.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-types.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-multitenancy.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-styling.md`
- `.claude/skills/frontend-dev-guidelines/resources/patterns-components.md`
- `.claude/skills/frontend-dev-guidelines/resources/testing-guide.md`

## Phase 1: Build & Test (Objective Gate)

```bash
cd frontend && npm run build
cd frontend && npm test -- --watchAll=false
```

If either fails, the audit overall status is automatically `fail`. Record the errors and stop.

## Phase 2: File Inventory

List all changed/in-scope files. Classify each as:

- **Page** (`pages/*.tsx`)
- **Component** (`components/**/*.tsx`)
- **Hook** (`lib/hooks/api/*.ts`)
- **Service** (`services/api/*.ts`)
- **Schema** (`lib/schemas/*.ts`)
- **Type** (`types/**/*.ts`)
- **Other**

## Phase 3: Mechanical Checks

For each in-scope file, run every applicable check.

### Anti-Pattern Checklist

| ID | Check | How to Verify | Pass Criteria |
|----|-------|---------------|---------------|
| FE-01 | No `any` type | Grep file for `: any` and `as any` | Zero matches (excluding `null as any` cast workarounds — those are also fails) |
| FE-02 | No manual class concatenation | Grep for `className={"` followed by `+` or template-string concatenation | Zero matches; `cn()` used instead |
| FE-03 | No direct API client calls in components | Grep components/pages for `import .* from "@/lib/api/client"` | Zero matches; service layer used |
| FE-04 | No inline Zod schemas in components | Grep components for `z.object(`, `z.string(`, etc. | Zero matches except `.refine()` cross-field validations |
| FE-05 | No spinners for content loading | Grep for `animate-spin` | Allowed only on submit buttons; content uses Skeleton |
| FE-06 | No hardcoded colors | Grep for class names matching `bg-(white|black|gray-\d|red-\d|...)` | Zero matches; semantic classes (`bg-background`, etc.) used |
| FE-07 | No state mutation | Grep for `\.push(`, `\.splice(`, `\.sort(` followed by setState | Zero matches; immutable updates only |
| FE-08 | No default exports for components | Grep for `export default function` in component files | Zero matches; named exports only |
| FE-09 | Tenant guard in hooks | Read each hook in `lib/hooks/api/` | Each hook either takes explicit `tenant` parameter or uses `useTenant()`; query hooks have `enabled: !!tenant?.id` |
| FE-10 | Tenant ID in query keys | Read query key factories | All keys for tenant-scoped resources include `tenant?.id \|\| 'no-tenant'` |
| FE-11 | Error handling with `createErrorFromUnknown` | Grep for `.catch(` in async operations | Each catch uses `createErrorFromUnknown()` and surfaces via toast or error state |

### Architecture Checklist

| ID | Check | How to Verify | Pass Criteria |
|----|-------|---------------|---------------|
| FE-12 | JSON:API model shape | Read `types/models/` files | All models follow `{ id: string, attributes: {...} }` structure |
| FE-13 | Service extends `BaseService` (when applicable) | Read `services/api/` files | Concrete services extend `BaseService` or use the documented direct-client pattern |
| FE-14 | Query key factory uses `as const` | Read query key factories | Keys are `[...] as const` |
| FE-15 | Forms use `react-hook-form` + `zodResolver` | Read form components | `useForm({ resolver: zodResolver(schema) })` pattern |
| FE-16 | Schema in `lib/schemas/` with inferred type | Read schema files | Each `z.object(...)` is paired with `export type X = z.infer<typeof schema>` |

### Testing Checklist

| ID | Check | How to Verify | Pass Criteria |
|----|-------|---------------|---------------|
| FE-17 | Tests exist for changed components | List test files matching changed components | At least one test file per non-trivial component change |
| FE-18 | Mocks updated when services changed | Diff `__mocks__/` against changed services | Mocks reflect interface changes |

## Phase 4: Produce Audit Artifacts

If invoked from a task folder context, append to `docs/tasks/<task-folder>/audit.md` (so the combined review lives in one file).

If invoked standalone, write to `docs/audits/frontend/audit.md`.

### audit.md format

```markdown
# Frontend Audit — <area or branch>

- **Audit Scope:** ...
- **Guidelines Source:** frontend-dev-guidelines skill
- **Date:** YYYY-MM-DD
- **Build:** PASS/FAIL
- **Tests:** X passed, Y failed
- **Overall:** PASS / NEEDS-WORK / FAIL

## Build & Test Results

[Verbatim summary]

## File Inventory

[Bulleted list of files audited and their classification]

## Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | (no matches found) |
| FE-02 | No manual class concatenation | FAIL | components/foo.tsx:34 |
| ... | ... | ... | ... |

## Architecture Checklist
[Same format]

## Testing Checklist
[Same format]

## Summary

### Blocking (must fix)
- [Bulleted list of FAIL items with IDs]

### Non-Blocking (should fix)
- [Bulleted list of WARN items with IDs]
```

## Rules for Status Assignment

- **PASS**: Build passes, tests pass, zero FAIL checks.
- **NEEDS-WORK**: Build and tests pass, but FAIL checks exist.
- **FAIL**: Build fails or tests fail.

A single FAIL check prevents overall PASS.
```

- [ ] **Step 3: Verify the file**

Run: `head -10 .claude/agents/frontend-guidelines-reviewer.md && wc -l .claude/agents/frontend-guidelines-reviewer.md`

Expected: frontmatter with `name: frontend-guidelines-reviewer`, file is 100+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/agents/frontend-guidelines-reviewer.md
git commit -m "feat: add frontend-guidelines-reviewer agent for TS/React audits"
```

---

## Task 10: Shrink commands/audit-plan.md to wrapper

**Files:**
- Modify: `.claude/commands/audit-plan.md`

- [ ] **Step 1: Replace the file content**

Use the Write tool to overwrite `.claude/commands/audit-plan.md` with:

```markdown
---
description: Verify a plan was faithfully implemented — dispatches the plan-adherence-reviewer agent
argument-hint: Task folder name under docs/tasks/ (e.g., "task-044-superpowers-integration")
---

Dispatch the `plan-adherence-reviewer` agent against the task folder: **$ARGUMENTS**.

Pass the task folder path so the agent can locate `plan.md`, run the audit, and write findings to `docs/tasks/$ARGUMENTS/audit.md`.

After the agent completes, summarize the findings to the user — completion rate, blocking issues, and recommended next steps.
```

- [ ] **Step 2: Verify the change**

Run: `wc -l .claude/commands/audit-plan.md && cat .claude/commands/audit-plan.md`

Expected: file is 8–12 lines, references `plan-adherence-reviewer`, no longer contains the full audit instructions.

- [ ] **Step 3: Commit**

```bash
git add .claude/commands/audit-plan.md
git commit -m "refactor: shrink /audit-plan to thin wrapper around plan-adherence-reviewer agent"
```

---

## Task 11: Shrink commands/backend-audit.md to wrapper

**Files:**
- Modify: `.claude/commands/backend-audit.md`

- [ ] **Step 1: Replace the file content**

Overwrite `.claude/commands/backend-audit.md` with:

```markdown
---
description: Adversarially audit a Go service against backend developer guidelines — dispatches the backend-guidelines-reviewer agent
argument-hint: Path to the service to audit (e.g., "services/auth-service")
---

Dispatch the `backend-guidelines-reviewer` agent against: **$ARGUMENTS**.

Pass the service path so the agent can run the build/test gate, the DOM-* / SUB-* checklists, and (if auth-related) SEC-* checks. The agent writes `audit.md` and `audit.json` under `docs/audits/<service-name>/` (or under the active task folder if invoked from a feature branch with a `plan.md`).

After the agent completes, summarize PASS / NEEDS-WORK / FAIL status and any blocking items.
```

- [ ] **Step 2: Verify the change**

Run: `wc -l .claude/commands/backend-audit.md && cat .claude/commands/backend-audit.md`

Expected: file is 8–12 lines, references `backend-guidelines-reviewer`.

- [ ] **Step 3: Commit**

```bash
git add .claude/commands/backend-audit.md
git commit -m "refactor: shrink /backend-audit to thin wrapper around backend-guidelines-reviewer agent"
```

---

## Task 12: Create agents/todo-scanner.md

**Files:**
- Create: `.claude/agents/todo-scanner.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/agents/todo-scanner.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/agents/todo-scanner.md`:

```markdown
---
name: todo-scanner
description: |
  Use this agent to scan the entire Home Hub codebase for TODO/FIXME/XXX/HACK markers and unimplemented stubs, categorize by service, prioritize, and update docs/TODO.md. Heavy file-scanning work isolated from the main agent's context window.

  <example>
  Context: User wants a fresh inventory of incomplete work.
  user: "/review-todos"
  assistant: "Dispatching todo-scanner to refresh docs/TODO.md."
  </example>

  <example>
  Context: After landing a large feature, the user wants to know what TODO debt was left behind.
  user: "What TODOs are left in recipe-service?"
  assistant: "Let me dispatch todo-scanner to give you a current inventory."
  </example>
model: inherit
---

You are a codebase analyst performing a comprehensive review of the Home Hub project to identify incomplete work.

## Process

### Phase 1: Discovery (Run in Parallel)

Launch three parallel exploration tasks:

1. **Find all TODO markers**
   - Search the entire codebase for: TODO, FIXME, XXX, HACK, and similar markers.
   - For each finding, note: file path, line number, content, and surrounding context.
   - Check all file types: Go, TypeScript, JavaScript, JSON, YAML, etc.

2. **Find unimplemented/stub code**
   - Search for patterns indicating incomplete implementations:
     - Functions returning nil/null without doing work
     - Empty function bodies or placeholder implementations
     - "not implemented" / "NotImplemented" strings/errors
     - Panic statements with "not implemented" messages
     - Functions that log warnings about missing implementation
     - Commented-out code blocks that may indicate incomplete work
   - Focus on Go files but also check TypeScript/JavaScript.
   - Note the file, function name, and what appears to be missing.

3. **Analyze project structure**
   - Identify the different domains/services in the codebase.
   - Understand service directory structure and purposes.
   - This enables organizing findings by domain.

### Phase 2: Analysis

After discovery completes:

1. **Categorize findings by domain/service**
   - Group all TODOs and incomplete implementations by the service they belong to.
   - Identify cross-cutting concerns that affect multiple services.

2. **Prioritize findings**
   - **Critical**: Core service functionality broken or missing.
   - **High Priority**: Features incomplete but not blocking basic functionality.
   - **Medium Priority**: Quality/polish issues, performance optimizations.
   - **Low Priority**: Minor cosmetic issues, documentation.

3. **Identify patterns**
   - Note areas with concentrated incomplete work.
   - Identify systemic issues vs. one-off TODOs.

### Phase 3: Update docs/TODO.md

Read the existing `docs/TODO.md` (if any), then update it with the comprehensive findings.

```markdown
# Home Hub Project TODO

This document tracks planned features and improvements for the Home Hub project.

---

## Priority Summary

### Critical
- [ ] **Item** — Brief description

### High Priority
- [ ] **Item** — Brief description

---

## Services

### Service Name
- [ ] Description of TODO (`file/path.go:123`)
- [ ] Description of TODO (`file/path.go:456`)

[Continue for all services alphabetically]

---

## Libraries

### library-name
- [ ] Description (`file/path.go:123`)

---

## Notes

- Summary statistics
- Important context
```

**Guidelines for updating:**
- Preserve any manually-curated items that aren't from inline TODOs.
- Include file path and line number for each inline TODO.
- Use bold for critical/blocking items.
- Group related items under subsections when a service has many items.
- Keep descriptions concise but informative.
- Add summary statistics at the end (total TODOs, most concentrated areas).

### Output

After updating the file, return a brief summary to the caller:
- Total number of TODOs/incomplete items found
- Top 3–5 most critical items
- Services with the most incomplete work
- Any new items added since the last review (if determinable)
```

- [ ] **Step 3: Verify the file**

Run: `head -10 .claude/agents/todo-scanner.md && wc -l .claude/agents/todo-scanner.md`

Expected: frontmatter with `name: todo-scanner`, 80+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/agents/todo-scanner.md
git commit -m "feat: add todo-scanner agent (replaces /review-todos body)"
```

---

## Task 13: Shrink commands/review-todos.md to wrapper

**Files:**
- Modify: `.claude/commands/review-todos.md`

- [ ] **Step 1: Replace the file content**

Overwrite `.claude/commands/review-todos.md` with:

```markdown
---
description: Scan the codebase for TODO/FIXME markers and unimplemented stubs; updates docs/TODO.md — dispatches the todo-scanner agent
---

Dispatch the `todo-scanner` agent.

The agent runs a full-repo scan, categorizes findings by service and priority, and updates `docs/TODO.md`. After it completes, surface the summary it returns: total findings, top critical items, and services with the most concentrated incomplete work.
```

- [ ] **Step 2: Verify the change**

Run: `wc -l .claude/commands/review-todos.md && cat .claude/commands/review-todos.md`

Expected: file is 6–10 lines, references `todo-scanner`.

- [ ] **Step 3: Commit**

```bash
git add .claude/commands/review-todos.md
git commit -m "refactor: shrink /review-todos to thin wrapper around todo-scanner agent"
```

---

## Task 14: Create agents/service-documentation.md

**Files:**
- Create: `.claude/agents/service-documentation.md`

- [ ] **Step 1: Confirm file does not exist**

Run: `ls .claude/agents/service-documentation.md 2>&1 || echo "absent"`

Expected: `absent`.

- [ ] **Step 2: Create the file**

Write the following content to `.claude/agents/service-documentation.md`:

```markdown
---
name: service-documentation
description: |
  Use this agent to generate or update documentation for one specific Home Hub service. Treats code as the single source of truth, follows DOCS.md and CLAUDE.md, makes no inferences about future behavior. Operates only within the target service directory.

  <example>
  Context: User wants to refresh docs for a service after a feature landed.
  user: "/service-doc auth-service"
  assistant: "Dispatching service-documentation agent against services/auth-service."
  </example>

  <example>
  Context: After a large refactor of recipe-service.
  user: "Re-document recipe-service from the current code."
  assistant: "Dispatching service-documentation agent."
  </example>
model: inherit
---

You are the Home Hub Documentation Agent.

## Authoritative Inputs

- `CLAUDE.md` (architecture and coding conventions)
- `DOCS.md` (documentation contract — required structure for service docs)
- The source code for the target service

## Strict Rules

You MUST:
- Follow `DOCS.md` exactly.
- Treat code as the single source of truth.
- Document only what exists in code.
- Preserve existing documentation structure and tone.
- Ask before adding new sections or files.
- Use precise, factual language.

You MUST NOT:
- Infer intent or future behavior.
- Improve, refactor, or rationalize design.
- Propose alternatives or enhancements.
- Merge documentation concerns across services.
- Modify code.

## Task

Generate or update documentation for the service specified in the invocation argument.

Argument shape: either a service name (`auth-service`) or a service path (`services/auth-service`). Resolve to the path under `services/`.

## Scope

- Operate only within the target service directory.
- Create missing required documentation files if necessary (per `DOCS.md`).
- Update existing documentation to match current code.

## Output

- Updated documentation files only.
- No commentary, no analysis, no recommendations.
- If a required doc file cannot be produced from the available code, ask a single targeted question and stop.
```

- [ ] **Step 3: Verify the file**

Run: `head -10 .claude/agents/service-documentation.md && wc -l .claude/agents/service-documentation.md`

Expected: frontmatter with `name: service-documentation`, 40+ lines.

- [ ] **Step 4: Commit**

```bash
git add .claude/agents/service-documentation.md
git commit -m "feat: add service-documentation agent (proper frontmatter, replaces /service-doc body)"
```

---

## Task 15: Shrink commands/service-doc.md to wrapper

**Files:**
- Modify: `.claude/commands/service-doc.md`

- [ ] **Step 1: Replace the file content**

Overwrite `.claude/commands/service-doc.md` with:

```markdown
---
description: Generate or update documentation for one Home Hub service — dispatches the service-documentation agent
argument-hint: Service name or path (e.g., "auth-service" or "services/auth-service")
---

Dispatch the `service-documentation` agent against: **$ARGUMENTS**.

The agent treats code as the single source of truth, follows `DOCS.md`, and operates only within the target service directory. It outputs only updated doc files — no commentary, no analysis.
```

- [ ] **Step 2: Verify the change**

Run: `wc -l .claude/commands/service-doc.md && cat .claude/commands/service-doc.md`

Expected: file is 8–12 lines, references `service-documentation`.

- [ ] **Step 3: Commit**

```bash
git add .claude/commands/service-doc.md
git commit -m "refactor: shrink /service-doc to thin wrapper around service-documentation agent"
```

---

## Task 16: Add frontend-dev-guidelines entry to skill-rules.json

**Files:**
- Modify: `.claude/skills/skill-rules.json`

- [ ] **Step 1: Read current file**

Run: `cat .claude/skills/skill-rules.json`

Confirm only `backend-dev-guidelines` is present under `skills`.

- [ ] **Step 2: Replace the file content**

Overwrite `.claude/skills/skill-rules.json` with the following (existing backend entry preserved verbatim, new frontend entry added):

```json
{
  "version": "1.0",
  "description": "Skill activation triggers for Claude Code. Controls when skills automatically suggest or block actions.",
  "skills": {
    "backend-dev-guidelines": {
      "type": "domain",
      "enforcement": "suggest",
      "priority": "high",
      "description": "Backend development patterns for Go",
      "promptTriggers": {
        "keywords": [
          "backend",
          "backend development",
          "microservice",
          "API",
          "endpoint",

          "middleware"
        ],

        "intentPatterns": [
          "(create|add|implement|build).*?(route|endpoint|API|controller|service|repository)",
          "(fix|handle|debug).*?(error|exception|backend)",

          "(add|implement).*?(middleware|validation|error.*?handling)",
          "(organize|structure|refactor).*?(backend|service|API)",
          "(how to|best practice).*?(backend|route|controller|service)"
        ]
      },
      "fileTriggers": {

        "pathPatterns": [],
        "pathExclusions": [],
        "contentPatterns": []
      }
    },
    "frontend-dev-guidelines": {
      "type": "domain",
      "enforcement": "suggest",
      "priority": "high",
      "description": "Frontend development patterns for React/TypeScript",
      "promptTriggers": {
        "keywords": [
          "frontend",
          "react",
          "tsx",
          "component",
          "hook",
          "form",
          "tailwind",
          "shadcn",
          "react query",
          "tanstack",
          "zod"
        ],
        "intentPatterns": [
          "(create|add|implement|build).*?(component|page|hook|form|dialog|table)",
          "(fix|handle|debug).*?(render|hydration|state|hook)",
          "(add|implement).*?(validation|schema|zod|form)",
          "(organize|structure|refactor).*?(frontend|component|hook|page)",
          "(how to|best practice).*?(frontend|component|hook|react|typescript)"
        ]
      },
      "fileTriggers": {
        "pathPatterns": [
          "frontend/**/*.ts",
          "frontend/**/*.tsx",
          "**/components/**/*.tsx",
          "**/pages/**/*.tsx",
          "**/lib/hooks/**/*.ts",
          "**/lib/schemas/**/*.ts",
          "**/services/api/**/*.ts"
        ],
        "pathExclusions": [
          "**/*.test.ts",
          "**/*.test.tsx"
        ],
        "contentPatterns": []
      }
    }
  },
  "notes": {
    "enforcement_types": {
      "suggest": "Skill suggestion appears but doesn't block execution",
      "block": "Requires skill to be used before proceeding (guardrail)",
      "warn": "Shows warning but allows proceeding"
    },
    "priority_levels": {
      "critical": "Highest - Always trigger when matched",
      "high": "Important - Trigger for most matches",
      "medium": "Moderate - Trigger for clear matches",
      "low": "Optional - Trigger only for explicit matches"
    },
    "customization": {
      "pathPatterns": "Adjust to match your project structure (blog-api, auth-service, etc.)",

      "keywords": "Add domain-specific terms relevant to your project",

      "intentPatterns": "Use regex for flexible user intent matching"
    }

  }
}
```

- [ ] **Step 3: Validate JSON**

Run: `python3 -m json.tool < .claude/skills/skill-rules.json > /dev/null && echo "valid JSON"`

Expected: `valid JSON`. If error, fix syntax and re-run.

- [ ] **Step 4: Smoke test the hook with a frontend prompt**

Run:
```bash
cd /home/tumidanski/source/pers/home-hub && \
echo '{"prompt": "I want to add a new React component for the dashboard"}' | \
CLAUDE_PROJECT_DIR=$(pwd) python3 .claude/hooks/skill-activation-prompt.py
```

Expected output includes:
```
🎯 SKILL ACTIVATION CHECK
...
📚 RECOMMENDED SKILLS:
  → frontend-dev-guidelines
```

- [ ] **Step 5: Smoke test the hook with a backend prompt (regression check)**

Run:
```bash
cd /home/tumidanski/source/pers/home-hub && \
echo '{"prompt": "I need to add a new endpoint to the auth-service"}' | \
CLAUDE_PROJECT_DIR=$(pwd) python3 .claude/hooks/skill-activation-prompt.py
```

Expected output includes `→ backend-dev-guidelines` (backend rule still works).

- [ ] **Step 6: Commit**

```bash
git add .claude/skills/skill-rules.json
git commit -m "chore: add frontend-dev-guidelines entry to skill-rules.json"
```

---

## Task 17: Remove commands/dev-docs.md

**Files:**
- Delete: `.claude/commands/dev-docs.md`

- [ ] **Step 1: Confirm no remaining references in the repo**

Run: `grep -rn "dev-docs" --include="*.md" --include="*.json" --include="*.sh" --include="*.py" .`

Expected: matches only inside `docs/tasks/task-044-superpowers-integration/` (this plan and design) and any prior task folders that historically referenced the old command. No matches inside `.claude/` outside of `.claude/commands/dev-docs.md` itself.

If unexpected references exist in active commands or agents, update them before deleting.

- [ ] **Step 2: Delete the file**

Run: `rm .claude/commands/dev-docs.md`

- [ ] **Step 3: Verify deletion**

Run: `ls .claude/commands/dev-docs.md 2>&1 || echo "deleted"`

Expected: `deleted` (or `No such file`).

- [ ] **Step 4: Commit**

```bash
git add -A .claude/commands/
git commit -m "chore: remove /dev-docs command (replaced by /design-task + /plan-task)"
```

---

## Task 18: Remove agents/documentation.md

**Files:**
- Delete: `.claude/agents/documentation.md`

- [ ] **Step 1: Confirm no remaining references in the repo**

Run: `grep -rn "agents/documentation\b\|^.*documentation\.md" --include="*.md" .claude/ docs/ CLAUDE.md`

Expected: no matches in active commands or agents (the new `service-doc.md` wrapper references `service-documentation`, not `documentation`). Some references in `docs/tasks/task-044-superpowers-integration/` are expected (this plan and design).

- [ ] **Step 2: Delete the file**

Run: `rm .claude/agents/documentation.md`

- [ ] **Step 3: Verify deletion**

Run: `ls .claude/agents/documentation.md 2>&1 || echo "deleted"`

Expected: `deleted`.

- [ ] **Step 4: Commit**

```bash
git add -A .claude/agents/
git commit -m "chore: remove agents/documentation.md (renamed to service-documentation.md)"
```

---

## Task 19: End-to-end smoke test

**Files:**
- Create: `docs/tasks/task-099-smoke-test/prd.md` (temporary, to be deleted)

This task verifies the new workflow produces artifacts in the expected locations.

- [ ] **Step 1: Create a throwaway PRD by hand**

Skip the interactive `/spec-task` invocation; manually create the smoke-test folder and a tiny PRD to feed `/design-task`.

```bash
mkdir -p docs/tasks/task-099-smoke-test
```

Write the following to `docs/tasks/task-099-smoke-test/prd.md`:

```markdown
# Smoke Test — Product Requirements Document

Version: v1
Status: Draft
Created: YYYY-MM-DD

## 1. Overview

Throwaway PRD used to verify the four-phase workflow produces artifacts in `docs/tasks/<task-folder>/`. This folder is deleted at the end of the smoke test.

## 2. Goals

- Verify `/design-task` writes `design.md` into the task folder.
- Verify `/plan-task` writes `plan.md` and `context.md` into the task folder.

Non-goals:
- Verify execution.

## 3. Acceptance Criteria

- After running `/design-task task-099-smoke-test`, `docs/tasks/task-099-smoke-test/design.md` exists.
- After running `/plan-task task-099-smoke-test`, `docs/tasks/task-099-smoke-test/plan.md` and `context.md` exist.
```

- [ ] **Step 2: Verify command files are discoverable**

Run: `ls .claude/commands/ | sort`

Expected output includes (alphabetically): `audit-plan.md`, `backend-audit.md`, `design-task.md`, `execute-task.md`, `plan-task.md`, `recipe-to-cooklang.md`, `review-todos.md`, `service-doc.md`, `spec-task.md`. Should NOT include `dev-docs.md`.

- [ ] **Step 3: Verify agent files are discoverable**

Run: `ls .claude/agents/ | sort`

Expected output (alphabetically): `backend-guidelines-reviewer.md`, `frontend-guidelines-reviewer.md`, `plan-adherence-reviewer.md`, `service-documentation.md`, `todo-scanner.md`. Should NOT include `documentation.md`.

- [ ] **Step 4: Manual session smoke test**

In a fresh `/clear`'d session, the user invokes `/design-task task-099-smoke-test` and confirms:

a. The command starts and validates the input (PRD exists).
b. The brainstorming skill is invoked.
c. After interactive Q&A and approval, `docs/tasks/task-099-smoke-test/design.md` is written.
d. The handoff message instructs the user to `/clear` and run `/plan-task task-099-smoke-test`.

Then in another fresh session, the user invokes `/plan-task task-099-smoke-test` and confirms:

a. Both `prd.md` and `design.md` are detected.
b. The writing-plans skill is invoked.
c. `docs/tasks/task-099-smoke-test/plan.md` and `context.md` are written.

This step is performed manually by the user — the executing agent should pause and ask the user to perform the smoke test sessions, then continue once they confirm success.

- [ ] **Step 5: Clean up the smoke test folder**

After the user confirms success:

```bash
rm -rf docs/tasks/task-099-smoke-test
```

- [ ] **Step 6: Commit (only if smoke test passed)**

If any artifacts were committed during the smoke test, remove them with:

```bash
git status docs/tasks/task-099-smoke-test
# if files staged, unstage and clean
```

The smoke-test folder should NOT end up in git history. If it accidentally was committed during the manual phase, revert those commits before this point.

No commit is required for this verification task — it is a pure check.

---

## Self-Review Notes

This plan was self-reviewed against the design (`design.md`) using the writing-plans skill's checklist:

1. **Spec coverage** — every section of the design has at least one task:
   - Workflow architecture → Tasks 3, 4, 5, 6 (commands)
   - Code review path → Tasks 7, 8, 9, 10, 11
   - Maintenance commands → Tasks 12, 13, 14, 15
   - Hook + skill-rules.json → Task 16
   - Documentation → Tasks 1, 2
   - Removals → Tasks 17, 18
   - Acceptance criteria → Task 19 (smoke test)

2. **Placeholder scan** — no TBD, TODO, "implement later", or "fill in details" in any step. Every code/text content step shows the exact content.

3. **Type/name consistency** — agent names used in slash command wrappers match the agent file `name:` frontmatter:
   - `plan-adherence-reviewer` (Task 7 / Task 10)
   - `backend-guidelines-reviewer` (Task 8 / Task 11)
   - `frontend-guidelines-reviewer` (Task 9, used by `superpowers:requesting-code-review`)
   - `todo-scanner` (Task 12 / Task 13)
   - `service-documentation` (Task 14 / Task 15)
