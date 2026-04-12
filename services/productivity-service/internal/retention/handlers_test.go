package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task/restoration"
	sr "github.com/jtumidanski/home-hub/shared/go/retention"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&task.Entity{}, &restoration.Entity{}, &sr.RunEntity{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCompletedTasksReap(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	old := time.Now().Add(-400 * 24 * time.Hour)
	young := time.Now().Add(-10 * 24 * time.Hour)

	mk := func(completedAt time.Time) uuid.UUID {
		id := uuid.New()
		db.Create(&task.Entity{
			Id: id, TenantId: tenantID, HouseholdId: householdID,
			Title: "x", Status: "done",
			CompletedAt: &completedAt,
			CreatedAt:   time.Now(), UpdatedAt: time.Now(),
		})
		return id
	}
	oldID := mk(old)
	mk(young)

	// Cascade target.
	db.Create(&restoration.Entity{
		Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
		TaskId: oldID, CreatedByUserId: uuid.New(), CreatedAt: time.Now(),
	})

	res, err := CompletedTasks{}.Reap(context.Background(), db, sr.Scope{
		TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID,
	}, 365, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 1 {
		t.Errorf("scanned = %d, want 1", res.Scanned)
	}
	// 1 task + 1 restoration = 2 deletes
	if res.Deleted != 2 {
		t.Errorf("deleted = %d, want 2", res.Deleted)
	}

	var remaining int64
	db.Model(&task.Entity{}).Count(&remaining)
	if remaining != 1 {
		t.Errorf("remaining tasks = %d, want 1", remaining)
	}
	var rcount int64
	db.Model(&restoration.Entity{}).Count(&rcount)
	if rcount != 0 {
		t.Errorf("remaining restorations = %d, want 0", rcount)
	}
}

func TestDeletedTasksRestoreWindowBoundary(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now()

	at := func(d time.Duration) *time.Time { v := now.Add(d); return &v }

	mk := func(deletedAt *time.Time) uuid.UUID {
		id := uuid.New()
		db.Create(&task.Entity{
			Id: id, TenantId: tenantID, HouseholdId: householdID,
			Title: "x", Status: "pending",
			DeletedAt: deletedAt,
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		})
		return id
	}
	mk(at(-31 * 24 * time.Hour)) // past 30-day window
	mk(at(-29 * 24 * time.Hour)) // inside window — should not be reaped
	mk(nil)                      // not deleted at all

	res, err := DeletedTasksRestoreWindow{}.Reap(context.Background(), db, sr.Scope{
		TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID,
	}, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 1 {
		t.Errorf("scanned = %d, want 1", res.Scanned)
	}
	var remaining int64
	db.Model(&task.Entity{}).Count(&remaining)
	if remaining != 2 {
		t.Errorf("remaining = %d, want 2", remaining)
	}
}
