# Recipe Service Latency Fix — Task Checklist

Last Updated: 2026-04-06

---

## Phase 1: Connection Pool Tuning [S]

- [ ] **1.1** In `shared/go/database/database.go`, after `gorm.Open` succeeds, obtain `sqlDB, err := db.DB()` and configure `sqlDB.SetMaxOpenConns(25)`, `sqlDB.SetMaxIdleConns(25)`, `sqlDB.SetConnMaxLifetime(30 * time.Minute)`, `sqlDB.SetConnMaxIdleTime(5 * time.Minute)`. Log a warning (don't fatal) if `db.DB()` errors.
- [ ] **1.2** (Optional) Extend `database.Config` with `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `ConnMaxIdleTime` fields. Apply config values when non-zero, fall back to the defaults from 1.1 otherwise. Document the math (services × MaxOpenConns ≤ Postgres `max_connections`) in a comment above the defaults.
- [ ] **1.3** Run `go build ./...` from the repo root and any service-level Docker builds for services that consume `shared/go/database` to confirm the shared change compiles everywhere.

**Acceptance**: All services using `shared/go/database` build cleanly. No behavior change on idle. Slow-log noise on the recipe-service drops at least somewhat once the change is deployed (baseline measurement).

---

## Phase 2: Batch Fetch Primitives [M]

### 2A: Planner

- [ ] **2A.1** In `services/recipe-service/internal/planner/provider.go`, add `getByRecipeIDs(recipeIDs []uuid.UUID) func(db *gorm.DB) ([]Entity, error)` that emits `WHERE recipe_id IN (?)`. Short-circuit to empty slice + nil error when input is empty.
- [ ] **2A.2** In `services/recipe-service/internal/planner/processor.go`, add `GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error)` that calls the provider, builds models via the existing `Make` function, and keys the map by `recipe_id`. Empty input returns an empty map without hitting the DB.

### 2B: Normalization

- [ ] **2B.1** In `services/recipe-service/internal/normalization/provider.go`, add `getByRecipeIDs(recipeIDs []uuid.UUID)` that emits `WHERE recipe_id IN (?) ORDER BY position ASC`. Short-circuit on empty input.
- [ ] **2B.2** In `services/recipe-service/internal/normalization/processor.go`, add `GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID][]Model, error)` returning a map keyed by recipe ID, preserving the per-recipe position ordering.

### 2C: Ingredient

- [ ] **2C.1** In `services/recipe-service/internal/ingredient/provider.go`, add `GetByIDs(ids []uuid.UUID) func(db *gorm.DB) ([]Entity, error)` that uses `Preload("Aliases").Where("id IN (?)", ids)`. Short-circuit on empty input.
- [ ] **2C.2** In `services/recipe-service/internal/ingredient/processor.go`, add `GetByIDs(ids []uuid.UUID) (map[uuid.UUID]Model, error)` returning a map keyed by ingredient ID with aliases populated.
- [ ] **2C.3** Verify with logging during a test run that `Preload("Aliases")` with an `IN` parent query produces exactly **one** child query (`canonical_ingredient_aliases WHERE canonical_ingredient_id IN (?)`), not N. If GORM emits N child queries, fall back to a manual second query that fetches aliases by IN and zips them onto the entities in Go.

### 2D: Recipe

- [ ] **2D.1** In `services/recipe-service/internal/recipe/provider.go`, add `getByIDs(ids []uuid.UUID)` that emits `WHERE id IN (?) AND deleted_at IS NULL` with `Preload("Tags")`. Short-circuit on empty input.
- [ ] **2D.2** In `services/recipe-service/internal/recipe/processor.go`, add `GetByIDs(ids []uuid.UUID) (map[uuid.UUID]Model, error)` returning a map keyed by recipe ID.

### 2E: Tests

- [ ] **2E.1** Add unit or integration tests covering each new `…ByIDs` / `…ByRecipeIDs` helper: empty input, single ID, multiple IDs (verify map keys, ordering for slice maps, and correct preload population for the ingredient case).

**Acceptance**: All new helpers compile, tests pass, and no existing test regresses. Manual test: hit a real DB and confirm exactly one SQL query per helper invocation (two for the ingredient helper because of the aliases preload).

---

## Phase 3: Caller Refactors [M]

### 3A: ConsolidateIngredients (highest impact)

- [ ] **3A.1** In `services/recipe-service/internal/export/processor.go`, refactor `ConsolidateIngredients` to first collect all `recipeIDs` from `items`.
- [ ] **3A.2** Replace per-item `normProc.GetByRecipeID` with a single `normalization.GetByRecipeIDs` call into a `map[uuid.UUID][]normalization.Model`.
- [ ] **3A.3** Replace per-item `plannerProc.GetByRecipeID` (inside `effectiveMultiplier` / `getServingsYield`) with a single `planner.GetByRecipeIDs` call into a `map[uuid.UUID]planner.Model`. Pass the map (or a closure capturing it) into the helpers, or inline the multiplier logic to avoid the per-item lookup.
- [ ] **3A.4** Replace per-item `recipeProc.Get` (used both in `effectiveMultiplier` fallback and in the unresolved cooklang fallback path) with a single `recipe.GetByIDs` call into a `map[uuid.UUID]recipe.Model`.
- [ ] **3A.5** First pass: walk items in memory using the maps to compute multipliers and accumulate quantities. Track distinct `canonID`s seen across the whole plan in a `map[uuid.UUID]struct{}`.
- [ ] **3A.6** After the first pass, call `ingredient.GetByIDs(canonIDsSlice)` exactly once to fetch all canonicals + aliases.
- [ ] **3A.7** Second pass: walk the `resolved` accumulator and fill in `displayName`, `name`, `unitFamily`, and category info from the canonical map. Preserve the existing sort and accumulator semantics byte-for-byte.
- [ ] **3A.8** Run the existing `internal/export` tests; verify they still pass. Manually exercise `GET /meals/plans/{id}/ingredients` against a populated plan and confirm output is identical to before.

### 3B: BuildItemsResponse

- [ ] **3B.1** In `services/recipe-service/internal/plan/processor.go`, refactor `BuildItemsResponse` to collect distinct `recipeIDs` from `items` first.
- [ ] **3B.2** Call `recipe.GetByIDs` and `planner.GetByRecipeIDs` once each into local maps.
- [ ] **3B.3** Walk items in memory using the maps to populate `RestItemModel`s. Preserve the existing "(deleted recipe)" fallback when an ID is missing from the recipe map.
- [ ] **3B.4** Run plan integration tests; verify they still pass. Manually exercise `GET /meals/plans/{id}` and confirm response matches pre-refactor output.

### 3C: Recipe List Enrichment

- [ ] **3C.1** In `services/recipe-service/internal/recipe/processor.go`, add `BuildListEnrichments(models []Model) []ListEnrichment` that:
  - Collects all recipe IDs from `models`.
  - Calls `normalization.GetByRecipeIDs` and `planner.GetByRecipeIDs` once.
  - Walks `models` in order and constructs each `ListEnrichment` from the maps.
  - Computes `TotalIngredients`, `ResolvedIngredients`, `PlannerReady`, `Classification` from in-memory data only.
- [ ] **3C.2** In `services/recipe-service/internal/recipe/resource.go`, replace the per-recipe `BuildListEnrichment` loop in `listHandler` (around lines 113-124) with a single call to `BuildListEnrichments`. Keep the existing `include_usage` branch (already batched via `GetRecipeUsage`).
- [ ] **3C.3** Decide whether `BuildListEnrichment` (singular) still has callers. If not, delete it. If yes, leave it unchanged.
- [ ] **3C.4** Run recipe integration tests; verify they still pass. Manually exercise `GET /recipes?page[size]=20` and confirm the response matches pre-refactor output (especially `plannerReady`, `classification`, `totalIngredients`, `resolvedIngredients`).

---

## Phase 4: Verification & Monitoring [S]

- [ ] **4.1** Add a temporary GORM callback (or context-bound counter) that increments a per-request query count. Log the count from each refactored handler at INFO level along with the request path. Plan to remove after rollout.
- [ ] **4.2** Manually exercise the three pages and assert query counts:
  - `GET /recipes?page[size]=20` → ≤ 6 queries
  - `GET /meals/plans/{id}` → ≤ 5 queries
  - `GET /meals/plans/{id}/ingredients` → ≤ 6 queries (independent of plan size)
- [ ] **4.3** Run all existing recipe-service tests (`go test ./services/recipe-service/...`) and confirm no regressions.
- [ ] **4.4** Deploy to the dev environment and tail logs for `SLOW SQL >= 200ms` lines from recipe-service for at least 10 minutes of normal usage. Confirm they drop to ≤ 1/min in steady state.
- [ ] **4.5** Verify p95 wall-clock latency on the three endpoints has dropped by ≥ 80% (eyeball from logs or a quick manual timing if no metrics dashboard exists).
- [ ] **4.6** Remove the temporary query-count callback and logs from 4.1 once the rollout is verified.

---

## Phase 5 (Optional): Tenant Callback Reflection Cache [S]

- [ ] **5.1** In `shared/go/database/tenant_callbacks.go`, introduce a package-level `var hasTenantIDCache sync.Map` keyed by `reflect.Type`.
- [ ] **5.2** In `hasTenantIDField`, look up `t` in the cache before walking fields. On miss, walk the fields, store the boolean result, and return it.
- [ ] **5.3** Run all services' build and tests to confirm nothing relied on the per-call reflection.
- [ ] **5.4** Confirm in dev that no tests now fail due to stale cache state across test runs (the cache is process-scoped, so this should be fine).

**Acceptance**: All builds and tests pass; no observable behavior change other than reduced per-query overhead.

---

## Done Criteria

- [ ] All Phase 1 / 2 / 3 / 4 checkboxes complete.
- [ ] No `SLOW SQL` log entries from recipe-service in steady-state dev usage of the recipes list, recipe detail, or meal-planning pages.
- [ ] No regressions in existing recipe-service integration tests.
- [ ] No API contract changes; frontend continues to work without any code changes.
