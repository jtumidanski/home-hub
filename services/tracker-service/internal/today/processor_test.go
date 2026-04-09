package today

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
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
	require.NoError(t, db.AutoMigrate(&trackingitem.Entity{}, &schedule.Entity{}, &entry.Entity{}))
	return db
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

// makeItem inserts a tracking item owned by the given user and returns its id.
func makeItem(t *testing.T, db *gorm.DB, userID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&trackingitem.Entity{
		Id:        id,
		TenantId:  uuid.New(),
		UserId:    userID,
		Name:      name,
		ScaleType: "numeric",
		Color:     "blue",
	}).Error)
	return id
}

func setSchedule(t *testing.T, db *gorm.DB, itemID uuid.UUID, days []int) {
	t.Helper()
	_, err := schedule.CreateSnapshot(db, itemID, days, time.Now().UTC().AddDate(0, -1, 0))
	require.NoError(t, err)
}

func numericValue(n int) json.RawMessage {
	b, _ := json.Marshal(map[string]int{"count": n})
	return b
}

func TestToday_ReturnsItemsScheduledForGivenDayOfWeek(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	userID := uuid.New()
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC) // Wednesday — Weekday() == 3
	wedItem := makeItem(t, db, userID, "wed-only")
	setSchedule(t, db, wedItem, []int{3})

	monItem := makeItem(t, db, userID, "mon-only")
	setSchedule(t, db, monItem, []int{1})

	result, err := p.Today(userID, now)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, wedItem, result.Items[0].Id())
}

func TestToday_EmptyScheduleMeansEveryDay(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	userID := uuid.New()
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	item := makeItem(t, db, userID, "everyday")
	setSchedule(t, db, item, []int{}) // empty schedule == every day

	result, err := p.Today(userID, now)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, item, result.Items[0].Id())
}

func TestToday_PairsExistingEntriesForToday(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	userID := uuid.New()
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	day := now.Truncate(24 * time.Hour)

	itemWithEntry := makeItem(t, db, userID, "with-entry")
	setSchedule(t, db, itemWithEntry, []int{})
	itemWithoutEntry := makeItem(t, db, userID, "without-entry")
	setSchedule(t, db, itemWithoutEntry, []int{})

	require.NoError(t, db.Create(&entry.Entity{
		Id:             uuid.New(),
		TenantId:       uuid.New(),
		UserId:         userID,
		TrackingItemId: itemWithEntry,
		Date:           day,
		Value:          numericValue(7),
	}).Error)

	result, err := p.Today(userID, now)
	require.NoError(t, err)
	assert.Len(t, result.Items, 2, "both items should be returned")
	require.Len(t, result.Entries, 1, "only the item with a logged entry should appear in entries")
	assert.Equal(t, itemWithEntry, result.Entries[0].TrackingItemID())
}

func TestToday_ExcludesItemsWithoutSchedule(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	userID := uuid.New()
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	makeItem(t, db, userID, "no-schedule") // never call setSchedule

	result, err := p.Today(userID, now)
	require.NoError(t, err)
	assert.Empty(t, result.Items, "items without an effective schedule snapshot are excluded")
	assert.Empty(t, result.Entries)
}

func TestToday_OnlyReturnsItemsForRequestedUser(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)

	user1 := uuid.New()
	user2 := uuid.New()
	myItem := makeItem(t, db, user1, "mine")
	otherItem := makeItem(t, db, user2, "theirs")
	setSchedule(t, db, myItem, []int{})
	setSchedule(t, db, otherItem, []int{})

	result, err := p.Today(user1, now)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, myItem, result.Items[0].Id())
}

func TestTransform_EmitsExpectedJsonApiShape(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	userID := uuid.New()
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	item := makeItem(t, db, userID, "shape-check")
	setSchedule(t, db, item, []int{})

	result, err := p.Today(userID, now)
	require.NoError(t, err)

	doc := Transform(result)
	assert.Equal(t, "tracker-today", doc.Data.Type)
	assert.Equal(t, "2026-04-08", doc.Data.Attributes.Date)
	require.Len(t, doc.Data.Relationships.Items.Data, 1)
	assert.Equal(t, "trackers", doc.Data.Relationships.Items.Data[0].Type)
	assert.NotNil(t, doc.Data.Relationships.Entries.Data, "entries.data must be a non-nil slice so JSON marshals to []")

	body, err := MarshalDocument(doc)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"type":"tracker-today"`)
	assert.Contains(t, string(body), `"date":"2026-04-08"`)
}
