package week

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

// setupTestDB stands up an in-memory SQLite for the week processor. Tenant
// callbacks are registered because every read flows through the shared tenant
// scope. AutoMigrate is sufficient — the production migration only runs the
// same statement.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}))
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	require.NoError(t, err)
	return v
}

func TestProcessor_EnsureExists_NormalizesAndCreates(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()

	// Wednesday — should normalize to Monday 2026-04-06.
	wed := mustParse(t, "2026-04-08")
	e, err := p.EnsureExists(tenantID, userID, wed)
	require.NoError(t, err)
	assert.Equal(t, "2026-04-06", e.WeekStartDate.Format("2006-01-02"))
	assert.NotEqual(t, uuid.Nil, e.Id)
}

func TestProcessor_EnsureExists_IsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	mon := mustParse(t, "2026-04-06")

	first, err := p.EnsureExists(tenantID, userID, mon)
	require.NoError(t, err)

	second, err := p.EnsureExists(tenantID, userID, mon)
	require.NoError(t, err)
	assert.Equal(t, first.Id, second.Id, "second call must return the existing row")
}

func TestProcessor_PatchRestDayFlags(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	mon := mustParse(t, "2026-04-06")

	cases := []struct {
		name    string
		flags   []int
		wantErr error
	}{
		{"empty resets flags", []int{}, nil},
		{"valid set", []int{0, 6}, nil},
		{"out of range high", []int{7}, ErrInvalidRestDay},
		{"out of range low", []int{-1}, ErrInvalidRestDay},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.PatchRestDayFlags(tenantID, userID, mon, tc.flags)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.flags, m.RestDayFlags())
		})
	}
}

func TestProcessor_Get_ReturnsNotFoundForUnseededWeek(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	_, userID := uuid.New(), uuid.New()
	mon := mustParse(t, "2026-04-06")

	_, err := p.Get(userID, mon)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProcessor_Get_NormalizesNonMondayInput(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()

	// Seed via Wednesday so the row's start_date is the normalized Monday.
	_, err := p.EnsureExists(tenantID, userID, mustParse(t, "2026-04-08"))
	require.NoError(t, err)

	// A Sunday input must round back to the same Monday.
	got, err := p.Get(userID, mustParse(t, "2026-04-12"))
	require.NoError(t, err)
	assert.Equal(t, "2026-04-06", got.WeekStartDate().Format("2006-01-02"))
}
