# Recipe Service Latency Fix — Implementation Plan

Last Updated: 2026-04-06

---

## Executive Summary

The recipe-service is hitting multi-second response times on the recipes list, recipe detail, and meal-planning pages. Production logs (Postgres slow-query log, threshold 200ms) show dozens of small per-row queries firing serially per request, with individual primary-key lookups landing between 200ms and 1300ms. The root cause is classic N+1 query patterns in three hot paths, amplified by an untuned Go database/sql connection pool that lets each in-flight request thrash the small default idle pool.

This plan eliminates the N+1 patterns by introducing batch (`WHERE … IN (?)`) fetch primitives in the `planner`, `normalization`, `ingredient`, and `recipe` packages, then rewires the three offending callers (`recipe.listHandler`, `plan.BuildItemsResponse`, `export.ConsolidateIngredients`) to fetch up-front and assemble responses in memory. It also tunes the shared GORM connection pool so individual queries stop suffering from connection churn.

The work is split into four phases: (1) connection pool tuning (smallest blast radius, largest immediate win), (2) batch fetch primitives, (3) caller refactors, (4) verification with query counts and slow-log monitoring. No API contract changes — this is a pure backend performance refactor.

---

## Current State Analysis

### What we observe in the logs

A single page load from the meal-planning UI produces a trace like:

```
plan_items WHERE plan_week_id = ?                                 (1 query)
recipe_planner_configs WHERE recipe_id = ?                        (per recipe)
recipe_ingredients WHERE recipe_id = ?                            (per recipe)
canonical_ingredient_aliases WHERE canonical_ingredient_id = ?    (per ingredient)
canonical_ingredients WHERE id = ?                                (per ingredient)
…repeat for every plan item, every recipe ingredient
```

For a 12-item plan with ~30 distinct ingredients this produces 100+ serial round trips. Individual queries log at 200ms–1300ms even for indexed point lookups, which is symptomatic of connection-pool starvation rather than query-planner issues.

### Identified N+1 sources

#### N+1 #1 — Recipes list page (`GET /recipes`)
`services/recipe-service/internal/recipe/resource.go:113-124` calls `proc.BuildListEnrichment(m)` once per recipe in the page. Each call (`internal/recipe/processor.go:371-389`) issues:

- `normProc.GetByRecipeID(m.Id())` → `recipe_ingredients WHERE recipe_id = ?`
- `plannerProc.GetByRecipeID(m.Id())` → `recipe_planner_configs WHERE recipe_id = ?`

A 20-recipe page = ~40 serial round trips.

#### N+1 #2 — Plan detail (`GET /meals/plans/{planId}`)
`plan.BuildItemsResponse` (`internal/plan/processor.go:257-301`) loops over plan items and calls per item:

- `recipeProc.Get(item.RecipeID())`
- `plannerProc.GetByRecipeID(item.RecipeID())`

12 items in a week = ~24 serial round trips.

#### N+1 #3 — Plan ingredients consolidation (`GET /meals/plans/{planId}/ingredients`)
The dominant offender. `export.ConsolidateIngredients` (`internal/export/processor.go:65-242`) does:

- Per plan item: `effectiveMultiplier` may invoke `plannerProc.GetByRecipeID` and `recipeProc.Get` (`processor.go:245-266`)
- Per plan item: `normProc.GetByRecipeID(item.RecipeID())`
- **Per resolved ingredient**: `ingredientProc.Get(canonID)` (`processor.go:158`), which uses `ingredient.GetByID` with `Preload("Aliases")` (`internal/ingredient/provider.go:12-16`) — this issues *two* serial queries per ingredient (one against `canonical_ingredients`, one against `canonical_ingredient_aliases`).

Net for a 12-item / 30-ingredient plan: ~100 serial round trips.

### Recipe detail (`GET /recipes/{id}`)
Comparatively healthy: `GetIngredients` (1 query) + `BuildDetailEnrichment` → `planner.GetByRecipeID` (1 query) + the recipe row itself ≈ 3 queries. It only feels slow because the DB is being saturated by the other endpoints above; once the big offenders are batched the detail page should require no further work.

### Why individual queries are slow

- **Untuned connection pool** in `shared/go/database/database.go:50`: GORM is opened with default settings, so the underlying `database/sql` pool uses the Go defaults (`MaxIdleConns=2`, `MaxOpenConns=0`/unlimited, no `ConnMaxLifetime`). With many serial queries this churns connections, causing connection-establishment overhead to be charged against query latency.
- **Reflection on every query** in `shared/go/database/tenant_callbacks.go:55-81`: `hasTenantIDField` walks struct fields with reflection on every Query/Update/Delete. Not the dominant cost, but multiplied across N+1 it compounds. Easily cached by `reflect.Type`.

---

## Proposed Future State

After implementation:

- `GET /recipes?page[size]=20` issues ~5 queries total regardless of page size: recipes page, tags, batched recipe_ingredients, batched planner_configs, optional batched usage.
- `GET /meals/plans/{id}` issues ~4 queries: plan, items, batched recipes, batched planner configs.
- `GET /meals/plans/{id}/ingredients` issues ~5 queries regardless of plan size: items, batched recipes, batched recipe_ingredients, batched planner_configs, batched canonical_ingredients (with aliases preloaded by IN).
- Individual query latency for indexed point lookups drops below the 200ms slow-log threshold; the `provider.go:18 SLOW SQL` log lines disappear from the recipe-service logs in the steady state.
- API contracts are unchanged; no frontend or migration work required.

---

## Implementation Phases

### Phase 1 — Connection Pool Tuning (Foundation)

Smallest change, biggest immediate win. Configure the underlying `database/sql` pool in the shared connect helper so all services benefit. This phase alone should noticeably drop baseline latency even before any N+1 batching lands, and gives a clean baseline to measure subsequent phases against.

**Key decisions:**
- Tune in `shared/go/database/database.go` so every service inherits sane defaults.
- Make the limits configurable via the existing `Config` struct (with sensible defaults) so individual services can tune if needed, but require no service-side changes to get the default.
- Defaults targeting a small Postgres (`max_connections` ~100): `MaxOpenConns=25`, `MaxIdleConns=25`, `ConnMaxLifetime=30m`, `ConnMaxIdleTime=5m`.

### Phase 2 — Batch Fetch Primitives

Add `…ByIDs` / `…ByRecipeIDs` providers that take a slice and emit a single `WHERE … IN (?)` query, returning a map keyed for O(1) lookup by callers. Each is independent and can be unit-tested in isolation. No callers are modified in this phase.

**New providers:**
- `planner.GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error)`
- `normalization.GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID][]Model, error)`
- `ingredient.GetByIDs(ids []uuid.UUID) (map[uuid.UUID]Model, error)` — uses `Preload("Aliases")`; GORM folds the preload into a single second query, so this collapses 2×N queries into exactly 2 total.
- `recipe.GetByIDs(ids []uuid.UUID) (map[uuid.UUID]Model, error)` — used by plan detail and consolidation.

**Conventions:**
- All map keys are the entity's primary or foreign key as appropriate.
- Empty input slice returns an empty map without hitting the DB.
- Tenant scoping continues to flow through the existing GORM tenant callbacks — no new tenant logic.
- Each new provider gets a focused unit test (or integration test where one already exists in the package).

### Phase 3 — Rewire Callers

Refactor the three offending request paths to use the new batch primitives. Each caller change is independent of the others and can be PR'd separately.

**3A: `export.ConsolidateIngredients`** — highest-impact, do first.
1. Collect all `recipeIDs` from plan items.
2. Call batched `normalization.GetByRecipeIDs`, `planner.GetByRecipeIDs`, `recipe.GetByIDs` once each.
3. Walk items in memory using the maps to compute multipliers and accumulate quantities.
4. Collect distinct `canonID`s across the entire plan after the walk; call `ingredient.GetByIDs` once.
5. Second walk fills in display name, unit family, and category from the canonical map.

**3B: `plan.BuildItemsResponse`**
1. Collect distinct `recipeIDs` from items.
2. Batch fetch `recipe.GetByIDs` and `planner.GetByRecipeIDs`.
3. Build `RestItemModel`s from the maps; no per-item DB calls.

**3C: `recipe.listHandler` / `BuildListEnrichment`**
1. After fetching the page of recipe models, collect their IDs.
2. Add `recipe.BuildListEnrichments(models []Model) []ListEnrichment` that calls `normalization.GetByRecipeIDs` and `planner.GetByRecipeIDs` once.
3. Replace the per-recipe loop in `listHandler` with a single call to the batch variant.
4. Keep `BuildListEnrichment` (singular) around only if other callers need it; otherwise delete to avoid bit-rot.

### Phase 4 — Verification & Monitoring

1. Add a structured log line at the end of each refactored handler reporting query count from a per-request counter (a tiny GORM callback can increment a context-bound counter; can be removed after rollout if too noisy).
2. Manually exercise the three pages and assert query count is in the expected single-digits range.
3. Watch the recipe-service logs for `SLOW SQL >= 200ms` entries; they should drop to near-zero in steady state.
4. Run the existing integration tests in `recipe-service` to confirm no regressions.

### Phase 5 (optional) — Tenant Callback Reflection Cache

`shared/go/database/tenant_callbacks.go:55-81` walks struct fields on every query. Add a `sync.Map` keyed by `reflect.Type` storing the `bool` result of `hasTenantIDField`. Free win, but only worth doing after Phase 1–4 land and the metrics are clean.

---

## Risk Assessment & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Pool limits set too low cause request queueing under load | Low | Medium | Defaults target small Postgres (~100 max connections); make the limits config-driven so they can be raised without a code change. |
| Pool limits set too high overwhelm Postgres `max_connections` | Low | High | Defaults are conservative; document the math (services × MaxOpenConns ≤ Postgres max). |
| Batch IN queries hit Postgres parameter limit (~32k) | Very Low | Low | All call sites batch by request scope (≤ a page of recipes / a week's plan items); no caller approaches the limit. |
| Refactor changes ordering / multiplicity semantics | Medium | High | Keep accumulator logic in `ConsolidateIngredients` byte-for-byte equivalent; only the data-fetch shape changes. Existing integration tests cover ordering. |
| Empty `IN ()` causes a SQL error or full-table scan | Medium | Medium | Each batch helper short-circuits to an empty map when input is empty. |
| GORM `Preload("Aliases")` with `IN` still issues one alias query per parent | Very Low | Medium | Verify with logging — GORM 2.x batches preloads into a single `IN` query. If not, fall back to a manual second query. |
| Tenant callback reflection cache breaks tests that assert struct introspection | Very Low | Low | Phase 5 is optional; defer if it causes any test churn. |

---

## Success Metrics

**Quantitative:**
- `GET /recipes?page[size]=20` query count drops from ~40 to ≤ 6.
- `GET /meals/plans/{id}` query count drops from ~25 to ≤ 5.
- `GET /meals/plans/{id}/ingredients` query count drops from ~100+ to ≤ 6 (independent of plan size).
- `SLOW SQL >= 200ms` log entries from recipe-service drop to ≤ 1/min in steady state.
- p95 wall-clock latency for the three endpoints drops by at least 80%.

**Qualitative:**
- Meal-planning UI feels responsive (subjective but the trigger for this work).
- Other services sharing the Postgres instance also benefit from reduced contention.

---

## Required Resources & Dependencies

- **Code:** recipe-service only; no frontend, migration, or shared-library breaking changes (Phase 1 touches `shared/go/database` but is additive).
- **Infra:** none. Postgres `max_connections` should be confirmed but no change is required.
- **Testing:** existing recipe-service integration tests cover the affected handlers; rely on them plus manual verification with query-count logging.
- **Rollout:** behind a single deploy. No feature flag needed since semantics are unchanged.

---

## Effort Sizing

| Phase | Effort | Notes |
|-------|--------|-------|
| 1 — Pool tuning | S | A few lines in `shared/go/database/database.go` plus optional `Config` plumbing. |
| 2 — Batch primitives | M | Four new providers, each with a small test. |
| 3A — Consolidate refactor | M | Largest individual change; needs careful preservation of accumulator semantics. |
| 3B — Plan items refactor | S | Mechanical. |
| 3C — Recipe list refactor | S | Mechanical. |
| 4 — Verification | S | Add query-count callback, manual testing, optionally remove the callback after. |
| 5 — Reflection cache | S | Optional. |

Total: M overall.
