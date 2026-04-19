# Task 043: Key Context

Last Updated: 2026-04-19

---

## One-line summary

Turn the Workout `Summary` tab into a `Weekly Review` that renders the per-day planned-vs-actual breakdown (already computed server-side) and supports prev / next / jump-to-populated-week navigation.

---

## Key Files

### Backend — to modify

| File | Change |
|---|---|
| `services/workout-service/internal/week/provider.go` | Add `GetSoonestNextWithItems(db, userID, after)` mirroring `GetMostRecentPriorWithItems` at line 26. |
| `services/workout-service/internal/summary/processor.go` | `Processor.Build` populates `previousPopulatedWeek` and `nextPopulatedWeek`; `buildActualSummary` emits `setRows` array when `performance.mode == per_set` (currently collapses at lines ~487–505). |
| `services/workout-service/internal/summary/rest.go` | Response struct gains the two pointer fields and the nested `setRows` on actual summary. |
| `services/workout-service/internal/weekview/resource.go` | Add handler for `GET /workouts/weeks/nearest?reference=&direction=`. |
| `services/workout-service/internal/weekview/rest.go` | Register new route. |

### Backend — tests to add

| File | Purpose |
|---|---|
| `services/workout-service/internal/week/provider_test.go` (or adjacent) | Unit tests for `GetSoonestNextWithItems`: hit, miss, boundary, cross-user. |
| `services/workout-service/internal/summary/processor_test.go` | Processor tests covering the two pointer fields and per-set `setRows`. |
| `services/workout-service/internal/weekview/resource_test.go` | REST tests for `/weeks/nearest`: 200, 404, 400, cross-user. |

### Frontend — to modify / create

| File | Change |
|---|---|
| `frontend/src/pages/WorkoutSummaryPage.tsx` | Rename to `WorkoutReviewPage.tsx` and replace body with the Review implementation. |
| `frontend/src/App.tsx` | Add `/app/workouts/review[/:weekStart]` routes; make `/app/workouts/summary[/:weekStart]` `<Navigate replace>` redirects. |
| `frontend/src/components/features/workout/workout-shell.tsx` | Sidebar tab: `Summary` → `Review`; link to `/app/workouts/review`. |
| `frontend/src/lib/hooks/api/use-workouts.ts` | Extend `useWorkoutWeekSummary` response shape; add `useWorkoutNearestPopulatedWeek(reference, direction, enabled?)`. |
| `frontend/src/types/models/workout.ts` | Add `previousPopulatedWeek`, `nextPopulatedWeek`, and `actualSummary.setRows`. |
| `frontend/src/components/features/workout/review-day-block.tsx` (new) | Per-day grid block. |
| `frontend/src/components/features/workout/review-item-card.tsx` (new) | Per-item card with kind-aware planned-vs-actual body and status badge. |

---

## Key Decisions

- **Per-set JSON key rename**: the PRD example uses `sets` both for the scalar count and the per-set array. Cannot coexist in JSON. Implementation uses `setRows` for the array (api-contracts.md §1 naming note).
- **Pending stat**: computed client-side as `planned − performed − skipped`. No backend field added (PRD §9 open question: if the backend later exposes `totalPendingItems`, swap at that point).
- **Rename strategy**: `/summary` paths stay live as redirects. The old page filename is renamed (not kept as a shim) — redirects handle URL back-compat.
- **Nearest-endpoint home**: handler lives in `weekview`, not `week`, to avoid the import cycle called out at `services/workout-service/internal/week/resource.go:10`.
- **Target-met `✓`**: gated strictly on `planned != null && actual != null`. The predicate differs by kind (see plan.md Task 4.4). No `✓` on pending, skipped, or planned-value-missing states.
- **Nav header**: uses default `navigate()` (push) semantics so browser back works; URL is path-based, not query-string, so bookmarks are shareable.
- **Single PR**: no staged rollout needed. All new fields are additive; new endpoint is new; redirects protect existing bookmarks.

---

## Dependencies

### Depends on (existing)

- `workout.weeks`, `workout.planned_items`, `workout.performances`, `workout.performance_sets` tables — all in place from task-027.
- `(user_id, week_start_date)` composite index on `workout.weeks` — in place from task-027.
- `GetMostRecentPriorWithItems` at `services/workout-service/internal/week/provider.go:26` — reference pattern for the symmetric helper.
- `WorkoutWeekPage` header component — reference pattern for prev/next navigation.
- Existing JWT auth middleware on `/api/v1/workouts/*` — covers the new endpoint without new wiring.

### Does not depend on

- Any cross-service call — all queries are local to `workout-service`.
- nginx / k3s ingress changes — `/api/v1/workouts/*` already routes here.
- `docs/architecture.md` updates.
- Database migrations.

### Blocks / deferred

- Cross-week progression charts and week-over-week deltas are explicitly deferred (PRD §2, task-027 §2).
- PR detection, streaks, badges — deferred.
- Calendar/date picker UI, swipe gestures, CSV export — deferred.

---

## Open Questions (from PRD §9)

- **Tab label**: if a future dashboard tile wants `Summary`, we may reclaim the label. Not a blocker.
- **Mobile nav**: buttons only in v1. Swipe is a small follow-up if repeatedly requested.
- **Pending stat**: client-side derivation works; swap to backend field later if one is added.

---

## Non-obvious gotchas

- **`buildActualSummary` batching**: if per-set rows are fetched per performance (N+1), the summary p95 blows past the 100ms SLO for weeks with many items. Fetch all performance-sets for the week in one query keyed by performance IDs.
- **Rest days vs empty days**: the per-day grid renders seven columns unconditionally. Rest days get a pill; empty non-rest days get `Nothing scheduled`. Missing this visual distinction will confuse users.
- **`✓` on missing data**: do not render `✓` when either side of the comparison is null — it would look like the user hit a target they never set.
- **Strikethrough + status badge duplication**: skipped items need *both* strikethrough (visual) and a `Skipped` badge (screen-reader + color-blind). One without the other is a regression.
- **`WorkoutShell` tab state**: when changing the tab label/link, verify the active-tab highlight still works on both `/review` and the old `/summary` path during the redirect.
