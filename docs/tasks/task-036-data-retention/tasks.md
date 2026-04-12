# Task-036 — Tasks Checklist

Last Updated: 2026-04-11

Effort key: S = ≤½ day · M = 1–2 days · L = 3–5 days · XL = 1+ week

---

## Phase 1 — Foundation (no behavior change)

- [ ] **P1.1** [L] Create `shared/go/retention` package
  - [ ] `loop.go` — jittered ticker (±10%) with graceful shutdown
  - [ ] `lock.go` — `pg_try_advisory_xact_lock` helper keyed on `hash(tenant_id, category)`
  - [ ] `policy_client.go` — 5-minute TTL cache + **mandatory cache-miss safety** (sentinel error, never 0 days)
  - [ ] `audit.go` — `retention_runs` row writer
  - [ ] `metrics.go` — Prometheus counters (`retention_scanned_total`, `retention_deleted_total`, `retention_run_duration_seconds`, `retention_run_failures_total`)
  - [ ] Unit tests for each file
- [ ] **P1.2** [M] account-service: `retention_policy_overrides` migration + compiled-in defaults
  - [ ] Migration matches `data-model.md` §6.1 exactly
  - [ ] `internal/retention/categories.go` with the PRD §4.2 default table
  - [ ] `Category` enum with `IsHouseholdScoped()` / `IsUserScoped()` / `Validate(days)` (365-day cap on `*_restore_window`, 3650-day cap otherwise)
  - [ ] Unit tests: unknown category, out-of-range, scope dispatch
- [ ] **P1.3** [M] account-service: `GET /api/v1/retention-policies`
  - [ ] Reads overrides for caller's household + user, merges with defaults
  - [ ] Response shape from `api-contracts.md` §1; per-category `source` annotation
  - [ ] Authorization: caller must belong to the household
  - [ ] Tests: empty overrides, single override, mixed scopes

## Phase 2 — Reapers (one service at a time)

Order: **productivity → recipe → tracker → workout → calendar.** Soak ~6h after each before moving on.

### productivity-service

- [ ] **P2.prod.1** [S] `retention_runs` migration
- [ ] **P2.prod.2** [S] Wire `retention.Loop` in `cmd/main.go`; configurable interval (default 6h ±10%)
- [ ] **P2.prod.3** [M] Category handlers + cascade
  - [ ] `productivity.completed_tasks`: hard-delete tasks where `completed_at < now() - window`
  - [ ] `productivity.deleted_tasks_restore_window`: hard-delete soft-deleted tasks past restore window
  - [ ] Cascade: `task → subtasks, reminders, task_restorations` in one transaction
  - [ ] Boundary tests; cascade integrity test; per-tenant failure isolation test
- [ ] **P2.prod.4** [S] Prometheus metrics exposed
- [ ] **P2.prod.5** [M] Concurrency integration test (two reapers, advisory lock)
- [ ] **P2.prod.6** [S] Deploy + ~6h soak; `retention_run_failures_total == 0`

### recipe-service

- [ ] **P2.rec.1** [S] `retention_runs` migration
- [ ] **P2.rec.2** [S] Wire `retention.Loop`
- [ ] **P2.rec.3** [M] Category handlers + cascade
  - [ ] `recipe.deleted_recipes_restore_window`: hard-delete soft-deleted recipes past restore window
  - [ ] `recipe.restoration_audit`: trim `recipe_restorations` past audit window
  - [ ] Cascade: `recipe → ingredients, instructions, recipe_restorations, meal-plan slot references` (preserve meal plan, clear slot only)
  - [ ] Boundary + cascade tests
- [ ] **P2.rec.4** [S] Prometheus metrics
- [ ] **P2.rec.5** [M] Concurrency integration test
- [ ] **P2.rec.6** [S] Deploy + soak

### tracker-service

- [ ] **P2.tra.1** [S] `retention_runs` migration
- [ ] **P2.tra.2** [S] Wire `retention.Loop`
- [ ] **P2.tra.3** [M] Category handlers + cascade
  - [ ] `tracker.entries`: reap `tracking_entries` by `entry_date` (no upward cascade)
  - [ ] `tracker.deleted_items_restore_window`: reap soft-deleted `tracking_items` past restore window
  - [ ] Cascade: `tracking_item → tracking_entries`
  - [ ] Boundary + cascade tests
- [ ] **P2.tra.4** [S] Prometheus metrics
- [ ] **P2.tra.5** [M] Concurrency integration test
- [ ] **P2.tra.6** [S] Deploy + soak

### workout-service

- [ ] **P2.wo.1** [S] `retention_runs` migration
- [ ] **P2.wo.2** [S] Wire `retention.Loop`
- [ ] **P2.wo.3** [L] Category handlers + cascade (largest cascade tree)
  - [ ] `workout.performances`: reap `performances` + `performance_sets` by `performed_at`
  - [ ] `workout.deleted_catalog_restore_window`: reap soft-deleted themes/regions/exercises past restore window
  - [ ] Cascade order: `theme → regions → exercises → performances → performance_sets`, one transaction per top-level parent
  - [ ] Boundary + multi-level cascade tests
- [ ] **P2.wo.4** [S] Prometheus metrics
- [ ] **P2.wo.5** [M] Concurrency integration test
- [ ] **P2.wo.6** [S] Deploy + soak

### calendar-service

- [ ] **P2.cal.1** [S] `retention_runs` migration
- [ ] **P2.cal.2** [S] Wire `retention.Loop`
- [ ] **P2.cal.3** [M] Category handler
  - [ ] `calendar.past_events`: reap `calendar_events` where `end_time < now() - window`. Leaf-level; no cascade.
  - [ ] Boundary tests; verify future events are untouched
- [ ] **P2.cal.4** [S] Prometheus metrics
- [ ] **P2.cal.5** [M] Concurrency integration test
- [ ] **P2.cal.6** [S] Deploy + soak

## Phase 3 — Write APIs and manual purge

- [ ] **P3.1** [M] account-service: `PATCH /api/v1/retention-policies/household/:household_id`
  - [ ] Sparse map; `null` deletes the override row
  - [ ] Household admin role check
  - [ ] Per-category bounds validation
  - [ ] Round-trip tests; 400 / 403 / 404 cases
- [ ] **P3.2** [S] account-service: `PATCH /api/v1/retention-policies/user`
  - [ ] Same shape, scoped to caller
  - [ ] Tests
- [ ] **P3.3** [M] `POST /internal/retention/purge` per service (parallelizable)
  - [ ] productivity-service
  - [ ] recipe-service
  - [ ] tracker-service
  - [ ] workout-service
  - [ ] calendar-service
  - [ ] Each: internal token auth, rate-limit (1 / (tenant, category) / 60s), `dry_run` support (transaction rollback), audit row written, returns `{run_id, scanned, deleted, dry_run, duration_ms}`, 409 on lock contention, 503 on unavailable policy
- [ ] **P3.4** [M] account-service: `POST /api/v1/retention-policies/purge` fan-out
  - [ ] Authorize, look up owning service, forward to internal endpoint
  - [ ] Returns 202 with correlation id
  - [ ] End-to-end test
- [ ] **P3.5** [M] account-service: `GET /api/v1/retention-runs` aggregated audit
  - [ ] Fan out to each reaper-owning service, aggregate, paginate
  - [ ] Filters: `category`, `trigger`; cursor pagination
  - [ ] Tenant-scoped

## Phase 4 — Package-service refactor (highest-risk phase)

- [ ] **P4.1** [S] Confirm `package.archive_window` and `package.archived_delete_window` defaults match current env-var values exactly
- [ ] **P4.2** [M] Add policy client + reaper alongside existing cleanup loop, **dry-run mode hardcoded**
- [ ] **P4.3** [S code, 1 week wall clock] Diff comparison: zero diff across at least 3 reaper cycles
- [ ] **P4.4** [M] Switch over
  - [ ] Delete `services/package-service/internal/poller/cleanup.go`
  - [ ] Remove env-var config knobs
  - [ ] Reaper now operates in normal (non-dry-run) mode
  - [ ] Spot check: `SELECT count(*) FROM packages WHERE archived_at < now() - interval '30 days'` returns 0

## Phase 5 — UI

- [ ] **P5.1** [L] "Data Retention" settings page
  - [ ] New page in household settings layout
  - [ ] "Household data" + "My personal data" sections
  - [ ] Per-row: label, current value, source badge, bounded number input, days unit
  - [ ] TanStack Query against `GET` + `PATCH` endpoints
  - [ ] Zod validation matching API bounds
- [ ] **P5.2** [M] Recent-purges panel powered by `GET /api/v1/retention-runs`
  - [ ] Shows last 20 entries grouped by category
  - [ ] Empty state
- [ ] **P5.3** [M] Per-category "Purge now" button + confirmation modal
  - [ ] Calls `POST /api/v1/retention-policies/purge`
  - [ ] Surfaces 429 rate-limit errors
- [ ] **P5.4** [M] Shrink-warning modal with dry-run preview
  - [ ] When user lowers a category below current effective value, send `dry_run: true` first
  - [ ] Show "this will permanently delete approximately N rows on the next reaper run"
  - [ ] Only send `PATCH` after explicit confirmation; cancel aborts

## Phase 6 — Documentation & Closeout

- [ ] **P6.1** [S] `docs/architecture.md` — new "Retention Framework" section
- [ ] **P6.2** [S] Walk PRD §10 acceptance-criteria checklist; mark each item with evidence (PR, test, dashboard)

## PRD §10 Acceptance Criteria — final verification

- [ ] `retention_policy_overrides` table exists in account-service per §6.1
- [ ] System defaults from §4.2 compiled into account-service and returned when no override
- [ ] `GET /api/v1/retention-policies` returns fully-resolved policy with per-category source
- [ ] `PATCH` endpoints accept partial updates and validate min/max bounds
- [ ] `POST /api/v1/retention-policies/purge` fans out and returns 202 with correlation id
- [ ] `GET /api/v1/retention-runs` returns paginated, tenant-scoped audit feed
- [ ] `shared/go/retention` package exists with all required helpers
- [ ] productivity, recipe, tracker, workout, calendar, package each run a reaper writing `retention_runs`
- [ ] Cascade rules in §4.5 implemented and tested per service
- [ ] Reapers skip tenants with unavailable policy and log a warning
- [ ] Reapers honor advisory locks (verified by integration test)
- [ ] Manual purge endpoints exist, are rate-limited, and emit audit rows with `trigger = 'manual'`
- [ ] Manual purge endpoints accept `dry_run` flag (transaction rollback path)
- [ ] Settings UI calls dry-run path before sending shrink `PATCH`
- [ ] Prometheus metrics exposed by each reaper-owning service
- [ ] `system.retention_audit` reaper trims old `retention_runs` rows
- [ ] UI "Data Retention" page exists with all required panels and affordances
- [ ] Package-service env-var config replaced by account-service policy lookup with no behavioral regression
- [ ] `docs/architecture.md` updated with retention framework section
