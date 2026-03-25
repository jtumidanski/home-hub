# Task 006 — Weather Enhancements: Task Checklist

Last Updated: 2026-03-25

---

## Phase 1: Backend — Open-Meteo Client Extension

- [x] **1.1** Add `HourlyData` and `HourlyUnits` structs to `openmeteo/types.go`; add `Hourly`/`HourlyUnits` fields to `ForecastResponse`
- [x] **1.2** Add `hourly=temperature_2m,weather_code,precipitation_probability` param to `FetchForecast()` in `openmeteo/client.go`
- [x] **1.3** Update `openmeteo/client_test.go` — mock responses include hourly data, verify parsing

## Phase 2: Backend — Domain Model, Processor, Cache

- [x] **2.1** Add `HourlyForecast` struct to `forecast/model.go` (Time, Temperature, WeatherCode, Summary, Icon, PrecipitationProbability)
- [x] **2.2** Add `HourlyForecast []HourlyForecast` field to `DailyForecast` struct
- [x] **2.3** Update `forecast/builder.go` — ensure hourly data flows through builder
- [x] **2.4** Update `transformResponse()` in `forecast/processor.go` — parse hourly arrays, group by day, map weather codes
- [x] **2.5** Add cache staleness check in `processor.go` — treat entries without hourly data as stale
- [x] **2.6** Verify entity JSONB round-trip in `forecast/entity.go` — hourly data serializes/deserializes correctly
- [x] **2.7** Add/update tests: `processor_test.go` (hourly parsing), `entity_test.go` (round-trip), `builder_test.go`

## Phase 3: Backend — REST Layer

- [x] **3.1** Add `HourlyRestModel` struct to `forecast/rest.go`
- [x] **3.2** Add `HourlyForecast []HourlyRestModel` field to `DailyRestModel`
- [x] **3.3** Update `TransformForecast()` — map hourly domain models to REST models
- [x] **3.4** Update `forecast/rest_test.go` — test hourly REST transformation
- [x] **3.5** Run `go build ./...` and `go test ./...` for weather-service — all pass
- [x] **3.6** Verify Docker build succeeds

## Phase 4: Frontend — Dashboard Forecast Preview

- [x] **4.1** Add `HourlyForecastEntry` interface to `types/models/weather.ts`
- [x] **4.2** Add `hourlyForecast` field to `WeatherDailyAttributes`
- [x] **4.3** Update `weather-widget.tsx` — add `useWeatherForecast()` hook call
- [x] **4.4** Render 4 mini forecast cards below current conditions (Today + 3 days)
- [x] **4.5** Each card shows: abbreviated day name, weather icon, high/low temp
- [x] **4.6** Handle forecast fetch failure gracefully — hide preview, keep current conditions

## Phase 5: Frontend — Expandable Hourly Breakdown

- [x] **5.1** Add expand/collapse state to `WeatherPage.tsx` (accordion-style, one at a time)
- [x] **5.2** Make day cards tappable to toggle expanded state
- [x] **5.3** Render hourly entries when expanded: time, weather icon, temperature, precipitation %
- [x] **5.4** Filter today's hourly entries to current hour onward (inclusive)
- [x] **5.5** Show all 24 hours for future days
- [x] **5.6** Verify mobile UX: legible at 320px, scrollable hourly list, tap-only (no swipe)

## Final Verification

- [x] **6.1** All backend tests pass (`go test ./...`)
- [x] **6.2** Docker build succeeds for weather-service
- [x] **6.3** Frontend builds without errors
- [ ] **6.4** Manual smoke test: dashboard preview + expandable hourly on weather page
