# Workout Tracker Bug Fixes — Implementation Plan

Last Updated: 2026-04-10

---

## Executive Summary

Three frontend-only UX bugs need fixing in the workout tracker. All have clear root causes, affect a single service (frontend), and require no backend changes. Total effort is Small — each fix is a few lines of code in well-understood files.

## Current State Analysis

| Bug | Symptom | Root Cause | File |
|-----|---------|------------|------|
| 1. Slow empty summary | Summary page takes several seconds to show "week is empty" | `useWorkoutWeekSummary` retries 404s 3× with backoff; sibling `useWorkoutWeek` already has `retry: false` | `frontend/src/lib/hooks/api/use-workouts.ts:317` |
| 2. UUID in dropdowns | Theme and Primary Region selects show UUIDs after selection | Radix `SelectValue` falls back to raw `value` when `SelectItem` children are not in the DOM during a re-render cycle | `frontend/src/pages/WorkoutExercisesPage.tsx:183-211` |
| 3. Start Fresh broken | Button creates week row but UI still shows empty-week prompt | `isEmpty` conflates 404 (no week) and 0-item week into the same branch | `frontend/src/pages/WorkoutWeekPage.tsx:59` |

## Proposed Changes

### Bug 1 — Add `retry: false` to `useWorkoutWeekSummary`

**File:** `frontend/src/lib/hooks/api/use-workouts.ts` (line 317)

Add `retry: false` to the `useQuery` options, matching the pattern already used by `useWorkoutWeek` at line 173.

Before:
```ts
return useQuery({
  queryKey: workoutKeys.summary(tenant, household, weekStart),
  queryFn: () => workoutService.getWeekSummary(tenant!, weekStart),
  enabled: !!tenant?.id && !!weekStart,
  staleTime: 30 * 1000,
});
```

After:
```ts
return useQuery({
  queryKey: workoutKeys.summary(tenant, household, weekStart),
  queryFn: () => workoutService.getWeekSummary(tenant!, weekStart),
  enabled: !!tenant?.id && !!weekStart,
  retry: false,
  staleTime: 30 * 1000,
});
```

### Bug 2 — Derive display text from data, not Radix auto-resolution

**File:** `frontend/src/pages/WorkoutExercisesPage.tsx` (lines 183-211)

Compute a `selectedThemeName` and `selectedRegionName` from the `themes`/`regions` props and the current `themeId`/`regionId` state. Pass these as explicit children to `SelectValue` so the trigger always shows a human name regardless of Radix DOM timing.

Before:
```tsx
<SelectValue placeholder="Select theme" />
```

After:
```tsx
<SelectValue placeholder="Select theme">
  {themes.find((t) => t.id === themeId)?.attributes.name ?? "Select theme"}
</SelectValue>
```

Same pattern for Primary Region.

### Bug 3 — Split `isEmpty` into `weekNotFound` vs `weekEmpty`

**File:** `frontend/src/pages/WorkoutWeekPage.tsx` (line 59)

Replace the single `isEmpty` flag with two:

Before:
```ts
const isEmpty = !!week.error || (week.data && items.length === 0);
```

After:
```ts
const weekNotFound = !!week.error;
```

Then update the JSX conditional (line 145):
- `weekNotFound` → render `<EmptyWeek>` (Copy Planned / Copy Actual / Start Fresh)
- Otherwise → render the 7-day planner grid (even with 0 items)

This means after Start Fresh succeeds, the week query refetch returns data (not 404), `weekNotFound` is `false`, and the planner grid renders with "+ Add exercise" buttons.

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Bug 2 fix: `SelectValue` children override may behave differently across Radix versions | Low | The Radix docs support explicit children on `SelectValue`; test visually after change |
| Bug 3 fix: showing the planner grid for an empty week may look odd without any items | Low | The grid already renders "+ Add exercise" per day — this is the intended UX after Start Fresh |
| Regression in copy-week flow | Low | Copy creates the week and adds items, so `weekNotFound` will be false and items will be populated — no change in behavior |

## Success Metrics

- Summary page for empty week loads in < 1 second (single request, no retries)
- Theme/Region selects always show human-readable names
- Start Fresh transitions to planner grid with day columns and add buttons
- Copy Planned / Copy Actual continue to work as before

## Timeline Estimate

**Total effort: S (Small)** — approximately 30 minutes of implementation + manual verification.
