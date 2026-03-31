package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTransform(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(uuid.New()).
		SetName("Dairy & Eggs").
		SetSortOrder(3).
		SetIngredientCount(12).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := Transform(m)

	if r.Id != id {
		t.Errorf("Id = %v, want %v", r.Id, id)
	}
	if r.Name != "Dairy & Eggs" {
		t.Errorf("Name = %q, want %q", r.Name, "Dairy & Eggs")
	}
	if r.SortOrder != 3 {
		t.Errorf("SortOrder = %d, want %d", r.SortOrder, 3)
	}
	if r.IngredientCount != 12 {
		t.Errorf("IngredientCount = %d, want %d", r.IngredientCount, 12)
	}
}

func TestTransformSlice(t *testing.T) {
	now := time.Now().UTC()

	models := make([]Model, 3)
	for i := 0; i < 3; i++ {
		m, err := NewBuilder().
			SetId(uuid.New()).
			SetTenantID(uuid.New()).
			SetName("Category " + string(rune('A'+i))).
			SetSortOrder(i + 1).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Build()
		if err != nil {
			t.Fatalf("unexpected error building model %d: %v", i, err)
		}
		models[i] = m
	}

	result := TransformSlice(models)

	if len(result) != 3 {
		t.Fatalf("TransformSlice length = %d, want 3", len(result))
	}
	for i, r := range result {
		if r.SortOrder != i+1 {
			t.Errorf("result[%d].SortOrder = %d, want %d", i, r.SortOrder, i+1)
		}
	}
}

func TestTransformSlice_Empty(t *testing.T) {
	result := TransformSlice([]Model{})
	if len(result) != 0 {
		t.Errorf("TransformSlice(empty) length = %d, want 0", len(result))
	}
}

func TestRestModel_JSONAPIInterface(t *testing.T) {
	id := uuid.New()
	r := RestModel{Id: id, Name: "Produce"}

	if r.GetName() != "ingredient-categories" {
		t.Errorf("GetName() = %q, want %q", r.GetName(), "ingredient-categories")
	}
	if r.GetID() != id.String() {
		t.Errorf("GetID() = %q, want %q", r.GetID(), id.String())
	}

	newID := uuid.New()
	if err := r.SetID(newID.String()); err != nil {
		t.Fatalf("SetID() error: %v", err)
	}
	if r.Id != newID {
		t.Errorf("after SetID, Id = %v, want %v", r.Id, newID)
	}
}
