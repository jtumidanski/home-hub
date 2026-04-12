package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/event"
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
	if err := db.AutoMigrate(&event.Entity{}, &sr.RunEntity{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPastEventsReapBoundary(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now()

	mkEvent := func(endTime time.Time) {
		db.Create(&event.Entity{
			Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
			ConnectionId: uuid.New(), SourceId: uuid.New(), UserId: uuid.New(),
			ExternalId: uuid.New().String(), Title: "Event",
			StartTime: endTime.Add(-time.Hour), EndTime: endTime,
			Visibility: "default", UserDisplayName: "User", UserColor: "#000000",
			CreatedAt: now, UpdatedAt: now,
		})
	}

	mkEvent(now.Add(-400 * 24 * time.Hour)) // old — should be reaped
	mkEvent(now.Add(-100 * 24 * time.Hour)) // inside 365-day window — should survive
	mkEvent(now.Add(7 * 24 * time.Hour))    // future — must survive

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID}
	res, err := PastEvents{}.Reap(context.Background(), db, scope, 365, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 1 {
		t.Errorf("deleted = %d, want 1", res.Deleted)
	}

	var remaining int64
	db.Model(&event.Entity{}).Count(&remaining)
	if remaining != 2 {
		t.Errorf("remaining events = %d, want 2 (recent + future)", remaining)
	}
}

func TestPastEventsDoesNotDeleteFutureEvents(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	now := time.Now()

	// Only future events exist.
	for i := 0; i < 3; i++ {
		db.Create(&event.Entity{
			Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID,
			ConnectionId: uuid.New(), SourceId: uuid.New(), UserId: uuid.New(),
			ExternalId: uuid.New().String(), Title: "Future",
			StartTime: now.Add(time.Duration(i+1) * 24 * time.Hour),
			EndTime:   now.Add(time.Duration(i+2) * 24 * time.Hour),
			Visibility: "default", UserDisplayName: "User", UserColor: "#000000",
			CreatedAt: now, UpdatedAt: now,
		})
	}

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeHousehold, ScopeId: householdID}
	res, err := PastEvents{}.Reap(context.Background(), db, scope, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 0 {
		t.Errorf("deleted = %d, want 0 — future events must never be reaped", res.Deleted)
	}
}
