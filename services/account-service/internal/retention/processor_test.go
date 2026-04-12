package retention

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	sharedretention "github.com/jtumidanski/home-hub/shared/go/retention"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestResolveAllUsesDefaults(t *testing.T) {
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
	hh := resolved.Household.Values[sharedretention.CatProductivityCompletedTasks]
	if hh.Days != 365 || hh.Source != "default" {
		t.Errorf("expected 365/default, got %+v", hh)
	}
	us := resolved.UserScope.Values[sharedretention.CatTrackerEntries]
	if us.Days != 730 || us.Source != "default" {
		t.Errorf("expected 730/default, got %+v", us)
	}
}

func TestApplyHouseholdPatchUpsertAndDelete(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	tenantID := uuid.New()
	householdID := uuid.New()

	v := 180
	patch := map[sharedretention.Category]*int{
		sharedretention.CatProductivityCompletedTasks: &v,
	}
	if err := p.ApplyHouseholdPatch(tenantID, householdID, patch); err != nil {
		t.Fatal(err)
	}
	resolved, _ := p.ResolveAll(tenantID, householdID, uuid.Nil)
	got := resolved.Household.Values[sharedretention.CatProductivityCompletedTasks]
	if got.Days != 180 || got.Source != "household" {
		t.Errorf("after upsert: %+v", got)
	}

	// delete via nil
	patch[sharedretention.CatProductivityCompletedTasks] = nil
	if err := p.ApplyHouseholdPatch(tenantID, householdID, patch); err != nil {
		t.Fatal(err)
	}
	resolved, _ = p.ResolveAll(tenantID, householdID, uuid.Nil)
	got = resolved.Household.Values[sharedretention.CatProductivityCompletedTasks]
	if got.Days != 365 || got.Source != "default" {
		t.Errorf("after delete: %+v", got)
	}
}

func TestApplyPatchScopeMismatch(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)
	v := 30
	patch := map[sharedretention.Category]*int{
		sharedretention.CatTrackerEntries: &v, // user-scoped
	}
	err := p.ApplyHouseholdPatch(uuid.New(), uuid.New(), patch)
	if err != ErrScopeMismatch {
		t.Errorf("expected ErrScopeMismatch, got %v", err)
	}
}

func TestApplyPatchInvalidDays(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)
	v := 4000
	patch := map[sharedretention.Category]*int{
		sharedretention.CatProductivityCompletedTasks: &v,
	}
	err := p.ApplyHouseholdPatch(uuid.New(), uuid.New(), patch)
	if err != ErrInvalidDays {
		t.Errorf("expected ErrInvalidDays, got %v", err)
	}
}
