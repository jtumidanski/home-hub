package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMake(t *testing.T) {
	t.Run("converts entity to model", func(t *testing.T) {
		id := uuid.New()
		tenantID := uuid.New()
		now := time.Now().UTC()

		e := Entity{
			Id:        id,
			TenantId:  tenantID,
			Name:      "Produce",
			SortOrder: 1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		m, err := Make(e)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if m.Id() != id {
			t.Errorf("Id() = %v, want %v", m.Id(), id)
		}
		if m.TenantID() != tenantID {
			t.Errorf("TenantID() = %v, want %v", m.TenantID(), tenantID)
		}
		if m.Name() != "Produce" {
			t.Errorf("Name() = %q, want %q", m.Name(), "Produce")
		}
		if m.SortOrder() != 1 {
			t.Errorf("SortOrder() = %d, want %d", m.SortOrder(), 1)
		}
	})

	t.Run("returns error for invalid entity", func(t *testing.T) {
		e := Entity{
			Id:        uuid.New(),
			TenantId:  uuid.New(),
			Name:      "", // invalid — empty name
			SortOrder: 0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err := Make(e)
		if err != ErrNameRequired {
			t.Errorf("expected ErrNameRequired, got %v", err)
		}
	})
}

func TestToEntity(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().UTC()

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetName("Frozen").
		SetSortOrder(6).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	e := m.ToEntity()

	if e.Id != id {
		t.Errorf("ToEntity().Id = %v, want %v", e.Id, id)
	}
	if e.TenantId != tenantID {
		t.Errorf("ToEntity().TenantId = %v, want %v", e.TenantId, tenantID)
	}
	if e.Name != "Frozen" {
		t.Errorf("ToEntity().Name = %q, want %q", e.Name, "Frozen")
	}
	if e.SortOrder != 6 {
		t.Errorf("ToEntity().SortOrder = %d, want %d", e.SortOrder, 6)
	}
}

func TestMake_ToEntity_RoundTrip(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().UTC()

	original := Entity{
		Id:        id,
		TenantId:  tenantID,
		Name:      "Spices & Seasonings",
		SortOrder: 10,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m, err := Make(original)
	if err != nil {
		t.Fatalf("Make() error: %v", err)
	}

	roundTripped := m.ToEntity()

	if roundTripped.Id != original.Id {
		t.Errorf("round-trip Id = %v, want %v", roundTripped.Id, original.Id)
	}
	if roundTripped.TenantId != original.TenantId {
		t.Errorf("round-trip TenantId = %v, want %v", roundTripped.TenantId, original.TenantId)
	}
	if roundTripped.Name != original.Name {
		t.Errorf("round-trip Name = %q, want %q", roundTripped.Name, original.Name)
	}
	if roundTripped.SortOrder != original.SortOrder {
		t.Errorf("round-trip SortOrder = %d, want %d", roundTripped.SortOrder, original.SortOrder)
	}
}

func TestEntity_TableName(t *testing.T) {
	e := Entity{}
	if e.TableName() != "ingredient_categories" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "ingredient_categories")
	}
}
