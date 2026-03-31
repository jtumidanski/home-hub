# Ingredient Category Grouping — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-31
---

## 1. Overview

The meal planner's consolidated ingredient list currently displays ingredients as a flat, unsorted list. When planning a full week of meals, this list can grow to dozens of items, making it difficult to use as a shopping reference.

This feature adds product-category grouping to ingredients. Each canonical ingredient gets a `category` attribute (e.g., Produce, Meats, Dairy, Pantry). The consolidated ingredient list — both in the frontend preview and the markdown export — renders ingredients grouped under category headers in a fixed "grocery store aisle" order. Categories are user-defined and extensible per tenant, with a set of sensible defaults seeded on creation.

Ingredients without a category (including all unresolved ingredients) appear in an explicit "Uncategorized" group to encourage users to categorize their ingredient registry over time. A bulk-edit UI allows efficient category assignment across many ingredients at once.

## 2. Goals

Primary goals:
- Group consolidated ingredients by product category in both frontend and markdown export
- Allow tenants to define and manage their own set of ingredient categories
- Provide a bulk-edit UI for efficiently assigning categories to canonical ingredients
- Display categories in a fixed grocery-aisle sort order
- Make uncategorized ingredients visually explicit to encourage categorization

Non-goals:
- Auto-classification of ingredients via heuristics or ML
- Per-recipe or per-plan category overrides (category lives on canonical ingredient only)
- Filtering or hiding specific categories from the ingredient list
- Custom sort order for categories (hardcoded aisle order for now)

## 3. User Stories

- As a meal planner, I want the consolidated ingredient list grouped by product category so I can shop aisle-by-aisle without scanning the whole list
- As a meal planner, I want the markdown export grouped by category so my printed/shared shopping list is organized
- As a household manager, I want to create custom categories that match how my grocery store is organized
- As a household manager, I want to bulk-assign categories to ingredients so I can quickly categorize my existing ingredient registry
- As a household manager, I want uncategorized ingredients clearly separated so I know which ingredients still need attention

## 4. Functional Requirements

### 4.1 Ingredient Categories (CRUD)

- Tenants can create, rename, and delete ingredient categories
- Each category has a `name` (unique per tenant) and a `sort_order` (integer, determines display position)
- Default categories are auto-seeded when the `GET /ingredient-categories` endpoint returns empty for a tenant: Produce, Meats & Seafood, Dairy & Eggs, Bakery & Bread, Pantry & Dry Goods, Frozen, Beverages, Snacks & Sweets, Condiments & Sauces, Spices & Seasonings, Other
- Default sort order follows grocery-aisle convention (Produce first, Other last)
- Deleting a category nullifies the `category_id` on any ingredients referencing it (they become Uncategorized)
- Categories are scoped to tenant (not household)

### 4.2 Category Assignment on Canonical Ingredients

- The canonical ingredient model gains an optional `category_id` foreign key
- Category can be set via the existing `POST /ingredients` create endpoint (optional `categoryId` field)
- Category can be set/updated via the existing `PATCH /ingredients/{id}` endpoint (add `categoryId` to the update payload)
- The `GET /ingredients` list response includes the category name for display
- The ingredient detail page shows the current category and allows changing it

### 4.3 Bulk Category Assignment

- A dedicated bulk-edit UI allows selecting multiple ingredients and assigning them to a category in one action
- The bulk-edit UI supports filtering by: uncategorized only, specific category, search by name
- A new API endpoint accepts a list of ingredient IDs and a target category ID, applying the update atomically
- The UI shows a count of uncategorized ingredients as a prompt to encourage categorization

### 4.4 Consolidated Ingredient Grouping

- The `ConsolidateIngredients` pipeline includes category information in its output
- The `GET /meals/plans/{planId}/ingredients` response includes `category_name` and `category_sort_order` on each ingredient
- Unresolved ingredients and ingredients with no category get `category_name: null` and sort last
- The frontend ingredient preview renders ingredients under category section headers
- Empty categories (no ingredients in the current plan) are not shown
- Within each category group, ingredients are sorted alphabetically by display name

### 4.5 Markdown Export Grouping

- The markdown export renders category headers (e.g., `## Produce`) with ingredients listed beneath
- Categories appear in sort-order sequence
- Uncategorized ingredients appear under an `## Uncategorized` header at the end
- Empty categories are omitted from the export

## 5. API Surface

### New Endpoints

#### `GET /api/v1/ingredient-categories`

List all categories for the tenant.

Response (JSON:API):
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "ingredient-categories",
      "attributes": {
        "name": "Produce",
        "sort_order": 1,
        "ingredient_count": 24,
        "created_at": "2026-03-31T00:00:00Z",
        "updated_at": "2026-03-31T00:00:00Z"
      }
    }
  ]
}
```

#### `POST /api/v1/ingredient-categories`

Create a new category. Assigns the next sort_order value automatically.

Request:
```json
{
  "data": {
    "type": "ingredient-categories",
    "attributes": {
      "name": "Produce"
    }
  }
}
```

#### `PATCH /api/v1/ingredient-categories/{id}`

Update category name or sort_order.

Request:
```json
{
  "data": {
    "type": "ingredient-categories",
    "attributes": {
      "name": "Fresh Produce",
      "sort_order": 1
    }
  }
}
```

#### `DELETE /api/v1/ingredient-categories/{id}`

Delete a category. Nullifies `category_id` on all ingredients referencing it.

#### `POST /api/v1/ingredients/bulk-categorize`

Assign a category to multiple ingredients at once.

Request:
```json
{
  "data": {
    "type": "ingredient-bulk-categorize",
    "attributes": {
      "ingredient_ids": ["uuid1", "uuid2", "uuid3"],
      "category_id": "uuid"
    }
  }
}
```

Response: 204 No Content on success.

### Modified Endpoints

#### `PATCH /api/v1/ingredients/{id}`

Add optional `category_id` field to the update payload. Pass `null` to clear the category.

#### `GET /api/v1/ingredients`

Response now includes `category_id` and `category_name` in attributes.

#### `GET /api/v1/meals/plans/{planId}/ingredients`

Response now includes `category_name` (string or null) and `category_sort_order` (integer or null) in each ingredient's attributes.

## 6. Data Model

### New Table: `ingredient_categories`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, FK to tenants, indexed |
| name | VARCHAR(100) | NOT NULL |
| sort_order | INT | NOT NULL, DEFAULT 0 |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |
| **unique** | | (tenant_id, name) |

### Modified Table: `canonical_ingredients`

| Column | Type | Constraints |
|--------|------|-------------|
| category_id | UUID | NULLABLE, FK to ingredient_categories, SET NULL on delete |

### Migration Notes

- Add `ingredient_categories` table
- Add `category_id` column to `canonical_ingredients` (nullable, no backfill needed)
- Add foreign key with ON DELETE SET NULL
- Add index on `canonical_ingredients.category_id` for join performance

## 7. Service Impact

### recipe-service

- **New domain:** `ingredient/category` — model, processor, repository, REST resource for ingredient categories
- **Modified domain:** `ingredient` — add `category_id` field to model, update processor and repository for category-aware queries, add bulk-categorize endpoint
- **Modified domain:** `export` — include category info in `ConsolidatedIngredient`, group markdown output by category
- **Migration:** new table + column addition

### frontend

- **New components:** Category management page/section, bulk-edit UI
- **Modified components:** `ingredient-preview.tsx` — render grouped sections with category headers
- **Modified pages:** `IngredientsPage.tsx` — show category column, link to category management; `IngredientDetailPage.tsx` — category selector field
- **New types:** `IngredientCategory`, updated `PlanIngredientAttributes` with category fields
- **New API service methods:** category CRUD, bulk-categorize

## 8. Non-Functional Requirements

- **Performance:** Bulk-categorize must handle up to 200 ingredients in a single request without timeout
- **Multi-tenancy:** Categories are tenant-scoped; all queries filter by tenant_id
- **Data integrity:** Deleting a category must not orphan ingredients — SET NULL on delete ensures they revert to Uncategorized
- **Backward compatibility:** Existing plans and exports continue to work; ingredients without categories simply appear in the Uncategorized group

## 9. Open Questions

None — all questions resolved.

## 10. Acceptance Criteria

- [ ] Tenant can create, rename, and delete ingredient categories
- [ ] Default categories are auto-seeded on first access and are in grocery-aisle order
- [ ] Canonical ingredients can be assigned a category at creation time (POST) and via update (PATCH)
- [ ] Bulk-categorize endpoint updates multiple ingredients in one request
- [ ] Bulk-edit UI allows selecting ingredients and assigning categories with filtering
- [ ] `GET /ingredients` response includes category name
- [ ] Consolidated ingredient list response includes category name and sort order
- [ ] Frontend ingredient preview groups ingredients by category with section headers
- [ ] Uncategorized/unresolved ingredients appear in an explicit "Uncategorized" group at the end
- [ ] Markdown export groups ingredients by category with section headers
- [ ] Deleting a category sets affected ingredients to uncategorized
- [ ] Empty categories are not shown in the plan ingredient view or markdown export
- [ ] All data is tenant-scoped
