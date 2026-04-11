package today

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
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
	require.NoError(t, db.AutoMigrate(&week.Entity{}))
	return db
}

func TestToday_UsesCallerTimezoneForDateMath(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	// 02:30 UTC on Tuesday 2026-04-07 is 22:30 Monday 2026-04-06 in NYC.
	// The processor must report Monday (dayOfWeek=0), not Tuesday (dayOfWeek=1),
	// and the week start must anchor to Monday 2026-04-06.
	nyc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	localNow := time.Date(2026, 4, 7, 2, 30, 0, 0, time.UTC).In(nyc)
	require.Equal(t, time.Monday, localNow.Weekday())

	res, err := p.Today(uuid.New(), localNow)
	require.NoError(t, err)
	assert.Equal(t, 0, res.DayOfWeek, "Monday in ISO numbering is 0")
	assert.Equal(t, "2026-04-06", res.Date.Format("2006-01-02"))
	assert.Equal(t, "2026-04-06", res.WeekStartDate.Format("2006-01-02"))
	assert.Empty(t, res.Items)
}

func TestToday_UTCInputStillWorks(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC) // Wednesday
	res, err := p.Today(uuid.New(), now)
	require.NoError(t, err)
	assert.Equal(t, 2, res.DayOfWeek, "Wednesday is ISO day 2")
	assert.Equal(t, "2026-04-08", res.Date.Format("2006-01-02"))
	assert.Equal(t, "2026-04-06", res.WeekStartDate.Format("2006-01-02"))
}
