# Task 040 — Cooklang Sections & Notes: Task Checklist

Last Updated: 2026-04-15

---

## Phase 1 — Parser & Types (Backend)

- [ ] 1.1 Update `types.go`: remove `Metadata.Notes`, add `PositionalNote`, add `ParseResult.Notes`
- [ ] 1.2 Replace `stripNotes` with `stripNotesPositional` returning `(body, []PositionalNote)` with block-index counting
- [ ] 1.3 Add `matchSection` helper using regex `^=+\s*(.*?)\s*=*\s*$`; use in both `Parse` and `splitSteps`
- [ ] 1.4 Wire `ParseResult.Notes` in the `Parse` return value
- [ ] 1.5 Drop `parsed.Metadata.Notes` reads in `processor.go` (and anywhere else in-service)
- [ ] 1.6 Update `Validate` to skip `=+`-prefixed lines (not only `= `)

## Phase 2 — Parser Tests

- [ ] 2.1 Section variants producing `section == "Filling"`: `= Filling`, `==Filling==`, `== Filling`, `== Filling ==`, `=== Filling ===`, `=  Filling  =`
- [ ] 2.2 `==Dough==\nMix @flour{500%g}` → one step with `Section == "Dough"`, no `==` in segments
- [ ] 2.3 Bare `==` clears the active section for following steps
- [ ] 2.4 Note positions: before first step (0), between steps (1), after last step (len), immediately after a section header
- [ ] 2.5 Update existing tests that asserted on `Metadata.Notes`

## Phase 3 — Resource / JSON:API

- [ ] 3.1 Emit `attributes.notes` on `POST /recipes/parse` preview response
- [ ] 3.2 Re-parse `source` in the detail handler; emit `attributes.notes` on `GET /recipes/:id`
- [ ] 3.3 Confirm `metadata.notes` is no longer emitted on any response

## Phase 4 — Frontend Types & Data Flow

- [ ] 4.1 Remove `notes?: string[]` from `RecipeMetadata`; add `PositionalNote`; add `notes?` to `RecipeParseResult.attributes` and `RecipeDetailAttributes`
- [ ] 4.2 Surface `notes` from `use-cooklang-preview`
- [ ] 4.3 Pass `notes` from the recipe detail page into `RecipeSteps`

## Phase 5 — Rendering & Help

- [ ] 5.1 Build `RecipeNote` sub-component (italic muted text + left border + icon — not color-only)
- [ ] 5.2 Modify `RecipeSteps` to accept `notes?` and interleave at declared positions (including `position === steps.length`)
- [ ] 5.3 Pass `notes` through `CooklangPreview`
- [ ] 5.4 Add section-variant and `>` note-syntax rows to `cooklang-help.tsx`

## Phase 6 — Build, Test, Verify

- [ ] 6.1 `go build ./... && go test ./...` in `services/recipe-service` passes
- [ ] 6.2 Frontend `tsc --noEmit` + `npm run lint` + `npm run build` pass
- [ ] 6.3 `scripts/local-up.sh` smoke: recipe with `==Dough==`, `==Filling==`, bare `==`, and inline `>` notes renders correctly in preview and detail
- [ ] 6.4 No regression for recipes without sections or notes

---

## PRD Acceptance Criteria

- [ ] `==Dough==` produces `section == "Dough"`; literal `==` absent from rendered steps
- [ ] `= Filling`, `== Filling`, `== Filling ==`, `=== Filling ===`, `=  Filling  =` all produce `section == "Filling"`
- [ ] Bare `==` ends previous section and clears active section
- [ ] `> Don't burn the roux!` between two steps renders between them in preview and detail
- [ ] Note before first step renders above first step; note after last step renders below last step
- [ ] Help panel documents section variants and note syntax
- [ ] Parser tests cover all section forms and ≥4 note positions
- [ ] No new database migrations
- [ ] `metadata.notes` absent from all responses; root-level `notes` array present on parse and detail responses
- [ ] Recipes without sections or notes render identically to before
