# Domain

## Package

### Responsibility

Tracks a single package shipment within a household. Manages tracking number, carrier, status lifecycle, privacy, and delivery estimates.

### Core Models

**Model** (`tracking.Model`)

| Field              | Type        |
|--------------------|-------------|
| id                 | uuid.UUID   |
| tenantID           | uuid.UUID   |
| householdID        | uuid.UUID   |
| userID             | uuid.UUID   |
| trackingNumber     | string      |
| carrier            | string      |
| label              | *string     |
| notes              | *string     |
| status             | string      |
| private            | bool        |
| estimatedDelivery  | *time.Time  |
| actualDelivery     | *time.Time  |
| lastPolledAt       | *time.Time  |
| lastStatusChangeAt | *time.Time  |
| archivedAt         | *time.Time  |
| createdAt          | time.Time   |
| updatedAt          | time.Time   |

All fields on Model are immutable after construction. Access is through getter methods.

### Invariants

- One package per tracking number per household (unique constraint).
- Carrier values: `usps`, `ups`, `fedex`.
- Status values: `pre_transit`, `in_transit`, `out_for_delivery`, `delivered`, `exception`, `stale`, `archived`.
- Default status on creation is `pre_transit`.
- Tracking number and carrier are required.
- Maximum active (non-archived) packages per household is configurable (default 25).
- Private packages are redacted for non-owners in API responses.
- Only the package creator can update, delete, or toggle privacy.

### State Transitions

```
pre_transit → in_transit → out_for_delivery → delivered → archived
                                            → exception
Any active status → stale (after N days with no status change)
archived → delivered (via unarchive)
```

- Status transitions are driven by carrier API polling results.
- Stale transition occurs after a configurable number of days (default 14) with no status change.
- Archive transition occurs automatically for delivered packages after a configurable number of days (default 7).
- Hard delete occurs for archived packages after a configurable number of days (default 30).

### Processors

**Processor** (`tracking.Processor`)

| Method                                              | Description                                                  |
|-----------------------------------------------------|--------------------------------------------------------------|
| `ByIDProvider(id)`                                  | Returns a provider for a single package                      |
| `Create(tenantID, householdID, userID, attrs)`      | Creates a package with duplicate and limit checks, triggers initial poll |
| `Get(id)`                                           | Returns a single package by ID                               |
| `GetTrackingEvents(packageID)`                      | Returns tracking event models for a package                  |
| `List(householdID, includeArchived, statuses, ...)`  | Lists packages with optional status and ETA filters          |
| `Update(id, userID, attrs)`                         | Updates label, notes, carrier, or privacy (owner only)       |
| `Delete(id, userID)`                                | Deletes a package (owner only)                               |
| `Archive(id)`                                       | Sets status to archived with timestamp                       |
| `Unarchive(id)`                                     | Restores archived package to delivered status                |
| `Refresh(id)`                                       | Triggers a manual carrier poll with 5-minute cooldown        |
| `Summary(householdID)`                              | Returns counts: arriving today, in transit, exceptions       |
| `PollEntity(entity)`                                | Polls carrier API and updates entity (used by background worker) |

---

## Tracking Event

### Responsibility

Stores individual tracking events received from carrier APIs. Append-only audit log per package.

### Core Models

**Model** (`trackingevent.Model`)

| Field       | Type       |
|-------------|------------|
| id          | uuid.UUID  |
| packageID   | uuid.UUID  |
| timestamp   | time.Time  |
| status      | string     |
| description | string     |
| location    | *string    |
| rawStatus   | *string    |
| createdAt   | time.Time  |

All fields on Model are immutable after construction. Access is through getter methods.

### Invariants

- Events are append-only; never updated or individually deleted.
- Deduplicated by (package_id, timestamp, description).
- Description and status are required.
- Events are ordered by timestamp descending when queried.

---

## Carrier Detection

### Responsibility

Detects the carrier for a tracking number using regex pattern matching.

- UPS: starts with `1Z` followed by 16 alphanumeric characters.
- FedEx: 12, 15, or 20 digit numeric tracking numbers.
- USPS: 20-22 digit numeric or 13-character alphanumeric tracking numbers.
- Confidence levels: `high` (single match), `medium` (multiple matches), `low` (no match).

---

## Background Polling

### Responsibility

Periodically polls carrier APIs for status updates on active packages.

- Runs at a configurable interval (default: 30 minutes).
- Urgent interval (default: 15 minutes) for packages with `out_for_delivery` status.
- Respects per-carrier daily rate budgets (USPS: 1000, UPS: 250, FedEx: 500).
- Updates package status, estimated delivery, and actual delivery from carrier responses.
- Appends new tracking events from carrier responses.
- Uses `WithoutTenantFilter` context for cross-tenant polling.

---

## Background Cleanup

### Responsibility

Manages package lifecycle transitions on a daily schedule.

- Marks packages as `stale` after a configurable number of days (default 14) with no status change.
- Auto-archives `delivered` packages after a configurable number of days (default 7).
- Hard-deletes `archived` packages after a configurable number of days (default 30).

---

## Carrier Clients

### Responsibility

Implements the `CarrierClient` interface for USPS, UPS, and FedEx carrier tracking APIs.

- All three carriers use OAuth 2.0 client credentials flow for authentication.
- OAuth tokens are cached with thread-safe refresh and 60-second expiry buffer.
- Per-carrier daily rate budgets are enforced before each API call.
- Carrier-specific response formats are normalized into a common `TrackingResult` with status, ETA, and events.
- A shared HTTP client with 15-second timeout is injected into all carrier clients.
