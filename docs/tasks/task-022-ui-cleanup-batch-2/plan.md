# UI Cleanup Batch 2 — Implementation Plan

Last Updated: 2026-04-08

## Executive Summary

Eight discrete UI/UX defects spanning the Tasks page, Calendar, Cook Mode, Ingredient pages, and shared UI primitives. Six items are pure frontend tweaks; two items (recipe-name display in "Used in Recipes" and stale "uncategorized" badge) require coordinated changes in `recipe-service`. The work is low-risk and parallelizable, but item 8 also includes a schema cleanup (drop the orphaned `ingredient_categories` table) and item 6 introduces a new JOIN that should be verified for tenant scoping.

## Current State Analysis

- **Item 1 — Overdue logic** (`frontend/src/types/models/task.ts:45-51`): `isTaskOverdue` constructs `new Date(dueOn)` from a `YYYY-MM-DD` string, which parses as UTC midnight. Any user east of UTC sees due-today tasks as overdue.
- **Item 2/4 — Cursor affordances**: shadcn `Button` (`frontend/src/components/ui/button.tsx`) and `DialogClose` (`frontend/src/components/ui/dialog.tsx`) do not include `cursor-pointer`. Native `<button>` defaults to arrow cursor in browsers.
- **Item 3 — Complete-toggle icon** (`frontend/src/pages/TasksPage.tsx:101-118`, `task-card.tsx`): static `Check` icon for both states; user cannot tell what the click will do.
- **Item 5 — All-day events** (`frontend/src/components/features/calendar/all-day-event-row.tsx`): wraps events in `flex flex-wrap gap-1`, so each event shrinks to its title width.
- **Item 6 — "Used in Recipes" labels** (`frontend/src/pages/IngredientDetailPage.tsx:241-261`): renders `{ref.rawName} (recipe)` (the unparsed ingredient string from the recipe), not the recipe title. Backend `RecipeRef` (`services/recipe-service/internal/ingredient/provider.go:88-103`) only carries `recipeId` and `rawName`.
- **Item 7 — Cook Mode single-step view** (`frontend/src/components/features/recipes/cook-mode.tsx:191-210`): `flex-1 ... justify-center ... overflow-y-auto` triggers the flexbox+overflow centering trap, making the top of long step text unreachable.
- **Item 8 — Stale "uncategorized" badge**: `searchWithUsage` in `services/recipe-service/internal/ingredient/provider.go:147-149` joins a legacy local `ingredient_categories` table that is no longer the source of truth. Categories are managed by `category-service`. The detail page (`IngredientDetailPage.tsx:188-189`) avoids the bug by resolving names client-side via `useIngredientCategories`.

## Proposed Future State

- `isTaskOverdue` operates in local-calendar-day terms; today is never overdue.
- All `Button`s and `DialogClose`s show `cursor-pointer` on hover via a single primitive-level fix.
- Tasks page complete-toggle uses `Circle` / `CheckCircle2` paired with a context-sensitive `aria-label`, on both desktop and mobile.
- All-day calendar events render as full-width blocks within their day column; multi-day events span every covered day.
- "Used in Recipes" lists recipe titles (clickable) by way of a new `recipeName` field on `RecipeRef`, populated via a JOIN against the `recipes` table.
- Cook Mode single-step view scrolls from the top of long step text and still vertically centers short content.
- `recipe-service` no longer JOINs the local `ingredient_categories` table, the orphaned table is dropped via migration, and the Ingredients list resolves category names client-side from `useIngredientCategories`.

## Implementation Phases

### Phase A — Low-risk frontend-only fixes (parallel)

These can land independently and in any order.

#### A1. Fix `isTaskOverdue` date comparison (Item 1) — S
- Edit `frontend/src/types/models/task.ts:45-51`:
  - Parse `dueOn` as `YYYY-MM-DD` → `new Date(year, month - 1, day)` (local).
  - Build `today = new Date(); today.setHours(0,0,0,0)`.
  - Return `dueDate < today`.
- Create `frontend/src/types/models/__tests__/task.test.ts` (if missing) covering: today, yesterday, tomorrow, completed-with-past-due, null `dueOn`.
- **Acceptance:** all five test cases pass; manual smoke on Tasks page shows due-today tasks as "pending".

#### A2. Cursor affordances on primitives (Items 2 + 4) — S
- Add `cursor-pointer` to the base class string in `frontend/src/components/ui/button.tsx` `buttonVariants`.
- Add `cursor-pointer` to the `DialogClose`/X button in `frontend/src/components/ui/dialog.tsx`.
- ~~Verify `frontend/src/components/ui/sheet.tsx`~~ — verified: file does not exist in this codebase, no action.
- Patch the bare `<button>` in `IngredientsPage.tsx:185-193` (search-input "X") with `cursor-pointer`.
- Verify `disabled:` selector still neutralizes hover (do not show pointer on disabled).
- Snapshot risk verified: `frontend/src/components/ui/__tests__/` only contains `user-avatar.test.tsx` — no `Button`/`Dialog` snapshot tests to update.
- **Acceptance:** every surface listed in PRD §4.2 hovers as pointer; disabled buttons remain non-pointer; no per-call-site `cursor-pointer` additions.

#### A3. Context-sensitive complete-toggle icon (Item 3) — S
- In `frontend/src/pages/TasksPage.tsx:101-118` and `frontend/src/components/features/tasks/task-card.tsx`:
  - Import `Circle, CheckCircle2` from `lucide-react`.
  - Render `Circle` (muted) when incomplete, `CheckCircle2` (primary) when completed.
  - `aria-label`: `"Mark complete"` / `"Mark incomplete"`.
- Leave existing toasts unchanged.
- **Acceptance:** desktop and mobile show identical paired-icon behavior; aria-label flips with state.

#### A4. All-day events full-width (Item 5) — S
- In `frontend/src/components/features/calendar/all-day-event-row.tsx`, replace `flex flex-wrap gap-1` with a vertical-stack layout where each event is `w-full block` and `truncate`.
- Preserve `min-h-[28px]`, `px-1 py-1`, and the popover-on-click handler.
- Per-day rendering passed in from `calendar-grid.tsx:100-104` is unchanged — each day's instance of a multi-day event is independently full-width.
- **Acceptance:** PRD §10 acceptance for item 5 — full-width, multi-day spans, truncation.

#### A5. Cook Mode single-step view scroll fix (Item 7) — S
- In `frontend/src/components/features/recipes/cook-mode.tsx:191-210`, restructure the scroll container per PRD §4.6:
  ```tsx
  <div className="flex-1 overflow-y-auto px-6 md:px-12">
    <div className="min-h-full flex flex-col items-center justify-center py-6">
      {/* section header + step text */}
    </div>
  </div>
  ```
- Verify keyboard arrow nav, swipe handlers, and step counter are unchanged.
- **Acceptance:** long step text is scrollable from the top on both mobile and desktop; short steps remain vertically centered.

### Phase B — Backend changes for "Used in Recipes" (Item 6)

#### B1. Backend: extend `RecipeRef` with `recipeName` — M
- In `services/recipe-service/internal/ingredient/provider.go`:
  - Add `RecipeName string \`json:"recipeName"\`` to `RecipeRef` (line 88-91).
  - In `getIngredientRecipes` (line 93-103), JOIN against `recipes` on `recipe_id` and select `recipes.title AS recipe_name`. Verified column: `recipes.title` (`services/recipe-service/internal/recipe/entity.go:15`).
  - Add `recipes.deleted_at IS NULL` to exclude soft-deleted recipes (`recipes` uses `DeletedAt`, see `recipe/entity.go:23`).
  - Tenant scoping: the JOIN itself is tenant-safe (recipe → recipe_ingredients are co-owned). The handler-level tenant check on the URL `id` is a pre-existing gap **out of scope** for this task — see context.md "Tenant-scoping note".
- Update the JSON:API response (`resource.go:285-292` writes `refs` directly into `data` — the new `recipeName` field will appear automatically once the struct tag is added). Preserve `rawName`.
- Add a Go test (extend `processor_integration_test.go`) verifying the JOIN returns the recipe name and that an ingredient used in two recipes returns both names.
- **Acceptance:** `GET /ingredients/{id}/recipes` returns `recipeName` for each ref; tests pass.

#### B2. Frontend: render recipe names — S
- Add `recipeName: string` to `IngredientRecipeRef` in `frontend/src/types/models/ingredient.ts:55-58`.
- In `frontend/src/pages/IngredientDetailPage.tsx:241-261`, render `{ref.recipeName}` instead of `{ref.rawName} (recipe)`. Fall back to `"(unknown recipe)"` when empty.
- **Acceptance:** ingredient detail page shows recipe titles; clicking still navigates to `/app/recipes/{recipeId}`.

### Phase C — Stale "uncategorized" backend cleanup (Item 8)

#### C1. Backend: drop `categoryName` end-to-end from `recipe-service` ingredient layer — M
Verified via grep: the only readers/writers of `categoryName` in `recipe-service` are the call sites listed below. Per decision 4 in context.md, fully delete the field.
- `provider.go:147-149` — remove `LEFT JOIN ingredient_categories` and the `category_name` SELECT.
- `provider.go:115-124` — drop `CategoryName` field from `entityWithUsage` and `entityWithCategory`.
- `processor.go:234` — drop the `.WithCategoryName(e.CategoryName)` call from `SearchWithUsage`.
- `model.go:43,57,69` — delete `categoryName` field, `CategoryName()` getter, `WithCategoryName()` setter.
- `builder.go:22,38,59` — delete `categoryName` field, `SetCategoryName` method, builder copy line.
- `rest.go:15,32` — remove `CategoryName` JSON field from list / detail response structs.
- `rest.go:126,139` — remove `CategoryName: m.CategoryName()` from serializer call sites.
- **Acceptance:** `GET /ingredients` no longer JOINs `ingredient_categories`; `categoryId` is still returned; `go test ./...` passes.

#### C2. Backend: drop the orphaned `ingredient_categories` table — S
- In `services/recipe-service/internal/ingredient/entity.go`'s `Migration` function, after `AutoMigrate`, add `db.Exec("DROP TABLE IF EXISTS ingredient_categories")`.
- Verify via `Grep` that no other reference exists in `recipe-service` after C1.
- **Acceptance:** migration runs cleanly; no references to `ingredient_categories` remain in the service.

#### C3. Frontend: resolve category names client-side — S
- In `frontend/src/pages/IngredientsPage.tsx:249-253`, look up the name from the `categories` array (already loaded via `useIngredientCategories`) keyed by `ingredient.attributes.categoryId`. Mirror `IngredientDetailPage.tsx:188-189`.
- Drop `categoryName` from `CanonicalIngredientListItem` in `frontend/src/types/models/ingredient.ts`.
- Verify "uncategorized count" derivation in `IngredientsPage.tsx:64-71` still works (it relies on `total` and category counts, not `categoryName`).
- **Acceptance:** newly-categorized ingredients show their category on the list page without a hard refresh; category persists across hard refresh.

### Phase D — Verification

#### D1. Build & test all affected services — M
- `frontend`: `npm run build`, `npm test`.
- `recipe-service`: Docker build per CLAUDE.md guidance, plus `go test ./...`.
- Manual smoke: walk through all eight acceptance criteria from PRD §10.

## Risk Assessment & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `Button` `cursor-pointer` change cascades and shows pointer on disabled buttons | Low | Low | Verify `disabled:cursor-default` or rely on `disabled:pointer-events-none`; manually hover a disabled button before sign-off. |
| Snapshot tests on UI primitives break | Med | Low | Update snapshots; review the diff to ensure only `cursor-pointer` was added. |
| Recipe-name JOIN regresses query latency on `getIngredientRecipes` | Low | Low | Result set is already paginated and small; `recipes.id` is the PK. No new index required. |
| Dropping `ingredient_categories` deletes data still used by another code path | Low | High | C1 must precede C2; rely on the existing PRD-noted code search confirming the only readers are the SELECT/JOIN being removed. Re-grep before adding the `DROP TABLE`. |
| Local-date parsing in `isTaskOverdue` misses an edge case (DST, year rollover) | Low | Low | Cover edge cases with unit tests; use the explicit `(year, month-1, day)` constructor instead of string parsing. |
| Item 8 frontend list drops `categoryName` before backend stops sending it (or vice versa) | Med | Low | Land C1+C3 in the same PR; type drop in `ingredient.ts` is the forcing function. |
| Multi-day all-day calendar event regressions | Low | Med | Test with a 3-day event spanning a week boundary in the manual smoke. |

## Success Metrics

- All eight PRD §10 acceptance criteria pass manually.
- New unit tests (`isTaskOverdue`, recipe-name JOIN) pass in CI.
- No regressions in existing `frontend` and `recipe-service` test suites.
- No remaining references to `ingredient_categories` in `recipe-service`.

## Required Resources & Dependencies

- **Codebases:** `frontend/`, `services/recipe-service/`.
- **Tooling:** Node + Vite (frontend), Go toolchain + Docker (recipe-service), `scripts/local-up.sh` for end-to-end smoke.
- **External services:** none (no `category-service` change).
- **Skills:** `frontend-dev-guidelines`, `backend-dev-guidelines` (apply per service).

## Timeline / Effort Estimates

| Phase | Items | Effort |
|---|---|---|
| A — Frontend-only | 1, 2/4, 3, 5, 7 | 5×S |
| B — Used in Recipes | 6 | 1×M + 1×S |
| C — Stale category cleanup | 8 | 1×M + 2×S |
| D — Verification | — | 1×M |

Total: ~7×S, 3×M. Single-developer scope. **Decision: ship as one bundled PR** (see context.md decision 7). Commit boundaries can mirror phases for review readability.
