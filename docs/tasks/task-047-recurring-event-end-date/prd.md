# Recurring Calendar Event End Date — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-05-02
---

## 1. Overview

The calendar event creation form (`event-form-dialog.tsx`) lets users mark an event as recurring via a "Repeats" dropdown (Daily / Weekly / Weekdays / Monthly / Yearly), but provides no way to bound the series. The frontend ships rules like `RRULE:FREQ=WEEKLY` to the backend, which forwards them verbatim to Google Calendar. Per RFC 5545, an `RRULE` without `UNTIL` or `COUNT` repeats indefinitely. The result, observed in production: a "weekly" event intended to run May 6 → June 10 generated occurrences extending years into the future.

This task closes the gap by adding a series-end control to the create form, composing a properly terminated `RRULE` (`UNTIL=…` or `COUNT=…`) before submit, and adding a backend safety net that rejects open-ended recurrence rules so a future buggy client cannot reproduce this incident. No backend contract, schema, or migration changes are required beyond the new validation.

The edit flow remains unchanged; per task-013, modifying the recurrence rule of an existing series is out of scope and stays out of scope here.

## 2. Goals

Primary goals:
- Users can create a recurring event that terminates on a chosen date or after a chosen number of occurrences
- "Never" remains available as a deliberate, explicit choice — never the silent default
- Compose RFC 5545–compliant `RRULE` strings (correct `UNTIL` formatting and timezone handling)
- Defense-in-depth: `calendar-service` rejects open-ended `RRULE`s by default

Non-goals:
- Editing the recurrence rule of an existing series (still out of scope, per task-013)
- Custom RRULE builder beyond the existing presets (no `INTERVAL`, `BYMONTH`, custom `BYDAY` selectors, etc.)
- Surfacing or visualizing recurrence terms on the event detail popover
- Backfill or cleanup of pre-existing runaway series
- Other calendar providers

## 3. User Stories

- As a household member, I want to specify when a recurring event stops repeating, so a six-week volleyball season doesn't fill my calendar through 2030.
- As a household member, I want to specify a number of occurrences (e.g., "10 sessions of physical therapy"), because some series are counted, not dated.
- As a household member, I want "Never" to be a deliberate choice with an explicit warning, so I cannot accidentally create an unbounded series.
- As a household member, I want the end-date control to default to a sensible bounded value (one year out), so the safe path is the easy path.
- As a developer, I want the backend to reject open-ended recurrence rules by default, so a UI bug cannot reproduce the prior incident.

## 4. Functional Requirements

### 4.1 Frontend — "Ends" Control

- The event creation form (`event-form-dialog.tsx`) gains an "Ends" control directly below the "Repeats" dropdown.
- The "Ends" control is rendered only when "Repeats" is set to a non-empty value (i.e., not "Does not repeat").
- The control is a radio group with three options, modeled after Google Calendar's UX:
  - **On** *(date picker)* — series ends on the chosen date, inclusive of that date
  - **After** *(integer input)* `occurrences` — series ends after N occurrences (1 ≤ N ≤ 730)
  - **Never** — series has no terminator
- When the user switches between options, the corresponding sub-control receives focus.
- When the user changes "Repeats" from a recurring value to "Does not repeat", the "Ends" state is reset.

### 4.2 Defaults

- When the user first selects a recurring value from "Repeats", the "Ends" control defaults to **On**, with the date pre-filled as `start date + 1 year`.
- If the user changes the start date after selecting a recurrence, and the "Ends" control is in **On** mode and still holds the auto-default value, the end date updates to `new start date + 1 year`. If the user has manually edited the end date, it is left untouched.

### 4.3 "Never" Confirmation

- Selecting **Never** reveals an inline warning: "This event will repeat forever. Are you sure?" with a checkbox "I understand this event has no end date."
- Submit is blocked while "Never" is selected and the checkbox is unchecked. The submit button shows the current disabled state (no new state needed).

### 4.4 Validation

- **On**: the end date must be strictly after the event's start date.
- **On**: the end date must be no more than 5 years after the event's start date.
- **After**: occurrence count must be an integer between 1 and 730 inclusive.
- All validation lives in the form's Zod schema and surfaces via the existing `<FormMessage />` mechanism.

### 4.5 RRULE Composition

The frontend composes the final `recurrence` array (array of one string) immediately before submit:

| "Ends" mode | Composition |
|-------------|-------------|
| **On**      | Append `;UNTIL=<utc>` to the selected preset `RRULE`. `<utc>` is the chosen date at `23:59:59` in the user's IANA timezone, converted to UTC, formatted as `YYYYMMDDTHHMMSSZ`. |
| **After**   | Append `;COUNT=<N>` to the selected preset `RRULE`. |
| **Never**   | Send the preset `RRULE` unchanged. |

`UNTIL` is inclusive of the selected day: composing it as end-of-day in local time means an occurrence falling anywhere on the picked date is included regardless of the event's start time.

The user's IANA timezone is the same `Intl.DateTimeFormat().resolvedOptions().timeZone` already used elsewhere in the form.

### 4.6 Backend Safety Net

- The `POST .../events` and (where relevant) `PATCH .../events/{id}` handlers in `calendar-service` validate every entry of the `recurrence` array.
- For any `RRULE:` entry, the rule must contain either `UNTIL=` or `COUNT=`. Other recurrence component types (`EXDATE`, `RDATE`, etc.) are not affected by this check.
- A rule that is open-ended is rejected with HTTP `422 Unprocessable Entity` and a JSON:API error whose `code` is `recurrence_unbounded` and whose `detail` explains that an end date or occurrence count is required.
- A rule whose `UNTIL` parses to a value more than 5 years after the event's start is rejected with `422` and `code = recurrence_too_long`.
- A rule whose `COUNT` is outside `[1, 730]` is rejected with `422` and `code = recurrence_count_out_of_range`.
- Validation parses each `RRULE` line case-insensitively for component names (`UNTIL`, `COUNT`) per RFC 5545. The parser does not need to fully validate every RRULE component — only the bounding parts.

### 4.7 Tests

Frontend:
- Unit tests for the RRULE composition function covering all three modes, multiple presets, and a non-UTC timezone (e.g., `America/New_York`) to confirm correct UTC conversion of `UNTIL`.
- Form tests covering: Ends control hidden until a recurrence is selected; defaults populate correctly; switching start date updates auto-default; "Never" requires the checkbox; validation messages appear for out-of-range values.

Backend:
- Handler tests for create (and update if applicable) covering: open-ended rule → 422 `recurrence_unbounded`; UNTIL > 5 years → 422 `recurrence_too_long`; COUNT out of range → 422 `recurrence_count_out_of_range`; valid bounded rules pass through.

## 5. API Surface

No new endpoints. No request/response shape changes.

Modified behavior on existing endpoints:

### `POST /api/v1/calendar/connections/{connId}/calendars/{calId}/events`

Adds three new error conditions:

| Status | Code                            | Condition                                                                  |
|--------|---------------------------------|----------------------------------------------------------------------------|
| 422    | `recurrence_unbounded`          | A submitted `RRULE` lacks both `UNTIL` and `COUNT`                         |
| 422    | `recurrence_too_long`           | A submitted `RRULE`'s `UNTIL` is more than 5 years after the event's start |
| 422    | `recurrence_count_out_of_range` | A submitted `RRULE`'s `COUNT` is outside `[1, 730]`                        |

### `PATCH /api/v1/calendar/connections/{connId}/events/{eventId}`

Same three error conditions when the request includes a `recurrence` array. The PATCH endpoint does not currently expose `recurrence` to the form-driven edit flow (per task-013), so this is purely a defensive guard at the handler boundary; no UI change ships in this task.

## 6. Data Model

No schema changes. No migrations.

The composed `RRULE` string is sent in the existing `recurrence` array attribute and persisted by Google Calendar. The local `calendar_events` table continues to store individual instances expanded by Google's `singleEvents=true`.

## 7. Service Impact

### frontend

- `frontend/src/lib/schemas/calendar-event.schema.ts`
  - Extend `eventFormSchema` with `endsMode: "on" | "after" | "never"`, `endsOnDate: string`, `endsAfterCount: number`, `endsNeverConfirmed: boolean`
  - Add cross-field refinements implementing §4.4
  - Update `RECURRENCE_OPTIONS` and `createEventDefaults` to seed the new fields
  - New exported `composeRecurrenceRule(preset, mode, endsOnDate, endsAfterCount, startDate, startTime, timeZone)` function used by the form's submit handler
- `frontend/src/components/features/calendar/event-form-dialog.tsx`
  - Render the new "Ends" control (radio + sub-controls + warning checkbox) when `recurrence !== ""`
  - Call `composeRecurrenceRule` to build the array sent in `recurrence`
  - Auto-update `endsOnDate` when the user changes `startDate` and has not manually edited the end date
- `frontend/src/components/features/calendar/__tests__/` — new test file(s) per §4.7

### calendar-service

- New validator (e.g., `internal/event/recurrence_validator.go`) implementing §4.6
- Wire validator into the create handler in `internal/event/rest.go`
- Wire validator into the update handler when `recurrence` is supplied (defensive; not user-reachable today)
- Surface validation errors as JSON:API `422` errors using existing error helpers
- Tests per §4.7

No changes to:
- `internal/googlecal/` — Google's API accepts whatever bounded `RRULE` we forward
- Database schema, migrations, or processors

### Other services

None.

## 8. Non-Functional Requirements

### Performance

- RRULE composition and validation are O(small) string operations with negligible impact on request latency.

### Security & Tenancy

- No new attack surface. All existing tenant, household, and connection-ownership scoping remains in effect.
- The 5-year cap and 730-occurrence cap also serve as cheap rate-limits against malicious or buggy clients trying to create runaway series.

### Observability

- Backend rejections (`recurrence_unbounded`, `recurrence_too_long`, `recurrence_count_out_of_range`) are logged at INFO with the connection ID and the offending rule, so we can detect any future client regression.

### Compatibility

- Pre-existing recurring events stored in Google Calendar (including any open-ended ones the user has not cleaned up) are unaffected. The new validator runs only on create/update requests originating from Home Hub; it does not retroactively reject anything coming back from Google during sync.

## 9. Open Questions

None — all questions resolved during spec interview.

## 10. Acceptance Criteria

- [ ] When the create form's "Repeats" dropdown is set to anything other than "Does not repeat", an "Ends" radio control appears with three options: On, After, Never
- [ ] The "Ends" control defaults to **On** with the date set to start date + 1 year
- [ ] Changing the start date updates the auto-defaulted end date but leaves a user-edited end date alone
- [ ] Selecting **Never** shows an explicit warning and a confirmation checkbox; submit is blocked until the checkbox is checked
- [ ] **On** dates earlier than start date show a validation error
- [ ] **On** dates more than 5 years after start date show a validation error
- [ ] **After** counts outside `[1, 730]` show a validation error
- [ ] A "Weekly" event from May 6 to June 10 created via the form generates a finite series ending on June 10 in Google Calendar
- [ ] A "Daily" event with **After 5 occurrences** generates exactly 5 instances in Google Calendar
- [ ] `UNTIL` values are correctly converted to UTC for non-UTC user timezones
- [ ] `POST .../events` returns `422` with code `recurrence_unbounded` when an `RRULE` lacks both `UNTIL` and `COUNT`
- [ ] `POST .../events` returns `422` with code `recurrence_too_long` when `UNTIL` is more than 5 years out
- [ ] `POST .../events` returns `422` with code `recurrence_count_out_of_range` when `COUNT` is outside `[1, 730]`
- [ ] Frontend unit tests cover all branches of the RRULE composition function
- [ ] Backend handler tests cover all three new 422 error paths and the happy path
- [ ] No database migration is added
- [ ] The edit dialog continues to omit any recurrence-rule controls (unchanged from task-013)
