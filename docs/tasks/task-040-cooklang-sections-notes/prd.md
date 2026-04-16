# Cooklang Sections & Notes — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-15
---

## 1. Overview

The Cooklang specification supports two recipe-structure features that Home Hub does not fully handle today: **sections** (e.g., `==Dough==`) for separating components prepared independently, and **notes** (lines beginning with `>`) for inline anecdotes and cooking insights.

Section headers are partially supported: the parser recognizes only the form `= Name` (single leading `=` plus space). The spec-conformant variants `==Dough==`, `===`, bare `==`, etc. are silently treated as step text and corrupt the rendered recipe. Notes are parsed and stripped from steps, but they are collected into a single global metadata list, never displayed in the UI, and not persisted on the saved recipe — so a user who writes `> Don't burn the roux!` next to a step sees nothing about it after saving.

This task brings the parser and rendering pipeline into spec compliance for both features, exposes notes inline at the position they appeared in the source, and updates the in-app Cooklang syntax help.

## 2. Goals

Primary goals:
- Recognize all spec-compliant section header forms in the parser.
- Render notes inline between steps at the position they appear in the source, in both the live preview and the saved recipe view.
- Surface notes on the saved recipe by re-parsing the stored source on read (no schema changes).
- Update the Cooklang help panel to document section variants and the notes syntax.

Non-goals:
- Grouping the ingredients list by section (display stays flat; existing global ingredient dedup behavior is unchanged).
- Schema changes to persist sections or notes as first-class entities.
- Section anchors / table of contents / jump-to-section UX in cook mode or detail view.
- Editing sections or notes via structured UI (raw Cooklang source remains the source of truth).
- Changes to shopping list, meal plan export, or any consumer of recipe data outside the recipe-service detail/preview responses.

## 3. User Stories

- As a recipe author, I want to use any spec-compliant section form (`==Dough==`, `=== Filling ===`, bare `==`) so that recipes copied from third-party Cooklang sources render correctly without manual editing.
- As a cook, I want to read `> Don't burn the roux!` next to the step it warns about so that critical guidance isn't buried at the bottom of the recipe or lost entirely.
- As a recipe author, I want the live preview and the saved recipe to show notes the same way so that what I see while editing matches what I see while cooking.
- As a new user, I want the in-app syntax help to document sections and notes so that I don't have to leave the app to find the spec.

## 4. Functional Requirements

### 4.1 Section parsing

- The parser MUST recognize a section header on any line whose trimmed content matches the regex `^=+\s*(.*?)\s*=*$`, where:
  - The line starts with one or more `=` characters.
  - An optional name follows (may contain spaces).
  - An optional run of trailing `=` characters closes the header.
- The number of leading and trailing `=` characters has no semantic meaning; all forms produce the same `Step.Section` value.
- An empty section name (e.g., `==`, `===`, `= =`) is valid and resets the active section to empty (acts as a divider — subsequent steps have no section attached).
- Section header lines MUST NOT be emitted as steps and MUST NOT contribute text to adjacent steps.
- A section header MUST start a new logical block even when not preceded by a blank line (matches existing `splitSteps` behavior for the `= ` form).
- The active section applies to every step that follows it until another section header appears.

### 4.2 Note parsing

- A note is any line whose trimmed content begins with `>`. The note text is the trimmed remainder after the leading `>`.
- Notes MUST be captured **positionally** rather than collected into a single global list. A note's position is "between block N and block N+1" in the post-comment-stripped source.
- Consecutive note lines (no blank line separating them) MAY be merged into a single multi-line note for that position, OR emitted as separate notes at the same position. (Implementation choice; the rendering is equivalent.)
- Notes MUST NOT be emitted as steps and MUST NOT contribute text to adjacent steps.
- The existing global `Metadata.Notes` field is removed — notes no longer appear in metadata. (The frontend type and any backend consumers are updated accordingly.)

### 4.3 Live preview rendering

- The Cooklang preview panel (`cooklang-preview.tsx`) MUST render notes interleaved with steps at their parsed positions, visually distinct from steps (e.g., italic muted text with a left border accent — exact styling at implementer discretion within existing design tokens).
- A note that appears before the first step MUST render above the first step. A note that appears after the last step MUST render below the last step.
- The ingredients list rendering is unchanged.

### 4.4 Saved recipe rendering

- The recipe detail page MUST render notes inline between steps using the same component used by the live preview.
- Notes MUST be derived by re-parsing the stored `source` field on each detail read; they are not persisted in a column or join table.
- The cook mode views (all-steps and single-step) are out of scope for note rendering in this task — notes appear only in the detail page and the live preview.

### 4.5 Help text

- `cooklang-help.tsx` MUST document:
  - Section header variants: `= Name`, `== Name ==`, `===`, etc., with a note that the number of `=` and the trailing `=` are optional.
  - Note syntax: `> Note text` placed on its own line between steps.

## 5. API Surface

No new endpoints. Modified response shapes:

### `POST /recipes/parse` (live preview)

- `attributes.metadata.notes` — REMOVED.
- `attributes.steps[]` — unchanged shape.
- `attributes.notes` — NEW. An array of `{ position: number, text: string }` entries. `position` is the zero-based index of the step the note appears **before**; a note appearing after the last step uses `position == steps.length`.

### `GET /recipes/:id` (saved recipe detail)

- `attributes.notes` — NEW. Same shape as above. Derived by parsing `attributes.source` server-side.

### Errors

No new error cases. Malformed section headers (e.g., `=` followed by garbage) are treated as section headers per the regex; this is intentionally permissive and matches the spec's "anything goes" tone.

## 6. Data Model

No schema changes. No new entities, columns, or migrations.

The `Metadata.Notes []string` field on the parser's `Metadata` struct is removed. A new top-level `Notes []PositionalNote` field is added to `ParseResult`, where `PositionalNote = { Position int, Text string }`.

## 7. Service Impact

### `recipe-service`

- `internal/recipe/cooklang/parser.go`:
  - Replace the `= ` / `=` section detection with the regex-based form.
  - Replace `stripNotes` with positional note collection that preserves ordering relative to step blocks.
- `internal/recipe/cooklang/types.go`:
  - Remove `Notes []string` from `Metadata`.
  - Add `Notes []PositionalNote` to `ParseResult`; add `PositionalNote` type.
- `internal/recipe/cooklang/parser_test.go`:
  - Add cases for each section variant (`==Dough==`, `=== Filling ===`, bare `==`, single `=`, with/without trailing whitespace).
  - Add cases for inline notes at start, between steps, after last step, and across section boundaries.
  - Update existing tests that asserted on `Metadata.Notes`.
- `internal/recipe/processor.go`:
  - Stop reading `parsed.Metadata.Notes` (field no longer exists). No other behavior change — notes are not persisted on the entity.
- `internal/recipe/resource.go`:
  - Include the new `Notes` field on the parse-preview JSON:API response.
  - Re-parse `source` when serving recipe detail and include the `Notes` array on the detail response. (Acceptable cost; parser is in-process and existing detail handlers already do non-trivial work.)

### `frontend`

- `src/types/models/recipe.ts`:
  - Remove `notes?: string[]` from `RecipeMetadata`.
  - Add `PositionalNote` type and `notes?: PositionalNote[]` to `RecipeParseResult.attributes` and `RecipeDetailAttributes`.
- `src/lib/hooks/use-cooklang-preview.ts`:
  - Track and return the new `notes` array.
- `src/components/features/recipes/cooklang-preview.tsx`:
  - Pass notes through to the steps renderer.
- `src/components/features/recipes/recipe-steps.tsx`:
  - Accept an optional `notes` prop and interleave them at their declared positions.
  - A small new sub-component (e.g., `RecipeNote`) renders a single note with consistent styling.
- Recipe detail page (wherever steps are rendered for the saved recipe) — pass `notes` from the detail attributes into `RecipeSteps`.
- `src/components/features/recipes/cooklang-help.tsx`:
  - Add documentation rows for section variants and the `>` note syntax.

No changes to cook mode components in this task.

## 8. Non-Functional Requirements

- **Performance**: re-parsing the source on each detail read is acceptable. Source size is bounded by `MaxSourceSize = 64 KB` and the parser is pure-Go with no I/O. No caching layer needed.
- **Multi-tenancy**: no new data is persisted; existing tenant scoping on recipe reads is unaffected.
- **Backward compatibility**: removal of `metadata.notes` from the parse response is a breaking shape change. The frontend and any backend caller MUST be updated together. There are no external API consumers of recipe-service today (verify during implementation), so no deprecation window is required. If an external consumer is found, surface for re-discussion before removing the field.
- **Observability**: no new metrics or logs required. Parse failures already surface through `ParseError`.
- **Accessibility**: the inline note styling MUST preserve sufficient contrast in both light and dark themes and MUST NOT rely on color alone to distinguish notes from steps (e.g., include an icon or border).

## 9. Open Questions

None — all scope decisions resolved during spec conversation. Implementation choice on whether consecutive `>` lines merge into one note vs. multiple is left to the implementer; the rendering is equivalent.

## 10. Acceptance Criteria

- [ ] A recipe source containing `==Dough==` produces a step with `section == "Dough"`; the literal `==` does not appear in any rendered step.
- [ ] All of these forms produce a section named `Filling`: `= Filling`, `== Filling`, `== Filling ==`, `=== Filling ===`, `=  Filling  =`.
- [ ] A bare `==` (no name) ends the previous section and clears the active section for following steps.
- [ ] A recipe source with `> Don't burn the roux!` on a line between two steps renders the note visually between those two steps in both the live preview and the saved detail view.
- [ ] A note appearing before the first step renders above the first step; a note after the last step renders below the last step.
- [ ] The Cooklang help panel shows the new section variants and note syntax.
- [ ] Parser unit tests cover every section form listed above and at least four note positions (start, between two steps, end, immediately following a section header).
- [ ] No new database migrations are introduced.
- [ ] `metadata.notes` is no longer present on any parse or detail response; `notes` array is present at the response root for both endpoints.
- [ ] Existing recipes (no sections, no notes) render identically to before this change.
