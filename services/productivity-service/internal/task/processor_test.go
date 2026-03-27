package task

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

	ownerID := uuid.New()
	m, err := p.Create(uuid.New(), uuid.New(), "Test Task", "Some notes", nil, false, &ownerID)
	require.NoError(t, err)
	require.Equal(t, "Test Task", m.Title())
	require.Equal(t, "pending", m.Status())
	require.Equal(t, "Some notes", m.Notes())
	require.False(t, m.RolloverEnabled())
	require.Equal(t, &ownerID, m.OwnerUserID())
}

func TestUpdate_StatusTransitions(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	m, err := p.Create(uuid.New(), uuid.New(), "Task", "", nil, false, nil)
	require.NoError(t, err)

	tests := []struct {
		name            string
		newStatus       string
		expectCompleted bool
		expectCompletedAt bool
	}{
		{"complete task", "completed", true, true},
		{"reopen task", "pending", false, false},
	}

	current := m
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updated, err := p.Update(current.Id(), "Task", "", tc.newStatus, nil, false, nil, userID)
			require.NoError(t, err)
			require.Equal(t, tc.expectCompleted, updated.IsCompleted())
			if tc.expectCompletedAt {
				require.NotNil(t, updated.CompletedAt())
			} else {
				require.Nil(t, updated.CompletedAt())
			}
			current = updated
		})
	}
}

func TestSoftDelete_And_Restore(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), "Delete Me", "", nil, false, nil)
	require.NoError(t, err)

	require.NoError(t, p.Delete(m.Id()))

	deleted, err := p.ByIDProvider(m.Id())()
	require.NoError(t, err)
	require.True(t, deleted.IsDeleted())

	require.NoError(t, p.Restore(m.Id()))

	restored, err := p.ByIDProvider(m.Id())()
	require.NoError(t, err)
	require.False(t, restored.IsDeleted())
}

func TestRestore_Errors(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tests := []struct {
		name    string
		setup   func() uuid.UUID
		wantErr error
	}{
		{
			name: "not deleted",
			setup: func() uuid.UUID {
				m, _ := p.Create(uuid.New(), uuid.New(), "Not Deleted", "", nil, false, nil)
				return m.Id()
			},
			wantErr: ErrNotDeleted,
		},
		{
			name: "restore window expired",
			setup: func() uuid.UUID {
				m, _ := p.Create(uuid.New(), uuid.New(), "Old Delete", "", nil, false, nil)
				oldTime := time.Now().UTC().Add(-4 * 24 * time.Hour)
				db.Model(&Entity{}).Where("id = ?", m.Id()).Update("deleted_at", oldTime)
				return m.Id()
			},
			wantErr: ErrRestoreWindow,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.setup()
			err := p.Restore(id)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
