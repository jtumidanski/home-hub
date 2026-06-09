# Reminder Data Retention — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-06-09
---

## 1. Overview

The Home Hub data retention framework (shipped in task-036) automatically ages out
old records for every household and user entity — tasks, recipes, tracker entries,
workout performances, calendar events, packages, and dashboards. Each is registered
as a shared retention `Category` with a default window, a per-service reaper handler,
and a per-tenant/scope policy override surfaced in the Data Retention settings page.

**Reminders are the one household entity the framework never covered.** The
`reminders` table (`productivity-service`) has no soft-delete column, no retention
category, no reaper handler, and no row in the settings UI. Dismissing or snoozing a
reminder only sets state flags (`last_dismissed_at`, `last_snoozed_until`); nothing
ever removes the row. As a result the `reminders`, `reminder_dismissals`, and
`reminder_snoozes` tables grow without bound. The task-036 audit explicitly noted
reminders were left out — they are standalone household entities (no `task_id` FK),
so they were not swept up by the task cascade, and they never got their own category.

This task closes that gap by adding reminders to the existing shared retention
framework, following the same two-category pattern already used for tasks: a primary
"aging" category that soft-deletes reminders past their useful life, and a
"restore window" category that hard-deletes the soft-deleted rows after a grace
period (cascading to `reminder_dismissals` and `reminder_snoozes`).

## 2. Goals

Primary goals:

- Old reminders are removed automatically, bounding the growth of the `reminders`,
  `reminder_dismissals`, and `reminder_snoozes` tables.
- Reminders are configurable from the Data Retention settings page exactly like every
  other entity, with per-household policy overrides honored.
- Removal is two-staged with a recovery grace period: a reminder is first soft-deleted
  (hidden, recoverable), then hard-deleted only after the restore window elapses.
- The implementation mirrors the established task retention pattern so it is
  unsurprising to maintainers and reuses the shared framework end-to-end.

Non-goals:

- No recurring-reminder regeneration, scheduling, or notification-delivery changes.
- No change to dismiss / snooze behavior or their API endpoints.
- No user-facing "trash / restore reminders" UI. Soft-delete here is reaper-driven,
  not user-initiated; the restore window is an internal safety buffer, not a UX feature.
- No change to how reminders are created, read, edited, dismissed, or snoozed beyond
  excluding soft-deleted rows from existing read paths.

## 3. User Stories

- As a household member, I want stale reminders to stop accumulating so my data stays
  tidy and the app stays fast.
- As a household admin, I want to see and adjust how long reminders are kept from the
  Data Retention settings page, just like I can for tasks and recipes.
- As a privacy-conscious user, I want dismissed and long-abandoned reminders to be
  permanently deleted on a predictable schedule rather than living in the database
  forever.
- As a maintainer, I want reminder retention to use the same shared framework and
  patterns as every other entity so there is nothing bespoke to reason about.

## 4. Functional Requirements

### 4.1 New retention categories

Two new household-scoped categories are added to the shared retention registry
(`shared/go/retention/category.go`):

| Category constant | String value | Default window | Scope | Max days |
| --- | --- | --- | --- | --- |
| `CatProductivityReminders` | `productivity.reminders` | 365 days | household | 3650 |
| `CatProductivityDeletedRemindersRestoreWindow` | `productivity.deleted_reminders_restore_window` | 30 days | household | 365 |

- `CatProductivityReminders` is the primary "aging" category. Its default of **365
  days** was selected for consistency with `productivity.completed_tasks` and
  `calendar.past_events`.
- `CatProductivityDeletedRemindersRestoreWindow` is the restore window. Its **30-day**
  default follows the convention of all four existing `*_restore_window` categories
  (recipe, tracker, workout, productivity-tasks). The `_restore_window` suffix
  automatically caps `MaxDays()` at 365 via existing logic in `category.go`.
- Both categories must be added to `Defaults` and `scopeKindOf` so the existing
  `All()`, `HouseholdCategories()`, validation, and account-service enumeration pick
  them up with no further wiring.

### 4.2 Primary reaper — soft-delete aged reminders

A new `Reminders` handler is registered in
`productivity-service/internal/retention/wire.go` alongside the existing task
handlers. It implements the shared `CategoryHandler` interface.

- **Category:** `CatProductivityReminders`.
- **Scope discovery:** distinct `(tenant_id, household_id)` from the `reminders`
  table, household-scoped (mirrors `CompletedTasks.DiscoverScopes`).
- **Reap criterion** (cutoff = `now() - window`):

  ```
  deleted_at IS NULL
  AND (
        (last_dismissed_at IS NOT NULL AND last_dismissed_at < cutoff)
     OR (scheduled_for < cutoff)
  )
  ```

  A reminder is soft-deleted when it has been **dismissed** longer than the window
  ago, **or** its scheduled time is more than the window in the past (abandoned /
  never-dismissed). Both branches use the same configured window.
- **Action:** set `deleted_at = now()` on matching rows (soft delete). Do **not** hard
  delete and do **not** touch `reminder_dismissals` / `reminder_snoozes` at this stage.
- Respect `dryRun`: when set, count matching rows but perform no writes (shared
  framework contract).
- Return `ReapResult{Scanned, Deleted}` where `Deleted` = rows soft-deleted.

> **Accepted trade-off:** the `scheduled_for < cutoff` branch can soft-delete a
> reminder that was never dismissed and is still rendered as "overdue" in the UI.
> With a 365-day window this only affects reminders overdue by more than a year, which
> are effectively abandoned, and the restore window provides a recovery buffer. This
> was an explicit product decision.

### 4.3 Restore-window reaper — hard-delete with cascade

A second handler, `DeletedRemindersRestoreWindow`, hard-deletes soft-deleted
reminders after the restore window and cascades to child tables.

- **Category:** `CatProductivityDeletedRemindersRestoreWindow`.
- **Scope discovery:** same as 4.2.
- **Reap criterion** (cutoff = `now() - window`):

  ```
  deleted_at IS NOT NULL AND deleted_at < cutoff
  ```

- **Action:** within a single transaction, delete the matching reminder rows and
  cascade-delete dependent rows in `reminder_dismissals` and `reminder_snoozes`
  (matched by `reminder_id`). This mirrors `cascadeDeleteTasks` →
  `task_restorations`.
- Respect `dryRun`.
- Return `ReapResult{Scanned, Deleted}` where `Deleted` = total rows removed across
  all three tables.

### 4.4 Read-path exclusion of soft-deleted reminders

Because reminders adopt the same plain `deleted_at *time.Time` column as tasks (not
GORM's automatic soft-delete), every existing read path that lists or fetches
reminders MUST be updated to exclude rows where `deleted_at IS NOT NULL`:

- Reminder list / fetch queries in `reminder/` (and any "active reminders" /
  summary aggregation in `summary/` that reads the `reminders` table).
- Dismiss and snooze handlers that load a reminder by id should treat a soft-deleted
  reminder as not found.

The exact set of affected queries is enumerated during the design phase; acceptance
requires that no soft-deleted reminder is returned by any API read path.

### 4.5 Policy overrides

Per-household overrides for both new categories work automatically through the shared
`PolicyClient` / account-service `retention_policy_overrides` mechanism. No
productivity-service or account-service code changes are required beyond registering
the categories (4.1) — account-service enumerates via `HouseholdCategories()`.

### 4.6 Settings UI

`frontend/src/pages/DataRetentionPage.tsx` must display human-readable labels for both
new categories in `CATEGORY_LABELS`:

- `productivity.reminders` → "Reminders"
- `productivity.deleted_reminders_restore_window` → "Deleted reminders (restore window)"

The page already fetches the full category set from the API and renders each row
generically, so only the label map needs updating.

## 5. API Surface

No new or modified public endpoints in productivity-service. Reminder CRUD, dismiss,
and snooze endpoints are unchanged except that soft-deleted reminders are no longer
returned (4.4).

Retention is exercised through the existing shared framework surfaces:

- **Account-service** `GET /api/v1/retention-policies` automatically includes the two
  new categories (with `source: "default"` until overridden) and accepts overrides for
  them via the existing override endpoint. Validation (1 ≤ days ≤ MaxDays) is enforced
  by the shared `Category.Validate`.
- **Productivity-service** internal retention endpoints (mounted by
  `sr.MountInternalEndpoints`) gain the two categories automatically, including the
  internal purge / dry-run trigger guarded by the internal token.

Request/response shapes are unchanged JSON:API; the category list simply grows by two
entries.

## 6. Data Model

### 6.1 `reminders` — add soft-delete column

Add a nullable `deleted_at` column, mirroring the `tasks` table exactly:

- Entity field: `DeletedAt *time.Time \`gorm:"index"\`` on
  `reminder.Entity` (`reminder/entity.go`).
- Model field `deletedAt *time.Time` with accessor, builder setter, and round-trip
  through `ToEntity` / `Make`, mirroring `task` model plumbing.
- Migration: additive `AutoMigrate` adds the nullable, indexed column. No backfill —
  existing rows have `deleted_at = NULL` (not soft-deleted), which is correct.

> Note: this is a **plain nullable timestamp with an index**, NOT GORM's
> `gorm.DeletedAt`. This matches the task implementation, which filters
> `deleted_at IS NULL` manually rather than relying on GORM's automatic scope. Adopting
> the same approach keeps reminders consistent with tasks and keeps the reaper queries
> explicit.

### 6.2 Cascade tables (unchanged schema)

- `reminder_dismissals` (`reminder_id` FK) — hard-deleted in the restore-window
  cascade.
- `reminder_snoozes` (`reminder_id` FK) — hard-deleted in the restore-window cascade.

No schema change to these tables; they are cascade targets only.

### 6.3 Multi-tenancy

All reaper queries filter by `tenant_id` (and `household_id` for scope), consistent
with the rest of the framework. Soft-delete and cascade operate strictly within a
single scope per transaction.

## 7. Service Impact

| Service | Change |
| --- | --- |
| `shared/go/retention` | Add two `Category` constants + `Defaults` + `scopeKindOf` entries (`category.go`). No logic changes — enumeration, validation, and MaxDays are derived. |
| `services/productivity-service` | Add `deleted_at` to reminder entity/model + migration; add `Reminders` and `DeletedRemindersRestoreWindow` reaper handlers; register both in `retention.Setup`/`wire.go`; exclude soft-deleted reminders from all read paths (reminder, summary, dismiss/snooze lookups). |
| `services/account-service` | None (auto-enumerates the new categories). Add/extend tests asserting the two categories appear in resolved policies. |
| `frontend` | Add two `CATEGORY_LABELS` entries in `DataRetentionPage.tsx`. |

## 8. Non-Functional Requirements

- **Performance:** reaper queries must be index-backed. Add the `deleted_at` index
  (6.1); existing indexes cover `scheduled_for` and `tenant_id`/`household_id`. Reaping
  runs in the existing jittered background loop and per-scope transactions — no new
  loops or schedulers.
- **Safety:** the two-stage soft-delete → restore-window design guarantees no reminder
  is hard-deleted without first spending the full restore window soft-deleted, giving a
  recovery buffer if the aging criterion misfires. All deletes are tenant- and
  scope-scoped and run inside transactions; the restore-window cascade is atomic across
  the three tables.
- **Observability:** both categories emit the shared Prometheus retention metrics
  (scanned/deleted per category) and write audit rows to the productivity-service
  `retention_runs` table automatically through the shared reaper — no extra
  instrumentation required.
- **Multi-tenancy:** every query is tenant-scoped; policy overrides are per
  household/scope.
- **Backward compatibility:** the migration is additive and non-destructive; default
  windows are long (365/30 days) so no existing reminder is removed immediately on
  deploy.

## 9. Open Questions

1. **Restore-window default (30 days):** chosen to match the four existing
   `*_restore_window` categories. Confirm 30 days is the desired grace period for
   reminders specifically, or adjust.
2. **Summary/aggregation read paths:** the design phase must enumerate exactly which
   queries in `summary/` (and any cross-entity reads) touch the `reminders` table so
   the `deleted_at IS NULL` filter is applied everywhere. Are there read paths outside
   `productivity-service` that query reminders directly? (Expected: none — reminders
   are owned solely by productivity-service.)
3. **Dismiss/snooze on a soft-deleted reminder:** confirm the desired behavior is
   "treat as not found" (proposed) versus silently no-op.

## 10. Acceptance Criteria

- [ ] `productivity.reminders` (365d default, household) and
      `productivity.deleted_reminders_restore_window` (30d default, household) exist in
      `shared/go/retention/category.go` with `Defaults` and `scopeKindOf` entries.
- [ ] `Category.Validate`, `All()`, and `HouseholdCategories()` return the two new
      categories; restore-window `MaxDays()` is 365, primary is 3650.
- [ ] `reminders` table has a nullable, indexed `deleted_at` column added via additive
      migration; the entity/model round-trips it like `tasks`.
- [ ] The `Reminders` reaper soft-deletes (sets `deleted_at`) reminders that are
      dismissed-aged OR scheduled-past beyond the window, and only those; it does not
      touch already-soft-deleted rows or child tables.
- [ ] The `DeletedRemindersRestoreWindow` reaper hard-deletes reminders whose
      `deleted_at` is older than the restore window and cascades to
      `reminder_dismissals` and `reminder_snoozes` atomically in one transaction.
- [ ] Both reapers honor `dryRun` (count, no writes) and are registered in
      `retention.Setup` so the background loop runs them.
- [ ] No API read path (list, fetch, summary, dismiss/snooze lookup) returns a
      soft-deleted reminder.
- [ ] Account-service `GET /retention-policies` includes both categories with correct
      defaults and accepts valid per-household overrides for them.
- [ ] `DataRetentionPage.tsx` shows "Reminders" and "Deleted reminders (restore
      window)" labels for the two categories.
- [ ] Both new categories emit retention metrics and write `retention_runs` audit rows.
- [ ] `go build ./...` and `go test ./...` pass for `shared/go/retention`,
      `productivity-service`, and `account-service`; frontend type-check/tests pass.
- [ ] Docker builds succeed for productivity-service and account-service (shared
      library change).
