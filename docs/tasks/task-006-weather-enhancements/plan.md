# Task 006 — Weather Enhancements: Implementation Plan

Last Updated: 2026-03-25

---

## Executive Summary

This plan covers two enhancements to the weather feature: (1) upgrading the dashboard weather widget to show a 4-day forecast preview, and (2) adding expandable hourly breakdowns to each day on the weather page. Both require extending the weather-service to fetch hourly data from Open-Meteo and include it in the forecast endpoint response.

The work spans three layers — Open-Meteo client, weather-service domain/REST, and frontend UI — and is structured as four sequential phases.

---

## Current State Analysis

### Backend (weather-service)
- **Open-Meteo client** (`internal/openmeteo/client.go`): Fetches `current` + `daily` variables only. No hourly data.
- **Domain model** (`internal/forecast/model.go`): `DailyForecast` struct has date, high/low temp, weather code, summary, icon. No hourly data.
- **Entity/cache** (`internal/forecast/entity.go`): `JSONForecastData` is `[]DailyForecast` stored as JSONB. No hourly nesting.
- **Processor** (`internal/forecast/processor.go`): `transformResponse()` parses daily arrays only.
- **REST layer** (`internal/forecast/rest.go`): `DailyRestModel` maps 1:1 to `DailyForecast`. No hourly field.
- **Weather code mapping** (`internal/weathercode/weathercode.go`): Already maps all WMO codes — reusable for hourly entries.
- **Background refresh** (`internal/refresh/refresh.go`): Calls `RefreshCache()` which re-fetches full forecast. Will automatically pick up hourly data once the client is extended.

### Frontend
- **Dashboard widget** (`components/features/weather/weather-widget.tsx`): Shows current conditions only. Does not fetch forecast.
- **Weather page** (`pages/WeatherPage.tsx`): Renders 7 flat day cards. No expand/collapse. No hourly data.
- **Types** (`types/models/weather.ts`): `WeatherDailyAttributes` has no `hourlyForecast` field.
- **Hooks** (`lib/hooks/api/use-weather.ts`): `useWeatherForecast()` exists and returns `WeatherDaily[]`.
- **WeatherIcon** (`components/common/weather-icon.tsx`): Supports all needed icon keys already.

---

## Proposed Future State

### Backend
- Open-Meteo client requests hourly variables: `temperature_2m`, `weather_code`, `precipitation_probability`.
- New `HourlyForecast` struct added to domain model.
- `DailyForecast` gains `HourlyForecast []HourlyForecast` field.
- `transformResponse()` parses hourly arrays, groups by day, attaches to each `DailyForecast`.
- `DailyRestModel` gains `HourlyForecast` field serialized in JSON:API response.
- Cache JSONB naturally grows — no migration needed. Entries missing hourly data treated as stale.

### Frontend
- Dashboard widget fetches forecast in addition to current, renders 4 mini forecast cards.
- Weather page day cards become expandable. Expanded view shows hourly entries.
- `WeatherDailyAttributes` type gains `hourlyForecast` array.
- New `HourlyForecastEntry` type defined.

---

## Implementation Phases

### Phase 1: Backend — Open-Meteo Client Extension
**Goal:** Fetch hourly data from Open-Meteo alongside daily data.

1. **Extend Open-Meteo types** (`internal/openmeteo/types.go`)
   - Add `Hourly` and `HourlyUnits` structs to `ForecastResponse`
   - `HourlyData`: `Time []string`, `Temperature []float64`, `WeatherCode []int`, `PrecipitationProbability []int`
   - Effort: **S**

2. **Extend `FetchForecast()` request** (`internal/openmeteo/client.go`)
   - Add `hourly=temperature_2m,weather_code,precipitation_probability` query parameter
   - Effort: **S**

3. **Update client tests** (`internal/openmeteo/client_test.go`)
   - Update mock server responses to include hourly data
   - Verify hourly fields are parsed correctly
   - Effort: **S**

### Phase 2: Backend — Domain Model, Processor, Cache
**Goal:** Parse hourly data into domain models and persist in cache.

4. **Add `HourlyForecast` struct** (`internal/forecast/model.go`)
   - Fields: `Time string`, `Temperature float64`, `WeatherCode int`, `Summary string`, `Icon string`, `PrecipitationProbability int`
   - Add `HourlyForecast []HourlyForecast` field to `DailyForecast`
   - Effort: **S**

5. **Update builder** (`internal/forecast/builder.go`)
   - Ensure `DailyForecast` with hourly data flows through builder correctly
   - Effort: **S**

6. **Update `transformResponse()`** (`internal/forecast/processor.go`)
   - Parse hourly arrays from `openmeteo.ForecastResponse`
   - Group hourly entries by date (match against daily date strings)
   - Map each hourly weather code to summary + icon via `weathercode.Lookup()`
   - Attach grouped hourly arrays to each `DailyForecast`
   - Effort: **M**

7. **Add cache staleness check** (`internal/forecast/processor.go`)
   - In `GetForecast()` / `GetCurrent()`, treat cache entries where `forecastData[0].HourlyForecast` is nil/empty as stale (triggers re-fetch)
   - Effort: **S**

8. **Update entity JSON handling** (`internal/forecast/entity.go`)
   - `JSONForecastData` marshaling already uses `[]DailyForecast` — the new `HourlyForecast` field will serialize automatically since it's a struct field
   - Verify round-trip: entity → model → entity preserves hourly data
   - Effort: **S**

9. **Update processor and entity tests**
   - Test `transformResponse()` with hourly data
   - Test entity round-trip with hourly data
   - Test cache staleness detection for entries without hourly data
   - Effort: **M**

### Phase 3: Backend — REST Layer
**Goal:** Expose hourly data in the forecast endpoint response.

10. **Add hourly REST model** (`internal/forecast/rest.go`)
    - New `HourlyRestModel` struct: `Time`, `Temperature`, `WeatherCode`, `Summary`, `Icon`, `PrecipitationProbability`
    - Add `HourlyForecast []HourlyRestModel` to `DailyRestModel`
    - Update `TransformForecast()` to map hourly domain models to REST models
    - Effort: **S**

11. **Update REST tests** (`internal/forecast/rest_test.go`)
    - Test `TransformForecast()` includes hourly data
    - Verify JSON:API serialization includes `hourlyForecast` attribute
    - Effort: **S**

12. **Build and test full service**
    - Run `go build` and `go test ./...` for weather-service
    - Verify Docker build succeeds
    - Effort: **S**

### Phase 4: Frontend — Dashboard Forecast Preview
**Goal:** Add 4-day mini forecast cards to the dashboard weather widget.

13. **Update TypeScript types** (`types/models/weather.ts`)
    - Add `HourlyForecastEntry` interface: `time`, `temperature`, `weatherCode`, `summary`, `icon`, `precipitationProbability`
    - Add `hourlyForecast: HourlyForecastEntry[]` to `WeatherDailyAttributes`
    - Effort: **S**

14. **Update dashboard weather widget** (`components/features/weather/weather-widget.tsx`)
    - Fetch forecast data using `useWeatherForecast()` hook alongside existing `useCurrentWeather()`
    - Add a row of 4 mini forecast cards below current conditions
    - Each card: abbreviated day name (or "Today"), weather icon, high temp, low temp
    - Handle forecast fetch failure gracefully (hide preview row, keep current conditions)
    - Effort: **M**

### Phase 5: Frontend — Expandable Hourly Breakdown
**Goal:** Make weather page day cards expandable with hourly detail.

15. **Update weather page** (`pages/WeatherPage.tsx`)
    - Add expand/collapse state for day cards (local state, accordion-style — one at a time)
    - On tap, toggle expanded state for that day
    - When expanded, render hourly entries below the day summary
    - Each hourly entry: time (formatted to household timezone), weather icon, temperature, precipitation probability
    - For today: filter to current hour onward (inclusive)
    - For future days: show all 24 hours
    - Effort: **L**

16. **Verify mobile UX**
    - Mini forecast cards legible at 320px viewport
    - Hourly breakdown scrollable within expanded card
    - Tap-only interaction (no swipe gestures)
    - Effort: **S**

---

## Risk Assessment and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Open-Meteo hourly response format mismatch | Blocks Phase 1 | Low | Verify against live API response before coding; Open-Meteo docs are well-maintained |
| Forecast payload size increase (~10-15 KB) | Performance concern | Low | Acceptable per PRD analysis; monitor response times |
| Cache entries without hourly data persist | Stale UI data | Medium | Explicit staleness check treats missing hourly data as cache miss |
| Hourly grouping misaligns with daily dates | Incorrect data display | Low | Use string date prefix matching; test with timezone-crossing data |
| Dashboard making two API calls | Perceived slowness | Low | Both calls serve from cache (sub-100ms); fire in parallel |

---

## Success Metrics

- Dashboard widget shows 4-day preview with correct day names, icons, and temperatures
- Weather page days expand/collapse on tap with smooth interaction
- Expanded hourly view shows correct time, icon, temperature, and precipitation
- Today's hourly view shows only remaining hours
- All backend tests pass including new hourly data tests
- All services build successfully (Go + Docker)
- No regressions to existing current weather or forecast functionality

---

## Required Resources and Dependencies

### Internal Dependencies
- `weathercode.Lookup()` — already exists, reusable for hourly entries
- `WeatherIcon` component — already supports all needed icon keys
- `useWeatherForecast()` hook — already exists, returns from cache
- Household timezone — already available in household context

### External Dependencies
- Open-Meteo Forecast API — hourly variables: `temperature_2m`, `weather_code`, `precipitation_probability`
- No new API keys or services required

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|--------------|
| Phase 1: Client Extension | S | None |
| Phase 2: Domain & Cache | M | Phase 1 |
| Phase 3: REST Layer | S | Phase 2 |
| Phase 4: Dashboard Preview | M | Phase 3 |
| Phase 5: Hourly Breakdown | L | Phase 3 |

Phases 4 and 5 can run in parallel after Phase 3 completes.
