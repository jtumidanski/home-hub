# Category Ingredient Count — Task Checklist

Last Updated: 2026-04-09

---

## Phase 1: Backend — recipe-service

- [ ] **1.1** Add `countByCategory` provider query to `ingredient/provider.go`
  - Query: `SELECT category_id, COUNT(*) FROM canonical_ingredients WHERE tenant_id = ? AND category_id IS NOT NULL GROUP BY category_id`
  - Returns `map[uuid.UUID]int`
  - Acceptance: compiles, returns correct counts for test data

- [ ] **1.2** Add `CountByCategory` method to `ingredient/processor.go`
  - Wraps provider with `p.db.WithContext(p.ctx)`
  - Acceptance: callable from handler with tenant ID

- [ ] **1.3** Add `RestCategorySummary` model to `ingredient/rest.go`
  - Fields: `Id`, `Name`, `SortOrder`, `IngredientCount`, `CreatedAt`, `UpdatedAt`
  - JSON tags: `name`, `sort_order`, `ingredient_count`, `created_at`, `updated_at`
  - Implements `GetName() → "categories"`, `GetID()`, `SetID()`
  - Acceptance: marshals to expected JSON:API shape

- [ ] **1.4** Add `categorySummaryHandler` and register route in `ingredient/resource.go`
  - Handler: extract tenant + access token, call `categoryclient.ListCategories`, call `proc.CountByCategory`, merge, return JSON:API
  - Route: `GET /ingredients/category-summary`
  - Error handling: 502 if category-service fails, 500 if DB query fails
  - Acceptance: returns categories with accurate `ingredient_count` values

- [ ] **1.5** Update `ingredient.InitializeRoutes` signature and `cmd/main.go`
  - Add `catClient *categoryclient.Client` parameter to `InitializeRoutes`
  - Pass `catClient` in `cmd/main.go`
  - Acceptance: compiles, `catClient` available in handler

## Phase 2: Frontend

- [ ] **2.1** Update `listCategories` in `frontend/src/services/api/ingredient.ts`
  - Change path from `"/categories"` to `"/ingredients/category-summary"`
  - Acceptance: frontend fetches from recipe-service endpoint

## Phase 3: Verification

- [ ] **3.1** `go build ./...` passes for recipe-service
- [ ] **3.2** `go test ./...` passes for recipe-service
- [ ] **3.3** Category manager shows accurate ingredient counts
- [ ] **3.4** Delete confirmation dialog shows correct count
- [ ] **3.5** Uncategorized badge on ingredients page is accurate
- [ ] **3.6** Category CRUD (create, update, delete) still works via category-service
