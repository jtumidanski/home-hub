package month

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

// All tests use January 2025 (a fully-past month relative to today's date in
// the test environment) so completion math is unaffected by `time.Now()`.
const testMonth = "2025-01"

var monthStart = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
var monthEnd = time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

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

// seedItem creates a tracking item directly in the DB with full control over
// timestamps and writes a single schedule snapshot effective on snapshotDate.
func seedItem(t *testing.T, db *gorm.DB, userID uuid.UUID, name, scaleType string, scaleConfig json.RawMessage, sched []int, createdAt time.Time, deletedAt *time.Time, snapshotDate time.Time) trackingitem.Entity {
	t.Helper()
	item := trackingitem.Entity{
		Id:          uuid.New(),
		TenantId:    uuid.New(),
		UserId:      userID,
		Name:        name,
		ScaleType:   scaleType,
		ScaleConfig: scaleConfig,
		Color:       "blue",
		SortOrder:   1,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		DeletedAt:   deletedAt,
	}
	require.NoError(t, db.Create(&item).Error)

	schedJSON, _ := json.Marshal(sched)
	snap := schedule.Entity{
		Id:             uuid.New(),
		TrackingItemId: item.Id,
		Schedule:       schedJSON,
		EffectiveDate:  snapshotDate,
		CreatedAt:      createdAt,
	}
	require.NoError(t, db.Create(&snap).Error)
	return item
}

// addSnapshot writes an additional schedule snapshot for an existing item.
func addSnapshot(t *testing.T, db *gorm.DB, itemID uuid.UUID, sched []int, effectiveDate time.Time) {
	t.Helper()
	schedJSON, _ := json.Marshal(sched)
	snap := schedule.Entity{
		Id:             uuid.New(),
		TrackingItemId: itemID,
		Schedule:       schedJSON,
		EffectiveDate:  effectiveDate,
		CreatedAt:      effectiveDate,
	}
	require.NoError(t, db.Create(&snap).Error)
}

func seedEntry(t *testing.T, db *gorm.DB, item trackingitem.Entity, date time.Time, value json.RawMessage, skipped bool) {
	t.Helper()
	e := entry.Entity{
		Id:             uuid.New(),
		TenantId:       item.TenantId,
		UserId:         item.UserId,
		TrackingItemId: item.Id,
		Date:           date,
		Value:          value,
		Skipped:        skipped,
		CreatedAt:      date,
		UpdatedAt:      date,
	}
	require.NoError(t, db.Create(&e).Error)
}

func numericValue(count int) json.RawMessage {
	b, _ := json.Marshal(map[string]int{"count": count})
	return b
}

func sentimentValue(rating string) json.RawMessage {
	b, _ := json.Marshal(map[string]string{"rating": rating})
	return b
}

func rangeValue(v int) json.RawMessage {
	b, _ := json.Marshal(map[string]int{"value": v})
	return b
}

func TestComputeMonthSummary_AllFilledMarksComplete(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	item := seedItem(t, db, userID, "Daily", "numeric", nil, []int{}, monthStart.AddDate(0, 0, -7), nil, monthStart.AddDate(0, 0, -7))
	for d := monthStart; !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
		seedEntry(t, db, item, d, numericValue(1), false)
	}

	summary, items, entries, err := p.ComputeMonthSummary(userID, testMonth)
	require.NoError(t, err)
	assert.True(t, summary.Complete)
	assert.Equal(t, 31, summary.Completion.Expected)
	assert.Equal(t, 31, summary.Completion.Filled)
	assert.Equal(t, 0, summary.Completion.Skipped)
	assert.Equal(t, 0, summary.Completion.Remaining)
	assert.Len(t, items, 1)
	assert.Len(t, entries, 31)
}

func TestComputeMonthSummary_PartialNotComplete(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	item := seedItem(t, db, userID, "Daily", "numeric", nil, []int{}, monthStart.AddDate(0, 0, -7), nil, monthStart.AddDate(0, 0, -7))
	for i := 0; i < 20; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), numericValue(2), false)
	}
	// 5 skips
	for i := 20; i < 25; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), nil, true)
	}

	summary, _, _, err := p.ComputeMonthSummary(userID, testMonth)
	require.NoError(t, err)
	assert.False(t, summary.Complete)
	assert.Equal(t, 31, summary.Completion.Expected)
	assert.Equal(t, 20, summary.Completion.Filled)
	assert.Equal(t, 5, summary.Completion.Skipped)
	assert.Equal(t, 6, summary.Completion.Remaining)
}

func TestComputeMonthSummary_MidMonthCreateLimitsExpectedRange(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	// Item created on Jan 15, 2025: only days 15..31 should count (17 days).
	createdAt := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	seedItem(t, db, userID, "Mid-month", "numeric", nil, []int{}, createdAt, nil, createdAt)

	summary, _, _, err := p.ComputeMonthSummary(userID, testMonth)
	require.NoError(t, err)
	assert.Equal(t, 17, summary.Completion.Expected)
}

func TestComputeMonthSummary_MidMonthDeleteCapsExpectedRange(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	// Item active Dec 25 .. Jan 15 inclusive: only Jan 1..15 (15 days) should count.
	createdAt := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	item := seedItem(t, db, userID, "Was active", "numeric", nil, []int{}, createdAt, &deletedAt, createdAt)
	// At least one historical entry within the month so the deleted item is
	// preserved in the summary (a delete-and-recreate with no entries should
	// drop the ghost row — see the no-entries test below).
	seedEntry(t, db, item, time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), numericValue(1), false)

	summary, items, _, err := p.ComputeMonthSummary(userID, testMonth)
	require.NoError(t, err)
	assert.Equal(t, 15, summary.Completion.Expected)
	assert.Len(t, items, 1, "soft-deleted items with historical entries this month must still appear")
}

func TestComputeMonthSummary_DeletedItemWithoutEntriesIsDropped(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	// Soft-deleted item whose lifetime overlaps the month but never logged any
	// entries this month — the calendar must not render a ghost row for it.
	createdAt := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	seedItem(t, db, userID, "Deleted, no entries", "numeric", nil, []int{}, createdAt, &deletedAt, createdAt)

	summary, items, _, err := p.ComputeMonthSummary(userID, testMonth)
	require.NoError(t, err)
	assert.Empty(t, items, "soft-deleted items with no entries this month must be dropped")
	assert.Equal(t, 0, summary.Completion.Expected)
}

func TestComputeMonthSummary_ScheduleChangeUsesPriorSnapshotForEarlierDays(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	// Initial snapshot: every day, effective Dec 25 2024.
	createdAt := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	item := seedItem(t, db, userID, "Schedule change", "numeric", nil, []int{}, createdAt, nil, createdAt)

	// Snapshot 2: Sundays only, effective Jan 15 2025.
	addSnapshot(t, db, item.Id, []int{0}, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))

	summary, _, _, err := p.ComputeMonthSummary(userID, testMonth)
	require.NoError(t, err)
	// Jan 1..14 = 14 days under "every day" snapshot.
	// Jan 15..31 with Sunday-only: Jan 19, Jan 26 = 2 days.
	assert.Equal(t, 16, summary.Completion.Expected)
}

func TestComputeMonthSummary_InvalidMonthFormat(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, _, _, err := p.ComputeMonthSummary(uuid.New(), "not-a-month")
	assert.ErrorIs(t, err, ErrInvalidMonth)
}

func TestComputeReport_RejectsIncompleteMonth(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	// Daily item, only one entry — month is not complete.
	item := seedItem(t, db, userID, "Daily", "numeric", nil, []int{}, monthStart.AddDate(0, 0, -7), nil, monthStart.AddDate(0, 0, -7))
	seedEntry(t, db, item, monthStart, numericValue(1), false)

	_, err := p.ComputeReport(userID, testMonth)
	assert.ErrorIs(t, err, ErrMonthIncomplete)
}

func TestComputeReport_NumericStats(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	item := seedItem(t, db, userID, "Drinks", "numeric", nil, []int{}, monthStart.AddDate(0, 0, -7), nil, monthStart.AddDate(0, 0, -7))
	// 31 entries: 10 days with count=2 (=20), 21 days with count=0.
	for i := 0; i < 31; i++ {
		count := 0
		if i < 10 {
			count = 2
		}
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), numericValue(count), false)
	}

	report, err := p.ComputeReport(userID, testMonth)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)

	var stats NumericStats
	require.NoError(t, json.Unmarshal(report.Items[0].Stats, &stats))
	assert.Equal(t, 31, stats.ExpectedDays)
	assert.Equal(t, 31, stats.FilledDays)
	assert.Equal(t, 20, stats.Total)
	assert.Equal(t, 10, stats.DaysWithEntriesAboveZero)
	assert.InDelta(t, 20.0/31.0, stats.DailyAverage, 0.01)
	require.NotNil(t, stats.Min)
	require.NotNil(t, stats.Max)
	assert.Equal(t, 0, stats.Min.Count)
	assert.Equal(t, 2, stats.Max.Count)
}

func TestComputeReport_SentimentStats(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	item := seedItem(t, db, userID, "Mood", "sentiment", nil, []int{}, monthStart.AddDate(0, 0, -7), nil, monthStart.AddDate(0, 0, -7))
	// 31 entries: 16 positive, 10 neutral, 5 negative.
	for i := 0; i < 16; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), sentimentValue("positive"), false)
	}
	for i := 16; i < 26; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), sentimentValue("neutral"), false)
	}
	for i := 26; i < 31; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), sentimentValue("negative"), false)
	}

	report, err := p.ComputeReport(userID, testMonth)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)

	var stats SentimentStats
	require.NoError(t, json.Unmarshal(report.Items[0].Stats, &stats))
	assert.Equal(t, 16, stats.Positive)
	assert.Equal(t, 10, stats.Neutral)
	assert.Equal(t, 5, stats.Negative)
	assert.InDelta(t, 16.0/31.0, stats.PositiveRatio, 0.01)
}

func TestComputeReport_RangeStats(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	cfg := json.RawMessage(`{"min":0,"max":100}`)
	item := seedItem(t, db, userID, "Sleep", "range", cfg, []int{}, monthStart.AddDate(0, 0, -7), nil, monthStart.AddDate(0, 0, -7))
	// 31 entries with values 70..100, average = 85.
	for i := 0; i < 31; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), rangeValue(70+i), false)
	}

	report, err := p.ComputeReport(userID, testMonth)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)

	var stats RangeStats
	require.NoError(t, json.Unmarshal(report.Items[0].Stats, &stats))
	assert.InDelta(t, 85.0, stats.Average, 0.1)
	require.NotNil(t, stats.Min)
	require.NotNil(t, stats.Max)
	assert.Equal(t, 70, stats.Min.Value)
	assert.Equal(t, 100, stats.Max.Value)
	assert.Greater(t, stats.StdDev, 0.0)
}

func TestComputeReport_SoftDeletedItemAppearsInHistoricalReport(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	// Active Dec 25 2024 .. Jan 31 2025; entries fill all 31 January days.
	createdAt := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)
	item := seedItem(t, db, userID, "Old habit", "numeric", nil, []int{}, createdAt, &deletedAt, createdAt)
	for i := 0; i < 31; i++ {
		seedEntry(t, db, item, monthStart.AddDate(0, 0, i), numericValue(1), false)
	}

	report, err := p.ComputeReport(userID, testMonth)
	require.NoError(t, err)
	require.Len(t, report.Items, 1)
	assert.Equal(t, "Old habit", report.Items[0].Name)
}
