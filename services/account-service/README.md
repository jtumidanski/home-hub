# Account Service

The account service manages multi-tenant identity and access for the Home Hub platform. It owns tenants, households, memberships, user preferences, and invitations, and provides a resolved application context that combines these into a single response for the frontend.

## External Dependencies

- **PostgreSQL** — primary data store, schema `account`
- **Auth Service** — JWKS endpoint for JWT validation
- **Auth Database** — `auth.users` table, joined for member display names

## Runtime Configuration

| Variable      | Default                                                              | Description              |
|---------------|----------------------------------------------------------------------|--------------------------|
| `DB_HOST`     | `postgres.home`                                                      | PostgreSQL host          |
| `DB_PORT`     | `5432`                                                               | PostgreSQL port          |
| `DB_USER`     | `home_hub`                                                           | Database user            |
| `DB_PASSWORD` | (empty)                                                              | Database password        |
| `DB_NAME`     | `home_hub`                                                           | Database name            |
| `PORT`        | `8080`                                                               | HTTP listen port         |
| `JWKS_URL`    | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json`        | JWT Key Set URL          |

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
