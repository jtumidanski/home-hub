package household

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migration
	if err := Migration()(db); err != nil {
		t.Fatalf("Failed to run migration: %v", err)
	}

	return db
}

func createTestHousehold(t *testing.T, db *gorm.DB, name string) Entity {
	t.Helper()

	entity := Entity{
		Id:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&entity).Error; err != nil {
		t.Fatalf("Failed to create test household: %v", err)
	}

	return entity
}

func TestGetById_Exists(t *testing.T) {
	db := setupTestDB(t)

	// Create test household
	testHousehold := createTestHousehold(t, db, "Test Household")

	// Fetch using provider
	provider := GetById(db)(testHousehold.Id)
	model, err := provider()

	if err != nil {
		t.Fatalf("GetById() failed: %v", err)
	}

	if model.Id() != testHousehold.Id {
		t.Errorf("Id = %v, want %v", model.Id(), testHousehold.Id)
	}
	if model.Name() != testHousehold.Name {
		t.Errorf("Name = %v, want %v", model.Name(), testHousehold.Name)
	}
}

func TestGetById_NotFound(t *testing.T) {
	db := setupTestDB(t)

	// Try to fetch non-existing household
	nonExistingId := uuid.New()
	provider := GetById(db)(nonExistingId)
	_, err := provider()

	if err == nil {
		t.Error("Expected error for non-existing household, got nil")
	}
}

func TestGetAll_EmptyTable(t *testing.T) {
	db := setupTestDB(t)

	provider := GetAll(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	if len(models) != 0 {
		t.Errorf("GetAll() returned %d households, want 0", len(models))
	}
}

func TestGetAll_MultipleHouseholds(t *testing.T) {
	db := setupTestDB(t)

	// Create multiple test households
	household1 := createTestHousehold(t, db, "Household 1")
	household2 := createTestHousehold(t, db, "Household 2")
	household3 := createTestHousehold(t, db, "Household 3")

	provider := GetAll(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	if len(models) != 3 {
		t.Fatalf("GetAll() returned %d households, want 3", len(models))
	}

	// Verify all households are present (order may vary)
	ids := []uuid.UUID{household1.Id, household2.Id, household3.Id}
	foundIds := make([]uuid.UUID, len(models))
	for i, model := range models {
		foundIds[i] = model.Id()
	}

	for _, expectedId := range ids {
		found := false
		for _, foundId := range foundIds {
			if foundId == expectedId {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected household with ID %v not found in results", expectedId)
		}
	}
}

func TestProvider_Timestamps(t *testing.T) {
	db := setupTestDB(t)

	// Create household
	before := time.Now().Add(-1 * time.Second)
	testHousehold := createTestHousehold(t, db, "Test Household")
	after := time.Now().Add(1 * time.Second)

	// Fetch and verify timestamps
	provider := GetById(db)(testHousehold.Id)
	model, err := provider()

	if err != nil {
		t.Fatalf("GetById() failed: %v", err)
	}

	if model.CreatedAt().Before(before) || model.CreatedAt().After(after) {
		t.Errorf("CreatedAt %v is outside expected range [%v, %v]", model.CreatedAt(), before, after)
	}

	if model.UpdatedAt().Before(before) || model.UpdatedAt().After(after) {
		t.Errorf("UpdatedAt %v is outside expected range [%v, %v]", model.UpdatedAt(), before, after)
	}
}

func TestGetAll_PreservesAllFields(t *testing.T) {
	db := setupTestDB(t)

	testHousehold := createTestHousehold(t, db, "Test Household")

	provider := GetAll(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	if len(models) != 1 {
		t.Fatalf("GetAll() returned %d households, want 1", len(models))
	}

	model := models[0]

	// Verify all fields are preserved
	if model.Id() != testHousehold.Id {
		t.Errorf("Id = %v, want %v", model.Id(), testHousehold.Id)
	}
	if model.Name() != testHousehold.Name {
		t.Errorf("Name = %v, want %v", model.Name(), testHousehold.Name)
	}
	if model.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if model.UpdatedAt().IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}
