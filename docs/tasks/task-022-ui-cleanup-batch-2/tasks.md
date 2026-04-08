# UI Cleanup Batch 2 — Task Checklist

Last Updated: 2026-04-08

## Phase A — Frontend-only fixes

### A1. Fix `isTaskOverdue` date comparison (Item 1) — S
- [ ] Update `frontend/src/types/models/task.ts:45-51` to parse `dueOn` as a local date and compare against local-midnight `today`.
- [ ] Create `frontend/src/types/models/__tests__/task.test.ts` (if missing).
- [ ] Add unit tests: today / yesterday / tomorrow / completed-with-past-due / null `dueOn`.
- [ ] Run `npm test` — all tests pass.
- [ ] Manual: due-today task displays as "pending" on Tasks page.

### A2. Cursor affordances on primitives (Items 2 + 4) — S
- [ ] Add `cursor-pointer` to `buttonVariants` base in `frontend/src/components/ui/button.tsx`.
- [ ] Add `cursor-pointer` to `DialogClose`/X in `frontend/src/components/ui/dialog.tsx`.
- [x] ~~Verify `sheet.tsx`~~ — file does not exist in this codebase.
- [ ] Patch bare `<button>` search "X" in `IngredientsPage.tsx:185-193` with `cursor-pointer`.
- [ ] Verify disabled buttons do not show pointer cursor.
- [x] ~~Update broken snapshot tests~~ — verified: only `user-avatar.test.tsx` exists, no risk.
- [ ] Confirm no per-call-site `cursor-pointer` overrides were added.
- [ ] Manual hover smoke: row actions, "Clear Filters", page header buttons, dialog X, dialog action buttons.

### A3. Context-sensitive complete-toggle icon (Item 3) — S
- [ ] In `TasksPage.tsx:101-118`, swap static `Check` for `Circle` (incomplete) / `CheckCircle2` (completed).
- [ ] Apply identical change in `frontend/src/components/features/tasks/task-card.tsx`.
- [ ] Set `aria-label` to `"Mark complete"` / `"Mark incomplete"` based on state.
- [ ] Verify existing toasts unchanged.
- [ ] Manual: toggle a task on desktop and mobile; icon and aria-label flip.

### A4. All-day events full-width (Item 5) — S
- [ ] Replace `flex flex-wrap gap-1` in `all-day-event-row.tsx` with vertical-stack `w-full block` events.
- [ ] Preserve `min-h-[28px]`, `px-1 py-1`, popover-on-click handler.
- [ ] Verify `truncate` on long titles (no wrapping).
- [ ] Manual: single-day event spans column width; multi-day event spans every covered day; long title truncates.

### A5. Cook Mode single-step view scroll fix (Item 7) — S
- [ ] Restructure `cook-mode.tsx:191-210` `SingleStepView` per PRD pattern (`min-h-full` inner wrapper).
- [ ] Remove `justify-center` from outer scroll container.
- [ ] Verify keyboard arrow nav, swipe handlers, step counter unchanged.
- [ ] Manual: long step text scrolls from top on mobile and desktop; short text remains vertically centered.

## Phase B — Used in Recipes (Item 6)

### B1. Backend: extend `RecipeRef` with `recipeName` — M
- [ ] Add `RecipeName string \`json:"recipeName"\`` to `RecipeRef` in `provider.go:88-91`.
- [ ] In `getIngredientRecipes`, `JOIN recipes ON recipes.id = recipe_ingredients.recipe_id`, select `recipes.title AS recipe_name`.
- [ ] Add `recipes.deleted_at IS NULL` to exclude soft-deleted recipes.
- [x] ~~Tenant scoping check~~ — JOIN is tenant-safe by construction; handler-level URL-id tenant gap is pre-existing and **out of scope** (see context.md tenant-scoping note).
- [ ] No serializer change needed — `resource.go:285-292` writes `refs` directly into `data`; the new struct field appears automatically.
- [ ] Add Go test in `processor_integration_test.go`: ingredient used in N recipes returns N names.
- [ ] `go test ./...` passes; Docker build succeeds.

### B2. Frontend: render recipe names — S
- [ ] Add `recipeName: string` to `IngredientRecipeRef` in `frontend/src/types/models/ingredient.ts:55-58`.
- [ ] In `IngredientDetailPage.tsx:241-261`, render `{ref.recipeName}` (fallback `"(unknown recipe)"`).
- [ ] Manual: ingredient detail page shows recipe titles; click navigates to `/app/recipes/{recipeId}`.

## Phase C — Stale "uncategorized" cleanup (Item 8)

### C1. Backend: drop `categoryName` end-to-end (verified call sites) — M
- [ ] `provider.go:147-149` — remove `LEFT JOIN ingredient_categories` and `category_name` SELECT.
- [ ] `provider.go:115-124` — drop `CategoryName` from `entityWithUsage` and `entityWithCategory`.
- [ ] `processor.go:234` — remove `.WithCategoryName(e.CategoryName)` from `SearchWithUsage`.
- [ ] `model.go:43,57,69` — delete `categoryName` field, `CategoryName()` getter, `WithCategoryName()` setter.
- [ ] `builder.go:22,38,59` — delete `categoryName` field, `SetCategoryName` method, builder copy line.
- [ ] `rest.go:15,32` — remove `CategoryName` JSON field from list/detail response structs.
- [ ] `rest.go:126,139` — remove `CategoryName: m.CategoryName()` from serializer call sites.
- [ ] `go test ./...` passes.

### C2. Backend: drop orphaned `ingredient_categories` table — S
- [ ] Re-grep `recipe-service` for any remaining `ingredient_categories` references after C1.
- [ ] Add `db.Exec("DROP TABLE IF EXISTS ingredient_categories")` to `entity.go` `Migration` after `AutoMigrate`.
- [ ] Docker build succeeds and migration runs cleanly.

### C3. Frontend: client-side category resolution — S
- [ ] In `IngredientsPage.tsx:249-253`, look up category name via `useIngredientCategories` keyed by `ingredient.attributes.categoryId`.
- [ ] Drop `categoryName` from `CanonicalIngredientListItem` in `frontend/src/types/models/ingredient.ts`.
- [ ] Verify uncategorized count derivation in `IngredientsPage.tsx:64-71` still works.
- [ ] Manual: categorize an ingredient on detail page → list page shows category without hard refresh; persists across hard refresh.

## Phase D — Verification

### D1. Build & test all affected services — M
- [ ] `npm run build` (frontend) succeeds.
- [ ] `npm test` (frontend) passes.
- [ ] `go test ./...` (recipe-service) passes.
- [ ] Docker build for recipe-service succeeds (per CLAUDE.md).
- [ ] Walk PRD §10 acceptance criteria end-to-end via `scripts/local-up.sh`.
- [ ] All eight acceptance criteria signed off.
