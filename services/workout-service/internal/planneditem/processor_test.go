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

// seedExercise inserts a strength exercise tied to the supplied user.
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

// softDeleteExercise marks an exercise row as soft-deleted.
func softDeleteExercise(t *testing.T, db *gorm.DB, id uuid.UUID) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.Model(&exercise.Entity{}).Where("id = ?", id).Update("deleted_at", &now).Error)
}

// seedWeek inserts a week row and returns its ID.
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

func TestProcessor_Add_Rejections(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, db *gorm.DB, tenantID, userID, weekID uuid.UUID) AddInput
		wantErr error
	}{
		{
			name: "soft-deleted exercise",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID, weekID uuid.UUID) AddInput {
				exID := seedExercise(t, db, tenantID, userID)
				softDeleteExercise(t, db, exID)
				return AddInput{ExerciseID: exID, DayOfWeek: 0}
			},
			wantErr: ErrExerciseDeleted,
		},
		{
			name: "exercise from other user",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID, weekID uuid.UUID) AddInput {
				otherUser := uuid.New()
				exID := seedExercise(t, db, tenantID, otherUser)
				return AddInput{ExerciseID: exID, DayOfWeek: 0}
			},
			wantErr: ErrExerciseMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID, userID := uuid.New(), uuid.New()
			weekID := seedWeek(t, db, tenantID, userID)
			in := tc.setup(t, db, tenantID, userID, weekID)
			p := newProcessor(t, db)
			_, err := p.Add(tenantID, userID, weekID, in)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestProcessor_BulkAdd(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(t *testing.T, db *gorm.DB, tenantID, userID, weekID uuid.UUID) []AddInput
		wantErr   bool
		wantCount int
		check     func(t *testing.T, models []Model)
	}{
		{
			name: "multi-day happy path",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID, weekID uuid.UUID) []AddInput {
				exID := seedExercise(t, db, tenantID, userID)
				return []AddInput{
					{ExerciseID: exID, DayOfWeek: 0},
					{ExerciseID: exID, DayOfWeek: 0},
					{ExerciseID: exID, DayOfWeek: 3},
				}
			},
			wantCount: 3,
			check: func(t *testing.T, models []Model) {
				assert.Equal(t, 0, models[0].Position())
				assert.Equal(t, 1, models[1].Position(), "second add on day 0 should slot in position 1")
				assert.Equal(t, 0, models[2].Position(), "first add on day 3 restarts positions")
			},
		},
		{
			name: "atomic rollback on failure",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID, weekID uuid.UUID) []AddInput {
				exID := seedExercise(t, db, tenantID, userID)
				deletedExID := seedExercise(t, db, tenantID, userID)
				softDeleteExercise(t, db, deletedExID)
				return []AddInput{
					{ExerciseID: exID, DayOfWeek: 0},
					{ExerciseID: exID, DayOfWeek: 1},
					{ExerciseID: deletedExID, DayOfWeek: 2}, // poisons the batch
				}
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID, userID := uuid.New(), uuid.New()
			weekID := seedWeek(t, db, tenantID, userID)
			inputs := tc.setup(t, db, tenantID, userID, weekID)
			p := newProcessor(t, db)
			models, err := p.BulkAdd(tenantID, userID, weekID, inputs)
			if tc.wantErr {
				require.Error(t, err)
				// Nothing should have been committed because the transaction rolled back.
				var count int64
				require.NoError(t, db.Model(&Entity{}).Where("week_id = ?", weekID).Count(&count).Error)
				assert.Equal(t, int64(0), count, "BulkAdd must be transactional — partial inserts are forbidden")
				return
			}
			require.NoError(t, err)
			require.Len(t, models, tc.wantCount)
			if tc.check != nil {
				tc.check(t, models)
			}
		})
	}
}

func TestProcessor_Reorder(t *testing.T) {
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

	t.Run("bogus item rolls back entire batch", func(t *testing.T) {
		bogus := uuid.New()
		err := p.Reorder(weekID, []ReorderEntry{
			{ItemID: first.Id(), DayOfWeek: 2, Position: 1},
			{ItemID: second.Id(), DayOfWeek: 2, Position: 0},
			{ItemID: bogus, DayOfWeek: 2, Position: 2},
		})
		require.ErrorIs(t, err, ErrNotFound)

		var rows []Entity
		require.NoError(t, db.Where("week_id = ?", weekID).Find(&rows).Error)
		for _, r := range rows {
			assert.Equal(t, 0, r.DayOfWeek, "Reorder must be atomic — failed batch should not persist any update")
		}
	})

	t.Run("happy path applies atomically", func(t *testing.T) {
		require.NoError(t, p.Reorder(weekID, []ReorderEntry{
			{ItemID: first.Id(), DayOfWeek: 2, Position: 1},
			{ItemID: second.Id(), DayOfWeek: 2, Position: 0},
		}))
		var rows []Entity
		require.NoError(t, db.Where("week_id = ?", weekID).Find(&rows).Error)
		byID := map[uuid.UUID]Entity{rows[0].Id: rows[0], rows[1].Id: rows[1]}
		assert.Equal(t, 1, byID[first.Id()].Position)
		assert.Equal(t, 0, byID[second.Id()].Position)
		for _, r := range rows {
			assert.Equal(t, 2, r.DayOfWeek)
		}
	})
}

func TestProcessor_Reorder_Validation(t *testing.T) {
	cases := []struct {
		name    string
		day     int
		pos     int
		wantErr error
	}{
		{"invalid day", 9, 0, ErrInvalidDayOfWeek},
		{"negative position", 0, -1, ErrInvalidPosition},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			p := newProcessor(t, db)
			err := p.Reorder(uuid.New(), []ReorderEntry{{ItemID: uuid.New(), DayOfWeek: tc.day, Position: tc.pos}})
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
