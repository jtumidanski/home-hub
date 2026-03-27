# Calendar Event CRUD — Task Checklist

Last Updated: 2026-03-27

---

## Phase 1: Backend Data Model & OAuth Scope Upgrade

- [ ] **1.1** Add `write_access` boolean column to `calendar_connections` entity (default `false`)
- [ ] **1.2** Add `write_access` to connection model, builder, and REST response
- [ ] **1.3** Add `google_calendar_id` string column to `calendar_events` entity
- [ ] **1.4** Add `google_calendar_id` to event model and builder
- [ ] **1.5** Populate `google_calendar_id` from `source.ExternalID()` during sync upsert
- [ ] **1.6** Change OAuth scope from `calendar.readonly` to `calendar` in Google client
- [ ] **1.7** Add `reauthorize` flag support to authorize endpoint
- [ ] **1.8** Update callback handler to detect re-authorization and update existing connection (tokens + `write_access = true`)
- [ ] **1.9** Set `write_access = true` for new connections created after scope upgrade
- [ ] **1.10** Build and verify calendar-service compiles and starts

## Phase 2: Google Calendar Write Client Methods

- [ ] **2.1** Add `InsertEvent()` method — creates event via Google Calendar API, handles timed/all-day/recurrence
- [ ] **2.2** Add `UpdateEvent()` method — patches event, supports single instance vs all (base event ID extraction)
- [ ] **2.3** Add `DeleteEvent()` method — deletes event, supports single instance vs all
- [ ] **2.4** Add request/response types for Google Calendar write operations

## Phase 3: Post-Mutation Sync & Backend Endpoints

- [ ] **3.1** Add public `SyncConnection()` method to sync engine (bypasses manual sync cooldown)
- [ ] **3.2** Wire sync trigger function into event resource handlers via dependency injection
- [ ] **3.3** Implement create event handler — validate ownership, write access, call Google API, trigger sync, return event
- [ ] **3.4** Implement update event handler — validate ownership, write access, look up external IDs, call Google API, trigger sync, return event
- [ ] **3.5** Implement delete event handler — validate ownership, write access, look up external IDs, call Google API, trigger sync, return 204
- [ ] **3.6** Add `sourceId`, `connectionId`, `isRecurring` to event list response
- [ ] **3.7** Register new routes in `main.go`
- [ ] **3.8** Build and verify calendar-service with all new endpoints

## Phase 4: Verify Recurring Event Sync

- [ ] **4.1** Verify `external_id` stores full instance-specific IDs from Google (`singleEvents=true` expansion)
- [ ] **4.2** Confirm unique constraint on `(source_id, external_id)` handles recurring instances correctly
- [ ] **4.3** Document findings and any adjustments made

## Phase 5: Frontend — Types, API Service, Hooks

- [ ] **5.1** Add `writeAccess` to `CalendarConnectionAttributes` type
- [ ] **5.2** Add `sourceId`, `connectionId`, `isRecurring` to `CalendarEventAttributes` type
- [ ] **5.3** Add event mutation request/response types
- [ ] **5.4** Add `createEvent()`, `updateEvent()`, `deleteEvent()` API service methods
- [ ] **5.5** Add `reauthorizeGoogle()` API service method
- [ ] **5.6** Add `useCreateEvent()` mutation hook with events query invalidation
- [ ] **5.7** Add `useUpdateEvent()` mutation hook with events query invalidation
- [ ] **5.8** Add `useDeleteEvent()` mutation hook with events query invalidation
- [ ] **5.9** Add `useReauthorizeCalendar()` mutation hook

## Phase 6: Frontend — Event Form Dialog

- [ ] **6.1** Create Zod validation schema for event form (title required, end >= start, field length limits)
- [ ] **6.2** Build EventFormDialog component with create/edit modes
- [ ] **6.3** Implement all-day toggle (switches between date-only and date+time inputs)
- [ ] **6.4** Implement recurrence selector dropdown (None, Daily, Weekly, Weekdays, Monthly, Yearly)
- [ ] **6.5** Implement calendar picker dropdown (user's visible sources with write access; read-only in edit mode)
- [ ] **6.6** Wire form submission to create/update mutation hooks
- [ ] **6.7** Add loading states and error handling (toasts)

## Phase 7: Frontend — Calendar Page Integration

- [ ] **7.1** Build recurring scope dialog ("This event only" / "All events in series")
- [ ] **7.2** Add "Add Event" button to calendar page header (hidden if no write access)
- [ ] **7.3** Add click-to-create on calendar grid empty slots (pre-fill start/end time)
- [ ] **7.4** Extend EventPopover with Edit button (owner-only, write-access-only)
- [ ] **7.5** Extend EventPopover with Delete button + confirmation dialog (owner-only, write-access-only)
- [ ] **7.6** Wire Edit button to EventFormDialog in edit mode (pre-filled, with recurring scope prompt)
- [ ] **7.7** Wire Delete button to delete mutation (with recurring scope prompt)
- [ ] **7.8** Build re-authorization banner for connections without write access
- [ ] **7.9** Add error state handling (no connections, no write access, Google write denied, network errors)

## Final Verification

- [ ] **V.1** End-to-end: Create timed event → appears on Google Calendar → appears in Home Hub after sync
- [ ] **V.2** End-to-end: Create all-day event → appears correctly
- [ ] **V.3** End-to-end: Create recurring event → instances appear correctly
- [ ] **V.4** End-to-end: Edit event (single instance of recurring) → change reflected
- [ ] **V.5** End-to-end: Edit all events in recurring series → changes reflected
- [ ] **V.6** End-to-end: Delete event → removed from both Google and Home Hub
- [ ] **V.7** End-to-end: Delete all events in recurring series → all removed
- [ ] **V.8** Re-authorization flow → existing connection upgraded, sync state preserved
- [ ] **V.9** Non-owner cannot see Edit/Delete buttons
- [ ] **V.10** Connection without write access shows banner, hides Add Event button
- [ ] **V.11** Docker build succeeds for calendar-service and frontend
