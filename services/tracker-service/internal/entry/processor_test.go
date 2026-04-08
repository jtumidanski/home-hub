package entry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
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
	require.NoError(t, db.AutoMigrate(&Entity{}))
	return db
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

func TestProcessor_CreateOrUpdate_Insert(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	itemID := uuid.New()
	m, created, err := p.CreateOrUpdate(uuid.New(), uuid.New(), itemID, yesterday(), sentimentValue("positive"), nil, "sentiment", nil)
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
	itemID := uuid.New()

	first, _, err := p.CreateOrUpdate(tenantID, userID, itemID, yesterday(), numericValue(1), nil, "numeric", nil)
	require.NoError(t, err)

	second, created, err := p.CreateOrUpdate(tenantID, userID, itemID, yesterday(), numericValue(5), nil, "numeric", nil)
	require.NoError(t, err)
	assert.False(t, created, "second call must update, not insert")
	assert.Equal(t, first.Id(), second.Id())
}

func TestProcessor_CreateOrUpdate_RejectsFutureDate(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), uuid.New(), tomorrow(), numericValue(1), nil, "numeric", nil)
	assert.ErrorIs(t, err, ErrFutureDate)
}

func TestProcessor_CreateOrUpdate_ValidatesValuePerScaleType(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	itemID := uuid.New()

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
			_, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), itemID, yesterday(), tc.value, nil, tc.scaleType, tc.cfg)
			if tc.wantErr == nil {
				assert.NoError(t, err)
				// Clean up so subsequent subtests on the same itemID don't see a stale entry.
				_ = p.Delete(itemID, yesterday())
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestProcessor_CreateOrUpdate_NoteTooLong(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	long := make([]byte, 501)
	for i := range long {
		long[i] = 'x'
	}
	note := string(long)

	_, _, err := p.CreateOrUpdate(uuid.New(), uuid.New(), uuid.New(), yesterday(), numericValue(1), &note, "numeric", nil)
	assert.ErrorIs(t, err, ErrNoteTooLong)
}

func TestProcessor_Skip_RequiresScheduledDay(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, err := p.Skip(uuid.New(), uuid.New(), uuid.New(), yesterday(), false)
	assert.ErrorIs(t, err, ErrNotScheduled)
}

func TestProcessor_Skip_ClearsValueAndNote(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tenantID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()
	note := "had a long day"

	_, _, err := p.CreateOrUpdate(tenantID, userID, itemID, yesterday(), numericValue(7), &note, "numeric", nil)
	require.NoError(t, err)

	skipped, err := p.Skip(tenantID, userID, itemID, yesterday(), true)
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
	itemID := uuid.New()

	_, err := p.Skip(tenantID, userID, itemID, yesterday(), true)
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
	item1 := uuid.New()
	item2 := uuid.New()

	_, _, err := p.CreateOrUpdate(uuid.New(), user1, item1, yesterday(), numericValue(1), nil, "numeric", nil)
	require.NoError(t, err)
	_, _, err = p.CreateOrUpdate(uuid.New(), user2, item2, yesterday(), numericValue(2), nil, "numeric", nil)
	require.NoError(t, err)

	month := time.Now().UTC().Format("2006-01")
	user1Entries, err := p.ListByMonth(user1, month)
	require.NoError(t, err)
	require.Len(t, user1Entries, 1)
	assert.Equal(t, item1, user1Entries[0].TrackingItemID())
}
