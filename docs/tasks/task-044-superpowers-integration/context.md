# Superpowers Integration — Implementation Context

Last Updated: 2026-04-21

## What's being built

The integration of the `superpowers` plugin (v5.0.7, already installed and enabled) into the Home Hub project's existing `.claude/` setup. See `design.md` in this folder for the full design.

## Authoritative inputs

- **Design:** `docs/tasks/task-044-superpowers-integration/design.md`
- **Existing project tooling:** `.claude/commands/`, `.claude/agents/`, `.claude/skills/`, `.claude/hooks/`, `.claude/skills/skill-rules.json`, `.claude/settings.json`
- **Plugin contents (read-only reference):** `~/.claude/plugins/cache/claude-plugins-official/superpowers/5.0.7/`
  - Skills: `skills/<name>/SKILL.md`
  - Agents: `agents/code-reviewer.md`
- **Project conventions:** `CLAUDE.md` (project root)
- **Existing audit/review templates being absorbed:**
  - `commands/audit-plan.md` (becomes `agents/plan-adherence-reviewer.md`)
  - `commands/backend-audit.md` (becomes `agents/backend-guidelines-reviewer.md`)
  - `commands/review-todos.md` (becomes `agents/todo-scanner.md`)
  - `commands/service-doc.md` + `agents/documentation.md` (merge into `agents/service-documentation.md`)
- **Existing PRD template being preserved:** `commands/spec-task.md`
- **Existing task structure being adopted:** `docs/tasks/task-NNN-slug/`

## Key decisions (locked in design)

| ID | Decision |
|----|----------|
| D1 | Superpowers is the default workflow; phases are explicit commands |
| D2 | Artifact home stays `docs/tasks/task-NNN-slug/`; superpowers default locations are overridden via `CLAUDE.md` |
| D3 | Each phase command (`/spec-task`, `/design-task`, `/plan-task`, `/execute-task`) is invoked from a `/clear`'d session — no auto-handoff |
| D4 | Code review = three modular reviewer agents dispatched in parallel |
| D5 | `review-todos` and `service-doc` promoted to agents; slash commands kept as one-line wrappers |
| D6 | Hook + `skill-rules.json` keep narrow domain-skill role; superpowers self-activates; add `frontend-dev-guidelines` entry |
| D7 | `recipe-to-cooklang` and the backend/frontend guideline skills' resources are untouched |

## Verification model

There are no unit tests for `.claude/` config files. Verification per task is by:

1. **File presence + correct frontmatter** (use `Read` to inspect).
2. **JSON validity** for `skill-rules.json` (`python3 -m json.tool < .claude/skills/skill-rules.json`).
3. **Hook smoke test** for skill-rules changes (pipe a sample JSON prompt into `python3 .claude/hooks/skill-activation-prompt.py` and verify the expected banner / silence).
4. **Live invocation smoke test** at the end (open a fresh session, run the new flow on a tiny task to confirm artifacts land in the right folders).

## Conventions for agent files

Every agent file MUST have YAML frontmatter:

```yaml
---
name: <kebab-case-name>
description: |
  <when to invoke this agent — used for proactive delegation and Skill tool listing>
  <Include 1-2 example scenarios as <example> tags>
model: inherit
---
```

The `description` should be specific enough that the main agent can decide when to delegate. Reference: `~/.claude/plugins/cache/claude-plugins-official/superpowers/5.0.7/agents/code-reviewer.md` for the established style.

## Commit discipline

Commit after each completed task (per superpowers convention). Each task in this plan ends with a commit step. Use conventional-commit-style prefixes: `feat:`, `chore:`, `docs:`, `refactor:`. Co-author tag per Claude Code defaults.

## Out of scope

- Migrating any existing `docs/tasks/task-001` … `task-043` content to the new structure (they remain as-is).
- Changing `commands/recipe-to-cooklang.md`.
- Changing the content of `skills/backend-dev-guidelines/resources/*` or `skills/frontend-dev-guidelines/resources/*`.
- Touching `settings.json` (superpowers plugin already enabled).
- Adding tests/CI for the `.claude/` directory.
