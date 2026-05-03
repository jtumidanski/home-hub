# Task 047 — Context

Quick reference for executing agents. Authoritative spec is `prd.md` + `design.md`.

## Goal in one sentence

Add a series-end control ("On" / "After" / "Never") to the event creation form so recurring events terminate, and add a backend safety net that rejects open-ended `RRULE`s with a stable `recurrence_*` JSON:API error code.

## Key files

### Frontend (no DB, no API shape change)

| Path | What |
|---|---|
| `frontend/src/lib/calendar/recurrence.ts` | **NEW** — `composeRecurrenceRule`, `formatUntilUTC`, `eventStartInstant`, `EndsMode` type |
| `frontend/src/lib/schemas/calendar-event.schema.ts` | Extend `eventFormSchema` and `createEventDefaults` with `endsMode`, `endsOnDate`, `endsAfterCount`, `endsNeverConfirmed`, `endsOnDateUserEdited` + cross-field refinements |
| `frontend/src/components/features/calendar/event-form-dialog.tsx` | Add **Ends** control + 3 effects + replace inline `[values.recurrence]` with `composeRecurrenceRule(...)` call |
| `frontend/src/components/features/calendar/__tests__/recurrence.test.ts` | **NEW** — pure-logic tests for composition + DST |
| `frontend/src/components/features/calendar/__tests__/event-form-dialog.test.tsx` | **NEW** — form-level interaction tests |

### Backend — `services/calendar-service`

| Path | What |
|---|---|
| `services/calendar-service/internal/event/recurrence_validator.go` | **NEW** — `ValidateRecurrence`, `parseRRULE`, `RecurrenceError` |
| `services/calendar-service/internal/event/recurrence_validator_test.go` | **NEW** — table-driven validator tests |
| `services/calendar-service/internal/event/rest.go` | Add `Recurrence *[]string `json:"recurrence"`` to `UpdateEventRequest` |
| `services/calendar-service/internal/event/resource.go` | Wire validator into `createEventHandler` and `updateEventHandler` via shared `validateRecurrenceOrWriteError` helper |

## Locked architectural decisions (from design §2)

1. **Backend RRULE parser:** hand-rolled scanner; only extracts `UNTIL` and `COUNT` case-insensitively; ignores everything else.
2. **`UpdateEventRequest.Recurrence`:** `*[]string` so we can distinguish "not supplied" from "empty array"; processor does *not* consume it. Defense-in-depth only.
3. **FE composition function** lives in `frontend/src/lib/calendar/recurrence.ts`, not in the schema file.
4. **"User edited end date?"** tracked via explicit form field `endsOnDateUserEdited` set true on the input's `onChange` *before* delegating to `field.onChange`.
5. **5-year cap:** `until.Sub(start) > 5*365*24h + 24h` on both sides (1-day cushion absorbs DST/leap drift).

## Limits & error codes

| Bound | Value | BE error code |
|---|---|---|
| `UNTIL` later than start by | `> 5*365*24h + 24h` | `recurrence_too_long` |
| `COUNT` range | `[1, 730]` (inclusive) | `recurrence_count_out_of_range` |
| Missing both `UNTIL` and `COUNT` on an `RRULE:` line | — | `recurrence_unbounded` |

Error helper: `server.WriteJSONAPIError(w, 422, code, "Validation Error", detail, "")` (already exists in `shared/go/server/response.go:41`).

## RRULE composition matrix (FE)

| Mode | Output (length-1 array) |
|---|---|
| `on` | `[`${preset};UNTIL=${formatUntilUTC(endsOnDate, timeZone)}`]` — `UNTIL` is end-of-day in user's timezone, converted to UTC, formatted `YYYYMMDDTHHMMSSZ` |
| `after` | `[`${preset};COUNT=${endsAfterCount}`]` |
| `never` | `[preset]` |
| preset === `""` | `undefined` |

`formatUntilUTC` MUST derive the offset for the *chosen* date via `Intl.DateTimeFormat` parts, not `Date.prototype.getTimezoneOffset()` of "now" (DST safety).

## Test conventions

- **Frontend:** `vitest`, run from `frontend/` via `npm test`. Component tests use `@testing-library/react` + `@testing-library/user-event`; mock `sonner`, `useCreateEvent`, `useUpdateEvent` like `frontend/src/components/features/tasks/__tests__/create-task-dialog.test.tsx`.
- **Backend:** `go test ./services/calendar-service/internal/event/...` from the repo root. The validator tests are pure-function table-driven. There is **no httptest harness** for `createEventHandler` today — the design's "handler tests" are satisfied by validator unit tests; do not invent a new harness for this task.

## Out of scope (do not touch)

- `internal/googlecal/`, `internal/event/processor.go`, `internal/event/builder.go`
- DB schema, migrations
- Edit-form recurrence controls (per task-013)
- Mapping `recurrence_*` codes to specific UI toasts
- Validating non-`RRULE:` recurrence components (`EXDATE`, `RDATE`)
