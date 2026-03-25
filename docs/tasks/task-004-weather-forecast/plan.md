# Weather Forecast — Implementation Plan

Last Updated: 2026-03-25

---

## Executive Summary

Add weather visibility to Home Hub: current conditions on the dashboard and a 7-day forecast page. This requires extending the household model with location fields (account-service), creating a new weather-service (Go microservice following existing patterns), and building frontend components (dashboard widget, weather page, geocoding autocomplete in household settings). Weather data is sourced from Open-Meteo (free, keyless), cached in PostgreSQL, and refreshed via an in-process background ticker.

---

## Current State Analysis

- **Household model** has `name`, `timezone`, `units` — no location fields
- **No weather service** exists; this is a greenfield service
- **Service pattern** is well-established: model → entity → builder → resource → rest → processor → provider → administrator
- **Frontend** uses React + Vite + shadcn/ui with a singleton ApiClient, JSON:API service classes, and TanStack React Query hooks
- **Infrastructure** uses docker-compose with nginx reverse proxy, path-prefix routing, and GitHub Actions CI
- **No background job pattern** exists in the codebase — the ticker will be the first

---

## Proposed Future State

- Household model includes `latitude`, `longitude`, `location_name` (nullable)
- New `weather-service` with `weather` PostgreSQL schema, serving cached weather data
- Background goroutine refreshes weather cache every 30 minutes (configurable)
- Frontend dashboard has a weather widget; new `/weather` page shows 7-day forecast
- Household settings includes geocoding autocomplete for location
- nginx routes `/api/v1/weather` to weather-service

---

## Implementation Phases

### Phase 1: Account Service — Household Location Extension

Extend the household domain in account-service to support latitude, longitude, and location_name fields. This is the foundation — weather-service and frontend both depend on location being available on the household.

**Why first:** Everything downstream (weather fetching, frontend display) depends on households having a location. This is the smallest, most contained change and can be verified independently.

### Phase 2: Weather Service — Core Backend

Create the new weather-service following existing service patterns. This includes:
- Service scaffolding (go.mod, cmd/main.go, config, Dockerfile)
- Open-Meteo client (forecast + geocoding)
- WMO weather code mapping (summary + icon)
- Weather cache domain (model, entity, CRUD)
- REST endpoints (current, forecast, geocoding)
- Background refresh ticker

**Why second:** The backend must exist before the frontend can consume it. The service is self-contained — it reads household context from JWT/headers and caches weather data independently.

### Phase 3: Infrastructure

Wire the new service into docker-compose, nginx, go.work, CI workflows, and build scripts.

**Why third:** Needed to run and test the full stack locally. Can be done in parallel with Phase 2 but logically follows service creation.

### Phase 4: Frontend — Weather Features

Build the frontend components:
- Weather API service class
- React Query hooks for weather data
- Dashboard weather widget
- Weather page with 7-day forecast
- Geocoding autocomplete in household settings
- Sidebar navigation entry

**Why last:** Depends on both the account-service changes (location fields) and weather-service (API endpoints) being in place.

---

## Detailed Tasks

### Phase 1: Account Service — Household Location Extension

#### 1.1 Extend Household Model (S)
- Add `latitude *float64`, `longitude *float64`, `locationName *string` private fields to `Model`
- Add getter methods: `Latitude()`, `Longitude()`, `LocationName()`
- Add `HasLocation() bool` convenience method
- **File:** `services/account-service/internal/household/model.go`
- **Acceptance:** Model compiles, getters return correct values, nil-safe

#### 1.2 Extend Household Entity (S)
- Add `Latitude *float64`, `Longitude *float64`, `LocationName *string` to `Entity` with GORM tags (nullable)
- Update `Make()` to map new fields from entity to model
- Migration adds columns on startup via AutoMigrate
- **File:** `services/account-service/internal/household/entity.go`
- **Acceptance:** GORM AutoMigrate creates nullable columns, Make() maps correctly

#### 1.3 Extend Household Builder (S)
- Add `SetLatitude(*float64)`, `SetLongitude(*float64)`, `SetLocationName(*string)` methods
- Add validation in `Build()`: if either lat or lon is set, both must be set; lat must be -90..90; lon must be -180..180
- **File:** `services/account-service/internal/household/builder.go`
- **Acceptance:** Partial coordinates rejected, out-of-range rejected, nil/nil accepted, valid pair accepted

#### 1.4 Extend Household Resource (S)
- Add `Latitude *float64`, `Longitude *float64`, `LocationName *string` to `RestModel` with JSON tags
- Add same fields to `UpdateRequest`
- Update `Transform()` to map new fields
- **File:** `services/account-service/internal/household/resource.go`
- **Acceptance:** JSON:API responses include new fields, update requests accept them

#### 1.5 Extend Household REST / Administrator (S)
- Update `updateHandler` to pass new fields through to processor
- Update `administrator.go` update function to persist new fields
- **File:** `services/account-service/internal/household/rest.go`, `administrator.go`
- **Acceptance:** PATCH `/api/v1/households/{id}` accepts and persists lat/lon/locationName

#### 1.6 Unit Tests for Location Fields (M)
- Test builder validation (partial coords, out-of-range, valid, nil)
- Test entity Make() with nullable fields
- Test resource Transform() includes new fields
- **Acceptance:** All new validation paths covered

#### 1.7 Build & Verify Account Service (S)
- Run `go build ./...` and `go test ./...` for account-service
- **Acceptance:** Compiles and all tests pass

---

### Phase 2: Weather Service — Core Backend

#### 2.1 Service Scaffolding (M)
- Create `services/weather-service/` directory structure:
  ```
  cmd/main.go
  internal/config/config.go
  internal/openmeteo/      (API client)
  internal/weathercode/    (WMO mapping)
  internal/forecast/       (cache domain)
  internal/geocoding/      (search proxy)
  internal/refresh/        (background ticker)
  go.mod
  Dockerfile
  ```
- `go.mod` with module path `github.com/jtumidanski/home-hub/services/weather-service`
- Import shared modules: server, auth, database, logging, tenant, model
- `config.go` with DB, Port, JWKSURL, RefreshInterval, CacheTTL env vars
- `cmd/main.go` following existing startup pattern (logging, tracing, DB, auth, routes, server.Run)
- **Acceptance:** Service compiles and starts (no routes yet)

#### 2.2 Open-Meteo Forecast Client (M)
- Create `internal/openmeteo/client.go` with HTTP client
- `FetchForecast(lat, lon, units, timezone string) (*ForecastResponse, error)` — calls Open-Meteo Forecast API
- `SearchPlaces(query string) ([]Place, error)` — calls Open-Meteo Geocoding API
- Response structs matching Open-Meteo JSON
- Rate limiting: 1 req/sec via a simple time-based throttle
- **File:** `services/weather-service/internal/openmeteo/`
- **Acceptance:** Can fetch weather for known coordinates, can search for "New York"

#### 2.3 WMO Weather Code Mapping (S)
- Create `internal/weathercode/weathercode.go`
- `Lookup(code int) (summary string, icon string)` — maps WMO code to human summary + Lucide icon key
- Full mapping table per api-contracts.md
- **Acceptance:** All WMO codes from the mapping table return correct summary and icon

#### 2.4 Forecast Cache Domain — Model & Entity (M)
- **model.go:** `forecast.Model` with fields: id, tenantID, householdID, latitude, longitude, units, currentData (struct), forecastData ([]DailyForecast), fetchedAt
- `CurrentData` struct: temperature, weatherCode, summary, icon
- `DailyForecast` struct: date, highTemp, lowTemp, weatherCode, summary, icon
- **entity.go:** GORM entity with JSONB columns for current_data and forecast_data, Migration() function
- **builder.go:** Fluent builder with validation
- Schema: `weather`
- **Files:** `services/weather-service/internal/forecast/`
- **Acceptance:** Entity compiles, Migration() creates table, Make() converts correctly

#### 2.5 Forecast Cache — Provider & Administrator (M)
- **provider.go:** `getByHouseholdID(householdID)` query
- **administrator.go:** `upsert(db, ...)` — insert or update cache by household_id; `deleteByHouseholdID(db, householdID)`
- **processor.go:** `GetCurrent(householdID)`, `GetForecast(householdID)`, `RefreshCache(householdID, lat, lon, units, tz)`, `InvalidateCache(householdID)`
- Processor calls Open-Meteo client, maps WMO codes, persists to cache
- On cache miss, triggers synchronous fetch
- **Files:** `services/weather-service/internal/forecast/`
- **Acceptance:** Cache is populated on first request, served from DB on subsequent requests

#### 2.6 Forecast REST Endpoints (M)
- **resource.go:** `CurrentRestModel`, `DailyRestModel` with JSON:API fields (including `icon`)
- **rest.go:** `InitializeRoutes(db, openMeteoClient)` registering:
  - `GET /weather/current` → currentHandler
  - `GET /weather/forecast` → forecastHandler
- Handlers extract tenant_id + household_id from context, call processor, return JSON:API response
- 404 if household has no cached location, 502 if upstream fails and no cache
- Temperature unit label ("°C" or "°F") derived from cached `units` field
- **Files:** `services/weather-service/internal/forecast/`
- **Acceptance:** Endpoints return correct JSON:API, 404 on no location, stale cache served with fetchedAt

#### 2.7 Geocoding REST Endpoint (M)
- **resource.go:** `GeocodingRestModel` with name, country, admin1, latitude, longitude
- **rest.go:** `InitializeRoutes(openMeteoClient)` registering:
  - `GET /weather/geocoding?q={query}` → searchHandler
- Validates query length >= 2 chars
- Calls Open-Meteo geocoding client, transforms results to JSON:API
- **Files:** `services/weather-service/internal/geocoding/`
- **Acceptance:** Returns place results for valid queries, 400 for short queries

#### 2.8 Background Refresh Ticker (M)
- Create `internal/refresh/refresh.go`
- `StartRefreshLoop(ctx context.Context, db *gorm.DB, client *openmeteo.Client, interval time.Duration, logger logrus.FieldLogger)`
- Runs on a `time.Ticker` at configurable interval
- Queries all weather_cache rows, re-fetches from Open-Meteo using stored lat/lon/units
- Spaces requests at ~1/sec to respect rate limits
- Logs errors per-household but continues processing
- Stops cleanly on context cancellation (for graceful shutdown)
- Launched from `cmd/main.go` as a goroutine before `server.Run()`
- **Acceptance:** Cache rows are refreshed periodically, errors don't crash the loop

#### 2.9 Wire Up main.go (S)
- Connect all pieces in `cmd/main.go`: config, DB, Open-Meteo client, auth validator, routes, refresh goroutine, server
- Pass context with cancel for graceful shutdown of ticker
- **Acceptance:** Service starts, all endpoints registered, ticker running

#### 2.10 Unit Tests (L)
- Weather code mapping: all codes
- Open-Meteo client: mock HTTP responses
- Cache processor: mock DB + client, test cache hit/miss/refresh/invalidate
- REST handlers: mock processor, test response shapes and error cases
- Geocoding handler: mock client, test response shapes
- **Acceptance:** Coverage on all non-trivial logic paths

#### 2.11 Build & Verify Weather Service (S)
- `go build ./...` and `go test ./...`
- **Acceptance:** Compiles and all tests pass

---

### Phase 3: Infrastructure

#### 3.1 Add to go.work (S)
- Add `./services/weather-service` to `go.work`
- Run `go work sync`
- **Acceptance:** `go work sync` succeeds

#### 3.2 Dockerfile (S)
- Create `services/weather-service/Dockerfile` following existing pattern
- Copy shared modules + weather-service, build, minimal alpine runtime image
- **Acceptance:** `docker build` succeeds

#### 3.3 Docker Compose Entry (S)
- Add `weather-service` to `deploy/compose/docker-compose.yml`
- Same pattern as other services: build context, expose 8080, DB + JWKS env vars
- Add to nginx depends_on
- **Acceptance:** `docker-compose up` starts weather-service

#### 3.4 Nginx Route (S)
- Add upstream `weather-service` and `location /api/v1/weather` to `deploy/compose/nginx.conf`
- **Acceptance:** Requests to `/api/v1/weather/*` proxy to weather-service

#### 3.5 Build Script (S)
- Create `scripts/build-weather.sh`
- Add weather-service to `scripts/build-all.sh`
- **Acceptance:** `./scripts/build-weather.sh` succeeds

#### 3.6 CI Workflows (M)
- Add weather-service detection and build/test/lint/docker jobs to `.github/workflows/pr.yml`
- Add weather-service build and push to `.github/workflows/main.yml`
- **Acceptance:** CI detects weather-service changes and runs appropriate jobs

#### 3.7 K8s Manifest (S)
- Create `deploy/k8s/weather-service.yaml` (Deployment + Service)
- Update ingress for `/api/v1/weather` path
- **Acceptance:** Manifest is valid YAML with correct selectors and env vars

#### 3.8 Update Architecture Docs (S)
- Add weather-service to `docs/architecture.md` routing table and service list
- **Acceptance:** Docs reflect the new service

---

### Phase 4: Frontend — Weather Features

#### 4.1 Weather API Service Class (S)
- Create `frontend/src/services/api/weather.ts`
- Methods: `getCurrent()`, `getForecast()`, `searchPlaces(query)`
- Uses existing `api` singleton with tenant headers
- Export `weatherService` instance
- **Acceptance:** Service class compiles, methods match API contract

#### 4.2 React Query Hooks (S)
- Create `frontend/src/lib/hooks/api/use-weather.ts`
- `useCurrentWeather()` — queries `/weather/current`, enabled when household has location
- `useWeatherForecast()` — queries `/weather/forecast`
- `useGeocodingSearch(query)` — queries `/weather/geocoding`, debounced, enabled when query >= 2 chars
- **Acceptance:** Hooks return typed data, handle loading/error states

#### 4.3 Weather Icon Component (S)
- Create `frontend/src/components/common/weather-icon.tsx`
- Maps `icon` key string (from API) to Lucide React component
- Lookup: sun → Sun, cloud-sun → CloudSun, cloud → Cloud, etc.
- Fallback to Cloud icon for unknown keys
- **Acceptance:** Renders correct Lucide icon for each key

#### 4.4 Dashboard Weather Widget (M)
- Create `frontend/src/components/features/weather/weather-widget.tsx`
- States: loading (skeleton), no-location (CTA to settings), data (current temp + H/L + icon + location name + absolute timestamp), error/stale (show data with warning)
- Card-based layout per user's mobile UI preferences
- Tappable/clickable → navigates to `/weather`
- Add widget to DashboardPage
- **Acceptance:** Widget renders all states correctly, navigates to weather page

#### 4.5 Weather Page (M)
- Create `frontend/src/pages/WeatherPage.tsx`
- 7-day forecast: card per day with day name, date, icon, summary, H/L
- Today visually distinguished
- No-location state with link to settings
- Mobile: full-width stacked cards, tap-friendly
- **Acceptance:** Page renders 7 days, today highlighted, responsive layout

#### 4.6 Geocoding Autocomplete Component (M)
- Create `frontend/src/components/features/weather/location-search.tsx`
- Autocomplete input with debounced search (300ms)
- Dropdown shows: name, admin region, country
- On select: populates lat/lon/locationName
- Clear button to remove location
- Keyboard navigable (up/down/enter)
- **Acceptance:** Search returns results, selection populates fields, clear works

#### 4.7 Integrate Location into Household Settings (S)
- Add `LocationSearch` component to household create/edit forms
- Wire to PATCH `/api/v1/households/{id}` with latitude, longitude, locationName
- Show current location if set
- **Acceptance:** Location can be set and updated from household settings

#### 4.8 Add Weather Route & Navigation (S)
- Add `<Route path="weather" element={<WeatherPage />} />` to App.tsx under `/app`
- Add "Weather" item to sidebar navigation
- **Acceptance:** `/app/weather` renders WeatherPage, sidebar links to it

#### 4.9 Frontend Tests (M)
- Weather widget: render states (loading, no-location, data, stale)
- Weather page: render 7 days, today highlighting
- Weather icon: correct icon per key
- Geocoding autocomplete: debounce, select, clear
- **Acceptance:** All component tests pass

#### 4.10 Frontend Build & Verify (S)
- `npm run build` and `npm run test`
- **Acceptance:** Compiles with no type errors, all tests pass

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Open-Meteo rate limiting | Medium | Medium | 1 req/sec throttle in client, cache TTL prevents excessive fetching |
| Open-Meteo downtime | Low | Medium | Serve stale cache, 502 only when no cache exists |
| Background ticker memory leak | Low | High | Context cancellation on shutdown, no goroutine leaks |
| JSONB schema drift | Low | Low | Structured Go types for JSONB marshal/unmarshal |
| Geocoding returns irrelevant results | Low | Low | Limit to 10 results, display country/region for disambiguation |
| Household location set without timezone match | Medium | Low | Out of scope — users manage timezone independently |
| Cross-tenant data leak via background job | Low | High | Background job reads from weather_cache only; API endpoints enforce tenant scope via middleware |

---

## Success Metrics

- Dashboard widget displays current temperature for a household with a location set
- 7-day forecast page renders correctly with icons and temperatures
- Geocoding autocomplete finds and selects a city successfully
- Weather cache is populated and refreshed on schedule
- Cache invalidation triggers correctly when location changes
- All endpoints enforce tenant/household scoping
- No Open-Meteo rate limit errors in normal operation
- All services build and tests pass in CI

---

## Required Resources and Dependencies

| Resource | Notes |
|----------|-------|
| Open-Meteo Forecast API | `api.open-meteo.com/v1/forecast` — free, no key |
| Open-Meteo Geocoding API | `geocoding-api.open-meteo.com/v1/search` — free, no key |
| PostgreSQL | Existing instance, new `weather` schema |
| Shared Go modules | server, auth, database, logging, tenant, model |
| Lucide React icons | Already available via shadcn — Sun, CloudSun, Cloud, CloudFog, CloudDrizzle, CloudRain, Snowflake, CloudLightning |

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|-------------|
| Phase 1: Account Service | S–M | None |
| Phase 2: Weather Service | L | Phase 1 (for household model understanding, not blocking) |
| Phase 3: Infrastructure | S–M | Phase 2 (Dockerfile needs service code) |
| Phase 4: Frontend | L | Phase 1 + Phase 2 + Phase 3 |

Phases 1 and 2 can be developed in parallel. Phase 3 can begin once Phase 2 has the Dockerfile. Phase 4 requires all backend work complete.
