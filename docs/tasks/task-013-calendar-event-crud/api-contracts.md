# Calendar Event CRUD — API Contracts

## New Endpoints

### Create Event

```
POST /api/v1/calendar/connections/{connectionId}/calendars/{calendarId}/events
```

Creates an event on the specified Google Calendar via the Google Calendar API, then triggers a sync to populate the local database.

**Path Parameters:**
- `connectionId` — UUID of the user's calendar connection
- `calendarId` — UUID of the calendar source (from `GET /connections/{id}/calendars`)

**Request Body (JSON:API):**
```json
{
  "data": {
    "type": "calendar-events",
    "attributes": {
      "title": "Team Standup",
      "start": "2026-03-28T09:00:00-05:00",
      "end": "2026-03-28T09:30:00-05:00",
      "allDay": false,
      "location": "Conference Room A",
      "description": "Daily sync meeting",
      "recurrence": ["RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"]
    }
  }
}
```

**Field Details:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `title` | string | Yes | Max 1024 characters |
| `start` | ISO 8601 datetime or date | Yes | Datetime for timed events, date (`2026-03-28`) for all-day |
| `end` | ISO 8601 datetime or date | No | Defaults to start + 1 hour (timed) or start + 1 day (all-day) |
| `allDay` | boolean | No | Default `false`. When `true`, `start`/`end` are date-only |
| `location` | string | No | Max 1024 characters |
| `description` | string | No | Max 8192 characters |
| `recurrence` | string[] | No | Array of RFC 5545 RRULE strings. Omit for single events |

**Response: `201 Created`**
```json
{
  "data": {
    "type": "calendar-events",
    "id": "event-uuid",
    "attributes": {
      "title": "Team Standup",
      "start": "2026-03-28T09:00:00-05:00",
      "end": "2026-03-28T09:30:00-05:00",
      "allDay": false,
      "location": "Conference Room A",
      "description": "Daily sync meeting",
      "visibility": "default",
      "userDisplayName": "Jane",
      "userColor": "#4285F4",
      "isOwner": true,
      "sourceId": "source-uuid"
    }
  }
}
```

**Error Responses:**

| Status | Code | Condition |
|--------|------|-----------|
| 403 | `WRITE_ACCESS_REQUIRED` | Connection does not have write scope |
| 403 | `NOT_CONNECTION_OWNER` | Authenticated user does not own this connection |
| 403 | `GOOGLE_WRITE_DENIED` | Google API rejected the write (read-only calendar) |
| 404 | `CONNECTION_NOT_FOUND` | Connection does not exist or is not in user's household |
| 404 | `CALENDAR_NOT_FOUND` | Calendar source does not exist for this connection |
| 422 | `VALIDATION_ERROR` | Missing required fields, invalid dates, end before start |

---

### Update Event

```
PATCH /api/v1/calendar/connections/{connectionId}/events/{eventId}
```

Updates an existing event via the Google Calendar API.

**Path Parameters:**
- `connectionId` — UUID of the user's calendar connection
- `eventId` — UUID of the local event record

**Request Body (JSON:API):**
```json
{
  "data": {
    "type": "calendar-events",
    "id": "event-uuid",
    "attributes": {
      "title": "Updated Standup",
      "start": "2026-03-28T10:00:00-05:00",
      "end": "2026-03-28T10:30:00-05:00",
      "allDay": false,
      "location": "Room B",
      "description": "Moved to 10am",
      "scope": "single"
    }
  }
}
```

All attribute fields are optional — only provided fields are updated (partial update).

**Additional Field:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `scope` | string | No | `"single"` (default) or `"all"`. Only relevant for recurring events |

**Scope behavior for recurring events:**
- `"single"` — Updates only this specific instance. Google creates an exception to the recurrence.
- `"all"` — Updates the entire recurring series (all future instances reflect the change).

**Response: `200 OK`** — Same shape as create response with updated fields.

**Error Responses:**

| Status | Code | Condition |
|--------|------|-----------|
| 403 | `WRITE_ACCESS_REQUIRED` | Connection does not have write scope |
| 403 | `NOT_EVENT_OWNER` | Event does not belong to this connection |
| 404 | `CONNECTION_NOT_FOUND` | Connection not found |
| 404 | `EVENT_NOT_FOUND` | Event not found |
| 422 | `VALIDATION_ERROR` | Invalid dates, end before start |

---

### Delete Event

```
DELETE /api/v1/calendar/connections/{connectionId}/events/{eventId}?scope=single
```

Deletes an event via the Google Calendar API.

**Path Parameters:**
- `connectionId` — UUID of the user's calendar connection
- `eventId` — UUID of the local event record

**Query Parameters:**

| Param | Type | Required | Notes |
|-------|------|----------|-------|
| `scope` | string | No | `"single"` (default) or `"all"`. Only relevant for recurring events |

**Scope behavior for recurring events:**
- `"single"` — Deletes only this specific instance.
- `"all"` — Deletes the entire recurring series.

**Response: `204 No Content`**

**Error Responses:**

| Status | Code | Condition |
|--------|------|-----------|
| 403 | `WRITE_ACCESS_REQUIRED` | Connection does not have write scope |
| 403 | `NOT_EVENT_OWNER` | Event does not belong to this connection |
| 404 | `CONNECTION_NOT_FOUND` | Connection not found |
| 404 | `EVENT_NOT_FOUND` | Event not found |

---

## Modified Endpoints

### Get Connections (updated response)

```
GET /api/v1/calendar/connections
```

**New attribute in response items:**
```json
{
  "data": [
    {
      "type": "calendar-connections",
      "id": "conn-uuid",
      "attributes": {
        "provider": "google",
        "email": "jane@gmail.com",
        "displayName": "Jane",
        "status": "connected",
        "color": "#4285F4",
        "lastSyncAt": "2026-03-27T12:00:00Z",
        "eventCount": 42,
        "writeAccess": true
      }
    }
  ]
}
```

The `writeAccess` field indicates whether the connection was authorized with write scope.

---

### Authorize Google (updated request)

```
POST /api/v1/calendar/connections/google/authorize
```

**Updated request body:**
```json
{
  "data": {
    "type": "calendar-connections",
    "attributes": {
      "redirectUri": "https://app.example.com/app/calendar",
      "reauthorize": true
    }
  }
}
```

When `reauthorize` is `true`:
- The generated Google OAuth URL includes `prompt=consent` to force the consent screen
- The `access_type=offline` parameter ensures a new refresh token is issued
- The callback handler updates the existing connection's tokens and sets `write_access = true`

---

### Get Events (updated response)

```
GET /api/v1/calendar/events?start=...&end=...
```

**New attribute in response items:**
```json
{
  "data": [
    {
      "type": "calendar-events",
      "id": "event-uuid",
      "attributes": {
        "title": "Team Standup",
        "start": "2026-03-28T09:00:00-05:00",
        "end": "2026-03-28T09:30:00-05:00",
        "allDay": false,
        "location": "Conference Room A",
        "description": "Daily sync meeting",
        "visibility": "default",
        "userDisplayName": "Jane",
        "userColor": "#4285F4",
        "isOwner": true,
        "sourceId": "source-uuid",
        "connectionId": "conn-uuid",
        "isRecurring": false
      }
    }
  ]
}
```

New fields:
- `sourceId` — UUID of the calendar source, needed for routing create requests
- `connectionId` — UUID of the connection, needed for routing edit/delete requests
- `isRecurring` — Whether this event is part of a recurring series
