# Meal Cook Tracking & Frequency Sort — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-06-14
---

## 1. Overview

Home Hub already records every recipe a household schedules into a weekly meal plan as a
`plan_item` row (`services/recipe-service/internal/planitem/entity.go`). Today the
`recipe-service` can *compute* a per-recipe "usage" signal on demand — a count of
`plan_items` and the most recent planned day — by aggregating that table
(`services/recipe-service/internal/recipe/provider.go:115`). The recipe selector in the
meal planner already displays this as "used Nx" and "last used {date}"
(`frontend/src/components/features/meals/recipe-selector.tsx:93`).

What is missing is the ability to *act* on that signal. A user filling a meal slot cannot
order candidate recipes by how often they've been cooked, so it's hard to either lean on
favorites or deliberately rotate in dishes they haven't made in a while. This feature turns
the existing read-only usage data into a first-class, **sortable** dimension — ascending and
descending — in the two places where a user browses recipes: the meal-slot recipe selector
and the main recipes list page.

Per product decision, **"scheduled" is treated as "cooked"**: scheduling a recipe into any
meal slot is the event that increments its cook count. The count is an **all-time total**, and
it remains **computed on the fly** from `plan_items` rather than being denormalized onto the
recipe row. This keeps the data self-correcting (editing or deleting plan items naturally
adjusts the count) and avoids a migration/backfill.

## 2. Goals

Primary goals:
- Let a user sort recipes by cook frequency, **both** most-cooked-first and least-cooked-first,
  when choosing a recipe for a meal slot.
- Provide the same frequency sort on the main recipes list page.
- Keep the cook count derived from `plan_items` (no persisted counter, no migration).
- Preserve correctness of the count under pagination — sorting by frequency must order the
  *entire* result set, not just the current page.

Non-goals:
- A separate "mark as cooked" action or any cooked/not-cooked state on plan items. (Scheduling
  == cooking for this feature.)
- Time-windowed or recency-weighted counts (e.g., "last 6 months"). The count is all-time.
- Dashboard widgets, charts, trend lines, or any analytics/visualization surface.
- Per-household leaderboards, sharing, or exporting statistics.
- Changing how `last used` date is computed or displayed (it already exists and stays as is).
- Denormalizing `cook_count` / `last_cooked` onto the `recipes` table.

## 3. User Stories

- As a meal planner, I want to sort the recipe picker for a slot by most-cooked, so I can
  quickly drop in a reliable favorite.
- As a meal planner, I want to sort the recipe picker by least-cooked, so I can rediscover
  recipes I haven't made in a while and add variety.
- As a household member browsing all recipes, I want to sort the recipes list by how often
  we've cooked each one, so I can see our staples and our neglected recipes at a glance.
- As a household member, I want the cook count I sort by to reflect every recipe in our
  collection (not just the current page), so the ordering is trustworthy.

## 4. Functional Requirements

### 4.1 Cook frequency definition
- FR-1: A recipe's **cook count** is the number of `plan_items` rows that reference it
  (`plan_items.recipe_id = recipe.id`), across all weekly plans, all-time. This matches the
  existing `usageCount` aggregation (`provider.go:115`).
- FR-2: A recipe never scheduled has a cook count of `0` and MUST be sortable/displayable as
  such (it must not be omitted from the list).
- FR-3: The count is computed at query time. Editing a plan item's recipe, deleting a plan
  item, or deleting a plan week MUST be reflected in the count on the next query with no extra
  bookkeeping.

### 4.2 Sorting — recipe-service list endpoint
- FR-4: The recipe list endpoint MUST accept a sort parameter selecting cook frequency in two
  directions: most-cooked-first (descending) and least-cooked-first (ascending). See §5 for the
  exact parameter shape.
- FR-5: When sorting by cook frequency, the ordering MUST apply to the full filtered result set
  **before pagination**. The current implementation enriches usage only for the already-paged
  page (`resource.go:106-125`); that is insufficient for sorting and MUST be changed so the
  aggregation participates in the ordered, paginated query (e.g., `LEFT JOIN` the per-recipe
  `plan_items` count, `COALESCE`d to 0, into the list query).
- FR-6: Frequency sort MUST compose with existing filters: `search`, `tag[]`, `classification`,
  `normalizationStatus`, and `plannerReady`. Sorting changes order only; it MUST NOT change
  which recipes match.
- FR-7: A deterministic tie-breaker MUST be defined for recipes with equal counts (e.g., by
  recipe `title` ascending, then `id`) so paginated results are stable across pages.
- FR-8: The endpoint MUST retain its current default ordering when no frequency sort is
  requested, so existing callers are unaffected.
- FR-9: When frequency sort is requested, the response MUST include each recipe's cook count
  (`usageCount`) without requiring the caller to also pass `include_usage=true` (so the UI can
  display the number it sorted by). The existing `include_usage` behavior remains available for
  callers that want usage without sorting.

### 4.3 Recipe selector (meal planner)
- FR-10: The recipe selector (`recipe-selector.tsx`) MUST offer a control to sort the candidate
  list by cook frequency, ascending and descending.
- FR-11: The selector already shows "used Nx" / "last used {date}"; this MUST continue to work
  and reflect the same count used for sorting.
- FR-12: Changing the sort MUST re-query / re-order the full candidate set (respecting the
  active search and classification filters), not just reorder the current page in memory, so the
  result is correct across pagination.
- FR-13: The selector's default behavior (when the user has not chosen a frequency sort) MUST be
  unchanged from today.

### 4.4 Recipes list page
- FR-14: The main recipes list page MUST offer a frequency sort control (ascending and
  descending) consistent in label/affordance with the selector's.
- FR-15: The recipes list page MUST display each recipe's cook count when frequency sort is the
  active sort (so the user sees the value driving the order). Showing it always is acceptable if
  it fits the existing card/row layout.
- FR-16: Frequency sort on the list page MUST compose with the page's existing
  search/tag/classification filters.

### 4.5 Cross-cutting
- FR-17: All cook-count queries MUST be scoped to the requesting tenant/household, consistent
  with existing recipe and plan-item scoping. A recipe's count MUST only include `plan_items`
  belonging to the same household.
- FR-18: Sort labels MUST make direction unambiguous to the user (e.g., "Most cooked" /
  "Least cooked" rather than a bare "frequency").

## 5. API Surface

Service: `recipe-service`. Endpoint: existing `GET /api/v1/recipes` (list).

### 5.1 New query parameter
Introduce a sort parameter following JSON:API sort conventions where practical. Proposed:

```
GET /api/v1/recipes?sort=usageCount        # least cooked first (ascending)
GET /api/v1/recipes?sort=-usageCount       # most cooked first (descending)
```

- `sort=usageCount` → ascending (0 / least-cooked first).
- `sort=-usageCount` → descending (most-cooked first).
- Absent / any other value → current default ordering (FR-8).
- The exact token (`usageCount` vs. `cookCount` vs. `usage`) is to be finalized in design; it
  MUST match the attribute name the UI reads. This PRD uses `usageCount` to align with the
  existing response attribute (`rest.go:26`).

Backend touch points (current state to change):
- `services/recipe-service/internal/recipe/resource.go:82-89` — parse the new `sort` param into
  `ListFilters`.
- `services/recipe-service/internal/recipe/provider.go` — the List query must `LEFT JOIN`/subquery
  the per-recipe `plan_items` count (`COALESCE(count,0)`), order by it with a tie-breaker, and
  paginate after ordering (FR-5, FR-7).
- `services/recipe-service/internal/recipe/resource.go:106-125` — ensure `usageCount` is
  populated in the response when sorting by frequency, independent of `include_usage` (FR-9).

### 5.2 Response shape (unchanged structure)
JSON:API `recipes` resources. Each item's attributes already include:

```jsonc
{
  "type": "recipes",
  "id": "<uuid>",
  "attributes": {
    "title": "…",
    "usageCount": 7,                 // count used for sorting (0 when never scheduled)
    "lastUsedDate": "2026-05-30",    // existing, optional
    // …existing recipe attributes…
  }
}
```

Meta block (`total`, `page`, `pageSize`) is unchanged (`resource.go:146`).

### 5.3 Error cases
- Unknown `sort` value: do not error; fall back to default order (lenient, matches existing
  permissive query handling).
- No new error codes introduced.

No changes to the plan-item write endpoints (`POST/PATCH/DELETE
/api/v1/meals/plans/{planId}/items`) — scheduling already creates the `plan_items` rows the
count reads from.

## 6. Data Model

**No schema changes.** The cook count is derived from the existing `plan_items` table:

- `plan_items.recipe_id` → `recipes.id` (existing relationship).
- Count = `COUNT(plan_items)` grouped by `recipe_id`, scoped to the household.

This is the decision recorded in scoping: **stay computed, no persisted counter, no migration,
no backfill.** Aggregation already exists at `provider.go:115` (`getRecipeUsageFromPlanItems`)
and must be lifted into the orderable list query for sorting (it currently runs only on the
post-pagination page slice).

Multi-tenancy: all queries continue to scope by `tenant_id` / `household_id` consistent with the
service's existing data access (FR-17).

## 7. Service Impact

### `recipe-service` (Go)
- Recipe list query gains an optional frequency ordering backed by a `plan_items` aggregation
  joined into the main query (not a post-pagination enrichment).
- `ListFilters` gains a sort field; `resource.go` parses the `sort` param.
- Response enrichment ensures `usageCount` is present when sorting by frequency.
- No new entities, no migration.

### `frontend` (TypeScript / React)
- `frontend/src/components/features/meals/recipe-selector.tsx` — add a sort control (most/least
  cooked); pass the sort to the recipes query.
- Recipes list page (main browse) — add the same sort control and surface the count.
- `frontend/src/lib/hooks/api/use-recipes.ts` — thread the `sort` parameter through.
- `frontend/src/types/models/recipe.ts` — `usageCount` / `lastUsedDate` already typed
  (`recipe.ts:59-74`); confirm no change needed.

### No impact
- Other services (auth, calendar, dashboard, etc.) are untouched.
- Plan-item creation flow is untouched.

## 8. Non-Functional Requirements

- **Performance:** Sorting by frequency must remain a single paginated query. The `plan_items`
  aggregation should be a join/subquery, not N+1 per recipe. For the expected household-scale
  data volumes this is acceptable; if profiling shows cost, an index on
  `plan_items(recipe_id)` (and household scoping columns) may be added — but no denormalized
  counter (per decision).
- **Correctness under pagination:** Ordering happens before `LIMIT/OFFSET`; ties broken
  deterministically (FR-7) so no recipe is duplicated or skipped across pages.
- **Multi-tenancy:** Counts include only same-household `plan_items` (FR-17).
- **Backward compatibility:** Default ordering and all existing query params behave exactly as
  before when no frequency sort is requested (FR-8, FR-13).
- **Observability:** Reuse existing service logging; no new metrics required.
- **Testing:** Backend tests for ascending/descending order, zero-count recipes included,
  tie-breaker stability across pages, and filter composition. Frontend tests for the sort
  control and that the displayed count matches the sort. Any date/timezone-sensitive assertions
  must pass under `TZ=UTC` (CI runs UTC).

## 9. Open Questions

- **Sort token name:** Finalize the query token (`usageCount` vs `cookCount` vs `usage`) and the
  matching UI label set during design. PRD assumes `usageCount` to match the existing attribute.
- **List-page count visibility:** Always show the cook count on the recipes list page, or only
  when frequency sort is active? (FR-15 allows either; pick during design based on layout.)
- **Selector sort persistence:** Should the chosen sort direction persist across slots / sessions,
  or reset to default each time the selector opens? (Default assumption: resets to default.)
- **Pagination in the selector:** Confirm whether the selector currently paginates or loads a
  capped set; if capped, ensure frequency sort still operates server-side over the full set
  (FR-12).

## 10. Acceptance Criteria

- [ ] `GET /api/v1/recipes?sort=-usageCount` returns recipes ordered most-cooked first across the
      entire filtered set; `sort=usageCount` returns least-cooked first.
- [ ] Recipes never scheduled appear with `usageCount: 0` and are included in both sort
      directions (not dropped).
- [ ] Frequency sort composes correctly with `search`, `tag[]`, `classification`,
      `normalizationStatus`, and `plannerReady` filters.
- [ ] Ordering is applied before pagination; results are stable across pages via a deterministic
      tie-breaker (no duplicates or omissions when paging through equal-count recipes).
- [ ] `usageCount` is present in the response when sorting by frequency without requiring
      `include_usage=true`.
- [ ] Cook counts include only `plan_items` from the requesting household (verified by a
      cross-household test).
- [ ] Default list ordering and existing query params are unchanged when no frequency sort is
      passed.
- [ ] The meal-planner recipe selector offers most-cooked / least-cooked sorting; selecting
      either reorders the full candidate set (respecting active search/classification), and the
      displayed "used Nx" matches the sort value.
- [ ] The recipes list page offers the same frequency sort and surfaces the cook count.
- [ ] Backend builds and tests pass for `recipe-service`; frontend builds and tests pass,
      including under `TZ=UTC`.
- [ ] Docker build verified for `recipe-service` (no shared-library changes expected, but verify
      if any are made).
