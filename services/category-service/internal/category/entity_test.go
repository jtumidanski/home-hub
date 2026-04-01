package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		entity   Entity
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name: "success with all fields",
			entity: Entity{
				Id:        id,
				TenantId:  tenantID,
				Name:      "Produce",
				SortOrder: 1,
				CreatedAt: now,
				UpdatedAt: now,
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, id, m.Id())
				assert.Equal(t, tenantID, m.TenantID())
				assert.Equal(t, "Produce", m.Name())
				assert.Equal(t, 1, m.SortOrder())
				assert.Equal(t, now, m.CreatedAt())
				assert.Equal(t, now, m.UpdatedAt())
			},
		},
		{
			name:   "zero sort order",
			entity: Entity{Name: "Other", SortOrder: 0},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, 0, m.SortOrder())
			},
		},
		{
			name:    "empty name returns error",
			entity:  Entity{Name: ""},
			wantErr: ErrNameRequired,
		},
		{
			name:    "name too long returns error",
			entity:  Entity{Name: string(make([]byte, 101))},
			wantErr: ErrNameTooLong,
		},
		{
			name:    "negative sort order returns error",
			entity:  Entity{Name: "Test", SortOrder: -1},
			wantErr: ErrInvalidSortOrder,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := Make(tc.entity)
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

func TestModel_ToEntity_RoundTrip(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().Truncate(time.Second)

	original := Entity{
		Id:        id,
		TenantId:  tenantID,
		Name:      "Dairy & Eggs",
		SortOrder: 3,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m, err := Make(original)
	assert.NoError(t, err)

	roundTripped := m.ToEntity()
	assert.Equal(t, original.Id, roundTripped.Id)
	assert.Equal(t, original.TenantId, roundTripped.TenantId)
	assert.Equal(t, original.Name, roundTripped.Name)
	assert.Equal(t, original.SortOrder, roundTripped.SortOrder)
	assert.Equal(t, original.CreatedAt, roundTripped.CreatedAt)
	assert.Equal(t, original.UpdatedAt, roundTripped.UpdatedAt)
}
