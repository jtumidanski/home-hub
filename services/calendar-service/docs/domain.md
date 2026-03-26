# Domain

## Connection

### Responsibility

Manages Google Calendar OAuth connections per user per household. Handles OAuth token lifecycle, connection status tracking, and user color assignment.

### Core Models

**Model** (`connection.Model`)

| Field              | Type       |
|--------------------|------------|
| id                 | uuid.UUID  |
| tenantID           | uuid.UUID  |
| householdID        | uuid.UUID  |
| userID             | uuid.UUID  |
| provider           | string     |
| status             | string     |
| email              | string     |
| accessToken        | string     |
| refreshToken       | string     |
| tokenExpiry        | time.Time  |
| userDisplayName    | string     |
| userColor          | string     |
| lastSyncAt         | *time.Time |
| lastSyncEventCount | int        |
| createdAt          | time.Time  |
| updatedAt          | time.Time  |

All fields on Model are immutable after construction. Access is through getter methods.

### Invariants

- One connection per provider per user per household (unique constraint).
- Tokens are encrypted with AES-256-GCM before persistence, never exposed via API.
- Status values: `connected`, `disconnected`, `syncing`, `error`.
- Color assigned from 8-color palette based on creation order within household, wrapping on overflow.

### Processors

**Processor** (`connection.Processor`)

| Method                            | Description                                           |
|-----------------------------------|-------------------------------------------------------|
| `ByIDProvider(id)`                | Returns a provider for a single connection            |
| `ByUserAndHousehold(userID, hID)` | Lists connections for a user in a household           |
| `AllConnected()`                  | Lists all connected connections across all tenants    |
| `Create(...)`                     | Creates a new connection with auto-assigned color     |
| `UpdateStatus(id, status)`        | Updates connection status                             |
| `UpdateTokens(id, token, expiry)` | Updates encrypted access token and expiry             |
| `UpdateSyncInfo(id, eventCount)`  | Updates last sync time and event count                |
| `Delete(id)`                      | Deletes a connection                                  |
| `CheckManualSyncAllowed(conn)`    | Checks if manual sync is within 5-minute cooldown     |

---

## Source

### Responsibility

Tracks individual Google Calendars associated with a connection, with visibility toggles for the household calendar view.

### Core Models

**Model** (`source.Model`)

| Field        | Type      |
|--------------|-----------|
| id           | uuid.UUID |
| tenantID     | uuid.UUID |
| householdID  | uuid.UUID |
| connectionID | uuid.UUID |
| externalID   | string    |
| name         | string    |
| primary      | bool      |
| visible      | bool      |
| color        | string    |
| syncToken    | string    |
| createdAt    | time.Time |
| updatedAt    | time.Time |

### Invariants

- One entry per Google Calendar per connection (unique on connection_id + external_id).
- Sync token is per source calendar for incremental sync.
- Visibility toggle controls whether events appear in the household calendar API response.

### Processors

| Method                                       | Description                                    |
|----------------------------------------------|------------------------------------------------|
| `ListByConnection(connectionID)`             | Lists all sources for a connection             |
| `CreateOrUpdate(tenantID, hID, connID, ...)` | Creates or updates a source from Google data   |
| `ToggleVisibility(id, visible)`              | Toggles source visibility                      |
| `UpdateSyncToken(id, syncToken)`             | Updates the incremental sync token             |
| `ClearSyncToken(id)`                         | Clears sync token to force full sync           |
| `DeleteByConnection(connectionID)`           | Deletes all sources for a connection           |

---

## Event

### Responsibility

Stores synced calendar events from external providers, scoped by tenant, household, and user. Provides privacy-masked query results.

### Core Models

**Model** (`event.Model`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantID        | uuid.UUID |
| householdID     | uuid.UUID |
| connectionID    | uuid.UUID |
| sourceID        | uuid.UUID |
| userID          | uuid.UUID |
| externalID      | string    |
| title           | string    |
| description     | string    |
| startTime       | time.Time |
| endTime         | time.Time |
| allDay          | bool      |
| location        | string    |
| visibility      | string    |
| userDisplayName | string    |
| userColor       | string    |
| createdAt       | time.Time |
| updatedAt       | time.Time |

### Invariants

- One event per source per external ID (unique on source_id + external_id).
- Privacy masking applied server-side: private/confidential events show as "Busy" with null description/location for non-owners.
- User display name and color are denormalized from the connection for query efficiency.
- Maximum query range is 90 days.

### Processors

| Method                                       | Description                                      |
|----------------------------------------------|--------------------------------------------------|
| `QueryByHouseholdAndTimeRange(hID, start, end)` | Queries visible events with 90-day range limit |
| `Upsert(entity)`                             | Inserts or updates an event by source+external ID |
| `DeleteBySourceAndExternalIDs(sID, extIDs)`  | Deletes cancelled events                        |
| `DeleteByConnection(connectionID)`           | Deletes all events for a connection              |
| `DeleteBySource(sourceID)`                   | Deletes all events for a source                  |
| `CountByConnection(connectionID)`            | Counts events for a connection                   |

---

## OAuth State

### Responsibility

Temporary storage for OAuth CSRF state validation. Short-lived records (10-minute expiry).

### Core Models

**Model** (`oauthstate.Model`)

| Field       | Type      |
|-------------|-----------|
| id          | uuid.UUID |
| tenantID    | uuid.UUID |
| householdID | uuid.UUID |
| userID      | uuid.UUID |
| redirectURI | string    |
| expiresAt   | time.Time |
| createdAt   | time.Time |

### Processors

| Method                    | Description                                          |
|---------------------------|------------------------------------------------------|
| `Create(tID, hID, uID, redirectURI)` | Creates a new state with 10-minute expiry |
| `ValidateAndConsume(stateID)` | Validates state exists, not expired, then deletes |
| `CleanupExpired()`        | Deletes all expired states                           |

---

## Background Sync

### Responsibility

Periodically syncs events from Google Calendar for all connected connections.

- Runs at a configurable interval (default: 15 minutes).
- Stagers sync operations with 0-60 second random jitter per connection.
- Refreshes access tokens when expired.
- Refreshes Google Calendar list on each cycle.
- Uses incremental sync tokens per source calendar.
- Falls back to full sync when sync token is invalidated.
- Marks connections as disconnected on token refresh failure.
- Retries on Google API 429/5xx with exponential backoff (1s, 2s, 4s, max 3 retries).
- Cleans up expired OAuth states on each cycle.
- Stops cleanly on context cancellation (graceful shutdown).

---

## Token Encryption

### Responsibility

Encrypts and decrypts OAuth tokens using AES-256-GCM with a 32-byte key provided via environment variable.

- Nonces are randomly generated per encryption operation.
- Ciphertext is base64-encoded for database storage.
- Key is decoded from base64 on startup.
