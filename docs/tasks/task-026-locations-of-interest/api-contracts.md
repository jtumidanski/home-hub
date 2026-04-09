# API Contracts — Locations of Interest

All endpoints prefixed with `/api/v1`. JWT auth required. Bodies follow JSON:API.

## Resource Type

`location-of-interest`

| Attribute  | Type    | Notes                                              |
|------------|---------|----------------------------------------------------|
| label      | string  | Optional friendly name, max 64 chars               |
| placeName  | string  | Geocoded display name (e.g. "Paris, Île-de-France, France") |
| latitude   | float64 | Required                                           |
| longitude  | float64 | Required                                           |
| createdAt  | string  | ISO 8601                                           |
| updatedAt  | string  | ISO 8601                                           |

---

## GET /api/v1/locations-of-interest

Lists all locations of interest for the active household.

**Parameters:** none.

**Response 200:**

```json
{
  "data": [
    {
      "type": "location-of-interest",
      "id": "9f1b...",
      "attributes": {
        "label": "Mom's House",
        "placeName": "Paris, Île-de-France, France",
        "latitude": 48.8566,
        "longitude": 2.3522,
        "createdAt": "2026-04-09T12:00:00Z",
        "updatedAt": "2026-04-09T12:00:00Z"
      }
    }
  ]
}
```

**Errors:**

| Status | Condition          |
|--------|--------------------|
| 401    | Missing/invalid JWT |

---

## POST /api/v1/locations-of-interest

Creates a new location of interest for the active household.

**Request:**

```json
{
  "data": {
    "type": "location-of-interest",
    "attributes": {
      "label": "Beach House",
      "placeName": "Outer Banks, NC, USA",
      "latitude": 35.5582,
      "longitude": -75.4665
    }
  }
}
```

`label` is optional; all other attributes are required. `label` is trimmed and rejected if longer than 64 characters after trimming. `latitude` and `longitude` are normalized to 4 decimal places before persistence.

After the row is persisted, the server synchronously fetches current weather and the 7-day forecast for the new coordinates and writes the corresponding `weather_caches` row. If that upstream call fails, the location is still created and the response is still `201`; the cache will be filled by the next refresh tick or first user view.

**Response 201:** the created `location-of-interest` resource (same shape as list entries).

**Errors:**

| Status | Condition                                                      |
|--------|----------------------------------------------------------------|
| 400    | Missing required attribute, invalid coordinates, label too long |
| 401    | Missing/invalid JWT                                            |
| 409    | Household has reached the 10-location cap. Error message: `"Households can save up to 10 locations of interest. Remove one to add another."` |

---

## PATCH /api/v1/locations-of-interest/{id}

Updates the friendly label of a saved location. Coordinates and place name are immutable.

**Request:**

```json
{
  "data": {
    "type": "location-of-interest",
    "id": "9f1b...",
    "attributes": {
      "label": "Renamed"
    }
  }
}
```

A `label` of `""` (empty string) clears the friendly label.

**Response 200:** the updated resource.

**Errors:**

| Status | Condition                                                  |
|--------|------------------------------------------------------------|
| 400    | Label exceeds 64 chars, or unsupported attribute provided  |
| 401    | Missing/invalid JWT                                        |
| 404    | Location not found, or not owned by the active household   |

---

## DELETE /api/v1/locations-of-interest/{id}

Deletes a location of interest and its corresponding cache row in `weather_caches`.

**Response 204:** no body.

**Errors:**

| Status | Condition                                                |
|--------|----------------------------------------------------------|
| 401    | Missing/invalid JWT                                      |
| 404    | Location not found, or not owned by the active household |

---

## Modified: GET /api/v1/weather/current

**New optional parameter:**

| Name       | In    | Type   | Required | Description                                        |
|------------|-------|--------|----------|----------------------------------------------------|
| locationId | query | uuid   | no       | If set, fetch weather for this saved location.     |
| latitude   | query | float  | conditional | Required when `locationId` is omitted.          |
| longitude  | query | float  | conditional | Required when `locationId` is omitted.          |
| units      | query | string | no       | Inherited from household if omitted.               |
| timezone   | query | string | no       | Inherited from household if omitted.               |

When `locationId` is provided, the server resolves it to coordinates from `locations_of_interest` (verifying ownership by the active household), and any `latitude`/`longitude` query parameters are ignored. Cache is keyed by `(household_id, location_id)`.

**Errors (additions):**

| Status | Condition                                              |
|--------|--------------------------------------------------------|
| 404    | `locationId` not found or not owned by the household   |

---

## Modified: GET /api/v1/weather/forecast

Same modifications as `/weather/current`. Same error additions.

---

## Unchanged

`GET /api/v1/weather/geocoding` is unchanged — the frontend reuses it for the place-search step when adding a new location of interest.
