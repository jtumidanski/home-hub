# Plan Audit — task-047-recurring-event-end-date

**Plan Path:** docs/tasks/task-047-recurring-event-end-date/plan.md
**Audit Date:** 2026-05-03
**Branch:** feature/task-047-recurring-event-end-date
**Base Branch:** main (commit 804881c)

## Executive Summary

All 14 plan tasks were implemented. Every expected file in the plan was created or
modified, the new acceptance-criteria-driven behavior is present in code, and both
frontend and backend test suites pass cleanly (Vitest: 547/547; `go test
./services/calendar-service/...`: all packages OK). The four caller-disclosed
deviations (Zod `z.number()` instead of `z.coerce.number()`, the `form.tsx`
`useFormField` refactor, the disambiguating `selector` in the form-dialog tests,
and the in-line wiring of Tasks 11-13) are all valid, well-scoped adaptations
that do not weaken the spec; in particular, the schema deviation is the correct
fix for Zod v4 input/output divergence under react-hook-form's `Resolver`
invariant. Recommendation: ready to merge.

## Task Completion

| #  | Task                                                       | Status | Evidence / Notes |
|----|------------------------------------------------------------|--------|------------------|
| 1  | Create `recurrence.ts` skeleton + `EndsMode` + `eventStartInstant` | DONE | `frontend/src/lib/calendar/recurrence.ts:1-13`; tests at `frontend/src/components/features/calendar/__tests__/recurrence.test.ts`; commit `dd16050` |
| 2  | `formatUntilUTC` with DST-correct offset                   | DONE | `frontend/src/lib/calendar/recurrence.ts:19-64`; tests cover EST/EDT/UTC; commit `e4b8d28` |
| 3  | `composeRecurrenceRule` covering all three modes + presets | DONE | `frontend/src/lib/calendar/recurrence.ts:66-84`; commit `4a6b086`. Interim type-safety fix `0ce5e5a` (per disclosure #3) is mechanical hardening for `noUncheckedIndexedAccess` and does not alter behavior |
| 4  | Extend `eventFormSchema` with Ends fields and refinements  | DONE (with disclosed deviation) | `frontend/src/lib/schemas/calendar-event.schema.ts:5-69` (5 refinements as planned); test `frontend/src/lib/schemas/__tests__/calendar-event.schema.test.ts` (81 lines, 9 specs all pass); commit `c597024`. Deviation: line 20 uses `z.number().int()` instead of `z.coerce.number().int()` — disclosed; the form-dialog input passes `Number(e.target.value)` (`event-form-dialog.tsx:405`) so the integer arrives pre-coerced; all 9 schema tests pass with no changes needed |
| 5  | Render the "Ends" control + warning checkbox               | DONE | `frontend/src/components/features/calendar/event-form-dialog.tsx:320-466`; renders only when `!isEdit` and `recurrence !== ""`; commit `aa2138a` |
| 6  | Add the three auto-update effects                          | DONE | `event-form-dialog.tsx:94-139` — `useRef`+`startDate`+`endsOnDateUserEdited` watches and the two `useEffect` blocks (recurrence-toggle reset, start-date propagation) match the plan; commit `d1b7f00` |
| 7  | Wire `composeRecurrenceRule` into the submit handler       | DONE | `event-form-dialog.tsx:13` import; `event-form-dialog.tsx:200-208` call in the create branch; commit `5fb2a47` |
| 8  | Form-level interaction tests                               | DONE (with disclosed deviation) | `frontend/src/components/features/calendar/__tests__/event-form-dialog.test.tsx` (215 lines, 9 specs incl. the planned 8 plus the "resets Ends fields" test). All pass. Deviations (disclosed): tests use `getByLabelText("End date", { selector: "input[aria-label='End date']" })` to disambiguate from the event-level End-date FormLabel; `form.tsx` was modified (`frontend/src/components/ui/form.tsx:38-67`) to consume `FormItemContext` for stable IDs (canonical shadcn pattern, fixes a pre-existing label/input association bug); commit `3da09a6` |
| 9  | `parseRRULE` (case-insensitive UNTIL/COUNT extraction)     | DONE | `services/calendar-service/internal/event/recurrence_validator.go:32-72`; tests `recurrence_validator_test.go:8-58` cover all 7 planned cases; commit `f8db6b6` |
| 10 | `ValidateRecurrence` covering all three error codes        | DONE | `recurrence_validator.go:74-112`; tests `recurrence_validator_test.go:70-112` cover all 13 planned cases including zero-start skip; commit `6da584e` |
| 11 | Add `Recurrence` field to `UpdateEventRequest`             | DONE | `services/calendar-service/internal/event/rest.go:45` — `Recurrence  *[]string \`json:"recurrence"\``; commit `a87e8dc` |
| 12 | Wire validator into `createEventHandler`                   | DONE | Helper at `resource.go:222-232`; call site at `resource.go:114-122`; handler-scenario test in `recurrence_validator_test.go:114-142`; commit `db75d92`. Per disclosure #4, the controller applied the change in-line rather than via subagent — no behavior delta |
| 13 | Wire validator into `updateEventHandler`                   | DONE | `resource.go:156-166` — guarded by `if input.Recurrence != nil`, parses optional RFC3339 start, calls `validateRecurrenceOrWriteError`; commit `2c09e17` |
| 14 | Whole-stack verification                                   | DONE | This audit re-ran both build commands and both test suites; all green (see Build & Test Results below). Docker build step (Task 14 Step 3) was not re-run by the auditor; per the plan's own self-note, the task does not touch shared libraries so it is a defensive check, and the equivalent compile invariant is exercised by `go build ./services/calendar-service/...` passing |

**Completion Rate:** 14/14 tasks (100%)
**Skipped without approval:** 0
**Partial implementations:** 0

## Skipped / Deferred Tasks

None. Two interim fix commits (`0ce5e5a` recurrence type-safe indexing, `966e522`
test-factory model-shape alignment) are not in the plan but are pure
sustainability fixes that do not change behavior — both are explicitly disclosed
in the audit request (deviation #3).

## Build & Test Results

| Service / Workspace                | Build | Tests | Notes |
|------------------------------------|-------|-------|-------|
| services/calendar-service (Go)     | PASS  | PASS  | `go build ./services/calendar-service/...` clean; `go test ./services/calendar-service/... -count=1` — all 9 packages OK including `internal/event` (which contains `TestParseRRULE`, `TestValidateRecurrence`, `TestValidateRecurrence_HandlerScenarios`) |
| frontend (Vite/TS)                 | PASS  | PASS  | `npm --prefix frontend run build` (`tsc -b && vite build`) clean; `npm --prefix frontend test` — 75 test files / 547 tests pass, no failures |

## Plan Adherence Verification — Acceptance Criteria (PRD §10)

| §10  | Acceptance criterion | Implemented as |
|------|----------------------|----------------|
| 10.1 | Ends control appears for recurring | `event-form-dialog.tsx:344` (`recurrence !== "" &&`); test "hides..." / "shows..." |
| 10.2 | Default On + start+1y | `event-form-dialog.tsx:122-124` (recurrence-on -> sets `endsOnDate = addOneYear(startDate)`, `endsOnDateUserEdited = false`); test "shows the Ends control and seeds end date to start + 1y" |
| 10.3 | Auto-update vs user-edited end date | `event-form-dialog.tsx:135-139` plus `:374-377` set `endsOnDateUserEdited = true` on edit; tests "auto-updates" and "leaves a user-edited end date alone" |
| 10.4 | Never warning + confirmation | `event-form-dialog.tsx:441-462`; schema refinement at `calendar-event.schema.ts:63-68`; test "blocks submit until the Never confirmation checkbox is checked" |
| 10.5 | On < start error | Schema refinement `calendar-event.schema.ts:35-44`; schema test "rejects mode=on with an end date before the start date" |
| 10.6 | On > 5y error | Schema refinement `calendar-event.schema.ts:46-54`; schema test "rejects mode=on with an end date more than 5 years out" |
| 10.7 | After count out of [1,730] | Schema refinement `calendar-event.schema.ts:56-61`; schema tests "count = 0" / "count = 731" |
| 10.8 | Weekly May6→Jun10 finite | Test "submits an UNTIL-terminated RRULE for mode=on" asserts `RRULE:FREQ=WEEKLY;UNTIL=20260611T...Z` |
| 10.9 | Daily After 5 | Test "submits a COUNT-terminated RRULE for mode=after" asserts `["RRULE:FREQ=DAILY;COUNT=5"]` |
| 10.10 | UTC conversion non-UTC zones | `formatUntilUTC` at `recurrence.ts:19-64`; tests "EST winter", "EDT summer", "UTC" |
| 10.11 | 422 `recurrence_unbounded` | `recurrence_validator.go:87-92`; test cases "open-ended" + "malformed UNTIL -> unbounded" |
| 10.12 | 422 `recurrence_too_long` | `recurrence_validator.go:101-108`; test case "until > 5y" + handler scenario "until 5y+2d" |
| 10.13 | 422 `recurrence_count_out_of_range` | `recurrence_validator.go:94-100`; test cases "count zero" / "count 731" |
| 10.14 | FE composition coverage | `recurrence.test.ts` 5 specs covering all three modes + DST + empty preset |
| 10.15 | BE handler tests for 422 paths | Substituted by `TestValidateRecurrence_HandlerScenarios` at the validator boundary, as design §3.2 anticipates (no httptest harness exists in this package); covers all three error codes plus 2 happy paths |
| 10.16 | No DB migration | None present in diff; no `*.sql` in changed files |
| 10.17 | Edit dialog unchanged | `event-form-dialog.tsx:320` and `:496` both gated on `!isEdit` so neither the recurrence dropdown nor the Ends control renders in edit mode |

## Disclosed Deviations — Assessment

1. **`z.number().int()` vs plan's `z.coerce.number().int()`** — Acceptable. The
   `Resolver` invariant in react-hook-form requires input and output types to be
   identical, and Zod v4's `coerce` helpers split those types. The form-dialog
   already coerces at the React boundary
   (`event-form-dialog.tsx:405` — `onChange={(e) => countField.onChange(Number(e.target.value))}`),
   so by the time the value reaches the schema it is already a number. All 9
   schema specs and all 9 form-dialog interaction specs pass with this shape.

2. **`form.tsx` `useFormField` refactor + dialog markup adjustment** —
   Acceptable. The original `useFormField` called `React.useId()` per call,
   which produced a fresh ID for every consumer in a `<FormItem>` (the
   `<label htmlFor>` and the input received different IDs), breaking
   `getByLabelText` for any field rendered inside a wrapping radio `<label>`.
   The fix (consume `FormItemContext` for a stable ID) matches the canonical
   shadcn pattern and corrects a pre-existing latent bug. The dialog's radio
   `<label>` was split off the date/number inputs to keep label association
   unambiguous; the `aria-label="End date"` selector in the tests is needed
   because the event-level "End date" `FormLabel` (line 296) shares the same
   text and would otherwise be ambiguous.

3. **Two interim sustainability commits (`0ce5e5a`, `966e522`)** — Acceptable.
   The first hardens `formatUntilUTC` against `noUncheckedIndexedAccess` and the
   second updates the test factory to current `CalendarConnection`/
   `CalendarSource` shapes (the plan's example used outdated field names).
   Neither alters behavior or test assertions.

4. **Tasks 11-13 applied inline rather than via subagent** — Acceptable. The
   resulting code matches the plan literally (struct field, helper signature,
   call sites, RFC3339 start parsing, `connID` passed to log/error builder); no
   behavior delta.

## Overall Assessment

- **Plan Adherence:** FULL
- **Recommendation:** READY_TO_MERGE

## Action Items

None required. Optional follow-up suggestions (non-blocking):

1. The `form.tsx` `useFormField` change has codebase-wide implications because
   every form-control consumer relies on it. The frontend test suite (547 tests)
   passes, suggesting no regressions, but a quick visual smoke of any form view
   that nests inputs inside wrapping `<label>` elements would be worth a
   follow-up issue.
2. Consider whether `endsAfterCount` should round-trip a coercion at the schema
   level for consistency with other numeric fields elsewhere in the codebase
   (defensive only — the input handler already covers the only entry point).
