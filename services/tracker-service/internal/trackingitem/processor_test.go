package trackingitem

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
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
	require.NoError(t, db.AutoMigrate(&Entity{}, &schedule.Entity{}))
	return db
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestProcessor_Create_WritesInitialScheduleSnapshot(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tenantID := uuid.New()
	userID := uuid.New()
	m, err := p.Create(tenantID, userID, "Running", "sentiment", "blue", nil, []int{1, 3, 5}, 0)
	require.NoError(t, err)
	assert.Equal(t, "Running", m.Name())
	assert.Equal(t, 1, m.SortOrder(), "first item should auto-assign sort_order 1")

	history, err := p.GetScheduleHistory(m.Id())
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, []int{1, 3, 5}, history[0].Schedule())
}

func TestProcessor_Create_DuplicateNameReturns409Error(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()

	_, err := p.Create(tenantID, userID, "Running", "sentiment", "blue", nil, nil, 0)
	require.NoError(t, err)

	_, err = p.Create(tenantID, userID, "Running", "numeric", "red", nil, nil, 0)
	assert.ErrorIs(t, err, ErrDuplicateName)
}

func TestProcessor_Create_DuplicateNameAllowedAcrossUsers(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tenantID := uuid.New()

	_, err := p.Create(tenantID, uuid.New(), "Running", "sentiment", "blue", nil, nil, 0)
	require.NoError(t, err)

	_, err = p.Create(tenantID, uuid.New(), "Running", "sentiment", "red", nil, nil, 0)
	require.NoError(t, err)
}

func TestProcessor_Create_RangeRequiresConfig(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, err := p.Create(uuid.New(), uuid.New(), "Sleep", "range", "blue", nil, nil, 0)
	assert.ErrorIs(t, err, ErrRangeConfigRequired)

	cfg, _ := json.Marshal(RangeConfig{Min: 0, Max: 100})
	_, err = p.Create(uuid.New(), uuid.New(), "Sleep", "range", "blue", cfg, nil, 0)
	require.NoError(t, err)
}

func TestProcessor_Update_SameDayScheduleChangeOverwritesSnapshot(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), "Running", "sentiment", "blue", nil, []int{1, 3, 5}, 0)
	require.NoError(t, err)

	// Editing the schedule on the same calendar day as creation must upsert the
	// existing snapshot rather than failing with a unique-constraint error.
	newSched := []int{2, 4}
	_, err = p.Update(m.Id(), nil, nil, &newSched, nil, nil)
	require.NoError(t, err)

	history, err := p.GetScheduleHistory(m.Id())
	require.NoError(t, err)
	require.Len(t, history, 1, "same-day schedule edit should overwrite, not append")
	assert.Equal(t, newSched, history[0].Schedule())
}

func TestProcessor_Update_LaterDayScheduleChangeCreatesNewSnapshot(t *testing.T) {
	// On a different effective date, schedule changes should create a new snapshot
	// so historical month math sees the prior schedule. We simulate "different day"
	// by directly seeding a back-dated snapshot, then calling Update.
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), "Running", "sentiment", "blue", nil, []int{1, 3, 5}, 0)
	require.NoError(t, err)

	// Backdate the auto-created snapshot to a week ago so the next Update
	// creates a fresh one for "today".
	require.NoError(t,
		db.Exec(`UPDATE schedule_snapshots SET effective_date = date('now', '-7 days') WHERE tracking_item_id = ?`, m.Id()).Error,
	)

	newSched := []int{2, 4}
	_, err = p.Update(m.Id(), nil, nil, &newSched, nil, nil)
	require.NoError(t, err)

	history, err := p.GetScheduleHistory(m.Id())
	require.NoError(t, err)
	require.Len(t, history, 2)
	assert.Equal(t, []int{1, 3, 5}, history[0].Schedule())
	assert.Equal(t, newSched, history[1].Schedule())
}

func TestProcessor_Update_NoScheduleChangePreservesSnapshots(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), "Running", "sentiment", "blue", nil, []int{1, 3, 5}, 0)
	require.NoError(t, err)

	newName := "Jogging"
	_, err = p.Update(m.Id(), &newName, nil, nil, nil, nil)
	require.NoError(t, err)

	history, err := p.GetScheduleHistory(m.Id())
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestProcessor_Update_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()

	_, err := p.Create(tenantID, userID, "Running", "sentiment", "blue", nil, nil, 0)
	require.NoError(t, err)
	other, err := p.Create(tenantID, userID, "Cycling", "sentiment", "red", nil, nil, 0)
	require.NoError(t, err)

	collide := "Running"
	_, err = p.Update(other.Id(), &collide, nil, nil, nil, nil)
	assert.ErrorIs(t, err, ErrDuplicateName)
}

func TestProcessor_Delete_SoftDeletesAndExcludesFromList(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	m, err := p.Create(uuid.New(), userID, "Running", "sentiment", "blue", nil, nil, 0)
	require.NoError(t, err)

	require.NoError(t, p.Delete(m.Id()))

	listed, err := p.List(userID)
	require.NoError(t, err)
	assert.Empty(t, listed, "soft-deleted items must not appear in List")

	// Soft-deleted items can be reached via the include-deleted query (used by month summary).
	all, err := GetAllByUserIncludeDeleted(userID)(db)()
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.NotNil(t, all[0].DeletedAt)
}

func TestProcessor_Delete_AllowsNameReuse(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()

	m, err := p.Create(tenantID, userID, "Running", "sentiment", "blue", nil, nil, 0)
	require.NoError(t, err)
	require.NoError(t, p.Delete(m.Id()))

	_, err = p.Create(tenantID, userID, "Running", "sentiment", "red", nil, nil, 0)
	require.NoError(t, err, "deleted name should be reusable")
}

func TestProcessor_Create_AutoIncrementsSortOrder(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tenantID := uuid.New()
	userID := uuid.New()

	m1, err := p.Create(tenantID, userID, "A", "sentiment", "blue", nil, nil, 0)
	require.NoError(t, err)
	m2, err := p.Create(tenantID, userID, "B", "sentiment", "red", nil, nil, 0)
	require.NoError(t, err)

	assert.Equal(t, 1, m1.SortOrder())
	assert.Equal(t, 2, m2.SortOrder())
}
