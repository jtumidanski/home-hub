# Category Ingredient Count — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-09

---

## 1. Overview

The ingredients page has a "Categories" management view (`category-manager.tsx`) that displays each category with a count of how many canonical ingredients are assigned to it. This count is always displayed as "0 ingredients" regardless of actual assignments.

The root cause is straightforward: the frontend reads `cat.attributes.ingredient_count` from the category-service `GET /api/v1/categories` response, but category-service's `RestModel` does not include an `ingredient_count` field — it only returns `name`, `sort_order`, `created_at`, and `updated_at`. The frontend falls back to `?? 0`, so every category shows "0 ingredients".

The challenge is architectural: categories live in `category-service`, but ingredient-to-category assignments live in `recipe-service` (`canonical_ingredients.category_id`). Category-service has no visibility into ingredient counts. The solution is to add a new endpoint in recipe-service that fetches categories (via its existing `categoryclient`) and enriches them with ingredient counts from its own database, then point the frontend's list-categories call at this new endpoint.

## 2. Goals

Primary goals:
- Display accurate ingredient counts next to each category in the category manager view.
- Display accurate uncategorized ingredient count on the ingredients page header badge.
- Ensure the delete confirmation dialog shows the correct count of ingredients that will become uncategorized.

Non-goals:
- No changes to category CRUD operations — create, update, and delete continue to go directly to category-service.
- No changes to the category-service codebase.
- No changes to how ingredients are assigned to categories.
- No new database tables or migrations.

## 3. User Stories

- As a household member managing ingredient categories, I want to see how many ingredients belong to each category so I can understand how my ingredients are organized.
- As a household member about to delete a category, I want the confirmation dialog to accurately tell me how many ingredients will become uncategorized so I can make an informed decision.

## 4. Functional Requirements

### 4.1 New recipe-service endpoint: enriched category list

Add a `GET /api/v1/ingredients/category-summary` endpoint to recipe-service that:

1. Fetches categories from category-service using the existing `categoryclient.ListCategories` (forwarding the caller's auth context).
2. Queries the local `canonical_ingredients` table for ingredient counts grouped by `category_id` for the authenticated tenant: `SELECT category_id, COUNT(*) FROM canonical_ingredients WHERE tenant_id = ? AND category_id IS NOT NULL GROUP BY category_id`.
3. Merges the counts into the category response, adding an `ingredient_count` field to each category's attributes.
4. Returns the enriched list using JSON:API format matching the existing category-service response shape, with the added `ingredient_count` field.

If the `categoryclient.ListCategories` call fails, the endpoint returns a 502 with an error message and logs the failure at error level.

### 4.2 Response shape

The response must match the existing category-service shape with one additional attribute:

```json
{
  "data": [
    {
      "type": "categories",
      "id": "<uuid>",
      "attributes": {
        "name": "Produce",
        "sort_order": 1,
        "ingredient_count": 12,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    }
  ]
}
```

Categories with no assigned ingredients have `ingredient_count: 0`.

### 4.3 Frontend: switch list-categories call

Update `ingredientService.listCategories` (in `frontend/src/services/api/ingredient.ts`) to call the new recipe-service endpoint (`/ingredients/category-summary`) instead of category-service's `/categories`.

The frontend already reads `ingredient_count` from the response — no component changes are needed.

Category CRUD mutations (`createCategory`, `updateCategory`, `deleteCategory`) continue to call category-service's `/categories` endpoint directly as they do today.

### 4.4 Nginx routing

No nginx changes required. The new endpoint lives under `/api/v1/ingredients/...` which already routes to recipe-service. Category CRUD at `/api/v1/categories/...` continues to route to category-service.

## 5. API Surface

### New endpoint

**`GET /api/v1/ingredients/category-summary`**

Returns all categories for the authenticated tenant enriched with ingredient counts.

**Parameters**: None.

**Request body**: None.

**Response model**: See section 4.2.

**Error conditions**:

| Status | Condition |
|--------|-----------|
| 502    | Failed to fetch categories from category-service |
| 500    | Failed to query ingredient counts |

### Existing endpoints

No changes to existing endpoints.

## 6. Data Model

No schema changes. No migrations.

The endpoint uses a read-only aggregate query against the existing `canonical_ingredients.category_id` column.

## 7. Service Impact

- **recipe-service** — New endpoint in the ingredient package: handler, provider query for category counts, and a REST model for the enriched category response. Uses the existing `categoryclient` dependency.
- **category-service** — No changes.
- **frontend** — One-line change in `ingredientService.listCategories` to call the new endpoint path. No component or type changes needed.

## 8. Non-Functional Requirements

- **Performance** — The endpoint issues two calls: one HTTP call to category-service and one `GROUP BY` query on an indexed column (`idx_canonical_ingredient_category`). Both are fast at household scale. No per-category queries.
- **Multi-tenancy** — The ingredient count query is scoped by `tenant_id`. Categories are fetched using the caller's auth context (forwarded via `categoryclient`). No cross-tenant data exposure.
- **Caching** — The frontend React Query hook already has a 5-minute `staleTime`. No additional caching needed.
- **Backward compatibility** — This adds a new endpoint; no existing endpoints are modified. The frontend change is purely which endpoint is called for listing categories.

## 9. Open Questions

None.

## 10. Acceptance Criteria

- [ ] `GET /api/v1/ingredients/category-summary` returns categories with accurate `ingredient_count` values matching the number of canonical ingredients assigned to each category for the tenant.
- [ ] Categories with no assigned ingredients show `ingredient_count: 0`.
- [ ] The category manager view displays correct counts (e.g., "12 ingredients") next to each category name.
- [ ] The delete confirmation dialog shows the correct count of affected ingredients.
- [ ] The "uncategorized" badge on the ingredients page header shows the correct count (derived from total minus sum of category counts).
- [ ] Category CRUD (create, update, delete) continues to work correctly via category-service.
- [ ] When category-service is unreachable, the new endpoint returns an appropriate error (not a silent failure).
- [ ] `go build ./...` and `go test ./...` pass for recipe-service.
- [ ] No changes to category-service.
