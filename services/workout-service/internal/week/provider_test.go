package week

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// seedWeekWithItem inserts a week row and one planned_item owned by it. The
// item is bare — only the fields the nearest-populated-week queries read
// (tenant_id, user_id, week_id) are set.
func seedWeekWithItem(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID, start time.Time) uuid.UUID {
	t.Helper()
	now := time.Now().UTC()
	wkID := uuid.New()
	require.NoError(t, db.Create(&Entity{
		Id:            wkID,
		TenantId:      tenantID,
		UserId:        userID,
		WeekStartDate: start,
		RestDayFlags:  json.RawMessage("[]"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error)
	require.NoError(t, db.Create(&planneditem.Entity{
		Id:         uuid.New(),
		TenantId:   tenantID,
		UserId:     userID,
		WeekId:     wkID,
		ExerciseId: uuid.New(),
		DayOfWeek:  0,
		Position:   0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)
	return wkID
}

func seedBareWeek(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID, start time.Time) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.Create(&Entity{
		Id:            uuid.New(),
		TenantId:      tenantID,
		UserId:        userID,
		WeekStartDate: start,
		RestDayFlags:  json.RawMessage("[]"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error)
}

func providerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&planneditem.Entity{}))
	return db
}

func TestGetSoonestNextWithItems_HitsAndMisses(t *testing.T) {
	db := providerTestDB(t)
	tenantID, userID := uuid.New(), uuid.New()

	wk1 := mustParse(t, "2026-03-30")
	wk2 := mustParse(t, "2026-04-13")
	wk3 := mustParse(t, "2026-04-20")
	seedWeekWithItem(t, db, tenantID, userID, wk1)
	seedWeekWithItem(t, db, tenantID, userID, wk2)
	seedWeekWithItem(t, db, tenantID, userID, wk3)

	t.Run("strictly-greater hit returns first populated week", func(t *testing.T) {
		ref := mustParse(t, "2026-04-06")
		got, err := GetSoonestNextWithItems(db, userID, ref)
		require.NoError(t, err)
		assert.Equal(t, "2026-04-13", got.WeekStartDate.Format("2006-01-02"))
	})

	t.Run("boundary excludes the reference week itself", func(t *testing.T) {
		// reference exactly matches wk2 — the call must skip it and return wk3.
		got, err := GetSoonestNextWithItems(db, userID, wk2)
		require.NoError(t, err)
		assert.Equal(t, "2026-04-20", got.WeekStartDate.Format("2006-01-02"))
	})

	t.Run("miss returns gorm ErrRecordNotFound", func(t *testing.T) {
		// reference past the last populated week.
		_, err := GetSoonestNextWithItems(db, userID, mustParse(t, "2026-04-27"))
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("ignores weeks with no planned items", func(t *testing.T) {
		// Seed a later empty week. The helper should continue skipping past it
		// because the INNER JOIN filters out weeks with no planned_items.
		otherUser := uuid.New()
		seedBareWeek(t, db, tenantID, otherUser, mustParse(t, "2026-05-04"))
		_, err := GetSoonestNextWithItems(db, otherUser, mustParse(t, "2026-04-27"))
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestGetSoonestNextWithItems_CrossUserIsolation(t *testing.T) {
	db := providerTestDB(t)
	tenantID := uuid.New()
	userA, userB := uuid.New(), uuid.New()

	// Only userA has a populated future week.
	seedWeekWithItem(t, db, tenantID, userA, mustParse(t, "2026-04-13"))

	_, err := GetSoonestNextWithItems(db, userB, mustParse(t, "2026-04-06"))
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "userB must not see userA's populated week")
}

func TestGetMostRecentPriorWithItems_Boundary(t *testing.T) {
	db := providerTestDB(t)
	tenantID, userID := uuid.New(), uuid.New()

	seedWeekWithItem(t, db, tenantID, userID, mustParse(t, "2026-03-30"))
	seedWeekWithItem(t, db, tenantID, userID, mustParse(t, "2026-04-13"))

	// reference exactly equals the later week — prior must skip it.
	got, err := GetMostRecentPriorWithItems(db, userID, mustParse(t, "2026-04-13"))
	require.NoError(t, err)
	assert.Equal(t, "2026-03-30", got.WeekStartDate.Format("2006-01-02"))
}
