# Task 040 — Cooklang Sections & Notes: Implementation Plan

Last Updated: 2026-04-15

---

## Executive Summary

Bring the Home Hub Cooklang parser into spec compliance for two features the current implementation handles incorrectly: **section headers** (multi-`=` forms like `==Dough==`) and **inline notes** (`> Don't burn the roux!`). The parser currently recognizes only `= Name` sections (dropping all other variants into step text) and collects notes into a global metadata list that the UI never renders. After this task, all spec-compliant section forms produce correct `Step.Section` values, and notes are captured positionally and rendered inline between steps in both the live preview and the saved detail view. No schema changes — the detail endpoint re-parses `source` on each read. The Cooklang help panel is updated to document both features.

---

## Current State Analysis

### Parser — `services/recipe-service/internal/recipe/cooklang/parser.go`

- `Parse` pipeline: `extractMetadata → stripComments → stripNotes → splitSteps → parseBlock`.
- `stripNotes` (lines 151–166): removes any line beginning with `>` and appends the text to `meta.Notes` (a single global `[]string`). No positional tracking.
- Section detection (lines 37–40 in `Parse`, lines 291–294 in `splitSteps`): matches only `"= "` prefix or bare `"="`. Anything else (e.g., `==Dough==`) falls through to `parseBlock` and gets emitted as step text.

### Types — `services/recipe-service/internal/recipe/cooklang/types.go`

- `Metadata.Notes []string` field carries global notes.
- `ParseResult` has no notes field of its own.
- `Step.Section` already exists — no change needed.

### Resource/processor layer

- `internal/recipe/resource.go` and `internal/recipe/processor.go` surface `Metadata.Notes` as `metadata.notes` on JSON:API responses. Processor may read it into entity fields; needs verification.

### Frontend

- `src/types/models/recipe.ts` has `notes?: string[]` under `RecipeMetadata`.
- `src/lib/hooks/use-cooklang-preview.ts` surfaces the parse result.
- `src/components/features/recipes/cooklang-preview.tsx` renders preview steps.
- `src/components/features/recipes/recipe-steps.tsx` renders steps for the detail page. Neither component renders notes today.
- `src/components/features/recipes/cooklang-help.tsx` lacks section-variant and note-syntax docs.

---

## Proposed Future State

### Parser

- Section headers: one regex `^=+\s*(.*?)\s*=*\s*$` recognized in both `splitSteps` (block-boundary detection) and `Parse` (header consumption). Empty captured name clears the active section.
- Notes: a new `stripNotesPositional` function walks the post-`stripComments` source, tracking how many step-blocks precede each note line, and returns `[]PositionalNote{Position, Text}` plus a body with note lines removed. Position is counted against *step blocks* (sections and empty lines don't advance position).
- Types: drop `Metadata.Notes`; add `ParseResult.Notes []PositionalNote` and a `PositionalNote struct { Position int; Text string }`.

### API

- `POST /recipes/parse` and `GET /recipes/:id` each expose `attributes.notes: [{position, text}]` at the attribute root. `attributes.metadata.notes` is removed. Detail endpoint re-parses `source` on read — no persistence.

### Frontend

- `RecipeParseResult.attributes` and `RecipeDetailAttributes` gain `notes?: PositionalNote[]`; `RecipeMetadata.notes` is removed.
- A new `RecipeNote` sub-component renders one note in italic muted text with a left border accent and an icon (not color-only, for a11y).
- `RecipeSteps` accepts an optional `notes` prop and interleaves them at their positions. `CooklangPreview` passes through. The recipe detail page passes `notes` from the fetched attributes into `RecipeSteps`. Cook mode is explicitly out of scope.
- `cooklang-help.tsx` gains rows for section variants and the `>` note syntax.

---

## Implementation Phases

### Phase 1 — Parser & Types (backend)

Rewrite `stripNotes` into positional form, update section detection, change `Metadata`/`ParseResult` shapes, adjust any processor call sites that read `Metadata.Notes`.

### Phase 2 — Parser Tests

Exhaustively cover section variants and note positions. Update tests that asserted on `Metadata.Notes`.

### Phase 3 — Resource/JSON:API

Expose `notes` on the parse preview response and on the detail response (re-parsing source on read).

### Phase 4 — Frontend Types & Data Flow

Update type definitions, `use-cooklang-preview` hook, and detail-page data flow so `notes` reaches the step renderer.

### Phase 5 — Rendering & Help

Build `RecipeNote`, interleave notes in `RecipeSteps`, update `CooklangPreview`, and add help-panel rows.

### Phase 6 — Build, Test, Verify

Run service build (`scripts/local-up.sh` or `go build` + `go test` in recipe-service) and frontend type-check/build. Manually verify via docker-compose a recipe round-trips: preview during edit → save → detail → matching inline notes.

---

## Detailed Tasks

### Phase 1 — Parser & Types

**1.1** Update `types.go`:
- Remove `Notes []string` field from `Metadata`.
- Add `PositionalNote struct { Position int \`json:"position"\`; Text string \`json:"text"\` }`.
- Add `Notes []PositionalNote \`json:"notes,omitempty"\`` to `ParseResult`.

**1.2** In `parser.go`, replace `stripNotes(body, &metadata)` with a positional collector that returns `(bodyWithoutNoteLines string, notes []PositionalNote)`. Counting rule: when walking the source line-by-line, maintain a `blockIndex` counter that advances by 1 each time we *close* a non-empty, non-section, non-note block (mirrors `splitSteps` block counting). A note line's position = current `blockIndex`.

**1.3** Replace the `= ` / `=` section checks in `Parse` and `splitSteps` with a single helper `matchSection(trimmed string) (name string, ok bool)` backed by `regexp.MustCompile(`^=+\s*(.*?)\s*=*\s*$`)`. Trim the captured name; empty name is valid and clears `currentSection`.

**1.4** Wire `ParseResult.Notes` in the `Parse` return.

**1.5** Update `processor.go` and any other in-service reader of `parsed.Metadata.Notes` to drop the reference. If the processor was persisting notes onto the entity, remove that path (PRD: not persisted on entity).

**1.6** Update `Validate` to use the new section detection so a `==Dough==` line isn't falsely scanned for ingredient markers.

**Acceptance:** Parser returns positional notes and recognizes all section forms. `go build ./...` and existing tests (unchanged ones) still pass.

### Phase 2 — Parser Tests

**2.1** In `parser_test.go`, add section-variant cases producing `section == "Filling"` for: `= Filling`, `==Filling==`, `== Filling`, `== Filling ==`, `=== Filling ===`, `=  Filling  =`.

**2.2** Add a section-variant case: `==Dough==\nMix @flour{500%g}` → one step with `Section == "Dough"`, no `==` in segments.

**2.3** Add a bare `==` case: steps before have section, steps after have empty section.

**2.4** Add note-position cases (at minimum):
- Note before first step → `Position == 0`.
- Note between step 1 and step 2 → `Position == 1`.
- Note after last step → `Position == len(steps)`.
- Note immediately following a section header (before any step of that section) → position equals the index of the following step.

**2.5** Update any existing test that asserted on `Metadata.Notes` to assert on `ParseResult.Notes`.

**Acceptance:** `go test ./internal/recipe/cooklang/...` green with the new cases.

### Phase 3 — Resource/JSON:API

**3.1** In `internal/recipe/resource.go`, extend the preview response attributes to include `notes` from `ParseResult.Notes`.

**3.2** In the detail handler, after loading the recipe, call `cooklang.Parse(entity.Source)` and include the resulting `Notes` on the response attributes. Do not include `Ingredients`/`Steps` from the re-parse if they are already persisted — only the new `notes` field.

**3.3** Ensure `metadata.notes` is no longer emitted anywhere.

**Acceptance:** curl against a running recipe-service returns `attributes.notes` on both endpoints; `metadata.notes` is absent.

### Phase 4 — Frontend Types & Data Flow

**4.1** In `src/types/models/recipe.ts`:
- Remove `notes?: string[]` from `RecipeMetadata`.
- Add `export type PositionalNote = { position: number; text: string }`.
- Add `notes?: PositionalNote[]` to `RecipeParseResult.attributes` and `RecipeDetailAttributes`.

**4.2** Update `src/lib/hooks/use-cooklang-preview.ts` to surface `notes` alongside steps/ingredients/metadata.

**4.3** Update the recipe detail page to pull `notes` from the detail attributes and pass into `RecipeSteps`.

**Acceptance:** `tsc --noEmit` clean; no dangling `metadata.notes` references.

### Phase 5 — Rendering & Help

**5.1** Create a `RecipeNote` sub-component (collocated with `recipe-steps.tsx`) rendering a single note: italic muted text, left-border accent using an existing semantic token (`border-l-4 border-muted-foreground/40`), and a small icon (e.g., `MessageSquareText` from lucide-react) so contrast isn't color-only.

**5.2** Modify `RecipeSteps` to accept `notes?: PositionalNote[]`. When iterating steps at index `i`, first render every note with `position === i`, then render the step. After the loop, render notes whose position equals `steps.length`.

**5.3** Update `CooklangPreview` to pass the `notes` prop through to `RecipeSteps` / the step renderer it uses.

**5.4** Update `cooklang-help.tsx`: add an entry for section header variants (show `= Name`, `==Name==`, `===Name===`, `==` bare, with a one-line explanation that the number and trailing `=` are cosmetic), and an entry for notes (`> Note text` on its own line, appears between steps).

**Acceptance:** manual QA — a recipe with `==Dough==` and an inline `> ...` note renders correctly in both preview and detail; light and dark themes both legible; screen reader announces the note distinctly.

### Phase 6 — Build, Test, Verify

**6.1** `cd services/recipe-service && go build ./... && go test ./...`.

**6.2** `cd frontend && npm run build` (or `tsc --noEmit` + `npm run lint`).

**6.3** `scripts/local-up.sh`; open the recipe editor; paste a recipe exercising all section variants and inline notes; save; reload the detail page; confirm parity with the preview.

**6.4** Verify existing recipes (no sections, no notes) render identically — no layout regression.

**Acceptance:** all PRD acceptance criteria checked.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Position counting drifts from `splitSteps` block counting, producing off-by-one note placement | Med | Med | Extract block-counting into a shared helper used by both `stripNotesPositional` and `splitSteps`, or unit-test with each note case. |
| Removing `Metadata.Notes` breaks an undiscovered backend consumer | Low | Med | Grep all services for `Metadata.Notes` / `metadata.notes` before deletion; surface any external consumer per PRD §8. |
| Re-parsing source on every detail read adds latency | Low | Low | Parser is pure Go, ≤64KB; PRD explicitly accepts the cost. If hot, cache later. |
| Regex-based section detection over-matches (e.g., markdown `===` heading-underline line) | Low | Low | PRD explicitly accepts permissive behavior; document in help panel. |
| Frontend renders notes with color-only styling | Low | Low (a11y) | Include an icon and border in `RecipeNote` per PRD §8 a11y clause. |

---

## Success Metrics

- All PRD §10 acceptance criteria met.
- Parser tests cover all enumerated section variants + ≥4 note positions.
- Zero new migrations.
- No regressions in existing recipe render tests or visual QA of note-free recipes.

---

## Required Resources and Dependencies

- No new Go modules or npm packages.
- No infra/deploy changes.
- Dev workflow: `scripts/local-up.sh` for end-to-end verification.

Depends on nothing in flight. Touches recipe-service and frontend only.

---

## Timeline Estimates

| Phase | Effort |
|------|--------|
| 1 — Parser & Types | M |
| 2 — Parser Tests | M |
| 3 — Resource/JSON:API | S |
| 4 — Frontend Types & Data Flow | S |
| 5 — Rendering & Help | M |
| 6 — Build, Test, Verify | S |

**Total:** ~1–1.5 engineering days.
