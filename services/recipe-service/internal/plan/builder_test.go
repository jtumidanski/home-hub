package plan

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	createdBy := uuid.New()
	startsOn := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)

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
					SetStartsOn(startsOn).
					SetName("My Meal Plan").
					SetLocked(true).
					SetCreatedBy(createdBy).
					SetCreatedAt(now).
					SetUpdatedAt(now)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, id, m.Id())
				assert.Equal(t, tenantID, m.TenantID())
				assert.Equal(t, householdID, m.HouseholdID())
				assert.Equal(t, startsOn, m.StartsOn())
				assert.Equal(t, "My Meal Plan", m.Name())
				assert.True(t, m.Locked())
				assert.Equal(t, createdBy, m.CreatedBy())
				assert.Equal(t, now, m.CreatedAt())
				assert.Equal(t, now, m.UpdatedAt())
			},
		},
		{
			name: "minimal valid build with auto-generated name",
			setup: func() *Builder {
				return NewBuilder().
					SetStartsOn(startsOn)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, "Week of March 30, 2026", m.Name())
				assert.Equal(t, startsOn, m.StartsOn())
				assert.False(t, m.Locked())
				assert.Equal(t, uuid.Nil, m.Id())
			},
		},
		{
			name: "empty name gets auto-generated",
			setup: func() *Builder {
				return NewBuilder().
					SetStartsOn(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)).
					SetName("")
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, "Week of January 5, 2026", m.Name())
			},
		},
		{
			name: "explicit name preserved",
			setup: func() *Builder {
				return NewBuilder().
					SetStartsOn(startsOn).
					SetName("Custom Name")
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, "Custom Name", m.Name())
			},
		},
		{
			name: "zero starts_on returns ErrStartsOnRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetName("Test Plan")
			},
			wantErr: ErrStartsOnRequired,
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
			if tt.assertModel != nil {
				tt.assertModel(t, model)
			}
		})
	}
}
