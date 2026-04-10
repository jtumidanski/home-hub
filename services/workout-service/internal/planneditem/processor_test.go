package planneditem

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB spins up an in-memory SQLite for the planneditem processor.
// We bypass the per-package Migration helpers because they emit Postgres-only
// DDL (partial unique indexes, ALTER TABLE FK rewrites). AutoMigrate is
// sufficient for the behavioural tests below.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}, &exercise.Entity{}, &week.Entity{}))
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

// seedExercise inserts a strength exercise tied to the supplied user. Tests
// that need a planned-item parent can call this once and reuse the returned
// UUID. The exercise is created via direct entity insert so we don't depend
// on theme/region tables.
func seedExercise(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&exercise.Entity{
		Id:                 id,
		TenantId:           tenantID,
		UserId:             userID,
		Name:               "Bench Press " + id.String()[:8],
		Kind:               exercise.KindStrength,
		WeightType:         exercise.WeightTypeFree,
		ThemeId:            uuid.New(),
		RegionId:           uuid.New(),
		SecondaryRegionIds: json.RawMessage("[]"),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}).Error)
	return id
}

// softDeleteExercise marks an exercise row as soft-deleted so the planneditem
// processor's "reject soft-deleted exercise on add" path can be exercised.
func softDeleteExercise(t *testing.T, db *gorm.DB, id uuid.UUID) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.Model(&exercise.Entity{}).Where("id = ?", id).Update("deleted_at", &now).Error)
}

// seedWeek inserts a week row and returns its ID. Tests use it as the parent
// container for the planned items they create.
func seedWeek(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&week.Entity{
		Id:            id,
		TenantId:      tenantID,
		UserId:        userID,
		WeekStartDate: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
		RestDayFlags:  json.RawMessage("[]"),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}).Error)
	return id
}

func TestProcessor_Add_AssignsNextPositionPerDay(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()
	weekID := seedWeek(t, db, tenantID, userID)
	exID := seedExercise(t, db, tenantID, userID)

	// Three adds on the same day should land at positions 0, 1, 2.
	for i := 0; i < 3; i++ {
		m, err := p.Add(tenantID, userID, weekID, AddInput{ExerciseID: exID, DayOfWeek: 0})
		require.NoError(t, err)
		assert.Equal(t, i, m.Position(), "expected sequential auto-assigned position")
	}

	// An item on a different day should restart at 0.
	m, err := p.Add(tenantID, userID, weekID, AddInput{ExerciseID: exID, DayOfWeek: 1})
	require.NoError(t, err)
	assert.Equal(t, 0, m.Position())
}

func TestProcessor_Add_RejectsSoftDeletedExercise(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()
	weekID := seedWeek(t, db, tenantID, userID)
	exID := seedExercise(t, db, tenantID, userID)
	softDeleteExercise(t, db, exID)

	_, err := p.Add(tenantID, userID, weekID, AddInput{ExerciseID: exID, DayOfWeek: 0})
	assert.ErrorIs(t, err, ErrExerciseDeleted, "soft-deleted exercise must yield 422 ErrExerciseDeleted")
}

func TestProcessor_Add_RejectsExerciseFromOtherUser(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()
	weekID := seedWeek(t, db, tenantID, userA)
	exID := seedExercise(t, db, tenantID, userB) // belongs to userB

	_, err := p.Add(tenantID, userA, weekID, AddInput{ExerciseID: exID, DayOfWeek: 0})
	assert.ErrorIs(t, err, ErrExerciseMismatch)
}

func TestProcessor_BulkAdd_AtomicallyRollsBackOnFailure(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()
	weekID := seedWeek(t, db, tenantID, userID)
	exID := seedExercise(t, db, tenantID, userID)
	deletedExID := seedExercise(t, db, tenantID, userID)
	softDeleteExercise(t, db, deletedExID)

	inputs := []AddInput{
		{ExerciseID: exID, DayOfWeek: 0},
		{ExerciseID: exID, DayOfWeek: 1},
		{ExerciseID: deletedExID, DayOfWeek: 2}, // poisons the batch
	}
	_, err := p.BulkAdd(tenantID, userID, weekID, inputs)
	require.Error(t, err, "expected the deleted-exercise row to fail the batch")

	// Nothing should have been committed because the transaction rolled back.
	var count int64
	require.NoError(t, db.Model(&Entity{}).Where("week_id = ?", weekID).Count(&count).Error)
	assert.Equal(t, int64(0), count, "BulkAdd must be transactional — partial inserts are forbidden")
}

func TestProcessor_BulkAdd_MultiDayHappyPath(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()
	weekID := seedWeek(t, db, tenantID, userID)
	exID := seedExercise(t, db, tenantID, userID)

	inputs := []AddInput{
		{ExerciseID: exID, DayOfWeek: 0},
		{ExerciseID: exID, DayOfWeek: 0},
		{ExerciseID: exID, DayOfWeek: 3},
	}
	models, err := p.BulkAdd(tenantID, userID, weekID, inputs)
	require.NoError(t, err)
	require.Len(t, models, 3)
	assert.Equal(t, 0, models[0].Position())
	assert.Equal(t, 1, models[1].Position(), "second add on day 0 should slot in position 1")
	assert.Equal(t, 0, models[2].Position(), "first add on day 3 restarts positions")
}

func TestProcessor_Reorder_AppliesAtomically(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()
	weekID := seedWeek(t, db, tenantID, userID)
	exID := seedExercise(t, db, tenantID, userID)

	first, err := p.Add(tenantID, userID, weekID, AddInput{ExerciseID: exID, DayOfWeek: 0})
	require.NoError(t, err)
	second, err := p.Add(tenantID, userID, weekID, AddInput{ExerciseID: exID, DayOfWeek: 0})
	require.NoError(t, err)

	// Move both to day 2 with swapped positions; include a bogus entry to
	// verify the transaction rolls back when one row is missing.
	bogus := uuid.New()
	err = p.Reorder(weekID, []ReorderEntry{
		{ItemID: first.Id(), DayOfWeek: 2, Position: 1},
		{ItemID: second.Id(), DayOfWeek: 2, Position: 0},
		{ItemID: bogus, DayOfWeek: 2, Position: 2},
	})
	require.ErrorIs(t, err, ErrNotFound, "bogus item id should fail with ErrNotFound")

	// Both legitimate updates must have rolled back — the rows still live on day 0.
	var rows []Entity
	require.NoError(t, db.Where("week_id = ?", weekID).Find(&rows).Error)
	for _, r := range rows {
		assert.Equal(t, 0, r.DayOfWeek, "Reorder must be atomic — failed batch should not persist any update")
	}

	// Happy-path reorder: only legitimate entries.
	require.NoError(t, p.Reorder(weekID, []ReorderEntry{
		{ItemID: first.Id(), DayOfWeek: 2, Position: 1},
		{ItemID: second.Id(), DayOfWeek: 2, Position: 0},
	}))
	require.NoError(t, db.Where("week_id = ?", weekID).Find(&rows).Error)
	byID := map[uuid.UUID]Entity{rows[0].Id: rows[0], rows[1].Id: rows[1]}
	assert.Equal(t, 1, byID[first.Id()].Position)
	assert.Equal(t, 0, byID[second.Id()].Position)
	for _, r := range rows {
		assert.Equal(t, 2, r.DayOfWeek)
	}
}

func TestProcessor_Reorder_RejectsInvalidDayOrPosition(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	weekID := uuid.New()
	itemID := uuid.New()

	err := p.Reorder(weekID, []ReorderEntry{{ItemID: itemID, DayOfWeek: 9, Position: 0}})
	assert.ErrorIs(t, err, ErrInvalidDayOfWeek)

	err = p.Reorder(weekID, []ReorderEntry{{ItemID: itemID, DayOfWeek: 0, Position: -1}})
	assert.ErrorIs(t, err, ErrInvalidPosition)
}
