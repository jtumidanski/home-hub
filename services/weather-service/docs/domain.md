# Domain

## Forecast Cache

### Responsibility

Caches weather data (current conditions and 7-day forecast) fetched from Open-Meteo, keyed by household. Serves cached data to avoid hitting rate limits and provides background refresh.

### Core Models

**Model** (`forecast.Model`)

| Field        | Type            |
|--------------|-----------------|
| id           | uuid.UUID       |
| tenantID     | uuid.UUID       |
| householdID  | uuid.UUID       |
| latitude     | float64         |
| longitude    | float64         |
| units        | string          |
| currentData  | CurrentData     |
| forecastData | []DailyForecast |
| fetchedAt    | time.Time       |
| createdAt    | time.Time       |
| updatedAt    | time.Time       |

**CurrentData** (embedded struct)

| Field       | Type    |
|-------------|---------|
| temperature | float64 |
| weatherCode | int     |
| summary     | string  |
| icon        | string  |

**DailyForecast** (embedded struct)

| Field           | Type    |
|-----------------|---------|
| date            | string  |
| highTemperature | float64 |
| lowTemperature  | float64 |
| weatherCode     | int     |
| summary         | string  |
| icon            | string  |

All fields on Model are immutable after construction. Access is through getter methods.

### Invariants

- One cache entry per household (unique on household_id).
- Cache stores coordinates and units at fetch time for self-healing invalidation.
- If cached coordinates/units don't match the request, cache is treated as a miss and re-fetched.

### Processors

**Processor** (`forecast.Processor`)

| Method                                                  | Description                                      |
|---------------------------------------------------------|--------------------------------------------------|
| `GetCurrent(tenantID, householdID, lat, lon, units, tz)` | Returns current weather from cache or fetches    |
| `GetForecast(tenantID, householdID, lat, lon, units, tz)` | Returns forecast from cache or fetches          |
| `RefreshCache(entity)`                                  | Re-fetches weather for an existing cache entry   |
| `InvalidateCache(householdID)`                          | Deletes the cache entry for a household          |
| `ByHouseholdIDProvider(householdID)`                    | Returns a provider for a single household's cache entry |
| `AllProvider()`                                         | Returns a provider for all cache entries (for background refresh) |

---

## Geocoding

### Responsibility

Proxies Open-Meteo's geocoding API for place search autocomplete. No persistence — results are returned directly.

---

## Weather Codes

### Responsibility

Maps WMO weather codes to human-readable summaries and Lucide icon keys. Pure lookup function with no state.

| Method               | Description                                    |
|----------------------|------------------------------------------------|
| `Lookup(code int)`   | Returns (summary string, icon string) for a WMO code |

---

## Background Refresh

### Responsibility

Periodically refreshes all weather cache entries using an in-process ticker goroutine.

- Runs at a configurable interval (default: 30 minutes).
- Queries all `weather_caches` rows and re-fetches from Open-Meteo.
- Respects rate limits (~1 request/second via the Open-Meteo client throttle).
- Errors are logged per-household but do not stop the loop.
- Stops cleanly on context cancellation (graceful shutdown).
