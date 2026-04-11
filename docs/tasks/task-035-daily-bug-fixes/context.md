# Task 035 — Context

Last Updated: 2026-04-10

## Key Files

### Bug 1 — Strength weight input
- `frontend/src/pages/WorkoutTodayPage.tsx:167–196` — strength row, `grid-cols-4` when `!isBw`. Weight + Unit live inside `!isBw &&`.
- `frontend/src/pages/WorkoutWeekPage.tsx:384–401` — inline planner edit row for strength.
- `services/workout-service/internal/weekview/projection.go:58` — `WeightType` projection from `ex.WeightType`. To rule out as root cause.
- `services/workout-service/internal/exercise/builder.go:22` — `WeightTypeImmutable` rule. **Do not** lift this in scope.

### Bug 2 — Timezone-aware `/today`
- `services/workout-service/internal/today/processor.go:45` — `time.Now().UTC().Truncate(24 * time.Hour)`.
- `services/workout-service/internal/today/resource.go` — HTTP boundary; will read `X-Timezone`.
- `services/tracker-service/internal/today/processor.go:40` — same UTC pattern.
- `services/tracker-service/internal/today/resource.go:26` — HTTP boundary.
- `services/tracker-service/internal/today/processor_test.go` — existing tests to update.
- `services/account-service/internal/household/rest.go:12` — household model already exposes `Timezone`. Source for fallback.
- `frontend/src/services/api/` — axios instance + existing interceptors. Add `X-Timezone` here.
- **Confirmed scope:** grep of `services/**/today/*.go` finds only workout-service and tracker-service `/today` endpoints. Re-grep at implementation time.

### Bug 3 — Tracker range slider
- `frontend/src/components/features/tracker/calendar-grid.tsx:356` — `RangeEditor` component.
- `frontend/src/components/features/tracker/calendar-grid.tsx:359–364` — buggy `useEffect` keyed only on `initial`, early-returns on undefined.
- `frontend/src/components/features/tracker/today-view.tsx:165, 169` — `RangeInput` with the same pattern, fix defensively.

### Bug 4 — Exercise dialog UUID regression
- `frontend/src/pages/WorkoutExercisesPage.tsx:189–191, 206–208` — already has the `find().attributes.name` fix from task-032 §4.2.
- `frontend/src/pages/WorkoutWeekPage.tsx:523, 536` — Add Exercise filter `<SelectValue placeholder="Theme/Region" />` with no children fallback. Most likely culprit.

## Key Decisions

- **Bundled task, single PR.** All four bugs were observed in the same dogfooding session and have surgical fixes. No separate branches.
- **Hybrid timezone strategy.** Header `X-Timezone` overrides the household default; UTC is only the last-resort fallback. Implementer chooses Option A (per-service helper) or Option B (middleware-attached `*time.Location`); Phase 1.1 of `plan.md` recommends Option A — small `internal/tz/resolver.go` per service — to avoid changes to shared middleware.
- **No schema, no migrations.** Household timezone field already exists.
- **`WeightTypeImmutable` stays.** Even if Bug 1 root cause is stale data, the fix path is a one-shot data correction in a follow-up task, not lifting the immutability rule.
- **Reproduce before patching Bugs 1 and 4.** Both have unknown root causes; do not guess.
- **Backend-first ordering** in Phase 1 reduces risk; the frontend `X-Timezone` interceptor is independently safe to ship since unaware services ignore the header.

## Dependencies

- Existing inter-service account-service client (for household timezone lookup).
- Existing tenant/household middleware for context plumbing.
- `scripts/local-up.sh` for local verification.
- A test household with a non-UTC timezone (or manual `X-Timezone` header).

## Out of Scope

- Per-user (vs. household) timezone preference UI.
- Shared cross-service timezone helper library.
- Workout logging form rework beyond minimum needed for Weight inputs.
- Backfilling historical workout logs missed because of the UTC bug.
- Lifting `WeightTypeImmutable`.
- Backfill / one-shot data correction for stale `weightType` (file as follow-up if needed).

## References

- PRD: `prd.md`
- Plan: `plan.md`
- Prior task referenced: task-032 §4.2 (original UUID fix)
- Project rules: `/home/tumidanski/source/home-hub/CLAUDE.md` — verify Docker builds, run tests for all affected services before reporting completion.
