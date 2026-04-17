# REST API

All endpoints are prefixed with `/api/v1`. Unless noted otherwise, endpoints require JWT authentication. Request and response bodies use JSON:API format. Tenant and household context are derived from the JWT.

## Endpoints

### GET /api/v1/calendar/connections

Returns calendar connections for the current user in the current household.

**Response:** JSON:API array of `calendar-connections` resources.

| Attribute          | Type    |
|--------------------|---------|
| provider           | string  |
| status             | string  |
| email              | string  |
| userDisplayName    | string  |
| userColor          | string  |
| writeAccess        | boolean |
| lastSyncAt         | string  |
| lastSyncEventCount | int     |
| createdAt          | string  |

---

### POST /api/v1/calendar/connections/google/authorize

Initiates Google Calendar OAuth consent flow. Returns a redirect URL.

**Request:** JSON:API `calendar-authorization-requests` resource.

| Attribute   | Type    | Required |
|-------------|---------|----------|
| redirectUri | string  | yes      |
| reauthorize | boolean | no       |

When `reauthorize` is true, the OAuth flow forces consent to re-acquire tokens (used to upgrade write access).

**Response:** JSON:API `calendar-authorization-responses` resource.

| Attribute    | Type   |
|--------------|--------|
| authorizeUrl | string |

---

### GET /api/v1/calendar/connections/google/callback

Handles the OAuth callback from Google. Not called directly by the frontend. **This endpoint does not require JWT authentication.**

**Parameters:**

| Name  | In    | Type   | Required |
|-------|-------|--------|----------|
| code  | query | string | yes      |
| state | query | string | yes      |

**Response:** 302 redirect to `/app/calendar?connected=true` on success, `/app/calendar?error=...` on failure.

---

### DELETE /api/v1/calendar/connections/{id}

Disconnects a calendar connection. Revokes the Google OAuth token and deletes all events and sources.

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition                         |
|--------|-----------------------------------|
| 403    | Connection belongs to another user |
| 404    | Connection not found              |

---

### POST /api/v1/calendar/connections/{id}/sync

Triggers an immediate sync for a connection. Rate-limited to once per 5 minutes.

**Response:** JSON:API `calendar-connections` resource with updated status.

**Error Conditions:**

| Status | Condition                                         |
|--------|---------------------------------------------------|
| 403    | Connection belongs to another user                |
| 404    | Connection not found                              |
| 429    | Rate limited (Retry-After: 300)                   |

---

### GET /api/v1/calendar/connections/{id}/calendars

Lists Google Calendars for a connection with visibility toggle state.

**Response:** JSON:API array of `calendar-sources` resources.

| Attribute | Type    |
|-----------|---------|
| name      | string  |
| primary   | boolean |
| visible   | boolean |
| color     | string  |

**Error Conditions:**

| Status | Condition                         |
|--------|-----------------------------------|
| 403    | Connection belongs to another user |
| 404    | Connection not found              |

---

### PATCH /api/v1/calendar/connections/{id}/calendars/{calendarId}

Toggles a Google Calendar's visibility on the household calendar.

**Request:** JSON:API `calendar-sources` resource.

| Attribute | Type    | Required |
|-----------|---------|----------|
| visible   | boolean | yes      |

**Response:** JSON:API `calendar-sources` resource with updated state.

**Error Conditions:**

| Status | Condition                         |
|--------|-----------------------------------|
| 403    | Connection belongs to another user |
| 404    | Connection or source not found    |

---

### GET /api/v1/calendar/events

Returns calendar events for the household within a time range. Merges events from all connected users. Privacy masking applied server-side.

The client computes both bounds from its local timezone and sends them as required RFC 3339 timestamps. The server does not infer a default range from a header or its own clock.

**Parameters:**

| Name  | In    | Type   | Required | Format    |
|-------|-------|--------|----------|-----------|
| start | query | string | yes      | RFC 3339  |
| end   | query | string | yes      | RFC 3339  |

**Constraints:** `end` must be strictly after `start`. Maximum range between start and end is 90 days.

**Error Conditions:**

| Status | Condition                                                       |
|--------|-----------------------------------------------------------------|
| 400    | Missing, empty, malformed `start`/`end`; `end` not after `start` |

**Response:** JSON:API array of `calendar-events` resources.

| Attribute       | Type    |
|-----------------|---------|
| title           | string  |
| description     | string  |
| startTime       | string  |
| endTime         | string  |
| allDay          | boolean |
| location        | string  |
| visibility      | string  |
| userDisplayName | string  |
| userColor       | string  |
| isOwner         | boolean |
| sourceId        | string  |
| connectionId    | string  |
| isRecurring     | boolean |

**Privacy Rules:**

- If visibility is `private` or `confidential` and requester is not the event owner: title replaced with "Busy", description and location set to null.
- `isOwner` indicates whether the requesting user owns the event.
- `userDisplayName` and `userColor` are always visible.

**Sorting:** All-day events first, then by startTime ascending.

**Error Conditions:**

| Status | Condition                     |
|--------|-------------------------------|
| 400    | Invalid date format or range exceeds 90 days |

---

### POST /api/v1/calendar/connections/{connId}/calendars/{calId}/events

Creates a new event on Google Calendar. Triggers a sync after creation.

**Request:** JSON:API `calendar-events` resource.

| Attribute   | Type     | Required |
|-------------|----------|----------|
| title       | string   | yes      |
| start       | string   | yes      |
| end         | string   | no       |
| allDay      | boolean  | no       |
| location    | string   | no       |
| description | string   | no       |
| recurrence  | string[] | no       |

For timed events, `start` and `end` use RFC 3339 format. If `end` is omitted, defaults to one hour after `start`. For all-day events, `start` and `end` use `YYYY-MM-DD` format.

**Response:** 201 Created (no body).

**Error Conditions:**

| Status | Condition                                      |
|--------|------------------------------------------------|
| 403    | Connection belongs to another user             |
| 403    | Connection does not have write access           |
| 403    | Google Calendar denied write (e.g. read-only calendar) |
| 404    | Connection or calendar not found               |
| 422    | Title is required                              |

---

### PATCH /api/v1/calendar/connections/{connId}/events/{eventId}

Updates an event on Google Calendar. Triggers a sync after update.

**Request:** JSON:API `calendar-events` resource. All fields are optional (partial update).

| Attribute   | Type    | Required |
|-------------|---------|----------|
| title       | string  | no       |
| start       | string  | no       |
| end         | string  | no       |
| allDay      | boolean | no       |
| location    | string  | no       |
| description | string  | no       |
| scope       | string  | no       |

When `scope` is `"all"` and the event is a recurring instance, the update targets the base recurring event.

**Response:** JSON:API `calendar-events` resource with updated state and privacy masking.

**Error Conditions:**

| Status | Condition                               |
|--------|-----------------------------------------|
| 403    | Connection belongs to another user      |
| 403    | Connection does not have write access    |
| 403    | Event does not belong to this connection |
| 404    | Connection or event not found           |

---

### DELETE /api/v1/calendar/connections/{connId}/events/{eventId}

Deletes an event on Google Calendar. Triggers a sync after deletion.

**Parameters:**

| Name  | In    | Type   | Required | Description                                        |
|-------|-------|--------|----------|----------------------------------------------------|
| scope | query | string | no       | `"all"` to delete entire recurring series           |

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition                               |
|--------|-----------------------------------------|
| 403    | Connection belongs to another user      |
| 403    | Connection does not have write access    |
| 403    | Event does not belong to this connection |
| 404    | Connection or event not found           |
