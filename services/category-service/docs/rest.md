# REST API

All endpoints require JWT authentication. Tenant ID is extracted from the authentication context.

Responses use JSON:API marshaling via `api2go`. The resource type is `categories`.

## Endpoints

### GET /api/v1/categories

Returns all categories for the authenticated tenant, ordered by sort order.

**Parameters**: None.

**Request body**: None.

**Response model**:

```json
{
  "data": [
    {
      "type": "categories",
      "id": "<uuid>",
      "attributes": {
        "name": "Produce",
        "sort_order": 1,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    }
  ]
}
```

**Error conditions**:

| Status | Condition         |
|--------|-------------------|
| 500    | Internal failure  |

---

### POST /api/v1/categories

Creates a new category for the authenticated tenant.

**Parameters**: None.

**Request model**:

```json
{
  "data": {
    "type": "categories",
    "attributes": {
      "name": "Deli"
    }
  }
}
```

**Response model**: Single `categories` resource (same shape as list item).

**Error conditions**:

| Status | Condition                                        |
|--------|--------------------------------------------------|
| 409    | `ErrDuplicateName` -- name already exists        |
| 422    | `ErrNameRequired` -- name is empty               |
| 422    | `ErrNameTooLong` -- name exceeds 100 characters  |
| 500    | Internal failure                                 |

---

### PATCH /api/v1/categories/{id}

Updates an existing category. Name and sort order are both optional.

**Parameters**:

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request model**:

```json
{
  "data": {
    "type": "categories",
    "id": "<uuid>",
    "attributes": {
      "name": "New Name",
      "sort_order": 5
    }
  }
}
```

Both `name` and `sort_order` are optional (`omitempty`).

**Response model**: Single `categories` resource.

**Error conditions**:

| Status | Condition                                            |
|--------|------------------------------------------------------|
| 404    | Category not found                                   |
| 409    | `ErrDuplicateName` -- name taken by another category |
| 422    | `ErrNameRequired` -- name is empty after trim        |
| 422    | `ErrNameTooLong` -- name exceeds 100 characters      |
| 422    | `ErrInvalidSortOrder` -- sort order is negative      |
| 500    | Internal failure                                     |

---

### DELETE /api/v1/categories/{id}

Deletes a category.

**Parameters**:

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request body**: None.

**Response**: 204 No Content.

**Error conditions**:

| Status | Condition            |
|--------|----------------------|
| 404    | Category not found   |
| 500    | Internal failure     |
