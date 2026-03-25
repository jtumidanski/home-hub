# Weather Forecast — API Contracts

## Account Service — Household Extension

### PATCH /api/v1/households/{id}

New optional fields added to the existing endpoint.

**Request (partial):**

```json
{
  "data": {
    "type": "households",
    "id": "<uuid>",
    "attributes": {
      "latitude": 40.7128,
      "longitude": -74.0060,
      "locationName": "New York, NY, United States"
    }
  }
}
```

**Validation rules:**
- If `latitude` is provided, `longitude` must also be provided (and vice versa).
- `latitude`: -90.0 to 90.0
- `longitude`: -180.0 to 180.0
- `locationName` is optional but recommended when setting coordinates.
- To clear location, send `latitude: null, longitude: null, locationName: null`.

**Response** includes the new fields in the `households` resource attributes alongside existing fields (name, timezone, units).

---

## Weather Service

### GET /api/v1/weather/current

Returns current conditions and today's forecast for the active household.

**Request:** No body. Household context derived from JWT + active household.

**Response:**

```json
{
  "data": {
    "type": "weather-current",
    "id": "<household-uuid>",
    "attributes": {
      "temperature": 72.5,
      "temperatureUnit": "°F",
      "summary": "Partly Cloudy",
      "icon": "cloud-sun",
      "weatherCode": 2,
      "highTemperature": 78.0,
      "lowTemperature": 55.0,
      "fetchedAt": "2026-03-25T14:30:00Z"
    }
  }
}
```

**Error responses:**

```json
// 404 — No location set
{
  "errors": [{
    "status": "404",
    "title": "No Location",
    "detail": "The active household does not have a location configured. Set a location in household settings."
  }]
}

// 502 — Upstream failure (no cache available)
{
  "errors": [{
    "status": "502",
    "title": "Weather Unavailable",
    "detail": "Unable to retrieve weather data. Please try again later."
  }]
}
```

---

### GET /api/v1/weather/forecast

Returns 7-day daily forecast for the active household.

**Request:** No body.

**Response:**

```json
{
  "data": [
    {
      "type": "weather-daily",
      "id": "2026-03-25",
      "attributes": {
        "date": "2026-03-25",
        "highTemperature": 78.0,
        "lowTemperature": 55.0,
        "temperatureUnit": "°F",
        "summary": "Partly Cloudy",
        "icon": "cloud-sun",
        "weatherCode": 2
      }
    },
    {
      "type": "weather-daily",
      "id": "2026-03-26",
      "attributes": {
        "date": "2026-03-26",
        "highTemperature": 65.0,
        "lowTemperature": 48.0,
        "temperatureUnit": "°F",
        "summary": "Rain",
        "icon": "cloud-rain",
        "weatherCode": 61
      }
    }
  ]
}
```

The array always contains 7 entries, starting with today (based on the household's timezone).

---

### GET /api/v1/weather/geocoding?q={query}

Proxies Open-Meteo's geocoding API for place search autocomplete.

**Request:**

| Parameter | Type   | Required | Constraints          |
|-----------|--------|----------|----------------------|
| q         | string | yes      | Minimum 2 characters |

**Response:**

```json
{
  "data": [
    {
      "type": "geocoding-results",
      "id": "5128581",
      "attributes": {
        "name": "New York",
        "country": "United States",
        "admin1": "New York",
        "latitude": 40.7128,
        "longitude": -74.006
      }
    },
    {
      "type": "geocoding-results",
      "id": "2643743",
      "attributes": {
        "name": "London",
        "country": "United Kingdom",
        "admin1": "England",
        "latitude": 51.5074,
        "longitude": -0.1278
      }
    }
  ]
}
```

Results are limited to 10 entries. The `id` is the Open-Meteo geocoding ID (integer, used as string for JSON:API compliance).

**Error responses:**

```json
// 400 — Bad query
{
  "errors": [{
    "status": "400",
    "title": "Invalid Query",
    "detail": "Search query must be at least 2 characters."
  }]
}
```

---

## Cache Invalidation on Location Change

When the frontend PATCHes a household's location, the weather cache for that household becomes stale. The weather-service handles this as follows:

- On `GET /api/v1/weather/current` or `GET /api/v1/weather/forecast`, the service compares the cached latitude/longitude/units against the household's current values (passed via the request context or derived from the cache record).
- If the cached coordinates or units differ from those on the household (detected because the cache row stores lat/lon/units at fetch time), the cache is treated as a miss and a fresh fetch is performed.
- If the household's location is cleared (latitude/longitude set to null), the weather endpoints return 404 and any existing cache row is deleted on the next background refresh cycle or on the next API call.

This approach avoids cross-service event buses — the weather-service is self-healing based on coordinate mismatches.

---

## Open-Meteo API Usage

### Forecast API

**Endpoint:** `https://api.open-meteo.com/v1/forecast`

**Parameters used:**

| Parameter        | Value                                              |
|------------------|----------------------------------------------------|
| latitude         | From household                                     |
| longitude        | From household                                     |
| current          | `temperature_2m,weather_code`                      |
| daily            | `temperature_2m_max,temperature_2m_min,weather_code` |
| temperature_unit | `celsius` or `fahrenheit` (from household units)   |
| timezone         | From household timezone field                      |
| forecast_days    | `7`                                                |

**No API key required.**

### Geocoding API

**Endpoint:** `https://geocoding-api.open-meteo.com/v1/search`

**Parameters used:**

| Parameter | Value              |
|-----------|--------------------|
| name      | User search query  |
| count     | `10`               |
| language  | `en`               |

**No API key required.**

---

## WMO Weather Code Mapping

The weather-service maps WMO weather codes from Open-Meteo to human-readable summaries:

| Code Range  | Summary                | Icon Key          |
|-------------|------------------------|-------------------|
| 0           | Clear                  | sun               |
| 1           | Mostly Clear           | sun               |
| 2           | Partly Cloudy          | cloud-sun         |
| 3           | Overcast               | cloud             |
| 45, 48      | Fog                    | cloud-fog         |
| 51, 53, 55  | Drizzle                | cloud-drizzle     |
| 56, 57      | Freezing Drizzle       | cloud-drizzle     |
| 61, 63, 65  | Rain                   | cloud-rain        |
| 66, 67      | Freezing Rain          | cloud-rain        |
| 71, 73, 75  | Snow                   | snowflake         |
| 77          | Snow Grains            | snowflake         |
| 80, 81, 82  | Rain Showers           | cloud-rain        |
| 85, 86      | Snow Showers           | snowflake         |
| 95          | Thunderstorm           | cloud-lightning   |
| 96, 99      | Thunderstorm with Hail | cloud-lightning   |

The `icon` key maps directly to Lucide icon component names used in the frontend (e.g., `cloud-rain` → `<CloudRain />`). The backend is the single source of truth for this mapping — the frontend renders the icon by name without duplicating WMO code logic.
