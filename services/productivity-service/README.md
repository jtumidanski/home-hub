# Productivity Service

The productivity service manages tasks and reminders for households in the Home Hub platform. It supports task lifecycle management with soft deletes and restoration, reminder scheduling with snooze and dismissal, optional owner assignment for tasks and reminders, and provides aggregated summary endpoints for dashboards.

## External Dependencies

- **PostgreSQL** — primary data store, schema `productivity`
- **Auth Service** — JWKS endpoint for JWT validation

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
