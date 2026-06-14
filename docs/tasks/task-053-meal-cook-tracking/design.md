# Meal Cook Tracking & Frequency Sort — Design

Task: task-053-meal-cook-tracking
Status: Draft for review
Created: 2026-06-14
Source PRD: `docs/tasks/task-053-meal-cook-tracking/prd.md`

---

## 1. Summary

Turn the existing read-only "cook count" (number of `plan_items` referencing a recipe)
into a **sortable** dimension on the recipe list endpoint, and expose that sort in the two
recipe-browsing surfaces: the meal-planner recipe selector and the main Recipes page. The
count stays **computed on the fly** — no persisted counter, no migration.

The core backend change is moving the `plan_items` aggregation **from a post-pagination
enrichment into the ordered, paginated list query** so the full filtered result set is
ordered by frequency before `LIMIT/OFFSET` (FR-5). The frontend adds a sort control to each
surface and threads a `sort` query parameter through the existing React Query hook.

---

## 2. Current state (verified against source)

| Concern | Where | Behavior today |
| --- | --- | --- |
| Recipe list query | `recipe/provider.go:44` `getAll` | Filters by search/tag/classification/plannerReady/normalizationStatus; `ORDER BY created_at DESC`; paginates. Returns `[]Entity`. |
| Tenant scoping | `shared/go/database/tenant_callbacks.go:46` | GORM `Before("gorm:query")` callback injects `WHERE tenant_id = ?` **only** on models whose struct has a `TenantId`/`tenant_id` column. |
| `recipes` table | `recipe/entity.go:11` | Has `TenantId` **and** `HouseholdId`. List query is therefore auto-scoped by `tenant_id` only — **not** household. |
| `plan_items` table | `planitem/entity.go:10` | Has **no** `tenant_id`/`household_id`. Has `idx_plan_item_recipe` on `recipe_id`. Scoping flows through `plan_week_id → plan_weeks`. |
| `plan_weeks` table | `plan/entity.go:10` | Has both `TenantId` and `HouseholdId`, with composite index `idx_plan_week_tenant_household`. |
| Usage aggregation | `recipe/provider.go:127` `getRecipeUsageFromPlanItems` | `SELECT recipe_id, MAX(day), COUNT(*) FROM plan_items WHERE recipe_id IN (?) GROUP BY recipe_id`. **No `plan_weeks` join — unscoped**; safe today only because `recipe_id`s are already tenant-scoped. Runs on the post-pagination page slice. |
| List handler enrichment | `recipe/resource.go:106-125` | Calls `GetRecipeUsage` **only** when `include_usage=true`, on the current page. |
| REST attribute | `recipe/rest.go:26`, `ListEnrichment` `rest.go:75` | `usageCount int64 json:"usageCount,omitempty"`, `lastUsedDate` already defined. |
| Frontend list service | `frontend/src/services/api/recipe.ts` `listRecipes` | Builds query params; **never sends `include_usage`**. |
| Frontend types | `frontend/src/types/models/recipe.ts` | `RecipeListAttributes.usageCount?: number` and `lastUsedDate?: string` already typed. No change needed. |
| Selector | `frontend/src/components/features/meals/recipe-selector.tsx:96` | Already renders `used {usageCount}x` and `last used {date}` — but these are **dormant** because the list service never requests usage. `pageSize: 50`, single page, no pagination UI. |
| Recipes page | `frontend/src/pages/RecipesPage.tsx:29` | Badge-based filter row; `useRecipes` with default page (1) / default pageSize (20); no pagination UI; no sort control. |
| Recipe card | `frontend/src/components/features/recipes/recipe-card.tsx` | Metadata row shows total time + tag badges; **no usage display**. |

**Key implication:** because the recipe list is scoped by `tenant_id` only (not household),
and `plan_items` carry no tenant column, the cook-count query must join `plan_weeks` to scope
explicitly. See §5 for the scoping decision.

---

## 3. Architecture decisions

### 3.1 Aggregate via LEFT JOIN on a grouped derived table (not a correlated subquery)

Two ways to order recipes by their `plan_items` count inside one paginated query:

- **A — Correlated scalar subquery in `SELECT`/`ORDER BY`.** Re-evaluated per candidate row;
  simpler SQL but no single place to also pull `MAX(day)`.
- **B — `LEFT JOIN (SELECT recipe_id, COUNT(*), MAX(day) ... GROUP BY recipe_id) u` (chosen).**
  Aggregation computed once over all `plan_items`, joined 1:1 to recipes, `COALESCE(count,0)`
  for never-scheduled recipes. Yields both the count and `last_used_day` in the same pass, and
  is exactly the shape the PRD suggests (§5 of PRD).

**Decision: B.** It computes the aggregate once, returns count + last-used together (so we
satisfy FR-9 and FR-11 without a second query), and the 1:1 join leaves the recipe row count
unchanged so `total` is unaffected.

### 3.2 Add the join only when frequency sort is requested

When no frequency sort is asked for, `getAll` keeps its current `ORDER BY created_at DESC`
and adds **no** join — existing callers and behavior are byte-for-byte unchanged (FR-8, FR-13).
The join, `Select`, and frequency `ORDER BY` are added only on the sort path.

### 3.3 Count definition unified across both code paths

The sort path's join becomes the single source of truth for the cook count and last-used date.
To avoid two divergent definitions (the join is `plan_weeks`-scoped; the legacy
`getRecipeUsageFromPlanItems` is unscoped), `getRecipeUsageFromPlanItems` is **also** updated to
join `plan_weeks` and apply the same tenant+household scoping. This keeps FR-11 exact (the
number displayed equals the number sorted on) and removes the latent cross-scope ambiguity in
the legacy `include_usage` path.

---

## 4. Backend design (`recipe-service`)

### 4.1 Sort parameter parsing — `recipe/resource.go`

`listHandler` parses `sort` and the requesting tenant/household:

```go
// in listHandler, alongside existing filter parsing
t := tenantctx.MustFromContext(r.Context())
filters := ListFilters{
    // …existing fields…
    TenantID:    t.Id(),
    HouseholdID: t.HouseholdId(),
    UsageSort:   parseUsageSort(r.URL.Query().Get("sort")),
}
```

```go
type UsageSort int

const (
    UsageSortNone UsageSort = iota // default order (created_at DESC)
    UsageSortAsc                   // least cooked first
    UsageSortDesc                  // most cooked first
)

func parseUsageSort(v string) UsageSort {
    switch v {
    case "usageCount":
        return UsageSortAsc
    case "-usageCount":
        return UsageSortDesc
    default:
        return UsageSortNone // lenient: unknown value → default (PRD §5.3)
    }
}
```

Token choice (resolves PRD Open Question §9): **`usageCount` / `-usageCount`**, matching the
existing JSON:API attribute (`rest.go:26`) and the already-typed frontend field.

### 4.2 `ListFilters` additions — `recipe/provider.go:34`

```go
type ListFilters struct {
    // …existing fields…
    TenantID    uuid.UUID
    HouseholdID uuid.UUID
    UsageSort   UsageSort
}
```

`TenantID`/`HouseholdID` are needed because the `plan_weeks` scoping in the join is **not**
applied by the tenant callback (the join targets `plan_items`/`plan_weeks`, not the `Entity`
model), so it must be passed explicitly.

### 4.3 Ordered query — `recipe/provider.go` `getAll`

`total` is still counted on the filtered set **before** any join (unchanged), so it is correct
and the 1:1 join would not affect it anyway. The `Find` branch diverges on `UsageSort`:

```go
// total: unchanged — count filtered recipes (no join)
var total int64
if err := query.Count(&total).Error; err != nil { return nil, nil, 0, err }

offset := (filters.Page - 1) * filters.PageSize

if filters.UsageSort == UsageSortNone {
    // EXISTING PATH — unchanged
    var entities []Entity
    err := query.Preload("Tags").Order("created_at DESC").
        Offset(offset).Limit(filters.PageSize).Find(&entities).Error
    return entities, nil, total, err
}

// FREQUENCY SORT PATH
usageSub := db.Table("plan_items AS pi").
    Select("pi.recipe_id AS recipe_id, COUNT(*) AS usage_count, MAX(pi.day) AS last_used_day").
    Joins("JOIN plan_weeks AS pw ON pw.id = pi.plan_week_id").
    Where("pw.tenant_id = ? AND pw.household_id = ?", filters.TenantID, filters.HouseholdID).
    Group("pi.recipe_id")

dir := "ASC"
if filters.UsageSort == UsageSortDesc { dir = "DESC" }
// tie-breaker (FR-7): stable across pages
order := fmt.Sprintf("COALESCE(u.usage_count, 0) %s, recipes.title ASC, recipes.id ASC", dir)

var rows []recipeWithUsage
err := query.
    Joins("LEFT JOIN (?) AS u ON u.recipe_id = recipes.id", usageSub).
    Select("recipes.*, COALESCE(u.usage_count, 0) AS usage_count, u.last_used_day").
    Preload("Tags").
    Order(order).
    Offset(offset).Limit(filters.PageSize).
    Find(&rows).Error
// split rows → ordered []Entity + usage map keyed by recipe id
```

New scan struct (embeds `Entity` so `Preload("Tags")` still works via `recipes.id`):

```go
type recipeWithUsage struct {
    Entity
    UsageCount  int64   `gorm:"column:usage_count"`
    LastUsedDay *string `gorm:"column:last_used_day"`
}
```

`getAll` signature becomes:

```go
func getAll(filters ListFilters) func(db *gorm.DB) ([]Entity, map[uuid.UUID]recipeUsageResult, int64, error)
```

returning `(orderedEntities, usageMap, total, err)`. `usageMap` is `nil` on the no-sort path.

**Column-qualification note (implementation guard):** the existing filter clauses use bare
`id`, `title`, `servings`, `deleted_at`. The derived table `u` exposes only `recipe_id`,
`usage_count`, `last_used_day`, and `recipes` is the only outer table with those other columns,
so no ambiguity arises. The tenant callback's injected `tenant_id = ?` likewise resolves to
`recipes.tenant_id` (only `recipes` has it at the outer level). Implementer should verify
generated SQL once and qualify `recipes.*` defensively if any DB complains.

### 4.4 Processor + handler wiring

- `Processor.List` (`processor.go:153`) returns the extra `usageMap`:
  `func (p *Processor) List(filters ListFilters) ([]Model, map[uuid.UUID]recipeUsageResult, int64, error)`.
  Models are built in slice order, preserving the SQL ordering (FR-5).
- `listHandler` (`resource.go:99`): when `usageMap != nil` (sort active), merge
  `LastUsedDate`/`UsageCount` into `enrichments` **regardless of `include_usage`** (FR-9). The
  existing `include_usage=true` branch stays for the non-sort case (now backed by the
  `plan_weeks`-scoped `getRecipeUsageFromPlanItems`).

### 4.5 Scoping of `getRecipeUsageFromPlanItems` — `provider.go:127`

Add the `plan_weeks` join + tenant/household filter so the legacy `include_usage` path matches
the sort path's definition:

```go
db.Table("plan_items AS pi").
    Select("pi.recipe_id, MAX(pi.day) AS last_used_day, COUNT(*) AS usage_count").
    Joins("JOIN plan_weeks AS pw ON pw.id = pi.plan_week_id").
    Where("pw.tenant_id = ? AND pw.household_id = ? AND pi.recipe_id IN ?", tenantID, householdID, recipeIDs).
    Group("pi.recipe_id").Find(&rows)
```

`GetRecipeUsage`/`getRecipeUsageFromPlanItems` gain `tenantID, householdID uuid.UUID` params,
passed from the handler's tenant context.

---

## 5. Multi-tenancy / household scoping decision (FR-17)

**Decision: scope the cook count by `plan_weeks.tenant_id` AND `plan_weeks.household_id`** from
the requesting context.

Rationale and the tension involved (flagged for review):

- FR-17 and the acceptance criterion are explicit: *"Cook counts include only `plan_items` from
  the requesting household."* `plan_weeks` carries `household_id`, so scoping by it is free
  (covered by `idx_plan_week_tenant_household`).
- **Caveat the reviewer should be aware of:** the recipe *list itself* is scoped by `tenant_id`
  **only** (the callback ignores `HouseholdId`; `provider.go:44` adds no household filter). So if
  two households ever share one tenant **and** can see each other's recipes, a recipe owned by
  household A would display `usageCount: 0` to household B (B has never scheduled it) even though
  it appears in B's list. This is the literal, FR-17-faithful behavior ("the requesting
  household's count"), and it also strictly improves on today's *unscoped* aggregation. If
  product instead wants a tenant-wide count, drop the `household_id` predicate in both queries —
  a one-line change. This design picks household scoping because the PRD demands it explicitly.

Cross-tenant isolation (the practical guarantee the "cross-household test" verifies) holds
either way via the `plan_weeks.tenant_id` predicate.

---

## 6. Frontend design

### 6.1 Thread `sort` through the data layer

- `RecipeListParams` (`services/api/recipe.ts`) and `UseRecipesParams`
  (`hooks/api/use-recipes.ts`): add `sort?: RecipeSort | undefined`.
- `RecipeService.listRecipes`: `if (params?.sort) query.set("sort", params.sort)`. Do **not**
  send `include_usage` — the sort auto-populates `usageCount` (FR-9).
- `useRecipes` query key already spreads `params` (`use-recipes.ts:40`), so changing `sort`
  produces a distinct cache entry and re-queries the full set server-side (FR-12).
- New type in `types/models/recipe.ts`:
  `export type RecipeSort = "usageCount" | "-usageCount";`

### 6.2 Sort control (shared affordance, FR-18)

A small labeled `Select` with three options, used in both surfaces:

| Label | `sort` value | Meaning |
| --- | --- | --- |
| Default | `undefined` | created_at DESC (unchanged) |
| Most cooked | `-usageCount` | descending |
| Least cooked | `usageCount` | ascending |

Labels are direction-explicit per FR-18. Default selection = "Default" so initial behavior is
unchanged (FR-13). Sort state is component-local and **resets when the control unmounts**
(resolves PRD Open Question §9 "selector sort persistence": reset to default each open).

### 6.3 Recipe selector — `recipe-selector.tsx`

Add the sort `Select` next to the existing classification `Select` (line 53). Local
`const [sort, setSort] = useState<RecipeSort | undefined>(undefined)`, passed into the
`useRecipes(...)` call (line 22). The existing `used {usageCount}x` / `last used {date}`
rendering (lines 93-98) lights up automatically once sort is active. `pageSize: 50` stays;
server-side ordering before `LIMIT` means the top 50 by frequency are returned (FR-12; resolves
PRD Open Question §9 "pagination in the selector": single capped page, sort is server-side).

### 6.4 Recipes page — `RecipesPage.tsx`

Add the same sort `Select` to the filter area (near the search box, line 77, so it reads as a
top-level control rather than a tag toggle). Local `sort` state passed to `useRecipes` (line 29).

### 6.5 Recipe card count display — `recipe-card.tsx` (FR-15)

Show a `cooked {usageCount}x` indicator in the metadata row (line 88, beside total time) **when
`usageCount` is present and > 0**. Since the list endpoint only returns `usageCount` when sort
is active, this naturally appears exactly when frequency sort is in effect — satisfying FR-15
("display the cook count when frequency sort is the active sort") without extra plumbing
(resolves PRD Open Question §9 "list-page count visibility": show-when-present).

---

## 7. Performance & migration (NFR)

- **Single paginated query**, not N+1: the aggregate is one grouped derived table joined once.
- **Indexes already exist** — `plan_items.idx_plan_item_recipe` (`planitem/entity.go:15`) and
  `plan_weeks.idx_plan_week_tenant_household` (`plan/entity.go:12`) cover the join and scope
  predicates. **No migration, no backfill, no denormalized counter** (PRD §6).
- `total` count query is untouched (no join), preserving its cost profile.
- **No shared-library changes** (`shared/go/database` is read but not modified), so no
  cross-service Docker rebuild is required — only `recipe-service`.

---

## 8. Testing strategy

### Backend (`recipe-service`)
- **Descending order:** most-cooked first across the full set.
- **Ascending order:** least-cooked first; zero-count recipes appear **first** and are not
  dropped (FR-2).
- **Tie-breaker stability (FR-7):** equal-count recipes ordered by `title` then `id`; paging
  through them (page 1 + page 2 with small `pageSize`) yields no duplicates or omissions.
- **Filter composition (FR-6):** sort + each of `search`, `tag[]`, `classification`,
  `plannerReady`, `normalizationStatus` filters the same set, only reordered.
- **Scoping (FR-17):** plan_items in another tenant **and** another household do not count
  (cross-tenant + cross-household cases).
- **FR-9:** `sort=-usageCount` returns `usageCount` in the response without `include_usage=true`.
- **FR-8:** no `sort` param → identical output/order to today.
- Unified definition: `include_usage=true` (no sort) returns the same per-recipe count as the
  sort path for the same data.

### Frontend (vitest + RTL)
- Selecting "Most cooked" / "Least cooked" calls `useRecipes` with `sort=-usageCount` /
  `usageCount` (assert the query param via mocked `recipeService.listRecipes`).
- Selector renders the sorted order from the mocked response and the displayed `used Nx`
  equals the `usageCount` driving the sort (FR-11).
- Recipes page renders the sort control and `RecipeCard` shows `cooked Nx` when `usageCount`
  is present.
- "Default" selection sends no `sort` param (FR-13).
- Any `lastUsedDate`/date assertions run green under `TZ=UTC` (CI is UTC).

### Build verification
- `recipe-service` Go build + tests; frontend build + tests (incl. `TZ=UTC`).
- Docker build for `recipe-service`.

---

## 9. Resolved open questions (PRD §9)

| Question | Decision |
| --- | --- |
| Sort token name | `usageCount` / `-usageCount` (matches existing attribute + FE type). |
| List-page count visibility | Show `cooked Nx` on the card **when present** (i.e., when frequency sort is active). |
| Selector sort persistence | Component-local; **resets to Default** each time the selector opens. |
| Pagination in selector | Selector loads a single capped page (`pageSize: 50`); sort is applied **server-side over the full set** before the cap. |

---

## 10. Out of scope (per PRD §2)

No "mark as cooked" action, no cooked/not-cooked state, no time-windowed/recency-weighted
counts, no dashboard widgets/charts, no leaderboards/export, no change to how `lastUsedDate`
is computed, and **no denormalized `cook_count`/`last_cooked` column**.

---

## 11. Touch-point checklist

Backend:
- `recipe/resource.go` — parse `sort`, pass tenant/household, merge sort-path usage into
  enrichments regardless of `include_usage`.
- `recipe/provider.go` — `UsageSort` type + parser, `ListFilters` fields, `getAll` join/order +
  new return shape + `recipeWithUsage` struct, `plan_weeks`-scope `getRecipeUsageFromPlanItems`.
- `recipe/processor.go` — `List` and `GetRecipeUsage` signature changes.

Frontend:
- `types/models/recipe.ts` — add `RecipeSort` type (attributes already include `usageCount`).
- `services/api/recipe.ts` — `sort` param on `RecipeListParams` + `listRecipes`.
- `lib/hooks/api/use-recipes.ts` — `sort` on `UseRecipesParams`.
- `components/features/meals/recipe-selector.tsx` — sort `Select` + state.
- `pages/RecipesPage.tsx` — sort `Select` + state.
- `components/features/recipes/recipe-card.tsx` — `cooked Nx` indicator.
