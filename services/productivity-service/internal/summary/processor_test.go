package summary

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
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
	require.NoError(t, db.AutoMigrate(&task.Entity{}))
	require.NoError(t, db.AutoMigrate(&reminder.Entity{}))
	return db
}

func TestTaskSummary(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, db *gorm.DB)
		expectPending int64
	}{
		{
			name:          "empty database returns zero pending",
			setup:         func(t *testing.T, db *gorm.DB) {},
			expectPending: 0,
		},
		{
			name: "counts pending tasks",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				p := task.NewProcessor(l, context.Background(), db)
				_, err := p.Create(uuid.New(), uuid.New(), "Task 1", "", nil, false)
				require.NoError(t, err)
				_, err = p.Create(uuid.New(), uuid.New(), "Task 2", "", nil, false)
				require.NoError(t, err)
			},
			expectPending: 2,
		},
		{
			name: "completed tasks are not counted as pending",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				p := task.NewProcessor(l, context.Background(), db)
				m, err := p.Create(uuid.New(), uuid.New(), "Task", "", nil, false)
				require.NoError(t, err)
				_, err = p.Update(m.Id(), "Task", "", "completed", nil, false, uuid.New())
				require.NoError(t, err)
			},
			expectPending: 0,
		},
		{
			name: "soft-deleted tasks are not counted as pending",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				p := task.NewProcessor(l, context.Background(), db)
				m, err := p.Create(uuid.New(), uuid.New(), "Task", "", nil, false)
				require.NoError(t, err)
				require.NoError(t, p.Delete(m.Id()))
			},
			expectPending: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tc.setup(t, db)

			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)
			s, err := proc.TaskSummary()
			require.NoError(t, err)
			require.Equal(t, tc.expectPending, s.PendingCount)
		})
	}
}

func TestReminderSummary(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, db *gorm.DB)
		expectDueNow   int64
		expectUpcoming int64
		expectSnoozed  int64
	}{
		{
			name:           "empty database returns zero counts",
			setup:          func(t *testing.T, db *gorm.DB) {},
			expectDueNow:   0,
			expectUpcoming: 0,
			expectSnoozed:  0,
		},
		{
			name: "counts due-now reminders",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				p := reminder.NewProcessor(l, context.Background(), db)
				past := time.Now().UTC().Add(-1 * time.Hour)
				_, err := p.Create(uuid.New(), uuid.New(), "Past Reminder", "", past)
				require.NoError(t, err)
			},
			expectDueNow:   1,
			expectUpcoming: 0,
			expectSnoozed:  0,
		},
		{
			name: "counts upcoming reminders",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				p := reminder.NewProcessor(l, context.Background(), db)
				future := time.Now().UTC().Add(1 * time.Hour)
				_, err := p.Create(uuid.New(), uuid.New(), "Future Reminder", "", future)
				require.NoError(t, err)
			},
			expectDueNow:   0,
			expectUpcoming: 1,
			expectSnoozed:  0,
		},
		{
			name: "dismissed reminders are not counted",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				p := reminder.NewProcessor(l, context.Background(), db)
				past := time.Now().UTC().Add(-1 * time.Hour)
				m, err := p.Create(uuid.New(), uuid.New(), "Dismissed", "", past)
				require.NoError(t, err)
				require.NoError(t, p.Dismiss(m.Id()))
			},
			expectDueNow:   0,
			expectUpcoming: 0,
			expectSnoozed:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tc.setup(t, db)

			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)
			s, err := proc.ReminderSummary()
			require.NoError(t, err)
			require.Equal(t, tc.expectDueNow, s.DueNowCount)
			require.Equal(t, tc.expectUpcoming, s.UpcomingCount)
			require.Equal(t, tc.expectSnoozed, s.SnoozedCount)
		})
	}
}

func TestDashboardSummary(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T, db *gorm.DB)
		expectPendingTasks int64
		expectDueReminders int64
	}{
		{
			name:               "empty database returns zero counts",
			setup:              func(t *testing.T, db *gorm.DB) {},
			expectPendingTasks: 0,
			expectDueReminders: 0,
		},
		{
			name: "aggregates pending tasks and due reminders",
			setup: func(t *testing.T, db *gorm.DB) {
				l, _ := test.NewNullLogger()
				tp := task.NewProcessor(l, context.Background(), db)
				_, err := tp.Create(uuid.New(), uuid.New(), "Task 1", "", nil, false)
				require.NoError(t, err)
				_, err = tp.Create(uuid.New(), uuid.New(), "Task 2", "", nil, false)
				require.NoError(t, err)

				rp := reminder.NewProcessor(l, context.Background(), db)
				past := time.Now().UTC().Add(-1 * time.Hour)
				_, err = rp.Create(uuid.New(), uuid.New(), "Due Reminder", "", past)
				require.NoError(t, err)
			},
			expectPendingTasks: 2,
			expectDueReminders: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tc.setup(t, db)

			l, _ := test.NewNullLogger()
			proc := NewProcessor(l, context.Background(), db)
			s, err := proc.DashboardSummary()
			require.NoError(t, err)
			require.Equal(t, tc.expectPendingTasks, s.PendingTaskCount)
			require.Equal(t, tc.expectDueReminders, s.DueReminderCount)
			require.False(t, s.GeneratedAt.IsZero())
		})
	}
}
