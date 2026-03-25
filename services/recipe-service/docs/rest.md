# Recipe Service — REST Endpoints

Base path: `/api/v1`

All endpoints require a valid JWT. Tenant and household scoping is automatic via JWT claims.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/recipes/parse` | Parse Cooklang source (preview, no persistence) |
| GET | `/recipes` | List recipes (paginated, searchable, filterable) |
| POST | `/recipes` | Create a recipe |
| GET | `/recipes/{id}` | Get recipe detail with parsed Cooklang |
| PATCH | `/recipes/{id}` | Update a recipe |
| DELETE | `/recipes/{id}` | Soft delete a recipe |
| POST | `/recipes/restorations` | Restore a soft-deleted recipe |
| GET | `/recipes/tags` | List all tags in use for the household |

## Query Parameters (List)

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `search` | string | — | Case-insensitive substring match on title |
| `tag` | string | — | Repeatable. Filter by tag(s) with AND semantics |
| `page[number]` | int | 1 | Page number |
| `page[size]` | int | 20 | Items per page (max 100) |

## Resource Types

- `recipes` — Recipe resource (list and detail variants)
- `recipe-tags` — Tag with usage count
- `recipe-parse` — Parse result (ingredients, steps, errors)
- `recipe-restorations` — Restoration request
