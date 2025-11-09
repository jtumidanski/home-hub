package household

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEntity_TableName(t *testing.T) {
	entity := Entity{}
	tableName := entity.TableName()

	if tableName != "households" {
		t.Errorf("TableName() = %v, want %v", tableName, "households")
	}
}

func TestMake_EntityToModel(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	entity := Entity{
		Id:        id,
		Name:      "Test Household",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	model, err := Make(entity)
	if err != nil {
		t.Fatalf("Make() failed: %v", err)
	}

	if model.Id() != id {
		t.Errorf("Id = %v, want %v", model.Id(), id)
	}
	if model.Name() != "Test Household" {
		t.Errorf("Name = %v, want %v", model.Name(), "Test Household")
	}
	if !model.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", model.CreatedAt(), createdAt)
	}
	if !model.UpdatedAt().Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", model.UpdatedAt(), updatedAt)
	}
}

func TestModel_ToEntity(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	model := Model{
		id:        id,
		name:      "Test Household",
		createdAt: createdAt,
		updatedAt: updatedAt,
	}

	entity := model.ToEntity()

	if entity.Id != id {
		t.Errorf("Id = %v, want %v", entity.Id, id)
	}
	if entity.Name != "Test Household" {
		t.Errorf("Name = %v, want %v", entity.Name, "Test Household")
	}
	if !entity.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", entity.CreatedAt, createdAt)
	}
	if !entity.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", entity.UpdatedAt, updatedAt)
	}
}

func TestRoundTrip_EntityToModelToEntity(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Entity{
		Id:        id,
		Name:      "Test Household",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Entity -> Model
	model, err := Make(original)
	if err != nil {
		t.Fatalf("Make() failed: %v", err)
	}

	// Model -> Entity
	restored := model.ToEntity()

	// Verify all fields match
	if restored.Id != original.Id {
		t.Errorf("Round-trip Id = %v, want %v", restored.Id, original.Id)
	}
	if restored.Name != original.Name {
		t.Errorf("Round-trip Name = %v, want %v", restored.Name, original.Name)
	}
	if !restored.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("Round-trip CreatedAt = %v, want %v", restored.CreatedAt, original.CreatedAt)
	}
	if !restored.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("Round-trip UpdatedAt = %v, want %v", restored.UpdatedAt, original.UpdatedAt)
	}
}

func TestRoundTrip_ModelToEntityToModel(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Model{
		id:        id,
		name:      "Test Household",
		createdAt: createdAt,
		updatedAt: updatedAt,
	}

	// Model -> Entity
	entity := original.ToEntity()

	// Entity -> Model
	restored, err := Make(entity)
	if err != nil {
		t.Fatalf("Make() failed: %v", err)
	}

	// Verify all fields match
	if restored.Id() != original.Id() {
		t.Errorf("Round-trip Id = %v, want %v", restored.Id(), original.Id())
	}
	if restored.Name() != original.Name() {
		t.Errorf("Round-trip Name = %v, want %v", restored.Name(), original.Name())
	}
	if !restored.CreatedAt().Equal(original.CreatedAt()) {
		t.Errorf("Round-trip CreatedAt = %v, want %v", restored.CreatedAt(), original.CreatedAt())
	}
	if !restored.UpdatedAt().Equal(original.UpdatedAt()) {
		t.Errorf("Round-trip UpdatedAt = %v, want %v", restored.UpdatedAt(), original.UpdatedAt())
	}
}

func TestMigration(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migration
	migrationFunc := Migration()
	if err := migrationFunc(db); err != nil {
		t.Fatalf("Migration() failed: %v", err)
	}

	// Verify table exists by attempting to query it
	var count int64
	if err := db.Table("households").Count(&count).Error; err != nil {
		t.Errorf("Failed to query households table after migration: %v", err)
	}

	// Verify we can insert a record
	id := uuid.New()
	entity := Entity{
		Id:        id,
		Name:      "Test Household",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&entity).Error; err != nil {
		t.Errorf("Failed to insert household after migration: %v", err)
	}

	// Verify we can query it back
	var retrieved Entity
	if err := db.Where("id = ?", id).First(&retrieved).Error; err != nil {
		t.Errorf("Failed to retrieve household after migration: %v", err)
	}

	if retrieved.Name != "Test Household" {
		t.Errorf("Retrieved name = %v, want %v", retrieved.Name, "Test Household")
	}
}
