# Calendar Event CRUD — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-27
---

## 1. Overview

Home Hub's calendar integration currently provides read-only sync from Google Calendar. Users can view their household's combined calendar but cannot create, edit, or delete events from within Home Hub — they must switch to Google Calendar to make changes and wait for the next sync cycle.

This feature adds full event CRUD (Create, Read, Update, Delete) capabilities to the calendar page. Users will be able to create timed, all-day, and recurring events on their synced Google Calendars directly from the Home Hub UI. All mutations write through to the Google Calendar API, maintaining Google as the authoritative source of truth. After each mutation, an internal sync is triggered to reconcile local state.

This also requires upgrading the Google OAuth scope from `calendar.readonly` to `calendar` (read-write), with a re-authorization flow for existing connections.

## 2. Goals

Primary goals:
- Allow users to create, edit, and delete events on their synced Google Calendars from the Home Hub calendar page
- Support timed events, all-day events, and recurring events
- Maintain Google Calendar as the source of truth — no local-only event data
- Provide a smooth re-authorization flow for existing connections upgrading to write scope

Non-goals:
- Attendee/invitee management
- Other calendar providers (Outlook, Apple, CalDAV)
- Drag-to-create or drag-to-resize on the calendar grid
- Conflict detection or scheduling suggestions
- Event notifications or reminders from Home Hub

## 3. User Stories

- As a household member, I want to create an event on my Google Calendar from the Home Hub calendar page so that I don't have to switch apps
- As a household member, I want to choose which of my visible calendars an event is added to so that events go to the right calendar
- As a household member, I want to edit an event I own (title, time, location, description) so that I can make corrections without leaving Home Hub
- As a household member, I want to delete an event I own so that cancelled plans are removed from the household view
- As a household member, I want to create all-day events so that I can mark holidays, trips, and deadlines
- As a household member, I want to create recurring events so that I can schedule repeating activities
- As an existing user, I want to be prompted to re-authorize my calendar connection so that I gain write access with minimal friction

## 4. Functional Requirements

### 4.1 OAuth Scope Upgrade

- Change the Google OAuth scope from `calendar.readonly` to `calendar` (full read-write access)
- Existing connections must be re-authorized to gain write access
- Add a `scope` or `write_access` field to connections to track whether the connection has write permission
- Display a prompt on the calendar page when a connection lacks write access, with a one-click re-authorize button
- Re-authorization reuses the existing OAuth flow but with `prompt=consent` to force Google's consent screen
- On successful re-authorization, update the connection's tokens and mark write access as granted
- Users who do not re-authorize retain full read-only functionality — no degradation

### 4.2 Event Creation

- "Add Event" button on the calendar page opens a creation dialog/form
- Clicking on an empty time slot on the calendar grid pre-fills the date/time in the creation form
- Required fields: title, start date/time
- Optional fields: end date/time (defaults to 1 hour after start), location, description
- All-day toggle: when enabled, date-only fields replace date/time fields
- Recurrence selector: none, daily, weekly, weekday (Mon-Fri), monthly, yearly (preset options only — no custom RRULE builder)
- Calendar picker: dropdown of the user's visible calendar sources; defaults to primary calendar
- Events are created via Google Calendar API — the local database is never written to directly
- After successful creation, trigger an internal sync for the connection (bypasses 5-minute manual sync cooldown)
- Show a loading state during creation and sync, then display the new event on the calendar

### 4.3 Event Editing

- Clicking an owned event opens a detail popover with an "Edit" button (existing popover, extended)
- Edit opens a dialog pre-filled with the event's current data
- Editable fields: title, start/end date/time, all-day toggle, location, description
- For recurring event instances, prompt: "Edit this event" vs "Edit all events in the series"
- Editing a recurring series allows changing event content (title, time, location, description) but not the recurrence rule itself
- Events are updated via Google Calendar API — local database is updated via post-mutation sync
- Only the event owner can edit (enforced server-side via connection ownership)
- Non-owners do not see the Edit button

### 4.4 Event Deletion

- Owned event popover includes a "Delete" button with confirmation dialog
- For recurring event instances, prompt: "Delete this event" vs "Delete all events in the series"
- Events are deleted via Google Calendar API — local database is updated via post-mutation sync
- Only the event owner can delete (enforced server-side via connection ownership)
- Non-owners do not see the Delete button

### 4.5 Recurring Event Fix

- Fix the existing sync upsert bug: the unique constraint on `(source_id, external_id)` must account for recurring event instances
- Google's `singleEvents=true` expansion returns instances with composite IDs (e.g., `base_event_id_20260327T100000Z`) — these are already unique per instance
- Verify that the external_id stored from Google already includes the instance suffix; if so, the upsert is correct and the bug may not exist. If not, adjust the sync to use the full instance ID

### 4.6 Post-Mutation Sync

- After any write operation (create, update, delete), trigger a targeted sync for the affected connection
- This sync is internal and bypasses the 5-minute manual sync cooldown
- The sync uses existing incremental sync token logic — only changed events are fetched
- The API response to the client is returned after both the Google API call and the sync complete, ensuring the UI reflects the current state

## 5. API Surface

### New Endpoints

#### `POST /api/v1/calendar/connections/{connectionId}/calendars/{calendarId}/events`

Create an event on a specific Google Calendar.

**Request:**
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

**Response:** `201 Created` with the synced event from the local database (after post-mutation sync).

**Errors:**
- `403 Forbidden` — Connection lacks write scope, or user doesn't own the connection
- `403 Forbidden` — Google API rejects write (calendar is read-only for the user)
- `404 Not Found` — Connection or calendar not found
- `422 Unprocessable Entity` — Validation errors (missing title, invalid dates, end before start)

#### `PATCH /api/v1/calendar/connections/{connectionId}/events/{eventId}`

Update an existing event.

**Request:**
```json
{
  "data": {
    "type": "calendar-events",
    "attributes": {
      "title": "Updated Title",
      "start": "2026-03-28T10:00:00-05:00",
      "end": "2026-03-28T10:30:00-05:00",
      "location": "Room B",
      "description": "Moved to 10am",
      "scope": "single"
    }
  }
}
```

The `scope` field applies to recurring events: `"single"` (this instance only) or `"all"` (entire series).

**Response:** `200 OK` with the updated event after post-mutation sync.

**Errors:**
- `403 Forbidden` — Not the event owner or connection lacks write scope
- `404 Not Found` — Event or connection not found
- `422 Unprocessable Entity` — Validation errors

#### `DELETE /api/v1/calendar/connections/{connectionId}/events/{eventId}`

Delete an event.

**Query Parameters:**
- `scope` — `single` (default) or `all` (for recurring events)

**Response:** `204 No Content`

**Errors:**
- `403 Forbidden` — Not the event owner or connection lacks write scope
- `404 Not Found` — Event or connection not found

### Modified Endpoints

#### `GET /api/v1/calendar/connections`

Response includes a new `writeAccess` boolean attribute per connection, indicating whether the connection has the write scope authorized.

#### `POST /api/v1/calendar/connections/google/authorize`

Accepts an optional `reauthorize` boolean in the request body. When `true`, adds `prompt=consent` and `access_type=offline` to the Google OAuth URL to force re-consent with the upgraded scope.

#### `GET /api/v1/calendar/events`

Response includes a `sourceId` attribute per event, needed for routing edit/delete requests to the correct connection and calendar.

## 6. Data Model

### Modified: `calendar_connections`

Add field:
- `write_access` (boolean, default `false`) — Tracks whether the connection has been authorized with write scope

On re-authorization with the new scope, set to `true`. New connections created after this feature ships will have `write_access = true` by default (since the scope is upgraded).

### Modified: `calendar_events`

Add field:
- `google_calendar_id` (string, nullable) — The Google Calendar ID (source external_id) the event belongs to. Needed to route update/delete requests to the correct Google Calendar. Populated during sync from the source's external_id.

### Verify: Recurring event instance IDs

Confirm that `external_id` in `calendar_events` already stores the full instance-specific ID from Google (e.g., `abc123_20260328T090000Z`). If it stores only the base event ID, the sync logic must be updated to use the full instance ID.

## 7. Service Impact

### calendar-service

- **Google Calendar client** (`internal/googlecal/client.go`): Add `InsertEvent()`, `UpdateEvent()`, `DeleteEvent()` methods
- **OAuth scope** (`internal/googlecal/client.go`): Change `calendar.readonly` to `calendar`
- **Connection model/entity**: Add `write_access` field
- **Connection resource**: Add `reauthorize` support to authorize endpoint, include `writeAccess` in response
- **Event model/entity**: Add `google_calendar_id` field
- **Event resource**: New create, update, delete handlers
- **Sync**: Populate `google_calendar_id` during sync; trigger post-mutation sync internally
- **Migration**: Add `write_access` column to `calendar_connections`, add `google_calendar_id` column to `calendar_events`

### frontend

- **Event creation dialog**: New component with form fields (title, date/time, all-day toggle, recurrence, location, description, calendar picker)
- **Event edit dialog**: Reuses creation form, pre-filled with event data
- **Event delete**: Confirmation dialog with recurring event scope choice
- **Event popover**: Add Edit/Delete buttons for owned events
- **Calendar page**: "Add Event" button, click-to-create on empty time slots
- **Connection status**: Re-authorize prompt for connections without write access
- **API service**: New methods for create, update, delete events
- **React Query hooks**: New mutations with post-success query invalidation

## 8. Non-Functional Requirements

### Performance
- Event creation should complete (including post-mutation sync) within 3 seconds under normal conditions
- Google Calendar API calls should timeout after 30 seconds (existing client behavior)

### Security
- Write operations enforce connection ownership: users can only create/edit/delete events through their own connections
- Event ownership is verified server-side before allowing edit/delete
- Write access flag prevents mutations on read-only connections
- All existing tenant and household scoping remains in effect

### Observability
- Log all write operations (create, update, delete) with connection ID, event ID, and user ID
- Log Google API errors with response status and body for debugging
- Track write operation counts for monitoring

### Multi-tenancy
- All new fields and queries respect existing tenant_id scoping
- Write operations are scoped to the user's connection within their household

## 9. Open Questions

None — all questions resolved.

## 10. Acceptance Criteria

- [ ] Existing connections without write scope show a re-authorize prompt on the calendar page
- [ ] Re-authorization flow upgrades the connection to write access without losing existing sync state
- [ ] New connections are created with write scope by default
- [ ] Users can create a timed event with title, start/end time, location, and description
- [ ] Users can create an all-day event
- [ ] Users can create a recurring event with preset recurrence options
- [ ] Users can select which visible calendar to add the event to
- [ ] Created events appear on the calendar after the post-mutation sync completes
- [ ] Users can edit events they own (title, time, location, description)
- [ ] Users can delete events they own
- [ ] Recurring event edit/delete prompts for "this event" vs "all events"
- [ ] Non-owners cannot see Edit/Delete buttons on events they don't own
- [ ] Write operations return appropriate errors for connections without write scope
- [ ] Google Calendar reflects all changes made through Home Hub (source of truth maintained)
- [ ] No local-only event data exists — all events flow through Google Calendar API and sync
