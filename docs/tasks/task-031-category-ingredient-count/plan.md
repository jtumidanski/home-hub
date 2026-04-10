# Category Ingredient Count — Implementation Plan

Last Updated: 2026-04-09

---

## Executive Summary

The category manager UI displays "0 ingredients" for every category because the `category-service` REST model has no `ingredient_count` field. The frontend already reads this field and falls back to `0`. The fix adds a new endpoint in `recipe-service` that proxies the category list from `category-service` and enriches it with ingredient counts from the local `canonical_ingredients` table, then updates the frontend to call this endpoint instead.

Total estimated effort: **Small** (~2-3 hours). Three files to create, two files to modify, no migrations.

---

## Current State Analysis

### What exists today

1. **category-service** `GET /api/v1/categories` returns `name`, `sort_order`, `created_at`, `updated_at` — no ingredient count.
2. **recipe-service** owns `canonical_ingredients` table with a `category_id` column (indexed at `idx_canonical_ingredient_category`) referencing category-service UUIDs.
3. **recipe-service** already has a `categoryclient` package that calls category-service, forwarding auth via `access_token` cookie and `X-Tenant-ID`/`X-Household-ID` headers.
4. **Frontend** `category-manager.tsx:113` renders `cat.attributes.ingredient_count ?? 0` — always 0 since the backend never provides it.
5. **Frontend** `ingredientService.listCategories()` calls `/categories` which nginx routes to category-service.

### Why it fails

Category-service has no access to `canonical_ingredients` (that's recipe-service's database). The frontend was built expecting an `ingredient_count` that was never implemented on the backend.

---

## Proposed Future State

A new `GET /api/v1/ingredients/category-summary` endpoint in recipe-service:
1. Calls `categoryclient.ListCategories` to get the full category list.
2. Runs `SELECT category_id, COUNT(*) FROM canonical_ingredients WHERE tenant_id = ? AND category_id IS NOT NULL GROUP BY category_id` against its own DB.
3. Merges counts into the category response, returning JSON:API with `ingredient_count` added.

Frontend switches its `listCategories` call from `/categories` to `/ingredients/category-summary`. Category CRUD stays on `/categories` (category-service).

---

## Implementation Phases

### Phase 1: Backend — recipe-service endpoint

#### 1.1 Add provider query for ingredient counts by category

**File**: `services/recipe-service/internal/ingredient/provider.go`

Add a new function:
```go
func countByCategory(tenantID uuid.UUID) func(db *gorm.DB) (map[uuid.UUID]int, error)
```

This runs the `GROUP BY category_id` query and returns a `map[category_id]count`.

**Effort**: S

#### 1.2 Add processor method

**File**: `services/recipe-service/internal/ingredient/processor.go`

Add:
```go
func (p *Processor) CountByCategory(tenantID uuid.UUID) (map[uuid.UUID]int, error)
```

Wraps the provider call with context.

**Effort**: S

#### 1.3 Add REST model for enriched category

**File**: `services/recipe-service/internal/ingredient/rest.go`

Add a `RestCategorySummary` struct:
```go
type RestCategorySummary struct {
    Id              uuid.UUID `json:"-"`
    Name            string    `json:"name"`
    SortOrder       int       `json:"sort_order"`
    IngredientCount int       `json:"ingredient_count"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

With `GetName() → "categories"` and `GetID()` methods for JSON:API marshaling.

**Effort**: S

#### 1.4 Add handler and register route

**File**: `services/recipe-service/internal/ingredient/resource.go`

Add `categorySummaryHandler` that:
1. Extracts tenant context and access token cookie.
2. Calls `categoryclient.ListCategories`.
3. Calls `proc.CountByCategory`.
4. Merges into `[]RestCategorySummary`.
5. Returns JSON:API response.

Register route: `api.HandleFunc("/ingredients/category-summary", ...)`.

**Effort**: S

#### 1.5 Pass `categoryclient` to ingredient routes

**File**: `services/recipe-service/cmd/main.go`

Change `ingredient.InitializeRoutes(db)` to `ingredient.InitializeRoutes(db, catClient)`.

Update `ingredient.InitializeRoutes` signature to accept `*categoryclient.Client`.

**Effort**: S

### Phase 2: Frontend — switch endpoint

#### 2.1 Update `listCategories` path

**File**: `frontend/src/services/api/ingredient.ts`

Change:
```ts
listCategories(tenant: Tenant) {
    return this.getList<IngredientCategory>(tenant, "/categories");
}
```
To:
```ts
listCategories(tenant: Tenant) {
    return this.getList<IngredientCategory>(tenant, "/ingredients/category-summary");
}
```

**Effort**: S

### Phase 3: Verification

#### 3.1 Build verification

Run `go build ./...` for recipe-service. Verify no compilation errors.

#### 3.2 Test verification

Run `go test ./...` for recipe-service. Verify no regressions.

#### 3.3 Manual verification

- Category manager shows accurate counts.
- Delete dialog shows accurate count warning.
- Uncategorized badge on ingredients page shows correct count.
- Category CRUD still works (routed to category-service).

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Auth forwarding fails for new endpoint | Low | Medium | Reuse exact same `accessTokenCookie` + tenant context pattern from `plan/resource.go` |
| Category CRUD breaks when switching frontend endpoint | Low | High | CRUD mutations stay on `/categories`; only the list call moves |
| `categoryclient.ListCategories` fails silently | Low | Medium | Return 502 with logged error, not a silent empty list |

---

## Success Metrics

- Category manager shows non-zero counts for categories with assigned ingredients.
- Uncategorized count badge is accurate.
- Delete confirmation shows correct ingredient count.
- No increase in error rates for category operations.

---

## Required Resources and Dependencies

- **Existing**: `categoryclient` package, `canonical_ingredients.category_id` index, tenant context middleware.
- **No new dependencies**: No new packages, services, or infrastructure.

---

## Timeline Estimate

| Phase | Effort |
|-------|--------|
| Phase 1: Backend (5 sub-tasks) | ~1.5 hours |
| Phase 2: Frontend (1 sub-task) | ~10 minutes |
| Phase 3: Verification | ~30 minutes |
| **Total** | **~2-3 hours** |
