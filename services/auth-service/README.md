# Auth Service

The auth service handles authentication for the Home Hub platform. It implements OIDC-based login via external identity providers, issues and manages JWT access tokens and refresh tokens, and exposes a JWKS endpoint for other services to verify tokens.

## External Dependencies

- **PostgreSQL** — primary data store, schema `auth`
- **External OIDC Provider** — currently Google, used for user authentication via OpenID Connect

## Runtime Configuration

| Variable           | Default                                                       | Description                        |
|--------------------|---------------------------------------------------------------|------------------------------------|
| `DB_HOST`          | `postgres.home`                                               | PostgreSQL host                    |
| `DB_PORT`          | `5432`                                                        | PostgreSQL port                    |
| `DB_USER`          | `home_hub`                                                    | Database user                      |
| `DB_PASSWORD`      | (empty)                                                       | Database password                  |
| `DB_NAME`          | `home_hub`                                                    | Database name                      |
| `PORT`             | `8080`                                                        | HTTP listen port                   |
| `JWT_PRIVATE_KEY`  | (required)                                                    | PEM-encoded RSA private key        |
| `JWT_KEY_ID`       | `home-hub-1`                                                  | Key ID included in JWT headers     |
| `OIDC_ISSUER_URL`  | `https://accounts.google.com`                                 | OIDC provider issuer URL           |
| `OIDC_CLIENT_ID`   | (required)                                                    | OAuth client ID                    |
| `OIDC_CLIENT_SECRET` | (required)                                                  | OAuth client secret                |

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
