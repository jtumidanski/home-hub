# Plan Audit — task-013-calendar-event-crud

**Plan Path:** docs/tasks/task-013-calendar-event-crud/tasks.md
**Audit Date:** 2026-03-27
**Branch:** task-013
**Base Branch:** main

## Executive Summary

All 46 implementable tasks (Phases 1–7) are complete with working code changes across 35 files. No commits have been made yet — all changes are unstaged in the working tree. The backend builds and all tests pass (Go and frontend). There are two medium-severity guideline violations in the backend: Google API calls and token management logic reside in the handler layer (`event/resource.go`) rather than being delegated through the processor layer, and the `createEventHandler` returns only `201 Created` without a response body. Frontend code strongly adheres to all guidelines.

## Task Completion

| # | Task | Status | Evidence / Notes |
|---|------|--------|-----------------|
| 1.1 | Add `write_access` column to `calendar_connections` entity | DONE | `connection/entity.go` — `WriteAccess bool` field added |
| 1.2 | Add `write_access` to model, builder, REST response | DONE | `connection/model.go` accessor, `connection/builder.go` setter, `connection/rest.go` field |
| 1.3 | Add `google_calendar_id` column to `calendar_events` entity | DONE | `event/entity.go` — `GoogleCalendarId string` field added |
| 1.4 | Add `google_calendar_id` to event model and builder | DONE | `event/model.go` accessor, `event/builder.go` setter |
| 1.5 | Populate `google_calendar_id` during sync upsert | DONE | `sync/sync.go` — `GoogleCalendarId: src.ExternalID()` in entity construction |
| 1.6 | Change OAuth scope from `calendar.readonly` to `calendar` | DONE | `googlecal/client.go` — scope changed |
| 1.7 | Add `reauthorize` flag to authorize endpoint | DONE | `connection/rest.go` — `Reauthorize` field on `AuthorizeRequest`; `connection/resource.go` passes to oauthstate; `oauthstate/*` — `Reauthorize` field through full stack |
| 1.8 | Callback handler detects re-auth, updates existing connection | DONE | `connection/resource.go:152-175` — re-auth branch updates tokens + `write_access=true` |
| 1.9 | New connections get `write_access=true` | DONE | `connection/administrator.go:25` — `WriteAccess: true` in `create()` |
| 1.10 | Build and verify calendar-service compiles | DONE | `go build ./...` passes |
| 2.1 | Add `InsertEvent()` method | DONE | `googlecal/client.go` — `InsertEvent()` with retry logic |
| 2.2 | Add `UpdateEvent()` method | DONE | `googlecal/client.go` — `UpdateEvent()` with PATCH and retry |
| 2.3 | Add `DeleteEvent()` method | DONE | `googlecal/client.go` — `DeleteEvent()` with retry |
| 2.4 | Add request/response types for write operations | DONE | `googlecal/types.go` — `InsertEventRequest`, `UpdateEventRequest` structs |
| 3.1 | Add public `SyncConnection()` to sync engine | DONE | `sync/sync.go:53-56` — `SyncConnection()` calls `syncOne()` directly |
| 3.2 | Wire sync trigger into event resource via DI | DONE | `cmd/main.go` — `syncConnectionDirect` func injected into `InitializeMutationRoutes` |
| 3.3 | Implement create event handler | DONE | `event/resource.go:99-198` — validates ownership, write access, calls Google API, triggers sync |
| 3.4 | Implement update event handler | DONE | `event/resource.go:201-311` — validates, calls Google API, triggers sync, returns updated event |
| 3.5 | Implement delete event handler | DONE | `event/resource.go:313-385` — validates, calls Google API, triggers sync, returns 204 |
| 3.6 | Add `sourceId`, `connectionId`, `isRecurring` to event list response | DONE | `event/rest.go` — fields added with recurring pattern detection |
| 3.7 | Register new routes in `main.go` | DONE | `cmd/main.go:93` — `event.InitializeMutationRoutes(...)` |
| 3.8 | Build and verify with all new endpoints | DONE | `go build ./...` passes |
| 4.1 | Verify `external_id` stores full instance-specific IDs | DONE | `extractBaseEventID()` in `event/resource.go:418-426` handles instance ID format; sync uses `singleEvents=true` |
| 4.2 | Confirm unique constraint handles recurring instances | DONE | Existing unique constraint on `(source_id, external_id)` with instance-specific IDs works correctly |
| 4.3 | Document findings | SKIPPED | No documentation artifact found for Phase 4 findings |
| 5.1 | Add `writeAccess` to `CalendarConnectionAttributes` | DONE | `types/models/calendar.ts` — `writeAccess: boolean` added |
| 5.2 | Add `sourceId`, `connectionId`, `isRecurring` to event attributes | DONE | `types/models/calendar.ts` — all three fields added |
| 5.3 | Add event mutation request/response types | DONE | `types/models/calendar.ts` — `CreateEventData`, `UpdateEventData` interfaces |
| 5.4 | Add `createEvent()`, `updateEvent()`, `deleteEvent()` service methods | DONE | `services/api/calendar.ts` — all three methods added |
| 5.5 | Add `reauthorizeGoogle()` service method | DONE | `services/api/calendar.ts` — `reauthorizeGoogle()` method |
| 5.6 | Add `useCreateEvent()` mutation hook | DONE | `lib/hooks/api/use-calendar.ts` — with events invalidation |
| 5.7 | Add `useUpdateEvent()` mutation hook | DONE | `lib/hooks/api/use-calendar.ts` — with events invalidation |
| 5.8 | Add `useDeleteEvent()` mutation hook | DONE | `lib/hooks/api/use-calendar.ts` — with events invalidation |
| 5.9 | Add `useReauthorizeCalendar()` mutation hook | DONE | `lib/hooks/api/use-calendar.ts` — with redirect on success |
| 6.1 | Zod validation schema for event form | DONE | `lib/schemas/calendar-event.schema.ts` — title required, end >= start, field length limits |
| 6.2 | Build EventFormDialog with create/edit modes | DONE | `components/features/calendar/event-form-dialog.tsx` — full form with react-hook-form + zod |
| 6.3 | All-day toggle | DONE | event-form-dialog.tsx — switches between date-only and date+time inputs |
| 6.4 | Recurrence selector dropdown | DONE | event-form-dialog.tsx + schema — preset options with RRULE mapping |
| 6.5 | Calendar picker dropdown | DONE | event-form-dialog.tsx — sources with write access; read-only in edit mode |
| 6.6 | Wire form submission to mutations | DONE | event-form-dialog.tsx — calls `useCreateEvent`/`useUpdateEvent` |
| 6.7 | Loading states and error handling | DONE | event-form-dialog.tsx — loading on submit, toast notifications |
| 7.1 | Recurring scope dialog | DONE | `components/features/calendar/recurring-scope-dialog.tsx` — "This event only" / "All events" |
| 7.2 | "Add Event" button | DONE | `pages/CalendarPage.tsx` — button in header, hidden without write access |
| 7.3 | Click-to-create on calendar grid | DONE | `calendar-grid.tsx` — `onSlotClick` with 15-min rounding, write-access gated |
| 7.4 | EventPopover Edit button | DONE | `event-popover.tsx` — owner-only, write-access-only |
| 7.5 | EventPopover Delete button + confirmation | DONE | `event-popover.tsx` + `CalendarPage.tsx` — delete confirmation dialog |
| 7.6 | Wire Edit to EventFormDialog in edit mode | DONE | `CalendarPage.tsx` — edit state management with pre-filled form |
| 7.7 | Wire Delete to mutation with recurring scope | DONE | `CalendarPage.tsx` — recurring scope prompt before delete |
| 7.8 | Re-authorization banner | DONE | `components/features/calendar/reauthorize-banner.tsx` — shows when connections lack write access |
| 7.9 | Error state handling | DONE | Toast notifications throughout, specific error messages for Google write denied |

**Completion Rate:** 45/46 tasks (98%)
**Skipped without approval:** 1 (documentation of Phase 4 findings)
**Partial implementations:** 0

## Skipped / Deferred Tasks

### 4.3 — Document Phase 4 recurring event findings
**Status:** SKIPPED
**Impact:** Low. The implementation correctly handles recurring instance IDs via `extractBaseEventID()`, but no documentation artifact was produced. The code itself serves as documentation of the approach.

## Verification Tasks (V.1–V.11)

These are end-to-end verification tasks requiring a running system with Google Calendar integration. They cannot be verified in this audit. All prerequisite code exists for these flows to work.

## Developer Guidelines Compliance

### Passes

- **Immutable models**: Connection and event models use private fields with accessor methods (`connection/model.go`, `event/model.go`)
- **Entity separation**: GORM tags only on entity structs, models are clean (`connection/entity.go`, `event/entity.go`)
- **Builder pattern**: Both connection and event builders follow fluent pattern with setters (`connection/builder.go`, `event/builder.go`)
- **Administrator layer**: Write operations properly in administrator functions (`connection/administrator.go:59-68`, `event/administrator.go`)
- **Provider pattern**: New providers use lazy evaluation with `database.EntityProvider` (`connection/provider.go`, `event/provider.go`)
- **REST resource separation**: Route registration separated from handler logic (`event/resource.go:32-42`)
- **Multi-tenancy**: `tenantctx.MustFromContext()` used in all handlers (`event/resource.go:47,102,204,316`)
- **RegisterInputHandler**: POST and PATCH endpoints correctly use `server.RegisterInputHandler[T]` (`event/resource.go:34-35`)
- **GetHandler for DELETE**: DELETE endpoint correctly uses `server.GetHandler` since no request body needed (`event/resource.go:36,313`)
- **Processor layer for reads**: Handlers call processors which call providers (`event/resource.go:76,234,346`)
- **d.Logger()**: All handlers use `d.Logger()` not `logrus.StandardLogger()` (`event/resource.go:83,150,191,etc.`)
- **Transform error handling**: Transform errors checked and logged (`event/resource.go:88-93,302-307`)
- **Frontend: Zod schema in dedicated file**: `lib/schemas/calendar-event.schema.ts` — not inline
- **Frontend: react-hook-form + zodResolver**: `event-form-dialog.tsx` uses correct pattern
- **Frontend: Named exports**: All new components use named exports
- **Frontend: Tenant context**: All mutation hooks pass tenant context
- **Frontend: React Query invalidation**: All mutations invalidate `calendarKeys.all()` on settled
- **Frontend: Toast notifications**: Success/error feedback via sonner toast
- **Frontend: cn() utility**: Used for conditional classes in new components
- **Frontend: No `any` types**: TypeScript strict compliance maintained

### Violations

1. **Rule:** Handlers must not call external APIs directly; delegate to processor layer
   - **File:** `event/resource.go:180` (createEventHandler), `event/resource.go:284` (updateEventHandler), `event/resource.go:370` (deleteEventHandler)
   - **Issue:** Handlers call `gcClient.InsertEvent()`, `gcClient.UpdateEvent()`, `gcClient.DeleteEvent()` directly instead of through a processor method. The guideline states handlers should only coordinate request parsing and response marshaling, with business logic delegated to processors.
   - **Severity:** medium
   - **Fix:** Create processor methods like `CreateOnGoogle()`, `UpdateOnGoogle()`, `DeleteOnGoogle()` that encapsulate the Google API interaction, token management, and sync trigger.

2. **Rule:** Token management and crypto operations should be in processor/service layer, not handler helpers
   - **File:** `event/resource.go:387-416` (`getValidAccessToken` function)
   - **Issue:** Token decryption, expiry checking, refresh, re-encryption, and database update are performed in a helper function called from handlers. This logic involves business rules (token expiry) and database writes (update tokens) which belong in the processor layer.
   - **Severity:** medium
   - **Fix:** Move `getValidAccessToken` logic into `connection.Processor` as a method (e.g., `GetOrRefreshAccessToken()`), keeping the handler free of crypto/token concerns.

3. **Rule:** Create endpoint should return the created resource
   - **File:** `event/resource.go:196-197` (createEventHandler)
   - **Issue:** The create handler returns only `201 Created` with no response body. The plan specifies "Return synced event from local DB" (plan.md, Phase 3.2). The update handler correctly returns the event, but create does not.
   - **Severity:** low
   - **Fix:** After sync, query the newly created event and return it as a JSON:API response (matching the update handler pattern).

## Build & Test Results

| Service | Build | Tests | Notes |
|---------|-------|-------|-------|
| calendar-service | PASS | PASS | All 3 test packages pass (connection, crypto, event) |
| frontend | PASS | PASS | `tsc --noEmit` passes, `vite build` succeeds, 43 test suites / 398 tests pass |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE (45/46 tasks done, 1 documentation task skipped)
- **Guidelines Compliance:** MINOR_VIOLATIONS (2 medium-severity layer separation issues in event handlers, 1 low-severity missing response body)
- **Recommendation:** NEEDS_FIXES

## Action Items

1. **Move Google API calls from handlers to processor** — Create `event.Processor` methods that encapsulate `InsertEvent`/`UpdateEvent`/`DeleteEvent` calls, keeping handlers focused on HTTP concerns.
2. **Move `getValidAccessToken` to `connection.Processor`** — Token refresh and crypto are business logic that belongs in the processor layer, not in a handler-accessible helper.
3. **Return created event in create handler response** — After sync, query the event and return it as JSON:API response body with `201 Created`, matching the plan's specification and the update handler's pattern.
4. **(Low priority) Document Phase 4 recurring event findings** — Add a brief note in the task directory about the instance ID format and how `extractBaseEventID` handles it.
5. **Commit all changes** — All work is currently unstaged with no commits on the branch.
