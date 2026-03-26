# Package Service

The package service tracks household package deliveries across USPS, UPS, and FedEx. It provides carrier auto-detection from tracking numbers, background status polling with adaptive intervals, package lifecycle management (archive/delete), and privacy controls for individual packages.

Each household member can add packages with tracking numbers. The service polls carrier APIs in the background to update status and estimated delivery dates. Private packages are redacted for non-owners in all API responses.

## External Dependencies

- **PostgreSQL** — persistent storage for packages and tracking events. Uses the `package` schema.
- **USPS Tracking API v3** — OAuth 2.0 client credentials, package tracking.
- **UPS Tracking API v1** — OAuth 2.0 client credentials, package tracking, 250/day rate limit.
- **FedEx Track API v1** — OAuth 2.0 client credentials, package tracking, 500/day rate limit.
- **Auth Service** — provides JWKS endpoint for JWT validation on protected routes.

## Runtime Configuration

| Variable                                | Description                                      | Default                                                       |
|-----------------------------------------|--------------------------------------------------|---------------------------------------------------------------|
| `DB_HOST`                               | PostgreSQL host                                  | `postgres.home`                                               |
| `DB_PORT`                               | PostgreSQL port                                  | `5432`                                                        |
| `DB_USER`                               | PostgreSQL user                                  | `home_hub`                                                    |
| `DB_PASSWORD`                           | PostgreSQL password                              |                                                               |
| `DB_NAME`                               | PostgreSQL database name                         | `home_hub`                                                    |
| `PORT`                                  | HTTP listen port                                 | `8080`                                                        |
| `JWKS_URL`                              | Auth service JWKS endpoint                       | `http://auth-service:8080/api/v1/auth/.well-known/jwks.json` |
| `USPS_CLIENT_ID`                        | USPS OAuth client ID                             |                                                               |
| `USPS_CLIENT_SECRET`                    | USPS OAuth client secret                         |                                                               |
| `UPS_CLIENT_ID`                         | UPS OAuth client ID                              |                                                               |
| `UPS_CLIENT_SECRET`                     | UPS OAuth client secret                          |                                                               |
| `FEDEX_API_KEY`                         | FedEx API key                                    |                                                               |
| `FEDEX_SECRET_KEY`                      | FedEx secret key                                 |                                                               |
| `FEDEX_SANDBOX`                         | Use FedEx sandbox environment                    | `false`                                                       |
| `PACKAGE_TOKEN_ENCRYPTION_KEY`          | Base64-encoded 32-byte AES-256 key               |                                                               |
| `PACKAGE_POLL_INTERVAL_MINUTES`         | Background poll interval in minutes              | `30`                                                          |
| `PACKAGE_POLL_INTERVAL_URGENT_MINUTES`  | Poll interval for out-for-delivery packages      | `15`                                                          |
| `PACKAGE_ARCHIVE_AFTER_DAYS`            | Days after delivery before auto-archive          | `7`                                                           |
| `PACKAGE_DELETE_AFTER_DAYS`             | Days after archive before hard delete            | `30`                                                          |
| `PACKAGE_STALE_AFTER_DAYS`             | Days with no status change before marking stale  | `14`                                                          |
| `PACKAGE_MAX_ACTIVE_PER_HOUSEHOLD`      | Maximum non-archived packages per household      | `25`                                                          |

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
