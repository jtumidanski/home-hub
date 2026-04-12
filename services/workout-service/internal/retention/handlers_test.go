package retention

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
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
		&theme.Entity{},
		&region.Entity{},
		&exercise.Entity{},
		&planneditem.Entity{},
		&performance.Entity{},
		&performance.SetEntity{},
		&sr.RunEntity{},
	); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPerformancesReapBoundary(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	old := now.Add(-400 * 24 * time.Hour)
	young := now.Add(-10 * 24 * time.Hour)

	mkPerf := func(createdAt time.Time) uuid.UUID {
		id := uuid.New()
		db.Create(&performance.Entity{
			Id: id, TenantId: tenantID, UserId: userID,
			PlannedItemId: uuid.New(), Status: "done", Mode: "summary",
			CreatedAt: createdAt, UpdatedAt: createdAt,
		})
		return id
	}

	oldPerfID := mkPerf(old)
	mkPerf(young) // should survive

	// Add performance_sets for the old performance.
	db.Create(&performance.SetEntity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		PerformanceId: oldPerfID, SetNumber: 1, Reps: 10, Weight: 50,
		CreatedAt: old,
	})
	db.Create(&performance.SetEntity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		PerformanceId: oldPerfID, SetNumber: 2, Reps: 8, Weight: 55,
		CreatedAt: old,
	})

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeUser, ScopeId: userID}
	res, err := Performances{}.Reap(context.Background(), db, scope, 365, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Scanned != 1 {
		t.Errorf("scanned = %d, want 1", res.Scanned)
	}
	// 2 sets + 1 performance = 3
	if res.Deleted != 3 {
		t.Errorf("deleted = %d, want 3", res.Deleted)
	}

	var perfCount int64
	db.Model(&performance.Entity{}).Count(&perfCount)
	if perfCount != 1 {
		t.Errorf("remaining performances = %d, want 1", perfCount)
	}
	var setCount int64
	db.Model(&performance.SetEntity{}).Count(&setCount)
	if setCount != 0 {
		t.Errorf("remaining sets = %d, want 0", setCount)
	}
}

func TestDeletedCatalogCascade(t *testing.T) {
	db := newDB(t)
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	past := now.Add(-60 * 24 * time.Hour)

	scope := sr.Scope{TenantId: tenantID, Kind: sr.ScopeUser, ScopeId: userID}

	// Create a soft-deleted theme past the 30-day window.
	themeID := uuid.New()
	db.Create(&theme.Entity{
		Id: themeID, TenantId: tenantID, UserId: userID, Name: "Old Theme",
		CreatedAt: past, UpdatedAt: past, DeletedAt: &past,
	})

	// A live theme that should survive.
	liveThemeID := uuid.New()
	db.Create(&theme.Entity{
		Id: liveThemeID, TenantId: tenantID, UserId: userID, Name: "Live Theme",
		CreatedAt: now, UpdatedAt: now,
	})

	// Region under the deleted theme (regions are independent entities but we
	// also test standalone region deletion).
	regionID := uuid.New()
	db.Create(&region.Entity{
		Id: regionID, TenantId: tenantID, UserId: userID, Name: "Old Region",
		CreatedAt: past, UpdatedAt: past, DeletedAt: &past,
	})

	// Exercise under the deleted theme.
	exerciseID := uuid.New()
	emptyJSON, _ := json.Marshal([]string{})
	db.Create(&exercise.Entity{
		Id: exerciseID, TenantId: tenantID, UserId: userID,
		Name: "Old Exercise", Kind: "weight", WeightType: "free",
		ThemeId: themeID, RegionId: regionID,
		SecondaryRegionIds: emptyJSON,
		CreatedAt:          past, UpdatedAt: past,
	})

	// Planned item linking exercise to performance.
	plannedItemID := uuid.New()
	db.Create(&planneditem.Entity{
		Id: plannedItemID, TenantId: tenantID, UserId: userID,
		WeekId: uuid.New(), ExerciseId: exerciseID, DayOfWeek: 1,
		CreatedAt: past, UpdatedAt: past,
	})

	// Performance + set referencing the exercise via planned_item.
	perfID := uuid.New()
	db.Create(&performance.Entity{
		Id: perfID, TenantId: tenantID, UserId: userID,
		PlannedItemId: plannedItemID, Status: "done", Mode: "summary",
		CreatedAt: past, UpdatedAt: past,
	})
	db.Create(&performance.SetEntity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		PerformanceId: perfID, SetNumber: 1, Reps: 10, Weight: 50,
		CreatedAt: past,
	})

	// A live region that should survive.
	liveRegionID := uuid.New()
	db.Create(&region.Entity{
		Id: liveRegionID, TenantId: tenantID, UserId: userID, Name: "Live Region",
		CreatedAt: now, UpdatedAt: now,
	})

	// Exercise under the live theme and live region (should survive).
	liveExerciseID := uuid.New()
	db.Create(&exercise.Entity{
		Id: liveExerciseID, TenantId: tenantID, UserId: userID,
		Name: "Live Exercise", Kind: "weight", WeightType: "free",
		ThemeId: liveThemeID, RegionId: liveRegionID,
		SecondaryRegionIds: emptyJSON,
		CreatedAt:          now, UpdatedAt: now,
	})

	res, err := DeletedCatalogRestoreWindow{}.Reap(context.Background(), db, scope, 30, false)
	if err != nil {
		t.Fatal(err)
	}

	if res.Scanned == 0 {
		t.Error("expected scanned > 0")
	}
	if res.Deleted == 0 {
		t.Error("expected deleted > 0")
	}

	// The deleted theme and region should be gone.
	var themeCount int64
	db.Model(&theme.Entity{}).Count(&themeCount)
	if themeCount != 1 {
		t.Errorf("remaining themes = %d, want 1 (the live theme)", themeCount)
	}

	var regionCount int64
	db.Model(&region.Entity{}).Count(&regionCount)
	if regionCount != 1 {
		t.Errorf("remaining regions = %d, want 1 (the live region)", regionCount)
	}

	// The exercise under the deleted theme should be gone; the live one stays.
	var exerciseCount int64
	db.Model(&exercise.Entity{}).Count(&exerciseCount)
	if exerciseCount != 1 {
		t.Errorf("remaining exercises = %d, want 1", exerciseCount)
	}

	// Performance and set under the deleted exercise should be gone.
	var perfCount int64
	db.Model(&performance.Entity{}).Count(&perfCount)
	if perfCount != 0 {
		t.Errorf("remaining performances = %d, want 0", perfCount)
	}
	var setCount int64
	db.Model(&performance.SetEntity{}).Count(&setCount)
	if setCount != 0 {
		t.Errorf("remaining sets = %d, want 0", setCount)
	}
}
