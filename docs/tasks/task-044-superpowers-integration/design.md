---
title: Superpowers Plugin Integration
status: Draft
created: 2026-04-21
---

# Superpowers Plugin Integration — Design

## Goal

Make the `superpowers` plugin (v5.0.7) the default development workflow for Home Hub while preserving the project's existing artifact discipline (numbered task folders, PRD-style requirements docs) and project-specific tooling (backend/frontend guidelines, audits, maintenance commands).

## Background

The project ships its own `.claude/` setup:

- **Commands:** `spec-task`, `dev-docs`, `audit-plan`, `backend-audit`, `review-todos`, `service-doc`, `recipe-to-cooklang`.
- **Skills:** `backend-dev-guidelines`, `frontend-dev-guidelines`.
- **Agents:** `documentation` (no frontmatter — currently a prompt template, not a real agent).
- **Hook:** `UserPromptSubmit` runs `skill-activation-prompt.py`, which reads `skill-rules.json` and emits a banner suggesting domain skills. Today only `backend-dev-guidelines` is wired in.

The `superpowers` plugin (5.0.7) provides:

- Skills for the full lifecycle: `brainstorming`, `writing-plans`, `executing-plans`, `subagent-driven-development`, `systematic-debugging`, `test-driven-development`, `verification-before-completion`, `using-git-worktrees`, `finishing-a-development-branch`, `requesting-code-review`, `receiving-code-review`, `dispatching-parallel-agents`, `writing-skills`, `using-superpowers`.
- A generic `code-reviewer` agent.
- Three deprecated slash commands (`/brainstorm`, `/write-plan`, `/execute-plan`) that just point users at the corresponding skills.

Two parallel sets of artifacts and conventions exist. This design merges them.

## Decisions

| # | Decision |
|---|---|
| D1 | Default workflow becomes the four-phase pipeline below, backed by superpowers skills. |
| D2 | Artifact home stays at `docs/tasks/task-NNN-slug/`. PRD shape is preserved. |
| D3 | Each phase is its own slash command, designed for invocation from a `/clear`'d session. No auto-handoff between phases. |
| D4 | Code review becomes three modular reviewer agents dispatched in parallel. |
| D5 | `review-todos` and `service-doc` are promoted to proper agents; slash commands stay as thin wrappers. |
| D6 | Hook + `skill-rules.json` keep their narrow domain-skill role. Superpowers skills self-activate. `frontend-dev-guidelines` is added to the rules file. |
| D7 | `recipe-to-cooklang` and the backend/frontend guideline skills' resources are untouched. |

## Workflow Architecture

Four explicit phases. Each command is invoked by the user from a fresh session.

```
/spec-task <idea>            → docs/tasks/task-NNN-slug/prd.md
/clear
/design-task <task-folder>   → docs/tasks/task-NNN-slug/design.md
/clear
/plan-task <task-folder>     → docs/tasks/task-NNN-slug/plan.md + context.md
/clear
/execute-task <task-folder>  → code + commits
```

### Phase responsibilities

- **`/spec-task <idea>`** — interactive PRD interview using the existing template (10 sections: overview, goals, user stories, functional requirements, API surface, data model, service impact, NFR, open questions, acceptance criteria). Output: `prd.md`.
- **`/design-task <task-folder>`** — loads `prd.md` as input context, then invokes `superpowers:brainstorming` for architecture, alternatives, and tradeoffs. Skips the brainstorming skill's default what/why questions because the PRD already answers them. Output: `design.md` in the same folder.
- **`/plan-task <task-folder>`** — loads `prd.md` + `design.md`, invokes `superpowers:writing-plans` to produce bite-sized TDD steps with exact file paths, code blocks, and commit boundaries. Output: `plan.md` + `context.md`.
- **`/execute-task <task-folder>`** — invokes `superpowers:subagent-driven-development` (default) or `superpowers:executing-plans` (fallback when subagents are unavailable). Each task runs in its own subagent context, with review checkpoints between tasks.

### Why explicit phases (no auto-chain)

A single-conversation auto-chain would accumulate the entire PRD interview, design Q&A, and planning conversation in one window. Risks:

- **Anchoring drift** — assumptions made during the PRD interview that didn't make it into the written PRD still bias later phases.
- **Auto-compaction loss** — as context grows past limits, summarization loses nuance the artifacts were supposed to capture cleanly.
- **Signal-to-noise** — the model reasons better when its only inputs are the deliberate artifacts, not the conversation that produced them.

The `/clear` discipline keeps each phase honest: it can only act on what's documented.

### Folder convention

`docs/tasks/task-NNN-slug/` is preserved (44 existing folders). Project `CLAUDE.md` adds an explicit override for superpowers' default spec/plan locations — both `brainstorming` and `writing-plans` skills explicitly state that user preferences for locations override their defaults.

Files per task:

- `prd.md` — what & why
- `design.md` — how & tradeoffs
- `plan.md` — bite-sized TDD steps
- `context.md` — key files, decisions, dependencies (quick reference for executing agents)
- (`audit.md`, `audit.json` — produced by code-review agents when run on this task's branch)

## Code Review Path

Three modular reviewer agents, dispatched in parallel via `superpowers:dispatching-parallel-agents`. Each invokable individually for ad-hoc checks.

### Agents

- **`agents/plan-adherence-reviewer.md`** — replaces `/audit-plan`. Walks task checkboxes against `git diff main...HEAD`; classifies each task as DONE / PARTIAL / SKIPPED / DEFERRED / NOT_APPLICABLE; cites file:line evidence; runs builds & tests on affected services; produces `audit.md` in the task folder.
- **`agents/backend-guidelines-reviewer.md`** — replaces `/backend-audit`. Runs the 20-item DOM-* checklist + sub-domain SUB-* checklist + security SEC-* checklist on changed Go packages. Adversarial mindset: default FAIL until proven otherwise, every PASS requires a file:line citation. Produces `audit.md` + `audit.json` (the JSON output is preserved for compliance tracking over time).
- **`agents/frontend-guidelines-reviewer.md`** — new. Mirrors backend audit for TypeScript/React using `frontend-dev-guidelines/resources/`. Checklist derived from the existing `anti-patterns.md` and pattern files.

### Orchestration

The user invokes `superpowers:requesting-code-review` (or a thin project wrapper). The orchestration layer:

1. Inspects `git diff` to determine which file types changed.
2. Dispatches whichever reviewers apply in parallel:
   - Always: `plan-adherence-reviewer` (if a `plan.md` exists in the task folder).
   - If Go files changed: `backend-guidelines-reviewer`.
   - If TS/React files changed: `frontend-guidelines-reviewer`.
3. Aggregates findings into a single combined report.

### Slash command updates

- `commands/audit-plan.md` → one-line wrapper that dispatches `plan-adherence-reviewer`.
- `commands/backend-audit.md` → one-line wrapper that dispatches `backend-guidelines-reviewer`.

## Maintenance Commands → Agents

### `agents/todo-scanner.md` (new)

Replaces the body of `/review-todos`. Whole-codebase scan for TODO / FIXME / XXX / HACK markers and unimplemented stubs; prioritization; updates `docs/TODO.md`. Heavy file reading isolated from the main agent's context window.

### `agents/service-documentation.md` (renamed from `documentation.md`)

The current `agents/documentation.md` lacks frontmatter, so it isn't a real Claude Code agent — it's a prompt template that overlaps with `/service-doc`. Fix: add proper frontmatter (`name`, `description`, `tools`, `model: inherit`), absorb the `/service-doc` body, rename for clarity. Generates or updates documentation for one service per invocation, following `DOCS.md` and `CLAUDE.md`.

### Slash command wrappers

`commands/review-todos.md` and `commands/service-doc.md` shrink to one-liners that dispatch the corresponding agent. Kept for `/` autocomplete discoverability.

## Hook + skill-rules.json

The hook stays. It does precise keyword/intent regex matching that's harder to express in a skill's `description` field, and produces a visible `🎯 SKILL ACTIVATION CHECK` banner that confirms a domain skill applies.

Changes:

- Add `frontend-dev-guidelines` entry to `skill-rules.json` (currently orphaned — the skill exists but the hook doesn't reference it). Mirror the backend entry's structure: keywords (`frontend`, `react`, `component`, `form`, `ui`, etc.) + intent patterns matching React/TS work.
- No entries added for superpowers skills. They self-activate via native skill matching plus the `using-superpowers` discipline that runs at conversation start.

Two non-overlapping suggestion mechanisms:

- **Hook:** project-specific domain skills only.
- **Native activation:** all superpowers skills.

## Documentation

- **`CLAUDE.md`** — add the four-phase workflow as the canonical project flow; add explicit override for superpowers' default spec/plan locations (`docs/tasks/task-NNN-slug/`); note the modular code-review pattern.
- **`docs/superpowers-integration.md`** — new. Explains when to invoke which command vs. agent vs. skill, for future reference. Acts as a quick-reference companion to `CLAUDE.md`.

## File Inventory

### New

| Path | Purpose |
|---|---|
| `commands/design-task.md` | Phase 2 wrapper — invokes superpowers:brainstorming |
| `commands/plan-task.md` | Phase 3 wrapper — invokes superpowers:writing-plans |
| `commands/execute-task.md` | Phase 4 wrapper — invokes superpowers:subagent-driven-development |
| `agents/plan-adherence-reviewer.md` | Replaces /audit-plan body |
| `agents/backend-guidelines-reviewer.md` | Replaces /backend-audit body |
| `agents/frontend-guidelines-reviewer.md` | Mirrors backend reviewer for TS/React |
| `agents/todo-scanner.md` | Replaces /review-todos body |
| `agents/service-documentation.md` | Renamed from documentation.md, with proper frontmatter |
| `docs/superpowers-integration.md` | When-to-use-what reference |

### Modified

| Path | Change |
|---|---|
| `commands/spec-task.md` | Update handoff note (was → /dev-docs; becomes → /clear then /design-task) |
| `commands/audit-plan.md` | Shrink to one-line agent wrapper |
| `commands/backend-audit.md` | Shrink to one-line agent wrapper |
| `commands/review-todos.md` | Shrink to one-line agent wrapper |
| `commands/service-doc.md` | Shrink to one-line agent wrapper |
| `skills/skill-rules.json` | Add `frontend-dev-guidelines` entry |
| `CLAUDE.md` | Add workflow section, location override, code-review pattern |

### Removed

| Path | Reason |
|---|---|
| `commands/dev-docs.md` | Role split between `/design-task` + `/plan-task` |
| `agents/documentation.md` | Renamed to `service-documentation.md` with frontmatter |

### Untouched

- `commands/recipe-to-cooklang.md`
- `skills/backend-dev-guidelines/**`
- `skills/frontend-dev-guidelines/**`
- `hooks/skill-activation-prompt.{sh,py}`
- `settings.json` (superpowers plugin already enabled)

## Acceptance Criteria

- All four phase commands exist and produce artifacts in `docs/tasks/task-NNN-slug/`.
- Running the four-phase flow on a sample task produces a complete prd.md → design.md → plan.md → executed code with zero references to the old `/dev-docs` or `dev-docs`-style artifacts.
- Code review on a feature branch dispatches the correct subset of three reviewer agents in parallel and produces a combined report.
- `frontend-dev-guidelines` skill banner appears when prompts contain frontend keywords (verified by manually triggering with a frontend-flavored prompt).
- `agents/documentation.md` no longer exists; `agents/service-documentation.md` has proper frontmatter and is invokable by name.
- `CLAUDE.md` and `docs/superpowers-integration.md` document the workflow such that someone who has never seen this repo could pick up the convention from those two files alone.
- `commands/dev-docs.md` is removed; no remaining references to it in the repo.

## Open Questions

None at design time. Implementation may surface details about how exactly `/design-task` and `/plan-task` invoke their underlying skills (passing the task folder as context, etc.) — those will be worked out in the implementation plan.
