# Recipe Management — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-25
---

## 1. Overview

Recipe management is the foundational feature for household meal planning. It enables household members to store, view, edit, and search recipes using the Cooklang format — a human-readable plain-text markup for recipes that encodes ingredients, quantities, cookware, and timers inline with step instructions.

The system stores recipes as raw Cooklang text and parses them server-side into structured data (ingredients list, step-by-step instructions). The API serves both the raw Cooklang source and parsed structured representation, allowing the frontend to render rich recipe views while preserving the original authored format.

This task introduces a new `recipe-service` microservice and corresponding frontend views for recipe list browsing (with search/filter), recipe detail viewing, and recipe creation/editing via a Cooklang text editor with validation.

## 2. Goals

Primary goals:
- Household members can create recipes using Cooklang syntax
- Household members can view a list of all recipes in their household
- Household members can view a single recipe with structured ingredient list and step-by-step instructions
- Household members can edit existing recipes
- Household members can search and filter recipes by name and tags
- Cooklang input is validated server-side before persistence
- Raw Cooklang source is preserved alongside parsed data

Non-goals:
- Meal plan generation or scheduling
- Recipe sharing across households
- Image upload or storage
- Nutritional information calculation
- Shopping list generation
- Recipe ratings or favorites
- Import from external recipe sources or URLs
- Recipe scaling (adjusting servings/quantities)

## 3. User Stories

- As a household member, I want to create a recipe by writing Cooklang text so that I can store structured recipes for my household
- As a household member, I want to see a list of all recipes in my household so that I can browse what's available
- As a household member, I want to search recipes by name so that I can quickly find a specific recipe
- As a household member, I want to filter recipes by tag so that I can find recipes by cuisine or category
- As a household member, I want to view a recipe's ingredient list and step-by-step instructions so that I can cook the meal
- As a household member, I want to edit an existing recipe so that I can correct mistakes or improve it
- As a household member, I want to delete a recipe that is no longer needed
- As a household member, I want validation feedback when my Cooklang syntax has errors so that I can fix them before saving

## 4. Functional Requirements

### 4.1 Recipe Storage

- Recipes are scoped to a tenant and household (tenant_id + household_id)
- Each recipe stores: title, description, raw Cooklang source, servings, prep time (minutes), cook time (minutes), source URL, and tags
- The raw Cooklang source is the authoritative representation — parsed data is derived from it
- Soft delete with restore window (consistent with productivity-service pattern)

### 4.2 Cooklang Parsing

- The server parses raw Cooklang text into structured data on read
- Parsed output includes:
  - Ordered list of ingredients with name, quantity, and unit
  - Ordered list of steps, each containing text segments and inline ingredient/cookware/timer references
- Parsing follows the Cooklang spec (https://cooklang.org/docs/spec/) for: ingredients (`@`), cookware (`#`), timers (`~`), comments (`--` and `[- -]`), and metadata
- Validation: the server validates Cooklang syntax on create/update and returns structured errors if the input is malformed
- Deduplication: ingredients appearing multiple times across steps are aggregated in the ingredient list with combined quantities where units match

### 4.3 Recipe CRUD

- Create: accepts title, description, Cooklang source, metadata fields, and tags; validates Cooklang; persists recipe
- Read (list): returns paginated recipe list with title, description, tags, prep/cook time; supports search by title and filter by tag
- Read (detail): returns full recipe including raw Cooklang source and parsed structured data (ingredients + steps)
- Update: accepts partial or full updates; re-validates Cooklang if source is changed
- Delete: soft delete with restore capability

### 4.4 Search and Filter

- Text search on recipe title (case-insensitive, substring match)
- Filter by one or more tags (AND semantics — recipe must have all specified tags)
- Search and filter are combinable
- Results are paginated

### 4.5 Tags

- Tags are free-form strings associated with a recipe (e.g., "italian", "quick", "vegetarian")
- A recipe can have zero or more tags
- Tags are stored normalized (lowercase, trimmed)
- An endpoint to list all tags in use within a household (for filter UI)

### 4.6 Frontend — Recipe List

- Displays a card-based list of recipes for the current household
- Each card shows: title, description (truncated), tags, prep + cook time
- Search bar for filtering by title
- Tag filter (multi-select from available tags)
- Tap a card to navigate to recipe detail
- FAB or button to create a new recipe

### 4.7 Frontend — Recipe Detail

- Displays recipe title, description, metadata (servings, prep time, cook time, source URL)
- Ingredient list section — each ingredient with quantity, unit, and name
- Step-by-step section — ordered steps with inline references rendered distinctly (e.g., ingredients highlighted)
- Edit button to navigate to edit form
- Delete button with confirmation

### 4.8 Frontend — Recipe Create/Edit

- Form with fields: title, description, servings, prep time, cook time, source URL, tags
- Cooklang text editor (multi-line textarea) for the recipe body
- **Live preview panel** alongside the editor that shows the parsed result as the user types:
  - Renders the ingredient list and step-by-step instructions in real time
  - Highlights ingredients, cookware, and timers with distinct styling (matching the detail view)
  - Updates on each keystroke (debounced, ~300ms) by calling a server-side parse/preview endpoint
  - Shows inline syntax errors if the Cooklang is malformed (e.g., unclosed blocks)
- Side-by-side layout on desktop (editor left, preview right); stacked on mobile (editor top, preview below)
- Brief syntax reference or cheat sheet accessible from the editor (e.g., collapsible help section or tooltip)
- On submit, server validates Cooklang and returns errors if invalid (server validation is authoritative)
- Validation errors displayed inline near the Cooklang editor
- On success, navigates to recipe detail view

## 5. API Surface

Base path: `/api/v1/recipes`

### 5.1 Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/recipes/parse` | Parse Cooklang source (preview endpoint) |
| POST | `/api/v1/recipes` | Create a recipe |
| GET | `/api/v1/recipes` | List recipes (paginated, searchable, filterable) |
| GET | `/api/v1/recipes/:id` | Get recipe detail with parsed Cooklang |
| PATCH | `/api/v1/recipes/:id` | Update a recipe |
| DELETE | `/api/v1/recipes/:id` | Soft delete a recipe |
| POST | `/api/v1/recipes/restorations` | Restore a soft-deleted recipe |
| GET | `/api/v1/recipes/tags` | List all tags in use for the household |

### 5.2 Query Parameters (List)

| Parameter | Type | Description |
|-----------|------|-------------|
| `search` | string | Case-insensitive substring match on title |
| `tag` | string (repeatable) | Filter by tag(s), AND semantics |
| `page[number]` | int | Page number (default 1) |
| `page[size]` | int | Page size (default 20, max 100) |

### 5.3 Resource Representations

**Recipe (list item):**
```json
{
  "type": "recipes",
  "id": "uuid",
  "attributes": {
    "title": "Pasta Carbonara",
    "description": "Classic Roman pasta dish",
    "servings": 4,
    "prep-time-minutes": 10,
    "cook-time-minutes": 20,
    "tags": ["italian", "pasta"],
    "created-at": "2026-03-25T12:00:00Z",
    "updated-at": "2026-03-25T12:00:00Z"
  }
}
```

**Recipe (detail):**
```json
{
  "type": "recipes",
  "id": "uuid",
  "attributes": {
    "title": "Pasta Carbonara",
    "description": "Classic Roman pasta dish",
    "servings": 4,
    "prep-time-minutes": 10,
    "cook-time-minutes": 20,
    "source-url": "https://example.com/carbonara",
    "tags": ["italian", "pasta"],
    "source": "Boil @water{2%l} in a #large pot{}...",
    "ingredients": [
      { "name": "water", "quantity": "2", "unit": "l" },
      { "name": "spaghetti", "quantity": "400", "unit": "g" },
      { "name": "guanciale", "quantity": "200", "unit": "g" }
    ],
    "steps": [
      {
        "number": 1,
        "segments": [
          { "type": "text", "value": "Boil " },
          { "type": "ingredient", "name": "water", "quantity": "2", "unit": "l" },
          { "type": "text", "value": " in a " },
          { "type": "cookware", "name": "large pot" },
          { "type": "text", "value": "..." }
        ]
      }
    ],
    "created-at": "2026-03-25T12:00:00Z",
    "updated-at": "2026-03-25T12:00:00Z"
  }
}
```

**Create/Update request body:**
```json
{
  "data": {
    "type": "recipes",
    "attributes": {
      "title": "Pasta Carbonara",
      "description": "Classic Roman pasta dish",
      "servings": 4,
      "prep-time-minutes": 10,
      "cook-time-minutes": 20,
      "source-url": "https://example.com/carbonara",
      "tags": ["italian", "pasta"],
      "source": "Boil @water{2%l} in a #large pot{}..."
    }
  }
}
```

### 5.4 Error Responses

Cooklang validation errors:
```json
{
  "errors": [
    {
      "status": "422",
      "title": "Invalid Cooklang syntax",
      "detail": "Unclosed ingredient block at line 3, column 15",
      "source": { "pointer": "/data/attributes/source" }
    }
  ]
}
```

## 6. Data Model

### 6.1 Recipes Table (`recipe.recipes`)

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, generated |
| tenant_id | UUID | NOT NULL, indexed |
| household_id | UUID | NOT NULL, indexed |
| title | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| source | TEXT | NOT NULL (raw Cooklang) |
| servings | INT | nullable |
| prep_time_minutes | INT | nullable |
| cook_time_minutes | INT | nullable |
| source_url | VARCHAR(2048) | nullable |
| deleted_at | TIMESTAMP | nullable (soft delete) |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- `(tenant_id, household_id)` — compound index for list queries
- `(tenant_id, household_id, title)` — for search
- `(tenant_id, household_id, deleted_at)` — for filtering soft-deleted

### 6.2 Recipe Tags Table (`recipe.recipe_tags`)

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, generated |
| recipe_id | UUID | NOT NULL, FK → recipes.id, indexed |
| tag | VARCHAR(100) | NOT NULL, normalized lowercase |

Indexes:
- `(recipe_id, tag)` — unique compound index
- `(tag)` — for tag listing/filtering

### 6.3 Recipe Restorations Table (`recipe.recipe_restorations`)

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, generated |
| recipe_id | UUID | NOT NULL, FK → recipes.id |
| restored_at | TIMESTAMP | NOT NULL |

## 7. Service Impact

### 7.1 recipe-service (new)

New microservice following the standard service pattern:

```
services/recipe-service/
  cmd/
    main.go
  internal/
    config/
    recipe/
      model.go
      entity.go
      builder.go
      processor.go
      provider.go
      resource.go
      rest.go
      cooklang/        # Cooklang parser package
        parser.go
        parser_test.go
  docs/
    domain.md
    rest.md
    storage.md
  Dockerfile
  go.mod
```

- Own schema: `recipe.*`
- Cooklang parser as an internal package under the recipe domain
- Standard auth middleware (JWT validation via JWKS)
- Standard tenant/household scoping

### 7.2 frontend

- New navigation entry for "Recipes"
- New pages: recipe list, recipe detail, recipe create/edit
- New API client hooks for recipe endpoints
- Tag filter component
- Live preview calls a server-side parse endpoint (debounced) — single Cooklang parser implementation in Go; no client-side parser needed

### 7.3 Infrastructure

- `deploy/compose/docker-compose.yml` — add recipe-service container
- `deploy/k8s/` — add recipe-service deployment, service, and ingress rule
- `nginx.conf` — add `/api/v1/recipes` proxy route
- `go.work` — add recipe-service module
- `.github/workflows/` — add CI detection for recipe-service
- `scripts/` — add `build-recipe.sh`
- `bruno/` — add recipe collection

## 8. Non-Functional Requirements

### 8.1 Performance
- Recipe list endpoint should respond within 200ms for up to 1000 recipes per household
- Cooklang parsing should add no more than 10ms overhead per recipe on detail reads

### 8.2 Security
- All endpoints require valid JWT
- All queries scoped by tenant_id and household_id from JWT claims
- Cooklang source is treated as untrusted input — no server-side execution, HTML-escaped on output

### 8.3 Observability
- Standard structured logging with request_id, user_id, tenant_id, household_id
- OpenTelemetry tracing on all endpoints

### 8.4 Multi-Tenancy
- All recipe data scoped by tenant_id + household_id
- GORM tenant callbacks enforce scoping (consistent with other services)

## 9. Open Questions

All resolved:

1. ~~**Cooklang spec completeness**~~ — v1 supports core syntax (@ingredients, #cookware, ~timers, comments). Metadata blocks and edge cases deferred.
2. ~~**Restore window**~~ — Same duration as productivity-service.
3. ~~**Tag management**~~ — Per-recipe tag editing only in v1. No rename/merge across recipes.

## 10. Acceptance Criteria

- [ ] New `recipe-service` builds, passes lint, and passes tests
- [ ] Recipe CRUD endpoints functional with proper JSON:API format
- [ ] Cooklang source is validated on create and update — invalid syntax returns 422 with descriptive errors
- [ ] Recipe detail endpoint returns both raw Cooklang source and parsed structured data (ingredients + steps)
- [ ] Recipe list supports pagination, title search, and tag filtering
- [ ] Tags endpoint returns all tags in use for the current household
- [ ] Soft delete and restore functional
- [ ] All endpoints enforce tenant_id + household_id scoping
- [ ] Frontend recipe list page displays cards with title, description, tags, and times
- [ ] Frontend recipe list supports search and tag filter
- [ ] Frontend recipe detail page shows ingredient list and step-by-step instructions with highlighted inline references
- [ ] Frontend recipe create/edit form includes Cooklang text editor with live preview panel
- [ ] Live preview renders ingredient list and steps in real time as the user types
- [ ] Live preview shows inline syntax errors for malformed Cooklang
- [ ] Service integrates into docker-compose, nginx routing, and CI
- [ ] Service documentation (domain.md, rest.md, storage.md) is complete
- [ ] Bruno collection for recipe endpoints exists
