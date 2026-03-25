# Weather Enhancements — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-25
---

## 1. Overview

This feature enhances the existing weather functionality in two areas. First, the dashboard weather widget is upgraded from a single-day summary to a richer card that includes a row of mini forecast cards showing today and the next three days. Second, the weather page's 7-day forecast gains expandable daily rows — tapping a day reveals an hourly breakdown with temperature, conditions, and precipitation probability.

These changes require the weather-service to fetch hourly forecast data from Open-Meteo in addition to the daily data it already retrieves, and to include that hourly data in the existing forecast endpoint response and cache.

## 2. Goals

Primary goals:
- Upgrade the dashboard weather widget to show a 4-day forecast preview (today + 3 days)
- Add expandable hourly breakdown to each day on the weather page
- Extend the weather-service to fetch and cache hourly forecast data from Open-Meteo
- Extend the existing forecast endpoint to return hourly data per day

Non-goals:
- Hourly data beyond the 7-day forecast window
- Separate hourly-only page or route
- Precipitation graphs or detailed charts
- Weather alerts or notifications
- Changes to household location or settings

## 3. User Stories

- As a household member, I want to see a quick forecast preview on the dashboard so I can plan the next few days at a glance without navigating to the weather page.
- As a household member, I want to expand a day on the weather page to see an hourly breakdown so I can plan activities around specific times.
- As a household member, I want to see the precipitation probability for each hour so I know when rain is likely.
- As a household member, I want today's hourly view to only show remaining hours (including current) so I see relevant information.

## 4. Functional Requirements

### 4.1 Dashboard Weather Widget — Forecast Preview

- Below the existing current conditions display, add a row of 4 mini forecast cards.
- Cards represent today (N), tomorrow (N+1), day after (N+2), and N+3.
- Each mini card shows:
  - Abbreviated day name (e.g., "Mon", "Tue") or "Today" for the first card
  - Weather condition icon
  - High temperature
  - Low temperature
- The dashboard fetches the forecast endpoint in addition to the current endpoint to populate the preview cards.
- If the forecast request fails, the current conditions area still displays; the preview row shows a subtle error state or is hidden.
- The existing "no location set" prompt behavior is unchanged.

### 4.2 Weather Page — Expandable Hourly Breakdown

- Each day card on the 7-day forecast becomes expandable (tap to toggle).
- When expanded, the card reveals a list of hourly entries for that day.
- Each hourly entry shows:
  - Time (formatted to household timezone, e.g., "2 PM", "14:00")
  - Weather condition icon
  - Temperature
  - Precipitation probability (e.g., "30%")
- For today: only hours from the current hour onward are displayed (including the current hour).
- For future days: all 24 hours are displayed.
- Only one day can be expanded at a time, or multiple — no strong preference, but collapsing others on expand keeps the page clean. Follow whichever pattern feels best for mobile usability.
- The expanded state is local UI state — not persisted.

### 4.3 Weather Service — Hourly Data Fetch

- The Open-Meteo Forecast API call is extended to request hourly variables in addition to the existing daily variables.
- Hourly variables requested: `temperature_2m`, `weather_code`, `precipitation_probability`.
- The hourly data covers the same 7-day window as the daily forecast.
- Unit system (metric/imperial) applies to hourly temperature values (same parameter already passed to Open-Meteo).

### 4.4 Forecast Endpoint Extension

- The existing `GET /api/v1/weather/forecast` response is extended.
- Each `weather-daily` resource gains a new `hourlyForecast` attribute containing an array of hourly entries.
- Each hourly entry contains: `time` (ISO 8601 datetime), `temperature`, `weatherCode`, `summary`, `icon`, `precipitationProbability`.
- The `temperatureUnit` on the daily resource applies to all hourly temperatures within it.
- This is a backwards-compatible addition — existing consumers that ignore `hourlyForecast` are unaffected.

### 4.5 Cache Schema Extension

- The `forecast_data` JSONB column in `weather_caches` is extended to include hourly data nested within each daily entry.
- No schema migration is needed — the JSONB structure simply grows. Existing cache rows without hourly data are treated as cache misses and re-fetched.
- The background refresh job automatically populates hourly data on its next cycle.

## 5. API Surface

### 5.1 GET /api/v1/weather/forecast (modified)

**Response:** JSON:API array of `weather-daily` resources (unchanged wrapper).

Each `weather-daily` resource now includes:

| Attribute              | Type   | Description                          | New? |
|------------------------|--------|--------------------------------------|------|
| date                   | string | ISO 8601 date (YYYY-MM-DD)          | No   |
| highTemperature        | float64| Daily high                           | No   |
| lowTemperature         | float64| Daily low                            | No   |
| temperatureUnit        | string | "°C" or "°F"                         | No   |
| summary                | string | Human-readable condition             | No   |
| icon                   | string | Icon key for frontend rendering      | No   |
| weatherCode            | int    | WMO weather code                     | No   |
| hourlyForecast         | array  | Array of hourly entries (see below)  | Yes  |

**Hourly entry shape** (within `hourlyForecast` array):

| Field                     | Type    | Description                        |
|---------------------------|---------|------------------------------------|
| time                      | string  | ISO 8601 datetime                  |
| temperature               | float64 | Hourly temperature                 |
| weatherCode               | int     | WMO weather code                   |
| summary                   | string  | Human-readable condition           |
| icon                      | string  | Icon key                           |
| precipitationProbability  | int     | Percentage (0-100)                 |

No changes to error conditions or other endpoints.

## 6. Data Model

### 6.1 Weather Service — Cache Extension

The `forecast_data` JSONB structure is extended from:

```json
[
  {
    "date": "2026-03-25",
    "highTemperature": 78.0,
    "lowTemperature": 55.0,
    "weatherCode": 2,
    "summary": "Partly Cloudy",
    "icon": "cloud-sun"
  }
]
```

To:

```json
[
  {
    "date": "2026-03-25",
    "highTemperature": 78.0,
    "lowTemperature": 55.0,
    "weatherCode": 2,
    "summary": "Partly Cloudy",
    "icon": "cloud-sun",
    "hourlyForecast": [
      {
        "time": "2026-03-25T00:00",
        "temperature": 58.0,
        "weatherCode": 1,
        "summary": "Mainly Clear",
        "icon": "sun",
        "precipitationProbability": 0
      }
    ]
  }
]
```

No new tables, columns, or indexes. Existing rows without `hourlyForecast` are treated as stale and re-fetched.

## 7. Service Impact

### 7.1 weather-service

- **Open-Meteo client**: Add hourly variables (`temperature_2m`, `weather_code`, `precipitation_probability`) to the forecast API request.
- **Domain model**: Add `HourlyForecast` struct and embed an array of it in `DailyForecast`.
- **Processor**: Update forecast parsing to include hourly data grouped by day.
- **REST layer**: Update `DailyRestModel` to include `hourlyForecast` in the JSON:API response.
- **Cache**: No schema change — the JSONB payload grows to include hourly data. Cache entries missing hourly data are re-fetched.
- **Background refresh**: Automatically picks up the new data on next cycle since it re-fetches the full forecast.

### 7.2 frontend

- **Dashboard weather widget**: Add a row of 4 mini forecast cards below current conditions. Fetch `/forecast` in addition to `/current`.
- **Weather page**: Make daily cards expandable. When expanded, render hourly entries. For today, filter to current hour onward.
- **API client**: Update forecast response types to include `hourlyForecast` array.
- **WeatherIcon component**: No changes needed — already supports all required icon keys.

### 7.3 No changes to

- account-service
- Infrastructure (no new services, routes, or deployments)

## 8. Non-Functional Requirements

### Performance
- Hourly data adds ~7x payload to the forecast response (168 hourly entries across 7 days). This is acceptable for a single-household response but should be monitored.
- The dashboard makes two parallel API calls (`/current` and `/forecast`). Both serve from cache so latency remains sub-100ms.
- The expanded hourly view renders on the client from already-fetched data — no additional API call on expand.

### Cache Size
- Estimated forecast_data JSONB size increase: from ~1 KB to ~10-15 KB per household. Acceptable for the expected household count.

### Mobile UX
- Mini forecast cards on the dashboard must be legible on small screens (minimum 320px viewport).
- Hourly breakdown should be scrollable within the expanded card if it overflows.
- Expand/collapse interaction uses tap only (no swipe gestures, per user preference).

### Security
- No new endpoints — existing auth and tenant scoping applies.

### Observability
- No new observability requirements beyond existing patterns.

## 9. Open Questions

None — all questions resolved during scoping.

## 10. Acceptance Criteria

- [ ] Dashboard weather widget shows a row of 4 mini forecast cards (today + 3 days) with day name, icon, and high/low temperature.
- [ ] Dashboard forecast preview fetches from the existing `/forecast` endpoint.
- [ ] Dashboard widget gracefully handles forecast fetch failure (current conditions still display).
- [ ] Weather page daily cards are expandable via tap.
- [ ] Expanded day shows hourly entries with time, icon, temperature, and precipitation probability.
- [ ] Today's hourly view shows only the current hour onward.
- [ ] Future days show all 24 hours.
- [ ] Weather-service fetches hourly variables from Open-Meteo (temperature_2m, weather_code, precipitation_probability).
- [ ] Forecast endpoint response includes `hourlyForecast` array in each daily resource.
- [ ] Hourly entries include weather code summary and icon (same mapping as daily).
- [ ] Cache stores hourly data in the existing `forecast_data` JSONB column.
- [ ] Existing cache entries without hourly data are treated as stale and re-fetched.
- [ ] Temperature units (metric/imperial) apply consistently to hourly data.
- [ ] All new backend code has unit tests.
- [ ] All affected services build and pass tests.
