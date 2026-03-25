# Weather Service

Provides current weather conditions, 7-day forecasts, and geocoding search for Home Hub households. Weather data is sourced from the Open-Meteo API and cached per household in PostgreSQL. A background refresh loop keeps cached data up to date.

## External Dependencies

- **PostgreSQL** — stores cached weather data in the `weather` schema
- **Open-Meteo API** — upstream provider for forecast and geocoding data
- **Auth Service** — JWKS endpoint for JWT validation

## Configuration

| Variable                  | Default                                                         | Description                  |
|---------------------------|-----------------------------------------------------------------|------------------------------|
| DB_HOST                   | postgres.home                                                   | PostgreSQL host              |
| DB_PORT                   | 5432                                                            | PostgreSQL port              |
| DB_USER                   | home_hub                                                        | PostgreSQL user              |
| DB_PASSWORD               |                                                                 | PostgreSQL password          |
| DB_NAME                   | home_hub                                                        | PostgreSQL database name     |
| PORT                      | 8080                                                            | HTTP listen port             |
| JWKS_URL                  | http://auth-service:8080/api/v1/auth/.well-known/jwks.json      | JWKS endpoint for JWT auth   |
| REFRESH_INTERVAL_MINUTES  | 30                                                              | Background refresh interval  |
| CACHE_TTL_MINUTES         | 30                                                              | Cache TTL                    |

## Documentation

- [Domain](docs/domain.md)
- [REST API](docs/rest.md)
- [Storage](docs/storage.md)
