# Shopping List — Migration Plan

## Overview

This feature requires a data migration to extract ingredient categories from the `recipe` schema into a new `category` schema owned by the category-service. This must be done carefully to avoid downtime or data loss.

## Phase 1: Deploy Category Service

### Step 1: Create category-service

- New service with `category` database schema.
- Implements full CRUD for categories.
- Seeds default categories on first access per tenant (existing 11 + 4 new non-food categories).
- Deploy and verify independently.

### Step 2: Migrate existing category data

Run a one-time migration script that:

1. Reads all rows from `recipe.ingredient_categories`.
2. Inserts them into `category.categories` preserving UUIDs, tenant_id, name, sort_order, and timestamps.
3. Verifies row counts match.

This can be a standalone Go script or SQL migration. UUIDs must be preserved so that existing `canonical_ingredients.category_id` references remain valid.

### Step 3: Verify data integrity

- Confirm all categories exist in category-service with correct IDs.
- Confirm recipe-service's existing ingredient categorization still resolves correctly (category_id UUIDs match).

## Phase 2: Update Recipe Service

### Step 4: Point recipe-service at category-service

- Add HTTP client to recipe-service for calling category-service.
- Replace internal category processor/provider calls with category-service API calls.
- Remove `/ingredient-categories` route registration from recipe-service.
- Remove the `ingredient/category/` domain package.
- Drop the FK constraint from `canonical_ingredients.category_id` → `ingredient_categories.id`. The column remains as an opaque UUID reference.

### Step 5: Update frontend

- Update ingredient category API base URL from `/api/v1/ingredient-categories` to `/api/v1/categories`.
- Update JSON:API type from `"ingredient-categories"` to `"categories"`.
- Verify ingredient management page, bulk categorize, and consolidated ingredient display all work.

### Step 6: Drop old table

- Once recipe-service is confirmed working against category-service, drop the `recipe.ingredient_categories` table.
- This is the last step and should only happen after Phase 2 is verified in production.

## Phase 3: Deploy Shopping Service

### Step 7: Create shopping-service

- New service with `shopping` database schema.
- Implements list and item CRUD, archiving, check/uncheck.
- Consumes category-service for category lookups.
- Consumes recipe-service for meal plan ingredient import.
- Deploy and verify.

### Step 8: Frontend integration

- Add shopping list pages and components.
- Add navigation entry.
- Add "Add to Shopping List" action on meal plan detail page.

## Rollback Strategy

- **Phase 1 rollback:** Category-service can be removed; recipe-service still owns its categories and continues to work unchanged.
- **Phase 2 rollback:** Re-register `/ingredient-categories` routes in recipe-service, restore the domain package. Category data still exists in both places.
- **Phase 3 rollback:** Shopping-service can be removed independently; no other service depends on it.

## Nginx Routing Additions

```nginx
# Category service
location /api/v1/categories {
    proxy_pass http://category-service:8080;
}

# Shopping service
location /api/v1/shopping {
    proxy_pass http://shopping-service:8080;
}
```
