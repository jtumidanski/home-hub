# Frontend Audit — task-053-meal-cook-tracking (cook-frequency sort)

- **Audit Scope:** Changed TS/React files in range `f8aff3a..a0b53fc` (cook-frequency sort feature)
- **Guidelines Source:** frontend-dev-guidelines skill
- **Date:** 2026-06-14
- **Build:** PASS
- **Tests:** 20 passed, 0 failed (5 changed test files)
- **Overall:** PASS

## Build & Test Results

- `npm run build` → built successfully (3434 modules; only pre-existing chunk-size warning).
- `vitest run` over the 5 changed test files → `Test Files 5 passed (5)`, `Tests 20 passed (20)`.
- `npm run lint` reports 9 errors, but ALL are in pre-existing untouched files (use-cooklang-preview.ts, DashboardDesigner.tsx, WorkoutReviewPage.test.tsx) — not attributable to this change.

## File Inventory

- `frontend/src/types/models/recipe.ts` — **Type** (adds `RecipeSort` union, line 59)
- `frontend/src/services/api/recipe.ts` — **Service** (threads `sort` param, lines 24, 52)
- `frontend/src/lib/hooks/api/use-recipes.ts` — **Hook** (adds `sort` to params, lines 6, 35)
- `frontend/src/components/features/recipes/recipe-sort-select.tsx` — **Component** (new shared sort select)
- `frontend/src/components/features/recipes/recipe-card.tsx` — **Component** (cooked Nx indicator, lines 93-97)
- `frontend/src/components/features/meals/recipe-selector.tsx` — **Component** (wires sort select, lines 22, 29, 69)
- `frontend/src/pages/RecipesPage.tsx` — **Page** (wires sort select, lines 25, 36, ~98)
- Test files: recipe-selector.test.tsx, recipe-card.test.tsx, recipe-sort-select.test.tsx, RecipesPage.test.tsx, recipe.test.ts — **Tests**

## Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | grep `: any`/`as any` over all changed files → zero matches |
| FE-02 | No manual class concatenation | PASS | `cn()` not needed; static class strings only (recipe-sort-select.tsx uses `className ?? "w-[140px]"` — a default fallback, not concatenation) |
| FE-03 | No direct API client calls in components | PASS | Components import `useRecipes` hook only (recipe-selector.tsx:8, RecipesPage.tsx:6); `@/lib/api/client` not imported in any changed component/page |
| FE-04 | No inline Zod schemas in components | PASS | No Zod in scope (feature has no forms) |
| FE-05 | No spinners for content loading | PASS | recipe-selector.tsx:73-76 uses `Skeleton`; no `animate-spin` in changed content |
| FE-06 | No hardcoded colors | PASS | New markup uses `text-muted-foreground` (recipe-card.tsx:94, recipe-sort-select.tsx). Pre-existing `text-green-500`/`text-yellow-500` in recipe-card.tsx:33,36 are NOT touched by this change |
| FE-07 | No state mutation | PASS | All sort state via `useState` + immutable setters (recipe-selector.tsx:22, RecipesPage.tsx:25); no `.push/.splice/.sort` introduced |
| FE-08 | No default exports for components | PASS | `export function RecipeSortSelect` (recipe-sort-select.tsx:12); all changed components use named exports |
| FE-09 | Tenant guard in hooks | PASS | use-recipes.ts:39 `useTenant()`; line 43 `enabled: !!tenant?.id && !!household?.id` (sort flows through existing guarded hook) |
| FE-10 | Tenant ID in query keys | PASS | recipeKeys.all includes `tenant?.id ?? "no-tenant"` (use-recipes.ts:14); list key spreads params incl. sort (line 41) so sort changes refetch correctly |
| FE-11 | Error handling with createErrorFromUnknown | PASS | No new catch blocks introduced; existing hook mutations use `getErrorMessage` + toast (use-recipes.ts:84, 100, 128) |

## Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-12 | JSON:API model shape | PASS | `RecipeSort` is a query-param union type, not a model; `RecipeListItem` keeps `{ id, type, attributes }` (recipe.ts:93-97); `usageCount?` added under attributes (recipe.ts:73) |
| FE-13 | Service extends BaseService | PASS | `class RecipeService extends BaseService` (recipe.ts:35); `listRecipes` uses documented direct-client pattern (`this.setTenant` + `api.get`) for custom query building — pre-existing, unchanged by this task |
| FE-14 | Query key factory uses `as const` | PASS | recipeKeys entries all `as const` (use-recipes.ts:14-22); list query key `[...] as const` (line 41) |
| FE-15 | Forms use react-hook-form + zodResolver | N/A | No forms in scope |
| FE-16 | Schema in lib/schemas with inferred type | N/A | No Zod schema in scope |

## Styling Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-19 | Interactive elements show cursor-pointer | PASS | Sort control is a shadcn `SelectTrigger`, which includes `cursor-pointer` in its base class (components/ui/select.tsx:42); `SelectItem` likewise (select.tsx:118). No raw clickable `<div>` introduced |

## Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PASS | recipe-sort-select.test.tsx (3 cases), recipe-card.test.tsx (+3 usageCount cases), recipe-selector.test.tsx (sort wiring), RecipesPage.test.tsx (sort wiring), recipe.test.ts (service sort param) — covers component, page, and service layers |
| FE-18 | Mocks updated when services changed | PASS | recipe.test.ts mocks `@/lib/api/client` and asserts query string contains/omits `sort=` (lines 16-39); hook-level mocks in component tests use `expect.objectContaining({ sort })` |

## Additional Verification (contract correctness)

- Backend contract confirmed against source: recipe-service reads `?sort` via `parseUsageSort` accepting exactly `usageCount` / `-usageCount` (`services/recipe-service/internal/recipe/usage_sort_test.go:10-11`, `provider.go:48-50`). Frontend `RecipeSort` union (recipe.ts:59) and emitted query string (recipe.ts:52) match exactly.
- `usageCount` JSON field matches backend `rest.go:26` (`UsageCount int64 json:"usageCount,omitempty"`). The card's `(attributes.usageCount ?? 0) > 0` guard (recipe-card.tsx:93) correctly handles the `omitempty`/undefined case.

## Summary

### Blocking (must fix)
- None.

### Non-Blocking (should fix)
- None.

Overall verdict: **PASS** — no Critical or Important issues. Build and all 20 tests green; the change is type-safe, tenant-scoped through the existing guarded hook, cache-correct (sort participates in the query key), uses semantic colors and named exports, and the sort values match the backend contract verified against source.
