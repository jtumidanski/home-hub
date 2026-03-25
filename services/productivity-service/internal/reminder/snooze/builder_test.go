package snooze

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
	snoozedUntil := now.Add(30 * time.Minute)

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
					SetDurationMinutes(30).
					SetSnoozedUntil(snoozedUntil).
					SetCreatedByUserID(validUserID).
					SetCreatedAt(now)
			},
			wantErr: nil,
		},
		{
			name: "missing reminderID returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetCreatedByUserID(validUserID).
					SetDurationMinutes(30).
					SetSnoozedUntil(snoozedUntil)
			},
			wantErr: ErrReminderIDRequired,
		},
		{
			name: "missing createdByUserID returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetReminderID(validReminderID).
					SetDurationMinutes(30).
					SetSnoozedUntil(snoozedUntil)
			},
			wantErr: ErrCreatedByRequired,
		},
		{
			name: "zero durationMinutes returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetReminderID(validReminderID).
					SetCreatedByUserID(validUserID).
					SetSnoozedUntil(snoozedUntil)
			},
			wantErr: ErrDurationMinutesRequired,
		},
		{
			name: "zero snoozedUntil returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetReminderID(validReminderID).
					SetCreatedByUserID(validUserID).
					SetDurationMinutes(30)
			},
			wantErr: ErrSnoozedUntilRequired,
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
			require.Equal(t, 30, m.DurationMinutes())
			require.Equal(t, validUserID, m.CreatedByUserID())
		})
	}
}
