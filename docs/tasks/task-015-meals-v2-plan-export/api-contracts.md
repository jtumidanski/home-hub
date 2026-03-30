# Meals v2 — API Contracts

## Base Path

All endpoints are under `/api/v1/meals`. Routed to `recipe-service` via nginx.

---

## Plan Management

### POST /api/v1/meals/plans

Create a new meal plan for a week.

**Request:**
```json
{
  "data": {
    "type": "plans",
    "attributes": {
      "starts_on": "2026-04-06",
      "name": "Week of April 6, 2026"
    }
  }
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| starts_on | string (date) | yes | Any date (ISO 8601); not constrained to a specific day of week |
| name | string | no | Auto-generated as "Week of {Month Day, Year}" if omitted |

**Response (201):**
```json
{
  "data": {
    "type": "plans",
    "id": "uuid",
    "attributes": {
      "starts_on": "2026-04-06",
      "name": "Week of April 6, 2026",
      "locked": false,
      "created_by": "uuid",
      "items": [],
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  }
}
```

**Errors:**
| Status | Condition |
|--------|-----------|
| 400 | starts_on missing or not a valid date |
| 409 | Plan already exists for this household + starts_on |

---

### GET /api/v1/meals/plans

List plans for the current household.

**Query Parameters:**
| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| starts_on | string (date) | no | Filter by exact week start date |
| page[number] | integer | no | Default 1 |
| page[size] | integer | no | Default 20, max 100 |

**Response (200):**
```json
{
  "data": [
    {
      "type": "plans",
      "id": "uuid",
      "attributes": {
        "starts_on": "2026-04-06",
        "name": "Week of April 6, 2026",
        "locked": true,
        "item_count": 7,
        "created_at": "2026-04-01T10:00:00Z",
        "updated_at": "2026-04-03T14:30:00Z"
      }
    }
  ]
}
```

Plans ordered by `starts_on` descending. List model includes `item_count` for summary display.

---

### GET /api/v1/meals/plans/{planId}

Get plan detail with all items.

**Response (200):**
```json
{
  "data": {
    "type": "plans",
    "id": "uuid",
    "attributes": {
      "starts_on": "2026-04-06",
      "name": "Week of April 6, 2026",
      "locked": false,
      "created_by": "uuid",
      "items": [
        {
          "id": "uuid",
          "day": "2026-04-06",
          "slot": "dinner",
          "recipe_id": "uuid",
          "recipe_title": "Chicken Tacos",
          "recipe_servings": 4,
          "recipe_classification": "dinner",
          "recipe_deleted": false,
          "serving_multiplier": 1.5,
          "planned_servings": null,
          "notes": "Double the garlic",
          "position": 0
        }
      ],
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  }
}
```

Items include denormalized recipe metadata (`recipe_title`, `recipe_servings`, `recipe_classification`, `recipe_deleted`) to avoid N+1 lookups in the UI. If a recipe has been soft-deleted, `recipe_deleted` is `true` and the title/servings reflect the last-known values.

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan not found |

---

### PATCH /api/v1/meals/plans/{planId}

Update plan name.

**Request:**
```json
{
  "data": {
    "type": "plans",
    "attributes": {
      "name": "Special Menu Week"
    }
  }
}
```

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 409 | Plan is locked |

---

### POST /api/v1/meals/plans/{planId}/lock

Lock the plan.

**Request:** Empty body.

**Response (200):** Updated plan resource with `locked: true`.

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 409 | Plan already locked |

---

### POST /api/v1/meals/plans/{planId}/unlock

Unlock the plan.

**Request:** Empty body.

**Response (200):** Updated plan resource with `locked: false`.

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 409 | Plan already unlocked |

---

### POST /api/v1/meals/plans/{planId}/duplicate

Duplicate a plan to another week.

**Request:**
```json
{
  "data": {
    "type": "plans",
    "attributes": {
      "starts_on": "2026-04-13"
    }
  }
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| starts_on | string (date) | yes | Target week start date |

Copies all plan items with day offsets adjusted to the target week. Item notes are not copied. New plan is created unlocked with an auto-generated name.

**Response (201):** New plan resource with copied items.

**Errors:**
| Status | Condition |
|--------|-----------|
| 400 | starts_on missing or not a valid date |
| 404 | Source plan not found |
| 409 | Plan already exists for target household + starts_on |

---

## Plan Items

### POST /api/v1/meals/plans/{planId}/items

Add a recipe to the plan.

**Request:**
```json
{
  "data": {
    "type": "plan-items",
    "attributes": {
      "day": "2026-04-06",
      "slot": "dinner",
      "recipe_id": "uuid",
      "serving_multiplier": 1.5,
      "planned_servings": null,
      "notes": "Double the garlic",
      "position": 0
    }
  }
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| day | string (date) | yes | Must fall within plan week (starts_on to starts_on + 6) |
| slot | string | yes | One of: breakfast, lunch, dinner, snack, side |
| recipe_id | string (UUID) | yes | Must reference active, non-deleted recipe in same tenant/household |
| serving_multiplier | decimal | no | Scaling factor for ingredients |
| planned_servings | integer | no | Takes precedence over serving_multiplier |
| notes | string | no | Free-form text |
| position | integer | no | Order within day+slot, default 0 |

**Response (201):** Plan item resource.

**Errors:**
| Status | Condition |
|--------|-----------|
| 400 | Invalid day (outside plan week), invalid slot, missing required fields |
| 404 | Plan not found, or recipe not found / deleted |
| 409 | Plan is locked |

---

### PATCH /api/v1/meals/plans/{planId}/items/{itemId}

Update a plan item.

**Request:**
```json
{
  "data": {
    "type": "plan-items",
    "attributes": {
      "day": "2026-04-07",
      "slot": "lunch",
      "serving_multiplier": 2.0,
      "notes": "Updated notes"
    }
  }
}
```

Any subset of updateable fields. `recipe_id` is not updatable — delete and re-add instead.

**Errors:**
| Status | Condition |
|--------|-----------|
| 400 | Invalid day or slot |
| 404 | Plan or item not found |
| 409 | Plan is locked |

---

### DELETE /api/v1/meals/plans/{planId}/items/{itemId}

Remove an item from the plan.

**Response:** 204 No Content.

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan or item not found |
| 409 | Plan is locked |

---

## Export

### GET /api/v1/meals/plans/{planId}/export/markdown

Generate markdown export of the plan with consolidated ingredients.

**Response (200):**
```
Content-Type: text/markdown; charset=utf-8
```

Body is raw markdown text (not JSON:API wrapped).

Emits `plan.exported` audit event on success.

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan not found |

---

## Ingredients (Convenience)

### GET /api/v1/meals/plans/{planId}/ingredients

Get consolidated ingredient list as structured JSON.

**Response (200):**
```json
{
  "data": [
    {
      "type": "plan-ingredients",
      "id": "canonical-ingredient-uuid",
      "attributes": {
        "name": "chicken breast",
        "display_name": "Chicken Breast",
        "quantity": 2.0,
        "unit": "pound",
        "unit_family": "weight",
        "resolved": true
      }
    },
    {
      "type": "plan-ingredients",
      "id": "generated-uuid",
      "attributes": {
        "name": "seasoning mix",
        "display_name": null,
        "quantity": 1.0,
        "unit": "packet",
        "unit_family": "",
        "resolved": false
      }
    }
  ]
}
```

Unresolved ingredients use a generated UUID as their ID and have `resolved: false`.

**Errors:**
| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
