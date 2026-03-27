# Calendar Event CRUD — Implementation Plan

Last Updated: 2026-03-27

---

## Executive Summary

Add full event CRUD (Create, Read, Update, Delete) capabilities to the calendar page. Users will create, edit, and delete events on their synced Google Calendars directly from Home Hub. All mutations write through to the Google Calendar API, with post-mutation sync to reconcile local state. This also requires upgrading the OAuth scope from `calendar.readonly` to `calendar` (read-write) with a re-authorization flow for existing connections.

---

## Current State Analysis

### Backend (calendar-service)

- **Google Calendar client** (`internal/googlecal/client.go`): Read-only — `ListCalendars()`, `ListEvents()`, `RefreshToken()`, `RevokeToken()`, `FetchUserEmail()`. Scope is `calendar.readonly`. Has retry logic with exponential backoff.
- **Connection model** (`internal/connection/`): Tracks OAuth tokens (encrypted AES-256-GCM), status, user assignment, color. No `write_access` field.
- **Event model** (`internal/event/`): Stores synced events with `external_id`, `source_id`, privacy masking. No `google_calendar_id` field.
- **Source model** (`internal/source/`): Stores calendar sources with `external_id` (Google Calendar ID), sync tokens, visibility.
- **Sync engine** (`internal/sync/sync.go`): Background sync every 15 min with jitter. Incremental sync via Google sync tokens. Handles cancelled events. No public method for targeted single-connection sync from handlers.
- **Ownership checks**: Enforced in handlers via `conn.UserID() != t.UserId()` pattern.
- **Event endpoints**: Only `GET /api/v1/calendar/events` exists (list by time range).

### Frontend

- **CalendarPage**: Read-only view with date navigation, connection management, source visibility toggles.
- **EventPopover**: Shows event details (title, time, location, description, owner). No Edit/Delete buttons.
- **CalendarGrid**: 7-column (desktop) / 3-column (mobile) layout with overlapping event positioning. No click-to-create.
- **React Query hooks**: `useCalendarConnections`, `useCalendarSources`, `useCalendarEvents`, `useTriggerSync`. No mutation hooks for events.
- **Types**: `CalendarEventAttributes` lacks `sourceId`, `connectionId`, `isRecurring`.
- **Form patterns**: react-hook-form + zod + shadcn Dialog used extensively in tasks/reminders — same pattern applies here.

---

## Proposed Future State

### Backend

- Google Calendar client gains `InsertEvent()`, `UpdateEvent()`, `DeleteEvent()` methods
- OAuth scope upgraded to `calendar` (read-write)
- Connection entity gains `write_access` boolean column
- Event entity gains `google_calendar_id` column (populated during sync)
- Three new endpoints: POST create, PATCH update, DELETE delete
- Authorize endpoint supports `reauthorize` flag for scope upgrade
- Sync engine exposes method for post-mutation targeted sync (bypasses cooldown)
- Event list response includes `sourceId`, `connectionId`, `isRecurring`

### Frontend

- EventFormDialog: Shared create/edit form with title, date/time, all-day toggle, recurrence, location, description, calendar picker
- Event creation via "Add Event" button and click-on-grid
- Event editing via popover Edit button (owner-only, write-access-only)
- Event deletion via popover Delete button with confirmation
- Recurring event scope prompts ("this event" vs "all events")
- Re-authorization banner for connections without write access
- New React Query mutations with post-success invalidation

---

## Implementation Phases

### Phase 1: Backend Data Model & OAuth Scope Upgrade

Foundational changes that all subsequent work depends on.

**1.1 Add `write_access` column to `calendar_connections`**
- Add `WriteAccess bool` field to `connection/entity.go`
- Add `writeAccess` accessor to `connection/model.go`
- Add `writeAccess` to `connection/builder.go`
- GORM AutoMigrate adds column (default `false`)
- Effort: **S**

**1.2 Add `google_calendar_id` column to `calendar_events`**
- Add `GoogleCalendarId string` field to `event/entity.go`
- Add accessor to `event/model.go`
- GORM AutoMigrate adds column
- Effort: **S**

**1.3 Populate `google_calendar_id` during sync**
- In `sync/sync.go` `syncSource()`, set `GoogleCalendarId` to `src.ExternalID()` on each event entity during upsert
- Effort: **S**

**1.4 Upgrade OAuth scope to read-write**
- In `googlecal/client.go`, change `calendarScope` from `calendar.readonly` to `calendar`
- New connections automatically get write scope
- Effort: **S**

**1.5 Support re-authorization in authorize endpoint**
- In `connection/resource.go` authorize handler, accept `reauthorize` boolean in request body
- When `true`, add `prompt=consent` to Google OAuth URL
- In callback handler, detect re-authorization (existing connection for same user/provider) and update tokens + set `write_access = true` instead of creating new connection
- Effort: **M**

**1.6 Include `writeAccess` in connection list response**
- Update `connection/rest.go` to include `writeAccess` in JSON:API response
- Effort: **S**

### Phase 2: Google Calendar Write Client Methods

**2.1 Add `InsertEvent()` method**
- Endpoint: `POST https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events`
- Accept event struct (title, start, end, location, description, recurrence)
- Handle all-day vs timed events (date vs dateTime in Google API)
- Use existing `doWithRetry` for resilience
- Return created event ID
- Effort: **M**

**2.2 Add `UpdateEvent()` method**
- Endpoint: `PATCH https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events/{eventId}`
- Support `scope=single` (update instance) vs `scope=all` (update base event)
- For `scope=all` on recurring instances: extract base event ID from composite ID
- Effort: **M**

**2.3 Add `DeleteEvent()` method**
- Endpoint: `DELETE https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events/{eventId}`
- Support `scope=single` vs `scope=all` (same base event ID logic)
- Effort: **S**

### Phase 3: Post-Mutation Sync & Backend Endpoints

**3.1 Expose targeted sync method**
- Add `SyncConnection(ctx, conn)` public method on sync Engine
- Reuses existing `syncOne()` logic
- Bypasses the 5-minute manual cooldown (internal call, not rate-limited)
- Effort: **S**

**3.2 Create event endpoint**
- `POST /api/v1/calendar/connections/{connectionId}/calendars/{calendarId}/events`
- Validate: connection ownership, write access, calendar belongs to connection
- Parse request body, validate fields (title required, end >= start)
- Look up source to get Google Calendar ID (`source.ExternalID()`)
- Call `InsertEvent()` on Google
- Trigger post-mutation sync
- Return synced event from local DB
- Effort: **L**

**3.3 Update event endpoint**
- `PATCH /api/v1/calendar/connections/{connectionId}/events/{eventId}`
- Validate: connection ownership, write access, event belongs to connection
- Look up event to get `google_calendar_id` and `external_id`
- Call `UpdateEvent()` on Google
- Trigger post-mutation sync
- Return updated event
- Effort: **L**

**3.4 Delete event endpoint**
- `DELETE /api/v1/calendar/connections/{connectionId}/events/{eventId}?scope=single`
- Validate: connection ownership, write access, event belongs to connection
- Look up event to get `google_calendar_id` and `external_id`
- Call `DeleteEvent()` on Google
- Trigger post-mutation sync
- Return 204
- Effort: **M**

**3.5 Include `sourceId`, `connectionId`, `isRecurring` in event list response**
- Update `event/rest.go` to include these fields
- `isRecurring`: detect from `external_id` format (contains `_` suffix with timestamp pattern)
- Effort: **S**

### Phase 4: Verify Recurring Event Sync

**4.1 Verify recurring event instance IDs**
- Confirm that `external_id` in `calendar_events` stores full instance-specific IDs (e.g., `abc123_20260328T090000Z`)
- Google's `singleEvents=true` expansion already returns unique per-instance IDs
- Verify the unique constraint on `(source_id, external_id)` handles these correctly
- If IDs are base-only, update sync to use full instance ID
- Effort: **S**

### Phase 5: Frontend — Types, API, Hooks

**5.1 Update TypeScript types**
- Add `writeAccess` to `CalendarConnectionAttributes`
- Add `sourceId`, `connectionId`, `isRecurring` to `CalendarEventAttributes`
- Add event mutation request/response types
- Effort: **S**

**5.2 Add API service methods**
- `createEvent(tenant, connectionId, calendarId, data)` — POST
- `updateEvent(tenant, connectionId, eventId, data)` — PATCH
- `deleteEvent(tenant, connectionId, eventId, scope)` — DELETE
- `reauthorizeGoogle(tenant, redirectUri)` — POST with `reauthorize: true`
- Effort: **S**

**5.3 Add React Query mutation hooks**
- `useCreateEvent()` — invalidates events queries on success
- `useUpdateEvent()` — invalidates events queries on success
- `useDeleteEvent()` — invalidates events queries on success
- `useReauthorizeCalendar()` — redirects to Google OAuth
- Effort: **M**

### Phase 6: Frontend — Event Form Dialog

**6.1 Build EventFormDialog component**
- Shared dialog for create and edit modes
- Zod schema for validation (title required, end >= start)
- Fields: title, all-day toggle, start date/time, end date/time, recurrence selector, location, description, calendar picker
- All-day toggle swaps between date-only and date+time inputs
- Calendar picker: dropdown of user's visible sources with write access
- In edit mode: calendar picker is read-only, shows source name
- Loading state on submit button
- Effort: **XL**

**6.2 Recurrence selector**
- Dropdown with preset options: None, Daily, Weekly, Weekdays (Mon-Fri), Monthly, Yearly
- Maps selection to RRULE string array
- In edit mode for recurring events: shows current rule but not editable
- Effort: **M**

**6.3 Recurring event scope prompt**
- Alert dialog shown before edit/delete of recurring events
- Options: "This event only" / "All events in series"
- Returns scope value ("single" or "all") to parent
- Effort: **S**

### Phase 7: Frontend — Calendar Page Integration

**7.1 "Add Event" button**
- Add button to calendar page header (near navigation)
- Opens EventFormDialog in create mode
- Hidden/disabled if no connections have write access
- Effort: **S**

**7.2 Click-to-create on calendar grid**
- Click on empty time slot opens EventFormDialog
- Pre-fills start time (rounded to nearest 15 min) and end time (start + 1 hour)
- Only functional if user has write access
- Effort: **M**

**7.3 Extend EventPopover with Edit/Delete buttons**
- Show Edit button for owned events on connections with write access
- Show Delete button for owned events on connections with write access
- Edit opens EventFormDialog in edit mode, pre-filled
- Delete opens confirmation dialog
- For recurring events, show scope prompt before action
- Effort: **M**

**7.4 Re-authorization banner**
- Show banner on calendar page when user has connections without write access
- "Upgrade your calendar connection to add and edit events" with [Upgrade Access] button
- Button triggers re-authorization flow
- Banner disappears when all connections have write access
- Effort: **S**

**7.5 Error states and loading**
- Loading spinners during mutations
- Toast notifications for success/failure
- Specific error messages per scenario (write denied, network error, sync failure)
- Effort: **S**

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Re-authorization breaks existing connections | Medium | High | Update existing connection tokens on callback instead of creating new; preserve sync tokens and state |
| Google API rate limits during post-mutation sync | Low | Medium | Existing retry logic with backoff handles this; sync is incremental (sync token) |
| Recurring event ID format assumptions incorrect | Medium | Medium | Phase 4 explicitly verifies this before building on it; fallback to full sync if needed |
| OAuth scope change requires all users to re-auth | Certain | Low | By design — read-only still works; banner prompts upgrade without forcing |
| Post-mutation sync adds latency to write responses | Certain | Low | Incremental sync is fast (only changed events); acceptable for 3s target |
| Calendar picker shows calendars without write permission | Low | Medium | Google API returns error for read-only calendars; surface as specific toast error |

---

## Success Metrics

- Users can create, edit, and delete events without leaving Home Hub
- All mutations reflect on Google Calendar within seconds
- Post-mutation sync completes within 3 seconds
- Existing read-only users experience no degradation
- Re-authorization flow preserves existing sync state and data

---

## Required Resources and Dependencies

### External Dependencies
- Google Calendar API v3 (write endpoints)
- Google OAuth2 consent flow with scope upgrade

### Internal Dependencies
- Existing sync engine (extended, not replaced)
- Existing OAuth flow (extended with re-authorize support)
- shadcn/ui form components (Dialog, Form, Input, Select, Switch, Button)
- react-hook-form + zod (existing pattern)
- TanStack React Query (existing pattern)

---

## Timeline Estimates

| Phase | Description | Effort | Dependencies |
|-------|-------------|--------|-------------|
| 1 | Data Model & OAuth Scope | M | None |
| 2 | Google Calendar Write Client | M | Phase 1 |
| 3 | Backend Endpoints & Post-Mutation Sync | L | Phase 1, 2 |
| 4 | Verify Recurring Event Sync | S | Phase 1 |
| 5 | Frontend Types, API, Hooks | S | Phase 3 |
| 6 | Frontend Event Form Dialog | XL | Phase 5 |
| 7 | Frontend Calendar Page Integration | L | Phase 5, 6 |
