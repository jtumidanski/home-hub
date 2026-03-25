# REST API

All endpoints are prefixed with `/api/v1` and require JWT authentication. Request and response bodies use JSON:API format.

## Endpoints

### GET /api/v1/weather/current

Returns current weather conditions and today's high/low for the active household.

**Parameters:**

| Name      | In    | Type   | Required | Description                 |
|-----------|-------|--------|----------|-----------------------------|
| latitude  | query | float  | yes      | Household latitude          |
| longitude | query | float  | yes      | Household longitude         |
| units     | query | string | no       | "metric" or "imperial" (default: metric) |
| timezone  | query | string | no       | IANA timezone (default: UTC) |

**Response:** JSON:API `weather-current` resource.

| Attribute       | Type    |
|-----------------|---------|
| temperature     | float64 |
| temperatureUnit | string  |
| summary         | string  |
| icon            | string  |
| weatherCode     | int     |
| highTemperature | float64 |
| lowTemperature  | float64 |
| fetchedAt       | string  |

**Error Conditions:**

| Status | Condition                     |
|--------|-------------------------------|
| 404    | Missing latitude or longitude |
| 400    | Invalid coordinates           |
| 502    | Upstream API failure          |

---

### GET /api/v1/weather/forecast

Returns 7-day daily forecast for the active household.

**Parameters:** Same as `/weather/current`.

**Response:** JSON:API array of `weather-daily` resources.

| Attribute       | Type    |
|-----------------|---------|
| date            | string  |
| highTemperature | float64 |
| lowTemperature  | float64 |
| temperatureUnit | string  |
| summary         | string  |
| icon            | string  |
| weatherCode     | int     |

**Error Conditions:** Same as `/weather/current`.

---

### GET /api/v1/weather/geocoding

Returns place search results for autocomplete. Proxies Open-Meteo Geocoding API.

**Parameters:**

| Name | In    | Type   | Required | Description                |
|------|-------|--------|----------|----------------------------|
| q    | query | string | yes      | Search term (min 2 chars)  |

**Response:** JSON:API array of `geocoding-results` resources.

| Attribute | Type    |
|-----------|---------|
| name      | string  |
| country   | string  |
| admin1    | string  |
| latitude  | float64 |
| longitude | float64 |

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | Missing or short query |
| 502    | Upstream API failure   |
