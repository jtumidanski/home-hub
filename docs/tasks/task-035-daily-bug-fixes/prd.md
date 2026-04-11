# Daily Use Bug Fixes — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-10
---

## 1. Overview

Four bugs surfaced during a single day of dogfooding the app. Each is small in scope but actively breaks a daily workflow:

1. **Weight is uneditable on weight-bearing strength exercises.** On both `WorkoutTodayPage` and the `WorkoutWeekPage` planner row editor, only Sets and Reps are presented for the user's strength exercises — the Weight input is missing even though the exercise is configured as a weight-bearing (free-weight) movement.
2. **`/today` endpoints compute "today" in UTC.** A workout finished at ~8 PM Eastern is past midnight UTC, so the `Today` projection looks at the wrong day-of-week and the user's planned workout disappears. Affects `workout-service`, `tracker-service`, and any other service that exposes a `/today` endpoint.
3. **Tracker range slider retains the previous day's value.** When navigating between days in the tracker month grid (`features/tracker/calendar-grid.tsx`), the `RangeEditor` popover keeps the previously rendered day's slider value rather than resetting to "not set" for an unentered day.
4. **Add/Edit Exercise dialog shows raw UUIDs in Theme/Region triggers.** Despite the prior fix in task-032 §4.2, the Theme and Primary Region `SelectTrigger`s still display UUIDs instead of human-readable names in at least one of the exercise dialogs. The regression needs to be hunted down and re-fixed.

All four are in scope for a single bundled task because they were all observed in the same dogfooding session and have small, surgical fixes.

## 2. Goals

Primary goals:
- Make Weight editable everywhere a strength exercise's actuals or planned values can be entered, when the exercise is `weightType = "free"`.
- Make `/today` projections honor the household's local timezone, with an `X-Timezone` request header overriding the household default when present (hybrid model).
- Reset the tracker range slider's local state when the active day changes so a fresh day starts from the midpoint placeholder, matching the existing "not set" behavior.
- Re-fix the UUID-in-trigger regression in the exercise dialog(s) so the dropdowns always display names.

Non-goals:
- Designing a per-user (vs. household) timezone preference UI.
- Building a shared `account-service` timezone helper library — the resolution can be inlined per service for now.
- Reworking the workout logging form layout beyond the minimum needed to fit a Weight input.
- Backfilling or correcting historical workout logs that were missed because of the UTC bug.

## 3. User Stories

- As a user logging today's strength workout, I want to enter the weight I lifted alongside sets and reps so my actuals are accurate.
- As a user editing a planned strength exercise on the weekly planner, I want to set the planned weight from the inline edit row so I don't have to dig into a separate dialog.
- As a user finishing a workout in the evening, I want it to still count as "today" so I don't lose credit for the session.
- As a user filling in tracker entries day-by-day, I want each new day to start with a fresh slider so I don't accidentally save yesterday's value as today's.
- As a user creating or editing an exercise, I want the Theme and Primary Region dropdowns to show the names I selected, not opaque UUIDs.

## 4. Functional Requirements

### 4.1 Weight Input on Strength Exercises (Bug 1)

Two locations are affected:

- **`frontend/src/pages/WorkoutTodayPage.tsx`** — the expanded actuals form for `kind === "strength"` (lines 167–196). The Weight input and Unit select are gated behind `!isBw` where `isBw = item.weightType === "bodyweight"`.
- **`frontend/src/pages/WorkoutWeekPage.tsx`** — the inline planner edit row for `kind === "strength"` (lines 384–401). Same `!isBw` gating.

Required behavior:

- For any strength exercise where `item.weightType !== "bodyweight"` (i.e., `"free"` or any other future variant), the Weight input and Weight Unit selector MUST render and be editable.
- The investigator must determine why these inputs are not currently visible for the user. Likely root causes to check, in order:
  1. **Server projection of `weightType`**: confirm `services/workout-service/internal/weekview/projection.go:58` is correctly populating `WeightType` from `ex.WeightType` for the user's existing exercises. If older exercises have an empty string in the column (pre-default), the frontend `=== "bodyweight"` check is false and weight should show — so this is unlikely to be the cause but should be ruled out.
  2. **Frontend grid layout**: the strength row uses `grid-cols-4` on `WorkoutTodayPage` and `grid-cols-2` on `WorkoutWeekPage`. On a narrow mobile viewport, a 4-column grid may visually clip the third and fourth columns (Weight, Unit) below the visible area, or render them so narrow they appear absent. The grid must reflow to at least two rows on mobile so all four inputs are reachable.
  3. **Existing exercise data**: if the user's exercises were created before `weightType` had a default and contain a stale value (e.g., literal `"bodyweight"` from a buggy create flow), the fix is a one-time data correction plus making the weight type editable post-creation (currently `WeightTypeImmutable` per `services/workout-service/internal/exercise/builder.go:22`). **Do not** lift the immutability rule as part of this task — instead, surface the user's existing weightType in the exercise list so they can verify, and if a correction is needed, do it via a manual SQL fix or a follow-up task.

The acceptance check is: open a barbell/dumbbell strength exercise on both pages on a mobile viewport (<768px) and confirm Weight + Unit are visible and editable.

### 4.2 Timezone-aware `/today` Endpoints (Bug 2)

Affected services (any service exposing a `/today` or equivalent local-day endpoint):

- `services/workout-service/internal/today/processor.go:45` — currently does `now.UTC().Truncate(24 * time.Hour)`.
- `services/tracker-service/internal/today/processor.go:40` and `resource.go:26` — same UTC truncation pattern.
- Any other `/today` endpoint discovered during implementation. The implementation pass MUST grep for `time.Now().UTC()` near `today` packages and confirm coverage. Suspected candidates: `meals` / `productivity` (verify during implementation; if absent, no change).

Required behavior:

- A new shared helper resolves the effective timezone for a request using the **hybrid** strategy:
  1. If the request carries an `X-Timezone` header with a value that parses via `time.LoadLocation`, use it.
  2. Otherwise, fall back to the household timezone from `account-service` (already on the household model — see `services/account-service/internal/household/rest.go:12`).
  3. If neither is resolvable, fall back to UTC and log a warning (do not error).
- The helper should live somewhere reusable. Two acceptable options — implementer's choice:
  - **Option A**: a small helper package inside each service that calls account-service via the existing inter-service client to fetch the household's timezone. Cache per-request.
  - **Option B**: enrich the existing tenant/household middleware to attach a resolved `*time.Location` to the request context, so processors can call `tz.FromContext(ctx)` without extra plumbing.
- The frontend MUST attach `X-Timezone: <Intl.DateTimeFormat().resolvedOptions().timeZone>` to every API request via the existing axios interceptor in `frontend/src/services/api/`.
- All `Today` processors must accept an explicit `time.Location` (or read it from context) and compute "today" as `time.Now().In(loc).Truncate(...)` rather than `time.Now().UTC().Truncate(...)`. The same goes for `weekStart` derivation and `dayOfWeek` calculation — they must all be computed in the resolved local zone.
- The change must be reflected in the existing frontend "today" widgets that read these endpoints; no behavioral changes are expected because the server result is the source of truth.

Out of scope: per-user timezone preferences, UI for editing the household timezone (it already exists at the account-service level), and migrating historical entries to a different day.

### 4.3 Tracker Range Slider Reset (Bug 3)

- **`frontend/src/components/features/tracker/calendar-grid.tsx:356`** — `RangeEditor` component.
- The `useEffect` at lines 359–364 only updates `local` when `initial !== undefined`. When the user opens a popover for a new day where no entry exists, `initial` is undefined, so `local` keeps the value from the last cell that had an entry, AND `touched` stays `true` because the previous cell flipped it.

Required behavior:

- When the popover opens for a cell with no existing entry (`initial === undefined`), `local` MUST reset to `Math.round((min + max) / 2)` (the midpoint) and `touched` MUST reset to `false`, displaying "Not set" until the user moves the slider.
- The same fix should be applied defensively to `RangeInput` in `frontend/src/components/features/tracker/today-view.tsx:165` if it has the same pattern (it does — line 169).
- The fix should key off the `(itemId, date)` pair so navigating between days (or between items on the same day) always re-evaluates initial state.

### 4.4 Exercise Dialog UUID Regression (Bug 4)

- The fix from task-032 §4.2 (`themes.find((t) => t.id === themeId)?.attributes.name` inside `SelectValue`) is currently in place at `frontend/src/pages/WorkoutExercisesPage.tsx:189–191` and `:206–208`.
- The user is still seeing UUIDs in the trigger after selecting a value. The implementation pass must:
  1. Reproduce the bug — note the exact dialog (Create vs Edit, if an Edit dialog exists; the WorkoutWeekPage exercise picker filter; etc.) and the exact reproduction steps.
  2. Identify which `Select` is rendering the UUID. Candidate locations beyond the create dialog: the WeekPage `Add Exercise` filter selects (`WorkoutWeekPage.tsx:523, 536`) — these use `<SelectValue placeholder="Theme" />` with no explicit children fallback, so if the Radix `SelectItem` for the active value isn't mounted at the moment of value selection (e.g., due to async loading), Radix falls back to displaying the raw value.
  3. Apply the same `find().attributes.name` children pattern to every `Select` whose value is a UUID, including the WeekPage filters and any Edit-exercise dialog if one exists.

Acceptance: in every dialog where a Theme or Region is selected, the trigger shows the human-readable name within one render of the selection.

## 5. API Surface

- **New header (Bug 2)**: All HTTP endpoints in `workout-service`, `tracker-service`, and any other service identified during implementation accept an optional `X-Timezone` header containing an IANA timezone identifier (e.g., `America/New_York`). Invalid values are silently ignored and the household-default fallback is used.
- No new endpoints, no breaking request/response shape changes.

## 6. Data Model

- No schema changes.
- No data migrations. (See §4.1 note about not touching the `weightType` immutability rule.)

## 7. Service Impact

| Service | Change |
|---------|--------|
| **frontend** | Bug 1: surface Weight + Unit inputs for non-bodyweight strength exercises on `WorkoutTodayPage` and `WorkoutWeekPage` row editor; reflow grid for mobile if necessary. Bug 2: attach `X-Timezone` header in axios interceptor. Bug 3: reset `RangeEditor` and `RangeInput` local state on cell change. Bug 4: re-fix UUID regression in all relevant exercise dialogs. |
| **workout-service** | Bug 2: add timezone resolution helper / context plumbing; update `today/processor.go` and `today/resource.go` to use the resolved location for "now" and for `dayOfWeek` / `weekStart`. |
| **tracker-service** | Bug 2: same change as workout-service for `today/processor.go` and `today/resource.go`. |
| **other `/today` endpoints** | Bug 2: same change if discovered during implementation grep. |
| **account-service** | None directly. The household timezone field already exists; no schema or endpoint changes. |

## 8. Non-Functional Requirements

- The TZ resolution helper must not add a synchronous account-service round-trip on every request. If the household lookup is needed, it should be cached for the lifetime of the request context (or longer, if a per-tenant cache already exists in the service).
- Invalid `X-Timezone` values must never cause a 5xx — fall back gracefully with a warn-level log.
- All existing tests for `today` processors must be updated to pass an explicit `*time.Location` and continue to pass. Add at least one new test case per affected processor: "workout finished at 23:30 UTC but 19:30 Eastern is still today's day-of-week."
- Multi-tenancy: the timezone resolution is per-household (via `tenant_id` → `household_id`), consistent with existing tenant scoping rules.

## 9. Open Questions

- Bug 1: is the root cause #1 (server projection), #2 (mobile layout), or #3 (stale exercise data)? The implementation pass should reproduce the bug and pin it before applying a fix. If it turns out to be #3, the user needs a UI to verify or correct `weightType` on existing exercises — that may warrant a follow-up task rather than expanding scope here.
- Bug 4: which dialog is the offender? Open during reproduction.

## 10. Acceptance Criteria

- [ ] On a mobile viewport, opening a non-bodyweight strength exercise on `WorkoutTodayPage` shows Sets, Reps, Weight, and Unit inputs, all editable, and "Log & Done" persists the entered weight.
- [ ] On `WorkoutWeekPage`, clicking the inline edit pencil for a non-bodyweight strength exercise shows Sets, Reps, Weight, and Unit inputs, all editable, and Save persists the planned weight.
- [ ] A workout completed at 8 PM local time on a workout day still appears in the `GET /workouts/today` projection at 8 PM local time, regardless of the user's offset from UTC.
- [ ] Sending `X-Timezone: America/New_York` to `GET /workouts/today` at 03:30 UTC on a Tuesday returns Monday's items (because it is still Monday 22:30 in NYC).
- [ ] Sending an invalid `X-Timezone` (e.g., `Mars/Olympus`) does not 5xx; the response uses the household timezone or UTC fallback.
- [ ] The same TZ-aware behavior holds for `GET /trackers/today` and any other `/today` endpoints discovered during implementation.
- [ ] Opening the tracker month grid range-slider popover on a day with no entry shows the slider at the midpoint with "Not set" displayed, regardless of which cell was edited last.
- [ ] Selecting a Theme or Region in the New Exercise dialog (and any Edit dialog or filter that uses a UUID-valued Select) displays the human-readable name in the trigger within one render.
- [ ] All existing tests in `workout-service`, `tracker-service`, and the frontend continue to pass.
- [ ] At least one new test per affected `today` processor covers the late-evening / cross-midnight-UTC case.
