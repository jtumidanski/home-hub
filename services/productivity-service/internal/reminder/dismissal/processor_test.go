package dismissal

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	reminder "github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
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

func TestCreate_SuccessfulDismissal(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	ctx := context.Background()

	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	remProc := reminder.NewProcessor(l, ctx, db)
	scheduled := time.Now().UTC().Add(1 * time.Hour)
	rem, err := remProc.Create(tenantID, householdID, "Test Reminder", "Some notes", scheduled)
	require.NoError(t, err)

	dismissalProc := NewProcessor(l, ctx, db)
	m, err := dismissalProc.Create(tenantID, householdID, rem.Id(), userID)
	require.NoError(t, err)
	require.Equal(t, rem.Id(), m.ReminderID())
	require.Equal(t, tenantID, m.TenantID())
	require.Equal(t, householdID, m.HouseholdID())
	require.Equal(t, userID, m.CreatedByUserID())
	require.NotEqual(t, uuid.Nil, m.Id())
	require.False(t, m.CreatedAt().IsZero())
}

func TestCreate_NonExistentReminder_ReturnsError(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	ctx := context.Background()

	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	nonExistentReminderID := uuid.New()

	dismissalProc := NewProcessor(l, ctx, db)
	_, err := dismissalProc.Create(tenantID, householdID, nonExistentReminderID, userID)
	require.Error(t, err)
}
