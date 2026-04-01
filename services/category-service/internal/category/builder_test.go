package category

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name      string
		setup     func() *Builder
		wantErr   error
		validate  func(t *testing.T, m Model)
	}{
		{
			name:    "requires name",
			setup:   func() *Builder { return NewBuilder() },
			wantErr: ErrNameRequired,
		},
		{
			name: "rejects name over 100 characters",
			setup: func() *Builder {
				return NewBuilder().SetName(strings.Repeat("a", 101))
			},
			wantErr: ErrNameTooLong,
		},
		{
			name: "rejects negative sort order",
			setup: func() *Builder {
				return NewBuilder().SetName("Test").SetSortOrder(-1)
			},
			wantErr: ErrInvalidSortOrder,
		},
		{
			name: "success with all fields",
			setup: func() *Builder {
				return NewBuilder().
					SetId(id).
					SetTenantID(tenantID).
					SetName("Produce").
					SetSortOrder(1)
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, id, m.Id())
				assert.Equal(t, tenantID, m.TenantID())
				assert.Equal(t, "Produce", m.Name())
				assert.Equal(t, 1, m.SortOrder())
			},
		},
		{
			name: "accepts max length name",
			setup: func() *Builder {
				return NewBuilder().SetName(strings.Repeat("a", 100))
			},
			validate: func(t *testing.T, m Model) {
				assert.Len(t, m.Name(), 100)
			},
		},
		{
			name: "accepts zero sort order",
			setup: func() *Builder {
				return NewBuilder().SetName("Test").SetSortOrder(0)
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, 0, m.SortOrder())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.setup().Build()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			if tc.validate != nil {
				tc.validate(t, m)
			}
		})
	}
}
