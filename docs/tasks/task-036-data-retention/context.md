# Task-036 — Context

Last Updated: 2026-04-11

## Source documents

- `prd.md` — Full PRD (v1, Draft, 2026-04-11). Sections 1–10 + §9a forward note.
- `data-model.md` — Schemas for `retention_policy_overrides`, `retention_runs`; per-service reaper queries; cascade implementation note.
- `api-contracts.md` — JSON:API request/response shapes for all endpoints, plus internal endpoint contract and category-ownership map.
- `migration-plan.md` — Phase ordering, rollback notes, verification steps.

## Key files in the codebase (as of 2026-04-11)

### Will be created

- `shared/go/retention/` — new package. `loop.go`, `lock.go`, `policy_client.go`, `audit.go`, `metrics.go`.
- `services/account-service/internal/retention/` — categories table, defaults, overrides repository, JSON:API handlers, fan-out client.
- `services/<svc>/internal/retention/` for each of productivity, recipe, tracker, workout, calendar — category handlers + cascade logic.
- New migrations in account-service (`retention_policy_overrides`) and in each reaper-owning service (`retention_runs`).
- `frontend/src/...` — new "Data Retention" settings page, recent-purges panel, purge-now button, shrink-warning modal.

### Will be modified

- Each reaper-owning service `cmd/main.go` — wire `retention.Loop`.
- `services/package-service/internal/poller/cleanup.go` — **deleted in Phase 4**. Replaced by the shared reaper.
- `services/package-service/cmd/main.go` — replace cleanup loop wiring with shared reaper.
- `docs/architecture.md` — new "Retention Framework" section in Phase 6.

### Reference (read-only context)

- `services/package-service/internal/poller/cleanup.go` — existing three-stage cleanup. The behavioral baseline for Phase 4's dry-run diff.
- `services/package-service/internal/poller/poller.go` — existing ticker pattern; informs the `shared/go/retention` loop design.
- `shared/go/tenant/` — existing multi-tenant helpers; the reaper iterates tenants explicitly using these.
- `shared/go/database/` — existing GORM setup; the advisory-lock helper goes alongside the connection wrappers.
- `shared/go/server/` — existing JSON:API server scaffolding used by all services.

## Architectural decisions (from PRD §6, §9, and elsewhere)

1. **System defaults are compiled-in Go constants**, not DB rows. They version with the deployment and require no migration to update.
2. **Cascades are application-level inside a single transaction**, not DB `ON DELETE CASCADE`. Reason: the reaper needs to count rows for the audit, and we want to add per-relation safety checks later without schema changes.
3. **Cache-miss safety is mandatory.** A reaper that cannot reach account-service and has no cache **must skip** the run, never default to "0 days". Enforced in `shared/go/retention`.
4. **Advisory locks via `pg_try_advisory_xact_lock`** keyed on a hash of `(tenant_id, category)` so two replicas cannot reap the same scope simultaneously.
5. **Dry-run is a transaction-rollback, not a separate code path.** Same scan + cascade walk; the transaction is rolled back at the end. Audit row written with `dry_run = true`.
6. **Aggregated audit feed fans out at request time** in v1 (each service queries its own `retention_runs`). Revisit a central event log only if request latency becomes a problem.
7. **Audit table self-cleans** via the `system.retention_audit` reaper category in each service.
8. **Phase 4 is gated by a one-week dry-run soak** with zero diff vs. existing cleanup loop across at least 3 cycles.

## Default-value table (PRD §4.2 — restate for quick reference)

| Category | Default | Scope |
|---|---|---|
| `productivity.completed_tasks` | 365 | household |
| `productivity.deleted_tasks_restore_window` | 30 | household |
| `recipe.deleted_recipes_restore_window` | 30 | household |
| `recipe.restoration_audit` | 90 | household |
| `tracker.entries` | 730 | user |
| `tracker.deleted_items_restore_window` | 30 | user |
| `workout.performances` | 1825 | user |
| `workout.deleted_catalog_restore_window` | 30 | user |
| `calendar.past_events` | 365 | household |
| `package.archive_window` | 7 | household |
| `package.archived_delete_window` | 30 | household |
| `system.retention_audit` | 180 | household |

Bounds: 1–3650 days for normal categories; 1–365 days for soft-delete restore windows.

## Cascade rules (PRD §4.5 — restate for quick reference)

- **workout-service**: `theme → regions → exercises → performances → performance_sets`. One transaction per top-level parent.
- **recipe-service**: `recipe → ingredients, instructions, recipe_restorations, meal-plan slot references`. Meal plans preserved (slot cleared only).
- **productivity-service**: `task → subtasks, reminders, task_restorations`.
- **tracker-service**: `tracking_item → tracking_entries`. Entry-level reaping does NOT cascade upward.
- **calendar-service**: `calendar_event` is leaf-level.
- **package-service**: existing `package → tracking_events`.

## Category ownership (api-contracts.md §"Category ownership map")

| Category | Owning service |
|---|---|
| `productivity.completed_tasks`, `productivity.deleted_tasks_restore_window` | productivity-service |
| `recipe.deleted_recipes_restore_window`, `recipe.restoration_audit` | recipe-service |
| `tracker.entries`, `tracker.deleted_items_restore_window` | tracker-service |
| `workout.performances`, `workout.deleted_catalog_restore_window` | workout-service |
| `calendar.past_events` | calendar-service |
| `package.archive_window`, `package.archived_delete_window` | package-service |
| `system.retention_audit` | each reaper-owning service (self-reaping) |

## Phase ordering & dependencies (must not be reordered)

1. Phase 1 (foundation, no behavior change) → unblocks everything.
2. Phase 2 (reapers, one service at a time, productivity → recipe → tracker → workout → calendar) → must soak ~6h between services.
3. Phase 3 (write APIs + manual purge) → depends on at least one Phase 2 service for `POST /internal/retention/purge`.
4. Phase 4 (package-service refactor) → highest risk; gated by one-week dry-run soak with zero diff.
5. Phase 5 (UI) → can overlap with later Phase 2 work once Phase 3 endpoints exist.
6. Phase 6 (docs + closeout).

## Forward dependency

A future "delete my account / household" GDPR task (PRD §9a) will reuse the per-service cascade implementations from Phase 2, exposed as a `POST /internal/retention/purge-tenant` variant. That task's PRD must reference task-036.

## Open / monitoring items

- Performance budgets (PRD §8): each `(tenant, category)` reaper run must complete in under 60s p95. Adjust batch size or interval if exceeded.
- Aggregated audit fan-out latency: monitor `GET /api/v1/retention-runs` p95; revisit central event log if it exceeds reasonable bounds.
- Package-service Phase 4 diff: any non-zero diff is a hard blocker — investigate before switching over.
