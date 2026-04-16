# Dashboard Widget Skip/Partial Status — Context

Last Updated: 2026-04-16

---

## Key Files

### Files to Modify

| File | Purpose |
|------|---------|
| `frontend/src/components/features/trackers/habits-widget.tsx` | Habits dashboard widget — add skipped state rendering |
| `frontend/src/components/features/workouts/workout-widget.tsx` | Workout dashboard widget — add skipped and partial state rendering |

### Reference Files (read-only)

| File | Why |
|------|-----|
| `frontend/src/components/features/tracker/today-view.tsx` | Reference for how the full habits page renders skipped state (opacity-60 + "skipped" badge) |
| `frontend/src/pages/WorkoutTodayPage.tsx` | Reference for how the full workout page handles skip/partial status |
| `frontend/src/types/models/tracker.ts` | `TrackerEntryAttributes` type — confirms `skipped: boolean` field |
| `frontend/src/types/models/workout.ts` | `PerformanceStatus` type — confirms `"pending" \| "done" \| "skipped" \| "partial"` |
| `frontend/src/lib/hooks/api/use-trackers.ts` | `useTrackerToday()` hook — data shape reference |
| `frontend/src/lib/hooks/api/use-workouts.ts` | `useWorkoutToday()` hook — data shape reference |

### Related Task

| Task | Relevance |
|------|-----------|
| `docs/tasks/task-034-dashboard-widgets/prd.md` | Original PRD that created these widgets — confirms read-only, no-interaction design |

## Key Decisions

1. **Skipped = resolved, not pending.** Skipped items use strikethrough text (like done) to convey "no action needed." Distinguished from done by icon (SkipForward vs Check) and color (muted vs green).

2. **Partial = in progress, not resolved.** Partial items use normal text (like pending) since they still need attention. Distinguished from pending by icon (CircleDot vs Circle) and color (yellow vs muted).

3. **No backend changes.** Both APIs already return all necessary data — this is purely a frontend rendering change.

4. **Icon choices.** `SkipForward` for skipped (intuitive skip metaphor), `CircleDot` for partial (filled center conveys "started but not complete"). Neither icon is currently used in the project.

## Dependencies

- `lucide-react` — already installed as a project dependency; `SkipForward` and `CircleDot` are standard icons in the library
- No new npm packages required
- No backend service dependencies
