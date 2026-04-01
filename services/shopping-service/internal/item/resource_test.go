package item

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransform_MapsAllFields(t *testing.T) {
	id := uuid.New()
	listID := uuid.New()
	catID := uuid.New()
	qty := "1 cup"
	catName := "Dairy"
	sortOrder := 3
	now := time.Now().Truncate(time.Second)

	m, err := NewBuilder().
		SetId(id).
		SetListID(listID).
		SetName("Milk").
		SetQuantity(&qty).
		SetCategoryID(&catID).
		SetCategoryName(&catName).
		SetCategorySortOrder(&sortOrder).
		SetChecked(true).
		SetPosition(2).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	assert.NoError(t, err)

	r := Transform(m)
	assert.Equal(t, id, r.Id)
	assert.Equal(t, "Milk", r.Name)
	assert.Equal(t, &qty, r.Quantity)
	assert.Equal(t, &catID, r.CategoryId)
	assert.Equal(t, &catName, r.CategoryName)
	assert.Equal(t, &sortOrder, r.CategorySortOrder)
	assert.True(t, r.Checked)
	assert.Equal(t, 2, r.Position)
	assert.Equal(t, now, r.CreatedAt)
	assert.Equal(t, now, r.UpdatedAt)
}

func TestTransform_NilOptionalFields(t *testing.T) {
	m, err := NewBuilder().
		SetName("Bread").
		Build()
	assert.NoError(t, err)

	r := Transform(m)
	assert.Equal(t, "Bread", r.Name)
	assert.Nil(t, r.Quantity)
	assert.Nil(t, r.CategoryId)
	assert.Nil(t, r.CategoryName)
	assert.Nil(t, r.CategorySortOrder)
	assert.False(t, r.Checked)
}

func TestRestModel_GetName(t *testing.T) {
	r := RestModel{}
	assert.Equal(t, "shopping-items", r.GetName())
}

func TestRestModel_GetID(t *testing.T) {
	id := uuid.New()
	r := RestModel{Id: id}
	assert.Equal(t, id.String(), r.GetID())
}

func TestRestModel_SetID(t *testing.T) {
	id := uuid.New()
	r := &RestModel{}
	err := r.SetID(id.String())
	assert.NoError(t, err)
	assert.Equal(t, id, r.Id)
}

func TestRestModel_SetID_InvalidUUID(t *testing.T) {
	r := &RestModel{}
	err := r.SetID("bad")
	assert.Error(t, err)
}

func TestTransformSlice_Empty(t *testing.T) {
	result := TransformSlice([]Model{})
	assert.Empty(t, result)
}

func TestTransformSlice_Multiple(t *testing.T) {
	models := make([]Model, 2)
	for i := range models {
		m, err := NewBuilder().
			SetId(uuid.New()).
			SetName("Item " + string(rune('A'+i))).
			Build()
		assert.NoError(t, err)
		models[i] = m
	}

	result := TransformSlice(models)
	assert.Len(t, result, 2)
	for i, r := range result {
		assert.Equal(t, models[i].Id(), r.Id)
		assert.Equal(t, models[i].Name(), r.Name)
	}
}
