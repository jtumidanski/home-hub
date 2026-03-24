package restoration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
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
	db.AutoMigrate(&task.Entity{})
	db.AutoMigrate(&Entity{})
	return db
}

func TestRestorationProcessor_Create(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name      string
		setup     func(t *testing.T, db *gorm.DB) uuid.UUID
		wantErr   bool
		checkErr  func(t *testing.T, err error)
		checkResult func(t *testing.T, e Entity)
	}{
		{
			name: "successful restoration of a soft-deleted task",
			setup: func(t *testing.T, db *gorm.DB) uuid.UUID {
				l, _ := test.NewNullLogger()
				taskProc := task.NewProcessor(l, context.Background(), db)
				m, err := taskProc.Create(tenantID, householdID, "Restore Me", "", nil, false)
				require.NoError(t, err)
				err = taskProc.Delete(m.Id())
				require.NoError(t, err)
				return m.Id()
			},
			wantErr: false,
			checkResult: func(t *testing.T, e Entity) {
				require.NotEqual(t, uuid.Nil, e.Id)
				require.Equal(t, tenantID, e.TenantId)
				require.Equal(t, householdID, e.HouseholdId)
				require.Equal(t, userID, e.CreatedByUserId)
				require.False(t, e.CreatedAt.IsZero())
			},
		},
		{
			name: "restoration of non-existent task returns error",
			setup: func(t *testing.T, db *gorm.DB) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, task.ErrNotFound)
			},
		},
		{
			name: "restoration of non-deleted task returns ErrNotDeleted",
			setup: func(t *testing.T, db *gorm.DB) uuid.UUID {
				l, _ := test.NewNullLogger()
				taskProc := task.NewProcessor(l, context.Background(), db)
				m, err := taskProc.Create(tenantID, householdID, "Not Deleted", "", nil, false)
				require.NoError(t, err)
				return m.Id()
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, task.ErrNotDeleted)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			taskID := tt.setup(t, db)

			p := NewProcessor(l, context.Background(), db)
			result, err := p.Create(tenantID, householdID, taskID, userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}
