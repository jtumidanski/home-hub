# Calendar Event CRUD — Context

Last Updated: 2026-03-27

---

## Key Backend Files

### Google Calendar Client
- `services/calendar-service/internal/googlecal/client.go` — OAuth URLs, token exchange/refresh/revoke, ListCalendars, ListEvents, doWithRetry. **Add**: InsertEvent, UpdateEvent, DeleteEvent methods. **Modify**: Change `calendarScope` from `calendar.readonly` to `calendar`.

### Connection Domain
- `services/calendar-service/internal/connection/entity.go` — GORM entity, `calendar_connections` table. **Add**: `WriteAccess bool` column.
- `services/calendar-service/internal/connection/model.go` — Immutable domain model. **Add**: `writeAccess` field + accessor.
- `services/calendar-service/internal/connection/builder.go` — Builder pattern. **Add**: `WithWriteAccess()`.
- `services/calendar-service/internal/connection/processor.go` — Business logic, ownership check, rate limiting. **Add**: `UpdateWriteAccess()` method.
- `services/calendar-service/internal/connection/resource.go` — HTTP handlers. **Modify**: authorize handler to accept `reauthorize` flag; callback handler to update existing connection on re-auth.
- `services/calendar-service/internal/connection/rest.go` — JSON:API mapping. **Add**: `writeAccess` to response.
- `services/calendar-service/internal/connection/provider.go` — DB queries. **Add**: query to find existing connection by user+provider for re-auth detection.

### Event Domain
- `services/calendar-service/internal/event/entity.go` — GORM entity, `calendar_events` table. Unique constraint on `(source_id, external_id)`. **Add**: `GoogleCalendarId string` column.
- `services/calendar-service/internal/event/model.go` — Domain model with `IsPrivate()`. **Add**: `googleCalendarId` field + accessor.
- `services/calendar-service/internal/event/builder.go` — **Add**: `WithGoogleCalendarId()`.
- `services/calendar-service/internal/event/processor.go` — Business logic. **Add**: validation for create/update.
- `services/calendar-service/internal/event/resource.go` — HTTP handlers (only GET list). **Add**: POST create, PATCH update, DELETE delete handlers.
- `services/calendar-service/internal/event/rest.go` — JSON:API mapping with privacy transform. **Add**: `sourceId`, `connectionId`, `isRecurring` to response.
- `services/calendar-service/internal/event/provider.go` — DB queries. **Add**: query by ID with connection ownership validation.

### Source Domain
- `services/calendar-service/internal/source/entity.go` — `calendar_sources` table with `ExternalId` (Google Calendar ID) and `SyncToken`.
- `services/calendar-service/internal/source/model.go` — Model with accessors. `ExternalID()` returns the Google Calendar ID needed for API calls.
- `services/calendar-service/internal/source/provider.go` — DB queries. Need to look up source by ID to get `ExternalID()` for create endpoint.

### Sync Engine
- `services/calendar-service/internal/sync/sync.go` — Background sync loop, syncAll, syncOne, getValidAccessToken, refreshCalendarList, syncSource. **Add**: public `SyncConnection()` method for post-mutation sync.

### Service Entry Point
- `services/calendar-service/cmd/main.go` — Wires dependencies, registers routes, starts sync engine. **Modify**: register new event mutation routes, wire sync trigger for post-mutation use.

### Crypto
- `services/calendar-service/internal/crypto/crypto.go` — AES-256-GCM encryption for OAuth tokens. Used to decrypt tokens before Google API calls.

### Config
- `services/calendar-service/internal/config/config.go` — Environment variables. No changes needed.

---

## Key Frontend Files

### Types
- `frontend/src/types/models/calendar.ts` — `CalendarConnectionAttributes`, `CalendarSourceAttributes`, `CalendarEventAttributes`. **Add**: `writeAccess` to connection, `sourceId`/`connectionId`/`isRecurring` to event.

### API Service
- `frontend/src/services/api/calendar.ts` — API methods for connections, sources, events. **Add**: `createEvent()`, `updateEvent()`, `deleteEvent()`, `reauthorizeGoogle()`.

### React Query Hooks
- `frontend/src/lib/hooks/api/use-calendar.ts` — Query keys, queries, mutations. **Add**: `useCreateEvent()`, `useUpdateEvent()`, `useDeleteEvent()`, `useReauthorizeCalendar()`.

### Calendar Page
- `frontend/src/pages/CalendarPage.tsx` — Main page with date navigation, connection status, grid. **Add**: "Add Event" button, re-authorization banner.

### Calendar Grid
- `frontend/src/components/features/calendar/calendar-grid.tsx` — Week view with hour grid, event positioning. **Add**: click-on-empty-slot handler for pre-filled event creation.

### Event Popover
- `frontend/src/components/features/calendar/event-popover.tsx` — Modal showing event details. **Add**: Edit/Delete buttons (owner-only, write-access-only).

### Event Block
- `frontend/src/components/features/calendar/event-block.tsx` — Timed event display on grid. No changes needed.

### Connection Status
- `frontend/src/components/features/calendar/connection-status.tsx` — Per-connection status display. May show write access indicator.

### New Components (to create)
- `frontend/src/components/features/calendar/event-form-dialog.tsx` — Shared create/edit form dialog
- `frontend/src/components/features/calendar/recurring-scope-dialog.tsx` — "This event" vs "All events" prompt
- `frontend/src/components/features/calendar/reauthorize-banner.tsx` — Write access upgrade prompt

### Form Patterns (reference)
- `frontend/src/components/features/tasks/create-task-dialog.tsx` — Example of react-hook-form + zod + shadcn dialog pattern
- `frontend/src/components/features/tasks/edit-task-dialog.tsx` — Example of edit dialog with pre-filled form

---

## Key Decisions

1. **Google Calendar remains source of truth** — No local-only event data. All mutations go through Google API, then sync back.

2. **Post-mutation sync is synchronous to the request** — API response waits for both Google API call and sync to complete before returning. This ensures the UI reflects current state.

3. **Re-authorization updates existing connection** — Does not create a new connection. Preserves sync tokens, sources, and event data.

4. **Recurrence presets only** — No custom RRULE builder. Preset options: None, Daily, Weekly, Weekdays, Monthly, Yearly.

5. **Recurrence rule not editable** — When editing a recurring series, content fields (title, time, location, description) can change but recurrence pattern cannot.

6. **Event ownership via connection ownership** — An event is "owned" by the user whose connection synced it. Verified server-side via `event.ConnectionId → connection.UserID`.

7. **Write access tracked per connection** — Not per user or per household. Each connection independently tracks whether it has write scope.

8. **Calendar picker uses source UUID** — Create endpoint uses `calendarId` (source UUID) in path; handler looks up `source.ExternalID()` for Google API call.

---

## Dependencies Between Tasks

```
Phase 1 (Data Model + OAuth)
  ├── Phase 2 (Google Write Client) ← needs scope upgrade
  │     └── Phase 3 (Backend Endpoints) ← needs write methods
  │           └── Phase 5 (Frontend Types/API/Hooks) ← needs endpoints
  │                 ├── Phase 6 (Event Form Dialog) ← needs hooks
  │                 └── Phase 7 (Page Integration) ← needs hooks + dialog
  └── Phase 4 (Recurring Event Verification) ← independent investigation
```

---

## Google Calendar API Reference

### Insert Event
```
POST https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events
Authorization: Bearer {accessToken}

{
  "summary": "Event Title",
  "location": "Room A",
  "description": "Details",
  "start": { "dateTime": "2026-03-28T09:00:00-05:00" },  // timed
  "end": { "dateTime": "2026-03-28T10:00:00-05:00" },
  "recurrence": ["RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR"]
}

// All-day event uses "date" instead of "dateTime":
{
  "start": { "date": "2026-03-28" },
  "end": { "date": "2026-03-29" }   // exclusive end date
}
```

### Update Event (Patch)
```
PATCH https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events/{eventId}
// Same body fields, only provided fields are updated

// For recurring: use base event ID for "all", instance ID for "single"
```

### Delete Event
```
DELETE https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events/{eventId}
// For recurring: same ID logic as update
```

### Recurring Event Instance IDs
- Base event: `abc123`
- Instance: `abc123_20260328T090000Z`
- To edit/delete single instance: use full instance ID
- To edit/delete all: extract base ID (before `_` + timestamp suffix)
