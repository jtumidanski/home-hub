# UI Cleanup Batch 2 — Context

Last Updated: 2026-04-08

## Source Documents

- PRD: `docs/tasks/task-022-ui-cleanup-batch-2/prd.md`
- Companion task (prior batch): task-017 (UI Cleanup Batch 1)
- Project guidelines: `CLAUDE.md` (root)

## Key Files — Frontend

| Item | Path | Lines | Purpose |
|---|---|---|---|
| 1 | `frontend/src/types/models/task.ts` | 45-51 | `isTaskOverdue` — date-comparison fix |
| 1 | `frontend/src/types/models/__tests__/task.test.ts` | (new) | Unit tests for `isTaskOverdue` |
| 1 | `frontend/src/pages/TasksPage.tsx` | 23-27, 136-143 | Status filter dropdown + badge logic (consume `isTaskOverdue`, no change needed) |
| 2/4 | `frontend/src/components/ui/button.tsx` | `buttonVariants` base | Add `cursor-pointer` |
| 2/4 | `frontend/src/components/ui/dialog.tsx` | `DialogClose` | Add `cursor-pointer` |
| 2/4 | `frontend/src/components/ui/sheet.tsx` | close button | Verify + patch if applicable |
| 2/4 | `frontend/src/pages/IngredientsPage.tsx` | 185-193 | Bare `<button>` search "X" — add `cursor-pointer` |
| 3 | `frontend/src/pages/TasksPage.tsx` | 101-118 | Desktop complete-toggle column |
| 3 | `frontend/src/components/features/tasks/task-card.tsx` | (whole) | Mobile card complete-toggle |
| 5 | `frontend/src/components/features/calendar/all-day-event-row.tsx` | (whole) | All-day events full-width |
| 5 | `frontend/src/components/features/calendar/calendar-grid.tsx` | 100-104 | Caller — preserves per-day rendering |
| 6 | `frontend/src/pages/IngredientDetailPage.tsx` | 241-261 | "Used in Recipes" rendering |
| 6 | `frontend/src/types/models/...` (RecipeRef) | — | Add `recipeName` to TS type |
| 7 | `frontend/src/components/features/recipes/cook-mode.tsx` | 191-210 | `SingleStepView` scroll fix |
| 8 | `frontend/src/pages/IngredientsPage.tsx` | 64-71, 249-253 | Client-side category lookup; uncategorized count unchanged |
| 8 | `frontend/src/pages/IngredientDetailPage.tsx` | 188-189 | Reference pattern for client-side lookup |
| 8 | `frontend/src/types/models/ingredient.ts` | `CanonicalIngredientListItem` | Drop `categoryName` |

## Key Files — Backend (recipe-service)

| Item | Path | Lines | Purpose |
|---|---|---|---|
| 6 | `services/recipe-service/internal/ingredient/provider.go` | 88-103 | `RecipeRef` + `getIngredientRecipes` — add `recipeName` via JOIN against `recipes` |
| 6 | `services/recipe-service/internal/ingredient/rest.go` | — | JSON:API serializer — expose `recipeName` |
| 6 | `services/recipe-service/internal/ingredient/processor_integration_test.go` or new test | — | Verify JOIN returns recipe name |
| 8 | `services/recipe-service/internal/ingredient/provider.go` | 115-124, 147-149 | Remove `CategoryName` field + stale `LEFT JOIN ingredient_categories` |
| 8 | `services/recipe-service/internal/ingredient/model.go` | 43, 57, 69 | Drop `categoryName` from Model (if no other readers) |
| 8 | `services/recipe-service/internal/ingredient/builder.go` | 22, 38, 59 | Drop `categoryName` from Builder (if no other readers) |
| 8 | `services/recipe-service/internal/ingredient/rest.go` | 15, 32 | Drop `categoryName` from JSON:API list response |
| 8 | `services/recipe-service/internal/ingredient/entity.go` | `Migration` (line 34-44) | Add `DROP TABLE IF EXISTS ingredient_categories` |

## Key Decisions

1. **Single-primitive cursor fix.** Apply `cursor-pointer` once at the `Button` / `DialogClose` primitive level. Do **not** add per-call-site overrides for any button covered by the global fix. Bare `<button>` elements outside primitives are patched individually.

2. **Local-date parsing for `isTaskOverdue`.** Use `new Date(year, month-1, day)` and compare against `today` zeroed to local midnight. Do **not** rely on `new Date(string)` which is timezone-ambiguous.

3. **Backend cleanup, not category re-architecture.** Item 8 is fixed by removing the stale local JOIN and dropping the orphaned table. Categories remain managed exclusively by `category-service`.

4. **Fully delete `categoryName` from `recipe-service` ingredient layer (decision).** Per the project guideline against backwards-compat shims, drop the field everywhere it appears in `recipe-service/internal/ingredient/`:
   - `provider.go:117,123` — `CategoryName` on `entityWithUsage` and `entityWithCategory`.
   - `model.go:43,57,69` — `categoryName` field, `CategoryName()` getter, `WithCategoryName()` setter.
   - `builder.go:22,38,59` — `categoryName` field, `SetCategoryName` method, builder copy.
   - `processor.go:234` — `.WithCategoryName(e.CategoryName)` call in `SearchWithUsage`.
   - `rest.go:15,32,126,139` — `CategoryName` JSON field and serializer references.
   Verified via `grep`: these are the **only** readers/writers in `recipe-service`. No external consumer remains after the JSON field is removed. The frontend `CanonicalIngredientListItem.categoryName` is dropped in the same PR (see decision 5).

5. **Cook Mode flexbox-overflow centering trap.** Wrap the inner content in `min-h-full flex ... justify-center` and remove `justify-center` from the scroll container. Short content stays centered; long content scrolls from the top.

6. **Multi-day all-day events.** Each day's instance of a multi-day event is rendered independently and must be full-width on its own day column. Per-day bucketing in `calendar-grid.tsx` is preserved.

7. **PR strategy: one bundled PR.** All eight items ship in a single PR. C1+C3 must land together regardless, and bundling avoids the cross-PR coordination cost. Phases are organized for review readability via commit boundaries, not separate PRs.

8. **`recipes.title` is the column joined for item 6.** Verified in `services/recipe-service/internal/recipe/entity.go:15` — the field is `Title string` mapped to `recipes.title`. The B1 JOIN is:
   ```go
   query := db.Table("recipe_ingredients").
       Joins("JOIN recipes ON recipes.id = recipe_ingredients.recipe_id").
       Where("recipe_ingredients.canonical_ingredient_id = ?", canonicalIngredientID)
   // SELECT recipe_ingredients.recipe_id, recipe_ingredients.raw_name, recipes.title AS recipe_name
   ```
   `recipes.deleted_at IS NULL` should also be added so soft-deleted recipes do not appear (verified in `entity.go:23` — `recipes` uses soft-delete via `DeletedAt`).

## Tenant-scoping note (item 6)

`getIngredientRecipes` (`provider.go:93-103`) currently queries `recipe_ingredients` by `canonical_ingredient_id` only. The HTTP handler `ingredientRecipesHandler` (`resource.go:268-296`) does **not** verify that the URL-provided ingredient `id` belongs to the caller's tenant — this is a pre-existing gap, not introduced by item 6. The new `recipes` JOIN is itself tenant-safe because `recipe_ingredients` rows are owned by the recipe (same tenant by construction). **Decision: do not expand this task's scope to close the handler-level tenant gap.** Flagging here so it can be tracked separately.

## RecipeRef TS type location (item 6)

The matching TypeScript type is `IngredientRecipeRef` in `frontend/src/types/models/ingredient.ts:55-58`:
```ts
export interface IngredientRecipeRef {
  recipeId: string;
  rawName: string;
}
```
B2 adds `recipeName: string` to this interface. The two other `rawName` references in `frontend/src/types/models/recipe.ts:29,40` are unrelated forward-direction types (recipe → ingredients) and are not touched.

## Sheet primitive (item A2)

`frontend/src/components/ui/sheet.tsx` **does not exist** in this codebase (verified via `ls frontend/src/components/ui/`). A2 has nothing to patch there — remove that bullet from the work.

## Snapshot tests (item A2)

`frontend/src/components/ui/__tests__/` contains only `user-avatar.test.tsx`. There are **no** `Button` or `Dialog` snapshot tests, so the cursor-pointer change carries no snapshot-update risk.

## Dependencies & Sequencing

- **C1 must precede C2** (drop the JOIN before dropping the table).
- **C1 + C3 should land in the same PR** so the frontend never depends on a `categoryName` field that the backend stopped sending.
- **B1 must precede B2** (backend field exists before frontend reads it). Acceptable to ship in one PR.
- Phase A items are mutually independent.

## Testing Strategy

- **Unit tests:** `isTaskOverdue` (item 1), recipe-name JOIN (item 6).
- **Manual smoke (per PRD §10):**
  1. Tasks page on a day with a due-today task → "pending" badge.
  2. Hover every surface in PRD §4.2 → pointer cursor.
  3. Toggle a task complete/incomplete → icon swaps; aria-label flips.
  4. Calendar with single-day and multi-day all-day events → full width on every covered day.
  5. Ingredient detail page → "Used in Recipes" lists recipe titles.
  6. Cook Mode with a long step → scrollable from the top on mobile + desktop.
  7. Categorize an ingredient on its detail page → list page shows the category without hard refresh; persists across hard refresh.
- **Build verification (per CLAUDE.md):** frontend `npm run build` + Docker build for recipe-service.
- **Local env:** use `scripts/local-up.sh` / `scripts/local-down.sh` for end-to-end manual smoke.

## Open Questions

None — all resolved during PRD review (PRD §9).
