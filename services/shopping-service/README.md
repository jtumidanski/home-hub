# Shopping Service

The shopping service manages shopping lists and their items. It provides CRUD operations for lists, item management within lists, list archival, and bulk import of ingredients from meal plans.

## External Dependencies

- **PostgreSQL** - Primary data store, using the `shopping` schema
- **Category Service** - Resolves category details (name, sort order) when adding or updating items
- **Recipe Service** - Fetches consolidated meal plan ingredients for import
- **Auth Service** - JWT token validation via JWKS endpoint

## Configuration

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `postgres.home` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `home_hub` | Database user |
| `DB_PASSWORD` | *(empty)* | Database password |
| `DB_NAME` | `home_hub` | Database name |
| `PORT` | `8080` | HTTP server port |
| `JWKS_URL` | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json` | JWKS endpoint for JWT validation |
| `CATEGORY_SERVICE_URL` | `http://category-service:8080` | Category service base URL |
| `RECIPE_SERVICE_URL` | `http://recipe-service:8080` | Recipe service base URL |

Database schema: `shopping`

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
