# Context — Task 024

Last Updated: 2026-04-08

This file captures the codebase context that informed the PRD and the plan, so the implementer doesn't have to re-derive it.

> **Plan note (2026-04-08):** During plan creation, `services/recipe-service/internal/export/markdown.go:96-156` was re-read and found to ALREADY group the shopping list by category with an "Uncategorized" tail. The PRD §4.4 framing ("currently does NOT group by category") is therefore stale at the code level — the *grouping* exists, but it is **duplicated** between `markdown.go` and `processor.go`'s final-pass sort. Phase 4 of `plan.md` extracts a shared `GroupByCategory` helper to eliminate that drift risk, which is the actually-load-bearing piece of §4.4.

## Symptom

In the deployed environment, the meal plan "consolidated ingredients" preview shows every ingredient in a single ungrouped block (effectively all under "Uncategorized"), even though the same canonical ingredients are visibly categorized in the canonical ingredient admin and in the shopping list view. There are no errors in the browser console, so the failure is server-side and silent.

## Pipeline (already wired, end-to-end)

Frontend grouping (already implemented):
- `frontend/src/components/features/meals/ingredient-preview.tsx:22-70` — `useMemo` builds `CategoryGroup[]` from `data.attributes.category_name` / `category_sort_order`, sorts groups by sort order, and appends a final "Uncategorized" group for items with no category. Renders one `<h4>` per group. **No bug suspected here.**
- `frontend/src/types/models/meal-plan.ts` — `PlanIngredient` already declares `category_name` and `category_sort_order` fields.

Backend producer (already implemented):
- `services/recipe-service/internal/plan/resource.go:65` — route `GET /meals/plans/{planId}/ingredients` is wired to `getIngredientsHandler`, which calls `proc.ConsolidateIngredients(id, accessTokenCookie(r), catClient)` and marshals via `export.TransformIngredientSlice`.
- `services/recipe-service/internal/plan/processor.go:216` — `ConsolidateIngredients` delegates to `export.NewProcessor(...).ConsolidateIngredients(...)`.
- `services/recipe-service/internal/export/processor.go:65` — `Processor.ConsolidateIngredients` does the real work:
  - Lines 124-136: builds `categoryByID` lookup map by calling `p.catClient.ListCategories(pd.AccessToken)`. **The error from `ListCategories` is silently swallowed by an `if … err == nil` guard — no log statement on the error path.** This is suspect #1.
  - Lines 229-252: batch-fetches canonicals via `ingredientProc.GetByIDs(canonIDs)`, then for each accumulator looks up the canonical's `CategoryID()` in `categoryByID` to fill `acc.categoryName` / `acc.categorySortOrder`. **Missing canonical rows or missing category IDs are also silently dropped.** This is suspect #2 and #4.
- `services/recipe-service/internal/export/resource.go:11-67` — `RestIngredientModel` and `TransformIngredient` already serialize `category_name` and `category_sort_order` as nullable fields. `TransformIngredient` only sets them when `ci.CategoryName != ""`, which is correct.
- `services/recipe-service/internal/export/markdown.go` — markdown export. Currently does NOT group by category (this is the §4.4 parity gap).

Categoryclient:
- `services/recipe-service/internal/categoryclient/client.go:41` — `ListCategories` issues `GET {baseURL}/api/v1/categories` and forwards the user's `access_token` cookie. Decodes a JSON:API response into `[]Category{ID, Name, SortOrder}`.

Recipe-service wiring:
- `services/recipe-service/cmd/main.go:43,54` — `catClient := categoryclient.New(cfg.CategoryServiceURL)` and passed into `plan.InitializeRoutes(db, catClient)`. So `catClient` is non-nil at runtime as long as `CategoryServiceURL` is set in config.

## Auth boundary verification (question 4 from the design discussion)

`category-service`'s middleware setup (`services/category-service/cmd/main.go:33-38`) installs `sharedauth.Middleware`, which (`shared/go/auth/auth.go:114-147`):
- Reads the `access_token` cookie.
- Validates the JWT.
- Pulls `tenant_id` directly out of the JWT claims.
- Injects tenant context.

Conclusion: forwarding *only* the `access_token` cookie (which is what `categoryclient` does today) is sufficient for `category-service` to derive tenancy. **Tenant forwarding is NOT the bug**, and the PRD correctly rules it out.

## Top diagnostic suspects (ranked)

Based on the above, ordered most-to-least likely:

1. **Silent error from `categoryclient.ListCategories`** — `category-service` is unreachable, returning 401, returning a body the client can't decode, or some other error in the deployed env. The current `if err == nil` guard at `processor.go:131` swallows this entirely. Adding the error log from PRD §4.3 will likely diagnose and fix this in the same step.
2. **`canonical_ingredients.category_id` is NULL in the deployed DB** despite the admin UI appearing to assign categories. Would mean a write-path bug elsewhere. Quick check: `SELECT id, name, category_id FROM canonical_ingredients WHERE tenant_id = '<tenant>' LIMIT 20;` in the deployed recipe-service DB.
3. **Cross-tenant or deleted-category ID mismatch** — canonical references a category ID that no longer exists in the `category-service` response. Less likely but worth catching with the warn log from PRD §4.3.
4. **`ingredientProc.GetByIDs` returning a partial map** — least likely given the existing tests, but the warn log from PRD §4.3 will surface it.

## Useful files for the implementer

- `services/recipe-service/internal/export/processor.go` — primary fix site.
- `services/recipe-service/internal/export/markdown.go` — §4.4 parity work.
- `services/recipe-service/internal/categoryclient/client.go` — touch only if §4.1 diagnosis points here.
- `frontend/src/components/features/meals/ingredient-preview.tsx` — already correct; touch only if §4.1 says so.
- `docs/tasks/task-016-ingredient-categories/` — prior work that established the canonical ingredient → category linkage.
- `docs/tasks/task-019-shopping-list/` — prior work that uses category grouping in a similar UI; can be referenced for consistency.

## Decisions locked from the design discussion

| # | Question | Answer |
|---|---|---|
| 1 | Markdown export parity? | Yes — fix both preview and markdown export. |
| 2 | "Uncategorized" group when nothing categorized? | Yes — keep today's behavior. |
| 3 | Repro environment? | Deployed only. Not seen in local dev. No browser console errors. |
| 4 | Tenant forwarding? | Verified — cookie alone is sufficient. Not the bug. |
| 5 | Sort within group? | Alphabetical by display name (today's behavior). |
| 6 | Category service unreachable behavior? | Silent fallback to "Uncategorized" + server-side error log. No 5xx. |
| 7 | Task number? | task-024. |
