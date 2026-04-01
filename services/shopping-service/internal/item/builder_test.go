package item

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build_RequiresName(t *testing.T) {
	_, err := NewBuilder().Build()
	assert.ErrorIs(t, err, ErrNameRequired)
}

func TestBuilder_Build_RejectsLongName(t *testing.T) {
	_, err := NewBuilder().
		SetName(strings.Repeat("a", 256)).
		Build()
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestBuilder_Build_Success(t *testing.T) {
	id := uuid.New()
	listID := uuid.New()
	catID := uuid.New()
	qty := "2 lb"
	catName := "Produce"
	sortOrder := 1

	m, err := NewBuilder().
		SetId(id).
		SetListID(listID).
		SetName("Chicken breast").
		SetQuantity(&qty).
		SetCategoryID(&catID).
		SetCategoryName(&catName).
		SetCategorySortOrder(&sortOrder).
		SetChecked(false).
		SetPosition(0).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, id, m.Id())
	assert.Equal(t, listID, m.ListID())
	assert.Equal(t, "Chicken breast", m.Name())
	assert.Equal(t, &qty, m.Quantity())
	assert.Equal(t, &catID, m.CategoryID())
	assert.Equal(t, &catName, m.CategoryName())
	assert.Equal(t, &sortOrder, m.CategorySortOrder())
	assert.False(t, m.Checked())
	assert.Equal(t, 0, m.Position())
}

func TestBuilder_Build_MinimalItem(t *testing.T) {
	m, err := NewBuilder().
		SetName("Milk").
		Build()

	assert.NoError(t, err)
	assert.Equal(t, "Milk", m.Name())
	assert.Nil(t, m.Quantity())
	assert.Nil(t, m.CategoryID())
	assert.Nil(t, m.CategoryName())
	assert.Nil(t, m.CategorySortOrder())
	assert.False(t, m.Checked())
}
