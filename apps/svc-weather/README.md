# svc-weather

Weather microservice for Home Hub that provides current temperature and 7-day forecasts with spatial de-duplication and stale-while-revalidate caching.

## Features

- **Cache-Only Request Path**: HTTP requests NEVER call Open-Meteo directly; only read from cache for predictable <300ms latency
- **Spatial De-Duplication**: 5-digit geohash (~5km precision) means nearby households share weather cache
- **Stale-While-Revalidate**: Serve stale data up to 24h when provider unavailable
- **Redis or In-Memory**: Production uses Redis for shared cache; in-memory fallback for development
- **JSON:API Compliant**: All endpoints follow JSON:API specification

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        svc-weather                           │
├─────────────────────────────────────────────────────────────┤
│  HTTP Handlers (JSON:API)                                    │
│    ├─ GET /api/v1/weather/combined?householdId=:id          │
│    ├─ GET /api/v1/weather/current?householdId=:id           │
│    └─ GET /api/v1/weather/forecast?householdId=:id          │
├─────────────────────────────────────────────────────────────┤
│  Business Logic                                              │
│    ├─ WeatherProvider (cache-only reads)                    │
│    ├─ HouseholdResolver (lat/lon/tz lookup + 24h cache)     │
│    └─ GeoKeyGenerator (5-digit geohash)                     │
├─────────────────────────────────────────────────────────────┤
│  External Integrations                                       │
│    ├─ Open-Meteo HTTP Client (current + forecast)           │
│    ├─ Redis Cache Client (with in-memory fallback)          │
│    └─ svc-users HTTP Client (household lookup)              │
└─────────────────────────────────────────────────────────────┘
```

## API Endpoints

### Public Kiosk Endpoints

#### GET /api/v1/weather/combined

Get both current weather and 7-day forecast.

**Query Parameters:**
- `householdId` (required): UUID of the household

**Response:**
```json
{
  "jsonapi": { "version": "1.0" },
  "data": {
    "type": "weather",
    "id": "household:123",
    "attributes": {
      "current": {
        "temperature_c": 12.3,
        "observed_at": "2025-11-11T18:55:00Z",
        "stale": false,
        "age_seconds": 60
      },
      "daily": [
        { "date": "2025-11-11", "tmax_c": 14.1, "tmin_c": 5.2 },
        { "date": "2025-11-12", "tmax_c": 13.3, "tmin_c": 4.9 }
      ],
      "units": "celsius",
      "stale": false
    },
    "meta": {
      "source": "open-meteo",
      "timezone": "America/Detroit",
      "geokey": "dpjz9",
      "refreshed_at": "2025-11-11T19:00:00Z"
    }
  }
}
```

#### GET /api/v1/weather/current

Get current weather only.

**Query Parameters:**
- `householdId` (required): UUID of the household

#### GET /api/v1/weather/forecast

Get 7-day forecast only.

**Query Parameters:**
- `householdId` (required): UUID of the household
- `days` (optional): Number of days (1-14, default: 7)

### Admin Endpoints (Auth Required)

#### DELETE /admin/weather/cache

Purge cached weather data for a specific household.

**Query Parameters:**
- `householdId` (required): UUID of the household

**Response:** 204 No Content

#### POST /admin/weather/refresh

Trigger async refresh of weather data for a household.

**Query Parameters:**
- `householdId` (required): UUID of the household

**Response:** 202 Accepted

#### DELETE /admin/weather/cache/all

Purge all cached weather data (use with caution).

**Response:** 204 No Content

## Configuration

All configuration via environment variables:

### Redis
```bash
REDIS_URL=redis://localhost:6379/0  # Empty = in-memory fallback
```

### TTLs
```bash
CURRENT_TTL=5m                      # Current weather TTL
FORECAST_TTL=1h                     # Forecast TTL
STALE_MAX=24h                       # Max stale age before 503
```

### Refresh
```bash
REFRESH_JITTER=0.2                  # ±20% jitter for refresh intervals
```

### Geohash
```bash
GEOHASH_PREC=5                      # 5-digit geohash (~5km precision)
```

### Open-Meteo
```bash
OPENMETEO_BASE_URL=https://api.open-meteo.com/v1/forecast
OPENMETEO_TIMEOUT=10s               # HTTP timeout
```

### svc-users
```bash
SVC_USERS_BASE_URL=http://svc-users:8080
SVC_USERS_TIMEOUT=5s                # HTTP timeout
HOUSEHOLD_CACHE_TTL=24h             # Household location cache TTL
```

### Service
```bash
SERVICE_PORT=8080
LOG_LEVEL=info
```

## Building

```bash
go build -o svc-weather
```

## Running

```bash
./svc-weather
```

## Docker

```bash
docker build -t svc-weather:latest .
docker run -p 8080:8080 -e REDIS_URL=redis://redis:6379/0 svc-weather:latest
```

## Development

### Prerequisites
- Go 1.24+
- Redis 6.0+ (optional, falls back to in-memory)
- Access to svc-users service

### Local Testing
```bash
# Without Redis (in-memory fallback)
go run main.go config.go migrations.go

# With Redis
REDIS_URL=redis://localhost:6379/0 go run main.go config.go migrations.go
```

## Key Design Decisions

### 1. Cache-Only Request Path
HTTP requests NEVER call Open-Meteo directly; they only read from cache. This ensures:
- Predictable latency (P99 < 300ms)
- No rate limiting on provider
- Decoupled request path from external dependency

### 2. Spatial De-Duplication with Geohash
5-digit geohash (~5km precision) used as cache key instead of exact lat/lon:
- Households within ~5km share weather cache
- Reduces provider calls significantly
- ~5km radius acceptable for temperature accuracy

### 3. Stale-While-Revalidate
Serve stale data up to 24h when Open-Meteo unavailable:
- Service remains functional during provider outages
- Yesterday's forecast still useful vs. no data
- Staleness indicated in response (`"stale": true`)

### 4. Redis with In-Memory Fallback
- **Production**: Redis for shared cache across replicas (k8s HPA)
- **Development**: In-memory cache (simpler local dev)
- Automatic fallback ensures service doesn't crash on Redis failure

### 5. No Database Persistence
Weather service is cache-only; no PostgreSQL database:
- Weather data is ephemeral (only need current + 7-day forecast)
- Cache (Redis/in-memory) sufficient for TTL-based data
- Reduces complexity and resource usage

### 6. Household Resolution with 24h Cache
Cache household location data (lat/lon/tz) from svc-users for 24h:
- Household location rarely changes
- Reduces traffic to svc-users
- 24h staleness acceptable (location updates manual)

## Performance

- **P99 latency**: < 300ms for cached endpoints
- **Cache hit ratio target**: > 95% for current, > 98% for forecast
- **Provider call rate**: < 1 call per household per TTL period
- **Stale data rate**: < 1% under normal operation

## Dependencies

### External
- **Open-Meteo API**: Free tier, 10,000 calls/day, no API key required
- **Redis**: 6.0+, ~100MB per 1000 households (estimated)

### Internal
- **svc-users**: For household location resolution

## Notes

- All temperatures stored and cached in **Celsius**
- Display-level conversion to Fahrenheit happens in frontend
- No background scheduler implemented yet (manual refresh via admin endpoints)
- Future: Automatic background refresh workers

## Future Enhancements (Out of Scope for MVP)

- Background scheduler with refresh workers (5m current, 1h forecast)
- Multiple weather providers with automatic failover
- Richer weather data (precipitation, wind, humidity)
- Historical weather data for analytics
- Weather alerts and notifications
