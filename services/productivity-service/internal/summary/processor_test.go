package summary

import (
	"context"
	"testing"

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
