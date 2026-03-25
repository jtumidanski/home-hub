package task

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	dueOn := now.Add(24 * time.Hour)
	completedAt := now.Add(1 * time.Hour)
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	completedByUID := uuid.New()

	tests := []struct {
		name        string
		setup       func() *Builder
		wantErr     error
		assertModel func(t *testing.T, m Model)
	}{
		{
			name: "valid build with all fields",
			setup: func() *Builder {
				return NewBuilder().
					SetId(id).
					SetTenantID(tenantID).
					SetHouseholdID(householdID).
					SetTitle("Buy groceries").
					SetNotes("Milk, eggs, bread").
					SetStatus("in_progress").
					SetDueOn(&dueOn).
					SetRolloverEnabled(true).
					SetCompletedAt(&completedAt).
					SetCompletedByUID(&completedByUID).
					SetCreatedAt(now).
					SetUpdatedAt(now)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, id, m.id)
				require.Equal(t, tenantID, m.tenantID)
				require.Equal(t, householdID, m.householdID)
				require.Equal(t, "Buy groceries", m.title)
				require.Equal(t, "Milk, eggs, bread", m.notes)
				require.Equal(t, "in_progress", m.status)
				require.Equal(t, &dueOn, m.dueOn)
				require.True(t, m.rolloverEnabled)
				require.Equal(t, &completedAt, m.completedAt)
				require.Equal(t, &completedByUID, m.completedByUID)
				require.Equal(t, now, m.createdAt)
				require.Equal(t, now, m.updatedAt)
			},
		},
		{
			name: "empty title returns ErrTitleRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetTitle("")
			},
			wantErr: ErrTitleRequired,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "default status is pending",
			setup: func() *Builder {
				return NewBuilder().
					SetTitle("Take out trash")
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, "pending", m.status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			model, err := builder.Build()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			tt.assertModel(t, model)
		})
	}
}
