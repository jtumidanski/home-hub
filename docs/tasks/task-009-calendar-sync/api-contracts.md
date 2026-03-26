# Household Calendar — API Contracts

## Base Path

All endpoints are prefixed with `/api/v1/calendar`.

## Authentication

All endpoints require a valid JWT access token (cookie-based, same as other services). JWT is validated via JWKS from auth-service.

## Headers

| Header | Required | Description |
|--------|----------|-------------|
| `X-Tenant-ID` | Yes | Tenant scope |
| `X-Household-ID` | Yes | Household scope |

---

## Connections

### List Connections

Returns calendar connections for the current user in the current household.

```
GET /api/v1/calendar/connections
```

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "type": "calendar-connections",
      "attributes": {
        "provider": "google",
        "status": "connected",
        "email": "user@gmail.com",
        "lastSyncAt": "2026-03-26T10:30:00Z",
        "lastSyncEventCount": 42,
        "createdAt": "2026-03-20T08:00:00Z"
      }
    }
  ]
}
```

**Connection statuses:** `connected`, `disconnected`, `syncing`, `error`

---

### List Calendars for Connection

Returns the Google Calendars available for a connection, with their sync/visibility toggle state.

```
GET /api/v1/calendar/connections/{id}/calendars
```

**Response 200:**

```json
{
  "data": [
    {
      "id": "google-calendar-id",
      "type": "calendar-sources",
      "attributes": {
        "name": "Personal",
        "primary": true,
        "visible": true,
        "color": "#4285F4"
      }
    },
    {
      "id": "google-calendar-id-2",
      "type": "calendar-sources",
      "attributes": {
        "name": "Work",
        "primary": false,
        "visible": true,
        "color": "#EA4335"
      }
    },
    {
      "id": "google-calendar-id-3",
      "type": "calendar-sources",
      "attributes": {
        "name": "Holidays in United States",
        "primary": false,
        "visible": false,
        "color": "#34A853"
      }
    }
  ]
}
```

---

### Toggle Calendar Visibility

Enables or disables a specific Google Calendar from appearing on the household calendar.

```
PATCH /api/v1/calendar/connections/{id}/calendars/{calendarId}
```

**Request body:**

```json
{
  "data": {
    "id": "google-calendar-id",
    "type": "calendar-sources",
    "attributes": {
      "visible": false
    }
  }
}
```

**Response 200:** Returns the updated calendar source resource.

**Error 404:** Connection or calendar not found.

---

### Initiate Google OAuth

Starts the Google Calendar OAuth consent flow. Returns a redirect URL for the frontend to navigate to.

```
POST /api/v1/calendar/connections/google/authorize
```

**Request body:**

```json
{
  "data": {
    "type": "calendar-authorization-requests",
    "attributes": {
      "redirectUri": "https://homehub.example.com/app/calendar/callback"
    }
  }
}
```

**Response 200:**

```json
{
  "data": {
    "id": "state-uuid",
    "type": "calendar-authorization-responses",
    "attributes": {
      "authorizeUrl": "https://accounts.google.com/o/oauth2/v2/auth?..."
    }
  }
}
```

The `id` is the OAuth state parameter, stored server-side for CSRF validation.

---

### Google OAuth Callback

Handles the OAuth callback from Google. Exchanges the authorization code for tokens and creates the connection.

```
GET /api/v1/calendar/connections/google/callback?code={code}&state={state}
```

**Response 302:** Redirects to `/app/calendar?connected=true` on success, `/app/calendar?error=auth_failed` on failure.

The callback validates the `state` parameter against the stored state, exchanges the code for tokens, stores the tokens, creates a `calendar_connection` record, fetches the user's Google Calendar list to populate `calendar_sources`, and triggers an immediate async first sync. The user display name is extracted from JWT claims in the request cookie and stored on the connection.

---

### Disconnect Calendar

Disconnects a calendar connection. Revokes the Google OAuth token and deletes stored events for this connection.

```
DELETE /api/v1/calendar/connections/{id}
```

**Response 204:** No content.

**Error 404:** Connection not found or not owned by the current user.

---

### Trigger Manual Sync

Triggers an immediate sync for a specific connection. Returns the connection with updated sync status.

```
POST /api/v1/calendar/connections/{id}/sync
```

**Response 200:**

```json
{
  "data": {
    "id": "uuid",
    "type": "calendar-connections",
    "attributes": {
      "provider": "google",
      "status": "syncing",
      "email": "user@gmail.com",
      "lastSyncAt": "2026-03-26T10:30:00Z",
      "lastSyncEventCount": 42,
      "createdAt": "2026-03-20T08:00:00Z"
    }
  }
}
```

**Error 404:** Connection not found or not owned by the current user.

**Error 429:** Manual sync rate-limited — one sync per 5 minutes per connection. Response includes `Retry-After` header.

---

## Events

### List Events

Returns calendar events for the household within a time range. Merges events from all connected users in the household.

```
GET /api/v1/calendar/events?start={ISO8601}&end={ISO8601}
```

**Query parameters:**

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `start` | No | Start of current day | Start of time range (ISO 8601) |
| `end` | No | 7 days from start | End of time range (ISO 8601) |
| `timezone` | No | Household timezone | IANA timezone for time rendering context (informational — all times returned in UTC, frontend converts) |

**Constraints:**

- Maximum range between `start` and `end` is 90 days. Requests exceeding this return 400.

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "type": "calendar-events",
      "attributes": {
        "title": "Team Standup",
        "description": "Daily sync meeting",
        "startTime": "2026-03-26T09:00:00Z",
        "endTime": "2026-03-26T09:30:00Z",
        "allDay": false,
        "location": "Zoom",
        "visibility": "default",
        "userDisplayName": "Jane Doe",
        "userColor": "#4285F4",
        "isOwner": true
      }
    },
    {
      "id": "uuid",
      "type": "calendar-events",
      "attributes": {
        "title": "Busy",
        "description": null,
        "startTime": "2026-03-26T12:00:00Z",
        "endTime": "2026-03-26T13:00:00Z",
        "allDay": false,
        "location": null,
        "visibility": "private",
        "userDisplayName": "John Doe",
        "userColor": "#EA4335",
        "isOwner": false
      }
    }
  ]
}
```

**Privacy rules applied server-side:**

- If `visibility` is `private` or `confidential` AND the requesting user is NOT the event owner:
  - `title` is replaced with `"Busy"`
  - `description` is set to `null`
  - `location` is set to `null`
- `isOwner` indicates whether the requesting user owns the event
- `userDisplayName` and `userColor` are always visible regardless of privacy

**Sorting:** Events are sorted by `startTime` ascending, with all-day events first.

---

## Error Responses

All error responses follow JSON:API error format:

```json
{
  "errors": [
    {
      "status": "404",
      "title": "Not Found",
      "detail": "Calendar connection not found"
    }
  ]
}
```

Common error codes:

| Status | When |
|--------|------|
| 400 | Invalid request parameters (bad date format, missing fields) |
| 401 | Missing or invalid JWT |
| 403 | Connection belongs to another user |
| 404 | Resource not found |
| 409 | User already has a Google Calendar connection in this household |
| 429 | Manual sync rate-limited (once per 5 minutes per connection) |
| 500 | Internal error (Google API failure, database error) |
