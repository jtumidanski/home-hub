# Category Ingredient Count — Context

Last Updated: 2026-04-09

---

## Key Files

### Backend — recipe-service (changes needed)

| File | Purpose |
|------|---------|
| `services/recipe-service/internal/ingredient/provider.go` | Add `countByCategory` query |
| `services/recipe-service/internal/ingredient/processor.go` | Add `CountByCategory` method |
| `services/recipe-service/internal/ingredient/rest.go` | Add `RestCategorySummary` model |
| `services/recipe-service/internal/ingredient/resource.go` | Add handler + route registration |
| `services/recipe-service/cmd/main.go` | Pass `catClient` to ingredient routes |

### Backend — recipe-service (reference, no changes)

| File | Purpose |
|------|---------|
| `services/recipe-service/internal/categoryclient/client.go` | Existing category-service HTTP client — reuse `ListCategories` |
| `services/recipe-service/internal/ingredient/entity.go` | Entity with `CategoryId *uuid.UUID` field |
| `services/recipe-service/internal/ingredient/model.go` | Domain model |
| `services/recipe-service/internal/plan/resource.go:355` | `accessTokenCookie` helper — pattern to reuse |
| `services/recipe-service/internal/config/config.go` | `CategoryServiceURL` config already exists |

### Backend — category-service (no changes)

| File | Purpose |
|------|---------|
| `services/category-service/internal/category/rest.go` | Current `RestModel` — no `ingredient_count` field |
| `services/category-service/internal/category/resource.go` | Current list handler |

### Frontend (one change)

| File | Purpose |
|------|---------|
| `frontend/src/services/api/ingredient.ts:94` | `listCategories` — change path from `/categories` to `/ingredients/category-summary` |

### Frontend (reference, no changes)

| File | Purpose |
|------|---------|
| `frontend/src/types/models/ingredient.ts:59-65` | `IngredientCategoryAttributes` already declares `ingredient_count?: number` |
| `frontend/src/components/features/ingredients/category-manager.tsx:113` | Reads `cat.attributes.ingredient_count ?? 0` |
| `frontend/src/pages/IngredientsPage.tsx:66` | Derives uncategorized count from category counts |
| `frontend/src/lib/hooks/api/use-ingredient-categories.ts` | React Query hook — calls `ingredientService.listCategories` |

### Infrastructure (no changes)

| File | Purpose |
|------|---------|
| `deploy/compose/nginx.conf:161` | `/api/v1/categories` routes to category-service — stays as-is |
| `deploy/compose/docker-compose.yml` | `CATEGORY_SERVICE_URL` already configured for recipe-service |

---

## Key Decisions

1. **recipe-service enriches, category-service stays untouched** — Categories are category-service's domain; ingredient counts are recipe-service's domain. recipe-service already has a `categoryclient` to fetch categories. Adding a count query to its own DB and merging the two is the cleanest approach.

2. **New endpoint, not modifying existing** — Adding `/ingredients/category-summary` rather than modifying `/categories` avoids touching category-service and nginx routing.

3. **Category CRUD stays on category-service** — Only the list call (which needs enrichment) moves to recipe-service. Create, update, delete continue to call `/categories` (category-service) directly.

4. **Auth forwarding via cookie + headers** — Same pattern used by `categoryclient.ListCategories` in the plan/export flow: forward `access_token` cookie and `X-Tenant-ID`/`X-Household-ID` headers.

---

## Dependencies

- `categoryclient.Client` — Already exists, already tested in the plan/export code path.
- `canonical_ingredients.category_id` index (`idx_canonical_ingredient_category`) — Already exists, makes `GROUP BY category_id` fast.
- Tenant context middleware — Already applied to all `/api/v1/*` routes.

---

## Query Design

The ingredient count query:

```sql
SELECT category_id, COUNT(*) as count
FROM canonical_ingredients
WHERE tenant_id = $1
  AND category_id IS NOT NULL
GROUP BY category_id
```

Returns a map of `category_id → count`. Categories not in the result have 0 ingredients. Uses the existing `idx_canonical_ingredient_category` index.
