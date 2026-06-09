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

	dismissedOld := mk(at(-10*24*time.Hour), ptr(at(-400*24*time.Hour)), nil)
	scheduledOld := mk(at(-400*24*time.Hour), nil, nil)
	fresh := mk(at(-1*24*time.Hour), ptr(at(-1*24*time.Hour)), nil)
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
	var ad reminder.Entity
	db.Where("id = ?", alreadyDeleted).First(&ad)
	if ad.DeletedAt == nil || ad.DeletedAt.After(now.Add(-1*time.Hour)) {
		t.Error("already-soft-deleted reminder's deleted_at must not be overwritten")
	}

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
