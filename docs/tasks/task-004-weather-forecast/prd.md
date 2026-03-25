# Weather Forecast — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-25
---

## 1. Overview

Home Hub users want at-a-glance weather visibility for the location where their household resides. This feature adds current conditions and a 7-day forecast, tied to each household's geographic location.

Weather data is fetched from the Open-Meteo public API, cached in a new weather-service, and refreshed periodically via an in-process background ticker. The household model in account-service is extended with latitude/longitude fields, and users set their location through a city/place search with autocomplete powered by Open-Meteo's geocoding API.

Temperature and weather data are displayed in the household's configured unit system (metric or imperial), using the existing `units` field on the household model.

## 2. Goals

Primary goals:
- Display current temperature and today's min/max on the dashboard
- Provide a dedicated weather page with a 7-day forecast
- Allow users to set household location via city/place search with autocomplete
- Cache weather data to avoid hitting rate limits
- Refresh weather data periodically in the background
- Respect the household's unit preference (metric/imperial)

Non-goals:
- Hourly forecasts
- Weather alerts or push notifications
- Historical weather data
- Multiple locations per household
- Weather-based automations or triggers
- Per-user unit override (uses household-level `units`)

## 3. User Stories

- As a household member, I want to see the current temperature on my dashboard so I can plan my day at a glance.
- As a household member, I want to see today's high and low temperature on the dashboard so I know the expected range.
- As a household member, I want to view a 7-day forecast so I can plan for the week ahead.
- As a household member, I want each forecast day to show min/max temperature and a weather summary (rain, snow, clear, etc.) so I understand upcoming conditions.
- As a household owner, I want to set my household's location by searching for a city or place so the weather reflects where I live.
- As a household member, I want weather displayed in my household's preferred units (metric or imperial) so readings are intuitive.

## 4. Functional Requirements

### 4.1 Household Location

- Add `latitude` (float64, nullable) and `longitude` (float64, nullable) fields to the household model in account-service.
- Both fields must be set together or both null. Partial coordinates are invalid.
- The PATCH `/api/v1/households/{id}` endpoint accepts the new fields.
- Location fields are included in the household JSON:API resource response.

### 4.2 Geocoding / Place Search

- The weather-service exposes a geocoding search endpoint that proxies Open-Meteo's Geocoding API.
- The endpoint accepts a search query string and returns a list of matching places with name, country, admin region, latitude, and longitude.
- The frontend uses this endpoint to power an autocomplete input on the household settings page.
- When a user selects a place, the frontend PATCHes the household with the selected latitude/longitude.

### 4.3 Weather Data Retrieval

- The weather-service fetches weather data from the Open-Meteo Forecast API for each household that has a location set.
- Data fetched per household:
  - Current conditions: temperature, weather code (WMO)
  - Daily forecast (7 days): date, min temperature, max temperature, weather code
- The service converts Open-Meteo WMO weather codes into human-readable summary strings (e.g., "Clear", "Partly Cloudy", "Rain", "Snow") and a corresponding icon key (e.g., "sun", "cloud-sun", "cloud-rain", "snowflake"). Both are returned in API responses so the frontend does not duplicate the mapping logic.
- Unit system is determined by the household's `units` field: `metric` requests Celsius/km, `imperial` requests Fahrenheit/miles. The unit parameter is passed to Open-Meteo at fetch time.

### 4.4 Caching

- Weather data is cached in the weather-service database, keyed by household ID.
- Cache includes: current conditions, 7-day forecast, the unit system used, and a `fetched_at` timestamp.
- The weather API endpoint serves cached data. If no cache exists for a household, a synchronous fetch is triggered before responding.
- Cache TTL is configurable via environment variable (default: 30 minutes).

### 4.5 Background Refresh

- An in-process goroutine runs on a ticker (configurable interval, default: 30 minutes).
- On each tick, the service queries all households with locations set (cross-tenant — this is a system-level background job) and refreshes their weather cache.
- Stale entries (older than TTL) are prioritized, but all entries are refreshed per cycle.
- The ticker respects Open-Meteo rate limits by spacing requests (e.g., 1 request per second).
- Errors during refresh are logged but do not crash the service.

### 4.6 Cache Invalidation on Location Change

- When a household's location (latitude/longitude) is updated or cleared via the PATCH endpoint, the weather cache for that household must be invalidated.
- If the location is cleared (set to null), the cache row is deleted.
- If the location is changed to new coordinates, the existing cache row is deleted. The next weather API call for that household triggers a fresh synchronous fetch.

### 4.7 Weather API Endpoints

- `GET /api/v1/weather/current` — Returns current conditions and today's forecast for the active household.
- `GET /api/v1/weather/forecast` — Returns the 7-day forecast for the active household.
- `GET /api/v1/weather/geocoding?q={query}` — Returns geocoding search results.
- All endpoints except geocoding require a valid household context (tenant + household scoped via JWT).
- If the active household has no location set, weather endpoints return 404 with a descriptive error.

### 4.8 Dashboard Widget

- A weather card on the dashboard displays:
  - Current temperature (e.g., "72°F" or "22°C")
  - Today's high and low (e.g., "H: 78° L: 55°")
  - Weather condition summary with appropriate icon (e.g., sun, cloud, rain, snow)
- If no location is set, the widget displays a prompt to set location in household settings.
- The widget links to the full weather page.

### 4.9 Weather Page

- A dedicated `/weather` page displays the 7-day forecast.
- Each day shows:
  - Day of week
  - Weather condition icon and summary text
  - High temperature
  - Low temperature
- Today is visually distinguished from future days.
- If no location is set, the page displays a message with a link to household settings.

### 4.10 Household Settings — Location Input

- The household settings/edit form includes a location field.
- The field is an autocomplete search: the user types a city/place name, sees matching results, and selects one.
- On selection, latitude and longitude are populated and saved with the household.
- The current location (if set) is displayed as the place name. The place name is stored alongside coordinates for display purposes.

## 5. API Surface

### 5.1 Account Service Changes

#### PATCH /api/v1/households/{id} (modified)

New optional attributes in request body:

| Attribute     | Type    | Required |
|---------------|---------|----------|
| latitude      | float64 | no       |
| longitude     | float64 | no       |
| locationName  | string  | no       |

Validation: if either `latitude` or `longitude` is provided, both must be present. `latitude` must be between -90 and 90. `longitude` must be between -180 and 180.

Response includes the new fields in the `households` resource.

### 5.2 Weather Service Endpoints

#### GET /api/v1/weather/current

**Response:** JSON:API `weather-current` resource with ID matching household ID.

| Attribute       | Type   | Description                        |
|-----------------|--------|------------------------------------|
| temperature     | float64| Current temperature                |
| temperatureUnit | string | "°C" or "°F"                       |
| summary         | string | Human-readable condition           |
| icon            | string | Icon key for frontend rendering    |
| weatherCode     | int    | WMO weather code                   |
| highTemperature | float64| Today's forecasted high            |
| lowTemperature  | float64| Today's forecasted low             |
| fetchedAt       | string | ISO 8601 timestamp of last fetch   |

**Error Conditions:**

| Status | Condition                     |
|--------|-------------------------------|
| 404    | Household has no location set |
| 502    | Upstream API failure          |

#### GET /api/v1/weather/forecast

**Response:** JSON:API array of `weather-daily` resources.

| Attribute       | Type   | Description                      |
|-----------------|--------|----------------------------------|
| date            | string | ISO 8601 date (YYYY-MM-DD)       |
| highTemperature | float64| Daily high                       |
| lowTemperature  | float64| Daily low                        |
| temperatureUnit | string | "°C" or "°F"                     |
| summary         | string | Human-readable condition         |
| icon            | string | Icon key for frontend rendering  |
| weatherCode     | int    | WMO weather code                 |

**Error Conditions:**

| Status | Condition                     |
|--------|-------------------------------|
| 404    | Household has no location set |
| 502    | Upstream API failure          |

#### GET /api/v1/weather/geocoding?q={query}

**Parameters:**

| Name | In    | Type   | Required | Description                |
|------|-------|--------|----------|----------------------------|
| q    | query | string | yes      | Search term (min 2 chars)  |

**Response:** JSON:API array of `geocoding-results` resources.

| Attribute  | Type    | Description              |
|------------|---------|--------------------------|
| name       | string  | Place name               |
| country    | string  | Country name             |
| admin1     | string  | Primary admin region     |
| latitude   | float64 | Latitude                 |
| longitude  | float64 | Longitude                |

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | Missing or short query |
| 502    | Upstream API failure   |

## 6. Data Model

### 6.1 Account Service — Household Extension

Add to `account.households`:

| Column        | Type             | Constraints |
|---------------|------------------|-------------|
| latitude      | DOUBLE PRECISION | NULLABLE    |
| longitude     | DOUBLE PRECISION | NULLABLE    |
| location_name | TEXT             | NULLABLE    |

No new indexes required. Existing GORM AutoMigrate handles the column additions.

### 6.2 Weather Service — New Schema

Schema: `weather`

#### weather_caches

| Column           | Type             | Constraints |
|------------------|------------------|-------------|
| id               | UUID             | PRIMARY KEY |
| tenant_id        | UUID             | NOT NULL    |
| household_id     | UUID             | NOT NULL    |
| latitude         | DOUBLE PRECISION | NOT NULL    |
| longitude        | DOUBLE PRECISION | NOT NULL    |
| units            | TEXT             | NOT NULL    |
| current_data     | JSONB            | NOT NULL    |
| forecast_data    | JSONB            | NOT NULL    |
| fetched_at       | TIMESTAMP        | NOT NULL    |
| created_at       | TIMESTAMP        | NOT NULL    |
| updated_at       | TIMESTAMP        | NOT NULL    |

**Indexes:**

| Index Name              | Columns      | Type   |
|-------------------------|--------------|--------|
| idx_weather_household   | household_id | UNIQUE |
| idx_weather_tenant      | tenant_id    | INDEX  |

`current_data` JSONB stores: temperature, weather code.

`forecast_data` JSONB stores: array of daily entries (date, high, low, weather code).

Using JSONB allows the cache shape to evolve without schema migrations. The structured fields (tenant_id, household_id, coordinates, units, fetched_at) remain as columns for querying and indexing.

## 7. Service Impact

### 7.1 weather-service (new)

- New Go microservice following existing service code patterns
- Domains: `weathercache` (cache storage/retrieval), `geocoding` (search proxy), `forecast` (Open-Meteo client)
- Background refresh goroutine started on service boot
- Requires database connection (shared postgres, `weather` schema)
- Requires outbound HTTPS access to `api.open-meteo.com` and `geocoding-api.open-meteo.com`
- Validates JWT via JWKS (same as other services)
- Needs household data: the weather-service reads household location/units. Two options:
  - **Option A:** The weather-service calls account-service's API to list households with locations. This respects service boundaries but adds inter-service HTTP calls.
  - **Option B:** The background refresh job queries the weather cache table (which stores lat/lon/units copied at cache creation time). The cache is populated/updated when the weather endpoint is first called for a household. This avoids inter-service calls for the background job.
  - **Recommended: Option B** — the weather cache stores the coordinates and units at write time. The background job only refreshes existing cache entries. New households get their first cache entry on the first weather API call (which receives household context from the JWT + frontend request).

### 7.2 account-service

- Add `latitude`, `longitude`, `location_name` fields to household model, entity, builder, resource, and REST layer.
- Update PATCH handler to accept and validate the new fields.
- GORM AutoMigrate adds the new columns on startup.

### 7.3 frontend

- New dashboard weather widget component
- New `/weather` page with 7-day forecast
- Geocoding autocomplete component for household settings
- Update household settings form to include location field
- New API client functions for weather endpoints

### 7.4 Infrastructure

- New Dockerfile for weather-service
- New docker-compose entry
- New nginx route: `/api/v1/weather -> weather-service`
- New k8s manifest
- New CI workflow rules
- New build script (`scripts/build-weather.sh`)

## 8. Non-Functional Requirements

### Performance
- Weather API responses should serve from cache (sub-100ms typical).
- Geocoding search should respond within 500ms (proxied to Open-Meteo).
- Background refresh should handle 100+ households without exceeding Open-Meteo rate limits (rate-limited to ~1 req/sec).

### Security
- All weather endpoints require JWT authentication.
- Weather data is tenant-scoped and household-scoped.
- The geocoding endpoint requires authentication but not household context.
- No API keys are stored for Open-Meteo (it's keyless).

### Observability
- Standard structured logging (Logrus) with request_id, user_id, tenant_id, household_id.
- OpenTelemetry tracing on all endpoints and the background refresh job.
- Log warnings on upstream API failures with response details.

### Multi-Tenancy
- Weather cache entries include `tenant_id` and `household_id`.
- All queries are tenant-scoped via the standard tenant middleware.
- Background refresh operates cross-tenant (system job) but cache reads are scoped.

### Reliability
- If Open-Meteo is unavailable, stale cached data is served with the `fetchedAt` timestamp so the frontend can indicate staleness.
- The background refresh logs errors and continues to the next household on failure.
- The service starts and operates normally even if Open-Meteo is unreachable (endpoints return 502 for uncached households).

## 9. Open Questions

1. **Place name persistence:** Should the `location_name` be re-resolved from coordinates, or is storing it at selection time sufficient? (Current design: store at selection time.)
2. **Timezone for daily forecasts:** Open-Meteo accepts a timezone parameter. Should we use the household's `timezone` field? (Suggested: yes, so "today" aligns with the household's local time.)

## 10. Acceptance Criteria

- [ ] Household model includes latitude, longitude, and location_name fields.
- [ ] PATCH households endpoint accepts and validates location fields.
- [ ] Household settings page has a working geocoding autocomplete for location.
- [ ] Weather-service is created following existing service code patterns.
- [ ] Weather cache is invalidated when a household's location is changed or cleared.
- [ ] `GET /api/v1/weather/current` returns current conditions including icon key for the active household.
- [ ] `GET /api/v1/weather/forecast` returns a 7-day forecast including icon keys for the active household.
- [ ] `GET /api/v1/weather/geocoding?q=...` returns place search results.
- [ ] Weather data is cached in the database and served from cache.
- [ ] Background ticker refreshes weather cache at a configurable interval.
- [ ] Dashboard shows a weather widget with current temp, today's high/low, and condition.
- [ ] Dashboard widget prompts to set location if none is configured.
- [ ] Weather page at `/weather` shows a 7-day forecast with daily high/low and conditions.
- [ ] Weather data respects the household's `units` field (metric/imperial).
- [ ] Weather endpoints return 404 with a clear message when no location is set.
- [ ] All weather endpoints require JWT authentication and tenant/household context.
- [ ] Weather-service has Dockerfile, compose entry, nginx route, k8s manifest, and CI rules.
- [ ] All new code has unit tests.
- [ ] All affected services build and pass tests.
