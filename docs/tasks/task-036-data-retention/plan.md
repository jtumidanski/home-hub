# Task-036 — Data Retention & Purging — Implementation Plan

Last Updated: 2026-04-11
Status: Draft (planning phase — no implementation until approval)

---

## 1. Executive Summary

Home Hub services purge data inconsistently today: package-service has a real cleanup loop, calendar-service expires OAuth state, weather-service refreshes a cache, and **everything else either soft-deletes forever or never deletes at all**. Soft-deleted rows accumulate, time-series tables (tracker entries, workout performances, past calendar events) grow without bound, and households have no control over how long their data is kept.

This task delivers a unified retention framework:

- **account-service** becomes the source of truth for retention policies (system defaults + per-household / per-user overrides), exposing JSON:API endpoints.
- **shared/go/retention** provides ticker, advisory-lock, policy-client (with TTL cache + cache-miss safety), and audit-row writer building blocks.
- **productivity, recipe, tracker, workout, calendar, package** each get a reaper that consults policy and writes a per-run audit row, with explicit application-level cascade rules.
- A **manual purge** path (with dry-run preview) and a **Data Retention** UI surface let households see and act on what is happening.
- **package-service is refactored last**, behind a dry-run gate, to preserve current production behavior.

The rollout is staged so that no phase deletes data based on missing policy and so the package-service refactor cannot regress existing behavior.

## 2. Current State Analysis

### 2.1 What exists today

- **package-service** (`services/package-service/internal/poller/cleanup.go`) — three-stage cleanup loop (stale → archive → hard-delete) configured by environment variables. The only mature reaper in the project.
- **calendar-service** — expires OAuth state tokens; not user-meaningful retention.
- **weather-service** — cache TTL refresh; not user-meaningful retention.
- **productivity, recipe, tracker, workout, account, shopping, category** — soft-delete (`gorm.DeletedAt`) forever or never delete. No reaper. No upper bound on table growth.
- **shared/go/** — has `auth`, `database`, `http`, `logging`, `model`, `server`, `tenant`, `testing`. **No `retention` package.**
- **account-service** — owns households, users, tenants. Does not yet expose any retention surface.

### 2.2 Pain points the framework must solve

1. Time-series tables (`tracker_entries`, `performances`, `calendar_events`) grow unbounded.
2. Soft-deletes never become hard-deletes; "restore window" is implicit and unbounded.
3. Orphan risk: cascading children (e.g., `performance_sets` under `performances`) are not removed when their parent is reaped.
4. No audit trail — operators cannot see what was purged, households cannot trust the system.
5. Package-service config lives in env vars, drifts from other services, and is invisible to households.

### 2.3 Constraints inherited from the codebase

- Go services with GORM, JSON:API transport, multi-tenant via `shared/go/tenant`.
- Postgres only — advisory locks (`pg_try_advisory_xact_lock`) are available.
- DDD layout with immutable models, functional composition, per-service `cmd/main.go`.
- Frontend is React + Vite + shadcn/ui + TanStack Query + react-hook-form + Zod.

## 3. Proposed Future State

- A `retention_policy_overrides` table in account-service plus a compiled-in `map[Category]int` of system defaults from PRD §4.2.
- Public JSON:API on account-service:
  - `GET  /api/v1/retention-policies`
  - `PATCH /api/v1/retention-policies/household/:household_id`
  - `PATCH /api/v1/retention-policies/user`
  - `POST /api/v1/retention-policies/purge` (with `dry_run`)
  - `GET  /api/v1/retention-runs` (cross-service audit fan-out)
- Internal endpoint on each reaper-owning service: `POST /internal/retention/purge` (with `dry_run`).
- A `shared/go/retention` package providing:
  - jittered ticker loop with graceful shutdown
  - `pg_try_advisory_xact_lock` helper keyed on hash of `(tenant_id, category)`
  - policy client with 5-minute TTL cache and **mandatory cache-miss safety** (skip the run, never default to "0 days")
  - audit-row writer for per-service `retention_runs`
- Each of productivity, recipe, tracker, workout, calendar, package services:
  - new `retention_runs` table
  - reaper goroutine wired into `cmd/main.go`
  - per-category handlers with explicit cascade rules per PRD §4.5, all inside a single transaction per parent
  - Prometheus metrics (`retention_scanned_total`, `retention_deleted_total`, `retention_run_duration_seconds`, `retention_run_failures_total`)
- A new "Data Retention" page in the household settings UI: per-category editor with source badges, recent-purges panel, per-category "Purge now" with dry-run preview before any window-shrinking edit.
- `docs/architecture.md` updated with a Retention Framework section.

## 4. Implementation Phases

The phase ordering is non-negotiable. **Phase 4 (package-service refactor) must come after Phases 1–3** so the dry-run gate can run alongside the existing cleanup loop.

### Phase 1 — Foundation (no behavior change)

Goal: ship the building blocks and the read-only policy API. Nothing new is reaped.

1. `shared/go/retention` package (ticker, advisory lock, policy client, audit writer).
2. account-service: `retention_policy_overrides` migration + compiled-in defaults table.
3. account-service: `GET /api/v1/retention-policies` + authorization helpers.

Exit criteria: package builds for all services, account-service returns a fully-resolved policy with `source: default` for every category, no other service is touched.

### Phase 2 — Reapers (one service at a time)

For each of **productivity → recipe → tracker → workout → calendar** (in that order):

1. `retention_runs` table migration in the service.
2. Wire the `shared/go/retention` reaper loop in `cmd/main.go`.
3. Implement category handlers + cascade per PRD §4.5.
4. Prometheus metrics.
5. Unit tests per category (boundary values, cascade integrity).
6. Integration test: two reapers in parallel; advisory lock prevents double-processing.
7. Deploy. Watch metrics for one full reaper cycle (~6h) before proceeding to the next service.

Order rationale: productivity first (most active soft-delete usage, smallest blast radius). Calendar last (largest single-table volume, most dependent on policy correctness).

### Phase 3 — Write APIs and manual purge

1. account-service `PATCH` endpoints for household and user scopes.
2. account-service `POST /api/v1/retention-policies/purge` with fan-out.
3. `POST /internal/retention/purge` on each reaper-owning service (with `dry_run`, rate-limit).
4. account-service `GET /api/v1/retention-runs` aggregated audit feed.

### Phase 4 — Package-service refactor (highest-risk phase)

1. Confirm `package.archive_window` and `package.archived_delete_window` defaults match current env-var values **exactly**.
2. Add the policy client + reaper loop alongside the existing cleanup loop, in **dry-run mode** (logs only).
3. Run for one full week. Compare what the new reaper would delete vs what the old loop actually deleted.
4. Once the diff is zero across at least 3 reaper cycles, switch over: remove `internal/poller/cleanup.go` and rely on the shared reaper.
5. Remove env-var config knobs.

### Phase 5 — UI

1. New "Data Retention" settings page using the existing settings layout.
2. Recent-purges panel powered by `GET /api/v1/retention-runs`.
3. Per-category "Purge now" with confirmation modal.
4. Shrink-warning modal: when a user lowers a window below the current effective value, the UI calls the dry-run path for a row-count preview before sending the `PATCH`.

### Phase 6 — Documentation & Closeout

1. `docs/architecture.md` Retention Framework section.
2. Mark task-036 acceptance criteria checked off in `tasks.md`.
3. Note the forward dependency for the future "delete my account" task (PRD §9a).

## 5. Detailed Tasks

Tasks are numbered by phase. Effort: S = ≤½ day, M = 1–2 days, L = 3–5 days, XL = 1+ week.

### Phase 1 — Foundation

**P1.1 — `shared/go/retention` package** [L]
- Create `shared/go/retention/` with: `loop.go` (jittered ticker + graceful shutdown), `lock.go` (`pg_try_advisory_xact_lock` helper), `policy_client.go` (TTL cache + cache-miss safety), `audit.go` (`retention_runs` writer), `metrics.go` (Prometheus).
- Acceptance:
  - Loop respects context cancellation; jitter is ±10%.
  - Lock helper takes `(tenantID, category)` and returns acquired/false; never blocks.
  - Policy client returns cached value past TTL on backend failure; returns sentinel error if no cache exists; **never** returns 0 days as a fallback.
  - Audit writer takes a struct and writes one row.
  - Unit tests for each.
- Depends on: none.

**P1.2 — account-service `retention_policy_overrides` migration + defaults** [M]
- Add migration matching `data-model.md` §6.1 exactly.
- Define `internal/retention/categories.go` with the system defaults table from PRD §4.2.
- Add a `Category` enum type with `IsHouseholdScoped()` / `IsUserScoped()` / `Validate(days int)` (enforces the 365-day cap on `*_restore_window` categories and 3650-day cap on others).
- Acceptance:
  - Migration applies cleanly on a fresh DB and on a populated dev DB.
  - Defaults table compiles into the binary; no DB rows required to return defaults.
  - Unknown category → `ErrUnknownCategory`; out-of-range days → `ErrOutOfRange`.
- Depends on: none (can run in parallel with P1.1).

**P1.3 — account-service `GET /api/v1/retention-policies`** [M]
- Read overrides for the caller's household and user, merge with defaults, return the response shape from `api-contracts.md` §1.
- Wire through existing JSON:API conventions. Source = `default` | `household` | `user`.
- Acceptance:
  - With no overrides, every category reports `source: default` and the value from the table.
  - With one override, that category reports the new value and the correct source.
  - Authorization: caller must be a member of the household; user scope is implied to be self.
  - Tests: empty, one override, mixed.
- Depends on: P1.2.

### Phase 2 — Reapers (per service)

The same task template repeats for each of productivity, recipe, tracker, workout, calendar. Listed once with placeholders; multiply by five.

**P2.{svc}.1 — `retention_runs` table migration** [S]
- Migration matching `data-model.md` §6.2.
- Acceptance: clean apply on fresh + populated DB.

**P2.{svc}.2 — Wire reaper loop in `cmd/main.go`** [S]
- Construct `retention.Loop` with the service's category handlers, the policy client (pointed at account-service), and the audit writer.
- Configurable interval (env var, default 6h, ±10% jitter).
- Honor `ctx` for graceful shutdown.
- Acceptance: service starts; first tick logs structured "skipped: no tenants" if DB is empty; metrics endpoint exposes the four counters.

**P2.{svc}.3 — Category handlers + cascade** [M–L per service]
- Implement one handler per category owned by this service (see PRD §4.5 for the per-service cascade rules).
- Each handler:
  - Iterates tenants explicitly (no global query).
  - Acquires the advisory lock per `(tenant, category)`; skips if unavailable.
  - Runs scoped query → batches of 500 rows → cascade in one transaction per parent.
  - Writes one `retention_runs` row per (tenant, category) per tick.
  - On per-tenant failure: logs, records the error in the audit row, continues to next tenant.
- **Cascade-by-service** (from PRD §4.5):
  - **productivity**: `task → subtasks, reminders, task_restorations`.
  - **recipe**: `recipe → ingredients, instructions, recipe_restorations, meal-plan slot references` (preserve the meal plan, clear the slot only).
  - **tracker**: `tracking_item → tracking_entries`. Reaping `tracking_entries` does NOT cascade upward.
  - **workout**: `theme → regions → exercises → performances → performance_sets`. One transaction per top-level parent.
  - **calendar**: `calendar_event` is leaf-level.
- Acceptance:
  - Boundary tests: row exactly at the cutoff is not deleted; one second past is.
  - Cascade integrity test: deleting a parent removes all dependents in the same transaction; an error mid-cascade rolls back the entire parent.
  - Reaper continues after a single-tenant failure.
- Depends on: P1.1, P1.2, P1.3, P2.{svc}.1, P2.{svc}.2.

**P2.{svc}.4 — Prometheus metrics** [S]
- Expose `retention_scanned_total`, `retention_deleted_total`, `retention_run_duration_seconds`, `retention_run_failures_total`, labeled by `service`, `category`.
- Acceptance: `/metrics` shows the counters; integration test asserts they increment.

**P2.{svc}.5 — Concurrency integration test** [M]
- Run two reaper instances in parallel against the same DB; assert that for each `(tenant, category)` only one instance writes a `retention_runs` row per tick.
- Acceptance: test passes deterministically.

**P2.{svc}.6 — Deploy + soak** [S]
- Deploy to staging. Watch one full ~6h cycle. `retention_run_failures_total == 0`. Spot-check audit counts vs. row deltas.
- Acceptance: clean cycle observed before next service is started.

### Phase 3 — Write APIs and manual purge

**P3.1 — `PATCH /api/v1/retention-policies/household/:household_id`** [M]
- Sparse map of `category → days | null`. `null` deletes the override row.
- Auth: household admin role check.
- Validation: per-category bounds.
- Acceptance: PATCH then GET round-trips; bounds errors return 400; non-admin returns 403.
- Depends on: P1.3.

**P3.2 — `PATCH /api/v1/retention-policies/user`** [S]
- Same shape, scoped to caller. No admin check.
- Acceptance: PATCH then GET round-trips.
- Depends on: P3.1 (shared validation logic).

**P3.3 — `POST /internal/retention/purge` per service** [M, parallelizable across services]
- Internal endpoint on each reaper-owning service. Body: `{ tenant_id, scope_kind, scope_id, category, dry_run }`.
- Authorized via internal service token.
- Rate-limited: one per `(tenant, category)` per 60s.
- Dry-run path runs the full scan + cascade walk inside a transaction that is rolled back. Audit row is written with `dry_run = true`.
- Acceptance: returns `{run_id, scanned, deleted, dry_run, duration_ms}`; 409 if lock is held; 503 if policy unavailable; rate limit returns 429.
- Depends on: P2.{svc}.3.

**P3.4 — `POST /api/v1/retention-policies/purge` (account-service fan-out)** [M]
- Authorize the request, look up the owning service, forward to its internal endpoint.
- Returns 202 with a correlation id.
- Acceptance: end-to-end from public endpoint → owning service → audit row written.
- Depends on: P3.3 (at least one service).

**P3.5 — `GET /api/v1/retention-runs` aggregated audit** [M]
- Fan out to each reaper-owning service, query their `retention_runs`, aggregate, paginate.
- Tenant-scoped. Filter by `category` / `trigger`. Cursor pagination.
- Acceptance: returns runs from at least two services in one response; pagination cursor round-trips.
- Depends on: P2 complete for at least two services; ideally all five.

### Phase 4 — Package-service refactor (highest-risk phase)

**P4.1 — Confirm default values match current env-vars** [S]
- Read current production env-var values. Confirm `package.archive_window` and `package.archived_delete_window` defaults in account-service match exactly.
- Acceptance: documented match; mismatch is a blocker.

**P4.2 — Add reaper alongside existing cleanup loop, dry-run mode** [M]
- Wire `shared/go/retention` reaper in `cmd/main.go` with `dry_run = true` hardcoded.
- Both loops run; new reaper logs only.
- Acceptance: both loops observable in logs; no behavior change.

**P4.3 — One-week diff comparison** [S in code, 1 week wall clock]
- Compare new reaper's would-delete counts vs old loop's actual deletes.
- Acceptance: zero diff across at least 3 reaper cycles.

**P4.4 — Switch over** [M]
- Remove `services/package-service/internal/poller/cleanup.go`.
- Remove env-var config (`PACKAGE_ARCHIVE_DAYS`, `PACKAGE_DELETE_DAYS`, etc.).
- Reaper now operates in normal (non-dry-run) mode.
- Acceptance: spot check `SELECT count(*) FROM packages WHERE archived_at < now() - interval '30 days'` is 0.
- Depends on: P4.3.

### Phase 5 — UI

**P5.1 — "Data Retention" settings page** [L]
- New page in household settings layout.
- Two sections: "Household data" and "My personal data".
- Each row: category label, current value, source badge, number input bounded by min/max, days unit.
- Uses TanStack Query against `GET /api/v1/retention-policies` and `PATCH` endpoints.
- Validation via Zod against the same min/max as the API.
- Acceptance: edits persist; source badges update; bounds errors surfaced.
- Depends on: P1.3, P3.1, P3.2.

**P5.2 — Recent-purges panel** [M]
- Panel showing the last 20 entries from `GET /api/v1/retention-runs`, grouped by category.
- Acceptance: groups update on refresh; empty state when no runs.
- Depends on: P3.5.

**P5.3 — "Purge now" affordance with confirmation modal** [M]
- Per-category button. Confirmation modal before sending `POST /api/v1/retention-policies/purge`.
- Acceptance: success toast on 202; rate-limit (429) surfaced.
- Depends on: P3.4.

**P5.4 — Shrink-warning modal with dry-run preview** [M]
- When the user lowers a category below its current effective value, send `POST /api/v1/retention-policies/purge` with `dry_run: true` first; show "this will permanently delete approximately N rows on the next reaper run" in a confirmation modal; only send the `PATCH` after explicit confirmation.
- Acceptance: shrinking triggers the preview path; non-shrinking edits do not; cancel aborts the `PATCH`.
- Depends on: P5.1, P3.4.

### Phase 6 — Documentation & Closeout

**P6.1 — `docs/architecture.md` Retention Framework section** [S]
- New section describing: policy storage, distribution, reaper loop, cascade rules, audit, manual purge, dry-run.
- Acceptance: section reviewed and merged.

**P6.2 — Acceptance-criteria checklist closeout** [S]
- Walk PRD §10 checklist; mark each item; link evidence (PR, test, dashboard).

## 6. Risk Assessment and Mitigation

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|------------|--------|-----------|
| R1 | Reaper deletes data based on missing/stale policy | Low | Critical (data loss) | `shared/go/retention` policy client **must** return a sentinel error on cache-miss; reaper **must** skip the run. Enforced by unit test in P1.1. |
| R2 | Package-service refactor regresses production behavior | Medium | High | Phase 4 dry-run gate: one week of zero-diff before switchover. P4.1 explicitly verifies default-value parity. |
| R3 | Cascade-delete leaves orphans on partial failure | Low | High | One transaction per top-level parent; mid-cascade error rolls back the parent. Verified by unit test in P2.{svc}.3. |
| R4 | Two replicas reap the same `(tenant, category)` simultaneously | Medium | Medium (double work, audit duplication) | Postgres advisory lock keyed on `hash(tenant, category)`; second replica skips. Verified by P2.{svc}.5 integration test. |
| R5 | Reaper holds DB locks too long and starves user requests | Medium | Medium | Batch size 500, one transaction per batch, yield between batches. p95 < 60s per (tenant, category) per run. |
| R6 | A user shrinks a window and is surprised by mass deletion | Medium | Medium | P5.4 dry-run preview modal — explicit confirmation with row count. |
| R7 | account-service becomes a single point of failure for all reapers | Medium | Medium | 5-minute TTL cache + use-stale-on-failure. Reaper degrades gracefully (skips) instead of misbehaving. |
| R8 | Audit table grows unbounded | Low | Low | `system.retention_audit` reaper trims `retention_runs` per its own configured window. |
| R9 | Future "delete my account" task duplicates cascade logic | Medium | Low | PRD §9a forward note: that task reuses these per-service cascade implementations. Documented in P6.1. |

## 7. Success Metrics

Operational (per service, after Phase 2 completes):
- `retention_run_failures_total == 0` for at least one full ~6h cycle.
- `retention_run_duration_seconds` p95 per `(tenant, category)` < 60s.
- `retention_deleted_total` non-zero for time-series categories within 24h of activation.

Behavioral:
- Zero orphan rows in dependent tables (e.g., `performance_sets` with no parent `performance`). Verified by spot-check query after Phase 2 per service.
- Zero diff between old and new package-service reaper across 3 cycles before Phase 4 switchover.

Product:
- "Data Retention" settings page renders for at least one staging household; an edit round-trips through PATCH → GET.
- Manual purge from the UI completes end-to-end; `retention_runs` row appears in the recent-purges panel.

## 8. Required Resources and Dependencies

- **Codebase areas touched**: `shared/go/retention/` (new), `services/account-service/`, `services/productivity-service/`, `services/recipe-service/`, `services/tracker-service/`, `services/workout-service/`, `services/calendar-service/`, `services/package-service/`, `frontend/` (settings UI).
- **Infrastructure**: Postgres advisory locks (already available); Prometheus scraping (already in place); existing internal service-to-service token mechanism.
- **External coordination**: none — all changes are within the home-hub repo.
- **Skills**: `backend-dev-guidelines` for Go services, `frontend-dev-guidelines` for the settings UI.
- **Local verification**: `scripts/local-up.sh` for end-to-end Docker Compose runs after each phase.

## 9. Timeline Estimates

Calendar estimate assumes one engineer working primarily on this task. The Phase 4 wall-clock minimum is fixed by the one-week dry-run soak.

| Phase | Effort | Wall clock |
|---|---|---|
| Phase 1 — Foundation | L + M + M | ~1 week |
| Phase 2 — Reapers (×5 services) | L per service | ~2.5 weeks (sequential per PRD) |
| Phase 3 — Write APIs / manual purge | M × 5 | ~1 week |
| Phase 4 — Package-service refactor | M + 1-week soak + M | ~1.5 weeks |
| Phase 5 — UI | L + M + M + M | ~1.5 weeks |
| Phase 6 — Docs / closeout | S × 2 | ~1 day |
| **Total** | | **~7.5 weeks** |

Phases 2 and 5 can overlap once Phase 3 is complete, compressing the total to ~6 weeks if the UI work is done in parallel with the later Phase 2 services.

## 10. Out of Scope (recap from PRD §2 non-goals)

- GDPR-style "delete my entire account" cascade flow (tracked separately; see PRD §9a forward note).
- Retention for short-TTL infrastructure data (auth refresh tokens, calendar OAuth state, weather cache).
- Backup/restore from cold storage. Reaped data is gone.
- Cross-service event-driven reaping. Each service polls policy on its own schedule.
- A general-purpose settings service. Policy lives in account-service.
