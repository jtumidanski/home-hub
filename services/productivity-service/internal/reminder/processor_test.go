package reminder

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
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

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	scheduled := time.Now().UTC().Add(1 * time.Hour)
	ownerID := uuid.New()
	m, err := p.Create(uuid.New(), uuid.New(), "Test Reminder", "Notes", scheduled, &ownerID)
	require.NoError(t, err)
	require.Equal(t, "Test Reminder", m.Title())
	require.Equal(t, "Notes", m.Notes())
	require.Equal(t, &ownerID, m.OwnerUserID())
}

func TestIsActive(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tests := []struct {
		name         string
		scheduledFor time.Time
		expectActive bool
	}{
		{"future reminder not active", time.Now().UTC().Add(1 * time.Hour), false},
		{"past reminder is active", time.Now().UTC().Add(-1 * time.Hour), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.Create(uuid.New(), uuid.New(), tc.name, "", tc.scheduledFor, nil)
			require.NoError(t, err)
			require.Equal(t, tc.expectActive, m.IsActive())
		})
	}
}

func TestSnooze(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tests := []struct {
		name            string
		durationMinutes int
		expectError     bool
	}{
		{"valid 10 minutes", 10, false},
		{"valid 30 minutes", 30, false},
		{"valid 60 minutes", 60, false},
		{"invalid 15 minutes", 15, true},
		{"invalid 0 minutes", 0, true},
		{"invalid 45 minutes", 45, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			past := time.Now().UTC().Add(-1 * time.Hour)
			m, err := p.Create(uuid.New(), uuid.New(), tc.name, "", past, nil)
			require.NoError(t, err)

			snoozedUntil, err := p.Snooze(m.Id(), tc.durationMinutes)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, snoozedUntil.After(time.Now().UTC()))
			}
		})
	}
}

func TestDismiss(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	past := time.Now().UTC().Add(-1 * time.Hour)
	m, err := p.Create(uuid.New(), uuid.New(), "Dismiss Me", "", past, nil)
	require.NoError(t, err)
	require.True(t, m.IsActive())

	require.NoError(t, p.Dismiss(m.Id()))

	dismissed, err := p.ByIDProvider(m.Id())()
	require.NoError(t, err)
	require.False(t, dismissed.IsActive())
}
