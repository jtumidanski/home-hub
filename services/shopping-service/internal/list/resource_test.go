package list

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/item"
	"github.com/stretchr/testify/assert"
)

func TestTransform_MapsAllFields(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Second)
	archivedAt := now.Add(-time.Hour)

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(uuid.New()).
		SetHouseholdID(uuid.New()).
		SetName("Weekly Groceries").
		SetStatus("archived").
		SetArchivedAt(&archivedAt).
		SetCreatedBy(uuid.New()).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	assert.NoError(t, err)

	updated := m.WithCounts(8, 3)
	r := Transform(updated)

	assert.Equal(t, id, r.Id)
	assert.Equal(t, "Weekly Groceries", r.Name)
	assert.Equal(t, "archived", r.Status)
	assert.Equal(t, 8, r.ItemCount)
	assert.Equal(t, 3, r.CheckedCount)
	assert.NotNil(t, r.ArchivedAt)
	assert.Equal(t, archivedAt, *r.ArchivedAt)
	assert.Equal(t, now, r.CreatedAt)
	assert.Equal(t, now, r.UpdatedAt)
}

func TestTransform_ActiveList(t *testing.T) {
	m, err := NewBuilder().
		SetName("Shopping").
		Build()
	assert.NoError(t, err)

	r := Transform(m)
	assert.Equal(t, "active", r.Status)
	assert.Nil(t, r.ArchivedAt)
	assert.Equal(t, 0, r.ItemCount)
	assert.Equal(t, 0, r.CheckedCount)
	assert.Empty(t, r.Items)
}

func TestRestModel_GetName(t *testing.T) {
	r := RestModel{}
	assert.Equal(t, "shopping-lists", r.GetName())
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
	err := r.SetID("invalid")
	assert.Error(t, err)
}

func TestTransformWithItems(t *testing.T) {
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetName("Groceries").
		Build()
	assert.NoError(t, err)

	items := []item.RestModel{
		{Id: uuid.New(), Name: "Milk"},
		{Id: uuid.New(), Name: "Bread"},
	}

	r := TransformWithItems(m, items)
	assert.Equal(t, "Groceries", r.Name)
	assert.Len(t, r.Items, 2)
	assert.Equal(t, "Milk", r.Items[0].Name)
	assert.Equal(t, "Bread", r.Items[1].Name)
}

func TestTransformWithItems_EmptyItems(t *testing.T) {
	m, err := NewBuilder().
		SetName("Empty").
		Build()
	assert.NoError(t, err)

	r := TransformWithItems(m, []item.RestModel{})
	assert.Empty(t, r.Items)
}

func TestTransformSlice_Empty(t *testing.T) {
	result := TransformSlice([]Model{})
	assert.Empty(t, result)
}

func TestTransformSlice_Multiple(t *testing.T) {
	models := make([]Model, 3)
	for i := range models {
		m, err := NewBuilder().
			SetId(uuid.New()).
			SetName("List " + string(rune('A'+i))).
			Build()
		assert.NoError(t, err)
		models[i] = m
	}

	result := TransformSlice(models)
	assert.Len(t, result, 3)
	for i, r := range result {
		assert.Equal(t, models[i].Id(), r.Id)
		assert.Equal(t, models[i].Name(), r.Name)
	}
}
