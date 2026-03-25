# Recipe Management — API Contracts

## Base Path

`/api/v1/recipes`

All endpoints require a valid JWT. Tenant and household context are derived from the JWT claims and active context.

---

## POST /api/v1/recipes/parse

Parse Cooklang source text and return structured ingredients and steps. Used by the live preview editor. Does not persist anything.

**Request:**
```json
{
  "data": {
    "type": "recipe-parse",
    "attributes": {
      "source": "Bring @water{2%l} to a boil in a #large pot{}.\n\nCook @spaghetti{400%g} until al dente."
    }
  }
}
```

**Response: 200 OK**
```json
{
  "data": {
    "type": "recipe-parse",
    "attributes": {
      "ingredients": [
        { "name": "water", "quantity": "2", "unit": "l" },
        { "name": "spaghetti", "quantity": "400", "unit": "g" }
      ],
      "steps": [
        {
          "number": 1,
          "segments": [
            { "type": "text", "value": "Bring " },
            { "type": "ingredient", "name": "water", "quantity": "2", "unit": "l" },
            { "type": "text", "value": " to a boil in a " },
            { "type": "cookware", "name": "large pot" },
            { "type": "text", "value": "." }
          ]
        },
        {
          "number": 2,
          "segments": [
            { "type": "text", "value": "Cook " },
            { "type": "ingredient", "name": "spaghetti", "quantity": "400", "unit": "g" },
            { "type": "text", "value": " until al dente." }
          ]
        }
      ]
    }
  }
}
```

**Response: 200 OK (with errors)**

When the source has syntax errors, the endpoint still returns 200 with partial results plus an errors array. This allows the preview to show what it can while highlighting problems.

```json
{
  "data": {
    "type": "recipe-parse",
    "attributes": {
      "ingredients": [],
      "steps": [],
      "errors": [
        { "line": 1, "column": 15, "message": "Unclosed ingredient block" }
      ]
    }
  }
}
```

---

## POST /api/v1/recipes

Create a new recipe.

**Request:**
```json
{
  "data": {
    "type": "recipes",
    "attributes": {
      "title": "Pasta Carbonara",
      "description": "Classic Roman pasta dish with eggs, cheese, and guanciale",
      "servings": 4,
      "prep-time-minutes": 10,
      "cook-time-minutes": 20,
      "source-url": "https://example.com/carbonara",
      "tags": ["italian", "pasta", "quick"],
      "source": "Bring @water{2%l} to a boil in a #large pot{}.\n\nCook @spaghetti{400%g} until al dente, about ~{8%minutes}.\n\nMeanwhile, dice @guanciale{200%g} and fry in a #skillet{} until crispy.\n\nWhisk @eggs{3} with @pecorino romano{100%g} and @black pepper{1%tsp}.\n\nDrain pasta, toss with guanciale, then mix in egg mixture off heat."
    }
  }
}
```

**Validation rules:**
- `title`: required, max 255 characters
- `source`: required, must be valid Cooklang syntax
- `description`: optional, max 2000 characters
- `servings`: optional, positive integer
- `prep-time-minutes`: optional, non-negative integer
- `cook-time-minutes`: optional, non-negative integer
- `source-url`: optional, max 2048 characters
- `tags`: optional, array of strings, each max 100 characters, normalized to lowercase

**Response: 201 Created**
```json
{
  "data": {
    "type": "recipes",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "attributes": {
      "title": "Pasta Carbonara",
      "description": "Classic Roman pasta dish with eggs, cheese, and guanciale",
      "servings": 4,
      "prep-time-minutes": 10,
      "cook-time-minutes": 20,
      "source-url": "https://example.com/carbonara",
      "tags": ["italian", "pasta", "quick"],
      "source": "Bring @water{2%l} to a boil in a #large pot{}.\n\nCook @spaghetti{400%g} until al dente, about ~{8%minutes}.\n\nMeanwhile, dice @guanciale{200%g} and fry in a #skillet{} until crispy.\n\nWhisk @eggs{3} with @pecorino romano{100%g} and @black pepper{1%tsp}.\n\nDrain pasta, toss with guanciale, then mix in egg mixture off heat.",
      "ingredients": [
        { "name": "water", "quantity": "2", "unit": "l" },
        { "name": "spaghetti", "quantity": "400", "unit": "g" },
        { "name": "guanciale", "quantity": "200", "unit": "g" },
        { "name": "eggs", "quantity": "3", "unit": "" },
        { "name": "pecorino romano", "quantity": "100", "unit": "g" },
        { "name": "black pepper", "quantity": "1", "unit": "tsp" }
      ],
      "steps": [
        {
          "number": 1,
          "segments": [
            { "type": "text", "value": "Bring " },
            { "type": "ingredient", "name": "water", "quantity": "2", "unit": "l" },
            { "type": "text", "value": " to a boil in a " },
            { "type": "cookware", "name": "large pot" },
            { "type": "text", "value": "." }
          ]
        },
        {
          "number": 2,
          "segments": [
            { "type": "text", "value": "Cook " },
            { "type": "ingredient", "name": "spaghetti", "quantity": "400", "unit": "g" },
            { "type": "text", "value": " until al dente, about " },
            { "type": "timer", "quantity": "8", "unit": "minutes" },
            { "type": "text", "value": "." }
          ]
        },
        {
          "number": 3,
          "segments": [
            { "type": "text", "value": "Meanwhile, dice " },
            { "type": "ingredient", "name": "guanciale", "quantity": "200", "unit": "g" },
            { "type": "text", "value": " and fry in a " },
            { "type": "cookware", "name": "skillet" },
            { "type": "text", "value": " until crispy." }
          ]
        },
        {
          "number": 4,
          "segments": [
            { "type": "text", "value": "Whisk " },
            { "type": "ingredient", "name": "eggs", "quantity": "3", "unit": "" },
            { "type": "text", "value": " with " },
            { "type": "ingredient", "name": "pecorino romano", "quantity": "100", "unit": "g" },
            { "type": "text", "value": " and " },
            { "type": "ingredient", "name": "black pepper", "quantity": "1", "unit": "tsp" },
            { "type": "text", "value": "." }
          ]
        },
        {
          "number": 5,
          "segments": [
            { "type": "text", "value": "Drain pasta, toss with guanciale, then mix in egg mixture off heat." }
          ]
        }
      ],
      "created-at": "2026-03-25T12:00:00Z",
      "updated-at": "2026-03-25T12:00:00Z"
    }
  }
}
```

**Error: 422 Unprocessable Entity (Cooklang validation)**
```json
{
  "errors": [
    {
      "status": "422",
      "title": "Invalid Cooklang syntax",
      "detail": "Unclosed ingredient block at line 3, column 15: '@guanciale{200%g'",
      "source": { "pointer": "/data/attributes/source" }
    }
  ]
}
```

**Error: 422 Unprocessable Entity (field validation)**
```json
{
  "errors": [
    {
      "status": "422",
      "title": "Validation error",
      "detail": "Title is required",
      "source": { "pointer": "/data/attributes/title" }
    }
  ]
}
```

---

## GET /api/v1/recipes

List recipes for the current household. Returns summary data (no parsed Cooklang).

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `search` | string | — | Case-insensitive substring match on title |
| `tag` | string | — | Repeatable. Filter by tag(s) with AND semantics |
| `page[number]` | int | 1 | Page number |
| `page[size]` | int | 20 | Items per page (max 100) |

**Examples:**
```
GET /api/v1/recipes?search=pasta
GET /api/v1/recipes?tag=italian&tag=quick
GET /api/v1/recipes?search=pasta&tag=italian&page[number]=2&page[size]=10
```

**Response: 200 OK**
```json
{
  "data": [
    {
      "type": "recipes",
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "attributes": {
        "title": "Pasta Carbonara",
        "description": "Classic Roman pasta dish with eggs, cheese, and guanciale",
        "servings": 4,
        "prep-time-minutes": 10,
        "cook-time-minutes": 20,
        "tags": ["italian", "pasta", "quick"],
        "created-at": "2026-03-25T12:00:00Z",
        "updated-at": "2026-03-25T12:00:00Z"
      }
    }
  ],
  "meta": {
    "total": 42,
    "page": 1,
    "page-size": 20
  }
}
```

---

## GET /api/v1/recipes/:id

Get a single recipe with full detail including raw Cooklang source and parsed data.

**Response: 200 OK**

Same as the create response body above.

**Error: 404 Not Found**
```json
{
  "errors": [
    {
      "status": "404",
      "title": "Not found",
      "detail": "Recipe not found"
    }
  ]
}
```

---

## PATCH /api/v1/recipes/:id

Update a recipe. Only provided fields are updated.

**Request:**
```json
{
  "data": {
    "type": "recipes",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "attributes": {
      "title": "Pasta Carbonara (Updated)",
      "tags": ["italian", "pasta", "classic"]
    }
  }
}
```

If `source` is included, Cooklang is re-validated. If `source` is omitted, existing source is preserved.

**Response: 200 OK** — full recipe detail (same as GET detail).

---

## DELETE /api/v1/recipes/:id

Soft delete a recipe.

**Response: 204 No Content**

---

## POST /api/v1/recipes/restorations

Restore a soft-deleted recipe.

**Request:**
```json
{
  "data": {
    "type": "recipe-restorations",
    "attributes": {
      "recipe-id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

**Response: 200 OK** — full recipe detail.

**Error: 410 Gone** — if restore window has expired.

---

## GET /api/v1/recipes/tags

List all distinct tags in use for the current household.

**Response: 200 OK**
```json
{
  "data": [
    {
      "type": "recipe-tags",
      "id": "italian",
      "attributes": {
        "tag": "italian",
        "count": 12
      }
    },
    {
      "type": "recipe-tags",
      "id": "quick",
      "attributes": {
        "tag": "quick",
        "count": 8
      }
    }
  ]
}
```
