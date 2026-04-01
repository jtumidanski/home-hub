package item

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	catID := uuid.New()
	qty := "2 lb"
	catName := "Produce"
	sortOrder := 1

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
					SetListID(uuid.New()).
					SetName("Chicken breast").
					SetQuantity(&qty).
					SetCategoryID(&catID).
					SetCategoryName(&catName).
					SetCategorySortOrder(&sortOrder).
					SetChecked(false).
					SetPosition(0).
					Build()
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Chicken breast", m.Name())
				assert.Equal(t, &qty, m.Quantity())
				assert.Equal(t, &catID, m.CategoryID())
				assert.Equal(t, &catName, m.CategoryName())
				assert.Equal(t, &sortOrder, m.CategorySortOrder())
				assert.False(t, m.Checked())
				assert.Equal(t, 0, m.Position())
			},
		},
		{
			name: "minimal item",
			build: func() (Model, error) {
				return NewBuilder().SetName("Milk").Build()
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Milk", m.Name())
				assert.Nil(t, m.Quantity())
				assert.Nil(t, m.CategoryID())
				assert.Nil(t, m.CategoryName())
				assert.Nil(t, m.CategorySortOrder())
				assert.False(t, m.Checked())
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
