# Package Tracking — API Contracts

## Base Path

`/api/v1/packages`

All endpoints require a valid JWT with tenant and household context (except where noted).

---

## Resources

### Package Resource

```json
{
  "data": {
    "type": "packages",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "attributes": {
      "trackingNumber": "1Z999AA10123456784",
      "carrier": "ups",
      "label": "New keyboard",
      "notes": "Leave at back door",
      "status": "in_transit",
      "private": false,
      "estimatedDelivery": "2026-03-28",
      "actualDelivery": null,
      "lastPolledAt": "2026-03-26T14:30:00Z",
      "archivedAt": null,
      "createdAt": "2026-03-25T10:00:00Z",
      "updatedAt": "2026-03-26T14:30:00Z"
    },
    "relationships": {
      "user": {
        "data": { "type": "users", "id": "user-uuid" }
      },
      "trackingEvents": {
        "data": [
          { "type": "trackingEvents", "id": "event-uuid-1" },
          { "type": "trackingEvents", "id": "event-uuid-2" }
        ]
      }
    }
  }
}
```

### Redacted Package Resource (private, viewed by non-owner)

```json
{
  "data": {
    "type": "packages",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "attributes": {
      "trackingNumber": null,
      "carrier": "ups",
      "label": "Package",
      "notes": null,
      "status": null,
      "private": true,
      "estimatedDelivery": "2026-03-28",
      "actualDelivery": null,
      "lastPolledAt": null,
      "archivedAt": null,
      "createdAt": "2026-03-25T10:00:00Z",
      "updatedAt": "2026-03-26T14:30:00Z"
    },
    "relationships": {
      "user": {
        "data": { "type": "users", "id": "user-uuid" }
      }
    }
  }
}
```

### Tracking Event Resource

```json
{
  "type": "trackingEvents",
  "id": "event-uuid-1",
  "attributes": {
    "timestamp": "2026-03-26T08:15:00Z",
    "status": "in_transit",
    "description": "Departed FedEx location MEMPHIS, TN",
    "location": "Memphis, TN",
    "rawStatus": "DP"
  }
}
```

---

## Endpoints

### Create Package

```
POST /api/v1/packages
```

**Request Body:**
```json
{
  "data": {
    "type": "packages",
    "attributes": {
      "trackingNumber": "1Z999AA10123456784",
      "carrier": "ups",
      "label": "New keyboard",
      "notes": "Leave at back door",
      "private": false
    }
  }
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| trackingNumber | string | yes | Max 64 chars |
| carrier | string | yes | `usps`, `ups`, `fedex` |
| label | string | no | Max 255 chars |
| notes | string | no | Freeform text (delivery instructions, gift notes, etc.) |
| private | boolean | no | Default: false |

**Responses:**

| Status | Condition |
|--------|-----------|
| 201 Created | Package created, initial poll completed |
| 409 Conflict | Tracking number already exists in household |
| 422 Unprocessable Entity | Invalid tracking number, unsupported carrier, or household active package limit reached |

---

### List Packages

```
GET /api/v1/packages
```

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| filter[status] | string | all non-archived | Comma-separated statuses |
| filter[carrier] | string | all | Filter by carrier |
| filter[hasEta] | boolean | false | Only packages with ETA |
| filter[archived] | boolean | false | Include archived packages |
| sort | string | eta | `eta`, `-eta`, `createdAt`, `-createdAt` |

**Response:** `200 OK` — JSON:API array of package resources.

Private packages from other household members are included but redacted.

---

### Get Package

```
GET /api/v1/packages/{id}
```

**Response:** `200 OK` — Package resource with included tracking events.

**Errors:**

| Status | Condition |
|--------|-----------|
| 403 Forbidden | Private package owned by another user |
| 404 Not Found | Package not found in household |

---

### Update Package

```
PATCH /api/v1/packages/{id}
```

Only the package creator can update. Updatable fields: `label`, `notes`, `carrier`, `private`.

**Request Body:**
```json
{
  "data": {
    "type": "packages",
    "id": "uuid",
    "attributes": {
      "label": "Updated label",
      "notes": "Updated delivery instructions",
      "private": true
    }
  }
}
```

**Responses:**

| Status | Condition |
|--------|-----------|
| 200 OK | Updated successfully |
| 403 Forbidden | Not the package creator |
| 404 Not Found | Package not found |

---

### Delete Package

```
DELETE /api/v1/packages/{id}
```

Package creator or household admin/owner can delete.

**Response:** `204 No Content`

---

### Archive Package

```
POST /api/v1/packages/{id}/archive
```

**Response:** `200 OK` — Updated package with `archived` status.

---

### Unarchive Package

```
POST /api/v1/packages/{id}/unarchive
```

**Response:** `200 OK` — Package restored to `delivered` status.

---

### Refresh Tracking

```
POST /api/v1/packages/{id}/refresh
```

Triggers an immediate carrier API poll. Rate-limited to once per 5 minutes per package.

**Responses:**

| Status | Condition |
|--------|-----------|
| 200 OK | Refreshed successfully |
| 429 Too Many Requests | Last refresh was < 5 minutes ago |

**429 Response includes:**
```json
{
  "errors": [{
    "status": "429",
    "title": "Too Many Requests",
    "detail": "Package was last refreshed 2 minutes ago. Try again in 3 minutes.",
    "meta": {
      "retryAfterSeconds": 180
    }
  }]
}
```

---

### Package Summary

```
GET /api/v1/packages/summary
```

**Response:**
```json
{
  "data": {
    "type": "packageSummaries",
    "attributes": {
      "arrivingTodayCount": 2,
      "inTransitCount": 5,
      "exceptionCount": 0
    }
  }
}
```

Counts include private packages from all household members.

---

### Detect Carrier

```
GET /api/v1/packages/carriers/detect?trackingNumber=1Z999AA10123456784
```

No authentication required beyond valid JWT. Does not create a package.

**Response:**
```json
{
  "data": {
    "type": "carrierDetections",
    "attributes": {
      "trackingNumber": "1Z999AA10123456784",
      "detectedCarrier": "ups",
      "confidence": "high"
    }
  }
}
```

| Confidence | Meaning |
|------------|---------|
| high | Matches exactly one carrier pattern |
| medium | Matches multiple carrier patterns |
| low | No pattern match — user must select manually |

---

## Error Format

All errors follow JSON:API error format:

```json
{
  "errors": [{
    "status": "422",
    "title": "Unprocessable Entity",
    "detail": "Tracking number format is not recognized for carrier 'ups'"
  }]
}
```
