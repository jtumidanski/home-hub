# Recipe Service

The recipe service manages recipe storage and retrieval for households within the Home Hub platform. Recipes are stored using the Cooklang plain-text format and parsed server-side into structured ingredients and steps. The service supports full CRUD operations, tag-based categorization, search, pagination, and soft delete with a time-limited restore window.

## External Dependencies

- **PostgreSQL** — persistent storage (schema: `recipe`)
- **Auth service** — JWT validation via JWKS endpoint

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

## Documentation

- [Domain](docs/domain.md) — domain models, invariants, and processors
- [REST](docs/rest.md) — HTTP endpoints and resource types
- [Storage](docs/storage.md) — database tables, relationships, and indexes
