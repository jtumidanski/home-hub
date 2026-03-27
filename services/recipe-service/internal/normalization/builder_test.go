package normalization

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder_Build(t *testing.T) {
	t.Run("successful build with valid rawName", func(t *testing.T) {
		id := uuid.New()
		tenantID := uuid.New()
		householdID := uuid.New()
		recipeID := uuid.New()
		canonicalID := uuid.New()
		now := time.Now().UTC()

		m, err := NewBuilder().
			SetId(id).
			SetTenantID(tenantID).
			SetHouseholdID(householdID).
			SetRecipeID(recipeID).
			SetRawName("garlic cloves").
			SetRawQuantity("3").
			SetRawUnit("clove").
			SetPosition(2).
			SetCanonicalIngredientID(&canonicalID).
			SetCanonicalUnit("clove").
			SetNormalizationStatus(StatusMatched).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if m.Id() != id {
			t.Errorf("Id() = %v, want %v", m.Id(), id)
		}
		if m.TenantID() != tenantID {
			t.Errorf("TenantID() = %v, want %v", m.TenantID(), tenantID)
		}
		if m.HouseholdID() != householdID {
			t.Errorf("HouseholdID() = %v, want %v", m.HouseholdID(), householdID)
		}
		if m.RecipeID() != recipeID {
			t.Errorf("RecipeID() = %v, want %v", m.RecipeID(), recipeID)
		}
		if m.RawName() != "garlic cloves" {
			t.Errorf("RawName() = %q, want %q", m.RawName(), "garlic cloves")
		}
		if m.RawQuantity() != "3" {
			t.Errorf("RawQuantity() = %q, want %q", m.RawQuantity(), "3")
		}
		if m.RawUnit() != "clove" {
			t.Errorf("RawUnit() = %q, want %q", m.RawUnit(), "clove")
		}
		if m.Position() != 2 {
			t.Errorf("Position() = %d, want %d", m.Position(), 2)
		}
		if m.CanonicalIngredientID() == nil || *m.CanonicalIngredientID() != canonicalID {
			t.Errorf("CanonicalIngredientID() = %v, want %v", m.CanonicalIngredientID(), &canonicalID)
		}
		if m.CanonicalUnit() != "clove" {
			t.Errorf("CanonicalUnit() = %q, want %q", m.CanonicalUnit(), "clove")
		}
		if m.NormalizationStatus() != StatusMatched {
			t.Errorf("NormalizationStatus() = %q, want %q", m.NormalizationStatus(), StatusMatched)
		}
		if !m.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", m.CreatedAt(), now)
		}
		if !m.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", m.UpdatedAt(), now)
		}
	})

	t.Run("ErrRawNameRequired when rawName is empty", func(t *testing.T) {
		_, err := NewBuilder().
			SetRawQuantity("1").
			SetRawUnit("cup").
			Build()

		if err != ErrRawNameRequired {
			t.Fatalf("expected ErrRawNameRequired, got %v", err)
		}
	})

	t.Run("nil canonical ingredient ID", func(t *testing.T) {
		m, err := NewBuilder().
			SetRawName("mystery ingredient").
			SetNormalizationStatus(StatusUnresolved).
			Build()

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if m.CanonicalIngredientID() != nil {
			t.Errorf("CanonicalIngredientID() = %v, want nil", m.CanonicalIngredientID())
		}
	})
}
