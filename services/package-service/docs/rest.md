# REST API

All endpoints are prefixed with `/api/v1`. All endpoints require JWT authentication. Request and response bodies use JSON:API format. Tenant and household context are derived from the JWT.

## Endpoints

### POST /api/v1/packages

Creates a new package. Triggers an initial carrier API poll on creation.

**Request:** JSON:API `packages` resource.

| Attribute      | Type    | Required |
|----------------|---------|----------|
| trackingNumber | string  | yes      |
| carrier        | string  | yes      |
| label          | string  | no       |
| notes          | string  | no       |
| private        | boolean | no       |

**Response:** JSON:API `packages` resource. Status 201.

**Error Conditions:**

| Status | Condition                                      |
|--------|------------------------------------------------|
| 400    | Missing or invalid tracking number or carrier  |
| 409    | Tracking number already exists in household    |
| 422    | Household has reached max active package limit |

---

### GET /api/v1/packages

Lists packages for the current household.

**Parameters:**

| Name             | In    | Type   | Required | Description                              |
|------------------|-------|--------|----------|------------------------------------------|
| filter[archived] | query | string | no       | Set to `true` to include archived        |
| filter[status]   | query | string | no       | Comma-separated status values            |
| filter[hasEta]   | query | string | no       | Set to `true` to filter to packages with ETA |
| sort             | query | string | no       | Sort field                               |

**Response:** JSON:API array of `packages` resources.

| Attribute         | Type    |
|-------------------|---------|
| trackingNumber    | string  |
| carrier           | string  |
| label             | string  |
| notes             | string  |
| status            | string  |
| private           | boolean |
| estimatedDelivery | string  |
| actualDelivery    | string  |
| lastPolledAt      | string  |
| archivedAt        | string  |
| isOwner           | boolean |
| userId            | string  |
| createdAt         | string  |
| updatedAt         | string  |

**Privacy Rules:**

- If package is private and requester is not the owner: trackingNumber, notes, status, and lastPolledAt set to null; label replaced with "Package".
- `isOwner` indicates whether the requesting user created the package.

**Sorting:** Packages sorted by estimated delivery ascending (nulls last), then created at descending.

---

### GET /api/v1/packages/{id}

Returns a single package with tracking event history.

**Response:** JSON:API `packages` resource with additional `trackingEvents` attribute.

| Attribute      | Type   |
|----------------|--------|
| trackingEvents | array  |

Each tracking event contains:

| Attribute   | Type   |
|-------------|--------|
| timestamp   | string |
| status      | string |
| description | string |
| location    | string |
| rawStatus   | string |

**Error Conditions:**

| Status | Condition                                       |
|--------|-------------------------------------------------|
| 403    | Package is private and requester is not owner   |
| 404    | Package not found                               |

---

### PATCH /api/v1/packages/{id}

Updates a package. Only the package creator can update.

**Request:** JSON:API `packages` resource.

| Attribute | Type    | Required |
|-----------|---------|----------|
| label     | string  | no       |
| notes     | string  | no       |
| carrier   | string  | no       |
| private   | boolean | no       |

**Response:** JSON:API `packages` resource.

**Error Conditions:**

| Status | Condition                    |
|--------|------------------------------|
| 400    | Invalid carrier value        |
| 403    | Requester is not the owner   |
| 404    | Package not found            |

---

### DELETE /api/v1/packages/{id}

Deletes a package. Only the package creator can delete.

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition                    |
|--------|------------------------------|
| 403    | Requester is not the owner   |
| 404    | Package not found            |

---

### POST /api/v1/packages/{id}/archive

Archives a package. Sets status to `archived` with timestamp.

**Response:** JSON:API `packages` resource.

**Error Conditions:**

| Status | Condition         |
|--------|-------------------|
| 404    | Package not found |

---

### POST /api/v1/packages/{id}/unarchive

Restores an archived package to `delivered` status.

**Response:** JSON:API `packages` resource.

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | Package is not archived |
| 404    | Package not found      |

---

### POST /api/v1/packages/{id}/refresh

Triggers a manual carrier API poll for the package. Rate-limited to once per 5 minutes.

**Response:** JSON:API `packages` resource with updated data.

**Error Conditions:**

| Status | Condition                                           |
|--------|-----------------------------------------------------|
| 404    | Package not found                                   |
| 429    | Refresh cooldown not elapsed (retryAfterSeconds: 300) |

---

### GET /api/v1/packages/summary

Returns aggregate counts for the current household.

**Response:** JSON:API `packageSummaries` resource.

| Attribute          | Type |
|--------------------|------|
| arrivingTodayCount | int  |
| inTransitCount     | int  |
| exceptionCount     | int  |

---

### GET /api/v1/packages/carriers/detect

Detects the carrier for a tracking number using regex pattern matching.

**Parameters:**

| Name           | In    | Type   | Required |
|----------------|-------|--------|----------|
| trackingNumber | query | string | yes      |

**Response:** JSON:API `carrierDetections` resource.

| Attribute       | Type   |
|-----------------|--------|
| trackingNumber  | string |
| detectedCarrier | string |
| confidence      | string |

Confidence values: `high`, `medium`, `low`.

**Error Conditions:**

| Status | Condition                          |
|--------|------------------------------------|
| 400    | Missing trackingNumber parameter   |
