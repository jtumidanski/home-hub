# Recipe Service

The recipe service manages recipes, ingredients, meal planning, and ingredient consolidation for households within the Home Hub platform. Recipes are stored using the Cooklang plain-text format and parsed server-side into structured ingredients and steps. The service supports full recipe CRUD, tag-based categorization, search, pagination, soft delete with time-limited restore, canonical ingredient management with alias-based normalization, weekly meal plan creation with per-item serving control, and ingredient consolidation for shopping list export.

## External Dependencies

- **PostgreSQL** — persistent storage (schema: `recipe`)
- **Auth service** — JWT validation via JWKS endpoint
- **S3-compatible object store** (RustFS) — recipe image renditions
- **Category service** — ingredient category lookup for sorting and display

## Runtime Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `postgres.home` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `home_hub` | Database user |
| `DB_PASSWORD` | _(empty)_ | Database password |
| `DB_NAME` | `home_hub` | Database name |
| `PORT` | `8080` | HTTP listen port |
| `JWKS_URL` | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json` | JWKS endpoint for JWT validation |
| `CATEGORY_SERVICE_URL` | `http://category-service:8080` | Category service base URL |
| `S3_ENDPOINT` | _(empty)_ | Object store host:port for recipe images; empty `S3_ENDPOINT` disables image endpoints (503) |
| `S3_BUCKET` | `home-hub-recipes` | Bucket for recipe image renditions |
| `S3_ACCESS_KEY` | _(empty)_ | Object store access key |
| `S3_SECRET_KEY` | _(empty)_ | Object store secret key |
| `S3_USE_SSL` | `false` | Use TLS to reach the object store |
| `S3_REGION` | `us-east-1` | Object store region |

## Documentation

- [Domain](docs/domain.md) — domain models, invariants, and processors
- [REST](docs/rest.md) — HTTP endpoints and resource types
- [Storage](docs/storage.md) — database tables, relationships, and indexes
