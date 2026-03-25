# Recipe Service — Domain

## Overview

The recipe service manages recipe storage and retrieval for households. Recipes are stored using the Cooklang plain-text format and parsed server-side into structured ingredients and steps.

## Domains

### recipe

The core domain. Manages recipe CRUD operations, Cooklang parsing, tag management, and soft delete/restore.

**Key concepts:**
- **Recipe**: A household-scoped entity containing a title, description, raw Cooklang source, metadata (servings, prep/cook time, source URL), and tags
- **Cooklang source**: The authoritative representation of recipe content. Parsed on read to produce structured ingredients and steps
- **Tags**: Free-form strings (normalized lowercase) associated with a recipe for categorization and filtering
- **Soft delete**: Recipes are soft-deleted with a 3-day restore window

**Business rules:**
- Title and source are required
- Cooklang source is validated on create and update
- Tags are deduplicated and normalized to lowercase
- Ingredients are aggregated across steps (combined quantities when units match)
- All data is scoped by tenant_id and household_id

### cooklang (internal package)

Cooklang parser that converts plain-text recipe format into structured data.

**Supported syntax:**
- `@ingredient{quantity%unit}` — ingredient with quantity and unit
- `@ingredient{quantity}` — ingredient with quantity, no unit
- `@ingredient` — ingredient without quantity (single word)
- `#cookware{}` — cookware item
- `~{quantity%unit}` — timer
- `--` — line comment (stripped)
- `[- ... -]` — block comment (stripped)
- Blank lines separate steps
