# Retention Data Model

## New tables

### `retention_policy_overrides` (account-service)

```sql
CREATE TABLE retention_policy_overrides (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      uuid NOT NULL,
  scope_kind     text NOT NULL CHECK (scope_kind IN ('household','user')),
  scope_id       uuid NOT NULL,
  category       text NOT NULL,
  retention_days int  NOT NULL CHECK (retention_days BETWEEN 1 AND 3650),
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, scope_kind, scope_id, category)
);

CREATE INDEX idx_retention_overrides_scope
  ON retention_policy_overrides (tenant_id, scope_kind, scope_id);
```

A row exists only when the household or user has explicitly overridden a category. Absence means "use system default". Clearing an override is a `DELETE`, not setting `retention_days = NULL`.

System defaults are compiled into account-service Go code as a `map[Category]int` so they version with the deployment and require no DB writes on rollout.

### `retention_runs` (one per reaper-owning service)

```sql
CREATE TABLE retention_runs (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   uuid NOT NULL,
  scope_kind  text NOT NULL,
  scope_id    uuid NOT NULL,
  category    text NOT NULL,
  trigger     text NOT NULL CHECK (trigger IN ('scheduled','manual')),
  dry_run     boolean NOT NULL DEFAULT false,
  scanned     int  NOT NULL DEFAULT 0,
  deleted     int  NOT NULL DEFAULT 0,
  started_at  timestamptz NOT NULL,
  finished_at timestamptz,
  error       text
);

CREATE INDEX idx_retention_runs_tenant_started
  ON retention_runs (tenant_id, started_at DESC);

CREATE INDEX idx_retention_runs_category_started
  ON retention_runs (category, started_at DESC);
```

Each reaper-owning service gets its own copy. The `system.retention_audit` reaper is responsible for trimming old rows in its own table.

## Existing tables — what the reaper queries

### productivity-service

- `tasks` — reap where `deleted_at IS NOT NULL AND deleted_at < now() - restore_window` (cascade per §4.5: `subtasks`, `reminders`, `task_restorations`).
- `tasks` — reap where `completed_at IS NOT NULL AND completed_at < now() - completed_window`.

### recipe-service

- `recipes` — reap where `deleted_at IS NOT NULL AND deleted_at < now() - restore_window` (cascade: ingredients, instructions, `recipe_restorations`, meal-plan slot references).
- `recipe_restorations` — reap where `created_at < now() - audit_window`.

### tracker-service

- `tracking_entries` — reap where `entry_date < now() - entries_window`. Does **not** cascade upward.
- `tracking_items` — reap where `deleted_at IS NOT NULL AND deleted_at < now() - restore_window` (cascade: all `tracking_entries` for the item).

### workout-service

- `performances` + `performance_sets` — reap where `performed_at < now() - performances_window`. Sets cascade with their parent performance.
- `themes` / `regions` / `exercises` — reap where `deleted_at IS NOT NULL AND deleted_at < now() - catalog_window`.
  - Cascade order: `theme → regions → exercises → performances → performance_sets`.
  - Cascade always uses one transaction per top-level parent.

### calendar-service

- `calendar_events` — reap where `end_time < now() - past_events_window`. Leaf-level; no children.

### package-service

- Existing behavior, refactored: `packages` reap where `archived_at < now() - archived_delete_window`; cascade to `tracking_events`. Stale-marking and archive transition unchanged.

## ER summary

```
account-service
  retention_policy_overrides
    (tenant_id, scope_kind, scope_id, category) → retention_days

each reaper-owning service
  retention_runs
    (tenant_id, scope_kind, scope_id, category, trigger, scanned, deleted, ...)

  domain tables (existing)
    └── reaper queries with computed window from policy
        └── application-level cascade (no DB CASCADE)
```

## Cascade implementation note

Cascades are application-level inside a single transaction, **not** `ON DELETE CASCADE` at the DB level. This is deliberate:

- The reaper needs to count exactly how many rows it removed across the cascade for the audit row.
- DB cascades hide volume from the application and would silently inflate the audit `deleted` count or require triggers to populate it.
- Application cascades let us add per-relation safety checks later (e.g., "do not delete this exercise if a retained performance still references it") without DB schema changes.
