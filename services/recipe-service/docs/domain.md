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
