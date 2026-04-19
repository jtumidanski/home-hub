package summary

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
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func summaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(
		&theme.Entity{},
		&region.Entity{},
		&exercise.Entity{},
		&week.Entity{},
		&planneditem.Entity{},
		&performance.Entity{},
		&performance.SetEntity{},
	))
	return db
}

type fixture struct {
	tenantID   uuid.UUID
	userID     uuid.UUID
	themeID    uuid.UUID
	regionID   uuid.UUID
	exerciseID uuid.UUID
}

func seedFixture(t *testing.T, db *gorm.DB) fixture {
	t.Helper()
	now := time.Now().UTC()
	f := fixture{
		tenantID:   uuid.New(),
		userID:     uuid.New(),
		themeID:    uuid.New(),
		regionID:   uuid.New(),
		exerciseID: uuid.New(),
	}
	require.NoError(t, db.Create(&theme.Entity{
		Id: f.themeID, TenantId: f.tenantID, UserId: f.userID,
		Name: "Strength", SortOrder: 0, CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&region.Entity{
		Id: f.regionID, TenantId: f.tenantID, UserId: f.userID,
		Name: "Chest", SortOrder: 0, CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&exercise.Entity{
		Id: f.exerciseID, TenantId: f.tenantID, UserId: f.userID,
		Name: "Bench", Kind: exercise.KindStrength, WeightType: exercise.WeightTypeFree,
		ThemeId: f.themeID, RegionId: f.regionID,
		SecondaryRegionIds: json.RawMessage("[]"),
		CreatedAt:          now, UpdatedAt: now,
	}).Error)
	return f
}

// seedPopulatedWeek creates a week + one planned item owned by (tenant, user).
// Returns the week entity so callers can build a `week.Model` for the Build
// call.
func seedPopulatedWeek(t *testing.T, db *gorm.DB, f fixture, start time.Time) week.Entity {
	t.Helper()
	now := time.Now().UTC()
	wk := week.Entity{
		Id: uuid.New(), TenantId: f.tenantID, UserId: f.userID,
		WeekStartDate: start, RestDayFlags: json.RawMessage("[]"),
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&wk).Error)
	require.NoError(t, db.Create(&planneditem.Entity{
		Id: uuid.New(), TenantId: f.tenantID, UserId: f.userID,
		WeekId: wk.Id, ExerciseId: f.exerciseID,
		DayOfWeek: 0, Position: 0,
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	return wk
}

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	require.NoError(t, err)
	return v
}

func buildForWeek(t *testing.T, db *gorm.DB, wk week.Entity) RestModel {
	t.Helper()
	l, _ := test.NewNullLogger()
	m, err := week.Make(wk)
	require.NoError(t, err)
	proc := NewProcessor(l, context.Background(), db)
	out, err := proc.Build(m)
	require.NoError(t, err)
	return out
}

func TestBuild_PopulatesNavPointers_Neither(t *testing.T) {
	db := summaryTestDB(t)
	f := seedFixture(t, db)
	wk := seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-06"))

	rm := buildForWeek(t, db, wk)
	assert.Nil(t, rm.PreviousPopulatedWeek)
	assert.Nil(t, rm.NextPopulatedWeek)
}

func TestBuild_PopulatesNavPointers_PriorOnly(t *testing.T) {
	db := summaryTestDB(t)
	f := seedFixture(t, db)
	seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-03-30"))
	wk := seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-06"))

	rm := buildForWeek(t, db, wk)
	require.NotNil(t, rm.PreviousPopulatedWeek)
	assert.Equal(t, "2026-03-30", *rm.PreviousPopulatedWeek)
	assert.Nil(t, rm.NextPopulatedWeek)
}

func TestBuild_PopulatesNavPointers_NextOnly(t *testing.T) {
	db := summaryTestDB(t)
	f := seedFixture(t, db)
	wk := seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-06"))
	seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-13"))

	rm := buildForWeek(t, db, wk)
	assert.Nil(t, rm.PreviousPopulatedWeek)
	require.NotNil(t, rm.NextPopulatedWeek)
	assert.Equal(t, "2026-04-13", *rm.NextPopulatedWeek)
}

func TestBuild_PopulatesNavPointers_Both(t *testing.T) {
	db := summaryTestDB(t)
	f := seedFixture(t, db)
	seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-03-30"))
	wk := seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-06"))
	seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-13"))

	rm := buildForWeek(t, db, wk)
	require.NotNil(t, rm.PreviousPopulatedWeek)
	require.NotNil(t, rm.NextPopulatedWeek)
	assert.Equal(t, "2026-03-30", *rm.PreviousPopulatedWeek)
	assert.Equal(t, "2026-04-13", *rm.NextPopulatedWeek)
}

// TestBuild_PerSetRowsEmittedOnlyForPerSetMode asserts the per-set
// `setRows` array is present on per-set performances and absent on
// summary-mode performances.
func TestBuild_PerSetRowsEmittedOnlyForPerSetMode(t *testing.T) {
	db := summaryTestDB(t)
	f := seedFixture(t, db)
	wk := seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-06"))

	// Replace the seeded planned item's performance with a per-set one.
	var item planneditem.Entity
	require.NoError(t, db.Where("week_id = ?", wk.Id).First(&item).Error)

	now := time.Now().UTC()
	unit := "lb"
	perf := performance.Entity{
		Id: uuid.New(), TenantId: f.tenantID, UserId: f.userID,
		PlannedItemId: item.Id, Status: performance.StatusDone,
		Mode:       performance.ModePerSet,
		WeightUnit: &unit,
		CreatedAt:  now, UpdatedAt: now,
	}
	require.NoError(t, db.Create(&perf).Error)
	for i, row := range []struct{ reps int; weight float64 }{{10, 135}, {10, 140}, {8, 145}} {
		require.NoError(t, db.Create(&performance.SetEntity{
			Id: uuid.New(), TenantId: f.tenantID, UserId: f.userID,
			PerformanceId: perf.Id,
			SetNumber:     i + 1,
			Reps:          row.reps, Weight: row.weight,
			CreatedAt: now,
		}).Error)
	}

	rm := buildForWeek(t, db, wk)
	require.Len(t, rm.ByDay, 7)
	require.Len(t, rm.ByDay[0].Items, 1)
	actual, ok := rm.ByDay[0].Items[0].ActualSummary.(map[string]any)
	require.True(t, ok, "actualSummary must be a map")

	rows, ok := actual["setRows"].([]map[string]any)
	require.True(t, ok, "per-set performance must emit setRows array")
	require.Len(t, rows, 3)
	assert.Equal(t, 1, rows[0]["setNumber"])
	assert.Equal(t, 10, rows[0]["reps"])
	assert.Equal(t, 135.0, rows[0]["weight"])

	// Sanity: scalar sets/reps/weight remain for backwards compat.
	assert.Equal(t, 3, actual["sets"])
	assert.Equal(t, 10, actual["reps"])
	assert.Equal(t, 145.0, actual["weight"])
}

func TestBuild_SummaryModeOmitsSetRows(t *testing.T) {
	db := summaryTestDB(t)
	f := seedFixture(t, db)
	wk := seedPopulatedWeek(t, db, f, mustParseDate(t, "2026-04-06"))

	var item planneditem.Entity
	require.NoError(t, db.Where("week_id = ?", wk.Id).First(&item).Error)

	now := time.Now().UTC()
	sets, reps := 3, 10
	weight := 135.0
	unit := "lb"
	require.NoError(t, db.Create(&performance.Entity{
		Id: uuid.New(), TenantId: f.tenantID, UserId: f.userID,
		PlannedItemId: item.Id, Status: performance.StatusDone,
		Mode:       performance.ModeSummary,
		WeightUnit: &unit,
		ActualSets: &sets, ActualReps: &reps, ActualWeight: &weight,
		CreatedAt: now, UpdatedAt: now,
	}).Error)

	rm := buildForWeek(t, db, wk)
	actual, ok := rm.ByDay[0].Items[0].ActualSummary.(map[string]any)
	require.True(t, ok)
	_, hasSetRows := actual["setRows"]
	assert.False(t, hasSetRows, "summary-mode performance must not emit setRows")
}
