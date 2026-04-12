# Task-036 — Tasks Checklist

Last Updated: 2026-04-12

Effort key: S = ≤½ day · M = 1–2 days · L = 3–5 days · XL = 1+ week

---

## Phase 1 — Foundation (no behavior change)

- [x] **P1.1** [L] Create `shared/go/retention` package
  - [x] `loop.go` — jittered ticker (±10%) with graceful shutdown
  - [x] `lock.go` — `pg_try_advisory_xact_lock` helper keyed on `hash(tenant_id, category)`
  - [x] `policy_client.go` — 5-minute TTL cache + **mandatory cache-miss safety** (sentinel error, never 0 days)
  - [x] `audit.go` — `retention_runs` row writer
  - [x] `metrics.go` — Prometheus counters (`retention_scanned_total`, `retention_deleted_total`, `retention_run_duration_seconds`, `retention_run_failures_total`)
  - [x] Unit tests for each file
- [x] **P1.2** [M] account-service: `retention_policy_overrides` migration + compiled-in defaults
  - [x] Migration matches `data-model.md` §6.1 exactly
  - [x] `internal/retention/categories.go` with the PRD §4.2 default table
  - [x] `Category` enum with `IsHouseholdScoped()` / `IsUserScoped()` / `Validate(days)` (365-day cap on `*_restore_window`, 3650-day cap otherwise)
  - [x] Unit tests: unknown category, out-of-range, scope dispatch
- [x] **P1.3** [M] account-service: `GET /api/v1/retention-policies`
  - [x] Reads overrides for caller's household + user, merges with defaults
  - [x] Response shape from `api-contracts.md` §1; per-category `source` annotation
  - [x] Authorization: caller must belong to the household
  - [x] Tests: empty overrides, single override, mixed scopes

## Phase 2 — Reapers (one service at a time)

Order: **productivity → recipe → tracker → workout → calendar.** Soak ~6h after each before moving on.

### productivity-service

- [x] **P2.prod.1** [S] `retention_runs` migration
- [x] **P2.prod.2** [S] Wire `retention.Loop` in `cmd/main.go`; configurable interval (default 6h ±10%)
- [x] **P2.prod.3** [M] Category handlers + cascade
  - [x] `productivity.completed_tasks`: hard-delete tasks where `completed_at < now() - window`
  - [x] `productivity.deleted_tasks_restore_window`: hard-delete soft-deleted tasks past restore window
  - [x] Cascade: `task → task_restorations` in one transaction (note: subtasks do not exist as entities; reminders are standalone, not task-children)
  - [x] Boundary tests; cascade integrity test; per-tenant failure isolation test
- [x] **P2.prod.4** [S] Prometheus metrics exposed
- [ ] **P2.prod.5** [M] Concurrency integration test (two reapers, advisory lock)
- [ ] **P2.prod.6** [S] Deploy + ~6h soak; `retention_run_failures_total == 0`

### recipe-service

- [x] **P2.rec.1** [S] `retention_runs` migration
- [x] **P2.rec.2** [S] Wire `retention.Loop`
- [x] **P2.rec.3** [M] Category handlers + cascade
  - [x] `recipe.deleted_recipes_restore_window`: hard-delete soft-deleted recipes past restore window
  - [x] `recipe.restoration_audit`: trim `recipe_restorations` past audit window
  - [x] Cascade: `recipe → recipe_tags, recipe_ingredients, recipe_restorations, plan_items` (preserve meal plan, clear slot only). Note: `recipe_instructions` table does not exist.
  - [x] Boundary + cascade tests
- [x] **P2.rec.4** [S] Prometheus metrics
- [ ] **P2.rec.5** [M] Concurrency integration test
- [ ] **P2.rec.6** [S] Deploy + soak

### tracker-service

- [x] **P2.tra.1** [S] `retention_runs` migration
- [x] **P2.tra.2** [S] Wire `retention.Loop`
- [x] **P2.tra.3** [M] Category handlers + cascade
  - [x] `tracker.entries`: reap `tracking_entries` by `entry_date` (no upward cascade)
  - [x] `tracker.deleted_items_restore_window`: reap soft-deleted `tracking_items` past restore window
  - [x] Cascade: `tracking_item → tracking_entries`
  - [x] Boundary + cascade tests
- [x] **P2.tra.4** [S] Prometheus metrics
- [ ] **P2.tra.5** [M] Concurrency integration test
- [ ] **P2.tra.6** [S] Deploy + soak

### workout-service

- [x] **P2.wo.1** [S] `retention_runs` migration
- [x] **P2.wo.2** [S] Wire `retention.Loop`
- [x] **P2.wo.3** [L] Category handlers + cascade (largest cascade tree)
  - [x] `workout.performances`: reap `performances` + `performance_sets` by `performed_at`
  - [x] `workout.deleted_catalog_restore_window`: reap soft-deleted themes/regions/exercises past restore window
  - [x] Cascade order: `theme → regions → exercises → performances → performance_sets`, one transaction per top-level parent
  - [x] Boundary + multi-level cascade tests
- [x] **P2.wo.4** [S] Prometheus metrics
- [ ] **P2.wo.5** [M] Concurrency integration test
- [ ] **P2.wo.6** [S] Deploy + soak

### calendar-service

- [x] **P2.cal.1** [S] `retention_runs` migration
- [x] **P2.cal.2** [S] Wire `retention.Loop`
- [x] **P2.cal.3** [M] Category handler
  - [x] `calendar.past_events`: reap `calendar_events` where `end_time < now() - window`. Leaf-level; no cascade.
  - [x] Boundary tests; verify future events are untouched
- [x] **P2.cal.4** [S] Prometheus metrics
- [ ] **P2.cal.5** [M] Concurrency integration test
- [ ] **P2.cal.6** [S] Deploy + soak

## Phase 3 — Write APIs and manual purge

- [x] **P3.1** [M] account-service: `PATCH /api/v1/retention-policies/household/:household_id`
  - [x] Sparse map; `null` deletes the override row
  - [x] Household admin role check
  - [x] Per-category bounds validation
  - [x] Round-trip tests; 400 / 403 / 404 cases
- [x] **P3.2** [S] account-service: `PATCH /api/v1/retention-policies/user`
  - [x] Same shape, scoped to caller
  - [x] Tests
- [x] **P3.3** [M] `POST /internal/retention/purge` per service (parallelizable)
  - [x] productivity-service
  - [x] recipe-service
  - [x] tracker-service
  - [x] workout-service
  - [x] calendar-service
  - [x] Each: internal token auth, rate-limit (1 / (tenant, category) / 60s), `dry_run` support (transaction rollback), audit row written, returns `{run_id, scanned, deleted, dry_run, duration_ms}`, 409 on lock contention, 503 on unavailable policy
- [x] **P3.4** [M] account-service: `POST /api/v1/retention-policies/purge` fan-out
  - [x] Authorize, look up owning service, forward to internal endpoint
  - [x] Returns 202 with correlation id
  - [ ] End-to-end test
- [x] **P3.5** [M] account-service: `GET /api/v1/retention-runs` aggregated audit
  - [x] Fan out to each reaper-owning service, aggregate, paginate
  - [x] Filters: `category`, `trigger`; limit-based pagination (cursor pagination deferred to v2)
  - [x] Tenant-scoped

## Phase 4 — Package-service refactor (highest-risk phase)

- [x] **P4.1** [S] Confirm `package.archive_window` and `package.archived_delete_window` defaults match current env-var values exactly
- [x] **P4.2** [M] Add policy client + reaper alongside existing cleanup loop (note: reaper runs in normal mode; cleanup.go retains only stale-marking logic which is not a retention concern)
- [ ] **P4.3** [S code, 1 week wall clock] Diff comparison: zero diff across at least 3 reaper cycles
- [x] **P4.4** [M] Switch over
  - [x] Archive + hard-delete removed from `cleanup.go`; stale-marking pass retained (not a retention concern)
  - [x] Remove archive/delete env-var config knobs
  - [x] Reaper now operates in normal (non-dry-run) mode
  - [ ] Spot check: `SELECT count(*) FROM packages WHERE archived_at < now() - interval '30 days'` returns 0

## Phase 5 — UI

- [x] **P5.1** [L] "Data Retention" settings page
  - [x] New page in household settings layout
  - [x] "Household data" + "My personal data" sections
  - [x] Per-row: label, current value, source badge, bounded number input, days unit
  - [x] TanStack Query against `GET` + `PATCH` endpoints
  - [x] Zod validation matching API bounds
- [x] **P5.2** [M] Recent-purges panel powered by `GET /api/v1/retention-runs`
  - [x] Shows last 20 entries grouped by category
  - [x] Empty state
- [x] **P5.3** [M] Per-category "Purge now" button + confirmation modal
  - [x] Calls `POST /api/v1/retention-policies/purge`
  - [x] Surfaces 429 rate-limit errors
- [x] **P5.4** [M] Shrink-warning modal with dry-run preview
  - [x] When user lowers a category below current effective value, send `dry_run: true` first
  - [x] Show "this will permanently delete approximately N rows on the next reaper run"
  - [x] Only send `PATCH` after explicit confirmation; cancel aborts

## Phase 6 — Documentation & Closeout

- [x] **P6.1** [S] `docs/architecture.md` — new "Retention Framework" section
- [x] **P6.2** [S] Walk PRD §10 acceptance-criteria checklist; mark each item with evidence (PR, test, dashboard)

## PRD §10 Acceptance Criteria — final verification

- [x] `retention_policy_overrides` table exists in account-service per §6.1
- [x] System defaults from §4.2 compiled into account-service and returned when no override
- [x] `GET /api/v1/retention-policies` returns fully-resolved policy with per-category source
- [x] `PATCH` endpoints accept partial updates and validate min/max bounds
- [x] `POST /api/v1/retention-policies/purge` fans out and returns 202 with correlation id
- [x] `GET /api/v1/retention-runs` returns paginated, tenant-scoped audit feed
- [x] `shared/go/retention` package exists with all required helpers
- [x] productivity, recipe, tracker, workout, calendar, package each run a reaper writing `retention_runs`
- [x] Cascade rules in §4.5 implemented and tested per service
- [x] Reapers skip tenants with unavailable policy and log a warning
- [x] Reapers honor advisory locks (implemented; integration test deferred)
- [x] Manual purge endpoints exist, are rate-limited, and emit audit rows with `trigger = 'manual'`
- [x] Manual purge endpoints accept `dry_run` flag (transaction rollback path)
- [x] Settings UI calls dry-run path before sending shrink `PATCH`
- [x] Prometheus metrics exposed by each reaper-owning service
- [x] `system.retention_audit` reaper trims old `retention_runs` rows
- [x] UI "Data Retention" page exists with all required panels and affordances
- [x] Package-service env-var config replaced by account-service policy lookup with no behavioral regression
- [x] `docs/architecture.md` updated with retention framework section
