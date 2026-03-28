# Ingredient Normalization — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-27
---

## 1. Overview

Ingredient normalization extends the existing recipe-service (task-005) to transform raw Cooklang-parsed ingredients into a canonical internal representation. When a recipe is saved, its parsed ingredients are persisted with both raw values (exactly as authored) and normalized values (matched against a tenant-scoped canonical ingredient registry). Unresolved or ambiguous matches surface in the admin UI for manual correction, and accepted corrections feed back as aliases for future automatic matching.

This is the foundational layer for downstream meal planning and shopping-list generation. Without normalized ingredients, the system cannot aggregate "chicken breast" and "chicken breasts" across recipes into a single shopping-list line. This task builds the ingredient knowledge base that those features depend on.

The recipe model is also extended with planner-specific metadata (classification, eat-within window, gap days) stored as a separate planner configuration entity, and a "planner readiness" status that indicates whether a recipe has sufficient metadata and normalization for downstream use.

## 2. Goals

Primary goals:
- Persist parsed ingredients per recipe with raw values preserved (name, quantity, unit, position)
- Build a tenant-scoped canonical ingredient registry with alias support
- Automatically normalize ingredients on recipe create/update via exact match, alias match, and text normalization
- Record normalization status per ingredient (matched, alias_matched, unresolved, manually_confirmed)
- Normalize units into canonical representations within unit families (count, weight, volume)
- Enable manual review and correction of unresolved ingredients in the admin UI
- Learn from manual corrections by persisting alias mappings for future normalization
- Extend recipes with planner metadata (classification, eat_within_days, min_gap_days, max_consecutive_days)
- Compute planner-readiness status per recipe
- Emit audit events for recipe and normalization lifecycle actions
- Provide API endpoints for all ingredient management, normalization correction, and planner configuration

Non-goals:
- Shopping-list generation (depends on this, built separately)
- Meal plan scheduling/generation
- Ingredient-specific unit conversions (e.g., cups of flour to grams)
- Cross-tenant ingredient sharing or seeded defaults
- Nutritional information
- Centralized audit log service (events stored locally in recipe-service)
- Recipe scaling (adjusting quantities for different serving counts)

## 3. User Stories

- As a household member, I want my recipe's ingredients automatically matched to canonical ingredients so that downstream features (shopping lists, meal plans) can aggregate them correctly
- As a household member, I want to see which ingredients in my recipe are unresolved so I can fix them before using the recipe in a meal plan
- As a household member, I want to accept, reject, or reassign a suggested canonical ingredient match so the system normalizes correctly
- As a household member, I want my manual corrections saved as aliases so the system learns and doesn't ask me the same question twice
- As a household member, I want to browse and manage the canonical ingredient registry for my tenant so I can add missing ingredients or clean up duplicates
- As a household member, I want to attach planner metadata (classification, eat-within days, gap days) to a recipe so the meal planner can schedule it appropriately
- As a household member, I want to see at a glance whether a recipe is planner-ready so I know what still needs attention
- As a household member, I want the original Cooklang text preserved exactly as I wrote it, even after normalization runs
- As a household member, I want to re-normalize a recipe's ingredients without editing the source so that newly added canonical ingredients and aliases take effect on older recipes
- As a household member, I want to see normalization status in the live preview before saving so I know which ingredients will resolve
- As a household member, I want to filter my recipe list by planner readiness, classification, and normalization status so I can find recipes that need attention
- As a household member, I want to see which recipes use a canonical ingredient so I can understand the impact of changes

## 4. Functional Requirements

### 4.1 Recipe Ingredient Persistence

- On recipe create, the system parses the Cooklang source and persists each ingredient as a `recipe_ingredient` record
- Each record preserves raw parsed values: name, quantity, unit, and ordinal position in the recipe
- The raw Cooklang source is never modified by normalization

**Reconciliation on source update:** When a recipe's Cooklang source is updated, the system reconciles new parsed ingredients against existing `recipe_ingredient` records rather than blindly deleting and re-creating:

1. Parse the updated source to get the new ingredient list
2. For each new parsed ingredient, attempt to match an existing `recipe_ingredient` by `raw_name` (lowercased, trimmed)
3. If a match is found and the existing record is `manually_confirmed`, carry forward the `canonical_ingredient_id`, `normalization_status`, and `canonical_unit` to the new record
4. If a match is found and the existing record has any other status (`matched`, `alias_matched`, `unresolved`), re-run normalization (the canonical registry may have changed since last save)
5. If no match is found (new ingredient), run normalization from scratch
6. Remove any existing `recipe_ingredient` records whose `raw_name` no longer appears in the parsed source (ingredient was removed from Cooklang)
7. Update `position` values to reflect the new ordering from the parse

This ensures that manual confirmation work survives source edits. A user who fixes a typo in step text does not lose their ingredient normalization work.

### 4.2 Canonical Ingredient Registry

- A tenant-scoped registry of canonical ingredients, each with a unique name, optional display name, and unit family hint (count, weight, volume, or unspecified)
- Each canonical ingredient can have zero or more aliases (alternate names that map to it)
- Aliases are tenant-scoped and globally unique within a tenant (an alias cannot point to two different canonical ingredients)
- CRUD operations on canonical ingredients and their aliases
- Canonical ingredient names are stored normalized (lowercase, trimmed)

### 4.3 Re-Normalization

- The recipe detail page includes a "Re-normalize" action that re-runs the normalization pipeline for that recipe's existing ingredients without requiring a source edit
- This allows newly added canonical ingredients and aliases to take effect on older recipes
- Re-normalization preserves `manually_confirmed` statuses — only `unresolved` ingredients are re-evaluated
- Re-normalization emits a `recipe.renormalized` audit event with metadata recording how many ingredients changed status

### 4.4 Automatic Normalization Pipeline

On recipe create/update, after parsing, for each parsed ingredient the system attempts normalization in order:

1. **Exact canonical match** — raw ingredient name (lowercased, trimmed) matches a canonical ingredient name exactly. Status: `matched`
2. **Alias match** — raw name matches a known alias. Status: `alias_matched`
3. **Text normalization match** — apply basic text normalization (strip trailing 's' for plurals, collapse whitespace, strip leading/trailing articles) and re-attempt exact + alias match. Status: `matched` or `alias_matched`
4. **Unresolved** — no confident match found. Status: `unresolved`

Normalization is synchronous within the create/update request. The normalized canonical ingredient ID (if matched) and normalization status are stored on the recipe ingredient record.

### 4.5 Unit Normalization

- A static unit registry mapping common unit strings to canonical unit identities within families:
  - **Count**: each, piece, pcs, count, whole, clove, cloves, head, heads, bunch, bunches, sprig, sprigs, stalk, stalks, slice, slices, pinch, pinches, dash, dashes
  - **Weight**: g, gram, grams, kg, kilogram, kilograms, oz, ounce, ounces, lb, pound, pounds
  - **Volume**: ml, milliliter, milliliters, l, liter, liters, tsp, teaspoon, teaspoons, tbsp, tablespoon, tablespoons, cup, cups, fl oz, fluid ounce, fluid ounces
- Each recipe ingredient stores both the raw unit text and the resolved canonical unit identity
- Phase 1 only performs same-unit aggregation (already handled by the Cooklang parser). No cross-unit conversions
- Units that don't match any known mapping are stored with a null canonical unit (unresolved unit)

### 4.6 Manual Review and Correction

- The admin UI presents unresolved or ambiguous ingredient matches on the recipe detail/edit page
- For each unresolved ingredient, the user can:
  - Accept a suggested canonical ingredient (if the system offers one)
  - Search for and select a different canonical ingredient
  - Leave the ingredient unresolved
  - Create a new canonical ingredient inline and assign it
- Accepted corrections update the recipe ingredient's canonical reference and status to `manually_confirmed`
- Corrections are persisted immediately via API

**Suggestion ranking:** When the resolve dropdown opens for an unresolved ingredient, the system suggests canonical ingredients using ILIKE prefix matching on canonical names and alias names, ordered by usage count descending (most-used ingredients surface first). The user can type to refine the search. This is simple, requires no PostgreSQL extensions, and produces good results because popular ingredients are most likely correct.

### 4.7 Alias Learning

- When a user manually maps a raw ingredient name to a canonical ingredient, the system offers to save that mapping as an alias
- Saving an alias causes future recipes with the same raw ingredient name to auto-resolve
- Aliases can be managed (listed, deleted) through the canonical ingredient API
- An alias that conflicts with an existing canonical ingredient name is rejected

### 4.8 Recipe Planner Configuration

- A separate `recipe_planner_config` entity associated with a recipe, containing:
  - `servings_yield` (int, nullable) — overrides recipe servings for planning purposes
  - `classification` (string, nullable) — e.g., "breakfast", "lunch", "dinner", "snack", "side"
  - `eat_within_days` (int, nullable) — how many days leftovers last
  - `min_gap_days` (int, nullable) — minimum days before repeating this recipe
  - `max_consecutive_days` (int, nullable) — max days this recipe can appear consecutively
- Planner config is optional — a recipe without planner config is not planner-ready
- Planner config is created/updated via the recipe API (nested under recipe endpoints, not a separate resource)
- Planner config inherits tenant_id + household_id scoping from the parent recipe

### 4.9 Planner Readiness

A recipe is considered planner-ready when all of the following are true:
- Cooklang source parses successfully (already enforced on save)
- Planner config exists with at least `classification` set
- `servings` (from recipe or planner config `servings_yield`) is present

Unresolved ingredients do **not** block planner readiness. They block shopping-list generation only (a downstream concern).

The planner-readiness status is computed on read, not stored. The recipe detail API includes a `plannerReady` boolean and a `plannerIssues` array describing what's missing.

### 4.10 Audit Events

The system records audit events in a local `recipe_audit_events` table:

| Event | Description |
|-------|-------------|
| `recipe.created` | Recipe created |
| `recipe.updated` | Recipe updated |
| `recipe.deleted` | Recipe soft-deleted |
| `recipe.restored` | Recipe restored |
| `recipe.renormalized` | Recipe ingredients re-normalized |
| `normalization.corrected` | Ingredient normalization manually changed |
| `ingredient.alias_created` | Canonical ingredient alias created from manual correction |

Each event records: id, tenant_id, entity_type, entity_id, action, actor_id (from JWT), metadata (JSON), created_at.

Audit events are write-only in this task. No read API for audit logs is in scope (future task).

### 4.11 Frontend — Ingredient Normalization Panel

The recipe detail page and recipe form page are extended with an ingredient normalization panel:

- **Recipe Detail Page**: Below the existing ingredients list, show normalization status per ingredient:
  - Green check for `matched` / `alias_matched` / `manually_confirmed` with canonical name shown
  - Yellow warning for `unresolved` with a "Resolve" action
  - Summary badge: "5/8 ingredients resolved"
  - "Re-normalize" button that re-runs the normalization pipeline for unresolved ingredients (calls `/renormalize` endpoint). Useful after adding new canonical ingredients or aliases.
- **Recipe Form Page (Edit)**: After the live preview panel, show a normalization review section:
  - Lists ingredients with their normalization status
  - Unresolved ingredients show a search/select dropdown to pick a canonical ingredient
  - Option to create a new canonical ingredient inline
  - Checkbox to "Save as alias" when resolving (checked by default)
  - Changes submit independently from the recipe save (ingredient corrections are their own API calls)
- **Live Preview Normalization**: The existing Cooklang live preview (which calls `/recipes/parse`) now also shows normalization status per ingredient. This gives the user immediate feedback on which ingredients will auto-resolve before saving. Normalization status in preview is informational only — not persisted until save.
- **Responsive**: On mobile, normalization panel appears below the preview. On desktop, integrated into the right panel.

### 4.12 Frontend — Canonical Ingredient Management

A new admin page for managing the canonical ingredient registry:

- **Ingredient List**: Searchable, paginated list of canonical ingredients for the tenant
  - Shows name, display name, unit family, alias count, usage count (how many recipe ingredients reference it)
  - Search by name or alias
- **Ingredient Detail/Edit**: View and edit a canonical ingredient
  - Edit display name, unit family
  - Manage aliases (add, remove)
  - View which recipes reference this ingredient — clickable list linking to recipe detail pages (paginated via `/ingredients/:id/recipes`)
- **Create Ingredient**: Simple form with name, display name, unit family
- **Reassign & Delete**: When deleting a canonical ingredient that has references, offer to reassign all references to a different canonical ingredient before deletion (calls `/ingredients/:id/reassign`). This avoids forcing the user to manually resolve each recipe ingredient individually.
- **Empty State**: When the canonical ingredient registry is empty, show brief guidance explaining that ingredients are added automatically as recipes are created, and can also be added manually. No complex onboarding — just inline help text.
- Accessible from the recipe section navigation (e.g., sidebar sub-item under Recipes)

### 4.13 Frontend — Planner Configuration

The recipe form page is extended with a collapsible "Planner Settings" section:

- Fields: classification (dropdown), eat-within days (number), min gap days (number), max consecutive days (number), servings yield override (number)
- Shown on both create and edit
- Planner readiness badge on recipe detail page and recipe list cards

## 5. API Surface

### 5.1 Modified Endpoints

| Method | Path | Changes |
|--------|------|---------|
| POST | `/api/v1/recipes` | Now triggers ingredient normalization; response includes `ingredients` with normalization data, `plannerReady`, `plannerIssues` |
| GET | `/api/v1/recipes/:id` | Response includes `ingredients` with normalization data, `plannerConfig`, `plannerReady`, `plannerIssues` |
| PATCH | `/api/v1/recipes/:id` | Re-triggers normalization if source changed; accepts `plannerConfig` in request body |
| GET | `/api/v1/recipes` | List response includes `plannerReady`, `classification`, `resolvedIngredients`, and `totalIngredients` per recipe; supports new filter query params |
| POST | `/api/v1/recipes/parse` | Response now includes `normalization` array with per-ingredient match status (preview only, not persisted) |

### 5.2 New Endpoints — Ingredient Normalization

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/recipes/:id/ingredients/:ingredientId/resolve` | Submit normalization correction for a recipe ingredient |
| POST | `/api/v1/recipes/:id/renormalize` | Re-run normalization pipeline for all unresolved ingredients on a recipe |

**Resolve Request:**
```json
{
  "data": {
    "type": "ingredient-resolutions",
    "attributes": {
      "canonicalIngredientId": "uuid",
      "saveAsAlias": true
    }
  }
}
```

**Resolve Response:** Updated recipe ingredient with new normalization status.

### 5.3 New Endpoints — Canonical Ingredients

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ingredients` | List canonical ingredients (paginated, searchable) |
| POST | `/api/v1/ingredients` | Create canonical ingredient |
| GET | `/api/v1/ingredients/:id` | Get canonical ingredient with aliases |
| PATCH | `/api/v1/ingredients/:id` | Update canonical ingredient |
| DELETE | `/api/v1/ingredients/:id` | Delete canonical ingredient (fails if referenced) |
| POST | `/api/v1/ingredients/:id/aliases` | Add alias to canonical ingredient |
| DELETE | `/api/v1/ingredients/:id/aliases/:aliasId` | Remove alias |
| GET | `/api/v1/ingredients/:id/recipes` | List recipes that reference this canonical ingredient (paginated) |
| POST | `/api/v1/ingredients/:id/reassign` | Reassign all recipe ingredient references to a different canonical ingredient, then delete this one |

**Canonical Ingredient Resource:**
```json
{
  "type": "ingredients",
  "id": "uuid",
  "attributes": {
    "name": "chicken breast",
    "displayName": "Chicken Breast",
    "unitFamily": "weight",
    "aliasCount": 3,
    "usageCount": 12,
    "createdAt": "2026-03-27T12:00:00Z",
    "updatedAt": "2026-03-27T12:00:00Z"
  }
}
```

**Canonical Ingredient Detail (with aliases):**
```json
{
  "type": "ingredients",
  "id": "uuid",
  "attributes": {
    "name": "chicken breast",
    "displayName": "Chicken Breast",
    "unitFamily": "weight",
    "aliases": [
      { "id": "uuid", "name": "chicken breasts" },
      { "id": "uuid", "name": "boneless chicken breast" }
    ],
    "createdAt": "2026-03-27T12:00:00Z",
    "updatedAt": "2026-03-27T12:00:00Z"
  }
}
```

**Create/Update Request:**
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

**Add Alias Request:**
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

### 5.4 New Endpoints — Planner Configuration

Planner config is managed as a nested attribute on recipe create/update, not as a separate resource.

**Extended Recipe Create/Update Request:**
```json
{
  "data": {
    "type": "recipes",
    "attributes": {
      "title": "Pasta Carbonara",
      "source": "...",
      "plannerConfig": {
        "classification": "dinner",
        "eatWithinDays": 3,
        "minGapDays": 7,
        "maxConsecutiveDays": 1,
        "servingsYield": 4
      }
    }
  }
}
```

**Extended Recipe Detail Response:**
```json
{
  "type": "recipes",
  "id": "uuid",
  "attributes": {
    "title": "Pasta Carbonara",
    "ingredients": [
      {
        "id": "uuid",
        "rawName": "spaghetti",
        "rawQuantity": "400",
        "rawUnit": "g",
        "position": 1,
        "canonicalIngredientId": "uuid",
        "canonicalName": "spaghetti",
        "canonicalUnit": "gram",
        "canonicalUnitFamily": "weight",
        "normalizationStatus": "matched"
      },
      {
        "id": "uuid",
        "rawName": "pecorino romano",
        "rawQuantity": "100",
        "rawUnit": "g",
        "position": 2,
        "canonicalIngredientId": null,
        "canonicalName": null,
        "canonicalUnit": "gram",
        "canonicalUnitFamily": "weight",
        "normalizationStatus": "unresolved"
      }
    ],
    "plannerConfig": {
      "classification": "dinner",
      "eatWithinDays": 3,
      "minGapDays": 7,
      "maxConsecutiveDays": 1,
      "servingsYield": 4
    },
    "plannerReady": true,
    "plannerIssues": [],
    "source": "...",
    "steps": []
  }
}
```

### 5.5 Query Parameters

**Recipe List (new filters):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `plannerReady` | boolean | Filter by planner readiness (`true` or `false`) |
| `classification` | string | Filter by planner classification (e.g., `dinner`) |
| `normalizationStatus` | string | Filter recipes by ingredient normalization completeness: `complete` (all resolved), `incomplete` (has unresolved), or omit for no filter |

**Canonical Ingredients List:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `search` | string | Case-insensitive match on name or alias |
| `page[number]` | int | Page number (default 1) |
| `page[size]` | int | Page size (default 20, max 100) |

### 5.6 Error Cases

| Status | Condition |
|--------|-----------|
| 400 | Invalid canonical ingredient name (empty, too long) |
| 404 | Canonical ingredient or recipe ingredient not found |
| 409 | Alias conflicts with existing canonical ingredient name or existing alias |
| 409 | Delete canonical ingredient that is still referenced by recipe ingredients |
| 422 | Invalid unit family value |

## 6. Data Model

### 6.1 New Table: `recipe_ingredients`

Persisted parsed ingredients per recipe, with raw and normalized values.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, indexed |
| household_id | UUID | NOT NULL |
| recipe_id | UUID | NOT NULL, FK -> recipes.id ON DELETE CASCADE |
| raw_name | VARCHAR(255) | NOT NULL |
| raw_quantity | VARCHAR(100) | nullable |
| raw_unit | VARCHAR(100) | nullable |
| position | INT | NOT NULL |
| canonical_ingredient_id | UUID | nullable, FK -> canonical_ingredients.id ON DELETE SET NULL |
| canonical_unit | VARCHAR(50) | nullable |
| normalization_status | VARCHAR(30) | NOT NULL, default 'unresolved' |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- `(recipe_id)` — for recipe-scoped lookups
- `(tenant_id, household_id)` — for tenant scoping
- `(canonical_ingredient_id)` — for usage count queries

### 6.2 New Table: `canonical_ingredients`

Tenant-scoped ingredient registry.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| display_name | VARCHAR(255) | nullable |
| unit_family | VARCHAR(20) | nullable (count, weight, volume) |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- `(tenant_id, name)` — unique, for exact-match lookups
- `(tenant_id)` — for list queries

### 6.3 New Table: `canonical_ingredient_aliases`

Alternate names mapping to canonical ingredients.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| canonical_ingredient_id | UUID | NOT NULL, FK -> canonical_ingredients.id ON DELETE CASCADE |
| name | VARCHAR(255) | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |

Indexes:
- `(tenant_id, name)` — unique, for alias lookups
- `(canonical_ingredient_id)` — for listing aliases of an ingredient

### 6.4 New Table: `recipe_planner_configs`

Planner-specific metadata per recipe.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| recipe_id | UUID | NOT NULL, unique, FK -> recipes.id ON DELETE CASCADE |
| classification | VARCHAR(50) | nullable |
| servings_yield | INT | nullable |
| eat_within_days | INT | nullable |
| min_gap_days | INT | nullable |
| max_consecutive_days | INT | nullable |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- `(recipe_id)` — unique, one config per recipe

### 6.5 New Table: `recipe_audit_events`

Local audit log for recipe-service actions.

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL |
| entity_type | VARCHAR(50) | NOT NULL |
| entity_id | UUID | NOT NULL |
| action | VARCHAR(50) | NOT NULL |
| actor_id | UUID | NOT NULL |
| metadata | JSONB | nullable |
| created_at | TIMESTAMP | NOT NULL |

Indexes:
- `(tenant_id, entity_type, entity_id)` — for entity history queries
- `(tenant_id, action)` — for action-type queries
- `(created_at)` — for time-range queries

### 6.6 Modified: `recipes` table

No schema changes. Planner config is a separate table. The `ingredients` field in the API response is now sourced from `recipe_ingredients` rather than re-parsing Cooklang on every read.

## 7. Service Impact

### 7.1 recipe-service

**New domain: `internal/ingredient/`**
- `model.go` — Immutable CanonicalIngredient model with name, displayName, unitFamily, aliases
- `entity.go` — CanonicalIngredientEntity, CanonicalIngredientAliasEntity, migration
- `builder.go` — Builder with validation (name required, unit family must be valid enum)
- `processor.go` — CRUD, search, alias management
- `provider.go` — Database access (getByName, getByAlias, search)
- `resource.go` — Route registration for `/api/v1/ingredients` endpoints
- `rest.go` — JSON:API resource mappings

**New domain: `internal/normalization/`**
- `model.go` — RecipeIngredient model (raw + normalized values, status)
- `entity.go` — RecipeIngredientEntity, migration
- `builder.go` — Builder for RecipeIngredient
- `processor.go` — Normalization pipeline (exact match -> alias -> text normalization -> unresolved), manual correction, alias learning
- `provider.go` — Database access
- `resource.go` — Route registration for resolve endpoint
- `rest.go` — JSON:API mappings
- `unit_registry.go` — Static unit normalization map

**New domain: `internal/planner/`**
- `model.go` — PlannerConfig model
- `entity.go` — PlannerConfigEntity, migration
- `builder.go` — Builder with validation
- `processor.go` — Create/update/get planner config, readiness computation
- `provider.go` — Database access

**New domain: `internal/audit/`**
- `entity.go` — AuditEventEntity, migration
- `emitter.go` — Function to emit audit events (called from other processors)

**Modified: `internal/recipe/`**
- `processor.go` — After create/update, trigger normalization pipeline; include planner config in responses
- `resource.go` — Wire normalization and planner config into existing handlers
- `rest.go` — Extend RestDetailModel with normalized ingredients, planner config, planner readiness

### 7.2 frontend

**Modified pages:**
- `RecipeDetailPage.tsx` — Add normalization status display, planner readiness badge, planner config section
- `RecipeFormPage.tsx` — Add planner settings section, ingredient normalization review panel
- `RecipesPage.tsx` — Add planner-ready badge and classification on recipe cards; add filter controls for planner readiness, classification, and normalization completeness

**New components:**
- `components/features/recipes/ingredient-normalization-panel.tsx` — Normalization review and correction UI
- `components/features/recipes/ingredient-resolver.tsx` — Single ingredient resolution (search/select/create canonical)
- `components/features/recipes/planner-config-form.tsx` — Planner settings form fields
- `components/features/recipes/planner-ready-badge.tsx` — Status badge component

**New pages:**
- `pages/IngredientsPage.tsx` — Canonical ingredient list
- `pages/IngredientDetailPage.tsx` — Canonical ingredient detail with alias management

**New hooks:**
- `use-ingredients.ts` — CRUD hooks for canonical ingredients and aliases
- `use-ingredient-normalization.ts` — Hook for submitting normalization corrections

**Navigation:**
- Add "Ingredients" sub-item under Recipes section in sidebar

### 7.3 Infrastructure

No new services, containers, or routing changes. All new endpoints are under the existing recipe-service at `/api/v1/recipes` and `/api/v1/ingredients`.

## 8. Non-Functional Requirements

### 8.1 Performance
- Normalization pipeline should add no more than 50ms to recipe create/update for recipes with up to 50 ingredients
- Canonical ingredient search should respond within 100ms for registries up to 5,000 ingredients
- Unit registry is in-memory (static map), no database lookup required

### 8.2 Security
- All canonical ingredient data is scoped by tenant_id
- Recipe ingredient data inherits tenant_id + household_id scoping from the parent recipe
- Audit events record actor_id from JWT claims
- Canonical ingredient names are treated as untrusted input (length-limited, trimmed)

### 8.3 Observability
- Normalization pipeline logs: ingredient count, matched count, unresolved count per recipe save
- Audit events provide a persistent record of all state changes

### 8.4 Multi-Tenancy
- Canonical ingredients are tenant-scoped (shared across households within a tenant)
- Recipe ingredients are household-scoped (inherit from recipe)
- Aliases are tenant-scoped
- Audit events are tenant-scoped
- All queries enforce tenant scoping via GORM callbacks

## 9. Open Questions

All resolved:

1. ~~**Text normalization heuristics**~~ — Include naive pluralization (strip trailing 's') in phase 1. It catches the most common case (80/20 trade-off). Aliases cover the remaining edge cases.
2. ~~**Canonical ingredient seeding**~~ — Start empty. The registry grows organically through recipe entry, so it only contains ingredients the household actually uses.
3. ~~**Alias conflict resolution**~~ — Reject, don't merge. Merging is destructive and hard to undo. The user can manually reassign if needed.

## 10. Acceptance Criteria

### Backend
- [ ] `canonical_ingredients`, `canonical_ingredient_aliases`, `recipe_ingredients`, `recipe_planner_configs`, and `recipe_audit_events` tables are created via GORM AutoMigrate
- [ ] Canonical ingredient CRUD endpoints functional with JSON:API format
- [ ] Alias add/remove endpoints functional with conflict detection
- [ ] Recipe create/update triggers normalization pipeline and persists recipe ingredients
- [ ] Normalization pipeline attempts exact match, alias match, text normalization match in order
- [ ] Text normalization includes naive pluralization (strip trailing 's'), whitespace collapse, and article stripping
- [ ] Unit normalization resolves common unit strings to canonical identities
- [ ] Manual correction endpoint updates normalization status to `manually_confirmed`
- [ ] Alias learning: corrections with `saveAsAlias: true` create a new alias
- [ ] Re-normalize endpoint re-runs pipeline for unresolved ingredients, preserves manually_confirmed
- [ ] Parse preview endpoint includes normalization status per ingredient
- [ ] Recipe detail response includes normalized ingredients with status
- [ ] Recipe detail response includes planner config and planner readiness
- [ ] Recipe list response includes `plannerReady`, `classification`, `resolvedIngredients`, and `totalIngredients` per recipe
- [ ] Recipe list supports filtering by `plannerReady`, `classification`, and `normalizationStatus`
- [ ] Planner config create/update via recipe PATCH endpoint
- [ ] Canonical ingredient detail includes recipe usage list (`/ingredients/:id/recipes`)
- [ ] Canonical ingredient reassign-and-delete endpoint moves all references before deletion
- [ ] Source update reconciliation preserves `manually_confirmed` ingredient statuses
- [ ] Resolve dropdown suggestions use ILIKE prefix match ordered by usage count
- [ ] Audit events emitted for all specified actions (including `recipe.renormalized`)
- [ ] All endpoints enforce tenant_id scoping
- [ ] All data passes through immutable models and builders
- [ ] All tests pass (`go test ./... -count=1`)
- [ ] Service builds (`go build ./...`)

### Frontend
- [ ] Recipe detail page shows normalization status per ingredient with color-coded indicators
- [ ] Recipe detail page shows summary badge ("5/8 ingredients resolved")
- [ ] Recipe detail page includes "Re-normalize" button for unresolved ingredients
- [ ] Recipe form page includes ingredient normalization review panel
- [ ] Live preview shows normalization status per ingredient before save
- [ ] Unresolved ingredients show search/select to pick canonical ingredient
- [ ] User can create a new canonical ingredient inline during resolution
- [ ] "Save as alias" checkbox works and creates alias on resolve
- [ ] Recipe form page includes collapsible planner settings section
- [ ] Planner-ready badge and classification appear on recipe detail and recipe list cards
- [ ] Recipe list page supports filtering by planner readiness, classification, and normalization completeness
- [ ] Canonical ingredient list page with search and pagination
- [ ] Canonical ingredient detail page with alias management (add/remove)
- [ ] Canonical ingredient detail page shows linked recipe list
- [ ] Canonical ingredient delete shows reassign option when references exist
- [ ] Canonical ingredient list shows empty state guidance when registry is empty
- [ ] Recipe list cards show normalization summary ("5/8 resolved")
- [ ] "Ingredients" navigation item appears under Recipes section
- [ ] All new components are responsive (mobile + desktop)
