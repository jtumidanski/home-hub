# Recipe Service — REST Endpoints

Base path: `/api/v1`

All endpoints require a valid JWT. Tenant and household scoping is derived from JWT claims.

## Endpoints

### POST `/recipes/parse`

Parse Cooklang source text without persistence.

**Request model** (`recipe-parse`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| source | string | yes | Cooklang source text |

**Response model** (`recipe-parse`):

| Field | Type | Description |
|-------|------|-------------|
| ingredients | []Ingredient | Parsed ingredient list |
| steps | []Step | Parsed step list |
| metadata | Metadata | Extracted metadata |
| errors | []ParseError | Validation errors (if any) |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | Source exceeds 64 KB |

---

### GET `/recipes`

List recipes with pagination, search, and tag filtering.

**Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `search` | string | — | Case-insensitive substring match on title |
| `tag` | string | — | Repeatable. Filter by tag(s) with AND semantics |
| `page[number]` | int | 1 | Page number |
| `page[size]` | int | 20 | Items per page (max 100) |

**Response model** (`recipes` list):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Recipe ID |
| title | string | Recipe title |
| description | string | Optional description |
| servings | *int | Optional serving count |
| prepTimeMinutes | *int | Optional prep time in minutes |
| cookTimeMinutes | *int | Optional cook time in minutes |
| tags | []string | Tags |
| createdAt | timestamp | Creation time |
| updatedAt | timestamp | Last update time |

**Response meta:**

| Field | Type | Description |
|-------|------|-------------|
| total | int | Total matching recipes |
| page | int | Current page number |
| pageSize | int | Items per page |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

---

### POST `/recipes`

Create a new recipe.

**Request model** (`recipes`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | yes | Recipe title |
| description | string | no | Description |
| source | string | yes | Cooklang source text |
| servings | *int | no | Serving count |
| prepTimeMinutes | *int | no | Prep time in minutes |
| cookTimeMinutes | *int | no | Cook time in minutes |
| sourceUrl | string | no | Source URL |
| tags | []string | no | Tags |

**Response model** (`recipes` detail): see GET `/recipes/{id}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | Title is required |
| 400 | Source is required |
| 422 | Invalid Cooklang syntax (returns array of parse errors with line/column/message) |
| 500 | Internal error |

---

### GET `/recipes/{id}`

Get a single recipe with parsed Cooklang data.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Recipe ID (path) |

**Response model** (`recipes` detail):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Recipe ID |
| title | string | Recipe title |
| description | string | Optional description |
| servings | *int | Optional serving count |
| prepTimeMinutes | *int | Optional prep time in minutes |
| cookTimeMinutes | *int | Optional cook time in minutes |
| sourceUrl | string | Optional source URL |
| tags | []string | Tags |
| source | string | Raw Cooklang source |
| ingredients | []Ingredient | Parsed ingredients |
| steps | []Step | Parsed steps |
| createdAt | timestamp | Creation time |
| updatedAt | timestamp | Last update time |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Recipe not found |

---

### PATCH `/recipes/{id}`

Partially update a recipe.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Recipe ID (path) |

**Request model** (`recipes`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | no | Recipe title |
| description | string | no | Description |
| source | string | no | Cooklang source text |
| servings | *int | no | Serving count |
| prepTimeMinutes | *int | no | Prep time in minutes |
| cookTimeMinutes | *int | no | Cook time in minutes |
| sourceUrl | string | no | Source URL |
| tags | []string | no | Tags |

**Response model** (`recipes` detail): see GET `/recipes/{id}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Recipe not found |
| 422 | Invalid Cooklang syntax (when source is updated) |
| 500 | Internal error |

---

### DELETE `/recipes/{id}`

Soft-delete a recipe.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Recipe ID (path) |

**Response:** 204 No Content

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

---

### POST `/recipes/restorations`

Restore a soft-deleted recipe within the 3-day restore window.

**Request model** (`recipe-restorations`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| recipeId | string (UUID) | yes | ID of the recipe to restore |

**Response model** (`recipes` detail): see GET `/recipes/{id}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | recipeId is not a valid UUID |
| 400 | Recipe is not deleted |
| 404 | Recipe not found |
| 410 | Restore window expired |
| 500 | Internal error |

---

### GET `/recipes/tags`

List all tags in use across non-deleted recipes with usage counts.

**Response model** (`recipe-tags` list):

| Field | Type | Description |
|-------|------|-------------|
| tag | string | Tag value (used as ID) |
| count | int | Number of recipes using this tag |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

## Resource Types

| Type | Description |
|------|-------------|
| `recipes` | Recipe resource (list and detail variants) |
| `recipe-tags` | Tag with usage count |
| `recipe-parse` | Parse result (ingredients, steps, metadata, errors) |
| `recipe-restorations` | Restoration request |
