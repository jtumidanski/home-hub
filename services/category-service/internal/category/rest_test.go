package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransform_MapsAllFields(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().Truncate(time.Second)

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetName("Produce").
		SetSortOrder(1).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	assert.NoError(t, err)

	r, err := Transform(m)
	assert.NoError(t, err)
	assert.Equal(t, id, r.Id)
	assert.Equal(t, "Produce", r.Name)
	assert.Equal(t, 1, r.SortOrder)
	assert.Equal(t, now, r.CreatedAt)
	assert.Equal(t, now, r.UpdatedAt)
}

func TestTransform_GetName(t *testing.T) {
	r := RestModel{}
	assert.Equal(t, "categories", r.GetName())
}

func TestTransform_GetID(t *testing.T) {
	id := uuid.New()
	r := RestModel{Id: id}
	assert.Equal(t, id.String(), r.GetID())
}

func TestTransform_SetID(t *testing.T) {
	id := uuid.New()
	r := &RestModel{}
	err := r.SetID(id.String())
	assert.NoError(t, err)
	assert.Equal(t, id, r.Id)
}

func TestTransform_SetID_InvalidUUID(t *testing.T) {
	r := &RestModel{}
	err := r.SetID("not-a-uuid")
	assert.Error(t, err)
}

func TestTransformSlice_Empty(t *testing.T) {
	result, err := TransformSlice([]Model{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestTransformSlice_Multiple(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	models := make([]Model, 3)
	for i := range models {
		m, err := NewBuilder().
			SetId(uuid.New()).
			SetName("Category " + string(rune('A'+i))).
			SetSortOrder(i).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Build()
		assert.NoError(t, err)
		models[i] = m
	}

	result, err := TransformSlice(models)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	for i, r := range result {
		assert.Equal(t, models[i].Id(), r.Id)
		assert.Equal(t, models[i].Name(), r.Name)
		assert.Equal(t, models[i].SortOrder(), r.SortOrder)
	}
}
