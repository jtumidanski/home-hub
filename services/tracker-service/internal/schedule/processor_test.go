package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}))
	return db
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestBuilder_Build(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (Model, error)
		wantErr error
	}{
		{
			name: "valid",
			build: func() (Model, error) {
				return NewBuilder().
					SetTrackingItemID(uuid.New()).
					SetSchedule([]int{1, 3, 5}).
					SetEffectiveDate(time.Now()).
					Build()
			},
			wantErr: nil,
		},
		{
			name: "missing tracking item",
			build: func() (Model, error) {
				return NewBuilder().
					SetSchedule([]int{1}).
					SetEffectiveDate(time.Now()).
					Build()
			},
			wantErr: ErrTrackingItemRequired,
		},
		{
			name: "missing effective date",
			build: func() (Model, error) {
				return NewBuilder().
					SetTrackingItemID(uuid.New()).
					SetSchedule([]int{1}).
					Build()
			},
			wantErr: ErrEffectiveDateRequired,
		},
		{
			name: "out-of-range day",
			build: func() (Model, error) {
				return NewBuilder().
					SetTrackingItemID(uuid.New()).
					SetSchedule([]int{0, 7}).
					SetEffectiveDate(time.Now()).
					Build()
			},
			wantErr: ErrInvalidScheduleDay,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.build()
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestMakeAndToEntity_RoundTrip(t *testing.T) {
	itemID := uuid.New()
	effDate := time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	original, err := NewBuilder().
		SetId(uuid.New()).
		SetTrackingItemID(itemID).
		SetSchedule([]int{1, 3, 5}).
		SetEffectiveDate(effDate).
		SetCreatedAt(createdAt).
		Build()
	require.NoError(t, err)

	entity := original.ToEntity()
	restored, err := Make(entity)
	require.NoError(t, err)

	assert.Equal(t, original.Id(), restored.Id())
	assert.Equal(t, original.TrackingItemID(), restored.TrackingItemID())
	assert.Equal(t, original.Schedule(), restored.Schedule())
	assert.Equal(t, original.EffectiveDate(), restored.EffectiveDate())
}

func TestProcessor_CreateSnapshot_UpsertsSameDay(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	itemID := uuid.New()
	day := time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)

	first, err := p.CreateSnapshot(itemID, []int{1, 3}, day)
	require.NoError(t, err)

	second, err := p.CreateSnapshot(itemID, []int{2, 4}, day)
	require.NoError(t, err)

	assert.Equal(t, first.Id(), second.Id(), "same-day snapshot must upsert, not collide")
	assert.Equal(t, []int{2, 4}, second.Schedule())
}

func TestProcessor_CreateSnapshot_RejectsInvalidDays(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, err := p.CreateSnapshot(uuid.New(), []int{0, 9}, time.Now().UTC())
	assert.ErrorIs(t, err, ErrInvalidScheduleDay)
}

func TestProcessor_GetEffective_ReturnsLatestPriorSnapshot(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	itemID := uuid.New()
	_, err := p.CreateSnapshot(itemID, []int{1}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	_, err = p.CreateSnapshot(itemID, []int{2, 4}, time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	m, err := p.GetEffective(itemID, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.Equal(t, []int{2, 4}, m.Schedule())
}

func TestProcessor_GetEffective_NotFound(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, err := p.GetEffective(uuid.New(), time.Now().UTC())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProcessor_GetHistoriesByItems(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	a := uuid.New()
	b := uuid.New()
	_, err := p.CreateSnapshot(a, []int{1}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	_, err = p.CreateSnapshot(a, []int{2}, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	_, err = p.CreateSnapshot(b, []int{3}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	histories, err := p.GetHistoriesByItems([]uuid.UUID{a, b})
	require.NoError(t, err)
	assert.Len(t, histories[a], 2)
	assert.Len(t, histories[b], 1)
}
