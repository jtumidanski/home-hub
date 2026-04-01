# Shopping List — API Contracts

## Category Service

Base URL: `/api/v1/categories`
JSON:API type: `"categories"`

---

### List Categories

`GET /categories`

Returns all categories for the tenant, sorted by `sort_order`.

**Response: 200 OK**
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "categories",
      "attributes": {
        "name": "Produce",
        "sort_order": 0,
        "created_at": "2026-03-31T00:00:00Z",
        "updated_at": "2026-03-31T00:00:00Z"
      }
    }
  ]
}
```

### Create Category

`POST /categories`

**Request:**
```json
{
  "data": {
    "type": "categories",
    "attributes": {
      "name": "Household"
    }
  }
}
```

**Response: 201 Created** — same shape as list item.

**Errors:** 400 (name required), 409 (duplicate name).

### Update Category

`PATCH /categories/{id}`

**Request:**
```json
{
  "data": {
    "type": "categories",
    "id": "uuid",
    "attributes": {
      "name": "Household Supplies",
      "sort_order": 12
    }
  }
}
```

Both fields optional. **Response: 200 OK.**

**Errors:** 400 (validation), 404 (not found), 409 (duplicate name).

### Delete Category

`DELETE /categories/{id}`

**Response: 204 No Content.**

**Errors:** 404 (not found).

Consumers handle orphaned references. Shopping items retain their denormalized category_name/sort_order. Recipe-service nullifies canonical_ingredients.category_id locally.

---

## Shopping Service

Base URL: `/api/v1/shopping`

---

### List Shopping Lists

`GET /shopping/lists?status=active`

Query params:
- `status` — `active` (default) or `archived`

**Response: 200 OK**
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "shopping-lists",
      "attributes": {
        "name": "Weekly Groceries",
        "status": "active",
        "item_count": 24,
        "checked_count": 0,
        "archived_at": null,
        "created_at": "2026-03-31T00:00:00Z",
        "updated_at": "2026-03-31T00:00:00Z"
      }
    }
  ]
}
```

### Create Shopping List

`POST /shopping/lists`

**Request:**
```json
{
  "data": {
    "type": "shopping-lists",
    "attributes": {
      "name": "Costco Run"
    }
  }
}
```

**Response: 201 Created.**

**Errors:** 400 (name required).

### Get Shopping List Detail

`GET /shopping/lists/{id}`

Returns the list with all items, grouped by category in the response.

**Response: 200 OK**
```json
{
  "data": {
    "id": "uuid",
    "type": "shopping-lists",
    "attributes": {
      "name": "Weekly Groceries",
      "status": "active",
      "item_count": 3,
      "checked_count": 1,
      "archived_at": null,
      "created_at": "2026-03-31T00:00:00Z",
      "updated_at": "2026-03-31T00:00:00Z",
      "items": [
        {
          "id": "uuid",
          "name": "Chicken breast",
          "quantity": "2 lb",
          "category_id": "uuid",
          "category_name": "Meats & Seafood",
          "category_sort_order": 1,
          "checked": false,
          "position": 0,
          "created_at": "2026-03-31T00:00:00Z",
          "updated_at": "2026-03-31T00:00:00Z"
        },
        {
          "id": "uuid",
          "name": "Laundry detergent",
          "quantity": "1 big bottle",
          "category_id": "uuid",
          "category_name": "Household",
          "category_sort_order": 11,
          "checked": false,
          "position": 0,
          "created_at": "2026-03-31T00:00:00Z",
          "updated_at": "2026-03-31T00:00:00Z"
        },
        {
          "id": "uuid",
          "name": "Yellow onion",
          "quantity": "3",
          "category_id": "uuid",
          "category_name": "Produce",
          "category_sort_order": 0,
          "checked": true,
          "position": 0,
          "created_at": "2026-03-31T00:00:00Z",
          "updated_at": "2026-03-31T00:00:00Z"
        }
      ]
    }
  }
}
```

### Update Shopping List

`PATCH /shopping/lists/{id}`

**Request:**
```json
{
  "data": {
    "type": "shopping-lists",
    "id": "uuid",
    "attributes": {
      "name": "Renamed List"
    }
  }
}
```

**Response: 200 OK.**

**Errors:** 400 (validation), 404 (not found), 409 (list is archived).

### Delete Shopping List

`DELETE /shopping/lists/{id}`

Works for both active and archived lists.

**Response: 204 No Content.**

### Archive Shopping List

`POST /shopping/lists/{id}/archive`

Transitions list status from `active` to `archived`. Sets `archived_at` timestamp.

**Response: 200 OK** — returns updated list (without items).

**Errors:** 404 (not found), 409 (already archived).

### Unarchive Shopping List

`POST /shopping/lists/{id}/unarchive`

Transitions list status from `archived` back to `active`. Clears `archived_at` timestamp.

**Response: 200 OK** — returns updated list (without items).

**Errors:** 404 (not found), 409 (not archived).

---

### Add Item

`POST /shopping/lists/{id}/items`

**Request:**
```json
{
  "data": {
    "type": "shopping-items",
    "attributes": {
      "name": "Milk",
      "quantity": "1 gallon",
      "category_id": "uuid"
    }
  }
}
```

All fields except `name` are optional. If `category_id` is provided, the service fetches `category_name` and `category_sort_order` from category-service and denormalizes them onto the item.

**Response: 201 Created.**

**Errors:** 400 (name required), 404 (list not found), 409 (list is archived).

### Update Item

`PATCH /shopping/lists/{id}/items/{itemId}`

**Request:**
```json
{
  "data": {
    "type": "shopping-items",
    "id": "uuid",
    "attributes": {
      "name": "Whole milk",
      "quantity": "2 gallons",
      "category_id": "uuid",
      "position": 1
    }
  }
}
```

All fields optional. If `category_id` changes, re-fetches and updates denormalized category fields.

**Response: 200 OK.**

**Errors:** 404, 409 (archived).

### Remove Item

`DELETE /shopping/lists/{id}/items/{itemId}`

**Response: 204 No Content.**

**Errors:** 404, 409 (archived).

### Check/Uncheck Item

`PATCH /shopping/lists/{id}/items/{itemId}/check`

**Request:**
```json
{
  "data": {
    "type": "shopping-items",
    "id": "uuid",
    "attributes": {
      "checked": true
    }
  }
}
```

This is the only item mutation allowed on archived lists: **No** — checking is also blocked on archived lists. This endpoint only works on active lists.

**Response: 200 OK.**

**Errors:** 404, 409 (archived).

### Uncheck All Items

`POST /shopping/lists/{id}/items/uncheck-all`

Sets `checked = false` on all items in the list. Useful for resetting a list to re-shop.

**Response: 200 OK** — returns updated list detail (with all items).

**Errors:** 404 (list not found), 409 (list is archived).

---

### Import from Meal Plan

`POST /shopping/lists/{id}/import/meal-plan`

**Request:**
```json
{
  "data": {
    "type": "shopping-list-imports",
    "attributes": {
      "plan_id": "uuid"
    }
  }
}
```

**Behavior:**
1. Shopping-service calls `GET /api/v1/meals/plans/{planId}/ingredients` on recipe-service (service-to-service, forwarding tenant context).
2. For each consolidated ingredient:
   - `name` = `display_name` or `name`
   - `quantity` = `"{quantity} {unit}"` formatted as freeform string (e.g., "2 lb", "3", "1.5 cup")
   - `category_id` = resolved from ingredient's category via category-service lookup (if the ingredient has a category)
   - `category_name`, `category_sort_order` = denormalized from category-service
   - `checked` = false
3. Items are appended; existing items are untouched.
4. Extra quantities (cross-family overflow) become separate items.

**Response: 200 OK** — returns the updated list detail (with all items).

**Errors:** 400 (plan_id required), 404 (list or plan not found), 409 (list is archived).
