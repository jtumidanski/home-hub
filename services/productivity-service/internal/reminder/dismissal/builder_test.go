package dismissal

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	validID := uuid.New()
	validTenantID := uuid.New()
	validHouseholdID := uuid.New()
	validReminderID := uuid.New()
	validUserID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name    string
		setup   func() *Builder
		wantErr error
	}{
		{
			name: "valid build with all fields",
			setup: func() *Builder {
				return NewBuilder().
					SetId(validID).
					SetTenantID(validTenantID).
					SetHouseholdID(validHouseholdID).
					SetReminderID(validReminderID).
					SetCreatedByUserID(validUserID).
					SetCreatedAt(now)
			},
			wantErr: nil,
		},
		{
			name: "missing reminderID returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetCreatedByUserID(validUserID)
			},
			wantErr: ErrReminderIDRequired,
		},
		{
			name: "missing createdByUserID returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetReminderID(validReminderID)
			},
			wantErr: ErrCreatedByRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.setup().Build()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, Model{}, m)
				return
			}
			require.NoError(t, err)
			require.Equal(t, validReminderID, m.ReminderID())
			require.Equal(t, validUserID, m.CreatedByUserID())
		})
	}
}
