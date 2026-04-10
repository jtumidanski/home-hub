# Workout Tracker Bug Fixes — Tasks

Last Updated: 2026-04-10

---

## Bug 1: Summary empty-week latency

- [ ] **1.1** Add `retry: false` to `useWorkoutWeekSummary` in `frontend/src/lib/hooks/api/use-workouts.ts:317` — match the pattern from `useWorkoutWeek` (line 173) `[S]`
- [ ] **1.2** Verify: navigate to summary page for a week with no workouts → "No summary available" message appears immediately (< 1s) `[S]`

## Bug 2: UUID display in New Exercise dialog

- [ ] **2.1** In `ExerciseCreateDialog` (`frontend/src/pages/WorkoutExercisesPage.tsx`), update Theme `SelectValue` (line 187) to render the resolved theme name as explicit children `[S]`
- [ ] **2.2** Same file — update Primary Region `SelectValue` (line 202) to render the resolved region name as explicit children `[S]`
- [ ] **2.3** Verify: open New Exercise dialog, select a theme → trigger shows theme name, not UUID `[S]`
- [ ] **2.4** Verify: select a region → trigger shows region name, not UUID `[S]`

## Bug 3: Start Fresh flow

- [ ] **3.1** In `WorkoutWeekPage` (`frontend/src/pages/WorkoutWeekPage.tsx`), replace `isEmpty` (line 59) with `weekNotFound = !!week.error` `[S]`
- [ ] **3.2** Update JSX conditional (line 145) to use `weekNotFound` instead of `isEmpty` `[S]`
- [ ] **3.3** Verify: click Start Fresh on an unprovisioned week → toast appears, planner grid renders with 7 day columns and "+ Add exercise" buttons `[S]`
- [ ] **3.4** Verify: Copy Planned / Copy Actual still work correctly from the empty-week prompt `[S]`
- [ ] **3.5** Verify: navigating to a week that already has items still shows the planner grid normally `[S]`

## Final

- [ ] **4.1** Run frontend build (`npm run build` or equivalent) — no type errors or warnings `[S]`
- [ ] **4.2** Create PR with all three fixes `[S]`
