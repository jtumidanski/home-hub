# Recurring Calendar Event End Date — Design

Version: v1
Status: Draft
Created: 2026-05-02
Companion to: `prd.md`
---

## 1. Scope and Constraints (recap)

This design implements the PRD without changing any HTTP route, request/response shape, database schema, or sync pipeline. The work is two new validators (one frontend, one backend) plus a new "Ends" sub-form. The PRD's acceptance criteria are taken as given; this document covers only how those requirements are realized.

## 2. Open Architectural Decisions

The PRD is prescriptive about UX, error codes, and limits, but leaves several implementation decisions open. They are resolved here:

| # | Decision | Choice | Rationale |
|---|---|---|---|
| 1 | Backend RRULE parsing | Hand-rolled scanner | Matches the PRD's "doesn't need to fully validate every RRULE component"; ~50 LOC; no new dependencies |
| 2 | `UpdateEventRequest.Recurrence` field | Add `Recurrence *[]string` and run the validator when non-nil | Defense-in-depth as PRD §4.6 / §7 require. Field stays unwired in the processor — no Google propagation path is added — so this changes no user-facing behavior |
| 3 | Frontend composition function location | New file `frontend/src/lib/calendar/recurrence.ts`, not `calendar-event.schema.ts` | Schema files hold Zod schemas; composition is independent logic with its own unit tests. PRD §7 was illustrative |
| 4 | "User edited the end date?" detection | Explicit boolean form field `endsOnDateUserEdited`, set true on any user `onChange` against the end-date input | Robust against the user typing a value that happens to equal the auto-default; trivial to test |
| 5 | 5-year cap origin (FE/BE alignment) | Both sides use the full event start instant as the origin and compare with a 1-day cushion: `until.Sub(origin) > 5*365*24h + 24h` | Avoids drift where FE day-precision would let through values BE then rejects with full-time precision. The 730-occurrence cap is the real teeth |

## 3. Component Inventory

### 3.1 Frontend

**New `frontend/src/lib/calendar/recurrence.ts`**
Pure logic, no React or Zod:

```ts
export type EndsMode = "on" | "after" | "never";

export function composeRecurrenceRule(
  preset: string,                  // "" | "RRULE:FREQ=DAILY" | …
  mode: EndsMode,
  endsOnDate: string,              // YYYY-MM-DD; ignored unless mode === "on"
  endsAfterCount: number,          // ignored unless mode === "after"
  startDate: string,
  startTime: string,
  timeZone: string,                // IANA, e.g., "America/New_York"
): string[] | undefined;

export function eventStartInstant(
  startDate: string,
  startTime: string,
  allDay: boolean,
  timeZone: string,
): Date;

export function formatUntilUTC(
  endsOnDate: string,              // YYYY-MM-DD
  timeZone: string,                // IANA
): string;                         // "YYYYMMDDTHHMMSSZ"
```

`formatUntilUTC` takes the chosen day at `23:59:59` *in the user's IANA timezone*, converts to UTC, formats as `YYYYMMDDTHHMMSSZ`. Implementation derives the offset for that specific local date via `Intl.DateTimeFormat` parts so DST transitions on the chosen day produce the right offset (rather than `new Date()`'s "today's offset").

`composeRecurrenceRule`:
- `preset === ""` → `undefined`
- `mode === "on"` → `[`${preset};UNTIL=${formatUntilUTC(endsOnDate, timeZone)}`]`
- `mode === "after"` → `[`${preset};COUNT=${endsAfterCount}`]`
- `mode === "never"` → `[preset]`

**Modified `frontend/src/lib/schemas/calendar-event.schema.ts`**
Adds to `eventFormSchema`:

```ts
endsMode: z.enum(["on", "after", "never"]),
endsOnDate: z.string(),                  // empty when mode != "on"
endsAfterCount: z.coerce.number().int(),
endsNeverConfirmed: z.boolean(),
endsOnDateUserEdited: z.boolean(),
```

Cross-field `.refine()`s (each gated on `recurrence !== ""`):
- `endsMode === "on"`: `endsOnDate` parses; resulting Date > start instant; `until - start <= 5*365*24h + 24h` (mirrors backend per §2.5).
- `endsMode === "after"`: `1 ≤ endsAfterCount ≤ 730`.
- `endsMode === "never"`: `endsNeverConfirmed === true`. The unmet refinement makes the schema invalid, which surfaces an inline message *and* keeps the submit button disabled via the existing `form.formState.isSubmitting`/validity gate. (PRD §4.3 says no new disabled-state plumbing is required.)

`createEventDefaults` seeds `endsMode: "on"`, `endsOnDate: ""`, `endsAfterCount: 10`, `endsNeverConfirmed: false`, `endsOnDateUserEdited: false`. `endsOnDate` is filled lazily by the form when the user first chooses a recurring preset (see §3.3 effects).

**Modified `frontend/src/components/features/calendar/event-form-dialog.tsx`**
- Renders the **Ends** control (radio + sub-controls + warning checkbox) when `recurrence !== ""` and `!isEdit`. The edit dialog already omits the recurrence dropdown, so this control never appears in the edit flow — preserves task-013's invariant.
- Auto-update effects (described in §3.3).
- End-date input's `onChange` sets `endsOnDateUserEdited = true` *before* invoking `field.onChange(...)` so the next render sees the flag.
- Submit handler resolves `timeZone` once (already done) and replaces the literal `[values.recurrence]` line with `composeRecurrenceRule(values.recurrence, values.endsMode, values.endsOnDate, values.endsAfterCount, values.startDate, values.startTime, timeZone)`.

**New `frontend/src/components/features/calendar/__tests__/recurrence.test.ts`**
Table-driven tests for `composeRecurrenceRule` across:
- All three modes × `RRULE:FREQ=DAILY`, `WEEKLY`, `WEEKLY;BYDAY=MO,TU,WE,TH,FR`, `MONTHLY`, `YEARLY`.
- `America/New_York` timezone, both inside and outside DST, asserting `formatUntilUTC` returns `T035959Z` (winter, EST = UTC-5) and `T025959Z` (summer, EDT = UTC-4) for end-of-day.
- A non-recurring preset (`""`) returns `undefined`.

**New or extended `frontend/src/components/features/calendar/__tests__/event-form-dialog.test.tsx`**
- Ends control hidden when "Does not repeat" is selected.
- Selecting a recurring preset shows the Ends control and seeds end date to start + 1y.
- Switching start date updates auto-default; switching start date after the user manually edits the end date leaves it alone.
- Selecting "Never" reveals the warning + checkbox; submit disabled until checkbox is checked.
- Out-of-range "After" and "On" values produce the expected `<FormMessage />`.
- Switching back to "Does not repeat" resets `ends*` fields.

### 3.2 Backend (`calendar-service`)

**New `services/calendar-service/internal/event/recurrence_validator.go`**
Pure functions, no DB:

```go
package event

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"
)

const (
    maxUntilWindow  = 5*365*24*time.Hour + 24*time.Hour // 5y + 1-day cushion
    minOccurrences  = 1
    maxOccurrences  = 730
    codeUnbounded   = "recurrence_unbounded"
    codeTooLong     = "recurrence_too_long"
    codeCountRange  = "recurrence_count_out_of_range"
)

type RecurrenceError struct {
    Code    string
    Detail  string
    RuleRaw string  // for INFO logging at the call site
}

func (e *RecurrenceError) Error() string { return e.Code + ": " + e.Detail }

// ValidateRecurrence runs the §4.6 checks against every RRULE: line in the
// slice. EXDATE, RDATE, and other components are ignored. Returns the first
// error encountered or nil.
func ValidateRecurrence(recurrence []string, eventStart time.Time) *RecurrenceError

// parseRRULE extracts UNTIL and COUNT from a single "RRULE:..." line.
// Returns (untilUTC, count, hasUntil, hasCount, err). Component names are
// matched case-insensitively per RFC 5545.
func parseRRULE(line string) (until time.Time, count int, hasUntil, hasCount bool, err error)
```

`parseRRULE` strips a leading `RRULE:` (case-insensitive), splits on `;`, and for each `KEY=VALUE` pair compares the upper-cased key against `UNTIL` and `COUNT`. `UNTIL` accepts both the date form `20060102` and the date-time UTC form `20060102T150405Z`. `COUNT` is parsed via `strconv.Atoi`. Lines that do not start with `RRULE:` are ignored entirely (they are valid recurrence components but not subject to bounding rules). A truly malformed `UNTIL`/`COUNT` returns an error; the validator surfaces this as `recurrence_unbounded` rather than a fourth error code (the practical outcome — "we don't see a usable bound" — is the same).

**Modified `services/calendar-service/internal/event/rest.go`**
Add to `UpdateEventRequest`:

```go
Recurrence *[]string `json:"recurrence"`
```

`*[]string` to distinguish "field not supplied" from "supplied as empty array." The processor does not consume this field; it exists only so the validator at the boundary has something to validate.

**Modified `services/calendar-service/internal/event/resource.go`**

- After `if input.Title == ""` in `createEventHandler`, parse `input.Start` (RFC 3339) and call `ValidateRecurrence(input.Recurrence, eventStart)`. On a non-nil `*RecurrenceError`:
  - `d.Logger().WithFields(logrus.Fields{"connID": connID, "rule": e.RuleRaw}).Info("rejected recurrence")`
  - `server.WriteJSONAPIError(w, http.StatusUnprocessableEntity, e.Code, "Validation Error", e.Detail, "")`
  - `return`
- In `updateEventHandler`, if `input.Recurrence != nil`, do the same. Origin time is `*input.Start` parsed (if supplied), otherwise `time.Time{}`. `ValidateRecurrence` checks `eventStart.IsZero()` internally and skips the `UNTIL` window check in that case, still enforcing `recurrence_unbounded` and `recurrence_count_out_of_range`. This is acceptable because the field is unreachable today; the safety net still rejects the production-incident shape (open-ended) and the count cap.

A small helper inside `resource.go` — `validateRecurrenceOrWriteError(...)` — keeps the create/update handlers symmetric and avoids duplicating the logging + response shape.

**New `services/calendar-service/internal/event/recurrence_validator_test.go`**
Table-driven tests cover:
- Open-ended `RRULE:FREQ=WEEKLY` → `recurrence_unbounded`.
- `RRULE:FREQ=WEEKLY;UNTIL=20310101T000000Z` with `eventStart = 2026-05-06` (5y + ~8mo) → `recurrence_too_long`.
- `RRULE:FREQ=WEEKLY;COUNT=0` → `recurrence_count_out_of_range`.
- `RRULE:FREQ=WEEKLY;COUNT=731` → `recurrence_count_out_of_range`.
- Case-insensitive: `rrule:freq=weekly;until=20260601t000000z` is accepted.
- `UNTIL` date-only form `20260601` is accepted.
- An `EXDATE:` line alongside a valid `RRULE:` does not trigger a false positive.
- Empty input slice → nil.

**Extended `services/calendar-service/internal/event/rest_test.go`** *(or new `resource_test.go` if handler tests don't currently live in `rest_test.go`)*
Handler-level tests using the existing test harness: create with open-ended → 422 + `code: recurrence_unbounded`; create with `UNTIL=5y+2d` → 422 + `code: recurrence_too_long`; create with `COUNT=731` → 422 + `code: recurrence_count_out_of_range`; create with `COUNT=10` happy path is accepted by the validator (the existing Google-call mocking determines whether it returns 201 or fails downstream); same matrix for PATCH where applicable.

### 3.3 Form Effects (detail)

The auto-update behavior in PRD §4.2 requires three `useEffect`s in `event-form-dialog.tsx`. They are spelled out here because the order matters and the existing dialog file is dense:

```
const previousRecurrenceRef = useRef(defaults.recurrence);

// Effect A+B: react to recurrence transitions
useEffect(() => {
  const prev = previousRecurrenceRef.current;
  if (prev === recurrence) return;

  if (prev === "" && recurrence !== "") {
    const startInstant = parseStart(startDate, startTime, allDay, timeZone);
    form.setValue("endsOnDate", formatDate(addYears(startInstant, 1)));
    form.setValue("endsOnDateUserEdited", false);
  } else if (prev !== "" && recurrence === "") {
    form.setValue("endsMode", "on");
    form.setValue("endsOnDate", "");
    form.setValue("endsAfterCount", 10);
    form.setValue("endsNeverConfirmed", false);
    form.setValue("endsOnDateUserEdited", false);
  }
  previousRecurrenceRef.current = recurrence;
}, [recurrence]);

// Effect C: startDate changed and end-date is still auto-default
useEffect(() => {
  if (recurrence !== "" && endsMode === "on" && !endsOnDateUserEdited) {
    const startInstant = parseStart(startDate, startTime, allDay, timeZone);
    form.setValue("endsOnDate", formatDate(addYears(startInstant, 1)));
  }
}, [startDate]);
```

The end-date `<Input type="date">` calls `form.setValue("endsOnDateUserEdited", true)` in its `onChange` before delegating to `field.onChange`. `previousRecurrenceRef` is initialized to the form's initial recurrence value, so on mount `prev === recurrence` and neither branch fires.

## 4. Data Flow

```
┌──────────────────────┐    composeRecurrenceRule()    ┌──────────────────────┐
│ event-form-dialog    │ ────────────────────────────▶ │ recurrence: string[] │
│ (handleSubmit)       │                               │ (length 1 or undef)  │
└──────────────────────┘                               └──────────┬───────────┘
                                                                  │ POST/PATCH
                                                                  ▼
                                                        ┌──────────────────────┐
                                                        │ resource.go handler  │
                                                        │ ValidateRecurrence() │ ── 422 ──▶ JSON:API error
                                                        └──────────┬───────────┘
                                                                   │ ok
                                                                   ▼
                                                        ┌──────────────────────┐
                                                        │ Processor → googlecal│
                                                        │ (unchanged)          │
                                                        └──────────────────────┘
```

Two validation layers, identical semantics, consistent origin and bounds. The frontend prevents the bad request; the backend protects against a buggy or bypassed frontend.

## 5. Error Handling

| Layer | Trigger | Surface |
|---|---|---|
| Zod schema | Out-of-range "On" / "After"; missing Never confirmation | Inline `<FormMessage />`; submit button disabled (existing mechanism) |
| Submit handler | Network / API error from `mutateAsync` | Existing `toast.error(createErrorFromUnknown(...))` path. The 422 codes from §1 of PRD surface via `error.message`; future work could map `code: recurrence_*` to a more specific toast, but that's out of scope |
| Backend handler | Malformed RFC 3339 in `input.Start` | Existing 400 path (no change) |
| Backend handler | `*RecurrenceError` | INFO log + `WriteJSONAPIError(422, code, "Validation Error", detail, "")` |
| Backend handler | Non-RRULE recurrence components only | Validator returns nil; request proceeds (per PRD §4.6) |

## 6. Testing Strategy

Already enumerated per file in §3. Coverage targets the three branches of §4.6 on the backend and the three modes × DST behavior on the frontend. No integration tests against Google Calendar are added; the existing harness for `createEventHandler` covers the wire-up.

## 7. Risks and Mitigations

| Risk | Mitigation |
|---|---|
| FE/BE 5-year cap drift | Identical formula on both sides; 1-day cushion absorbs DST / leap jitter (§2.5) |
| `UNTIL` UTC formatting wrong on DST boundary | `formatUntilUTC` derives offset for the *chosen date* via `Intl.DateTimeFormat` parts, not `Date.prototype.getTimezoneOffset()` of "now" |
| User confused by "Never" option | Inline warning + explicit checkbox + disabled submit (PRD §4.3) |
| Recurrence field reaching processor on PATCH | Field is `*[]string` and the existing processor (`internal/event/processor.go`) does not read it. No Google call carries a recurrence change. The validator runs purely as a boundary check |
| Library-less RRULE parser misses an edge case | Restricted scope (`UNTIL`/`COUNT` only, case-insensitive); both date and date-time `UNTIL` forms accepted; tests cover the matrix in §3.2 |
| Logging the RRULE leaks data | RRULE strings contain only schedule metadata, no user content; INFO is appropriate per PRD §8 |

## 8. Out of Scope

- Editing the recurrence rule of an existing series (task-013).
- Custom RRULE builder UI beyond the existing presets.
- Mapping `recurrence_*` codes to specific UI toasts.
- Backfill of pre-existing runaway series.
- Validating non-RRULE recurrence components.
