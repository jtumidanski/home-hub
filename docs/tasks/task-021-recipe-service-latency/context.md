# Recipe Service Latency Fix — Context & Dependencies

Last Updated: 2026-04-06

---

## Trigger

Production logs from recipe-service show `SLOW SQL >= 200ms` lines firing many times per second on the recipes list, recipe detail, and meal-planning pages. Individual primary-key SELECT statements are landing at 200ms–1300ms each, and dozens of them stack up serially per request. The user reports the affected pages feel unresponsive in the web UI.

---

## Key Files — Backend (recipe-service)

### Recipe Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/internal/recipe/resource.go` | HTTP handlers including `listHandler` | Replace per-recipe `BuildListEnrichment` loop with batched `BuildListEnrichments`. |
| `services/recipe-service/internal/recipe/processor.go` | Includes `BuildListEnrichment`, `BuildDetailEnrichment` | Add `BuildListEnrichments(models []Model) []ListEnrichment` that batch-fetches ingredients + planner configs. |
| `services/recipe-service/internal/recipe/provider.go` | DB queries for recipes | Add `getByIDs(ids []uuid.UUID)` returning all matching rows with tags preloaded. |

### Plan Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/internal/plan/processor.go` | Includes `BuildItemsResponse` (lines 257-301) | Refactor to collect recipeIDs once and call batch `recipe.GetByIDs` and `planner.GetByRecipeIDs`. |
| `services/recipe-service/internal/plan/resource.go` | HTTP handlers | No structural change; only exercises the refactored processor. |

### Export Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/internal/export/processor.go` | `ConsolidateIngredients` (lines 65-242), `effectiveMultiplier` / `getServingsYield` (lines 245-266) | Largest refactor: batch fetch all recipes / planner_configs / recipe_ingredients up front, then walk items in memory; collect distinct canonical IDs across the whole plan and batch-fetch canonicals once. |

### Planner Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/internal/planner/provider.go` | DB queries for `recipe_planner_configs` | Add `getByRecipeIDs(recipeIDs []uuid.UUID)` returning all matches in one IN query. |
| `services/recipe-service/internal/planner/processor.go` | Wraps providers | Add `GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error)` returning a map keyed by recipe ID. |

### Normalization Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/internal/normalization/provider.go` | DB queries for `recipe_ingredients` | Add `getByRecipeIDs(recipeIDs []uuid.UUID)` returning all rows in one IN query. |
| `services/recipe-service/internal/normalization/processor.go` | Wraps providers | Add `GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID][]Model, error)` returning a map keyed by recipe ID. |

### Ingredient Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/internal/ingredient/provider.go` | `GetByID` uses `Preload("Aliases")` | Add `getByIDs(ids []uuid.UUID)` that uses the same preload and returns all matches in one IN query (GORM folds preload into a single second IN query). |
| `services/recipe-service/internal/ingredient/processor.go` | Wraps providers | Add `GetByIDs(ids []uuid.UUID) (map[uuid.UUID]Model, error)`. |

---

## Key Files — Shared

| File | Purpose | Changes Needed |
|------|---------|----------------|
| `shared/go/database/database.go` | `Connect` opens GORM with default pool settings | Tune `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `ConnMaxIdleTime` after `gorm.Open`. Optionally extend `Config` with pool fields and overridable defaults. |
| `shared/go/database/tenant_callbacks.go` | `hasTenantIDField` reflects on every query (lines 55-81) | Optional Phase 5: cache results by `reflect.Type` in a `sync.Map`. |
| `shared/go/database/provider.go` | Generic `Query` and `SliceQuery` helpers; line 18 / line 29 are where the slow-log call sites originate from | No changes; this is just where the slow logs are emitted from for context. |

---

## Key Decisions

- **No API contract changes.** This is a pure backend refactor. Frontend, schemas, and migrations are untouched.
- **Batching maps are returned, not slices.** Callers want O(1) lookup by recipe/canonical ID; converting from slices in every caller would defeat the purpose.
- **Tenant scoping stays in the existing GORM callbacks.** The new batch providers do not inject tenant filters manually — they rely on the same `tenant:query` callback already registered globally.
- **Empty input slices short-circuit.** Every `…ByIDs` helper returns an empty map without touching the DB if the input is empty, both to avoid `WHERE id IN ()` SQL errors and to keep the call site code simple.
- **Pool defaults are conservative.** `MaxOpenConns=25` × the number of services should sit comfortably under Postgres `max_connections`. Document this in `shared/go/database/database.go` so future services don't accidentally raise their limit blindly.
- **Connection-pool tuning ships first.** It is safe, high-leverage, and gives a clean baseline against which to measure the batching changes.
- **Reflection cache is optional and last.** Worth doing only after the metrics are clean enough to attribute any remaining latency to it.

---

## Dependencies & Ordering

- **Phase 1 (pool tuning)** has no dependencies and can ship first.
- **Phase 2 (batch primitives)** depends on nothing; can be parallel to Phase 1.
- **Phase 3A (consolidate refactor)** depends on the `planner`, `normalization`, `ingredient`, and `recipe` batch primitives from Phase 2.
- **Phase 3B (plan items refactor)** depends on the `recipe` and `planner` batch primitives.
- **Phase 3C (recipe list refactor)** depends on the `normalization` and `planner` batch primitives.
- **Phase 4 (verification)** depends on at least one of 3A/3B/3C landing.
- **Phase 5 (reflection cache)** is independent and optional.

Phases 3A/3B/3C are independent of each other and can land in any order or in parallel PRs.

---

## Reference: Slow-Log Pattern

The trace below is the canonical fingerprint of N+1 #3 (consolidate ingredients). Use it to verify the fix is complete by exercising `GET /meals/plans/{planId}/ingredients` and confirming the `canonical_ingredient_aliases` / `canonical_ingredients` per-id pairs have been replaced by a single IN query.

```
SELECT * FROM "plan_items" WHERE plan_week_id = ? ORDER BY day, position
SELECT * FROM "recipe_planner_configs" WHERE recipe_id = ? LIMIT 1     ← per item
SELECT * FROM "recipe_ingredients" WHERE recipe_id = ? ORDER BY position ← per item
SELECT * FROM "canonical_ingredient_aliases" WHERE canonical_ingredient_id = ? ← per ingredient
SELECT * FROM "canonical_ingredients" WHERE id = ? LIMIT 1             ← per ingredient
```

---

## Out of Scope

- Caching layers (Redis, in-memory). Premature; batching alone should be sufficient.
- Schema changes or new indexes. Existing indexes (`idx_recipe_ingredient_recipe`, `idx_alias_canonical_ingredient`, primary keys) are adequate once query volume drops.
- Recipe detail handler (`GET /recipes/{id}`). Already healthy at ~3 queries; will benefit indirectly from reduced DB contention.
- Refactoring `recipe.GetByID` callers other than the three identified hot paths.
- Connection-pool tuning for non-recipe services. The shared change benefits them all, but no service-specific tuning is in scope.
