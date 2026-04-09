# Tracker Service

The tracker service powers the Daily Tracker feature: per-user, customizable habit and wellness tracking with a monthly calendar grid and a completed-month dashboard report.

Each user defines their own tracking items (e.g., "Running", "Sleep quality", "Drinks") with a name, color, scale type, optional range bounds, and a weekly schedule. The service stores entries per item per day, computes month completion using versioned schedule snapshots, and generates per-item report stats once a month is complete. Data is scoped by tenant and user — there is no household sharing.

## External Dependencies

- **PostgreSQL** — persistent storage for tracking items, schedule snapshots, and entries. Uses the `tracker` schema.
- **Auth Service** — provides the JWKS endpoint for JWT validation on protected routes.

## Runtime Configuration

| Variable      | Description                       | Default                                                       |
|---------------|-----------------------------------|---------------------------------------------------------------|
| `DB_HOST`     | PostgreSQL host                   | `postgres.home`                                               |
| `DB_PORT`     | PostgreSQL port                   | `5432`                                                        |
| `DB_USER`     | PostgreSQL user                   | `home_hub`                                                    |
| `DB_PASSWORD` | PostgreSQL password               |                                                               |
| `DB_NAME`     | PostgreSQL database name          | `home_hub`                                                    |
| `PORT`        | HTTP listen port                  | `8080`                                                        |
| `JWKS_URL`    | Auth service JWKS endpoint        | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json` |

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
