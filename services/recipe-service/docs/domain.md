# Recipe Service — Domain

## Recipe

### Responsibility

Manages recipe CRUD operations, Cooklang parsing, tag management, and soft delete with time-limited restore.

### Core Models

**Model** — immutable domain model representing a recipe.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| tenantID | UUID | Owning tenant |
| householdID | UUID | Owning household |
| title | string | Recipe title |
| description | string | Optional description |
| source | string | Raw Cooklang source text |
| servings | *int | Optional serving count |
| prepTimeMinutes | *int | Optional prep time in minutes |
| cookTimeMinutes | *int | Optional cook time in minutes |
| sourceURL | string | Optional source URL |
| tags | []string | Categorization tags |
| deletedAt | *time.Time | Soft delete timestamp (nil if active) |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

**Builder** — constructs Model instances with validation.

### Invariants

- Title is required (`ErrTitleRequired`)
- Source is required (`ErrSourceRequired`)
- Cooklang source is validated on create and update; invalid syntax is rejected
- Tags are deduplicated and normalized to lowercase
- All data is scoped by tenant ID and household ID
- Soft-deleted recipes can only be restored within a 3-day window (`restoreWindowDays = 3`)
- A recipe that is not deleted cannot be restored (`ErrNotDeleted`)
- List pagination defaults: page 1, page size 20, max page size 100

### Processors

**Processor** — orchestrates recipe operations.

| Method | Description |
|--------|-------------|
| `Create` | Validates attrs and Cooklang syntax, derives metadata fields (tags, sourceURL, servings, prepTime, cookTime, title) from Cooklang metadata when not explicitly provided, persists entity, returns model and parse result |
| `Get` | Retrieves a recipe by ID (excludes soft-deleted), parses Cooklang source, returns model and parse result |
| `List` | Retrieves paginated recipes with optional search (case-insensitive title substring) and tag filtering (AND semantics), returns models and total count |
| `Update` | Applies partial updates to a recipe, re-validates Cooklang source if changed, re-derives tags and source URL from updated metadata, returns model and parse result |
| `Delete` | Soft-deletes a recipe by setting `deleted_at` |
| `Restore` | Restores a soft-deleted recipe within the 3-day window, records restoration event |
| `ListTags` | Returns all tags in use (across non-deleted recipes) with usage counts, ordered by count descending |
| `ParseSource` | Parses and validates Cooklang source without persistence, returns parse result with any errors |
| `ByIDProvider` | Returns a provider function that fetches a recipe by ID |

## Cooklang (internal package)

### Responsibility

Parses Cooklang plain-text recipe format into structured data (ingredients, steps, metadata).

### Core Models

**ParseResult** — output of parsing Cooklang source.

| Field | Type | Description |
|-------|------|-------------|
| Ingredients | []Ingredient | Aggregated ingredient list |
| Steps | []Step | Ordered recipe steps |
| Metadata | Metadata | Extracted metadata |
| Errors | []ParseError | Validation errors (if any) |

**Ingredient** — a single ingredient entry.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Ingredient name |
| Quantity | string | Quantity value |
| Unit | string | Unit of measure |

**Step** — a single recipe step.

| Field | Type | Description |
|-------|------|-------------|
| Number | int | Step number (1-indexed) |
| Section | string | Section name (if under a section header) |
| Segments | []Segment | Ordered segments composing the step text |

**Segment** — a typed portion of step text.

| Field | Type | Description |
|-------|------|-------------|
| Type | SegmentType | One of: text, ingredient, cookware, timer, reference |
| Value | string | Text content (for text segments) |
| Name | string | Name (for ingredient, cookware, reference segments) |
| Quantity | string | Quantity (for ingredient, timer segments) |
| Unit | string | Unit (for ingredient, timer segments) |
| Path | string | File path (for reference segments) |

**Metadata** — metadata extracted from `--- ... ---` blocks.

| Field | Type | Description |
|-------|------|-------------|
| Tags | []string | Recipe tags |
| Source | string | Source URL |
| Title | string | Recipe title |
| Servings | string | Serving count |
| PrepTime | string | Prep time |
| CookTime | string | Cook time |
| Notes | []string | Notes extracted from blockquote lines |
| Extra | map[string]string | Additional key-value pairs |

### Invariants

- Maximum source size: 64 KB (`MaxSourceSize`)
- Duplicate ingredients (by name, case-insensitive) are aggregated; quantities are summed when units match
- Metadata blocks are delimited by `---` on their own line
- Comments (`--` line, `[- ... -]` block) are stripped before parsing
- Blockquote lines (`> ...`) are stripped and collected as notes
- Section headers (`= SectionName`) create named sections; subsequent steps inherit the section name

### Supported Syntax

| Syntax | Description |
|--------|-------------|
| `@ingredient{quantity%unit}` | Ingredient with quantity and unit |
| `@ingredient{quantity}` | Ingredient with quantity, no unit |
| `@ingredient` | Single-word ingredient without quantity |
| `@multi word name{qty%unit}` | Multi-word ingredient name (brace required) |
| `@./path/to/recipe{}` | Recipe reference |
| `#cookware{}` | Cookware item |
| `~{quantity%unit}` | Timer |
| `--` | Line comment |
| `[- ... -]` | Block comment |
| `> text` | Note (blockquote) |
| `= Section` | Section header |
| `--- ... ---` | Metadata block |

### Processors

| Function | Description |
|----------|-------------|
| `Parse` | Parses Cooklang source into a `ParseResult` with ingredients, steps, and metadata |
| `Validate` | Checks Cooklang source for syntax errors (unclosed braces) without full parsing |
| `ParseMinutes` | Extracts integer minutes from strings like "20 minutes" or "35 min" |
| `ParseServings` | Extracts integer serving count from strings like "4" or "4 servings" |

## Ingredient

### Responsibility

Manages a canonical ingredient registry with display names, unit family classification, category assignment, and alias management.

### Core Models

**Model** — immutable canonical ingredient.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| tenantID | UUID | Owning tenant |
| name | string | Normalized lowercase name (unique per tenant) |
| displayName | string | Human-readable display name |
| unitFamily | string | Unit family classification: count, weight, volume, or empty |
| categoryID | *UUID | Optional category reference (external category-service) |
| categoryName | string | Denormalized category name (populated at query time) |
| aliases | []Alias | Associated alias names |
| aliasCount | int | Number of aliases |
| usageCount | int | Number of recipe ingredients referencing this ingredient |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

**Alias** — a name alias for a canonical ingredient.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| name | string | Alias name |

**Builder** — constructs Model instances with validation.

### Invariants

- Name is required (`ErrNameRequired`)
- Unit family must be count, weight, volume, or empty (`ErrInvalidUnitFamily`)
- Names are normalized to lowercase and trimmed
- Alias names cannot conflict with existing canonical ingredient names or other aliases within the same tenant (`ErrAliasConflict`)
- Search pagination defaults: page 1, page size 20, max page size 100 (200 for SearchWithUsage)

### Processors

**Processor** — orchestrates ingredient operations.

| Method | Description |
|--------|-------------|
| `Create` | Normalizes name to lowercase, validates, persists entity |
| `Get` | Retrieves a canonical ingredient by ID |
| `Search` | Searches ingredients by name or alias with pagination |
| `SearchWithUsage` | Searches ingredients with usage count and category name, supports category filtering |
| `Update` | Partially updates name, displayName, unitFamily, and/or categoryID |
| `Delete` | Nullifies references in recipe_ingredients (setting them to unresolved), then deletes the ingredient |
| `AddAlias` | Adds an alias after checking for conflicts with canonical names and existing aliases |
| `RemoveAlias` | Removes an alias by ID |
| `GetUsageCount` | Returns count of recipe ingredients referencing this canonical ingredient |
| `GetIngredientRecipes` | Returns paginated list of recipe references for a canonical ingredient |
| `Reassign` | Moves all recipe ingredient references from one canonical ingredient to another |
| `BulkCategorize` | Assigns a category to multiple ingredients in a single transaction |
| `Suggest` | Returns ingredients matching a prefix, ordered by usage count descending |
| `ByIDProvider` | Returns a provider function that fetches an ingredient by ID |

## Normalization

### Responsibility

Links raw recipe ingredients to canonical ingredients and manages resolution status.

### Core Models

**Model** — immutable recipe ingredient with normalization state.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| tenantID | UUID | Owning tenant |
| householdID | UUID | Owning household |
| recipeID | UUID | Parent recipe |
| rawName | string | Original ingredient name from Cooklang source |
| rawQuantity | string | Original quantity string |
| rawUnit | string | Original unit string |
| position | int | Position in recipe ingredient list |
| canonicalIngredientID | *UUID | Resolved canonical ingredient (nil if unresolved) |
| canonicalUnit | string | Resolved canonical unit name |
| normalizationStatus | Status | One of: matched, alias_matched, unresolved, manually_confirmed |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

**PreviewResult** — result of previewing normalization without persistence.

| Field | Type | Description |
|-------|------|-------------|
| RawName | string | Original ingredient name |
| Position | int | Position in ingredient list |
| Status | Status | Predicted normalization status |
| CanonicalIngredientID | *UUID | Matched canonical ingredient ID |
| CanonicalName | string | Matched canonical ingredient name |
| CanonicalUnit | string | Resolved canonical unit |
| CanonicalUnitFamily | string | Resolved unit family |

**Builder** — constructs Model instances with validation.

### Invariants

- Raw names are normalized to lowercase and trimmed before matching
- Matching strategy (in order): exact canonical name match, alias match, text-normalized exact match, text-normalized alias match
- Text normalization strips leading articles (the, a, an), collapses whitespace, and removes trailing "s" (simple depluralization)
- Manually confirmed resolutions are preserved during reconciliation
- Unit resolution uses the unit registry to map raw units to canonical forms

### Unit Registry

Maps raw unit strings to canonical unit identities with family classification.

| Family | Canonical Units |
|--------|----------------|
| count | each, piece, whole, clove, head, bunch, sprig, stalk, slice, pinch, dash |
| weight | gram, kilogram, ounce, pound |
| volume | teaspoon, tablespoon, cup, fluid ounce, milliliter, liter |

### Processors

**Processor** — orchestrates normalization operations.

| Method | Description |
|--------|-------------|
| `NormalizeIngredients` | Runs normalization pipeline on parsed ingredients for a recipe, persists results |
| `ReconcileIngredients` | Re-normalizes after recipe update, preserving manually confirmed resolutions and reusing existing entities by raw name |
| `ResolveIngredient` | Manually resolves an ingredient to a canonical ingredient, optionally creating an alias, emits audit events |
| `Renormalize` | Re-runs matching for all non-manually-confirmed ingredients in a recipe, emits audit event with summary |
| `GetByRecipeID` | Returns all normalization records for a recipe |
| `PreviewNormalization` | Previews normalization results without persistence |

## Planner

### Responsibility

Stores per-recipe configuration for meal planning: classification, servings yield, and scheduling constraints.

### Core Models

**Model** — immutable planner configuration for a recipe.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| recipeID | UUID | Associated recipe (unique index) |
| classification | string | Meal type classification (e.g., main, side, breakfast) |
| servingsYield | *int | Override servings yield for planning |
| eatWithinDays | *int | Scheduling constraint: eat within N days |
| minGapDays | *int | Scheduling constraint: minimum days between uses |
| maxConsecutiveDays | *int | Scheduling constraint: maximum consecutive days |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

**Readiness** — computed planner readiness status.

| Field | Type | Description |
|-------|------|-------------|
| Ready | bool | Whether the recipe is ready for planning |
| Issues | []string | List of issues preventing readiness |

**Builder** — constructs Model instances.

### Invariants

- One config per recipe (unique index on recipeID)
- Readiness requires classification to be set and servings to be available (from either planner config or recipe)

### Processors

**Processor** — orchestrates planner configuration.

| Method | Description |
|--------|-------------|
| `CreateOrUpdate` | Creates or updates planner config for a recipe (upsert semantics) |
| `GetByRecipeID` | Retrieves planner config by recipe ID |

**ComputeReadiness** — standalone function that evaluates whether a recipe is ready for meal planning.

## Plan

### Responsibility

Manages weekly meal plans with locking, duplication, and export capabilities.

### Core Models

**Model** — immutable weekly meal plan.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| tenantID | UUID | Owning tenant |
| householdID | UUID | Owning household |
| startsOn | time.Time | Week start date |
| name | string | Plan name (defaults to "Week of {date}") |
| locked | bool | Whether the plan is locked for editing |
| createdBy | UUID | User who created the plan |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

**Builder** — constructs Model instances with validation.

### Invariants

- startsOn is required (`ErrStartsOnRequired`)
- Name defaults to "Week of {date}" if not provided
- Unique constraint on (tenant_id, household_id, starts_on) — one plan per household per week (`ErrAlreadyExists`)
- Locked plans cannot have their name updated (`ErrLocked`)
- Lock/unlock operations validate current lock state (`ErrAlreadyLocked`, `ErrNotLocked`)
- All mutations emit audit events

### Processors

**Processor** — orchestrates plan operations.

| Method | Description |
|--------|-------------|
| `Create` | Creates a weekly plan after checking uniqueness |
| `Get` | Retrieves a plan by ID |
| `List` | Lists plans with optional starts_on filter and pagination |
| `UpdateName` | Updates plan name (rejects if locked) |
| `Lock` | Locks a plan (rejects if already locked) |
| `Unlock` | Unlocks a plan (rejects if not locked) |
| `Duplicate` | Creates a new plan for a target week, copies items with day offset |
| `ExportMarkdown` | Generates a markdown export of the plan with meal schedule and shopping list |
| `Delete` | Deletes a plan by ID |

## Plan Item

### Responsibility

Manages individual meal slot assignments within a weekly plan.

### Core Models

**Model** — immutable meal assignment within a plan.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| planWeekID | UUID | Parent plan |
| day | time.Time | Date of the meal |
| slot | string | Meal slot: breakfast, lunch, dinner, snack, or side |
| recipeID | UUID | Assigned recipe |
| servingMultiplier | *float64 | Optional serving multiplier |
| plannedServings | *int | Optional explicit planned servings |
| notes | *string | Optional notes |
| position | int | Ordering position within day+slot |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

**Builder** — constructs Model instances with validation.

### Invariants

- Day is required (`ErrDayRequired`)
- Slot must be one of: breakfast, lunch, dinner, snack, side (`ErrInvalidSlot`)
- Recipe ID is required (`ErrRecipeIDRequired`)
- Day must fall within the parent plan's week range (startsOn to startsOn+6 days) (`ErrDayOutOfRange`)
- Position auto-increments within day+slot if not provided
- Locked plans reject all item mutations
- All mutations emit audit events
- Notes are not copied during plan duplication

### Processors

**Processor** — orchestrates plan item operations.

| Method | Description |
|--------|-------------|
| `AddItem` | Adds a meal assignment to a plan, auto-assigns position if not provided |
| `UpdateItem` | Partially updates a plan item (day, slot, servingMultiplier, plannedServings, notes, position) |
| `RemoveItem` | Removes a plan item |
| `GetByPlanWeekID` | Returns all items for a plan, ordered by day then position |
| `CountByPlanWeekID` | Returns the count of items in a plan |
| `GetRecipeUsage` | Returns usage statistics (last used day, usage count) for a set of recipe IDs |
| `CopyItems` | Copies all items from one plan to another with a day offset (notes excluded) |

## Export

### Responsibility

Consolidates ingredients from meal plan items for shopping list generation and markdown export.

### Processors

**Processor** — orchestrates export operations.

| Method | Description |
|--------|-------------|
| `ConsolidateIngredients` | Builds a merged ingredient list from all items in a plan, aggregating quantities by canonical ingredient and base unit, with category-based sorting |
| `GenerateMarkdown` | Produces a markdown string with daily meal schedule and categorized shopping list |

### Consolidation Logic

- For each plan item, computes an effective multiplier from planned servings or serving multiplier
- Resolved ingredients are grouped by canonical ingredient ID and aggregated by base unit
- Base unit conversion: volume units convert to teaspoons, weight units convert to grams (metric) or ounces (imperial)
- Display units are chosen for readability (e.g., teaspoons promoted to tablespoons or cups)
- Unresolved ingredients are listed individually without aggregation
- Results are sorted by category sort order, then alphabetically by display name
- Uncategorized ingredients sort last

## Audit

### Responsibility

Records audit events for compliance and debugging across all domains.

### Processors

**Emit** — standalone function that writes an audit event record.

| Parameter | Type | Description |
|-----------|------|-------------|
| tenantId | UUID | Tenant scope |
| entityType | string | Entity type (e.g., recipe, plan, recipe_ingredient, canonical_ingredient) |
| entityId | UUID | Entity identifier |
| action | string | Action name (e.g., recipe.created, plan.locked, normalization.corrected) |
| actorId | UUID | User who performed the action |
| metadata | map[string]interface{} | Optional JSON metadata |
