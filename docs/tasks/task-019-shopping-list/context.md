# Task 019: Shopping List — Context

Last Updated: 2026-03-31

---

## Key Files

### Category Domain (to extract from recipe-service)
- `services/recipe-service/internal/ingredient/category/entity.go` — Current GORM entity
- `services/recipe-service/internal/ingredient/category/model.go` — Current domain model
- `services/recipe-service/internal/ingredient/category/builder.go` — Current builder
- `services/recipe-service/internal/ingredient/category/processor.go` — Current business logic + 11 default seeds
- `services/recipe-service/internal/ingredient/category/provider.go` — Current data access
- `services/recipe-service/internal/ingredient/category/resource.go` — Current HTTP handlers
- `services/recipe-service/internal/ingredient/category/rest.go` — Current JSON:API model
- `services/recipe-service/internal/ingredient/category/administrator.go` — Admin functions

### Meal Plan / Export (consumed by shopping import)
- `services/recipe-service/internal/plan/resource.go` — `getIngredientsHandler()` (~line 381)
- `services/recipe-service/internal/export/processor.go` — `ConsolidateIngredients()` logic

### Infrastructure
- `deploy/compose/docker-compose.yml` — Service definitions
- `deploy/compose/nginx.conf` — Route configuration
- `go.work` — Go workspace modules

### Frontend — Category (to migrate)
- `frontend/src/services/api/ingredient.ts` — Category API calls
- `frontend/src/lib/hooks/api/use-ingredient-categories.ts` — React Query hooks
- `frontend/src/lib/hooks/api/query-keys.ts` — Query key definitions
- `frontend/src/components/features/ingredients/category-manager.tsx` — Category management UI
- `frontend/src/types/models/ingredient.ts` — Category type definitions

### Frontend — Patterns to Follow
- `frontend/src/pages/MealsPage.tsx` — Example page with list + detail pattern
- `frontend/src/services/api/meals.ts` — Example service layer
- `frontend/src/lib/hooks/api/use-meals.ts` — Example hooks
- `frontend/src/components/features/meals/` — Example feature components

### Backend — Patterns to Follow
- `services/recipe-service/cmd/main.go` — Service entry point pattern
- `services/recipe-service/Dockerfile` — Dockerfile pattern
- `services/recipe-service/internal/config/` — Config pattern
- `services/recipe-service/internal/recipe/` — Complete domain pattern

### Shared Modules
- `shared/go/auth/` — JWT middleware
- `shared/go/database/` — GORM setup, migrations, tenant callbacks
- `shared/go/server/` — HTTP server, JSON:API helpers
- `shared/go/tenant/` — Tenant context

### CI/CD
- `.github/workflows/` — GitHub Actions workflows

---

## Key Decisions

1. **Category data denormalized on shopping items**: `category_name` and `category_sort_order` stored as snapshots on `shopping_items` to avoid cross-service calls on every list render. Tradeoff: stale data if categories are renamed.

2. **No FK constraints across services**: `category_id` on both `canonical_ingredients` and `shopping_items` is an opaque UUID — no database-level FK to the category-service's table.

3. **JWT forwarding for service-to-service calls**: Shopping-service forwards the user's JWT when calling recipe-service and category-service. No separate service account tokens.

4. **Import is append-only**: Meal plan import appends items to a list without deduplication or modification of existing items.

5. **Auto-seed on first access**: Category-service seeds 15 default categories (11 existing food + 4 new non-food) per tenant on first access, matching recipe-service's current behavior.

6. **Snapshot approach for categories**: Existing shopping items retain their category snapshot even if the category is renamed or deleted. Only new items reflect current state.

7. **No separate migration SQL**: Categories are seeded on demand by the category-service. The recipe-service category table can be dropped after the migration is verified.

---

## Dependencies Between Phases

```
Phase 1 (Category Service)
  |
  +--> Phase 2 (Recipe Migration) ------+
  |                                      |
  +--> Phase 3 (Shopping Service) -------+--> Phase 5 (Frontend Shopping)
  |                                      |
  +--> Phase 4 (Frontend Category) -----+
  |
  +--> Phase 6 (CI/CD & Docs) — can start after Phase 1, finalized after Phase 3
```

---

## API Contract Summary

### Category Service: `/api/v1/categories`
- `GET /categories` — List all (tenant-scoped)
- `POST /categories` — Create
- `PATCH /categories/{id}` — Update name/sort_order
- `DELETE /categories/{id}` — Delete

### Shopping Service: `/api/v1/shopping`
- `GET /lists` — List (query: `status=active|archived`)
- `POST /lists` — Create
- `GET /lists/{id}` — Detail with items
- `PATCH /lists/{id}` — Rename
- `DELETE /lists/{id}` — Delete
- `POST /lists/{id}/archive` — Archive
- `POST /lists/{id}/unarchive` — Unarchive
- `POST /lists/{id}/items` — Add item
- `PATCH /lists/{id}/items/{itemId}` — Update item
- `DELETE /lists/{id}/items/{itemId}` — Remove item
- `PATCH /lists/{id}/items/{itemId}/check` — Check/uncheck
- `POST /lists/{id}/items/uncheck-all` — Uncheck all
- `POST /lists/{id}/import/meal-plan` — Import from meal plan

---

## Default Categories (15 total)

### Existing Food (11)
Produce, Meat & Seafood, Dairy & Eggs, Bakery, Frozen, Canned & Jarred, Dry Goods & Pasta, Condiments & Sauces, Spices & Seasonings, Beverages, Snacks

### New Non-Food (4)
Household, Personal Care, Baby & Kids, Pet Supplies
