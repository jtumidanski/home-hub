# Task 006 — Weather Enhancements: Context

Last Updated: 2026-03-25

---

## Key Files

### Backend — weather-service

| File | Role | Changes Needed |
|------|------|----------------|
| `services/weather-service/internal/openmeteo/types.go` | Open-Meteo API response structs | Add `HourlyData` struct, add `Hourly` field to `ForecastResponse` |
| `services/weather-service/internal/openmeteo/client.go` | HTTP client for Open-Meteo | Add `hourly` query params to `FetchForecast()` |
| `services/weather-service/internal/openmeteo/client_test.go` | Client tests | Update mock responses with hourly data |
| `services/weather-service/internal/forecast/model.go` | Domain model | Add `HourlyForecast` struct, embed in `DailyForecast` |
| `services/weather-service/internal/forecast/builder.go` | Model builder | Ensure hourly data passes through |
| `services/weather-service/internal/forecast/processor.go` | Business logic, `transformResponse()` | Parse hourly data, group by day, attach to daily entries |
| `services/weather-service/internal/forecast/entity.go` | GORM entity, JSON marshaling | Verify hourly data round-trips through JSONB |
| `services/weather-service/internal/forecast/rest.go` | HTTP handlers, REST models | Add `HourlyRestModel`, extend `DailyRestModel` |
| `services/weather-service/internal/forecast/processor_test.go` | Processor tests | Test hourly parsing and grouping |
| `services/weather-service/internal/forecast/entity_test.go` | Entity tests | Test JSONB round-trip with hourly data |
| `services/weather-service/internal/forecast/rest_test.go` | REST tests | Test hourly REST model transformation |
| `services/weather-service/internal/weathercode/weathercode.go` | WMO code mapping | No changes — reused for hourly entries |
| `services/weather-service/internal/refresh/refresh.go` | Background refresh | No changes — automatically picks up new data |

### Frontend

| File | Role | Changes Needed |
|------|------|----------------|
| `frontend/src/types/models/weather.ts` | TypeScript types | Add `HourlyForecastEntry`, extend `WeatherDailyAttributes` |
| `frontend/src/components/features/weather/weather-widget.tsx` | Dashboard widget | Add forecast fetch, render 4 mini forecast cards |
| `frontend/src/pages/WeatherPage.tsx` | Weather page | Add expand/collapse, render hourly entries |
| `frontend/src/lib/hooks/api/use-weather.ts` | Weather hooks | No changes — `useWeatherForecast()` already exists |
| `frontend/src/components/common/weather-icon.tsx` | Icon component | No changes — all icons supported |
| `frontend/src/services/api/weather.ts` | API service | No changes — forecast endpoint unchanged |

---

## Key Decisions

1. **No database migration** — Hourly data is stored in the existing `forecast_data` JSONB column. The payload grows from ~1 KB to ~10-15 KB per household.

2. **Cache staleness for missing hourly data** — Existing cache entries without `hourlyForecast` are treated as stale and re-fetched. This is checked by inspecting whether the first daily entry has hourly data.

3. **Hourly grouping strategy** — Hourly data from Open-Meteo arrives as flat arrays (168 entries for 7 days). These are grouped by matching the date portion of each hourly timestamp against the daily date strings.

4. **Today's hourly filtering is frontend-only** — The backend returns all 24 hours for every day. The frontend filters today's hours to current hour onward using the client's local time.

5. **Dashboard makes two parallel API calls** — `/current` and `/forecast`. Both serve from cache so latency is sub-100ms.

6. **Accordion-style expand** — One day expanded at a time on the weather page to keep the page clean on mobile.

7. **Backwards-compatible API change** — Adding `hourlyForecast` to the existing `weather-daily` resource is additive. Existing consumers that ignore the field are unaffected.

---

## Dependencies Between Tasks

```
Phase 1 (Client) → Phase 2 (Domain/Cache) → Phase 3 (REST)
                                                  ↓
                                          ┌───────┴───────┐
                                          ↓               ↓
                                   Phase 4 (Dashboard)  Phase 5 (Hourly)
```

- Phases 1–3 are strictly sequential (each builds on the prior)
- Phases 4 and 5 are independent of each other (both depend on Phase 3)

---

## Open-Meteo API Reference

### Current Forecast Request (before)
```
GET https://api.open-meteo.com/v1/forecast
  ?latitude=...&longitude=...
  &current=temperature_2m,weather_code
  &daily=temperature_2m_max,temperature_2m_min,weather_code
  &temperature_unit=fahrenheit|celsius
  &timezone=...
  &forecast_days=7
```

### Extended Request (after)
```
GET https://api.open-meteo.com/v1/forecast
  ?latitude=...&longitude=...
  &current=temperature_2m,weather_code
  &daily=temperature_2m_max,temperature_2m_min,weather_code
  &hourly=temperature_2m,weather_code,precipitation_probability
  &temperature_unit=fahrenheit|celsius
  &timezone=...
  &forecast_days=7
```

### Hourly Response Shape (new)
```json
{
  "hourly": {
    "time": ["2026-03-25T00:00", "2026-03-25T01:00", ...],
    "temperature_2m": [58.0, 57.5, ...],
    "weather_code": [1, 1, ...],
    "precipitation_probability": [0, 5, ...]
  },
  "hourly_units": {
    "time": "iso8601",
    "temperature_2m": "°F",
    "weather_code": "wmo code",
    "precipitation_probability": "%"
  }
}
```

168 entries total (24 hours × 7 days).

---

## Testing Strategy

- **Unit tests**: All new domain logic (hourly parsing, grouping, REST transformation, entity round-trip)
- **Client tests**: Mock server returns hourly data in response
- **Build verification**: `go build` and `go test ./...` for weather-service, Docker build
- **Manual verification**: Dashboard preview and expandable hourly view via local docker-compose
