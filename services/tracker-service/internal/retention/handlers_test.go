package retention

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
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
	if err := db.AutoMigrate(
		&trackingitem.Entity{},
		&entry.Entity{},
		&sr.RunEntity{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestEntriesReapBoundary(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	itemID := uuid.New()

	val, _ := json.Marshal(3)
	db.Create(&trackingitem.Entity{
		Id: itemID, TenantId: tenantID, UserId: userID,
		Name: "Mood", ScaleType: "numeric", ScaleConfig: val,
		Color: "#ff0000", CreatedAt: now, UpdatedAt: now,
	})

	old := now.AddDate(0, 0, -800)
	recent := now.AddDate(0, 0, -10)
	db.Create(&entry.Entity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		TrackingItemId: itemID, Date: old, Value: val,
		CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&entry.Entity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		TrackingItemId: itemID, Date: recent, Value: val,
		CreatedAt: now, UpdatedAt: now,
	})

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeUser, ScopeId: userID}
	res, err := Entries{}.Reap(context.Background(), db, scope, 730, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 1 {
		t.Errorf("deleted = %d, want 1", res.Deleted)
	}

	// The tracking_item should still exist (no upward cascade).
	var itemCount int64
	db.Model(&trackingitem.Entity{}).Count(&itemCount)
	if itemCount != 1 {
		t.Errorf("remaining items = %d, want 1", itemCount)
	}
}

func TestDeletedItemsRestoreWindowCascade(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	past := now.Add(-31 * 24 * time.Hour)

	val, _ := json.Marshal(5)

	// Soft-deleted item past the 30-day window.
	oldItemID := uuid.New()
	db.Create(&trackingitem.Entity{
		Id: oldItemID, TenantId: tenantID, UserId: userID,
		Name: "Old", ScaleType: "numeric", ScaleConfig: val,
		Color: "#000", CreatedAt: now, UpdatedAt: now, DeletedAt: &past,
	})
	// Entries for the deleted item should be cascaded.
	db.Create(&entry.Entity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		TrackingItemId: oldItemID, Date: now.AddDate(0, 0, -5), Value: val,
		CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&entry.Entity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		TrackingItemId: oldItemID, Date: now.AddDate(0, 0, -6), Value: val,
		CreatedAt: now, UpdatedAt: now,
	})

	// Live item should survive.
	liveItemID := uuid.New()
	db.Create(&trackingitem.Entity{
		Id: liveItemID, TenantId: tenantID, UserId: userID,
		Name: "Live", ScaleType: "numeric", ScaleConfig: val,
		Color: "#fff", CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&entry.Entity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		TrackingItemId: liveItemID, Date: now, Value: val,
		CreatedAt: now, UpdatedAt: now,
	})

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeUser, ScopeId: userID}
	res, err := DeletedItemsRestoreWindow{}.Reap(context.Background(), db, scope, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 1 {
		t.Errorf("scanned = %d, want 1", res.Scanned)
	}
	// 2 entries + 1 item = 3
	if res.Deleted != 3 {
		t.Errorf("deleted = %d, want 3", res.Deleted)
	}

	var itemCount int64
	db.Model(&trackingitem.Entity{}).Count(&itemCount)
	if itemCount != 1 {
		t.Errorf("remaining items = %d, want 1", itemCount)
	}
	var entryCount int64
	db.Model(&entry.Entity{}).Count(&entryCount)
	if entryCount != 1 {
		t.Errorf("remaining entries = %d, want 1", entryCount)
	}
}
