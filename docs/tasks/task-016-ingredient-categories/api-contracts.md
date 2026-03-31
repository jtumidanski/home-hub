# Ingredient Categories — API Contracts

## New Endpoints

### List Categories

```
GET /api/v1/ingredient-categories
Authorization: Bearer <jwt>
```

Query parameters:
- None (categories are few enough to return all at once)

Auto-seed behavior: If the tenant has no categories, this endpoint automatically seeds the default category set (Produce, Meats & Seafood, Dairy & Eggs, Bakery & Bread, Pantry & Dry Goods, Frozen, Beverages, Snacks & Sweets, Condiments & Sauces, Spices & Seasonings, Other) before returning. This is idempotent — subsequent calls return the existing categories.

Response `200 OK`:
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "type": "ingredient-categories",
      "attributes": {
        "name": "Produce",
        "sort_order": 1,
        "ingredient_count": 24,
        "created_at": "2026-03-31T12:00:00Z",
        "updated_at": "2026-03-31T12:00:00Z"
      }
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "type": "ingredient-categories",
      "attributes": {
        "name": "Meats & Seafood",
        "sort_order": 2,
        "ingredient_count": 12,
        "created_at": "2026-03-31T12:00:00Z",
        "updated_at": "2026-03-31T12:00:00Z"
      }
    }
  ]
}
```

### Create Category

```
POST /api/v1/ingredient-categories
Authorization: Bearer <jwt>
Content-Type: application/vnd.api+json
```

Request:
```json
{
  "data": {
    "type": "ingredient-categories",
    "attributes": {
      "name": "Deli & Prepared"
    }
  }
}
```

Response `201 Created`:
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440010",
    "type": "ingredient-categories",
    "attributes": {
      "name": "Deli & Prepared",
      "sort_order": 12,
      "ingredient_count": 0,
      "created_at": "2026-03-31T12:00:00Z",
      "updated_at": "2026-03-31T12:00:00Z"
    }
  }
}
```

Error cases:
- `409 Conflict` — category name already exists for this tenant
- `422 Unprocessable Entity` — name is blank or exceeds 100 characters

### Update Category

```
PATCH /api/v1/ingredient-categories/{id}
Authorization: Bearer <jwt>
Content-Type: application/vnd.api+json
```

Request (all fields optional):
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

Response `200 OK`: Returns the updated category resource.

Error cases:
- `404 Not Found` — category does not exist or belongs to another tenant
- `409 Conflict` — updated name conflicts with existing category
- `422 Unprocessable Entity` — invalid input

### Delete Category

```
DELETE /api/v1/ingredient-categories/{id}
Authorization: Bearer <jwt>
```

Response `204 No Content`

Side effects:
- All canonical ingredients referencing this category have their `category_id` set to NULL (handled by DB ON DELETE SET NULL)

Error cases:
- `404 Not Found` — category does not exist or belongs to another tenant

### Bulk Categorize Ingredients

```
POST /api/v1/ingredients/bulk-categorize
Authorization: Bearer <jwt>
Content-Type: application/vnd.api+json
```

Request:
```json
{
  "data": {
    "type": "ingredient-bulk-categorize",
    "attributes": {
      "ingredient_ids": [
        "550e8400-e29b-41d4-a716-446655440020",
        "550e8400-e29b-41d4-a716-446655440021",
        "550e8400-e29b-41d4-a716-446655440022"
      ],
      "category_id": "550e8400-e29b-41d4-a716-446655440001"
    }
  }
}
```

Response `204 No Content`

Behavior:
- Updates `category_id` on all specified ingredients in a single transaction
- Silently skips ingredient IDs that don't exist or belong to a different tenant
- Maximum 200 ingredient IDs per request

Error cases:
- `404 Not Found` — category_id does not exist or belongs to another tenant
- `422 Unprocessable Entity` — ingredient_ids is empty or exceeds 200 items

---

## Modified Endpoints

### Create Ingredient (existing)

```
POST /api/v1/ingredients
```

Added optional field in attributes:
```json
{
  "data": {
    "type": "ingredients",
    "attributes": {
      "name": "chicken breast",
      "category_id": "550e8400-e29b-41d4-a716-446655440002"
    }
  }
}
```

`category_id` is optional. If omitted, the ingredient is created as uncategorized.

### Update Ingredient (existing)

```
PATCH /api/v1/ingredients/{id}
```

Added field in attributes:
```json
{
  "data": {
    "type": "ingredients",
    "attributes": {
      "category_id": "550e8400-e29b-41d4-a716-446655440001"
    }
  }
}
```

Pass `"category_id": null` to clear the category assignment.

### List Ingredients (existing)

```
GET /api/v1/ingredients
```

New query parameter:
- `filter[category_id]` — filter by category ID; use `null` to filter uncategorized

Added fields in response attributes:
```json
{
  "attributes": {
    "category_id": "550e8400-e29b-41d4-a716-446655440001",
    "category_name": "Produce"
  }
}
```

### Plan Ingredients (existing)

```
GET /api/v1/meals/plans/{planId}/ingredients
```

Added fields in response attributes:
```json
{
  "attributes": {
    "category_name": "Produce",
    "category_sort_order": 1
  }
}
```

Both fields are `null` for uncategorized/unresolved ingredients.
