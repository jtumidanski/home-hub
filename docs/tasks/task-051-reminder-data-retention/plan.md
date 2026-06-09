# Reminder Data Retention Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring `reminders` into the shared data-retention framework with a two-stage reaper (soft-delete aged reminders, then hard-delete + cascade after a restore window), configurable from the Data Retention settings page exactly like every other entity.

**Architecture:** Two new household-scoped `Category` constants in `shared/go/retention`; a plain nullable indexed `deleted_at` column on `reminders` (mirroring `tasks`, NOT `gorm.DeletedAt`); two `CategoryHandler` implementations in `productivity-service/internal/retention` registered in `wire.go`; a `deleted_at IS NULL` filter on every reminder read/lookup path; and two `CATEGORY_LABELS` entries in the frontend. Account-service, the reaper loop, metrics, audit rows, dry-run, and policy overrides all work unchanged because they enumerate categories dynamically.

**Tech Stack:** Go 1.x (multi-module via `go.work`), GORM, sqlite (in-memory test DB), testify; React + TypeScript + Vite + vitest for the frontend.

---

## File Structure

**`shared/go/retention/` (module: `shared/go/retention/go.mod`)**
- Modify `category.go` — add 2 `Category` constants + `Defaults` + `scopeKindOf` entries.
- Modify `category_test.go` — assert the two new categories' scope, defaults, and MaxDays.

**`services/productivity-service/` (module: `services/productivity-service/go.mod`)**
- Modify `internal/reminder/entity.go` — add `DeletedAt` field; round-trip through `ToEntity`/`Make`.
- Modify `internal/reminder/model.go` — add `deletedAt` field, `DeletedAt()` accessor, `IsDeleted()` helper.
- Modify `internal/reminder/builder.go` — add `deletedAt` field + `SetDeletedAt`.
- Create `internal/reminder/softdelete_test.go` — model round-trip + read/lookup-path exclusion tests.
- Modify `internal/reminder/provider.go` — `deleted_at IS NULL` on `getByID`, `getAll`, the 3 counts.
- Modify `internal/reminder/administrator.go` — `deleted_at IS NULL` on `update`, `dismiss`, `snooze`, `deleteByID`; add the missing zero-rows check to `snooze`.
- Modify `internal/retention/handlers.go` — add `Reminders` + `DeletedRemindersRestoreWindow` handlers + `cascadeDeleteReminders`.
- Create `internal/retention/reminder_handlers_test.go` — handler unit tests against an in-memory DB.
- Modify `internal/retention/wire.go` — register the two handlers.

**`services/account-service/` (module: `services/account-service/go.mod`)**
- Modify `internal/retention/processor_test.go` — assert both new categories resolve with defaults 365/30. (No production change — auto-enumerated.)

**`frontend/` (module: `frontend/package.json`)**
- Modify `src/pages/DataRetentionPage.tsx` — 2 `CATEGORY_LABELS` entries.

---

## Task 1: Shared retention categories

**Files:**
- Modify: `shared/go/retention/category.go`
- Test: `shared/go/retention/category_test.go`

- [ ] **Step 1: Write the failing test**

Add to `shared/go/retention/category_test.go`:

```go
func TestReminderCategories(t *testing.T) {
	// Known + household-scoped.
	for _, c := range []Category{CatProductivityReminders, CatProductivityDeletedRemindersRestoreWindow} {
		if !c.IsKnown() {
			t.Errorf("%s should be known", c)
		}
		if !c.IsHouseholdScoped() {
			t.Errorf("%s should be household-scoped", c)
		}
	}
	// Defaults.
	if Defaults[CatProductivityReminders] != 365 {
		t.Errorf("reminders default = %d, want 365", Defaults[CatProductivityReminders])
	}
	if Defaults[CatProductivityDeletedRemindersRestoreWindow] != 30 {
		t.Errorf("restore-window default = %d, want 30", Defaults[CatProductivityDeletedRemindersRestoreWindow])
	}
	// MaxDays: primary 3650, restore window capped at 365 by the suffix.
	if CatProductivityReminders.MaxDays() != 3650 {
		t.Errorf("reminders MaxDays = %d, want 3650", CatProductivityReminders.MaxDays())
	}
	if CatProductivityDeletedRemindersRestoreWindow.MaxDays() != 365 {
		t.Errorf("restore-window MaxDays = %d, want 365", CatProductivityDeletedRemindersRestoreWindow.MaxDays())
	}
	// Both appear in the household enumeration.
	found := map[Category]bool{}
	for _, c := range HouseholdCategories() {
		found[c] = true
	}
	if !found[CatProductivityReminders] || !found[CatProductivityDeletedRemindersRestoreWindow] {
		t.Errorf("HouseholdCategories missing reminder categories: %v", HouseholdCategories())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd shared/go/retention && go test ./... -run TestReminderCategories -v`
Expected: FAIL — `undefined: CatProductivityReminders`.

- [ ] **Step 3: Add the constants, defaults, and scope entries**

In `shared/go/retention/category.go`, add to the `const (...)` block (after `CatProductivityDeletedTasksRestoreWindow`):

```go
	CatProductivityReminders                     Category = "productivity.reminders"
	CatProductivityDeletedRemindersRestoreWindow Category = "productivity.deleted_reminders_restore_window"
```

Add to the `Defaults` map:

```go
	CatProductivityReminders:                     365,
	CatProductivityDeletedRemindersRestoreWindow: 30,
```

Add to the `scopeKindOf` map:

```go
	CatProductivityReminders:                     ScopeHousehold,
	CatProductivityDeletedRemindersRestoreWindow: ScopeHousehold,
```

(No other change. `All()`, `HouseholdCategories()`, `Validate`, `MaxDays()`, and the `_restore_window` 365-day cap all derive from these maps.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd shared/go/retention && go test ./...`
Expected: PASS (including the existing `TestDefaultsCoverage`, which now covers the two new categories).

- [ ] **Step 5: Commit**

```bash
git add shared/go/retention/category.go shared/go/retention/category_test.go
git commit -m "feat(retention): add reminder retention categories"
```

---

## Task 2: Reminder soft-delete column plumbing

**Files:**
- Modify: `services/productivity-service/internal/reminder/model.go`
- Modify: `services/productivity-service/internal/reminder/builder.go`
- Modify: `services/productivity-service/internal/reminder/entity.go`
- Test: `services/productivity-service/internal/reminder/softdelete_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `services/productivity-service/internal/reminder/softdelete_test.go`:

```go
package reminder

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeletedAtRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	e := Entity{
		Id:           uuid.New(),
		TenantId:     uuid.New(),
		HouseholdId:  uuid.New(),
		Title:        "x",
		ScheduledFor: now,
		DeletedAt:    &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	m, err := Make(e)
	require.NoError(t, err)
	require.NotNil(t, m.DeletedAt())
	require.True(t, m.IsDeleted())
	require.Equal(t, e, m.ToEntity())

	// Nil case: a live reminder is not deleted.
	e2 := e
	e2.DeletedAt = nil
	m2, err := Make(e2)
	require.NoError(t, err)
	require.Nil(t, m2.DeletedAt())
	require.False(t, m2.IsDeleted())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/productivity-service && go test ./internal/reminder/ -run TestDeletedAtRoundTrip -v`
Expected: FAIL — `Entity has no field DeletedAt` (compile error).

- [ ] **Step 3: Add the `deletedAt` field to the model**

In `internal/reminder/model.go`, add `deletedAt *time.Time` to the `Model` struct (after `lastSnoozedUntil`):

```go
	lastSnoozedUntil *time.Time
	deletedAt        *time.Time
	createdAt        time.Time
```

Add the accessor and helper (after `LastSnoozedUntil()`):

```go
func (m Model) DeletedAt() *time.Time { return m.deletedAt }
func (m Model) IsDeleted() bool        { return m.deletedAt != nil }
```

- [ ] **Step 4: Add the `deletedAt` field to the builder**

In `internal/reminder/builder.go`, add `deletedAt *time.Time` to the `Builder` struct (after `lastSnoozedUntil`):

```go
	lastSnoozedUntil *time.Time
	deletedAt        *time.Time
	createdAt        time.Time
```

Add the setter (after `SetLastSnoozedUntil`):

```go
func (b *Builder) SetDeletedAt(t *time.Time) *Builder { b.deletedAt = t; return b }
```

Carry it through `Build()` (after `lastSnoozedUntil: b.lastSnoozedUntil,`):

```go
		lastSnoozedUntil: b.lastSnoozedUntil,
		deletedAt:        b.deletedAt,
		createdAt:        b.createdAt,
```

- [ ] **Step 5: Add the `DeletedAt` column to the entity and round-trip it**

In `internal/reminder/entity.go`, add the field to `Entity` (after `LastSnoozedUntil`):

```go
	LastSnoozedUntil *time.Time
	DeletedAt        *time.Time `gorm:"index"`
	CreatedAt        time.Time  `gorm:"not null"`
```

Add to `ToEntity()` (in the `LastDismissedAt`/`LastSnoozedUntil` group):

```go
		LastDismissedAt: m.lastDismissedAt, LastSnoozedUntil: m.lastSnoozedUntil,
		DeletedAt: m.deletedAt,
		CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
```

Add to `Make()` (after `SetLastSnoozedUntil(e.LastSnoozedUntil)`):

```go
		SetLastSnoozedUntil(e.LastSnoozedUntil).
		SetDeletedAt(e.DeletedAt).
		SetCreatedAt(e.CreatedAt).
```

(`Migration` stays `db.AutoMigrate(&Entity{})` — the additive nullable indexed column is created automatically. No backfill: existing rows get `deleted_at = NULL`, which correctly means "not soft-deleted".)

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd services/productivity-service && go test ./internal/reminder/ -run TestDeletedAtRoundTrip -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add services/productivity-service/internal/reminder/model.go services/productivity-service/internal/reminder/builder.go services/productivity-service/internal/reminder/entity.go services/productivity-service/internal/reminder/softdelete_test.go
git commit -m "feat(reminder): add deleted_at soft-delete column"
```

---

## Task 3: Exclude soft-deleted reminders from read and lookup paths

**Files:**
- Modify: `services/productivity-service/internal/reminder/provider.go`
- Modify: `services/productivity-service/internal/reminder/administrator.go`
- Test: `services/productivity-service/internal/reminder/softdelete_test.go` (extend)

- [ ] **Step 1: Write the failing tests**

Append to `services/productivity-service/internal/reminder/softdelete_test.go`. (`setupTestDB` and `newTestProcessor` already exist in `processor_test.go` in this package; reuse them. Add these imports to the existing import block of `softdelete_test.go`: `"errors"`, `"github.com/jtumidanski/home-hub/shared/go/database"` is NOT needed here, but `"gorm.io/gorm"` IS.)

Update the import block at the top of `softdelete_test.go` to:

```go
import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)
```

Then append:

```go
// seedDeleted inserts a reminder whose deleted_at is set (soft-deleted) and
// returns its id. Uses the raw entity so it bypasses the read-path filters.
func seedDeleted(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	now := time.Now().UTC()
	id := uuid.New()
	require.NoError(t, db.Create(&Entity{
		Id:           id,
		TenantId:     uuid.New(),
		HouseholdId:  uuid.New(),
		Title:        "ghost",
		ScheduledFor: now.Add(-1 * time.Hour),
		DeletedAt:    &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error)
	return id
}

func TestSoftDeletedHiddenFromReads(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	id := seedDeleted(t, db)

	// getByID via the processor returns not-found.
	_, err := p.ByIDProvider(id)()
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "ByIDProvider err = %v", err)

	// getAll excludes it.
	all, err := p.AllProvider()()
	require.NoError(t, err)
	require.Empty(t, all)

	// All three counts exclude it (the seeded row is past-due, undismissed,
	// unsnoozed — it would count as due-now if not filtered).
	due, err := p.DueNowCount()
	require.NoError(t, err)
	require.Equal(t, int64(0), due)
	up, err := p.UpcomingCount()
	require.NoError(t, err)
	require.Equal(t, int64(0), up)
	sn, err := p.SnoozedCount()
	require.NoError(t, err)
	require.Equal(t, int64(0), sn)
}

func TestSoftDeletedNotFoundOnMutations(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	id := seedDeleted(t, db)

	_, err := p.Update(id, "new", "", time.Now().UTC(), nil)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Update err = %v", err)

	require.True(t, errors.Is(p.Dismiss(id), gorm.ErrRecordNotFound), "Dismiss")

	_, err = p.Snooze(id, 10)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Snooze err = %v", err)

	require.True(t, errors.Is(p.Delete(id), gorm.ErrRecordNotFound), "Delete")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/productivity-service && go test ./internal/reminder/ -run 'TestSoftDeleted' -v`
Expected: FAIL — soft-deleted rows are still returned/mutated (e.g. `AllProvider` returns 1 row, `Dismiss` succeeds).

- [ ] **Step 3: Add `deleted_at IS NULL` to the read paths**

In `internal/reminder/provider.go`:

`getByID`:

```go
func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND deleted_at IS NULL", id)
	})
}
```

`getAll`:

```go
func getAll() database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted_at IS NULL").
			Order("last_dismissed_at IS NULL DESC").Order("scheduled_for ASC")
	})
}
```

`countDueNow` — append `AND deleted_at IS NULL` to the `Where`:

```go
		Where("scheduled_for <= CURRENT_TIMESTAMP AND last_dismissed_at IS NULL AND (last_snoozed_until IS NULL OR last_snoozed_until <= CURRENT_TIMESTAMP) AND deleted_at IS NULL").
```

`countUpcoming`:

```go
		Where("scheduled_for > CURRENT_TIMESTAMP AND last_dismissed_at IS NULL AND deleted_at IS NULL").
```

`countSnoozed`:

```go
		Where("last_snoozed_until > CURRENT_TIMESTAMP AND last_dismissed_at IS NULL AND deleted_at IS NULL").
```

- [ ] **Step 4: Add `deleted_at IS NULL` to the lookup/mutation paths**

In `internal/reminder/administrator.go`:

`update` — gate the `First`:

```go
	if err := db.Where("id = ? AND deleted_at IS NULL", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
```

`dismiss` — gate the `Where` (the existing `RowsAffected == 0 → ErrRecordNotFound` then fires for a soft-deleted row):

```go
	result := db.Model(&Entity{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
```

`snooze` — gate the `Where` AND add the missing zero-rows check (parity with `dismiss`):

```go
func snooze(db *gorm.DB, id uuid.UUID, snoozedUntil time.Time) error {
	now := time.Now().UTC()
	result := db.Model(&Entity{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"last_snoozed_until": snoozedUntil,
		"updated_at":         now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
```

`deleteByID` — gate the delete and report not-found when nothing matched (so a hidden reminder is "not found" to a user delete rather than being hard-deleted out from under the restore window):

```go
func deleteByID(db *gorm.DB, id uuid.UUID) error {
	result := db.Where("id = ? AND deleted_at IS NULL", id).Delete(&Entity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd services/productivity-service && go test ./internal/reminder/ -v`
Expected: PASS (new soft-delete tests plus all existing reminder tests — `TestCreate`, `TestDismiss`, `TestSnooze`, `TestIsActive`, builder tests).

- [ ] **Step 6: Commit**

```bash
git add services/productivity-service/internal/reminder/provider.go services/productivity-service/internal/reminder/administrator.go services/productivity-service/internal/reminder/softdelete_test.go
git commit -m "feat(reminder): exclude soft-deleted reminders from read and lookup paths"
```

---

## Task 4: Primary reaper — soft-delete aged reminders

**Files:**
- Modify: `services/productivity-service/internal/retention/handlers.go`
- Test: `services/productivity-service/internal/retention/reminder_handlers_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `services/productivity-service/internal/retention/reminder_handlers_test.go`:

```go
package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder/dismissal"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder/snooze"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newReminderDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&reminder.Entity{}, &dismissal.Entity{}, &snooze.Entity{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRemindersReapSoftDeletes(t *testing.T) {
	db := newReminderDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now().UTC()
	at := func(d time.Duration) time.Time { return now.Add(d) }
	ptr := func(tm time.Time) *time.Time { return &tm }

	// Helper to insert a reminder with explicit scheduled/dismissed/deleted state.
	mk := func(scheduledFor time.Time, dismissedAt, deletedAt *time.Time) uuid.UUID {
		id := uuid.New()
		db.Create(&reminder.Entity{
			Id: id, TenantId: tenantID, HouseholdId: householdID,
			Title: "x", ScheduledFor: scheduledFor,
			LastDismissedAt: dismissedAt, DeletedAt: deletedAt,
			CreatedAt: now, UpdatedAt: now,
		})
		return id
	}

	// 1) Dismissed > window ago  -> soft-deleted.
	dismissedOld := mk(at(-10*24*time.Hour), ptr(at(-400*24*time.Hour)), nil)
	// 2) Scheduled > window ago, never dismissed -> soft-deleted.
	scheduledOld := mk(at(-400*24*time.Hour), nil, nil)
	// 3) Recently dismissed, recently scheduled -> untouched.
	fresh := mk(at(-1*24*time.Hour), ptr(at(-1*24*time.Hour)), nil)
	// 4) Already soft-deleted -> untouched (deleted_at not overwritten).
	alreadyDeleted := mk(at(-400*24*time.Hour), nil, ptr(at(-5*24*time.Hour)))

	res, err := Reminders{}.Reap(context.Background(), db, sr.Scope{
		TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID,
	}, 365, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 2 || res.Deleted != 2 {
		t.Errorf("got scanned=%d deleted=%d, want 2/2", res.Scanned, res.Deleted)
	}

	isDeleted := func(id uuid.UUID) bool {
		var e reminder.Entity
		db.Where("id = ?", id).First(&e)
		return e.DeletedAt != nil
	}
	if !isDeleted(dismissedOld) || !isDeleted(scheduledOld) {
		t.Error("aged reminders should be soft-deleted")
	}
	if isDeleted(fresh) {
		t.Error("fresh reminder should not be soft-deleted")
	}
	// alreadyDeleted keeps its original deleted_at (was -5d, not overwritten to now).
	var ad reminder.Entity
	db.Where("id = ?", alreadyDeleted).First(&ad)
	if ad.DeletedAt == nil || ad.DeletedAt.After(now.Add(-1*time.Hour)) {
		t.Error("already-soft-deleted reminder's deleted_at must not be overwritten")
	}

	// Child tables are never touched by the primary stage.
	var dismissals, snoozes int64
	db.Model(&dismissal.Entity{}).Count(&dismissals)
	db.Model(&snooze.Entity{}).Count(&snoozes)
	if dismissals != 0 || snoozes != 0 {
		t.Errorf("primary reap must not touch child tables; got d=%d s=%d", dismissals, snoozes)
	}
}

func TestRemindersDiscoverScopes(t *testing.T) {
	db := newReminderDB(t)
	tenantID := uuid.New()
	householdA := uuid.New()
	householdB := uuid.New()
	now := time.Now().UTC()
	for _, hh := range []uuid.UUID{householdA, householdA, householdB} {
		db.Create(&reminder.Entity{
			Id: uuid.New(), TenantId: tenantID, HouseholdId: hh,
			Title: "x", ScheduledFor: now, CreatedAt: now, UpdatedAt: now,
		})
	}
	scopes, err := Reminders{}.DiscoverScopes(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(scopes) != 2 {
		t.Errorf("distinct scopes = %d, want 2", len(scopes))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/productivity-service && go test ./internal/retention/ -run TestReminders -v`
Expected: FAIL — `undefined: Reminders`.

- [ ] **Step 3: Implement the `Reminders` handler**

In `internal/retention/handlers.go`, add the `reminder` import to the import block:

```go
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder/dismissal"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder/snooze"
```

Append the handler (after `cascadeDeleteTasks`, before `AuditTrim`):

```go
// Reminders soft-deletes reminders that are dismissed-aged or scheduled-past
// beyond the configured window. It is the reaper-driven soft-delete stage
// (reminders have no user-facing trash lifecycle); the restore-window handler
// hard-deletes them later. It touches no child tables.
type Reminders struct{}

func (Reminders) Category() sr.Category { return sr.CatProductivityReminders }

func (Reminders) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	type row struct {
		TenantId    uuid.UUID
		HouseholdId uuid.UUID
	}
	var rows []row
	if err := db.Table("reminders").
		Select("DISTINCT tenant_id, household_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]sr.Scope, 0, len(rows))
	for _, r := range rows {
		out = append(out, sr.Scope{TenantId: r.TenantId, Kind: sr.ScopeHousehold, ScopeId: r.HouseholdId})
	}
	return out, nil
}

func (Reminders) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(days) * 24 * time.Hour)

	r := tx.Table("reminders").
		Where("tenant_id = ? AND household_id = ? AND deleted_at IS NULL AND ((last_dismissed_at IS NOT NULL AND last_dismissed_at < ?) OR (scheduled_for < ?))",
			scope.TenantId, scope.ScopeId, cutoff, cutoff).
		Update("deleted_at", now)
	if r.Error != nil {
		return sr.ReapResult{}, r.Error
	}
	return sr.ReapResult{Scanned: int(r.RowsAffected), Deleted: int(r.RowsAffected)}, nil
}
```

> Note: like the task handlers, this does NOT branch on `dryRun`. `Reaper.RunOne` rolls the transaction back via `errDryRunRollback` when `dryRun` is true, so the UPDATE executes (yielding an accurate `RowsAffected` count) and is then discarded.

(The `dismissal`/`snooze` imports added in this step are consumed by Task 5's cascade — adding them now keeps the single import edit together. If your editor flags them as unused before Task 5, complete Task 5 in the same working session.)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd services/productivity-service && go test ./internal/retention/ -run TestReminders -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add services/productivity-service/internal/retention/handlers.go services/productivity-service/internal/retention/reminder_handlers_test.go
git commit -m "feat(retention): add primary reminder soft-delete reaper"
```

---

## Task 5: Restore-window reaper — hard-delete with cascade

**Files:**
- Modify: `services/productivity-service/internal/retention/handlers.go`
- Test: `services/productivity-service/internal/retention/reminder_handlers_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Append to `services/productivity-service/internal/retention/reminder_handlers_test.go`:

```go
func TestDeletedRemindersRestoreWindowCascade(t *testing.T) {
	db := newReminderDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now().UTC()
	at := func(d time.Duration) *time.Time { v := now.Add(d); return &v }

	mk := func(deletedAt *time.Time, withChildren bool) uuid.UUID {
		id := uuid.New()
		db.Create(&reminder.Entity{
			Id: id, TenantId: tenantID, HouseholdId: householdID,
			Title: "x", ScheduledFor: now.Add(-500 * 24 * time.Hour),
			DeletedAt: deletedAt, CreatedAt: now, UpdatedAt: now,
		})
		if withChildren {
			db.Create(&dismissal.Entity{
				Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
				ReminderId: id, CreatedByUserId: uuid.New(), CreatedAt: now,
			})
			db.Create(&snooze.Entity{
				Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
				ReminderId: id, DurationMinutes: 10, SnoozedUntil: now, CreatedByUserId: uuid.New(), CreatedAt: now,
			})
		}
		return id
	}

	expired := mk(at(-31*24*time.Hour), true) // past 30d window, has 1 dismissal + 1 snooze
	mk(at(-29*24*time.Hour), false)            // inside window — keep
	mk(nil, false)                             // not soft-deleted — keep

	res, err := DeletedRemindersRestoreWindow{}.Reap(context.Background(), db, sr.Scope{
		TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID,
	}, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 1 {
		t.Errorf("scanned = %d, want 1", res.Scanned)
	}
	// 1 reminder + 1 dismissal + 1 snooze = 3 rows removed.
	if res.Deleted != 3 {
		t.Errorf("deleted = %d, want 3", res.Deleted)
	}

	var reminders, dismissals, snoozes int64
	db.Model(&reminder.Entity{}).Count(&reminders)
	db.Model(&dismissal.Entity{}).Count(&dismissals)
	db.Model(&snooze.Entity{}).Count(&snoozes)
	if reminders != 2 {
		t.Errorf("remaining reminders = %d, want 2", reminders)
	}
	if dismissals != 0 || snoozes != 0 {
		t.Errorf("children of expired reminder not cascaded: d=%d s=%d", dismissals, snoozes)
	}

	// The expired reminder is gone; the inside-window one is not.
	var gone int64
	db.Model(&reminder.Entity{}).Where("id = ?", expired).Count(&gone)
	if gone != 0 {
		t.Error("expired reminder should have been hard-deleted")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/productivity-service && go test ./internal/retention/ -run TestDeletedRemindersRestoreWindow -v`
Expected: FAIL — `undefined: DeletedRemindersRestoreWindow`.

- [ ] **Step 3: Implement the restore-window handler and cascade**

In `internal/retention/handlers.go`, append (after the `Reminders` handler):

```go
// DeletedRemindersRestoreWindow hard-deletes reminders whose deleted_at is
// older than the restore window, cascading to reminder_dismissals and
// reminder_snoozes. Mirrors DeletedTasksRestoreWindow + cascadeDeleteTasks.
type DeletedRemindersRestoreWindow struct{}

func (DeletedRemindersRestoreWindow) Category() sr.Category {
	return sr.CatProductivityDeletedRemindersRestoreWindow
}

func (DeletedRemindersRestoreWindow) DiscoverScopes(ctx context.Context, db *gorm.DB) ([]sr.Scope, error) {
	return Reminders{}.DiscoverScopes(ctx, db)
}

func (DeletedRemindersRestoreWindow) Reap(ctx context.Context, tx *gorm.DB, scope sr.Scope, days int, dryRun bool) (sr.ReapResult, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)

	var ids []string
	if err := tx.Table("reminders").
		Where("tenant_id = ? AND household_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", scope.TenantId, scope.ScopeId, cutoff).
		Pluck("id", &ids).Error; err != nil {
		return sr.ReapResult{}, err
	}
	if len(ids) == 0 {
		return sr.ReapResult{}, nil
	}

	deleted, err := cascadeDeleteReminders(tx, ids)
	if err != nil {
		return sr.ReapResult{}, err
	}
	return sr.ReapResult{Scanned: len(ids), Deleted: deleted}, nil
}

// cascadeDeleteReminders removes the listed reminder ids and their dependent
// rows (reminder_snoozes, reminder_dismissals) inside the supplied tx, children
// first. Returns the total number of rows removed across all three tables.
func cascadeDeleteReminders(tx *gorm.DB, ids []string) (int, error) {
	var total int

	r := tx.Where("reminder_id IN ?", ids).Delete(&snooze.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("reminder_id IN ?", ids).Delete(&dismissal.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	r = tx.Where("id IN ?", ids).Delete(&reminder.Entity{})
	if r.Error != nil {
		return 0, r.Error
	}
	total += int(r.RowsAffected)

	return total, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd services/productivity-service && go test ./internal/retention/ -v`
Expected: PASS (all reminder handler tests plus the pre-existing task handler tests).

- [ ] **Step 5: Commit**

```bash
git add services/productivity-service/internal/retention/handlers.go services/productivity-service/internal/retention/reminder_handlers_test.go
git commit -m "feat(retention): add reminder restore-window reaper with cascade"
```

---

## Task 6: Register the reminder handlers in the reaper

**Files:**
- Modify: `services/productivity-service/internal/retention/wire.go`

- [ ] **Step 1: Register both handlers in `Setup`**

In `internal/retention/wire.go`, add the two handlers to the `sr.New(...)` call:

```go
	reaper := sr.New("productivity-service", db, pc, metrics, l,
		CompletedTasks{},
		DeletedTasksRestoreWindow{},
		Reminders{},
		DeletedRemindersRestoreWindow{},
		AuditTrim{},
	)
```

- [ ] **Step 2: Build the service to verify wiring compiles**

Run: `cd services/productivity-service && go build ./...`
Expected: success, no output.

- [ ] **Step 3: Run the full retention test suite**

Run: `cd services/productivity-service && go test ./internal/retention/ ./internal/reminder/ -v`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add services/productivity-service/internal/retention/wire.go
git commit -m "feat(retention): register reminder reapers in the productivity loop"
```

---

## Task 7: Account-service enumeration test

**Files:**
- Modify: `services/account-service/internal/retention/processor_test.go`

(No production change: `ResolveAll` merges DB overrides over `HouseholdCategories()`, which now includes the two new categories automatically. This task adds a regression test proving they resolve with the shipped defaults.)

- [ ] **Step 1: Write the test**

Append to `services/account-service/internal/retention/processor_test.go`:

```go
func TestResolveAllIncludesReminderCategories(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	resolved, err := p.ResolveAll(tenantID, householdID, userID)
	if err != nil {
		t.Fatal(err)
	}

	rem := resolved.Household.Values[sharedretention.CatProductivityReminders]
	if rem.Days != 365 || rem.Source != "default" {
		t.Errorf("reminders = %+v, want 365/default", rem)
	}
	rw := resolved.Household.Values[sharedretention.CatProductivityDeletedRemindersRestoreWindow]
	if rw.Days != 30 || rw.Source != "default" {
		t.Errorf("restore window = %+v, want 30/default", rw)
	}
}
```

- [ ] **Step 2: Run the test to verify it passes**

Run: `cd services/account-service && go test ./internal/retention/ -run TestResolveAllIncludesReminderCategories -v`
Expected: PASS (the categories are enumerated automatically from the shared registry).

- [ ] **Step 3: Run the full account-service retention suite**

Run: `cd services/account-service && go test ./internal/retention/ -v`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add services/account-service/internal/retention/processor_test.go
git commit -m "test(retention): assert reminder categories resolve in account-service"
```

---

## Task 8: Frontend settings labels

**Files:**
- Modify: `frontend/src/pages/DataRetentionPage.tsx`

- [ ] **Step 1: Add the two category labels**

In `frontend/src/pages/DataRetentionPage.tsx`, add to the `CATEGORY_LABELS` map (after the existing `"productivity.deleted_tasks_restore_window"` entry):

```ts
  "productivity.reminders": "Reminders",
  "productivity.deleted_reminders_restore_window": "Deleted reminders (restore window)",
```

(`categoryMax` already returns 365 for any `*_restore_window` key and 3650 otherwise; rows render and sort generically from the API response — no other change.)

- [ ] **Step 2: Type-check the frontend**

Run: `cd frontend && npm run build`
Expected: `tsc -b` and `vite build` succeed with no type errors.

- [ ] **Step 3: Run the frontend test suite**

Run: `cd frontend && npm run test`
Expected: PASS (no DataRetentionPage-specific test exists; this confirms the label change breaks nothing in the existing suite).

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/DataRetentionPage.tsx
git commit -m "feat(frontend): label reminder retention categories on settings page"
```

---

## Task 9: Full verification (builds, tests, Docker)

**Files:** none (verification only).

- [ ] **Step 1: Build and test all three Go modules**

Run:

```bash
cd shared/go/retention && go build ./... && go test ./...
cd ../../../services/productivity-service && go build ./... && go test ./...
cd ../account-service && go build ./... && go test ./...
```

Expected: all build and all tests PASS for each module.

- [ ] **Step 2: Frontend build + tests**

Run: `cd frontend && npm run build && npm run test`
Expected: type-check, build, and tests all PASS.

- [ ] **Step 3: Docker builds for the two affected services (shared library changed)**

Per `CLAUDE.md` ("Always verify Docker builds when changing shared libraries"), build the productivity-service and account-service images. From the repo root:

```bash
docker build -f services/productivity-service/Dockerfile -t hh-productivity-verify .
docker build -f services/account-service/Dockerfile -t hh-account-verify .
```

Expected: both images build successfully. (If the Dockerfiles expect a different build context or build args, consult `scripts/local-up.sh` for the canonical invocation and use that context instead. Do NOT mark this step complete until both images build.)

- [ ] **Step 4: Final acceptance-criteria sweep**

Confirm against `prd.md` §10 and `design.md`:
- Two categories exist with `Defaults` 365/30 and `scopeKindOf` household; restore-window `MaxDays()` is 365, primary 3650. (Task 1)
- `reminders` has a nullable indexed `deleted_at`; the model round-trips it. (Task 2)
- `Reminders` reaper soft-deletes only dismissed-aged OR scheduled-past rows; never touches already-deleted rows or child tables. (Task 4)
- `DeletedRemindersRestoreWindow` hard-deletes `deleted_at < cutoff` and cascades to both child tables atomically. (Task 5)
- Both handlers are registered in `wire.go`. (Task 6)
- No read/lookup path returns a soft-deleted reminder. (Task 3)
- Account-service resolves both categories with defaults. (Task 7)
- Frontend shows "Reminders" and "Deleted reminders (restore window)". (Task 8)

- [ ] **Step 5: Commit any final verification notes (if applicable)**

If steps 1–3 required no fixes, there is nothing to commit here. If a fix was needed, commit it with a descriptive message and re-run the relevant build/test before marking complete.

---

## Notes for the executor

- **Multi-module workspace:** the repo uses `go.work`; `shared/go/retention`, `services/productivity-service`, and `services/account-service` are separate modules. `cd` into each module directory to run its `go build`/`go test`. The workspace makes the productivity-service and account-service builds pick up the local `shared/go/retention` change automatically.
- **`dryRun` is intentionally ignored by the handlers** — the shared `Reaper.RunOne` rolls the transaction back when `dryRun` is true. This matches the existing task handlers; do not add a `dryRun` branch.
- **Plain `deleted_at`, not `gorm.DeletedAt`** — reads filter `deleted_at IS NULL` explicitly, mirroring the task implementation. Do not switch to GORM's automatic soft-delete.
- **Run tests, not just builds, before claiming completion** (project rule). Each task already pairs a build with its test run.
