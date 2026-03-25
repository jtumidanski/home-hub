# Weather Forecast — Context

Last Updated: 2026-03-25

---

## Key Files

### Account Service — Household Domain (to be modified)

| File | Purpose |
|------|---------|
| `services/account-service/internal/household/model.go` | Immutable domain model — add latitude, longitude, locationName |
| `services/account-service/internal/household/entity.go` | GORM entity + Migration — add nullable columns |
| `services/account-service/internal/household/builder.go` | Fluent builder — add setters + coordinate validation |
| `services/account-service/internal/household/resource.go` | JSON:API types — add fields to RestModel + UpdateRequest |
| `services/account-service/internal/household/rest.go` | Route handlers — update updateHandler for new fields |
| `services/account-service/internal/household/administrator.go` | Raw CRUD — update to persist new fields |
| `services/account-service/internal/household/processor.go` | Service logic — pass through new fields |
| `services/account-service/internal/household/provider.go` | Data access (no changes expected) |

### Weather Service — New Files (to be created)

| File | Purpose |
|------|---------|
| `services/weather-service/cmd/main.go` | Service entry point |
| `services/weather-service/internal/config/config.go` | Environment config |
| `services/weather-service/internal/openmeteo/client.go` | Open-Meteo HTTP client |
| `services/weather-service/internal/openmeteo/types.go` | API response structs |
| `services/weather-service/internal/weathercode/weathercode.go` | WMO code → summary + icon mapping |
| `services/weather-service/internal/forecast/model.go` | Cache domain model |
| `services/weather-service/internal/forecast/entity.go` | GORM entity (JSONB columns) |
| `services/weather-service/internal/forecast/builder.go` | Cache builder |
| `services/weather-service/internal/forecast/resource.go` | CurrentRestModel, DailyRestModel |
| `services/weather-service/internal/forecast/rest.go` | /weather/current, /weather/forecast handlers |
| `services/weather-service/internal/forecast/processor.go` | Cache read/write + Open-Meteo fetch |
| `services/weather-service/internal/forecast/provider.go` | DB queries |
| `services/weather-service/internal/forecast/administrator.go` | Upsert/delete cache |
| `services/weather-service/internal/geocoding/resource.go` | GeocodingRestModel |
| `services/weather-service/internal/geocoding/rest.go` | /weather/geocoding handler |
| `services/weather-service/internal/refresh/refresh.go` | Background ticker goroutine |
| `services/weather-service/go.mod` | Go module definition |
| `services/weather-service/Dockerfile` | Docker build |

### Frontend — New Files (to be created)

| File | Purpose |
|------|---------|
| `frontend/src/services/api/weather.ts` | Weather API service class |
| `frontend/src/lib/hooks/api/use-weather.ts` | React Query hooks |
| `frontend/src/components/common/weather-icon.tsx` | Icon key → Lucide component |
| `frontend/src/components/features/weather/weather-widget.tsx` | Dashboard widget |
| `frontend/src/components/features/weather/location-search.tsx` | Geocoding autocomplete |
| `frontend/src/pages/WeatherPage.tsx` | 7-day forecast page |

### Frontend — Existing Files (to be modified)

| File | Purpose |
|------|---------|
| `frontend/src/App.tsx` | Add /weather route |
| `frontend/src/services/api/index.ts` | Export weatherService |
| `frontend/src/pages/DashboardPage.tsx` | Add weather widget |
| Sidebar component (TBD) | Add "Weather" nav item |
| Household settings form (TBD) | Add location input |

### Infrastructure (to be created/modified)

| File | Purpose |
|------|---------|
| `deploy/compose/docker-compose.yml` | Add weather-service entry |
| `deploy/compose/nginx.conf` | Add /api/v1/weather upstream + location |
| `deploy/k8s/weather-service.yaml` | K8s deployment + service |
| `scripts/build-weather.sh` | Build script |
| `scripts/build-all.sh` | Add weather-service call |
| `.github/workflows/pr.yml` | Add weather-service CI detection |
| `.github/workflows/main.yml` | Add weather-service image push |
| `go.work` | Add weather-service module |
| `docs/architecture.md` | Add weather-service to routing + service list |

---

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| New weather-service (not extending existing service) | Follows strict service boundary architecture |
| Location on household model (not in weather-service) | Location is a household property; keeps account-service as source of truth |
| Household `units` field for temperature (no per-user override) | Simplifies v1; household-level units already exist |
| Open-Meteo (not OpenWeatherMap or NWS) | Free, keyless, global coverage, no rate-limit key management |
| JSONB for weather cache data | Allows cache shape to evolve without migrations; structured fields for querying |
| Option B for background refresh | Cache stores lat/lon/units; background job refreshes existing entries only; avoids inter-service HTTP calls |
| Self-healing cache invalidation | Compare cached lat/lon/units vs current on API call; no cross-service events needed |
| Backend owns WMO → icon mapping | Single source of truth; frontend just renders by icon key name |
| In-process ticker (not external cron) | Simplest approach; fits stateless service model; ticker per instance is acceptable |
| Absolute timestamps for "last updated" | User preference; shows exact time like "2:30 PM" |

---

## Dependencies

### Service Dependencies

```
weather-service → shared/go/server (HTTP server, handlers, JSON:API responses)
weather-service → shared/go/auth (JWT validation via JWKS)
weather-service → shared/go/database (GORM connection, tenant callbacks)
weather-service → shared/go/logging (Logrus structured logging)
weather-service → shared/go/tenant (Tenant context extraction)
weather-service → shared/go/model (Provider type, Map utilities)
weather-service → api.open-meteo.com (external, forecast + geocoding)
```

### Task Dependencies

```
Phase 1 (account-service) ──┐
                             ├──→ Phase 4 (frontend)
Phase 2 (weather-service) ──┤
                             │
Phase 3 (infrastructure) ───┘
         ↑
    Depends on Phase 2 (needs Dockerfile)
```

Phases 1 and 2 can be developed in parallel.

---

## External API Reference

### Open-Meteo Forecast API

```
GET https://api.open-meteo.com/v1/forecast
  ?latitude=40.71
  &longitude=-74.01
  &current=temperature_2m,weather_code
  &daily=temperature_2m_max,temperature_2m_min,weather_code
  &temperature_unit=fahrenheit  (or celsius)
  &timezone=America/New_York
  &forecast_days=7
```

### Open-Meteo Geocoding API

```
GET https://geocoding-api.open-meteo.com/v1/search
  ?name=New+York
  &count=10
  &language=en
```

---

## Patterns to Follow

### Go Service Pattern

Each domain in each service uses these files:
- `model.go` — immutable struct, private fields, getter methods
- `entity.go` — GORM struct, `Migration()`, `Make()` conversion
- `builder.go` — fluent builder, validation in `Build()`
- `processor.go` — business logic, depends on logger + context + db
- `provider.go` — database queries using `database.EntityProvider`
- `administrator.go` — raw CRUD (create/update/delete)
- `resource.go` — JSON:API `RestModel`, `CreateRequest`, `UpdateRequest`, `Transform()`
- `rest.go` — `InitializeRoutes()`, handler functions using `server.RegisterHandler`/`server.RegisterInputHandler`

### Frontend API Pattern

- Service class in `frontend/src/services/api/` using `api` singleton
- React Query hooks in `frontend/src/lib/hooks/api/`
- Components in `frontend/src/components/features/<domain>/`
- Pages in `frontend/src/pages/`
