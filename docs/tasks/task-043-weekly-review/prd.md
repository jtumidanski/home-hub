# Workout Weekly Review — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-19
---

## 1. Overview

The Workout section already has a Summary tab backed by `GET /api/v1/workouts/weeks/{weekStart}/summary`, but the frontend only renders three header counts (planned / performed / skipped) and per-theme / per-region totals. The per-day breakdown that the backend projection already computes — including each planned item's status and its planned-vs-actual values — is thrown away by the UI. There is also no way to navigate between weeks from the Summary screen: landing on an empty week gives a generic error with no prev/next controls.

This task turns that tab into a proper **Weekly Review**: surface the per-day breakdown (what was planned vs. what was actually done), add prev / next week navigation, and add a "jump to the previous / next populated week" affordance so the user can scroll through history without clicking past empty weeks one at a time.

The work is scoped to presentation plus a single new navigation helper endpoint. There is no new analytics, no new domain logic, no schema change. Cross-week progression charts remain deferred to a future task (see task-027 §2 non-goals).

## 2. Goals

Primary goals:
- Render the per-day breakdown (status, planned values, actual values) already returned by the summary endpoint, with planned-vs-actual rendered side-by-side
- Add prev / next week navigation to the Review page, matching the pattern used by `WorkoutWeekPage`
- Add "jump to previous populated week" and "jump to next populated week" controls that skip over empty weeks in one click
- Render a usable empty-week state on the Review page instead of the current generic error
- Rename the tab / route from "Summary" to "Review" to match how the user thinks about it

Non-goals (v1):
- Cross-week progression charts or time-series visualizations (deferred — see task-027 §2)
- Personal record (PR) detection, streak tracking, badges
- Editing plans or performances from the Review page (the Week page owns editing)
- A calendar / date picker UI — prev / next plus jump-to-populated is enough
- Mobile swipe gestures for week change — buttons are sufficient for v1
- Comparison to prior week (week-over-week delta) beyond what the existing per-theme / per-region block already shows
- Sharing, export, CSV download
- A separate "workouts done this year" view or multi-week aggregate
- Extending the summary projection itself — the shape it already returns is sufficient

## 3. User Stories

- As a user, I want to see, for each day of a given week, which exercises I planned and what I actually did, so that I can quickly review whether the week matched my plan.
- As a user, I want each planned item's status (done, partial, skipped, pending) to be visually obvious so I can scan a week at a glance.
- As a user, for strength items, I want planned sets/reps/weight rendered alongside the actual sets/reps/weight so I know whether I hit the target without doing math in my head.
- As a user, for cardio items, I want planned duration and distance rendered alongside the actual duration and distance.
- As a user, when a strength performance was logged per-set (varying reps / weight between sets), I want to see each individual set listed rather than a collapsed max.
- As a user, I want to click prev / next to move between adjacent weeks, the same way I can on the Week page.
- As a user, I want "jump to the last week that has data" and "jump to the next week that has data" controls so that scrolling through my history doesn't require clicking past weeks I never filled in.
- As a user, when I land on an empty week (past or future), I want a friendly empty state that still lets me navigate, rather than an error.

## 4. Functional Requirements

### 4.1 Route & Navigation

- Keep the existing routes (`/app/workouts/summary` and `/app/workouts/summary/:weekStart`) wired to the same page, rename the page component and file to `WorkoutReviewPage` / `WorkoutReviewPage.tsx`, and add new canonical routes `/app/workouts/review` and `/app/workouts/review/:weekStart`. The old `/summary` paths redirect to the corresponding `/review` path so existing bookmarks continue to work.
- The `WorkoutShell` sidebar tab `Summary` is renamed to `Review` and points at `/app/workouts/review`.
- When the user clicks the tab without a `:weekStart` parameter, the page lands on the current ISO week (Monday start, in the user's local time zone), matching the existing `currentWeekStart()` helper.
- Prev / Next buttons move by exactly 7 days and update the URL path (not query string) so each state is shareable and back-button friendly.

### 4.2 Week Header

- The top of the page shows a navigation row identical in spirit to `WorkoutWeekPage`'s: `« Prev` — `Week of YYYY-MM-DD` — `Next »`.
- A second row (or inline) shows two additional buttons:
  - `↞ Previous populated` — jumps to the most recent week that has planned items, strictly earlier than the currently displayed week. Disabled when none exists.
  - `Next populated ↠` — jumps to the soonest week that has planned items, strictly later than the currently displayed week. Disabled when none exists.
- Availability of the "populated" buttons comes from the summary endpoint response (see §5).
- When the target is the currently displayed week itself, the button is disabled.

### 4.3 Header Totals

- The existing three-up totals card (Planned / Performed / Skipped) is preserved.
- Add a fourth stat: `Pending` = `totalPlannedItems − totalPerformedItems − totalSkippedItems`. Computed client-side from the existing attributes; no backend change.
- The "Week of {date}" title is kept in the totals card as today.

### 4.4 Per-Day Breakdown

- Under the totals card, render a **Per Day** section with one block per day of the week, Mon → Sun, using the `byDay` array already returned by the summary projection.
- Each day block renders:
  - Day name (`Monday`, `Tuesday`, …)
  - A "Rest day" pill if `isRestDay` is true
  - The count of items on that day (e.g., `3 exercises`), or `No exercises` if empty
  - An ordered list of items (order = the order returned by the backend, which is `position` ascending per day)
- Days with zero items AND not marked as rest render a muted "Nothing scheduled" line — they must still appear so the week grid is visually consistent.
- Desktop: render the seven day blocks in a 7-column responsive grid (same breakpoint pattern as `WorkoutWeekPage`: 1 column on mobile, 7 on `md:` and up).
- Mobile: single column, days stacked top to bottom.

### 4.5 Per-Item Rendering

Each item inside a day block shows:

- Exercise name (bold)
- A status badge — one of:
  - `Done` (success color)
  - `Partial` (warning color)
  - `Skipped` (muted, strikethrough name)
  - `Pending` (muted, italicized)
- A two-line body showing planned vs actual, shape depending on `kind`:

  - **Strength (summary mode)**: `Planned: 3×10 @ 135 lb` / `Actual: 3×10 @ 140 lb ✓`
  - **Strength (per-set mode)**: `Planned: 3×10 @ 135 lb` / `Actual: set 1: 10 @ 135 · set 2: 8 @ 145 · set 3: 6 @ 155`
  - **Isometric**: `Planned: 3×60s` / `Actual: 3×55s`
  - **Cardio**: `Planned: 30:00 · 3.0 mi` / `Actual: 28:45 · 3.1 mi`

- A check (`✓`) is shown on the actual line when the actual value meets or exceeds the plan on the primary metric: strength → `actualWeight * actualReps * actualSets ≥ plannedWeight * plannedReps * plannedSets`; isometric → `actualSets * actualDuration ≥ plannedSets * plannedDuration`; cardio → `actualDistance ≥ plannedDistance` when both have distance, else `actualDuration ≥ plannedDuration`.
- When a planned value is missing (null), the check is not rendered (we can't know whether the target was met).
- When no actual exists (`Pending`), the actual line reads `Actual: —`.
- For `Skipped`, the actual line reads `Actual: Skipped` and the planned line is muted.
- Per-set rendering expands inline inside the same card — no click required. If the list is long (≥ 6 sets), it wraps to a second line rather than truncating.

### 4.6 Per-Set Data Availability

- The existing summary projection collapses per-set performances to a summary shape (`sets = count`, `reps = max reps`, `weight = max weight`) at §4.6 / processor.go:487–505.
- To render per-set detail per §4.5, the projection must additionally include the raw per-set rows when the performance mode is `per_set`. See §5.2 for the projection change.
- Non-strength kinds never have per-set data.

### 4.7 Empty-Week State

- When `GET /summary` returns `404` (no week row exists), render a card with:
  - The week date in the title
  - Copy: `No workouts logged for this week.`
  - The navigation header (prev / next / jump-to-populated) is still visible and functional above this card.
- The existing generic error text is removed.

### 4.8 Per-Theme and Per-Region Blocks

- The existing per-theme and per-region totals cards are preserved as-is. No visual change required in v1.

## 5. API Surface

Only one backend change is required; it is additive and backwards-compatible.

### 5.1 Extend `GET /api/v1/workouts/weeks/{weekStart}/summary`

Add two new top-level attributes to the response:

- `previousPopulatedWeek: "YYYY-MM-DD" | null` — the `week_start_date` of the most recent week strictly earlier than `weekStart` that has at least one `planned_items` row for this user; `null` if none.
- `nextPopulatedWeek: "YYYY-MM-DD" | null` — the `week_start_date` of the soonest week strictly later than `weekStart` that has at least one `planned_items` row for this user; `null` if none.

Implementation reuses the existing join pattern in `week.GetMostRecentPriorWithItems` (services/workout-service/internal/week/provider.go:26). A symmetric `GetSoonestNextWithItems` helper is added, and both are called from the summary processor. Both queries are indexed on `(user_id, week_start_date)` which already exists per task-027 data-model.

### 5.2 Include per-set rows in the summary projection

Extend each item in `byDay[].items[].actualSummary` so that when the underlying performance was in `per_set` mode, the projection additionally includes an array:

```
"sets": [
  { "setNumber": 1, "reps": 10, "weight": 135 },
  { "setNumber": 2, "reps": 8,  "weight": 145 }
]
```

The existing collapsed `sets`/`reps`/`weight` summary fields remain, so older clients (none yet, but architecturally) continue to work. The `weightUnit` continues to come from the performance, not per-set. This is added only when `mode == "per_set"`; in summary mode the `sets` array is omitted or `null`.

### 5.3 Summary endpoint behavior when week does not exist

Unchanged: still returns `404`. The page's empty-week renderer handles this. The two new `previous/nextPopulatedWeek` fields are only populated on the `200` path; on the `404` path, the frontend derives them by issuing a separate lookup — see §5.4.

### 5.4 Empty-week navigation lookup

When the page is on an empty week and needs to populate prev / next jump targets, the frontend calls a new lightweight endpoint:

- `GET /api/v1/workouts/weeks/nearest?reference=YYYY-MM-DD&direction=prev|next`
  - Returns `{ "weekStartDate": "YYYY-MM-DD" }` if a populated week exists in that direction relative to `reference`.
  - Returns `404` if none exists.
  - Authenticated as the JWT subject; never returns another user's data.
  - The `reference` value is normalized to the Monday of its ISO week server-side, same as existing week endpoints.

This endpoint is also usable from the populated-week flow, but the summary response in §5.1 is preferred there because it avoids a second round-trip.

### 5.5 Error Cases

- `400` — invalid `weekStart` or `reference` format; invalid `direction` value
- `404` — no populated week exists in the requested direction (nearest endpoint only); no week row (existing summary behavior)

## 6. Data Model

No schema changes.

Implementation uses existing tables:
- `workout.weeks`
- `workout.planned_items`
- `workout.performances`
- `workout.performance_sets`

Existing indexes on `(user_id, week_start_date)` on `workout.weeks` cover the new nearest-populated-week queries.

## 7. Service Impact

### 7.1 `workout-service`

- Add `week.GetSoonestNextWithItems(db, userID, after)` mirroring the existing `GetMostRecentPriorWithItems` helper.
- Extend `summary.Processor.Build` to populate `previousPopulatedWeek` and `nextPopulatedWeek` on its `RestModel` by calling both helpers.
- Extend `summary.buildActualSummary` to include the raw per-set rows when `p.Mode == performance.ModePerSet`; the existing collapsed fields stay.
- Add a new `weekview` route `GET /workouts/weeks/nearest` that reads `reference` and `direction` from query params and returns `{weekStartDate}`. This handler lives in `weekview` (not `week`) for the same import-cycle reason documented at services/workout-service/internal/week/resource.go:10.
- Tests: unit tests for the new helper; processor tests covering `previousPopulatedWeek` / `nextPopulatedWeek` values; projection tests covering the per-set array; REST tests for the nearest endpoint (200 and 404, cross-user rejection).

### 7.2 `frontend`

- Rename `frontend/src/pages/WorkoutSummaryPage.tsx` → `WorkoutReviewPage.tsx` and update `App.tsx` imports, route bindings, and the sidebar shell tab in `frontend/src/components/features/workout/workout-shell.tsx`.
- Add redirects: `/app/workouts/summary` → `/app/workouts/review`, `/app/workouts/summary/:weekStart` → `/app/workouts/review/:weekStart`.
- Implement the per-day section, per-item planned-vs-actual rendering (strength summary, strength per-set, isometric, cardio), status badge, and four-stat header (add `Pending`).
- Implement prev / next / jump-to-populated navigation using `previousPopulatedWeek` / `nextPopulatedWeek` from the summary response when on a populated week, or a call to `/weeks/nearest` when on an empty week.
- Extend `useWorkoutWeekSummary` hook typings in `frontend/src/lib/hooks/api/use-workouts.ts` (and the `WeekSummary` type in `frontend/src/types/models/workout.ts`) to include the new fields and per-set array.
- Add a hook `useWorkoutNearestPopulatedWeek(reference, direction)` that calls `/weeks/nearest`.
- Empty-week state and navigation preservation per §4.7.
- Keep accessibility in mind: status badges have both color and text label; buttons have proper labels.

### 7.3 No changes required

- nginx / k3s ingress (existing `/api/v1/workouts` route already covers the new endpoint path)
- `docs/architecture.md` (no new service, no routing change)
- auth / tenant plumbing (endpoint reuses existing middleware)
- Database migrations

## 8. Non-Functional Requirements

- **Multi-tenancy**: the new `/weeks/nearest` endpoint and the extended summary queries both filter by `tenant_id` and `user_id` from the JWT. Cross-user access is covered by an integration test.
- **Performance**: the nearest-populated-week queries are single-row lookups against an existing composite index. Budget: ≤ 50ms server-side p95 under typical load. The summary endpoint's p95 SLO from task-027 (100ms) is preserved.
- **Accessibility**: status badges include a text label, not color alone; keyboard navigation works for all buttons (prev/next/jump).
- **Responsive**: the 7-column day grid collapses to 1-column on mobile per the existing `md:` breakpoint pattern.
- **Browser back button**: changing weeks pushes a new history entry so browser back returns to the previous week.
- **Telemetry**: the frontend emits the same structured logs / errors it does today; no new observability work required.

## 9. Open Questions

- **Tab label**: the PRD specifies renaming `Summary` → `Review`. If there's a future dashboard tile that also shows weekly totals, we may want to reclaim "Summary" for that tile. Not a blocker.
- **Navigation on mobile**: v1 uses buttons. If swipe-between-weeks becomes a repeated request, that's a small follow-up.
- **Week-over-week deltas**: explicitly deferred. Re-open if the user finds the per-week review insufficient.
- **Pending stat definition**: defined client-side as `planned − performed − skipped`. If the backend ever decides to expose `totalPendingItems` directly, swap at that point.

## 10. Acceptance Criteria

- [ ] A new `GET /api/v1/workouts/weeks/nearest?reference=...&direction=prev|next` endpoint returns `{weekStartDate}` or `404`; cross-user access is rejected
- [ ] `week.GetSoonestNextWithItems` helper exists with unit tests mirroring the prior-direction helper
- [ ] Summary response includes `previousPopulatedWeek` and `nextPopulatedWeek` (string or null) on the `200` path
- [ ] Summary response includes per-set `sets` array when the underlying performance is in `per_set` mode; summary-mode performances do not include the array
- [ ] Route `/app/workouts/review` and `/app/workouts/review/:weekStart` render the review page; `/app/workouts/summary*` paths redirect to the `review` equivalents
- [ ] Sidebar tab reads "Review" and navigates to `/app/workouts/review`
- [ ] Landing on `/app/workouts/review` (no week param) shows the current ISO week
- [ ] Header shows prev / next / jump-to-previous-populated / jump-to-next-populated, disabled appropriately
- [ ] Totals card shows four stats: Planned / Performed / Skipped / Pending
- [ ] Per-day section renders all seven days Mon → Sun with rest-day pill, item count, and ordered items
- [ ] Each item shows a status badge and a planned-vs-actual body appropriate to its kind
- [ ] Per-set mode performances render each set individually in the actual line
- [ ] Target-met check (`✓`) is shown when actual ≥ planned on the primary metric and both are present
- [ ] Empty-week (404) state renders a friendly card with navigation controls still functional
- [ ] Prev / next / jump-to-populated change the URL path and push a history entry
- [ ] Frontend tests exercise the four kinds (strength summary, strength per-set, isometric, cardio) and the four statuses (done, partial, skipped, pending)
- [ ] Service integration tests exercise the summary response's new fields and the new nearest endpoint
