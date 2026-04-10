# Workout Tracker Bug Fixes — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-10
---

## 1. Overview

Three UX bugs in the workout tracker need to be resolved. All are frontend-side fixes with no backend changes required.

1. The weekly summary page takes several seconds to display "no workouts" for empty weeks because React Query retries the 404 response three times with exponential backoff before surfacing the error.
2. The "New exercise" dialog shows raw UUIDs in the Theme and Primary Region dropdowns after a selection is made, instead of the human-readable name.
3. The "Start Fresh" button creates the week row server-side but the UI continues to show the empty-week prompt instead of transitioning to the 7-day planner grid where exercises can be added.

These are small, self-contained fixes that can ship in a single PR.

## 2. Goals

Primary goals:
- Eliminate the multi-second delay when viewing the summary of an empty week
- Display theme/region names (not UUIDs) in the New Exercise dialog dropdowns
- Make "Start Fresh" transition the user to the planner grid so they can add exercises

Non-goals:
- Redesigning the summary endpoint to return 200 for missing weeks
- Adding new features or components beyond what's needed to fix these bugs
- Backend changes to workout-service

## 3. User Stories

- As a user, I want the summary page to immediately tell me a week is empty so I don't wait several seconds for an obvious answer.
- As a user, I want the Theme and Primary Region dropdowns to show names so I can confirm my selection at a glance.
- As a user, I want "Start Fresh" to show me the weekly planner with "Add exercise" buttons so I can begin building my week.

## 4. Functional Requirements

### 4.1 Summary Empty-Week Latency (Bug 1)

- The `useWorkoutWeekSummary` hook must not retry on 404 responses.
- When the summary endpoint returns 404, the summary page must immediately display the empty-week message ("No summary available — week is empty.").
- **Root cause**: `useWorkoutWeekSummary` in `frontend/src/lib/hooks/api/use-workouts.ts` (line 317) does not set `retry: false`. The sibling hook `useWorkoutWeek` (line 173) already does — this fix follows the same pattern.

### 4.2 UUID Display in New Exercise Dialog (Bug 2)

- After selecting a Theme or Primary Region in the New Exercise dialog, the `SelectTrigger` must display the selected item's name, never its UUID.
- The fix is in `ExerciseCreateDialog` within `frontend/src/pages/WorkoutExercisesPage.tsx` (lines 183–211).
- **Root cause**: Radix `SelectValue` falls back to rendering the raw `value` prop (a UUID) when it cannot resolve a matching `SelectItem` in the DOM — likely due to a render-cycle timing issue where the select items are momentarily absent during a data refetch. The fix should ensure the displayed text is always derived from the loaded theme/region data rather than relying solely on Radix's automatic value-to-label resolution.

### 4.3 Start Fresh Flow (Bug 3)

- After "Start Fresh" succeeds, the page must display the 7-day planner grid (with "+ Add exercise" buttons per day), not the empty-week prompt.
- The empty-week prompt (Copy Planned / Copy Actual / Start Fresh) must only appear when the week row does not exist (i.e., the week endpoint returns 404/error).
- When the week row exists but contains zero planned items, the planner grid must be shown.
- **Root cause**: In `frontend/src/pages/WorkoutWeekPage.tsx` (line 59), `isEmpty` conflates two states: `!!week.error` (no week row — 404) and `week.data && items.length === 0` (week exists, no items). The fix splits these into distinct conditions: the empty-week prompt shows only for the 404 case; the planner grid shows whenever the week row exists, even with zero items.

## 5. API Surface

No API changes. All fixes are frontend-only.

## 6. Data Model

No data model changes.

## 7. Service Impact

| Service | Change |
|---------|--------|
| **frontend** | Fix `useWorkoutWeekSummary` retry, fix `ExerciseCreateDialog` select display, fix `WorkoutWeekPage` empty-state logic |
| **workout-service** | None |

## 8. Non-Functional Requirements

- No new dependencies introduced.
- No changes to existing test coverage requirements beyond verifying the fixes.

## 9. Open Questions

None — all three bugs have clear root causes and straightforward fixes.

## 10. Acceptance Criteria

- [ ] Navigating to the summary page for an empty week shows the empty-state message within ~1 second (no multi-second retry delay).
- [ ] Selecting a Theme in the New Exercise dialog displays the theme name in the trigger, not a UUID.
- [ ] Selecting a Primary Region in the New Exercise dialog displays the region name in the trigger, not a UUID.
- [ ] Clicking "Start Fresh" on an empty week shows a toast confirmation and transitions to the 7-day planner grid with "+ Add exercise" buttons on each day.
- [ ] The empty-week prompt (Copy Planned / Copy Actual / Start Fresh) only appears when the week row does not yet exist (404 from the week endpoint).
- [ ] Existing copy-week and planner functionality is unaffected.
