package list

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (Model, error)
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name:    "requires name",
			build:   func() (Model, error) { return NewBuilder().Build() },
			wantErr: ErrNameRequired,
		},
		{
			name: "rejects long name",
			build: func() (Model, error) {
				return NewBuilder().SetName(strings.Repeat("a", 256)).Build()
			},
			wantErr: ErrNameTooLong,
		},
		{
			name: "success with all fields",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetName("Weekly Groceries").
					SetCreatedBy(uuid.New()).
					Build()
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Weekly Groceries", m.Name())
				assert.Equal(t, "active", m.Status())
				assert.Nil(t, m.ArchivedAt())
				assert.False(t, m.IsArchived())
			},
		},
		{
			name: "archived status",
			build: func() (Model, error) {
				return NewBuilder().
					SetName("Old List").
					SetStatus("archived").
					Build()
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "archived", m.Status())
				assert.True(t, m.IsArchived())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.build()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.validate(t, m)
		})
	}
}
