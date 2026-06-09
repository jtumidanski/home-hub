# Backend Audit — task-051-reminder-data-retention

- **Scope:** Go diff `ef2cd70` → `aa3c9ef` (`shared/go/retention`, `productivity-service/internal/reminder`, `productivity-service/internal/retention`, `account-service/internal/retention` regression test)
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-06-09
- **Build:** PASS
- **Tests:** PASS (productivity-service full suite, shared/go/retention, account-service retention)
- **Overall:** NEEDS-WORK (one Important inconsistency; everything else passes)

## Build & Test Results

- `shared/go/retention`: `go build ./...` PASS; `go test ./... -count=1` → `ok` (0.084s)
- `services/productivity-service`: `go build ./...` PASS; `go test ./... -count=1` → all packages `ok` (reminder, reminder/dismissal, reminder/snooze, retention, summary, task, task/restoration)
- `services/account-service`: `go test ./internal/retention/... -count=1` → `ok`

Objective gate PASSED — proceeded to checklist phase.

## Domain Discovery

- `internal/reminder` — domain package (`model.go` present). Diff modifies model/builder/entity/provider/administrator + adds `softdelete_test.go`.
- `internal/retention` — support/wiring package (no `model.go`, no `resource.go`; implements `sr.CategoryHandler`s). Audited against the established `task` retention pattern rather than DOM-*.
- `shared/go/retention` — shared library; constant + map additions only.
- `internal/reminder/snooze`, `internal/reminder/dismissal` — sub-domain (action-event) packages; touched indirectly via cascade + behavior change.

## Domain Checklist Results — internal/reminder

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists / Build validates | PASS | builder.go:34 `Build()` validates `title != ""`; new `SetDeletedAt` builder.go:124 |
| DOM-02 | ToEntity() method | PASS | entity.go:29; now maps `DeletedAt` entity.go:35 |
| DOM-03 | Make(Entity) (Model,error) | PASS | entity.go:40; sets `SetDeletedAt(e.DeletedAt)` entity.go:51 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:21 `NewProcessor(l logrus.FieldLogger, ...)` |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:34,71,97,132,160 all `NewProcessor(d.Logger(), ...)` |
| DOM-09 | Transform errors handled | PASS | resource.go:104-109, 144-149 check & log err |
| DOM-10 | Test DB registers tenant callbacks | PASS | reminder/processor_test.go:21 `RegisterTenantCallbacks`; `softdelete_test.go` reuses `setupTestDB`/`newTestProcessor` from it |
| DOM-11 | Providers use lazy database.Query | PASS | provider.go:7 `database.Query`, :14 `database.SliceQuery`; new filters appended inside the closures |
| DOM-12 | No os.Getenv in handlers | PASS | grep `os.Getenv` in resource.go → 0 matches |
| DOM-14 | Handlers call processor, not providers | PASS | resource.go handlers call `proc.*` only |
| DOM-15 | No direct db.Create/Save/Delete in handlers | PASS | grep in resource.go → 0 matches; writes go through processor → administrator |
| DOM-16 | administrator.go for writes | PASS | administrator.go `create/update/dismiss/snooze/deleteByID` |
| DOM-17 | Domain error → HTTP status | PARTIAL | reminder/resource.go maps `ErrRecordNotFound`→404 (Update :135, Delete :162). See IMPORTANT-1 for snooze sub-domain gap |
| DOM-20 | Table-driven / error-path tests | PASS | softdelete_test.go round-trip + read/mutation not-found cases; reminder_handlers_test.go table-style fixtures |

### Immutability / builder discipline
- Model fields stay private; `deletedAt` added with accessor `DeletedAt()` (model.go:195) and derived `IsDeleted()` (model.go:196). Builder is the sole mutation path. PASS.

### Soft-delete column type (deliberate, correct)
- `Entity.DeletedAt *time.Time` with `gorm:"index"` (entity.go:20) — NOT `gorm.DeletedAt`. This is the correct choice and matches `task.Entity.DeletedAt` (task/entity.go:22). Using `gorm.DeletedAt` would auto-rewrite the reaper's hard `Delete()` into an UPDATE and auto-filter reads, defeating both `cascadeDeleteReminders` and the explicit `deleted_at IS NULL` filters. PASS.
- Migration via `AutoMigrate` (entity.go:27) adds the indexed column. PASS.

## Retention Handlers — internal/retention

| Check | Status | Evidence |
|-------|--------|----------|
| Mirrors established `task` pattern | PASS | `Reminders`/`DeletedRemindersRestoreWindow` (handlers.go:120-218) structurally mirror `CompletedTasks`/`DeletedTasksRestoreWindow`/`cascadeDeleteTasks` (handlers.go:23-118) |
| Tenant + household scoping explicit | PASS | Reap WHERE clauses filter `tenant_id = ? AND household_id = ?` (handlers.go:151, 178); DiscoverScopes selects `DISTINCT tenant_id, household_id` (handlers.go:134-136) |
| Cascade runs inside reaper-supplied tx | PASS | `cascadeDeleteReminders(tx, ids)` deletes children first (snooze :199, dismissal :205) then parent (:211), all on `tx` |
| Cascade FK columns correct | PASS | `reminder_id IN ?` matches `snooze.Entity.ReminderId` / `dismissal.Entity.ReminderId` (both `gorm:"type:uuid;not null"`); parent `id IN ?` |
| Primary reap does not touch children | PASS | `Reminders.Reap` is a single `Update("deleted_at", now)` (handlers.go:150-153); verified by TestRemindersReapSoftDeletes asserting child counts == 0 |
| Already-soft-deleted not re-stamped | PASS | reap WHERE includes `deleted_at IS NULL` (handlers.go:151); test asserts deleted_at not overwritten (reminder_handlers_test.go:567-571) |
| dryRun handling consistent with pattern | PASS | reminder handlers ignore `dryRun`, matching `CompletedTasks.Reap`/`DeletedTasksRestoreWindow.Reap` which also ignore it |
| Registered in reaper | PASS | wire.go:25-26 adds `Reminders{}`, `DeletedRemindersRestoreWindow{}` |

## Shared Library — shared/go/retention

| Check | Status | Evidence |
|-------|--------|----------|
| New constants added | PASS | category.go: `CatProductivityReminders` (productivity.reminders), `CatProductivityDeletedRemindersRestoreWindow` |
| Defaults populated | PASS | `Defaults` map: 365 / 30 (matches task analogues) |
| Scope kind populated | PASS | `scopeKindOf`: both `ScopeHousehold` (household-scoped per spec) |
| Coverage tests | PASS | category_test.go TestReminderCategories asserts IsKnown, IsHouseholdScoped, defaults, MaxDays (3650 / 365 via restore-window suffix), HouseholdCategories enumeration; TestDefaultsCoverage iterates All() |

## Provider / read-filter correctness

| Check | Status | Evidence |
|-------|--------|----------|
| Soft-deleted hidden from single read | PASS | getByID adds `deleted_at IS NULL` (provider.go:9); test ByIDProvider → ErrRecordNotFound (softdelete_test.go:318) |
| Soft-deleted hidden from list | PASS | getAll adds `Where("deleted_at IS NULL")` (provider.go:16) |
| Soft-deleted excluded from all counts | PASS | countDueNow/countUpcoming/countSnoozed each gained `AND deleted_at IS NULL`; verified by softdelete_test.go:326-336 (seeded past-due row counts 0/0/0) |
| Mutations reject soft-deleted | PASS | update/dismiss/snooze/deleteByID gained `deleted_at IS NULL`; softdelete_test.go:339-352 asserts ErrRecordNotFound on all four |
| Tenant filtering preserved | PASS | filters are appended to provider closures invoked with `p.db.WithContext(p.ctx)` (processor.go:25,29); GORM tenant callback still injects `tenant_id` |

## Findings

### Important
- **IMPORTANT-1 — snooze of a missing/soft-deleted reminder now returns 500, not 404.**
  `snooze()` previously returned `nil` when no row matched; it now returns `gorm.ErrRecordNotFound` when `RowsAffected == 0` (administrator.go:71-73). That error propagates `reminder.Processor.Snooze` → `snooze.Processor.Create` (reminder/snooze/processor.go:32-34) → the snooze handler, which maps only the three validation errors and falls through to `StatusInternalServerError` (reminder/snooze/resource.go:36-44). By contrast the dismissal handler correctly maps `gorm.ErrRecordNotFound` → 404 (reminder/dismissal/resource.go:42-44), and the reminder delete/update handlers map it to 404 (resource.go:135, 162). Recommend adding the same `errors.Is(err, gorm.ErrRecordNotFound)` → 404 branch to reminder/snooze/resource.go for consistency. Note: the *new* behavior is still better than the *old* (old path returned 201 and created a snooze row referencing a non-existent reminder); this is exposing a pre-existing gap, not introducing data corruption.

### Non-Blocking / Observations
- **OBS-1** — `reminder_handlers_test.go:newReminderDB` (line 19) does not call `database.RegisterTenantCallbacks`. This is intentional and consistent with the established `handlers_test.go:newDB` (line 18): retention reap queries run on raw `db` with explicit `tenant_id = ?` predicates (never `db.WithContext(ctx)`), so the callback is irrelevant here. Not a violation.
- **OBS-2** — `Reminders.Reap` uses a single-column `Update("deleted_at", now)` and does not bump `updated_at` on soft-delete. This matches the reaper's row-level intent and the task soft-delete-by-reaper has no analogous touch; acceptable.

## Summary

### Blocking (must fix)
- None. Build green, tests green, no DOM-* hard failures, multi-tenancy/scoping and transaction handling verified.

### Should fix
- IMPORTANT-1: map `gorm.ErrRecordNotFound` → 404 in `internal/reminder/snooze/resource.go` so snooze matches dismissal/delete/update not-found semantics.
