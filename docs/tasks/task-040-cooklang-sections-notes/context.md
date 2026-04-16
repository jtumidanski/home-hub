# Task 040 — Cooklang Sections & Notes: Context

Last Updated: 2026-04-15

---

## Key Files

### Backend (`services/recipe-service`)

| File | Role |
|------|------|
| `internal/recipe/cooklang/parser.go` | `Parse`, `stripNotes` (lines 151–166), section detection (lines 37–40, 291–294), `Validate` note/section skip (line 198) — all need updating |
| `internal/recipe/cooklang/types.go` | `Metadata.Notes` removed; add `PositionalNote` + `ParseResult.Notes` |
| `internal/recipe/cooklang/parser_test.go` | Add section-variant + note-position cases; update any existing assertions on `Metadata.Notes` |
| `internal/recipe/processor.go` | Drop reads of `parsed.Metadata.Notes` (do not persist notes on the entity) |
| `internal/recipe/resource.go` | Emit `attributes.notes` on parse preview response; re-parse `source` on detail handler and emit `attributes.notes` |

### Frontend (`frontend`)

| File | Role |
|------|------|
| `src/types/models/recipe.ts` | Remove `RecipeMetadata.notes`; add `PositionalNote`, add `notes?` to `RecipeParseResult.attributes` and `RecipeDetailAttributes` |
| `src/lib/hooks/use-cooklang-preview.ts` | Surface the new `notes` array from the parse response |
| `src/components/features/recipes/cooklang-preview.tsx` | Pass `notes` through to `RecipeSteps` (or its step-renderer equivalent) |
| `src/components/features/recipes/recipe-steps.tsx` | Accept `notes?: PositionalNote[]`; interleave at declared positions; new `RecipeNote` sub-component |
| `src/components/features/recipes/cooklang-help.tsx` | Add section-variant and `>` note syntax rows |
| Recipe detail page (rendering `RecipeSteps` for saved recipe) | Pass `notes` from detail attributes into `RecipeSteps` |

---

## Key Decisions

1. **No persistence.** Notes are re-derived by parsing `source` on every detail read. The parser is in-process and the source is capped at 64KB — PRD explicitly accepts the cost.
2. **Position counts step blocks, not lines.** `PositionalNote.position` is "render this note before step N (zero-indexed)" — a note after the last step uses `position == steps.length`. Section headers and empty lines don't advance the counter.
3. **Single regex for all section forms.** `^=+\s*(.*?)\s*=*\s*$` drives both `splitSteps` block boundaries and header consumption in `Parse`. Empty captured name clears the active section (acts as divider).
4. **Notes carried at the attribute root, not under metadata.** Matches PRD §5; cleanly separates structural content from descriptive metadata.
5. **Shared `RecipeNote` component.** One rendering of a note is reused by preview and detail so styling stays in sync, per PRD §4.3/§4.4.
6. **Accessibility by composition.** Note styling uses italic + left border + icon — never color alone (PRD §8 a11y).
7. **Cook mode out of scope.** Per PRD §4.4, notes appear only in the preview and detail page in this task.

---

## Dependencies

- No new Go modules or npm packages.
- No database migrations.
- No cross-service changes — recipe-service and frontend only.

---

## Risks & Gotchas

- **Off-by-one on note position.** The block counter in `stripNotesPositional` must advance exactly when `splitSteps` closes a step block (not on section headers, not on empty lines). Recommend a shared helper.
- **`Metadata.Notes` removal.** Grep backend for any remaining reader before deleting. External API consumers: none known (PRD §8). Surface for re-discussion if found.
- **`Validate` function** currently skips lines starting with `"= "` — update it to skip any `=+`-prefixed line, else a `==Dough==` line could produce spurious marker errors.
- **Re-parse cost on detail.** Acceptable per PRD; revisit only if measurable latency surfaces.

---

## Verification Checklist

- `go build ./... && go test ./...` in `services/recipe-service`.
- `npm run build` (or `tsc --noEmit` + `npm run lint`) in `frontend`.
- `scripts/local-up.sh` smoke test: create a recipe with `==Dough==`, `==Filling==`, bare `==`, and an inline `> note`. Confirm preview and detail match.
- Accessibility: verify note contrast in both light and dark themes; verify a non-color distinguisher (icon/border) is visible.
