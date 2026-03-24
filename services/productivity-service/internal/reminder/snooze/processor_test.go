package snooze

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test db")

	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)

	require.NoError(t, db.AutoMigrate(&reminder.Entity{}))
	require.NoError(t, db.AutoMigrate(&Entity{}))

	return db
}

func createReminder(t *testing.T, db *gorm.DB) reminder.Model {
	t.Helper()
	l, _ := test.NewNullLogger()
	p := reminder.NewProcessor(l, context.Background(), db)
	past := time.Now().UTC().Add(-1 * time.Hour)
	m, err := p.Create(uuid.New(), uuid.New(), "Test Reminder", "Notes", past)
	require.NoError(t, err, "failed to create reminder")
	return m
}

func TestProcessor_Create(t *testing.T) {
	tests := []struct {
		name            string
		durationMinutes int
		setupReminder   bool
		wantErr         bool
	}{
		{
			name:            "successful snooze with valid duration (30 minutes)",
			durationMinutes: 30,
			setupReminder:   true,
			wantErr:         false,
		},
		{
			name:            "invalid duration (15 minutes) returns error",
			durationMinutes: 15,
			setupReminder:   true,
			wantErr:         true,
		},
		{
			name:            "snooze of non-existent reminder returns error",
			durationMinutes: 15,
			setupReminder:   false,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, context.Background(), db)

			tenantID := uuid.New()
			householdID := uuid.New()
			userID := uuid.New()

			var reminderID uuid.UUID
			if tt.setupReminder {
				m := createReminder(t, db)
				reminderID = m.Id()
			} else {
				reminderID = uuid.New()
			}

			m, err := p.Create(tenantID, householdID, reminderID, userID, tt.durationMinutes)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, Model{}, m)
			} else {
				require.NoError(t, err)
				require.Equal(t, reminderID, m.ReminderID())
				require.Equal(t, tt.durationMinutes, m.DurationMinutes())
				require.Equal(t, tenantID, m.TenantID())
				require.Equal(t, householdID, m.HouseholdID())
				require.Equal(t, userID, m.CreatedByUserID())
				require.True(t, m.SnoozedUntil().After(time.Now().UTC()), "snoozedUntil should be in the future")
				require.NotEqual(t, uuid.Nil, m.Id())
			}
		})
	}
}
