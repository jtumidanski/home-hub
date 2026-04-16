# Plan Audit — task-040-cooklang-sections-notes

**Plan Path:** docs/tasks/task-040-cooklang-sections-notes/tasks.md
**Audit Date:** 2026-04-15
**Branch:** recipe-ingestion
**Base Branch:** main

## Executive Summary

All 24 implementation tasks in Phases 1–5 are implemented in the working tree despite every checkbox in `tasks.md` being left unchecked. Backend builds and tests pass. Frontend type-check and lint pass. Phase 6 manual QA items (`6.3`, `6.4`) and PRD acceptance criteria checkboxes were not marked — these are human-verification steps that cannot be confirmed from static inspection but the code paths are in place. No developer-guidelines violations were found in modified code.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Types: remove Metadata.Notes, add PositionalNote + ParseResult.Notes | DONE | services/recipe-service/internal/recipe/cooklang/types.go:44-55 |
| 1.2 | stripNotesPositional with block-index counting | DONE | parser.go:171-206 |
| 1.3 | matchSection regex helper used in Parse and splitSteps | DONE | parser.go:75-90, 36, 335 |
| 1.4 | Wire ParseResult.Notes in Parse | DONE | parser.go:22, 67-72 |
| 1.5 | Drop parsed.Metadata.Notes reads | DONE | Grep shows no remaining readers in services/ |
| 1.6 | Validate skips `=+` headers | DONE | parser.go:241-243 (uses matchSection) |
| 2.1 | Section-variant tests → section == "Filling" | DONE | parser_test.go:451-471 (all six variants) |
| 2.2 | ==Dough==\nMix inline case | DONE | parser_test.go:473-481 |
| 2.3 | Bare `==` clears active section | DONE | parser_test.go:483-489 |
| 2.4 | Note positions (≥4 cases) | DONE | parser_test.go:491-525 (before/between/after/post-section = 4 positions) |
| 2.5 | Update existing tests asserting on Metadata.Notes | DONE | TestParse_Blockquotes and Stuffed Peppers now assert on result.Notes (parser_test.go:243-253, 402) |
| 3.1 | Emit attributes.notes on POST /recipes/parse | DONE | resource.go:52-67 (Notes field on RestParseModel with empty-slice fallback); rest.go:150 |
| 3.2 | Re-parse source + emit notes on GET /recipes/:id | DONE | processor.go:150 re-parses; rest.go:121-129 TransformDetail includes Notes; resource.go:227 wires through |
| 3.3 | metadata.notes no longer emitted | DONE | types.go Metadata struct has no Notes field |
| 4.1 | Remove RecipeMetadata.notes, add PositionalNote, add notes? to attrs types | DONE | frontend/src/types/models/recipe.ts:76-89, 153 |
| 4.2 | Surface notes from use-cooklang-preview | DONE | use-cooklang-preview.ts:11, 22, 57, 76 |
| 4.3 | Pass notes from detail page to RecipeSteps | DONE | RecipeDetailPage.tsx:170 |
| 5.1 | RecipeNote sub-component (italic + border + icon) | DONE | recipe-steps.tsx:10-17 (MessageSquareText + border-l-4 + italic + sr-only label) |
| 5.2 | RecipeSteps interleaves notes at declared positions incl. trailing | DONE | recipe-steps.tsx:24-65 |
| 5.3 | CooklangPreview passes notes through | DONE | cooklang-preview.tsx:10, 14, 51 |
| 5.4 | Help panel rows for section variants + `>` notes | DONE | cooklang-help.tsx:42-45 |
| 6.1 | go build + go test pass | DONE | `go build ./...` clean; `go test ./... -count=1` all packages ok |
| 6.2 | Frontend tsc + lint + build pass | PARTIAL | `tsc --noEmit` clean; `npm run lint` clean; `npm run build` not executed in this audit |
| 6.3 | local-up.sh smoke (sections + notes render) | SKIPPED | Manual QA — no evidence in repo; cannot verify statically |
| 6.4 | No regression for recipes without sections/notes | SKIPPED | Manual QA — no evidence; existing unit tests for section-free/note-free paths pass |

**Completion Rate:** 22 DONE / 1 PARTIAL / 2 SKIPPED (manual QA) = 22/25 fully verifiable (88%)
**Skipped without approval:** 0 (skipped items are manual QA gates, not code tasks)
**Partial implementations:** 1 (6.2 — `npm run build` not re-run; tsc + lint both pass)

## Skipped / Deferred Tasks

- **6.3 / 6.4 — docker-compose smoke test:** not possible to verify from static audit. Recommend the implementer run `scripts/local-up.sh` and paste a recipe exercising `==Dough==`, `==Filling==`, bare `==`, and `> …` notes; then reload a saved recipe to verify detail parity.
- **6.2 — `npm run build`:** tsc and lint pass; the full Vite build was not re-executed. Low risk but worth running before merge.
- **PRD acceptance-criteria checkboxes:** all left unchecked in tasks.md. The corresponding implementation is present; update the checkboxes after manual QA.

## Developer Guidelines Compliance

### Passes

- **Backend — Immutable models:** no domain model fields were mutated. `PositionalNote` is a transport/value type, appropriately placed in the cooklang package (types.go:44-47).
- **Backend — Resource/handler separation:** `resource.go` builds the REST model; parsing/transform logic stays in processor + rest.go (TransformDetail/RestParseModel).
- **Backend — Multi-tenancy:** no tenancy changes needed; existing `tenantctx.MustFromContext` usage preserved (resource.go:48).
- **Backend — Pure parser:** `stripNotesPositional` and `matchSection` are pure functions without side effects.
- **Frontend — Type safety:** `PositionalNote` exported and threaded through hook → page → component.
- **Frontend — A11y:** `RecipeNote` uses icon + left border (not color-only) + sr-only "Note: " prefix (recipe-steps.tsx:10-17) per PRD §8.
- **Frontend — React Query / form patterns:** no changes here; preview hook retains debounced fetch.

### Violations

None observed in new/modified code.

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| recipe-service (Go) | PASS | PASS | `go build ./...` clean; `go test ./... -count=1` all packages green incl. cooklang |
| frontend (TS/React) | PASS (tsc+lint) | N/A | `tsc --noEmit` clean; `eslint .` clean. `npm run build` not executed. |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — all code-level tasks implemented; only manual QA gates and the tasks.md checkboxes themselves remain.
- **Guidelines Compliance:** COMPLIANT
- **Recommendation:** NEEDS_REVIEW — ready to merge after: (1) running the Phase 6.3/6.4 docker-compose smoke test, (2) executing `npm run build`, and (3) checking the tasks.md + PRD acceptance checkboxes.

## Action Items

1. Run `scripts/local-up.sh` and manually QA a recipe exercising all section variants + inline `>` notes in both preview and detail views (tasks 6.3, 6.4).
2. Run `cd frontend && npm run build` to fully close task 6.2.
3. Tick the completed checkboxes in `docs/tasks/task-040-cooklang-sections-notes/tasks.md` and the PRD acceptance criteria.
