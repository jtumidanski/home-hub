# Household Package Tracking — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-26

---

## 1. Overview

Home Hub users manage busy households where multiple packages from different carriers arrive throughout the week. Currently there is no way to see what's coming, when it's expected, or whether something is delayed — users must check each carrier's website individually.

This feature introduces a new `package-service` that lets household members add tracking numbers, automatically polls carrier APIs for status and ETA updates, and surfaces upcoming deliveries in a dedicated list view, on the household calendar, and through a dashboard summary widget. The service integrates directly with free-tier APIs from USPS, UPS, and FedEx — no paid aggregator required.

Packages are household-scoped with an optional privacy flag, mirroring the calendar-service's handling of private events. Delivered packages are auto-archived and eventually purged to keep the list focused on what matters.

## 2. Goals

Primary goals:
- Allow household members to track packages across USPS, UPS, and FedEx from a single view
- Automatically poll carrier APIs for status updates and estimated delivery dates
- Surface delivery ETAs on the task-009 calendar view as a frontend overlay
- Provide a dashboard summary of in-transit and arriving-today packages
- Respect household privacy with a per-package "private" flag

Non-goals:
- Email or SMS auto-import of tracking numbers
- Push notifications or real-time carrier webhooks
- Outbound shipment / label creation
- Carrier support beyond USPS, UPS, and FedEx in v1
- Historical delivery analytics
- Proof-of-delivery photo storage

## 3. User Stories

- As a household member, I want to add a tracking number so that I can monitor a package's status without visiting the carrier's website
- As a household member, I want the carrier to be auto-detected from the tracking number format so that I don't have to look it up myself
- As a household member, I want to see a list of all in-transit packages with their current status and ETA so that I know what's arriving and when
- As a household member, I want to see other members' packages (unless marked private) so that I can receive deliveries on their behalf
- As a household member, I want to mark a package as private so that other members only see a "Package" placeholder without details
- As a household member, I want delivery ETAs to appear on the household calendar so that I can plan around arrivals
- As a household member, I want a dashboard widget showing today's deliveries and in-transit count so that I have a quick daily overview
- As a household member, I want delivered packages to auto-archive so that the active list stays focused on pending deliveries
- As a household member, I want to manually override the auto-detected carrier if detection was wrong

## 4. Functional Requirements

### 4.1 Package Entry

- Users add packages via a form with: tracking number (required), carrier (auto-populated, editable), label/description (optional), notes (optional, freeform text), private flag (default false)
- Carrier auto-detection rules:
  - **UPS**: starts with `1Z` followed by 16 alphanumeric characters
  - **FedEx**: 12, 15, or 20 digits
  - **USPS**: 20-22 digits, or starts with specific prefixes (e.g., `94`, `92`, `93`, `70`, `23`)
  - If ambiguous or unrecognized, prompt user to select carrier manually
- Frontend calls `GET /api/v1/packages/carriers/detect` on tracking number input blur/paste to auto-populate the carrier field; user can override before submitting
- On submission, the service immediately performs an initial tracking poll before responding
  - If the carrier returns "not found," the package is still created as `pre_transit` (carrier may not have scanned it yet) with a warning surfaced to the user: "Tracking number not yet recognized by carrier — it may update once the package is scanned"
- Duplicate tracking number within the same household is rejected

### 4.2 Status Polling

- Background polling runs on a configurable interval (default: 30 minutes)
- Polling frequency adapts based on package state:
  - `pre_transit` / `in_transit`: poll every 30 minutes
  - `out_for_delivery`: poll every 15 minutes
  - `delivered` / `exception`: stop polling
- Each carrier client normalizes responses into a common status model (see Data Model)
- Rate limiting: respect carrier API limits; implement per-carrier request budgets
  - USPS: no hard limit (stay under 1000/day for safety)
  - UPS: 250 requests/day budget
  - FedEx: 500 requests/day budget
- OAuth tokens are cached and refreshed before expiry
- Failed polls are retried with exponential backoff (max 3 retries per cycle)
- After 14 consecutive days with no status change, mark package as `stale` and stop polling

### 4.3 Package Lifecycle

| Status | Description |
|--------|-------------|
| `pre_transit` | Label created, not yet picked up by carrier |
| `in_transit` | In carrier network, moving toward destination |
| `out_for_delivery` | On delivery vehicle |
| `delivered` | Successfully delivered |
| `exception` | Delivery issue (held, returned, damaged, etc.) |
| `stale` | No update for 14+ days — polling stopped |
| `archived` | Auto-archived after delivery |

- Delivered packages transition to `archived` after N days (configurable, default 7)
- Archived packages are hard-deleted after M days (configurable, default 30)
- A daily cleanup job handles archive transitions and deletions
- Users can manually archive or delete packages at any time
- Users can unarchive packages (moves back to `delivered` status)

### 4.4 Package List View

- Route: `/app/packages`
- Default view shows active packages (non-archived) sorted by ETA ascending (soonest first)
- Toggle to show archived packages
- Each package card displays: carrier icon, tracking number (truncated), label, notes (if present), status badge, ETA (or delivered date), last updated timestamp, added-by user
- Private packages from other members show as "Package" with carrier icon and ETA only — no tracking number, label, notes, or status details
- Click to expand: full tracking event history, link to carrier website
- Quick actions: archive, delete, toggle privacy, edit label/notes

### 4.5 Calendar Overlay

- The frontend queries `GET /api/v1/packages?filter[status]=pre_transit,in_transit,out_for_delivery&filter[hasEta]=true` alongside calendar-service events
- Packages with an ETA render as all-day events on the calendar at the estimated delivery date
- Display format: carrier icon + label (or "Package" if private to viewer) + status badge
- If ETA changes (carrier updates), the calendar event moves automatically on next data fetch
- Clicking a calendar package event navigates to the package detail in `/app/packages`
- Calendar overlay uses a distinct visual style (e.g., dashed border, different color band) to differentiate from calendar events

### 4.6 Dashboard Summary

- New summary endpoint returns:
  - `arrivingTodayCount`: packages with ETA = today
  - `inTransitCount`: packages in `pre_transit`, `in_transit`, or `out_for_delivery`
  - `exceptionCount`: packages in `exception` status
- Dashboard widget displays these counts with appropriate icons
- Counts respect privacy: private packages from other members are included in counts but not individually listed

## 5. API Surface

Base path: `/api/v1/packages`

### Endpoints

#### `POST /api/v1/packages`
Create a new tracked package.

**Request:**
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

**Response:** `201 Created` — full package resource including initial tracking status

**Errors:**
- `409 Conflict` — duplicate tracking number in household
- `422 Unprocessable Entity` — invalid tracking number format or unsupported carrier

#### `GET /api/v1/packages`
List packages for the household.

**Query parameters:**
- `filter[status]` — comma-separated status values (default: all non-archived)
- `filter[carrier]` — filter by carrier
- `filter[hasEta]` — `true` to return only packages with an ETA
- `filter[archived]` — `true` to include archived packages
- `sort` — `eta`, `-eta`, `createdAt`, `-createdAt` (default: `eta`)

**Response:** `200 OK` — array of package resources. Private packages from other members return redacted attributes.

#### `GET /api/v1/packages/{id}`
Get a single package with full tracking event history.

**Response:** `200 OK` — package resource with included `trackingEvents` relationship.
Private packages from other members return `403 Forbidden`.

#### `PATCH /api/v1/packages/{id}`
Update package attributes (label, notes, carrier, private flag).

**Request:**
```json
{
  "data": {
    "type": "packages",
    "id": "uuid",
    "attributes": {
      "label": "Updated label",
      "notes": "Updated notes",
      "private": true
    }
  }
}
```

**Response:** `200 OK` — updated package resource.
Only the user who created the package can update it.

#### `DELETE /api/v1/packages/{id}`
Delete a package. Only the creator or a household admin/owner can delete.

**Response:** `204 No Content`

#### `POST /api/v1/packages/{id}/archive`
Manually archive a package.

**Response:** `200 OK` — updated package resource with `archived` status.

#### `POST /api/v1/packages/{id}/unarchive`
Restore an archived package to `delivered` status.

**Response:** `200 OK` — updated package resource.

#### `POST /api/v1/packages/{id}/refresh`
Trigger an immediate tracking poll for a single package. Rate-limited to once per 5 minutes per package.

**Response:** `200 OK` — updated package resource with latest tracking data.

**Errors:**
- `429 Too Many Requests` — if refreshed within the last 5 minutes

#### `GET /api/v1/packages/summary`
Dashboard summary for the household.

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

#### `GET /api/v1/packages/carriers/detect`
Detect carrier from a tracking number without creating a package.

**Query parameters:**
- `trackingNumber` — the tracking number to analyze

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

**Confidence levels:** `high` (single match), `medium` (matches multiple patterns), `low` (no match — user must select manually)

## 6. Data Model

### packages

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | |
| tenant_id | UUID | NOT NULL, INDEX | Tenant scope |
| household_id | UUID | NOT NULL, INDEX | Household scope |
| user_id | UUID | NOT NULL | User who added the package |
| tracking_number | VARCHAR(64) | NOT NULL | Carrier tracking number |
| carrier | VARCHAR(16) | NOT NULL | `usps`, `ups`, `fedex` |
| label | VARCHAR(255) | | Optional user description |
| notes | TEXT | | Freeform notes (e.g., delivery instructions) |
| status | VARCHAR(24) | NOT NULL, DEFAULT 'pre_transit' | Current package status |
| private | BOOLEAN | NOT NULL, DEFAULT false | Hide details from other members |
| estimated_delivery | DATE | | Latest ETA from carrier |
| actual_delivery | TIMESTAMPTZ | | When delivered |
| last_polled_at | TIMESTAMPTZ | | Last successful carrier poll |
| last_status_change_at | TIMESTAMPTZ | | For stale detection |
| archived_at | TIMESTAMPTZ | | When auto/manually archived |
| created_at | TIMESTAMPTZ | NOT NULL | |
| updated_at | TIMESTAMPTZ | NOT NULL | |

**Indexes:**
- `(tenant_id, household_id, status)` — primary list query
- `(tenant_id, household_id, tracking_number)` UNIQUE — duplicate prevention
- `(status, last_polled_at)` — polling scheduler query
- `(status, archived_at)` — cleanup job query

### tracking_events

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | |
| package_id | UUID | NOT NULL, FK → packages, INDEX | |
| timestamp | TIMESTAMPTZ | NOT NULL | Event time from carrier |
| status | VARCHAR(24) | NOT NULL | Normalized status at this point |
| description | VARCHAR(512) | NOT NULL | Human-readable event description |
| location | VARCHAR(255) | | City, State or facility name |
| raw_status | VARCHAR(128) | | Original carrier status code |
| created_at | TIMESTAMPTZ | NOT NULL | |

**Indexes:**
- `(package_id, timestamp DESC)` — event history query

### carrier_tokens (in-memory or encrypted storage)

OAuth tokens for each carrier API are stored encrypted in the database or environment. They are short-lived and refreshed automatically.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | |
| carrier | VARCHAR(16) | NOT NULL, UNIQUE | |
| access_token | TEXT | NOT NULL | Encrypted |
| expires_at | TIMESTAMPTZ | NOT NULL | |
| created_at | TIMESTAMPTZ | NOT NULL | |
| updated_at | TIMESTAMPTZ | NOT NULL | |

## 7. Service Impact

### New: package-service

- Full new microservice following existing patterns (cmd/, internal/, etc.)
- Owns package CRUD, carrier API clients, background polling, cleanup jobs
- Carrier clients: USPS Tracking API v3, UPS Tracking API v1, FedEx Track API v1
- OAuth 2.0 client credentials flow for all three carriers
- Background workers: polling scheduler, archive/delete cleanup
- PostgreSQL schema: `package`

### Frontend

- New sidebar nav entry "Packages" with badge (inTransitCount)
- New route `/app/packages` with list and detail views
- Package entry form with carrier auto-detect
- Calendar overlay: query package-service for packages with ETAs, render as styled all-day events alongside calendar-service events
- Dashboard summary card for package counts
- New React Query hooks: `usePackages`, `usePackage`, `useCreatePackage`, `useUpdatePackage`, `useDeletePackage`, `useArchivePackage`, `useUnarchivePackage`, `usePackageSummary`, `useRefreshPackage`, `useDetectCarrier`

### Infrastructure

- New `package-service` container in docker-compose and k8s
- Nginx route: `/api/v1/packages/*` → package-service
- New database: `package_service` (or schema within shared PG instance per project pattern)
- Environment variables for carrier API credentials (see Non-Functional Requirements)

### No changes required to existing services

Calendar integration is frontend-only. No cross-service API calls.

## 8. Non-Functional Requirements

### Performance
- Background polling must not exceed carrier rate limits
- Package list query should respond in < 200ms for up to 100 active packages per household
- Initial tracking poll on package creation should complete within 10 seconds (with timeout)

### Security
- Carrier API credentials stored as secrets (env vars, k8s secrets) — never logged
- OAuth tokens encrypted at rest if persisted to database
- Private packages enforce access control at the API layer — redacted in list responses, 403 on detail for non-owners
- Tracking numbers treated as PII-adjacent — not included in logs

### Observability
- Structured JSON logging with request_id, user_id, tenant_id, household_id (per project standard)
- OpenTelemetry tracing on API endpoints and carrier API calls
- Metrics: carrier API latency, poll success/failure rate, active package count per household

### Multi-tenancy
- All queries scoped by tenant_id and household_id
- Carrier API credentials are system-wide (not per-tenant) — shared infrastructure

### Configuration (Environment Variables)

| Variable | Description | Default |
|----------|-------------|---------|
| `USPS_CLIENT_ID` | USPS API OAuth client ID | required |
| `USPS_CLIENT_SECRET` | USPS API OAuth client secret | required |
| `UPS_CLIENT_ID` | UPS API OAuth client ID | required |
| `UPS_CLIENT_SECRET` | UPS API OAuth client secret | required |
| `FEDEX_API_KEY` | FedEx API key | required |
| `FEDEX_SECRET_KEY` | FedEx API secret key | required |
| `PACKAGE_POLL_INTERVAL` | Default polling interval | `30m` |
| `PACKAGE_POLL_INTERVAL_URGENT` | Polling interval for out-for-delivery | `15m` |
| `PACKAGE_ARCHIVE_AFTER_DAYS` | Days after delivery to auto-archive | `7` |
| `PACKAGE_DELETE_AFTER_DAYS` | Days after archive to hard-delete | `30` |
| `PACKAGE_STALE_AFTER_DAYS` | Days with no update before marking stale | `14` |
| `PACKAGE_MAX_ACTIVE_PER_HOUSEHOLD` | Max active (non-archived) packages per household | `25` |

## 9. Resolved Design Decisions

1. **Carrier expansion model** — Use a `CarrierClient` Go interface (provider pattern) from the start. Each carrier implements `Track(trackingNumber) → TrackingResult`. This is zero extra cost since the three carrier APIs need normalization anyway, and it makes adding carriers later (DHL, Amazon Logistics) straightforward.
2. **Tracking number validation** — The initial poll on creation attempts a carrier API call. If the carrier returns "not found," the package is still created as `pre_transit` (the carrier may not have scanned it yet) with a warning surfaced to the user. This avoids rejecting valid pre-ship labels.
3. **Household package limits** — Max 25 active (non-archived) packages per household (configurable via `PACKAGE_MAX_ACTIVE_PER_HOUSEHOLD`). This keeps all three carriers well within their daily rate budgets even with aggressive polling. `POST /api/v1/packages` returns `422` with a clear message when the limit is reached.

## 10. Open Questions

None — all design questions have been resolved.

## 11. Acceptance Criteria

- [ ] User can add a package by entering a tracking number; carrier is auto-detected on input blur/paste
- [ ] User can manually select/override the carrier
- [ ] User can add an optional label, notes, and set the package as private
- [ ] If the carrier returns "not found" on initial poll, the package is created as `pre_transit` with a warning shown to the user
- [ ] Duplicate tracking numbers within a household are rejected
- [ ] Creating a package when household is at the 25-package active limit returns 422 with a clear message
- [ ] Package list shows all active packages sorted by ETA with status, carrier, label, and notes
- [ ] Private packages from other members show redacted information ("Package" placeholder)
- [ ] Carrier clients implement a common `CarrierClient` interface (provider pattern)
- [ ] Background polling updates package status and ETA on the configured interval
- [ ] Polling frequency increases for out-for-delivery packages
- [ ] Polling stops for delivered, exception, and stale packages
- [ ] Delivered packages auto-archive after the configured number of days
- [ ] Archived packages are hard-deleted after the configured retention period
- [ ] Users can manually archive, unarchive, and delete packages
- [ ] Package ETAs appear as all-day events on the calendar view
- [ ] Calendar package events are visually distinct from calendar-service events
- [ ] Dashboard summary shows arriving-today, in-transit, and exception counts
- [ ] Manual refresh triggers an immediate carrier poll (rate-limited to 5 min)
- [ ] Carrier API credentials are stored as secrets and never logged
- [ ] Tracking numbers are not included in application logs
- [ ] All queries are scoped by tenant_id and household_id
- [ ] Service includes structured logging and OpenTelemetry tracing
