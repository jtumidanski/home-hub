package entry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
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
	require.NoError(t, db.AutoMigrate(&Entity{}, &trackingitem.Entity{}, &schedule.Entity{}))
	return db
}

// makeItem inserts a tracking item with the given scale type/config so the
// entry processor can resolve it via the cross-domain provider during writes.
// Returns the new item's id.
func makeItem(t *testing.T, db *gorm.DB, scaleType string, scaleConfig json.RawMessage) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&trackingitem.Entity{
		Id:          id,
		TenantId:    uuid.New(),
		UserId:      uuid.New(),
		Name:        "test-" + id.String()[:8],
		ScaleType:   scaleType,
		ScaleConfig: scaleConfig,
		Color:       "blue",
	}).Error)
	return id
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func sentimentValue(rating string) json.RawMessage {
	b, _ := json.Marshal(map[string]string{"rating": rating})
	return b
}

func numericValue(count int) json.RawMessage {
	b, _ := json.Marshal(map[string]int{"count": count})
	return b
}

func rangeValue(v int) json.RawMessage {
	b, _ := json.Marshal(map[string]int{"value": v})
	return b
}

func rangeConfig(min, max int) json.RawMessage {
	b, _ := json.Marshal(map[string]int{"min": min, "max": max})
	return b
}

func yesterday() string {
	return time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
}

func tomorrow() string {
	return time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
}

func todayUTC() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// scheduleEverydayFor inserts a snapshot effective today so the entry
// processor's "scheduled" lookup treats every day as scheduled.
func scheduleEverydayFor(t *testing.T, db *gorm.DB, itemID uuid.UUID) {
	t.Helper()
	_, err := schedule.CreateSnapshot(db, itemID, []int{}, time.Now().UTC().AddDate(0, -1, 0))
	require.NoError(t, err)
}

func TestProcessor_CreateOrUpdate_Insert(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	itemID := makeItem(t, db, "sentiment", nil)
	m, created, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), itemID, yesterday(), todayUTC(), sentimentValue("positive"), nil)
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, itemID, m.TrackingItemID())
	assert.False(t, m.Skipped())
}

func TestProcessor_CreateOrUpdate_UpdatePreservesIDClearsSkip(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tenantID := uuid.New()
	userID := uuid.New()
	itemID := makeItem(t, db, "numeric", nil)

	first, _, _, err := p.CreateOrUpdate(tenantID, userID, itemID, yesterday(), todayUTC(), numericValue(1), nil)
	require.NoError(t, err)

	second, created, _, err := p.CreateOrUpdate(tenantID, userID, itemID, yesterday(), todayUTC(), numericValue(5), nil)
	require.NoError(t, err)
	assert.False(t, created, "second call must update, not insert")
	assert.Equal(t, first.Id(), second.Id())
}

func TestProcessor_CreateOrUpdate_RejectsFutureDate(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	itemID := makeItem(t, db, "numeric", nil)
	_, _, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), itemID, tomorrow(), todayUTC(), numericValue(1), nil)
	assert.ErrorIs(t, err, ErrFutureDate)
}

func TestProcessor_CreateOrUpdate_RejectsMissingItem(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, _, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), uuid.New(), yesterday(), todayUTC(), numericValue(1), nil)
	assert.ErrorIs(t, err, ErrItemNotFound)
}

func TestProcessor_CreateOrUpdate_ValidatesValuePerScaleType(t *testing.T) {
	tests := []struct {
		name      string
		scaleType string
		cfg       json.RawMessage
		value     json.RawMessage
		wantErr   error
	}{
		{"sentiment ok", "sentiment", nil, sentimentValue("positive"), nil},
		{"sentiment bad rating", "sentiment", nil, sentimentValue("amazing"), ErrInvalidSentiment},
		{"numeric ok", "numeric", nil, numericValue(3), nil},
		{"numeric negative", "numeric", nil, numericValue(-1), ErrInvalidNumeric},
		{"range in bounds", "range", rangeConfig(0, 100), rangeValue(50), nil},
		{"range below min", "range", rangeConfig(0, 100), rangeValue(-1), ErrInvalidRange},
		{"range above max", "range", rangeConfig(0, 100), rangeValue(101), ErrInvalidRange},
		{"missing value", "numeric", nil, nil, ErrValueRequired},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			p := newTestProcessor(t, db)
			itemID := makeItem(t, db, tc.scaleType, tc.cfg)

			_, _, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), itemID, yesterday(), todayUTC(), tc.value, nil)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestProcessor_CreateOrUpdate_NoteTooLong(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	itemID := makeItem(t, db, "numeric", nil)

	long := make([]byte, 501)
	for i := range long {
		long[i] = 'x'
	}
	note := string(long)

	_, _, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), itemID, yesterday(), todayUTC(), numericValue(1), &note)
	assert.ErrorIs(t, err, ErrNoteTooLong)
}

func TestProcessor_Skip_RequiresScheduledDay(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	itemID := makeItem(t, db, "numeric", nil)
	// Snapshot the schedule to a specific day-of-week that won't match yesterday.
	yesterdayDow := int(time.Now().UTC().Add(-24 * time.Hour).Weekday())
	otherDow := (yesterdayDow + 3) % 7
	_, err := schedule.CreateSnapshot(db, itemID, []int{otherDow}, time.Now().UTC().AddDate(0, -1, 0))
	require.NoError(t, err)

	_, err = p.Skip(uuid.New(), uuid.New(), itemID, yesterday(), todayUTC())
	assert.ErrorIs(t, err, ErrNotScheduled)
}

func TestProcessor_Skip_ClearsValueAndNote(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tenantID := uuid.New()
	userID := uuid.New()
	itemID := makeItem(t, db, "numeric", nil)
	scheduleEverydayFor(t, db, itemID)
	note := "had a long day"

	_, _, _, err := p.CreateOrUpdate(tenantID, userID, itemID, yesterday(), todayUTC(), numericValue(7), &note)
	require.NoError(t, err)

	skipped, err := p.Skip(tenantID, userID, itemID, yesterday(), todayUTC())
	require.NoError(t, err)
	assert.True(t, skipped.Skipped())
	assert.Nil(t, skipped.Note())
	assert.Empty(t, skipped.Value())
}

func TestProcessor_RemoveSkip_DeletesSkippedEntry(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tenantID := uuid.New()
	userID := uuid.New()
	itemID := makeItem(t, db, "numeric", nil)
	scheduleEverydayFor(t, db, itemID)

	_, err := p.Skip(tenantID, userID, itemID, yesterday(), todayUTC())
	require.NoError(t, err)

	require.NoError(t, p.RemoveSkip(itemID, yesterday()))

	models, err := p.ListByMonth(userID, time.Now().UTC().Format("2006-01"))
	require.NoError(t, err)
	for _, m := range models {
		assert.NotEqual(t, itemID, m.TrackingItemID(), "entry should have been removed")
	}
}

func TestProcessor_Delete_IsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	require.NoError(t, p.Delete(uuid.New(), yesterday()))
}

func TestProcessor_ListByMonth_OnlyReturnsRequestedUserAndMonth(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	user1 := uuid.New()
	user2 := uuid.New()
	item1 := makeItem(t, db, "numeric", nil)
	item2 := makeItem(t, db, "numeric", nil)

	_, _, _, err := p.CreateOrUpdate(uuid.New(), user1, item1, yesterday(), todayUTC(), numericValue(1), nil)
	require.NoError(t, err)
	_, _, _, err = p.CreateOrUpdate(uuid.New(), user2, item2, yesterday(), todayUTC(), numericValue(2), nil)
	require.NoError(t, err)

	// Query for yesterday's month — guards against running on the 1st of a
	// month, when yesterday's entries land in the previous calendar month.
	month := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01")
	user1Entries, err := p.ListByMonth(user1, month)
	require.NoError(t, err)
	require.Len(t, user1Entries, 1)
	assert.Equal(t, item1, user1Entries[0].TrackingItemID())
}

func TestProcessor_ListByMonthWithScheduled_PairsScheduleProjection(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	userID := uuid.New()
	scheduledItem := makeItem(t, db, "numeric", nil)
	unscheduledItem := makeItem(t, db, "numeric", nil)

	// scheduledItem has empty schedule (= every day); unscheduledItem has a
	// schedule that excludes yesterday.
	scheduleEverydayFor(t, db, scheduledItem)
	yesterdayDow := int(time.Now().UTC().Add(-24 * time.Hour).Weekday())
	otherDow := (yesterdayDow + 3) % 7
	_, err := schedule.CreateSnapshot(db, unscheduledItem, []int{otherDow}, time.Now().UTC().AddDate(0, -1, 0))
	require.NoError(t, err)

	_, _, _, err = p.CreateOrUpdate(uuid.New(), userID, scheduledItem, yesterday(), todayUTC(), numericValue(1), nil)
	require.NoError(t, err)
	_, _, _, err = p.CreateOrUpdate(uuid.New(), userID, unscheduledItem, yesterday(), todayUTC(), numericValue(2), nil)
	require.NoError(t, err)

	// Query for yesterday's month — see TestProcessor_ListByMonth_OnlyReturnsRequestedUserAndMonth.
	results, err := p.ListByMonthWithScheduled(userID, time.Now().UTC().Add(-24*time.Hour).Format("2006-01"))
	require.NoError(t, err)
	require.Len(t, results, 2)

	scheduledByItem := make(map[uuid.UUID]bool)
	for _, r := range results {
		scheduledByItem[r.Entry.TrackingItemID()] = r.Scheduled
	}
	assert.True(t, scheduledByItem[scheduledItem], "everyday schedule should mark entry as scheduled")
	assert.False(t, scheduledByItem[unscheduledItem], "non-matching DOW should mark entry as not scheduled")
}
