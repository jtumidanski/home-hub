package item

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMake(t *testing.T) {
	id := uuid.New()
	listID := uuid.New()
	catID := uuid.New()
	qty := "2 lb"
	catName := "Produce"
	sortOrder := 1
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		entity   Entity
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name: "full entity",
			entity: Entity{
				Id: id, ListId: listID, Name: "Chicken breast",
				Quantity: &qty, CategoryId: &catID, CategoryName: &catName,
				CategorySortOrder: &sortOrder, Checked: true, Position: 5,
				CreatedAt: now, UpdatedAt: now,
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, id, m.Id())
				assert.Equal(t, listID, m.ListID())
				assert.Equal(t, "Chicken breast", m.Name())
				assert.Equal(t, &qty, m.Quantity())
				assert.Equal(t, &catID, m.CategoryID())
				assert.Equal(t, &catName, m.CategoryName())
				assert.Equal(t, &sortOrder, m.CategorySortOrder())
				assert.True(t, m.Checked())
				assert.Equal(t, 5, m.Position())
				assert.Equal(t, now, m.CreatedAt())
				assert.Equal(t, now, m.UpdatedAt())
			},
		},
		{
			name:   "minimal entity",
			entity: Entity{Name: "Milk"},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Milk", m.Name())
				assert.Nil(t, m.Quantity())
				assert.Nil(t, m.CategoryID())
				assert.Nil(t, m.CategoryName())
				assert.Nil(t, m.CategorySortOrder())
				assert.False(t, m.Checked())
				assert.Equal(t, 0, m.Position())
			},
		},
		{
			name:    "empty name returns error",
			entity:  Entity{Name: ""},
			wantErr: ErrNameRequired,
		},
		{
			name:    "name too long returns error",
			entity:  Entity{Name: string(make([]byte, 256))},
			wantErr: ErrNameTooLong,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := Make(tc.entity)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.validate(t, m)
		})
	}
}

func TestToEntity(t *testing.T) {
	id := uuid.New()
	listID := uuid.New()
	catID := uuid.New()
	qty := "1 lb"
	catName := "Meat"
	sortOrder := 2
	now := time.Now().Truncate(time.Second)

	m, err := NewBuilder().
		SetId(id).
		SetListID(listID).
		SetName("Steak").
		SetQuantity(&qty).
		SetCategoryID(&catID).
		SetCategoryName(&catName).
		SetCategorySortOrder(&sortOrder).
		SetChecked(true).
		SetPosition(3).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	require.NoError(t, err)

	e := m.ToEntity()
	assert.Equal(t, id, e.Id)
	assert.Equal(t, listID, e.ListId)
	assert.Equal(t, "Steak", e.Name)
	assert.Equal(t, &qty, e.Quantity)
	assert.Equal(t, &catID, e.CategoryId)
	assert.Equal(t, &catName, e.CategoryName)
	assert.Equal(t, &sortOrder, e.CategorySortOrder)
	assert.True(t, e.Checked)
	assert.Equal(t, 3, e.Position)
	assert.Equal(t, now, e.CreatedAt)
	assert.Equal(t, now, e.UpdatedAt)
}
