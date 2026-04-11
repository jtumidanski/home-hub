# Task 035 — Tasks Checklist

Last Updated: 2026-04-10 (updated after Bug 1 root cause confirmed)

Effort key: S = <2h, M = ½ day, L = 1 day, XL = >1 day.

## Phase 1 — Bug 2: Timezone-aware `/today`

- [x] **1.1** Add `internal/tz/resolver.go` to workout-service and tracker-service implementing `Resolve(ctx, headers, householdLookup) *time.Location`. Order: `X-Timezone` header → household timezone → UTC fallback + warn log. Cache resolved location on `context.Context`. *(M)*
  - Acceptance: unit tests cover header valid, header invalid, household fallback, UTC fallback. ✓
- [x] **1.2** workout-service `/today` wiring: read `X-Timezone` in `internal/today/resource.go`, compute "now"/`weekStart`/`dayOfWeek` in `*time.Location` inside `internal/today/processor.go:45`. Update existing tests; add the 21:30 ET / 02:30 UTC cross-midnight test case. *(S)*
- [x] **1.3** tracker-service `/today` wiring: same change for `internal/today/{processor.go:40, resource.go:26, processor_test.go}`. Add an analogous cross-midnight-UTC test. *(S)*
- [x] **1.4** Frontend `X-Timezone` header wired into `frontend/src/lib/api/client.ts` `buildHeaders` (used by every request). *(S)*
- [x] **1.5** Sweep confirmed: only `workout-service` and `tracker-service` expose `/today`. See Notes → `/today` endpoint sweep. *(S)*

## Phase 2 — Bug 1: Strength weight input

Root cause confirmed: stale data from user mis-click at creation time. `weightType` was set to `"bodyweight"` in the DB for the affected exercises. User has already corrected the bad rows by hand via SQL. No server-side defect exists. Remaining work is making the mistake **visible** so it doesn't silently recur.

- [x] **2.1** ~~Reproduce~~ — done. Root cause pinned as stale data. *(S)*
- [x] **2.2** ~~Inspect REST response~~ — done. `weightType: "bodyweight"` confirmed for both Concentration Curl (strength) and Long Run (cardio). *(S)*
- [x] **2.3** ~~Apply fix for the weightType rows~~ — user fixed directly in SQL. No code change for the data itself. *(S)*
- [x] **2.4** Weight-type badge added next to each strength exercise in `WorkoutExercisesPage.tsx`. *(S)*
- [x] **2.5** Week view "Add Exercise" dialog now has a two-step flow (pick → initial values). `ExerciseInitialValuesForm` collects Sets/Reps/Weight/Unit (or duration/distance for other kinds) and passes `planned` through `useAddPlannedItem`. *(S)*

## Phase 3 — Bug 3: Tracker range slider reset

- [x] **3.1** `RangeEditor` now takes `itemId`/`date` and the effect resets `local` to midpoint and `touched` to `false` when `initial === undefined`. Effect keyed on `(itemId, date, initial, min, max)`. *(S)*
- [x] **3.2** Same reset applied to `RangeInput` in `today-view.tsx`. *(S)*
- [ ] **3.3** Manual verification pending (requires running app). *(S)*

## Phase 4 — Bug 4: UUID-in-trigger regression

- [x] **4.1** ~~Reproduce~~ — pinned by user: Week view "Add Exercise" dialog theme/region filter selects at `WorkoutWeekPage.tsx:523, 536`. See "Notes → Bug 4 pinned." *(S)*
- [x] **4.2** Added explicit `themeLabel`/`regionLabel` children to the Theme/Region `SelectValue`s in `WorkoutWeekPage.tsx`. Sweep confirmed no other `placeholder="Theme"`/`placeholder="Region"` sites remain. *(S)*
- [ ] **4.3** Manual verification pending (requires running app). *(S)*

## Phase 5 — Verification & Build

- [x] **5.1** `go test ./...` green for `workout-service` and `tracker-service`. *(S)*
- [x] **5.2** Frontend `tsc -b` + `vitest run` (44 files / 404 tests) green. Lint parity preserved — same 3 pre-existing warnings/errors on `calendar-grid.tsx`/`today-view.tsx`, no new ones introduced. *(S)*
- [ ] **5.3** `scripts/local-up.sh` clean rebuild + smoke tests — pending. *(S)*
- [ ] **5.4** `X-Timezone` DevTools verification — pending (requires running app). *(S)*
- [ ] **5.5** Manual tick of PRD §10 criteria — pending live verification. *(S)*

## Notes

### Bug 1 reproduction (2026-04-10, from user)
- Reproduced on **desktop** — rules out the mobile-layout-clipping hypothesis.
- Reproduced with "concentration curl" added to the Week view via the Add Exercise dialog.
- Symptom in **Week view edit** and **Today view** for this exercise: only Sets and Reps are editable; no Weight input is rendered at all (not clipped — absent).
- Also observed: the Add Exercise dialog in Week view only lets the user pick theme/region/exercise — it does NOT accept initial sets/reps/weight when queueing an exercise into the week. User flagged this as a separate bug to fix alongside Bug 1.
- Plumbing check: `services/workout-service/internal/weekview/rest.go:67` serializes `weightType` (JSON tag `weightType`); `projection.go:58` sets it from `ex.WeightType`; the today processor reuses the same `ItemRest` struct; frontend `WeekItem` type includes `weightType: WeightType`. End-to-end wiring exists.
- Frontend gate: `WorkoutTodayPage.tsx:60` and `WorkoutWeekPage.tsx:330` both do `isBw = item.weightType === "bodyweight"`. Behaving correctly — the problem was the DB value, not the gate.
- **Confirmed (user, 2026-04-10):** DB showed `weightType = "bodyweight"` for both Concentration Curl (strength) and Long Run (cardio). User input error at creation time, not a code bug. User fixed the bad rows directly via SQL.
- **Resolution (user, 2026-04-10):** Option A + B from the recovery conversation — user did the SQL fix, task 2.4 will surface `weightType` on the Exercises list so future mis-clicks are visible. `WeightTypeImmutable` stays in place; the broader conversation about conditional immutability is deferred.

### Bug 4 pinned (2026-04-10, from user)
- Offender is the **Week view "Add Exercise" dialog** theme/region filter selects — exactly `WorkoutWeekPage.tsx:523` (Theme) and `:536` (Region).
- Confirmed by reading the file: both use `<SelectValue placeholder="Theme|Region" />` with no children fallback. If `themes.data?.data` or `regions.data?.data` is undefined on the render immediately after selection (likely due to async loading), Radix's `SelectItem` list is empty, so Radix falls back to displaying the raw UUID value.
- Fix: add explicit children to each `<SelectValue>` — `{themes.data?.data?.find(t => t.id === themeId)?.attributes.name ?? "All themes"}` / same for regions. No changes needed to the `WorkoutExercisesPage.tsx` create-exercise dialog (already fixed in task-032).

### Add-Exercise dialog missing sets/reps/weight (new sub-bug raised by user)
- In Week view, the Add Exercise dialog only captures theme/region/exercise. It should also accept initial sets/reps/weight so the user can queue a planned exercise in one shot rather than add-then-edit.
- **Decision needed:** fold into this task (Phase 2 expansion) or file as follow-up. Leaning toward folding in since it's adjacent to Bug 1 and the same dialog surface.

### `/today` endpoint sweep
- Confirmed 2026-04-10 during implementation: `/today` is only served by `services/workout-service/internal/today/resource.go` and `services/tracker-service/internal/today/resource.go`. No productivity/meals/calendar equivalents exist.
