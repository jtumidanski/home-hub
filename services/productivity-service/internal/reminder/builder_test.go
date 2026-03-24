package reminder

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
	validScheduledFor := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		setup       func(b *Builder) *Builder
		wantErr     error
		wantNoErr   bool
	}{
		{
			name: "valid build with all fields",
			setup: func(b *Builder) *Builder {
				return b.
					SetId(validID).
					SetTenantID(validTenantID).
					SetHouseholdID(validHouseholdID).
					SetTitle("Take out the bins").
					SetNotes("Don't forget recycling").
					SetScheduledFor(validScheduledFor)
			},
			wantNoErr: true,
		},
		{
			name: "empty title returns ErrTitleRequired",
			setup: func(b *Builder) *Builder {
				return b.
					SetTitle("").
					SetScheduledFor(validScheduledFor)
			},
			wantErr: ErrTitleRequired,
		},
		{
			name: "zero scheduledFor returns ErrScheduledForRequired",
			setup: func(b *Builder) *Builder {
				return b.
					SetTitle("Take out the bins").
					SetScheduledFor(time.Time{})
			},
			wantErr: ErrScheduledForRequired,
		},
		{
			name: "both title and scheduledFor missing returns ErrTitleRequired first",
			setup: func(b *Builder) *Builder {
				return b.
					SetTitle("").
					SetScheduledFor(time.Time{})
			},
			wantErr: ErrTitleRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setup(NewBuilder())
			model, err := b.Build()

			if tt.wantNoErr {
				require.NoError(t, err)
				require.NotEqual(t, Model{}, model)
				return
			}

			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, Model{}, model)
		})
	}
}
