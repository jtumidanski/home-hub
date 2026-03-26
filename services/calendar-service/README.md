# Calendar Service

The calendar service manages Google Calendar integration for households. It connects to Google Calendar via OAuth 2.0, syncs events on a configurable interval, and serves a unified household calendar view with per-user privacy controls.

Each household member can connect their own Google account. The service merges all members' events into a single calendar view, applying privacy masking so that private events appear as "Busy" to non-owners.

## External Dependencies

- **PostgreSQL** — persistent storage for connections, sources, events, and OAuth state. Uses the `calendar` schema.
- **Google Calendar API** — reads calendar lists and events via OAuth 2.0 with `calendar.readonly` and `email` scopes.
- **Auth Service** — provides JWKS endpoint for JWT validation on protected routes.

## Runtime Configuration

| Variable                        | Description                                      | Default                                                        |
|---------------------------------|--------------------------------------------------|----------------------------------------------------------------|
| `DB_HOST`                       | PostgreSQL host                                  | `postgres.home`                                                |
| `DB_PORT`                       | PostgreSQL port                                  | `5432`                                                         |
| `DB_USER`                       | PostgreSQL user                                  | `home_hub`                                                     |
| `DB_PASSWORD`                   | PostgreSQL password                              |                                                                |
| `DB_NAME`                       | PostgreSQL database name                         | `home_hub`                                                     |
| `PORT`                          | HTTP listen port                                 | `8080`                                                         |
| `JWKS_URL`                      | Auth service JWKS endpoint                       | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json`  |
| `GOOGLE_CALENDAR_CLIENT_ID`     | Google OAuth client ID                           |                                                                |
| `GOOGLE_CALENDAR_CLIENT_SECRET` | Google OAuth client secret                       |                                                                |
| `CALENDAR_TOKEN_ENCRYPTION_KEY` | Base64-encoded 32-byte AES-256 key for token encryption |                                                         |
| `SYNC_INTERVAL_MINUTES`         | Background sync interval in minutes              | `15`                                                           |

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
