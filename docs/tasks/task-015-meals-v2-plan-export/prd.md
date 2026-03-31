# Meals v2: Selected-Recipe Plan Generation + Markdown Ingredient Export — Product Requirements Document

Version: v2
Status: Draft
Created: 2026-03-29

---

## 1. Overview

Meals v2 adds weekly meal planning to the recipe-service. Users select existing recipes, assign them to days and meal slots on a weekly grid, save the plan, and export a consolidated ingredient list as markdown.

This builds directly on the recipe management (task-005) and ingredient normalization (task-014) foundations. The plan domain leverages normalized ingredient data for accurate consolidation and the planner configuration (classification, servings_yield) for recipe metadata.

This version intentionally avoids advanced automated planning logic — no repeat-spacing enforcement, leftovers optimization, opt-out handling, pantry subtraction, or async background generation. Those remain future concerns.

## 2. Goals

Primary goals:

- Allow users to create weekly meal plans by selecting existing recipes
- Support explicit day/slot assignment for plan items (multiple items per slot allowed)
- Provide ingredient consolidation across all plan items with serving scaling
- Export the plan and consolidated ingredients as markdown
- Retain plan history for browsing previous weeks
- Support duplicating a plan to another week
- Show recipe usage history in the recipe selector
- Auto-filter recipes by classification when adding to a slot
- Emit audit events for plan lifecycle actions

Non-goals:

- Free-form recipe ingestion or AI-assisted recipe parsing
- Async planner progress bar workflow
- Meal opt-outs or suggested alternatives
- Automatic leftovers generation
- Repeat-spacing and eat-within constraint solving
- Pantry subtraction or nutrition-aware planning
- Persistent export storage (generated on-demand only)
- Per-recipe ingredient breakdown in the markdown export
- CSV export changes
- New recipe listing endpoints (reuse existing `GET /api/v1/recipes`)

## 3. User Stories

- As an admin, I want to create a weekly meal plan by selecting recipes so that my household knows what meals are planned
- As an admin, I want to assign recipes to specific days and meal slots so that the plan reflects when each meal will be prepared
- As an admin, I want to adjust serving sizes per plan item so that the ingredient list reflects how many people I'm cooking for
- As an admin, I want to lock a finalized plan so that it isn't accidentally modified
- As an admin, I want to export the plan as markdown so that I can copy the ingredient list into a shopping app or share it
- As an admin, I want to browse previous week's plans so that I can reference or reuse past meal plans
- As an admin, I want to duplicate a plan to another week so that I can reuse a plan without re-entering every item
- As an admin, I want to see when I last used a recipe so that I can avoid repeating meals too frequently
- As an admin, I want the recipe selector to auto-filter by classification when I click a slot so that I see relevant recipes immediately
- As a viewer, I want to see the current week's meal plan so that I know what's being cooked

## 4. Functional Requirements

### 4.1 Plan Creation

The system shall allow a user to create a plan for a specified household and week.

Required fields:

- `tenant_id` — from JWT context
- `household_id` — from JWT context
- `starts_on` — any date identifying the start of the plan week (not constrained to Monday; the frontend determines the preferred week start day)
- At least one plan item (may be added after initial plan creation, empty plans are allowed)

The system shall auto-generate a plan name as "Week of {starts_on formatted as Month Day, Year}" (e.g., "Week of April 6, 2026"). Users may override this name.

The system shall enforce one plan per household per `starts_on` date. Attempting to create a duplicate shall return 409 Conflict.

New plans are created in an unlocked state.

### 4.2 Plan Editing

Users with Owner, Admin, or Editor roles may:

- Add plan items to an unlocked plan
- Remove plan items from an unlocked plan
- Update plan item day, slot, serving multiplier, or notes
- Rename the plan
- Lock the plan to prevent further edits
- Unlock a locked plan to resume editing

The system shall reject modifications to locked plans with 409 Conflict, except for the unlock action.

Viewer users may read plans but may not create, edit, lock, unlock, or export.

### 4.3 Plan Items

Each plan item represents a single recipe assigned to a day and slot within the plan week.

Fields:

- `id` — UUID
- `plan_week_id` — FK to plan_week
- `day` — date (must fall within the plan week: starts_on through starts_on + 6 days)
- `slot` — constrained enum: `breakfast`, `lunch`, `dinner`, `snack`, `side`
- `recipe_id` — FK to recipe (must be active, non-deleted at time of creation; no cascade on recipe deletion)
- `serving_multiplier` — decimal, nullable (default 1.0 if both multiplier and planned_servings are null)
- `planned_servings` — integer, nullable
- `notes` — text, nullable
- `position` — integer (ordering within a day+slot, allows multiple items per slot)

Serving scaling precedence:

1. If `planned_servings` is set, effective multiplier = `planned_servings / recipe.servings_yield`
2. If `serving_multiplier` is set, use it directly
3. If neither is set, default multiplier = 1.0

The system shall validate:

- `day` falls within the plan week range
- `slot` is one of the allowed enum values
- `recipe_id` references an active, non-deleted recipe in the same tenant/household (validated at creation time only)
- The plan is not locked

### 4.4 Deleted Recipe Handling

If a recipe is soft-deleted after being added to a plan, the plan item shall remain in the plan. The `recipe_id` FK shall not cascade on recipe deletion.

When displaying plan items, the system shall check if the referenced recipe is deleted and surface that state:

- The plan detail response shall include a `recipe_deleted` boolean on each item
- The UI shall show a visual warning on items referencing deleted recipes
- Deleted-recipe items are still included in ingredient consolidation using their last-known normalized ingredients
- Users may remove deleted-recipe items from unlocked plans but cannot add new items referencing deleted recipes

### 4.5 Plan Duplication

The system shall allow duplicating an existing plan to a different week.

Duplication shall:

- Create a new `plan_week` with the specified target `starts_on` date
- Copy all plan items from the source plan, adjusting each item's `day` by the same offset (e.g., if source starts Monday Apr 6 and target starts Sunday Apr 12, all days shift by +6)
- Set the new plan to unlocked regardless of source plan's lock state
- Auto-generate a new plan name based on the target week
- Validate that no plan already exists for the target household + starts_on

Duplication shall not copy notes from source plan items (start fresh).

### 4.6 Recipe Usage History

When listing recipes for the planner selector, the system shall include usage metadata:

- `last_used_date` — the most recent `plan_item.day` referencing this recipe (across all plans for this household), or null if never used
- `usage_count` — total number of plan items referencing this recipe

This data is computed via a query join against `plan_items` and returned as part of the existing recipe list response when requested via a query parameter (e.g., `?include_usage=true`).

### 4.7 Classification Auto-Filter

When the user clicks a slot cell in the planner UI to add a recipe, the recipe selector shall auto-apply a classification filter matching the slot:

| Slot | Auto-filter classification |
|------|---------------------------|
| breakfast | breakfast |
| lunch | lunch |
| dinner | dinner |
| snack | snack |
| side | side |

The user may clear the filter to see all recipes. This is a UI-only behavior — no backend changes required beyond the existing classification filter on the recipe list endpoint.

### 4.8 Plan History

Plans are retained indefinitely. The API supports listing plans by `starts_on` date, enabling week-by-week browsing. Plans are never automatically deleted.

### 4.9 Ingredient Consolidation

The system shall consolidate ingredients across all plan items in a plan for export.

For each plan item:

1. Read the recipe's normalized ingredients (`recipe_ingredients` table)
2. Apply the plan item's effective serving multiplier to scale quantities
3. If no normalized ingredient data exists for a recipe, fall back to raw parsed Cooklang ingredients

Consolidation rules (intentionally conservative for v2):

Ingredients consolidate only when ALL of the following are true:

- `canonical_ingredient_id` matches (both must be resolved)
- `canonical_unit` matches exactly
- Quantity is numerically aggregatable

Ingredients that cannot be consolidated (unresolved, mismatched units, no canonical mapping) are listed individually with their raw text.

Examples:

- yellow onion (2 count) + yellow onion (1 count) → 3 count yellow onion
- olive oil (2 tbsp) + olive oil (1 tbsp) → 3 tbsp olive oil
- onion (2 count) + onion (100 g) → listed separately (unit family mismatch)

### 4.10 Markdown Export

The system shall generate markdown on-demand (not persisted) from a plan.

Format:

```markdown
# Meal Plan: Week of April 6, 2026

## Monday (2026-04-06)
- **Breakfast:** Overnight Oats
- **Dinner:** Chicken Tacos (serves 6)

## Tuesday (2026-04-07)
- **Dinner:** Pasta Carbonara

## Consolidated Ingredients
- 2 lb chicken breast
- 3 yellow onion
- 4 tbsp olive oil
- 2 jar tomato sauce
- 400 g spaghetti
- 1 cup rolled oats

## Notes
- Some ingredients could not be fully consolidated and were listed as entered.
```

Rules:

- Top-level heading with plan name
- No household name (recipe-service does not own household metadata)
- Days listed in calendar order starting from `starts_on`
- Only days with assigned items are included
- Within each day, slots ordered: breakfast, lunch, dinner, snack, side
- Servings override shown in parentheses only when different from recipe default
- Consolidated ingredients as flat bullet list
- Display format: `{quantity} {unit} {ingredient_name}`
- Zero-quantity ingredients omitted
- Notes section included only when unresolved ingredients exist

### 4.11 Audit Events

The system shall emit audit events to the existing `recipe_audit_events` table using the established `audit.Emit` pattern.

New entity type: `"plan"`

New actions:

| Action | When |
|--------|------|
| `plan.created` | Plan created |
| `plan.updated` | Plan name changed |
| `plan.locked` | Plan locked |
| `plan.unlocked` | Plan unlocked |
| `plan.item_added` | Plan item added |
| `plan.item_updated` | Plan item modified |
| `plan.item_removed` | Plan item deleted |
| `plan.duplicated` | Plan duplicated to another week |
| `plan.exported` | Markdown export generated |

Metadata shall include relevant context (e.g., recipe_id for item events, format for export events).

## 5. API Surface

All endpoints are scoped to the current tenant and household from JWT context.

### 5.1 Plan Management

**POST /api/v1/meals/plans** — Create a plan

Request:
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

- `starts_on` required (any date; not constrained to a specific day of week)
- `name` optional, auto-generated if omitted
- Returns 201 with plan resource
- Returns 409 if plan already exists for this household + starts_on

**GET /api/v1/meals/plans** — List plans

Query parameters:
- `starts_on` — filter by exact week start date
- `page[number]`, `page[size]` — pagination (default page size 20, max 100)

Returns plan list ordered by `starts_on` descending.

**GET /api/v1/meals/plans/{planId}** — Get plan detail

Returns plan with all plan items and their recipe references.

**PATCH /api/v1/meals/plans/{planId}** — Update plan

Request:
```json
{
  "data": {
    "type": "plans",
    "attributes": {
      "name": "Special Week"
    }
  }
}
```

- Only `name` is updatable on the plan itself
- Returns 409 if plan is locked

**POST /api/v1/meals/plans/{planId}/lock** — Lock plan

- Returns 200 with updated plan
- Returns 409 if already locked

**POST /api/v1/meals/plans/{planId}/unlock** — Unlock plan

- Returns 200 with updated plan
- Returns 409 if already unlocked

**POST /api/v1/meals/plans/{planId}/duplicate** — Duplicate plan to another week

Request:
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

- `starts_on` required, target week start date
- Copies all items with day offsets adjusted to target week
- Notes on items are not copied
- New plan is created unlocked with auto-generated name
- Returns 201 with new plan resource
- Returns 409 if plan already exists for target household + starts_on

### 5.2 Plan Items

**POST /api/v1/meals/plans/{planId}/items** — Add item

Request:
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

- `day`, `slot`, `recipe_id` required
- Returns 201 with plan item resource
- Returns 409 if plan is locked
- Returns 400 if day outside plan week range or invalid slot
- Returns 404 if recipe not found or deleted

**PATCH /api/v1/meals/plans/{planId}/items/{itemId}** — Update item

- Any subset of: `day`, `slot`, `serving_multiplier`, `planned_servings`, `notes`, `position`
- Returns 409 if plan is locked

**DELETE /api/v1/meals/plans/{planId}/items/{itemId}** — Remove item

- Returns 204
- Returns 409 if plan is locked

### 5.3 Export

**GET /api/v1/meals/plans/{planId}/export/markdown** — Generate markdown

- Returns 200 with `Content-Type: text/markdown; charset=utf-8`
- Response body is the raw markdown string
- Emits `plan.exported` audit event

### 5.4 Ingredients (Optional Convenience Endpoint)

**GET /api/v1/meals/plans/{planId}/ingredients** — Get consolidated ingredients

Returns the consolidated ingredient list as a JSON:API resource collection, useful for the UI preview without generating full markdown.

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

## 6. Data Model

### 6.1 plan_week (new table in recipe schema)

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL, indexed |
| household_id | UUID | NOT NULL |
| starts_on | DATE | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| locked | BOOLEAN | NOT NULL, default false |
| created_by | UUID | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- Unique: `(tenant_id, household_id, starts_on)`
- Index: `(tenant_id, household_id)` for listing

### 6.2 plan_item (new table in recipe schema)

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| plan_week_id | UUID | NOT NULL, FK → plan_week.id ON DELETE CASCADE |
| day | DATE | NOT NULL |
| slot | VARCHAR(20) | NOT NULL |
| recipe_id | UUID | NOT NULL, FK → recipes.id (no cascade on delete) |
| serving_multiplier | DECIMAL(5,2) | nullable |
| planned_servings | INTEGER | nullable |
| notes | TEXT | nullable |
| position | INTEGER | NOT NULL, default 0 |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- Index: `(plan_week_id)` for item listing
- Index: `(recipe_id)` for recipe usage lookups

Constraints:
- `slot` must be one of: `breakfast`, `lunch`, `dinner`, `snack`, `side`
- `day` must be within `plan_week.starts_on` to `plan_week.starts_on + 6 days` (enforced in processor, not DB)

### 6.3 Existing tables used (no modifications)

- `recipes` — recipe lookup and validation
- `recipe_ingredients` — normalized ingredient data for consolidation
- `canonical_ingredients` — ingredient identity and display names
- `recipe_planner_configs` — classification and servings_yield
- `recipe_audit_events` — audit event storage (new entity_type "plan")

## 7. Service Impact

### 7.1 recipe-service

New internal packages:

| Package | Files | Purpose |
|---------|-------|---------|
| `internal/plan` | model.go, entity.go, builder.go, processor.go, provider.go, resource.go, rest.go | Plan week CRUD, locking |
| `internal/planitem` | model.go, entity.go, builder.go, processor.go, provider.go, resource.go, rest.go | Plan item CRUD |
| `internal/export` | processor.go, markdown.go | Ingredient consolidation and markdown generation |

The `export` package depends on:
- `plan` and `planitem` domains for plan data
- `normalization` domain for reading recipe ingredients
- `ingredient` domain for canonical ingredient display names
- `recipe` domain for recipe metadata (title, servings)

Cross-domain orchestration: The export processor coordinates reads across plan, planitem, recipe, normalization, and ingredient domains. This is a read-only aggregation (similar to dashboard summaries) so it may call multiple processors directly per the architecture guidelines.

### 7.2 frontend

New admin planner screen:
- Week selector (date picker; preferred start day configurable in frontend, defaults to Monday)
- Recipe search/select panel (reuses existing recipe list API with filters + `include_usage=true`)
- Auto-filter by classification when clicking a slot cell
- Day/slot grid for placing selected recipes (multiple items per slot supported)
- Serving multiplier/planned servings input per item
- Deleted recipe warning indicators on affected items
- Ingredient consolidation preview panel
- Save / Lock / Unlock / Duplicate / Export actions
- Markdown preview modal with copy-to-clipboard and download-as-file

### 7.3 Infrastructure

- Add nginx route: `/api/v1/meals` → recipe-service
- Update `docs/architecture.md` routing table

## 8. Non-Functional Requirements

### Performance

- Plan list: < 200ms
- Plan detail with items: < 300ms
- Ingredient consolidation + markdown export: < 500ms for typical weekly plan (≤ 21 items)

These are synchronous operations. No background processing needed for v2 plan sizes.

### Security

- All operations enforce tenant_id and household_id scoping from JWT
- Write operations (create, update, lock, export) require Owner, Admin, or Editor role
- Read operations (list, get) available to all authenticated household members including Viewer
- Plan items validate recipe_id belongs to same tenant/household

### Reliability

- Export generation is read-only — never mutates plan data
- If export fails, the plan remains intact
- Locked plans provide an explicit guard against accidental edits

### Observability

- Structured logging with request_id, user_id, tenant_id, household_id on all operations
- Audit events for all state-changing actions
- OpenTelemetry spans for plan operations and export generation

## 9. Open Questions

None — all questions resolved during spec review.

## 10. Acceptance Criteria

- [ ] An admin can create a plan for a week by specifying any start date (not constrained to Monday)
- [ ] Plan name is auto-generated as "Week of {date}" and can be overridden
- [ ] Empty plans (no items) can be saved
- [ ] Only one plan per household per week start date (unique constraint)
- [ ] Recipes can be added to a plan as items assigned to specific days and slots
- [ ] Multiple items are allowed per day+slot (ordered by position)
- [ ] Slot values are constrained to: breakfast, lunch, dinner, snack, side
- [ ] Day values are validated to fall within the plan week (starts_on to starts_on + 6)
- [ ] Recipe references are validated as active and non-deleted at creation time
- [ ] Plan items referencing subsequently deleted recipes remain visible with a `recipe_deleted` indicator
- [ ] Serving multiplier and planned servings are supported with correct precedence
- [ ] Plans can be locked to prevent edits and unlocked to resume
- [ ] Locked plans reject all modifications except unlock
- [ ] Plan history is retained and browsable by week
- [ ] A plan can be duplicated to another week with day offsets adjusted
- [ ] Recipe selector shows `last_used_date` and `usage_count` when `include_usage=true`
- [ ] Recipe selector auto-filters by classification matching the selected slot (UI behavior)
- [ ] Ingredient consolidation correctly merges matching canonical ingredients with matching units
- [ ] Unresolved or unit-mismatched ingredients are listed individually
- [ ] Serving scaling is applied to ingredient quantities during consolidation
- [ ] Markdown export includes plan heading, meals by day/slot, and consolidated ingredients (no household name)
- [ ] Markdown export includes Notes section when unresolved ingredients exist
- [ ] Markdown can be previewed, copied to clipboard, and downloaded as .md file in the UI
- [ ] Audit events are emitted for: plan created, updated, locked, unlocked, duplicated, item added/updated/removed, exported
- [ ] Viewer role users can read plans but cannot create, edit, lock, or export
- [ ] All data is scoped to tenant_id and household_id
