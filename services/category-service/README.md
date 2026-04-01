# Category Service

The category service manages grocery/shopping categories within Home Hub. Each tenant receives a set of default categories on first access, and can create, update, reorder, and delete categories as needed. Categories are scoped per tenant with a unique name constraint enforced at the database level.

## Dependencies

- **PostgreSQL** -- persistent storage (schema: `category`, table: `categories`)
- **Auth Service** -- JWKS endpoint for JWT validation

## Configuration

| Variable      | Default                                                      |
|---------------|--------------------------------------------------------------|
| `DB_HOST`     | `postgres.home`                                              |
| `DB_PORT`     | `5432`                                                       |
| `DB_USER`     | `home_hub`                                                   |
| `DB_PASSWORD` | *(empty)*                                                    |
| `DB_NAME`     | `home_hub`                                                   |
| `PORT`        | `8080`                                                       |
| `JWKS_URL`    | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json` |

The database schema is always set to `category`.

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
