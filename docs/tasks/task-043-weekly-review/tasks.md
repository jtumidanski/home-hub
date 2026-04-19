# Task 043: Task Checklist

Last Updated: 2026-04-19

---

## Phase 1 — Backend: Summary Projection Extensions

- [ ] **1.1** Add `week.GetSoonestNextWithItems(db, userID, after)` in `services/workout-service/internal/week/provider.go` mirroring `GetMostRecentPriorWithItems`; unit tests cover hit, miss, boundary (strictly greater than `after`), cross-user isolation [S]
- [ ] **1.2** `summary.Processor.Build` populates `previousPopulatedWeek` / `nextPopulatedWeek` by calling both helpers; processor tests cover the four combinations (prior-only, next-only, both, neither) [S]
- [ ] **1.3** `summary.buildActualSummary` emits a `setRows` array when `performance.mode == per_set`; preserve existing collapsed scalar fields; batch-fetch per-set rows for the whole week to avoid N+1; projection test asserts presence only in per-set case [M]

## Phase 2 — Backend: `/weeks/nearest` Endpoint

- [ ] **2.1** Add handler in `services/workout-service/internal/weekview/resource.go` (or adjacent file): parse `reference` (normalize to Monday-of-ISO-week) and `direction` (`prev`|`next`), dispatch to the appropriate provider helper, return JSON:API envelope with `type: "workoutWeekPointer"`; register route in `weekview/rest.go` [M]
- [ ] **2.2** Validation: `400` on missing/invalid `reference` or `direction`; `404` when no populated week exists in the requested direction; REST tests cover each error path [S]
- [ ] **2.3** Cross-user isolation integration test: user A never receives user B's populated weeks [S]

## Phase 3 — Frontend: Types, Hooks, Routing

- [ ] **3.1** Extend `WeekSummary` type in `frontend/src/types/models/workout.ts` with `previousPopulatedWeek`, `nextPopulatedWeek`, and `actualSummary.setRows?` [S]
- [ ] **3.2** Add `useWorkoutNearestPopulatedWeek(reference, direction, enabled?)` to `frontend/src/lib/hooks/api/use-workouts.ts`; React Query key includes reference + direction [S]
- [ ] **3.3** Rename `frontend/src/pages/WorkoutSummaryPage.tsx` → `WorkoutReviewPage.tsx`; add `/app/workouts/review[/:weekStart]` routes in `App.tsx`; convert `/app/workouts/summary[/:weekStart]` to `<Navigate replace>` redirects; update sidebar tab to `Review` in `components/features/workout/workout-shell.tsx` [S]

## Phase 4 — Frontend: Review Page UI

- [ ] **4.1** Four-stat totals card: Planned / Performed / Pending / Skipped (Pending computed client-side) [S]
- [ ] **4.2** Navigation header: `« Prev` / `Week of …` / `Next »` row + `↞ Previous populated` / `Next populated ↠` row; populated weeks read from summary response; empty weeks use two parallel `useWorkoutNearestPopulatedWeek` calls; buttons disabled when no target; URL push (not replace) so browser back works [M]
- [ ] **4.3** Per-day grid: seven blocks Mon → Sun from `byDay`; rest-day pill on `isRestDay`; item count or `Nothing scheduled`; `grid-cols-1 md:grid-cols-7` [M]
- [ ] **4.4** Per-item card: exercise name, status badge (`Done` / `Partial` / `Skipped` / `Pending`), kind-aware planned-vs-actual body (strength summary, strength per-set via `setRows`, isometric, cardio); `✓` when target met AND both sides non-null; skipped → strikethrough name + `Actual: Skipped`; pending → italic muted name + `Actual: —` [L]
- [ ] **4.5** Empty-week card: title has the week date, body reads `No workouts logged for this week.`; navigation header remains above and functional; parallel `nearest?prev` + `nearest?next` calls hydrate jump buttons [S]
- [ ] **4.6** Accessibility: `aria-label="Target met"` on `✓`; status badges carry text + color; day sections use `<section>` + `<h2>`; keyboard traversal works for every button [S]

## Phase 5 — Tests, Build, QA

- [ ] **5.1** Backend tests green: unit tests for `GetSoonestNextWithItems`, processor pointer fields, projection `setRows`, REST tests for `/weeks/nearest` (200, 404, 400, cross-user); `./scripts/test-all.sh` for workout-service passes [S]
- [ ] **5.2** Frontend tests green: component tests cover four kinds × four statuses; hook tests for `useWorkoutNearestPopulatedWeek`; navigation tests for URL push + browser back [M]
- [ ] **5.3** `./scripts/lint-all.sh`, `./scripts/build-all.sh`, `npm run build`; Docker builds succeed via `./scripts/local-up.sh` [S]
- [ ] **5.4** Manual QA on local stack: populated week (mixed statuses, per-set item renders each set), empty week (card + jump behavior), `/summary` redirect preserves `:weekStart`, sidebar tab reads `Review` [M]

---

## Acceptance Criteria (from PRD §10)

- [ ] `GET /api/v1/workouts/weeks/nearest?reference=…&direction=prev|next` returns `{weekStartDate}` or `404`; cross-user access rejected
- [ ] `week.GetSoonestNextWithItems` helper exists with unit tests mirroring the prior-direction helper
- [ ] Summary response includes `previousPopulatedWeek` and `nextPopulatedWeek` (string or null) on the `200` path
- [ ] Summary response includes per-set `setRows` array when performance mode is `per_set`; summary-mode performances do not
- [ ] `/app/workouts/review[/:weekStart]` renders the review page; `/app/workouts/summary*` paths redirect to the `review` equivalents
- [ ] Sidebar tab reads `Review` and navigates to `/app/workouts/review`
- [ ] Landing on `/app/workouts/review` (no week param) shows the current ISO week
- [ ] Header shows prev / next / jump-to-previous-populated / jump-to-next-populated, disabled appropriately
- [ ] Totals card shows four stats: Planned / Performed / Skipped / Pending
- [ ] Per-day section renders all seven days Mon → Sun with rest-day pill, item count, and ordered items
- [ ] Each item shows a status badge and a planned-vs-actual body appropriate to its kind
- [ ] Per-set mode performances render each set individually in the actual line
- [ ] Target-met `✓` is shown when actual ≥ planned on the primary metric and both are present
- [ ] Empty-week (404) state renders a friendly card with navigation controls still functional
- [ ] Prev / next / jump-to-populated change the URL path and push a history entry
- [ ] Frontend tests exercise the four kinds (strength summary, strength per-set, isometric, cardio) and the four statuses (done, partial, skipped, pending)
- [ ] Service integration tests exercise the summary response's new fields and the new nearest endpoint
