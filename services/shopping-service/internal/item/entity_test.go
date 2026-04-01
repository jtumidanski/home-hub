package item

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMake_FullEntity(t *testing.T) {
	id := uuid.New()
	listID := uuid.New()
	catID := uuid.New()
	qty := "2 lb"
	catName := "Produce"
	sortOrder := 1
	now := time.Now().Truncate(time.Second)

	e := Entity{
		Id:                id,
		ListId:            listID,
		Name:              "Chicken breast",
		Quantity:          &qty,
		CategoryId:        &catID,
		CategoryName:      &catName,
		CategorySortOrder: &sortOrder,
		Checked:           true,
		Position:          5,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	m, err := Make(e)
	assert.NoError(t, err)
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
}

func TestMake_MinimalEntity(t *testing.T) {
	e := Entity{
		Name: "Milk",
	}

	m, err := Make(e)
	assert.NoError(t, err)
	assert.Equal(t, "Milk", m.Name())
	assert.Nil(t, m.Quantity())
	assert.Nil(t, m.CategoryID())
	assert.Nil(t, m.CategoryName())
	assert.Nil(t, m.CategorySortOrder())
	assert.False(t, m.Checked())
	assert.Equal(t, 0, m.Position())
}

func TestMake_EmptyName_ReturnsError(t *testing.T) {
	e := Entity{Name: ""}
	_, err := Make(e)
	assert.ErrorIs(t, err, ErrNameRequired)
}

func TestMake_NameTooLong_ReturnsError(t *testing.T) {
	e := Entity{Name: string(make([]byte, 256))}
	_, err := Make(e)
	assert.ErrorIs(t, err, ErrNameTooLong)
}
