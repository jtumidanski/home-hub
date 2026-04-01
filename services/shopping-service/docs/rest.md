# Shopping REST API

All endpoints are prefixed with `/api/v1` and require JWT authentication.

## Endpoints

### GET /shopping/lists

List shopping lists filtered by status.

**Parameters**

| Name | In | Type | Default | Description |
|---|---|---|---|---|
| status | query | string | `active` | Filter by status (`active` or `archived`) |

**Response**

JSON:API collection of `shopping-lists` resources.

| Field | Type | Description |
|---|---|---|
| name | string | List name |
| status | string | `active` or `archived` |
| item_count | int | Total item count |
| checked_count | int | Checked item count |
| archived_at | string/null | ISO 8601 timestamp or null |
| created_at | string | ISO 8601 timestamp |
| updated_at | string | ISO 8601 timestamp |

---

### POST /shopping/lists

Create a new shopping list.

**Request** (JSON:API resource type: `shopping-lists`)

| Field | Type | Required | Description |
|---|---|---|---|
| name | string | yes | List name |

**Response**: `201 Created` with the created `shopping-lists` resource.

**Errors**

| Status | Condition |
|---|---|
| 422 | Name is empty or exceeds 255 characters |

---

### GET /shopping/lists/{id}

Get a single shopping list with its items.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Response**

JSON:API `shopping-lists` resource with an `items` field containing an array of `shopping-items` resources.

`shopping-items` fields:

| Field | Type | Description |
|---|---|---|
| name | string | Item name |
| quantity | string/null | Quantity description |
| category_id | UUID/null | Category reference |
| category_name | string/null | Category name |
| category_sort_order | int/null | Category sort order |
| checked | bool | Checked state |
| position | int | Sort position |
| created_at | string | ISO 8601 timestamp |
| updated_at | string | ISO 8601 timestamp |

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |

---

### PATCH /shopping/lists/{id}

Update a shopping list name.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Request** (JSON:API resource type: `shopping-lists`)

| Field | Type | Required | Description |
|---|---|---|---|
| name | string | yes | New list name |

**Response**: Updated `shopping-lists` resource.

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |
| 409 | List is archived |
| 422 | Name is empty or exceeds 255 characters |

---

### DELETE /shopping/lists/{id}

Delete a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Response**: `204 No Content`

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |

---

### POST /shopping/lists/{id}/archive

Archive a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Response**: Updated `shopping-lists` resource with status `archived`.

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |
| 409 | List is already archived |

---

### POST /shopping/lists/{id}/unarchive

Unarchive a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Response**: Updated `shopping-lists` resource with status `active`.

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |
| 409 | List is not archived |

---

### POST /shopping/lists/{id}/items

Add an item to a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Request** (JSON:API resource type: `shopping-items`)

| Field | Type | Required | Description |
|---|---|---|---|
| name | string | yes | Item name |
| quantity | string | no | Quantity description |
| category_id | UUID | no | Category ID (resolved via category service) |
| position | int | no | Sort position (auto-assigned if omitted) |

**Response**: `201 Created` with the created `shopping-items` resource.

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |
| 409 | List is archived |
| 422 | Name is empty or exceeds 255 characters |

---

### PATCH /shopping/lists/{id}/items/{itemId}

Update a shopping item.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |
| itemId | path | UUID | Item ID |

**Request** (JSON:API resource type: `shopping-items`)

All fields are optional:

| Field | Type | Description |
|---|---|---|
| name | string | Item name |
| quantity | string | Quantity description |
| category_id | UUID | Category ID (resolved via category service) |
| position | int | Sort position |

**Response**: Updated `shopping-items` resource.

**Errors**

| Status | Condition |
|---|---|
| 404 | List or item not found |
| 409 | List is archived |
| 422 | Name is empty or exceeds 255 characters |

---

### DELETE /shopping/lists/{id}/items/{itemId}

Remove an item from a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |
| itemId | path | UUID | Item ID |

**Response**: `204 No Content`

**Errors**

| Status | Condition |
|---|---|
| 404 | List or item not found |
| 409 | List is archived |

---

### PATCH /shopping/lists/{id}/items/{itemId}/check

Toggle the checked state of a shopping item.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |
| itemId | path | UUID | Item ID |

**Request** (JSON:API resource type: `shopping-items`)

| Field | Type | Required | Description |
|---|---|---|---|
| checked | bool | yes | Checked state |

**Response**: Updated `shopping-items` resource.

**Errors**

| Status | Condition |
|---|---|
| 404 | List or item not found |
| 409 | List is archived |

---

### POST /shopping/lists/{id}/items/uncheck-all

Uncheck all items in a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Response**: Full `shopping-lists` resource with all items.

**Errors**

| Status | Condition |
|---|---|
| 404 | List not found |
| 409 | List is archived |

---

### POST /shopping/lists/{id}/import/meal-plan

Import ingredients from a meal plan into a shopping list.

**Parameters**

| Name | In | Type | Description |
|---|---|---|---|
| id | path | UUID | Shopping list ID |

**Request** (JSON:API resource type: `shopping-list-imports`)

| Field | Type | Required | Description |
|---|---|---|---|
| plan_id | UUID | yes | Meal plan ID to import from |

Fetches consolidated ingredients from the recipe service. Resolves category names against the category service. Creates one item per ingredient, plus separate items for extra quantities.

**Response**: Full `shopping-lists` resource with all items.

**Errors**

| Status | Condition |
|---|---|
| 400 | plan_id is missing or meal plan ingredients could not be fetched |
| 404 | List not found |
| 409 | List is archived |
