# Reminder Data Retention — Design

Status: Approved for planning
Created: 2026-06-09
PRD: `prd.md`
---

## 1. Summary

Bring `reminders` into the shared retention framework using the **same two-stage
pattern as tasks**, but with one deliberate divergence: for reminders the *primary*
reaper performs a **reaper-driven soft-delete** (there is no user-facing trash UI),
whereas the *restore-window* reaper performs a **hard-delete cascade** that mirrors
`DeletedTasksRestoreWindow` / `cascadeDeleteTasks` exactly.

The change is mechanical and additive:

1. Two new household-scoped `Category` constants in `shared/go/retention/category.go`.
2. A nullable, indexed `deleted_at` column on `reminders`, plumbed through the
   immutable model exactly like `tasks`.
3. Two `CategoryHandler` implementations in `productivity-service/internal/retention`,
   registered in `wire.go`.
4. A `deleted_at IS NULL` filter applied to every reminder read and lookup path.
5. Two `CATEGORY_LABELS` entries in `DataRetentionPage.tsx`.

Account-service, the reaper loop, advisory locking, metrics, audit rows, dry-run, and
policy overrides all work unchanged — they enumerate categories dynamically.

This document records the architecture, the verified read-path inventory (PRD open
question 2), the resolution of all three PRD open questions, and the alternatives
considered.

## 2. Verified context (ground truth)

All file:line references below were read directly from the worktree, not inferred.

### 2.1 Shared framework contract

`shared/go/retention/category.go`
- `Category` enum + `Defaults` + `scopeKindOf`. `All()`, `HouseholdCategories()`,
  `Validate`, `MaxDays()` all derive from these two maps — adding a category to both
  maps is sufficient for enumeration, validation, and the `_restore_window` 365-day cap.

`shared/go/retention/reaper.go`
- `CategoryHandler` interface: `Category()`, `DiscoverScopes(ctx, db) ([]Scope, error)`,
  `Reap(ctx, tx, scope, retentionDays int, dryRun bool) (ReapResult, error)`.
- `Scope{TenantId, Kind, ScopeId}`; `ReapResult{Scanned, Deleted}`.
- `Reaper.RunOne` wraps each `Reap` in a transaction, takes a per-scope advisory lock,
  and — critically — **rolls the transaction back via the `errDryRunRollback` sentinel
  when `dryRun` is true**. The handler does not have to branch on `dryRun`.
- `Reaper.RunTick` calls `DiscoverScopes` per handler and `RunOne(..., dryRun=false)`
  for each scope on the scheduled loop.

### 2.2 Task handlers — the template

`productivity-service/internal/retention/handlers.go`
- `CompletedTasks.Reap` (l.42): plucks ids by criterion, calls `cascadeDeleteTasks`.
  **Note: the task *primary* handler hard-deletes** (completed tasks already pass
  through a separate user-initiated soft-delete lifecycle).
- `DeletedTasksRestoreWindow.Reap` (l.74): plucks `deleted_at IS NOT NULL AND
  deleted_at < cutoff`, calls `cascadeDeleteTasks`. `DiscoverScopes` reuses
  `CompletedTasks{}.DiscoverScopes` (l.70).
- `cascadeDeleteTasks` (l.97): inside the supplied `tx`, deletes `task_restorations`
  (child) then `tasks` (parent), summing `RowsAffected`.
- **Neither task handler reads the `dryRun` parameter.** Counts come from real writes
  that the framework then rolls back.

`wire.go` `Setup` registers handlers via `sr.New("productivity-service", db, pc,
metrics, l, CompletedTasks{}, DeletedTasksRestoreWindow{}, AuditTrim{})`.

### 2.3 Task `deleted_at` plumbing — the model template

`task/entity.go`: `DeletedAt *time.Time \`gorm:"index"\``; round-tripped in `ToEntity`
and `Make` via `SetDeletedAt`. `task/model.go`: private `deletedAt`, accessor
`DeletedAt()`, convenience `IsDeleted()`. `task/builder.go`: `deletedAt` field +
`SetDeletedAt`. `task/provider.go`: read queries add `deleted_at IS NULL`
(`getAll(includeDeleted)` gates it behind a flag; `getByStatus`, the counts, etc.
hard-code it).

### 2.4 Reminder domain — current state (no soft-delete today)

`reminder/entity.go`: `Entity` has **no** `deleted_at`. `Migration` = `AutoMigrate(&Entity{})`.
`reminder/model.go`, `builder.go`: immutable model with builder; no deleted plumbing.

`reminder/provider.go` — every read of the `reminders` table:
- `getByID(id)` (l.9) — `WHERE id = ?` (no soft-delete filter).
- `getAll()` (l.15) — order only, no filter.
- `countDueNow(db)` (l.21), `countUpcoming(db)` (l.29), `countSnoozed(db)` (l.37).

`reminder/administrator.go` — every mutation that targets a reminder by id:
- `update` (l.29) — `First(&e)` by id, then `Save`.
- `dismiss` (l.45) — `Updates` `WHERE id = ?`; **already** returns
  `gorm.ErrRecordNotFound` when `RowsAffected == 0`.
- `snooze` (l.60) — `Updates` `WHERE id = ?`; **does not** check `RowsAffected`
  (silent no-op on a missing/soft-deleted row today).
- `deleteByID` (l.68) — hard delete `WHERE id = ?`.

`reminder/processor.go`: `ByIDProvider`/`AllProvider` wrap the provider funcs;
`Dismiss`/`Snooze` call `dismiss`/`snooze`; `DueNowCount`/`UpcomingCount`/`SnoozedCount`
call the counts.

`summary/processor.go`: `ReminderSummary` and `DashboardSummary` touch reminders
**only** through `remProc.DueNowCount/UpcomingCount/SnoozedCount` — i.e. transitively
through `provider.go`. There is no direct `reminders` query in `summary/`.

Child tables: `reminder/dismissal/entity.go` (`reminder_dismissals`, `ReminderId` FK),
`reminder/snooze/entity.go` (`reminder_snoozes`, `ReminderId` FK). No DB-level
`ON DELETE CASCADE`; cascade is done explicitly in Go (as for tasks).

### 2.5 Account-service & frontend

`account-service/internal/retention/processor.go` `ResolveAll` (l.76, l.87) merges
DB overrides over `HouseholdCategories()` / `UserCategories()`. New household
categories appear automatically; **no account-service code change** is required.

`frontend/src/pages/DataRetentionPage.tsx`: `CATEGORY_LABELS` map (l.37) +
`categoryMax` helper that already returns 365 for any `*_restore_window` key. Rows
render generically from the API response, sorted by key.

## 3. Architecture

### 3.1 Categories (`shared/go/retention/category.go`)

Add to the const block, `Defaults`, and `scopeKindOf`:

| Constant | Value | Default | Scope | MaxDays |
| --- | --- | --- | --- | --- |
| `CatProductivityReminders` | `productivity.reminders` | 365 | household | 3650 |
| `CatProductivityDeletedRemindersRestoreWindow` | `productivity.deleted_reminders_restore_window` | 30 | household | 365 (auto via suffix) |

No other shared change. `All()`, `HouseholdCategories()`, `Validate`, `MaxDays()`,
account enumeration, and the productivity internal endpoints pick these up for free.

### 3.2 Soft-delete column (`reminder/entity.go`, `model.go`, `builder.go`)

Mirror the task plumbing field-for-field:
- `Entity`: add `DeletedAt *time.Time \`gorm:"index"\``; include in `ToEntity` and
  `Make` (`SetDeletedAt(e.DeletedAt)`).
- `Model`: private `deletedAt *time.Time`; accessor `DeletedAt()`; convenience
  `IsDeleted()`.
- `Builder`: `deletedAt` field + `SetDeletedAt(*time.Time)`; carried through `Build()`.

Migration is the existing `AutoMigrate(&Entity{})` — additive, nullable, indexed,
no backfill. Existing rows get `deleted_at = NULL` (correct: not soft-deleted).

This is a **plain nullable timestamp + manual filtering**, deliberately *not*
`gorm.DeletedAt`, to stay identical to tasks and keep reaper queries explicit.

### 3.3 Primary handler — `Reminders` (soft-delete UPDATE)

`productivity-service/internal/retention/handlers.go` (new struct, registered in
`wire.go`).

- `Category()` → `CatProductivityReminders`.
- `DiscoverScopes` → `SELECT DISTINCT tenant_id, household_id FROM reminders`,
  `Kind: ScopeHousehold` (structurally identical to `CompletedTasks.DiscoverScopes`,
  against `reminders`).
- `Reap` → a **single bulk UPDATE** inside `tx`, `cutoff = now().UTC() - days*24h`:

  ```
  UPDATE reminders
     SET deleted_at = <now>
   WHERE tenant_id = ? AND household_id = ?
     AND deleted_at IS NULL
     AND (
           (last_dismissed_at IS NOT NULL AND last_dismissed_at < cutoff)
        OR (scheduled_for < cutoff)
         )
  ```

  `ReapResult{Scanned: RowsAffected, Deleted: RowsAffected}` (same shape as
  `AuditTrim`). No pluck-then-mutate is needed because this stage does not cascade.
- Does **not** touch `reminder_dismissals` / `reminder_snoozes`.
- Does **not** branch on `dryRun`: the framework rolls the UPDATE back. The reported
  count still reflects the rows that *would* be soft-deleted (the UPDATE executes,
  then rolls back). See §5.1.

> **Why a soft-delete here, not a hard-delete like `CompletedTasks`:** reminders have
> no user-initiated trash lifecycle, so this reaper *is* the soft-delete. The hidden
> row then ages through the restore window before §3.4 hard-deletes it. This is the
> one intentional structural difference from the task handlers.

### 3.4 Restore-window handler — `DeletedRemindersRestoreWindow` (hard-delete cascade)

Mirrors `DeletedTasksRestoreWindow` + `cascadeDeleteTasks` precisely.

- `Category()` → `CatProductivityDeletedRemindersRestoreWindow`.
- `DiscoverScopes` → reuse `Reminders{}.DiscoverScopes` (as tasks reuse
  `CompletedTasks{}.DiscoverScopes`).
- `Reap` → pluck ids `WHERE tenant_id=? AND household_id=? AND deleted_at IS NOT NULL
  AND deleted_at < cutoff`; if none, return zero; else call
  `cascadeDeleteReminders(tx, ids)`. `ReapResult{Scanned: len(ids), Deleted: total
  across three tables}`.
- New `cascadeDeleteReminders(tx, ids)` — children first, then parent, summing
  `RowsAffected`, all in the supplied `tx`:

  ```
  DELETE FROM reminder_snoozes     WHERE reminder_id IN (ids)
  DELETE FROM reminder_dismissals  WHERE reminder_id IN (ids)
  DELETE FROM reminders            WHERE id          IN (ids)
  ```

  (`reminder_id`/`id` are `uuid`; pluck ids as `[]string` like the task handler.)

### 3.5 Read- and lookup-path exclusion

Because reminders use a manual `deleted_at` (not GORM's auto scope), every path that
returns or mutates a reminder by identity must exclude soft-deleted rows. The complete,
verified inventory (PRD OQ2 resolved):

| File:line | Function | Change |
| --- | --- | --- |
| `provider.go:9` | `getByID` | add `deleted_at IS NULL` to the `Where` |
| `provider.go:15` | `getAll` | add `Where("deleted_at IS NULL")` |
| `provider.go:21` | `countDueNow` | append `AND deleted_at IS NULL` |
| `provider.go:29` | `countUpcoming` | append `AND deleted_at IS NULL` |
| `provider.go:37` | `countSnoozed` | append `AND deleted_at IS NULL` |
| `administrator.go:29` | `update` | add `deleted_at IS NULL` to the `First` `Where` (yields `ErrRecordNotFound` for a soft-deleted row) |
| `administrator.go:45` | `dismiss` | add `deleted_at IS NULL` to the `Where` (existing `RowsAffected==0 → ErrRecordNotFound` then fires) |
| `administrator.go:60` | `snooze` | add `deleted_at IS NULL` to the `Where` **and** add the missing `RowsAffected==0 → gorm.ErrRecordNotFound` check, for parity with `dismiss` |
| `administrator.go:68` | `deleteByID` | add `deleted_at IS NULL` to the `Where` (a hidden reminder is "not found" to a user delete rather than being hard-deleted out from under the restore window) |

`summary/` needs no direct change — it reaches reminders only through the three
`count*` funcs above. No read path exists outside productivity-service (reminders are
owned solely by this service), so OQ2's expectation holds.

### 3.6 Frontend (`DataRetentionPage.tsx`)

Add to `CATEGORY_LABELS`:
- `"productivity.reminders": "Reminders"`
- `"productivity.deleted_reminders_restore_window": "Deleted reminders (restore window)"`

`categoryMax` already handles the 365 cap via the `_restore_window` suffix. No other
frontend change; rows render and sort generically.

## 4. Resolution of PRD open questions

1. **Restore-window default (30 days):** keep 30. It matches all four existing
   `*_restore_window` categories and the shared 365-day suffix cap. Adopted as the
   default rather than blocking — adjustable later via a per-household override with no
   code change.
2. **Summary / cross-entity read paths:** fully enumerated in §3.5. `summary/` touches
   reminders only via `countDueNow/Upcoming/Snoozed`; no direct `reminders` query and
   no reads outside productivity-service. Adding the filter to those three count
   functions covers all summary/dashboard surfaces.
3. **Dismiss/snooze on a soft-deleted reminder:** "treat as not found." `dismiss`
   already returns `ErrRecordNotFound` on zero rows; `snooze` gains both the
   `deleted_at IS NULL` filter and the missing zero-rows check so the two behave
   identically. `update` and `deleteByID` likewise return not-found.

## 5. Key decisions & trade-offs

### 5.1 `dryRun` is honored by the framework, not the handler
The PRD says each handler should "count matching rows but perform no writes" under
`dryRun`. The **established task pattern does not do this** — handlers ignore the
`dryRun` flag and `Reaper.RunOne` rolls the transaction back via `errDryRunRollback`.
We follow the established pattern: the net effect (no committed writes, accurate
counts) satisfies the PRD's intent and the observable contract, and reusing the
framework end-to-end is itself a primary PRD goal. **Decision: mirror tasks — do not
branch on `dryRun` in either reminder handler.**

### 5.2 Primary stage soft-deletes; only the restore stage cascades
Cascading `reminder_dismissals`/`reminder_snoozes` at soft-delete time was rejected: a
soft-deleted reminder must remain fully recoverable during the restore window, and its
dismiss/snooze history is part of that state. Children are deleted **only** in the
hard-delete cascade (§3.4). This is why the primary handler is a bare UPDATE.

### 5.3 Plain `*time.Time`, not `gorm.DeletedAt`
Using GORM's native soft-delete would auto-filter reads but diverge from tasks and hide
the filter from the explicit reaper queries. Rejected for consistency and
explicitness; we mirror the task column exactly.

### 5.4 Explicit Go cascade, not DB `ON DELETE CASCADE`
The codebase has no FK cascade constraints and deletes children explicitly
(`cascadeDeleteTasks`). We add `cascadeDeleteReminders` in the same style rather than
introducing a migration-managed FK, keeping the atomic delete visible in one place and
consistent with the existing handler.

### 5.5 Single bulk UPDATE vs pluck-then-update for the primary stage
The task primary handler plucks ids then cascades because it must touch a child table.
The reminder primary stage touches no child table, so a single UPDATE is simpler and
`RowsAffected` is an exact soft-delete count. Adopted.

### 5.6 Accepted edge cases
- **Overdue-but-rendered:** `scheduled_for < cutoff` can soft-delete a never-dismissed
  reminder still shown as "overdue." With a 365-day window this only affects reminders
  overdue by >1 year (effectively abandoned); the restore window is the recovery
  buffer. Per PRD, an explicit product decision.
- **Snoozed-but-ancient:** the criterion ignores `last_snoozed_until`, so a reminder
  scheduled >1 year ago but snoozed into the future would still be soft-deleted. This
  is vanishingly rare (snoozes are 10/30/60 minutes) and recoverable within the restore
  window; not worth special-casing.

## 6. Alternatives considered (and rejected)

- **A — Single category, hard-delete aged reminders directly (no restore window).**
  Simpler, one handler. Rejected: no recovery buffer, and the `scheduled_for` branch can
  catch still-overdue rows; the PRD mandates a two-stage safety buffer.
- **B — Reuse `gorm.DeletedAt` automatic soft-delete.** Rejected (§5.3).
- **C — Cascade children at soft-delete time.** Rejected (§5.2) — breaks recoverability.
- **D — Handler branches on `dryRun` to skip writes.** Rejected (§5.1) — diverges from
  the framework's rollback model and the task pattern.
- **E — DB-level `ON DELETE CASCADE` FK on the child tables.** Rejected (§5.4) —
  inconsistent with the existing explicit-cascade approach; adds migration risk.

## 7. Service impact

| Service | Change |
| --- | --- |
| `shared/go/retention` | 2 category constants + `Defaults` + `scopeKindOf` entries (`category.go`). No logic change. |
| `productivity-service` | `deleted_at` on reminder entity/model/builder + `AutoMigrate`; `Reminders` + `DeletedRemindersRestoreWindow` handlers + `cascadeDeleteReminders` in `handlers.go`; register both in `wire.go`; add `deleted_at IS NULL` across the 9 paths in §3.5 (incl. the `snooze` zero-rows fix). |
| `account-service` | None (auto-enumerates). Add/extend a test asserting both categories resolve with defaults 365/30. |
| `frontend` | 2 `CATEGORY_LABELS` entries in `DataRetentionPage.tsx`. |

## 8. Testing strategy

- **shared/go/retention:** assert `CatProductivityReminders` (3650 max) and
  `CatProductivityDeletedRemindersRestoreWindow` (365 max via suffix) are in `All()` /
  `HouseholdCategories()` and `Validate` correctly bounds them.
- **productivity-service retention:** unit tests for both handlers against an in-memory
  / test DB: (a) `Reminders.Reap` soft-deletes exactly the dismissed-aged and
  scheduled-past rows, leaves fresh and already-soft-deleted rows untouched, and never
  writes child tables; (b) `DeletedRemindersRestoreWindow.Reap` hard-deletes only
  `deleted_at < cutoff` rows and cascades to both child tables atomically; (c) both are
  tenant/scope-scoped; (d) `DiscoverScopes` returns distinct household scopes.
- **productivity-service reminder:** tests proving no read/lookup path returns a
  soft-deleted reminder — `getByID`/`getAll`, the three counts, and
  `update`/`dismiss`/`snooze`/`deleteByID` all return `ErrRecordNotFound` (or empty)
  for a soft-deleted row; include the new `snooze` zero-rows case.
- **account-service:** test that resolved household policies include both categories
  with defaults 365 and 30.
- **frontend:** existing `DataRetentionPage` test (or a label assertion) confirms the
  two new labels render.
- **Builds:** `go build ./...` + `go test ./...` for the three Go modules; frontend
  type-check/tests; Docker builds for productivity-service and account-service (shared
  library change per CLAUDE.md).

## 9. Out of scope (per PRD non-goals)

No recurring-reminder regeneration, no notification/scheduling changes, no change to
dismiss/snooze API semantics beyond not-found-on-soft-deleted, and no user-facing
trash/restore UI — the restore window is an internal safety buffer only.
