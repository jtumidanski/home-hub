# Migration Plan

The rollout is staged to avoid two failure modes: deleting data based on missing policy, and breaking package-service's existing cleanup behavior.

## Phase 1 — Foundation (no behavior change)

1. **Create `shared/go/retention` package**
   - Ticker loop with jitter and graceful shutdown.
   - Advisory-lock helper (Postgres `pg_try_advisory_xact_lock` keyed on hash of `(tenant_id, category)`).
   - Policy client with 5-minute TTL cache and the cache-miss safety guard.
   - Audit writer for `retention_runs`.
2. **Account-service: schema + defaults**
   - Migration adds `retention_policy_overrides`.
   - Compile system defaults from PRD §4.2 into a Go constant table.
3. **Account-service: read APIs**
   - `GET /api/v1/retention-policies` returning fully-resolved policy.
   - Authorization helpers for household-admin and user-scope checks.

At end of Phase 1: nothing is being reaped that was not already being reaped. Package-service is still reading env vars.

## Phase 2 — Reapers (one service at a time)

For each of productivity, recipe, tracker, workout, calendar:

1. Add `retention_runs` table migration.
2. Wire in `shared/go/retention` reaper loop in `cmd/main.go`.
3. Implement category handlers with cascade rules per PRD §4.5.
4. Add Prometheus metrics.
5. Add unit tests for each category handler (boundary values, cascade integrity).
6. Add an integration test that runs the reaper twice in parallel and verifies the advisory lock prevents double-processing.
7. Deploy the service. Watch metrics for one full reaper cycle (~6h) before moving to the next service.

Order: **productivity → recipe → tracker → workout → calendar**. Productivity first because it has the most active soft-delete usage and the smallest blast radius.

Each service deploy is independently safe to roll back: removing the reaper goroutine reverts behavior with no schema downgrade required (the `retention_runs` table can be left in place).

## Phase 3 — Account-service write APIs and manual purge

1. `PATCH /api/v1/retention-policies/household/:household_id` and `PATCH .../user`.
2. `POST /api/v1/retention-policies/purge` with fan-out to internal endpoints.
3. Add `POST /internal/retention/purge` to each reaper-owning service.
4. `GET /api/v1/retention-runs` aggregated audit feed.

## Phase 4 — Package-service refactor

This phase replaces existing env-var configuration with policy lookups. **Risk: regression of current production behavior.**

1. Default values in account-service for `package.archive_window` and `package.archived_delete_window` must match current env-var values exactly.
2. Add the policy client + reaper loop alongside the existing cleanup loop, in dry-run mode (logs only, no deletes).
3. Run for one full week. Compare what the new reaper would delete vs what the old cleanup loop actually deleted.
4. Once the diff is zero across at least 3 reaper cycles, switch over: remove the old `internal/poller/cleanup.go` and rely on the shared reaper.
5. Remove the env-var config knobs.

## Phase 5 — UI

1. New "Data Retention" settings page using existing settings layout.
2. Recent-purges panel powered by `GET /api/v1/retention-runs`.
3. Per-category "Purge now" affordance with confirmation modal.
4. Confirmation modal warning when shrinking a window (per §9 open question).

## Rollback

- Phases 1, 2, 3, 5 are independently revertable. Removing the reaper goroutine or UI components has no data side effects.
- Phase 4 has the highest risk; the dry-run gate is the protection. If the diff is non-zero, do not switch over — investigate first.
- The new tables (`retention_policy_overrides`, `retention_runs`) are additive and should not be dropped on rollback; their presence is harmless if unused.

## Verification

- Each phase's deploy is followed by a Prometheus check: `retention_run_failures_total == 0` for at least one full reaper cycle.
- A spot-check query on a staging tenant after each Phase 2 step confirms expected row deletions match audit counts.
- Phase 4 spot-check: `SELECT count(*) FROM packages WHERE archived_at < now() - interval '30 days'` should be 0 after the switchover.
