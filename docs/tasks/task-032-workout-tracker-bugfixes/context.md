# Workout Tracker Bug Fixes — Context

Last Updated: 2026-04-10

---

## Key Files

| File | Relevance |
|------|-----------|
| `frontend/src/lib/hooks/api/use-workouts.ts` | Bug 1 — `useWorkoutWeekSummary` (line 315) needs `retry: false` |
| `frontend/src/pages/WorkoutExercisesPage.tsx` | Bug 2 — `ExerciseCreateDialog` (lines 107-220), Theme select (lines 183-196), Region select (lines 198-211) |
| `frontend/src/pages/WorkoutWeekPage.tsx` | Bug 3 — `isEmpty` logic (line 59), `EmptyWeek` component (lines 221-251), planner grid conditional (line 145) |
| `frontend/src/pages/WorkoutSummaryPage.tsx` | Bug 1 — summary page error display (line 15); no changes needed here |
| `services/workout-service/internal/summary/rest.go` | Bug 1 — backend returns 404 for missing weeks (line 44); no changes needed |

## Existing Patterns

- `useWorkoutWeek` at `use-workouts.ts:173` already uses `retry: false` for the same 404-as-empty-state pattern. Bug 1 fix follows this precedent.
- The exercise list in `WorkoutExercisesPage` (line 59) already resolves theme/region names from IDs for the list display: `themes.data?.data.find((t) => t.id === e.attributes.themeId)`. Bug 2 fix applies the same lookup pattern to the create dialog's selects.

## Dependencies

- No cross-service dependencies. All changes are in the `frontend` package.
- No new packages or dependencies required.
- shadcn/ui `Select` wraps Radix `@radix-ui/react-select`.

## Decisions

| Decision | Rationale |
|----------|-----------|
| Frontend-only fixes (no backend changes) | The 404 behavior for missing weeks is correct and documented in architecture.md §3.12. The frontend should handle it gracefully. |
| Explicit `SelectValue` children over restructuring the component | Minimal change; avoids rearchitecting the dialog or adding derived state management |
| Split `isEmpty` into `weekNotFound` only | The planner grid with 0 items is a valid and useful state (shows "+ Add exercise" per day). No need for a third "empty but provisioned" UI variant. |
