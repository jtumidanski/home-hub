# Ingredient Normalization — API Contracts

All endpoints require a valid JWT with tenant_id and household_id claims. All responses follow JSON:API format.

---

## Canonical Ingredients

### List Canonical Ingredients

```
GET /api/v1/ingredients?search=chick&page[number]=1&page[size]=20
```

The `search` parameter performs ILIKE prefix matching on canonical ingredient names and alias names. Results are ordered by usage count descending (most-referenced ingredients first), which ensures the resolve dropdown surfaces the most likely matches at the top.

**Response 200:**
```json
{
  "data": [
    {
      "type": "ingredients",
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "attributes": {
        "name": "chicken breast",
        "displayName": "Chicken Breast",
        "unitFamily": "weight",
        "aliasCount": 2,
        "usageCount": 8,
        "createdAt": "2026-03-27T12:00:00Z",
        "updatedAt": "2026-03-27T12:00:00Z"
      }
    }
  ],
  "meta": {
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

### Create Canonical Ingredient

```
POST /api/v1/ingredients
```

**Request:**
```json
{
  "data": {
    "type": "ingredients",
    "attributes": {
      "name": "chicken breast",
      "displayName": "Chicken Breast",
      "unitFamily": "weight"
    }
  }
}
```

**Response 201:** Single ingredient resource (same as list item shape).

**Error 400:** Name is empty or exceeds 255 characters.
**Error 409:** Name conflicts with existing canonical ingredient in this tenant.
**Error 422:** Invalid unit family value (must be `count`, `weight`, `volume`, or omitted).

### Get Canonical Ingredient

```
GET /api/v1/ingredients/:id
```

**Response 200:**
```json
{
  "data": {
    "type": "ingredients",
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "attributes": {
      "name": "chicken breast",
      "displayName": "Chicken Breast",
      "unitFamily": "weight",
      "aliases": [
        { "id": "660e8400-e29b-41d4-a716-446655440001", "name": "chicken breasts" },
        { "id": "660e8400-e29b-41d4-a716-446655440002", "name": "boneless chicken breast" }
      ],
      "usageCount": 8,
      "createdAt": "2026-03-27T12:00:00Z",
      "updatedAt": "2026-03-27T12:00:00Z"
    }
  }
}
```

### Update Canonical Ingredient

```
PATCH /api/v1/ingredients/:id
```

**Request:**
```json
{
  "data": {
    "type": "ingredients",
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "attributes": {
      "displayName": "Chicken Breast (Boneless)",
      "unitFamily": "weight"
    }
  }
}
```

**Response 200:** Updated ingredient resource.

### Delete Canonical Ingredient

```
DELETE /api/v1/ingredients/:id
```

**Response 204:** No content.
**Error 409:** Ingredient is referenced by one or more recipe ingredients. Response includes the reference count:
```json
{
  "errors": [
    {
      "status": "409",
      "title": "Conflict",
      "detail": "Cannot delete ingredient referenced by 12 recipe ingredients. Use /reassign to move references first.",
      "meta": { "referenceCount": 12 }
    }
  ]
}
```

### Reassign and Delete Canonical Ingredient

```
POST /api/v1/ingredients/:id/reassign
```

Reassigns all `recipe_ingredient` references from this canonical ingredient to a target canonical ingredient, then deletes this one. Useful when consolidating duplicates or removing a canonical ingredient that is actively referenced.

**Request:**
```json
{
  "data": {
    "type": "ingredient-reassignments",
    "attributes": {
      "targetIngredientId": "550e8400-e29b-41d4-a716-446655440002"
    }
  }
}
```

**Response 200:**
```json
{
  "data": {
    "type": "ingredient-reassignments",
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "attributes": {
      "reassignedCount": 12,
      "targetIngredientId": "550e8400-e29b-41d4-a716-446655440002",
      "deleted": true
    }
  }
}
```

**Side effects:**
- All `recipe_ingredient` rows pointing to this canonical ingredient are updated to point to the target
- Normalization status on reassigned recipe ingredients is preserved (they keep `matched`, `manually_confirmed`, etc.)
- The source canonical ingredient is deleted
- Aliases from the deleted ingredient are **not** transferred (they are deleted via CASCADE). If desired, the user should manually add them to the target before reassigning.

**Error 400:** Target ingredient ID is the same as the source, or does not exist.
**Error 404:** Source canonical ingredient not found.

### Add Alias

```
POST /api/v1/ingredients/:id/aliases
```

**Request:**
```json
{
  "data": {
    "type": "ingredient-aliases",
    "attributes": {
      "name": "chicken breasts"
    }
  }
}
```

**Response 201:**
```json
{
  "data": {
    "type": "ingredient-aliases",
    "id": "660e8400-e29b-41d4-a716-446655440003",
    "attributes": {
      "name": "chicken breasts",
      "canonicalIngredientId": "550e8400-e29b-41d4-a716-446655440001",
      "createdAt": "2026-03-27T12:00:00Z"
    }
  }
}
```

**Error 409:** Alias name conflicts with an existing canonical ingredient name or alias in this tenant.

### Remove Alias

```
DELETE /api/v1/ingredients/:id/aliases/:aliasId
```

**Response 204:** No content.
**Error 404:** Alias not found or does not belong to this canonical ingredient.

---

## Ingredient Normalization Correction

### Resolve Recipe Ingredient

```
POST /api/v1/recipes/:id/ingredients/:ingredientId/resolve
```

**Request:**
```json
{
  "data": {
    "type": "ingredient-resolutions",
    "attributes": {
      "canonicalIngredientId": "550e8400-e29b-41d4-a716-446655440001",
      "saveAsAlias": true
    }
  }
}
```

**Response 200:**
```json
{
  "data": {
    "type": "recipe-ingredients",
    "id": "770e8400-e29b-41d4-a716-446655440001",
    "attributes": {
      "rawName": "chkn breast",
      "rawQuantity": "500",
      "rawUnit": "g",
      "position": 3,
      "canonicalIngredientId": "550e8400-e29b-41d4-a716-446655440001",
      "canonicalName": "chicken breast",
      "canonicalUnit": "gram",
      "canonicalUnitFamily": "weight",
      "normalizationStatus": "manually_confirmed"
    }
  }
}
```

**Side effects when `saveAsAlias: true`:**
- Creates alias "chkn breast" -> canonical ingredient "chicken breast"
- Future recipes with "chkn breast" will auto-resolve to `alias_matched`
- If alias already exists (pointing to same canonical), no-op
- If alias conflicts (pointing to different canonical), alias creation is skipped (no error — the correction still succeeds)

**Error 404:** Recipe or recipe ingredient not found.
**Error 400:** `canonicalIngredientId` is not a valid UUID or does not exist.

### Re-Normalize Recipe Ingredients

```
POST /api/v1/recipes/:id/renormalize
```

No request body required. Re-runs the normalization pipeline for all ingredients on this recipe that have `unresolved` status. Ingredients with `manually_confirmed` status are preserved.

**Response 200:**
```json
{
  "data": {
    "type": "recipe-normalizations",
    "id": "uuid",
    "attributes": {
      "recipeId": "880e8400-e29b-41d4-a716-446655440001",
      "totalIngredients": 8,
      "resolved": 6,
      "unresolved": 2,
      "changed": 3,
      "ingredients": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440001",
          "rawName": "pecorino romano",
          "normalizationStatus": "matched",
          "canonicalName": "pecorino romano",
          "previousStatus": "unresolved"
        }
      ]
    }
  }
}
```

The `ingredients` array only includes ingredients whose status changed during re-normalization.

**Error 404:** Recipe not found.

---

## Ingredient Usage (Recipes Referencing a Canonical Ingredient)

### List Recipes by Canonical Ingredient

```
GET /api/v1/ingredients/:id/recipes?page[number]=1&page[size]=20
```

Returns recipes that have at least one `recipe_ingredient` referencing this canonical ingredient.

**Response 200:**
```json
{
  "data": [
    {
      "type": "recipes",
      "id": "880e8400-e29b-41d4-a716-446655440001",
      "attributes": {
        "title": "Pasta Carbonara",
        "tags": ["italian", "pasta"],
        "createdAt": "2026-03-27T12:00:00Z"
      }
    }
  ],
  "meta": {
    "total": 12,
    "page": 1,
    "pageSize": 20
  }
}
```

**Error 404:** Canonical ingredient not found.

---

## Modified Recipe Endpoints

### Recipe Detail (Extended Response)

```
GET /api/v1/recipes/:id
```

**Response 200:** Existing fields plus:
```json
{
  "data": {
    "type": "recipes",
    "id": "uuid",
    "attributes": {
      "title": "Pasta Carbonara",
      "description": "Classic Roman pasta dish",
      "servings": 4,
      "prepTimeMinutes": 10,
      "cookTimeMinutes": 20,
      "sourceUrl": "https://example.com/carbonara",
      "tags": ["italian", "pasta"],
      "source": "...",
      "ingredients": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440001",
          "rawName": "spaghetti",
          "rawQuantity": "400",
          "rawUnit": "g",
          "position": 1,
          "canonicalIngredientId": "550e8400-e29b-41d4-a716-446655440010",
          "canonicalName": "spaghetti",
          "canonicalUnit": "gram",
          "canonicalUnitFamily": "weight",
          "normalizationStatus": "matched"
        }
      ],
      "steps": [],
      "plannerConfig": {
        "classification": "dinner",
        "servingsYield": 4,
        "eatWithinDays": 3,
        "minGapDays": 7,
        "maxConsecutiveDays": 1
      },
      "plannerReady": true,
      "plannerIssues": [],
      "createdAt": "2026-03-27T12:00:00Z",
      "updatedAt": "2026-03-27T12:00:00Z"
    }
  }
}
```

When planner config is absent: `plannerConfig` is `null`, `plannerReady` is `false`, `plannerIssues` includes `"missing planner configuration"`.

When classification is missing: `plannerReady` is `false`, `plannerIssues` includes `"classification is required"`.

### Recipe List (Extended Response)

```
GET /api/v1/recipes
```

Each list item gains:
```json
{
  "attributes": {
    "plannerReady": false,
    "classification": "dinner",
    "resolvedIngredients": 5,
    "totalIngredients": 8
  }
}
```

`resolvedIngredients` counts recipe ingredients with status `matched`, `alias_matched`, or `manually_confirmed`. `totalIngredients` is the total count. These are computed via a subquery or join, not stored.

**New filter query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `plannerReady` | string | `true` or `false` — filter by planner readiness |
| `classification` | string | Filter by planner classification (e.g., `dinner`, `breakfast`) |
| `normalizationStatus` | string | `complete` (all ingredients resolved) or `incomplete` (has unresolved) |

### Recipe Create/Update (Extended Request)

```
POST /api/v1/recipes
PATCH /api/v1/recipes/:id
```

**Extended request body:**
```json
{
  "data": {
    "type": "recipes",
    "attributes": {
      "title": "Pasta Carbonara",
      "source": "...",
      "plannerConfig": {
        "classification": "dinner",
        "servingsYield": 4,
        "eatWithinDays": 3,
        "minGapDays": 7,
        "maxConsecutiveDays": 1
      }
    }
  }
}
```

On create: if `plannerConfig` is provided, creates the planner config record.
On update: if `plannerConfig` is provided, upserts the planner config record. If `plannerConfig` is explicitly `null`, deletes the planner config.

**Response:** Same as recipe detail (includes normalized ingredients, planner config, planner readiness).

### Parse Preview (Extended Response)

```
POST /api/v1/recipes/parse
```

The existing parse preview endpoint now includes normalization status for each ingredient. This is computed on-the-fly against the current canonical ingredient registry but not persisted.

**Extended response:**
```json
{
  "data": {
    "type": "recipe-parse",
    "id": "parse",
    "attributes": {
      "ingredients": [
        { "name": "spaghetti", "quantity": "400", "unit": "g" },
        { "name": "pecorino romano", "quantity": "100", "unit": "g" }
      ],
      "steps": [],
      "metadata": {},
      "errors": [],
      "normalization": [
        {
          "rawName": "spaghetti",
          "normalizationStatus": "matched",
          "canonicalName": "spaghetti",
          "canonicalUnit": "gram",
          "canonicalUnitFamily": "weight"
        },
        {
          "rawName": "pecorino romano",
          "normalizationStatus": "unresolved",
          "canonicalName": null,
          "canonicalUnit": "gram",
          "canonicalUnitFamily": "weight"
        }
      ]
    }
  }
}
```

The `normalization` array is parallel to `ingredients` — same length, same order. Unit normalization is always applied (from the static registry). Ingredient name normalization depends on the canonical registry state at the time of the request.
