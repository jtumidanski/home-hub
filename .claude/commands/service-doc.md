# Command: Document Home Hub Service

## Purpose
Generate or update service documentation using the Home Hub Documentation Agent, following `DOCS.md` and the architectural constraints in `CLAUDE.md`.

This command operates on exactly one service at a time.

## Inputs
- `CLAUDE.md` (repo root)
- `DOCS.md` (repo root)
- `.claude/agents/documentation.md` (repo root)

## Instructions

Use the Home Hub Documentation Agent.

Authoritative inputs:
- CLAUDE.md
- DOCS.md
- agents/documentation.md
- The source code for the specified service

Task:
Generate or update documentation for the $ARGUMENTS service.

Scope:
- Operate only within the service directory
- Create missing required documentation files if necessary
- Update existing documentation to match current code
- Do not modify any code

Output requirements:
- Output updated documentation files only
- No commentary, no analysis, no recommendations
- If a required doc file cannot be produced from the available code, ask a single targeted question and stop
