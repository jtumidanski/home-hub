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

---

### POST `/recipes/{id}/ingredients/{ingredientId}/resolve`

Manually resolve a recipe ingredient to a canonical ingredient.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Recipe ID (path) |
| ingredientId | UUID | Recipe ingredient ID (path) |

**Request model** (`ingredient-resolutions`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| canonicalIngredientId | string (UUID) | yes | Target canonical ingredient ID |
| saveAsAlias | bool | no | Whether to save the raw name as an alias |

**Response model** (`recipe-ingredients`):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Recipe ingredient ID |
| rawName | string | Original ingredient name |
| rawQuantity | string | Original quantity |
| rawUnit | string | Original unit |
| position | int | Position in recipe |
| canonicalIngredientId | *UUID | Resolved canonical ingredient |
| canonicalName | string | Canonical ingredient name |
| canonicalUnit | string | Resolved canonical unit |
| canonicalUnitFamily | string | Resolved unit family |
| normalizationStatus | string | Resolution status |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | ingredientId or canonicalIngredientId is not a valid UUID |
| 404 | Recipe ingredient not found |
| 500 | Internal error |

---

### POST `/recipes/{id}/renormalize`

Re-run normalization for all non-manually-confirmed ingredients in a recipe.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Recipe ID (path) |

**Request model** (`recipe-renormalize`): empty body.

**Response model** (`recipe-ingredients` list): array of recipe ingredient models (see resolve endpoint response).

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

---

### GET `/ingredients`

List canonical ingredients with usage counts and category names.

**Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `search` | string | — | Case-insensitive substring match on name or alias |
| `filter[category_id]` | string | — | Filter by category ID. Use "null" for uncategorized |
| `page[number]` | int | 1 | Page number |
| `page[size]` | int | 20 | Items per page (max 200) |

**Response model** (`ingredients` list):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Ingredient ID |
| name | string | Canonical name |
| displayName | string | Display name |
| unitFamily | string | Unit family (count, weight, volume) |
| categoryId | *UUID | Category ID |
| categoryName | string | Category name |
| aliasCount | int | Number of aliases |
| usageCount | int | Number of recipe ingredients using this |
| createdAt | timestamp | Creation time |
| updatedAt | timestamp | Last update time |

**Response meta:**

| Field | Type | Description |
|-------|------|-------------|
| total | int | Total matching ingredients |
| page | int | Current page number |
| pageSize | int | Items per page |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

---

### GET `/ingredients/lookup`

Resolve a free-form ingredient name to a canonical ingredient for the calling tenant. Used by shopping-service to auto-categorize manually added items, but available to any caller. Tries exact name match, then alias match, then a normalized variant (strips leading articles `the`/`a`/`an` and a trailing plural `s`) against both names and aliases.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Ingredient name to resolve |

**Response model** (`ingredient-lookups`):

| Field | Type | Description |
|-------|------|-------------|
| name | string | Canonical name |
| display_name | string | Display name |
| category_id | *UUID | Category ID, or null if uncategorized |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | `name` query parameter is missing |
| 404 | No canonical ingredient (or alias / normalized variant) matches |
| 500 | Internal error |

---

### POST `/ingredients`

Create a canonical ingredient.

**Request model** (`ingredients`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Ingredient name (normalized to lowercase) |
| displayName | string | no | Display name |
| unitFamily | string | no | Unit family: count, weight, volume |
| categoryId | string (UUID) | no | Category ID |

**Response model** (`ingredients` detail): see GET `/ingredients/{id}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | Name is required |
| 422 | Invalid unit family |
| 422 | categoryId is not a valid UUID |
| 500 | Internal error |

---

### GET `/ingredients/{id}`

Get a canonical ingredient with aliases.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Ingredient ID (path) |

**Response model** (`ingredients` detail):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Ingredient ID |
| name | string | Canonical name |
| displayName | string | Display name |
| unitFamily | string | Unit family |
| categoryId | *UUID | Category ID |
| categoryName | string | Category name |
| aliases | []Alias | List of aliases (id, name) |
| createdAt | timestamp | Creation time |
| updatedAt | timestamp | Last update time |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Ingredient not found |

---

### PATCH `/ingredients/{id}`

Partially update a canonical ingredient.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Ingredient ID (path) |

**Request model** (`ingredients`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | no | Updated name |
| displayName | string | no | Updated display name |
| unitFamily | string | no | Updated unit family |
| categoryId | *string | no | Updated category ID (empty string to clear) |

**Response model** (`ingredients` detail): see GET `/ingredients/{id}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Ingredient not found |
| 422 | Invalid unit family |
| 422 | categoryId is not a valid UUID |
| 500 | Internal error |

---

### DELETE `/ingredients/{id}`

Delete a canonical ingredient. Nullifies all recipe ingredient references and sets them to unresolved.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Ingredient ID (path) |

**Response:** 204 No Content

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 409 | Ingredient is still referenced (when ErrHasReferences is returned) |
| 500 | Internal error |

---

### POST `/ingredients/{id}/aliases`

Add an alias to a canonical ingredient.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Ingredient ID (path) |

**Request model** (`ingredient-aliases`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Alias name |

**Response model** (`ingredients` detail): returns the updated ingredient with aliases.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 409 | Alias conflicts with existing canonical name or alias |
| 500 | Internal error |

---

### DELETE `/ingredients/{id}/aliases/{aliasId}`

Remove an alias from a canonical ingredient.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Ingredient ID (path) |
| aliasId | UUID | Alias ID (path) |

**Response:** 204 No Content

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | aliasId is not a valid UUID |
| 500 | Internal error |

---

### GET `/ingredients/{id}/recipes`

List recipe references for a canonical ingredient.

**Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| id | UUID | — | Ingredient ID (path) |
| `page[number]` | int | 1 | Page number |
| `page[size]` | int | 20 | Items per page (max 100) |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| data | []RecipeRef | Array of {recipeId, rawName} |
| meta.total | int | Total references |
| meta.page | int | Current page |
| meta.pageSize | int | Items per page |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

---

### POST `/ingredients/{id}/reassign`

Reassign all recipe ingredient references from one canonical ingredient to another, then delete the source ingredient.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | UUID | Source ingredient ID (path) |

**Request model** (`ingredient-reassignments`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| targetIngredientId | string (UUID) | yes | Target canonical ingredient ID |

**Response:**

| Field | Type | Description |
|-------|------|-------------|
| meta.reassigned | int | Number of references reassigned |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | targetIngredientId is not a valid UUID |
| 500 | Internal error |

---

### POST `/ingredients/bulk-categorize`

Assign a category to multiple canonical ingredients.

**Request model** (`ingredient-bulk-categorize`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| ingredient_ids | []string (UUID) | yes | List of ingredient IDs (max 200) |
| category_id | string (UUID) | yes | Category to assign |

**Response:** 204 No Content

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 422 | ingredient_ids is empty |
| 422 | ingredient_ids exceeds 200 items |
| 422 | category_id is not a valid UUID |
| 422 | Any ingredient_id is not a valid UUID |
| 500 | Internal error |

---

### POST `/meals/plans`

Create a weekly meal plan.

**Request model** (`plans`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| starts_on | string (YYYY-MM-DD) | yes | Week start date |
| name | string | no | Plan name (defaults to "Week of {date}") |

**Response model** (`plans` detail): see GET `/meals/plans/{planId}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | starts_on is required or invalid date format |
| 409 | A plan already exists for this household and week |
| 500 | Internal error |

---

### GET `/meals/plans`

List meal plans with pagination and optional date filter.

**Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `starts_on` | string (YYYY-MM-DD) | — | Filter by exact start date |
| `page[number]` | int | 1 | Page number |
| `page[size]` | int | 20 | Items per page (max 100) |

**Response model** (`plans` list):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Plan ID |
| starts_on | string | Week start date |
| name | string | Plan name |
| locked | bool | Lock status |
| item_count | int | Number of items in the plan |
| created_at | timestamp | Creation time |
| updated_at | timestamp | Last update time |

**Response meta:**

| Field | Type | Description |
|-------|------|-------------|
| total | int | Total matching plans |
| page | int | Current page number |
| pageSize | int | Items per page |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 500 | Internal error |

---

### GET `/meals/plans/{planId}`

Get a meal plan with its items enriched with recipe metadata.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Response model** (`plans` detail):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Plan ID |
| starts_on | string | Week start date |
| name | string | Plan name |
| locked | bool | Lock status |
| created_by | UUID | Creator user ID |
| items | []PlanItem | Meal assignments (see below) |
| created_at | timestamp | Creation time |
| updated_at | timestamp | Last update time |

**Plan item fields:**

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Item ID |
| day | string | Date (YYYY-MM-DD) |
| slot | string | Meal slot |
| recipe_id | UUID | Assigned recipe |
| recipe_title | string | Recipe title (or "(deleted recipe)") |
| recipe_servings | *int | Recipe serving count |
| recipe_classification | string | Planner classification |
| recipe_deleted | bool | Whether the recipe has been deleted |
| serving_multiplier | *float64 | Serving multiplier |
| planned_servings | *int | Explicit planned servings |
| notes | *string | Optional notes |
| position | int | Position within day+slot |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |

---

### PATCH `/meals/plans/{planId}`

Update a meal plan name.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Request model** (`plans`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Updated plan name |

**Response model** (`plans` detail): see GET `/meals/plans/{planId}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 409 | Plan is locked |
| 500 | Internal error |

---

### POST `/meals/plans/{planId}/lock`

Lock a meal plan.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Response model** (`plans` detail): see GET `/meals/plans/{planId}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 409 | Plan is already locked |
| 500 | Internal error |

---

### POST `/meals/plans/{planId}/unlock`

Unlock a meal plan.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Response model** (`plans` detail): see GET `/meals/plans/{planId}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 409 | Plan is not locked |
| 500 | Internal error |

---

### POST `/meals/plans/{planId}/duplicate`

Duplicate a meal plan to a new week, copying all items with adjusted dates.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Source plan ID (path) |

**Request model** (`plans`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| starts_on | string (YYYY-MM-DD) | yes | Target week start date |

**Response model** (`plans` detail): see GET `/meals/plans/{planId}` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | starts_on is invalid date format |
| 404 | Source plan not found |
| 409 | A plan already exists for the target week |
| 500 | Internal error |

---

### POST `/meals/plans/{planId}/items`

Add a meal assignment to a plan.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Request model** (`plan-items`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| day | string (YYYY-MM-DD) | yes | Date for the meal |
| slot | string | yes | Meal slot: breakfast, lunch, dinner, snack, side |
| recipe_id | string (UUID) | yes | Recipe to assign |
| serving_multiplier | *float64 | no | Serving multiplier |
| planned_servings | *int | no | Explicit planned servings |
| notes | *string | no | Optional notes |
| position | *int | no | Position (auto-assigned if omitted) |

**Response model** (`plan-items`):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Item ID |
| day | string | Date |
| slot | string | Meal slot |
| recipe_id | UUID | Recipe ID |
| serving_multiplier | *float64 | Serving multiplier |
| planned_servings | *int | Planned servings |
| notes | *string | Notes |
| position | int | Position |
| created_at | timestamp | Creation time |
| updated_at | timestamp | Last update time |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | day is invalid or outside plan week |
| 400 | recipe_id is not a valid UUID |
| 400 | Invalid slot value |
| 404 | Plan not found |
| 404 | Recipe not found or deleted |
| 409 | Plan is locked |
| 500 | Internal error |

---

### PATCH `/meals/plans/{planId}/items/{itemId}`

Partially update a plan item.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |
| itemId | UUID | Item ID (path) |

**Request model** (`plan-items`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| day | string (YYYY-MM-DD) | no | Updated date |
| slot | string | no | Updated meal slot |
| serving_multiplier | *float64 | no | Updated serving multiplier |
| planned_servings | *int | no | Updated planned servings |
| notes | *string | no | Updated notes |
| position | *int | no | Updated position |

**Response model** (`plan-items`): see POST `/meals/plans/{planId}/items` response.

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 400 | day is invalid or outside plan week |
| 400 | Invalid slot value |
| 404 | Plan not found |
| 404 | Plan item not found |
| 409 | Plan is locked |
| 500 | Internal error |

---

### DELETE `/meals/plans/{planId}/items/{itemId}`

Remove a plan item.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |
| itemId | UUID | Item ID (path) |

**Response:** 204 No Content

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |
| 404 | Plan item not found |
| 409 | Plan is locked |
| 500 | Internal error |

---

### GET `/meals/plans/{planId}/export/markdown`

Export a meal plan as markdown with daily schedule and shopping list.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Response:** `text/markdown; charset=utf-8`

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |

---

### GET `/meals/plans/{planId}/ingredients`

Get consolidated ingredients for a meal plan.

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| planId | UUID | Plan ID (path) |

**Response model** (`plan-ingredients` list):

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Generated ID |
| name | string | Ingredient name |
| display_name | *string | Display name |
| quantity | float64 | Consolidated quantity |
| unit | string | Display unit |
| unit_family | string | Unit family |
| resolved | bool | Whether the ingredient is resolved to a canonical ingredient |
| extra_quantities | []QuantityUnit | Additional quantity+unit pairs for cross-family grouping |
| category_name | *string | Category name |
| category_sort_order | *int | Category sort order |

**Error conditions:**

| Status | Condition |
|--------|-----------|
| 404 | Plan not found |

## Resource Types

| Type | Description |
|------|-------------|
| `recipes` | Recipe resource (list and detail variants) |
| `recipe-tags` | Tag with usage count |
| `recipe-parse` | Parse result (ingredients, steps, metadata, errors) |
| `recipe-restorations` | Restoration request |
| `recipe-ingredients` | Recipe ingredient with normalization status |
| `ingredient-resolutions` | Ingredient resolution request |
| `recipe-renormalize` | Renormalization request |
| `ingredients` | Canonical ingredient resource (list and detail variants) |
| `ingredient-aliases` | Alias creation request |
| `ingredient-reassignments` | Reassignment request |
| `ingredient-bulk-categorize` | Bulk categorization request |
| `plans` | Meal plan resource (list and detail variants) |
| `plan-items` | Plan item resource |
| `plan-ingredients` | Consolidated ingredient for export |
