# Reminder Data Retention — Context

Companion to `plan.md`. Captures the verified ground truth, key decisions, and
dependencies an executor needs without re-reading the whole codebase. All
file:line references were read directly from this worktree.

## Goal in one line

Add `reminders` to the shared retention framework using the same two-stage
pattern as tasks: a primary reaper that **soft-deletes** aged reminders, then a
restore-window reaper that **hard-deletes + cascades** to `reminder_dismissals`
and `reminder_snoozes` after a grace period.

## Key files & responsibilities

### Shared (`shared/go/retention`, own go.mod)
- `category.go` — `Category` enum + `Defaults` + `scopeKindOf` maps. `All()`,
  `HouseholdCategories()`, `Validate`, `MaxDays()` all derive from these two
  maps. The `_restore_window` suffix auto-caps `MaxDays()` at 365 (`category.go:101-106`).
  **Adding a category to both maps is the entire wiring** — enumeration,
  validation, account-service resolution, and the productivity internal
  endpoints pick it up for free.
- `category_test.go` — has `TestDefaultsCoverage` which already asserts every
  `All()` category is present in both maps, so the new categories are covered
  the moment they're added.
- `reaper.go` (not edited) — `CategoryHandler` interface
  (`Category()`, `DiscoverScopes(ctx, db)`, `Reap(ctx, tx, scope, days, dryRun)`);
  `Scope{TenantId, Kind, ScopeId}`; `ReapResult{Scanned, Deleted}`.
  **`Reaper.RunOne` rolls the tx back via `errDryRunRollback` when `dryRun` is
  true** — handlers do not branch on `dryRun`.

### productivity-service (`services/productivity-service`, own go.mod)
- `internal/retention/handlers.go` — the template. `CompletedTasks` (primary,
  hard-deletes) and `DeletedTasksRestoreWindow` (restore window) +
  `cascadeDeleteTasks` (children-first delete summing `RowsAffected`).
  **Neither task handler reads `dryRun`.** New reminder handlers go here.
- `internal/retention/handlers_test.go` — `newDB(t)` opens in-memory sqlite and
  AutoMigrates the task entities. The new reminder handler test
  (`reminder_handlers_test.go`) adds its own `newReminderDB(t)` migrating
  `reminder.Entity`, `dismissal.Entity`, `snooze.Entity`. These test DBs do NOT
  register tenant callbacks; queries pass `tenant_id`/`household_id` explicitly.
- `internal/retention/wire.go` — `Setup` registers handlers via
  `sr.New("productivity-service", db, pc, metrics, l, CompletedTasks{},
  DeletedTasksRestoreWindow{}, AuditTrim{})`. Add the two reminder handlers here.
- `internal/reminder/entity.go` — `Entity` (table `reminders`); `Migration` is
  `AutoMigrate(&Entity{})`; `ToEntity`/`Make` round-trip. **No `deleted_at`
  today.**
- `internal/reminder/model.go` / `builder.go` — immutable model + builder; no
  deleted plumbing today. `IsActive()` lives on the model.
- `internal/reminder/provider.go` — reads: `getByID` (l.9), `getAll` (l.15),
  `countDueNow` (l.21), `countUpcoming` (l.29), `countSnoozed` (l.37). None
  filter `deleted_at` today.
- `internal/reminder/administrator.go` — mutations: `update` (l.29, `First`),
  `dismiss` (l.45, already returns `ErrRecordNotFound` on 0 rows), `snooze`
  (l.60, **no** zero-rows check today), `deleteByID` (l.68, hard delete, no
  zero-rows check). All keyed by `id` only.
- `internal/reminder/processor.go` — wraps provider/administrator; `summary/`
  reaches reminders ONLY through `DueNowCount/UpcomingCount/SnoozedCount`, so
  filtering the three count funcs covers all summary/dashboard surfaces.
- `internal/reminder/processor_test.go` — defines `setupTestDB(t)` (registers
  tenant callbacks, AutoMigrates `reminder.Entity`) and `newTestProcessor(t, db)`.
  Reuse these from the new `softdelete_test.go` (same package).
- `internal/reminder/dismissal/entity.go` — `reminder_dismissals`, FK
  `ReminderId`. `internal/reminder/snooze/entity.go` — `reminder_snoozes`, FK
  `ReminderId`. No DB-level cascade; deletion is explicit Go (like tasks).

### Task plumbing template (mirror exactly)
- `task/model.go:21,37,40` — `deletedAt *time.Time`, `DeletedAt()`, `IsDeleted()`.
- `task/builder.go:26,46,66` — `deletedAt` field, `SetDeletedAt`, carried in `Build()`.
- `task/entity.go:22,38,55` — `DeletedAt *time.Time \`gorm:"index"\``, in
  `ToEntity`, `SetDeletedAt` in `Make`.
- `task/provider.go:20,28,...` — reads filter `deleted_at IS NULL` manually.

### account-service (`services/account-service`, own go.mod)
- `internal/retention/processor.go` `ResolveAll` merges DB overrides over
  `HouseholdCategories()`/`UserCategories()` — **no production change**; new
  categories appear automatically.
- `internal/retention/processor_test.go` — `setupTestDB` + existing
  `TestResolveAllUsesDefaults` show the assertion pattern (`resolved.Household.
  Values[cat].Days` / `.Source`). Add a reminder-categories test here.

### frontend (`frontend`, vite + vitest)
- `src/pages/DataRetentionPage.tsx:37` — `CATEGORY_LABELS` map; `categoryMax`
  (l.52) already handles the 365 cap via the `_restore_window` suffix. Rows
  render/sort generically. Only two label entries to add.
- No DataRetentionPage test exists; `npm run build` (tsc) + `npm run test`
  (vitest) are the gates.

## New categories

| Constant | String | Default | Scope | MaxDays |
| --- | --- | --- | --- | --- |
| `CatProductivityReminders` | `productivity.reminders` | 365 | household | 3650 |
| `CatProductivityDeletedRemindersRestoreWindow` | `productivity.deleted_reminders_restore_window` | 30 | household | 365 (auto via suffix) |

## Reap criteria

- **Primary (`Reminders`)** — single bulk `UPDATE reminders SET deleted_at = now`
  where `deleted_at IS NULL AND ((last_dismissed_at IS NOT NULL AND
  last_dismissed_at < cutoff) OR (scheduled_for < cutoff))`, scoped by
  tenant+household. `Scanned = Deleted = RowsAffected`. No child-table writes.
- **Restore window (`DeletedRemindersRestoreWindow`)** — pluck ids where
  `deleted_at IS NOT NULL AND deleted_at < cutoff`, then
  `cascadeDeleteReminders` (snoozes → dismissals → reminders, summing
  `RowsAffected`). `Scanned = len(ids)`, `Deleted = total across 3 tables`.

`cutoff = time.Now().UTC().Add(-days*24h)`.

## Key decisions (from design.md §5)

1. **`dryRun` honored by framework, not handler** — mirror tasks; do not branch.
2. **Primary stage soft-deletes only; cascade only in restore stage** — a
   soft-deleted reminder must stay fully recoverable (incl. dismiss/snooze
   history) during the restore window.
3. **Plain `*time.Time` + manual filtering, NOT `gorm.DeletedAt`** — identical
   to tasks; keeps reaper queries explicit.
4. **Explicit Go cascade, not DB `ON DELETE CASCADE`** — consistent with
   `cascadeDeleteTasks`.
5. **Single bulk UPDATE for the primary stage** (no pluck-then-update) — it
   touches no child table, so `RowsAffected` is an exact count.

## Read/lookup-path changes (9 sites)

`provider.go`: `getByID`, `getAll`, `countDueNow`, `countUpcoming`,
`countSnoozed` — add `deleted_at IS NULL`.
`administrator.go`: `update` (`First` Where), `dismiss` (Where; existing
zero-rows check fires), `snooze` (Where **+ add** zero-rows→`ErrRecordNotFound`),
`deleteByID` (Where **+ add** zero-rows→`ErrRecordNotFound`).

## PRD open questions — resolved

1. Restore-window default = **30 days** (matches the four existing
   `*_restore_window` categories).
2. Summary/cross-entity reads — only via the three `count*` funcs; no direct
   `reminders` query in `summary/`; no reads outside productivity-service.
3. Dismiss/snooze on a soft-deleted reminder = **treat as not found**.

## Dependencies & ordering

- Task 1 (shared categories) must land before Tasks 4–7 (they reference the
  constants and the resolved defaults).
- Task 2 (the `deleted_at` column) must land before Task 3 (read-path filters
  reference the column) and Tasks 4–5 (reaper queries reference it).
- The `dismissal`/`snooze` imports added to `handlers.go` in Task 4 are consumed
  by Task 5's cascade — do Tasks 4 and 5 in the same session to avoid a transient
  unused-import.
- Tasks 7 (account-service) and 8 (frontend) are independent of each other; both
  depend only on Task 1.

## Verification gates

- Per-module `go build ./... && go test ./...` for `shared/go/retention`,
  `services/productivity-service`, `services/account-service`.
- Frontend `npm run build` (tsc) + `npm run test` (vitest).
- Docker builds for productivity-service and account-service (shared library
  changed — CLAUDE.md rule). Consult `scripts/local-up.sh` for the canonical
  build context if the plain `docker build` invocation needs adjustment.
- **Run tests, not just builds** (project rule) before claiming completion.
