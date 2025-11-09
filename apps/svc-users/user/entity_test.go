package user

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

	if tableName != "users" {
		t.Errorf("TableName() = %v, want %v", tableName, "users")
	}
}

func TestMake_EntityToModel(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		entity Entity
		verify func(t *testing.T, model Model)
	}{
		{
			name: "entity with household",
			entity: Entity{
				Id:          id,
				Email:       "test@example.com",
				DisplayName: "Test User",
				HouseholdId: &householdId,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			verify: func(t *testing.T, model Model) {
				if model.Id() != id {
					t.Errorf("Id = %v, want %v", model.Id(), id)
				}
				if model.Email() != "test@example.com" {
					t.Errorf("Email = %v, want %v", model.Email(), "test@example.com")
				}
				if model.DisplayName() != "Test User" {
					t.Errorf("DisplayName = %v, want %v", model.DisplayName(), "Test User")
				}
				if !model.HasHousehold() {
					t.Error("Expected model to have household")
				}
				if *model.HouseholdId() != householdId {
					t.Errorf("HouseholdId = %v, want %v", model.HouseholdId(), householdId)
				}
				if !model.CreatedAt().Equal(createdAt) {
					t.Errorf("CreatedAt = %v, want %v", model.CreatedAt(), createdAt)
				}
				if !model.UpdatedAt().Equal(updatedAt) {
					t.Errorf("UpdatedAt = %v, want %v", model.UpdatedAt(), updatedAt)
				}
			},
		},
		{
			name: "entity without household",
			entity: Entity{
				Id:          id,
				Email:       "test@example.com",
				DisplayName: "Test User",
				HouseholdId: nil,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			verify: func(t *testing.T, model Model) {
				if model.Id() != id {
					t.Errorf("Id = %v, want %v", model.Id(), id)
				}
				if model.HasHousehold() {
					t.Error("Expected model to have no household")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := Make(tt.entity)
			if err != nil {
				t.Errorf("Make() unexpected error = %v", err)
				return
			}

			tt.verify(t, model)
		})
	}
}

func TestModel_ToEntity(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		model  Model
		verify func(t *testing.T, entity Entity)
	}{
		{
			name: "model with household",
			model: Model{
				id:          id,
				email:       "test@example.com",
				displayName: "Test User",
				householdId: &householdId,
				createdAt:   createdAt,
				updatedAt:   updatedAt,
			},
			verify: func(t *testing.T, entity Entity) {
				if entity.Id != id {
					t.Errorf("Id = %v, want %v", entity.Id, id)
				}
				if entity.Email != "test@example.com" {
					t.Errorf("Email = %v, want %v", entity.Email, "test@example.com")
				}
				if entity.DisplayName != "Test User" {
					t.Errorf("DisplayName = %v, want %v", entity.DisplayName, "Test User")
				}
				if entity.HouseholdId == nil {
					t.Error("Expected entity to have household")
				} else if *entity.HouseholdId != householdId {
					t.Errorf("HouseholdId = %v, want %v", entity.HouseholdId, householdId)
				}
				if !entity.CreatedAt.Equal(createdAt) {
					t.Errorf("CreatedAt = %v, want %v", entity.CreatedAt, createdAt)
				}
				if !entity.UpdatedAt.Equal(updatedAt) {
					t.Errorf("UpdatedAt = %v, want %v", entity.UpdatedAt, updatedAt)
				}
			},
		},
		{
			name: "model without household",
			model: Model{
				id:          id,
				email:       "test@example.com",
				displayName: "Test User",
				householdId: nil,
				createdAt:   createdAt,
				updatedAt:   updatedAt,
			},
			verify: func(t *testing.T, entity Entity) {
				if entity.Id != id {
					t.Errorf("Id = %v, want %v", entity.Id, id)
				}
				if entity.HouseholdId != nil {
					t.Error("Expected entity to have no household")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.model.ToEntity()
			tt.verify(t, entity)
		})
	}
}

func TestRoundTrip_EntityToModelToEntity(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Entity{
		Id:          id,
		Email:       "test@example.com",
		DisplayName: "Test User",
		HouseholdId: &householdId,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
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
	if restored.Email != original.Email {
		t.Errorf("Round-trip Email = %v, want %v", restored.Email, original.Email)
	}
	if restored.DisplayName != original.DisplayName {
		t.Errorf("Round-trip DisplayName = %v, want %v", restored.DisplayName, original.DisplayName)
	}
	if (restored.HouseholdId == nil) != (original.HouseholdId == nil) {
		t.Errorf("Round-trip HouseholdId nil mismatch: got %v, want %v", restored.HouseholdId, original.HouseholdId)
	}
	if restored.HouseholdId != nil && original.HouseholdId != nil && *restored.HouseholdId != *original.HouseholdId {
		t.Errorf("Round-trip HouseholdId = %v, want %v", *restored.HouseholdId, *original.HouseholdId)
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
	householdId := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Model{
		id:          id,
		email:       "test@example.com",
		displayName: "Test User",
		householdId: &householdId,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
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
	if restored.Email() != original.Email() {
		t.Errorf("Round-trip Email = %v, want %v", restored.Email(), original.Email())
	}
	if restored.DisplayName() != original.DisplayName() {
		t.Errorf("Round-trip DisplayName = %v, want %v", restored.DisplayName(), original.DisplayName())
	}
	if restored.HasHousehold() != original.HasHousehold() {
		t.Errorf("Round-trip HasHousehold = %v, want %v", restored.HasHousehold(), original.HasHousehold())
	}
	if restored.HasHousehold() && original.HasHousehold() && *restored.HouseholdId() != *original.HouseholdId() {
		t.Errorf("Round-trip HouseholdId = %v, want %v", *restored.HouseholdId(), *original.HouseholdId())
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
	if err := db.Table("users").Count(&count).Error; err != nil {
		t.Errorf("Failed to query users table after migration: %v", err)
	}

	// Verify we can insert a record
	id := uuid.New()
	entity := Entity{
		Id:          id,
		Email:       "test@example.com",
		DisplayName: "Test User",
		HouseholdId: nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&entity).Error; err != nil {
		t.Errorf("Failed to insert user after migration: %v", err)
	}

	// Verify we can query it back
	var retrieved Entity
	if err := db.Where("id = ?", id).First(&retrieved).Error; err != nil {
		t.Errorf("Failed to retrieve user after migration: %v", err)
	}

	if retrieved.Email != "test@example.com" {
		t.Errorf("Retrieved email = %v, want %v", retrieved.Email, "test@example.com")
	}
}

func TestMigration_UniqueEmailConstraint(t *testing.T) {
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

	// Insert first user
	user1 := Entity{
		Id:          uuid.New(),
		Email:       "duplicate@example.com",
		DisplayName: "User 1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Attempt to insert second user with same email
	user2 := Entity{
		Id:          uuid.New(),
		Email:       "duplicate@example.com", // Same email
		DisplayName: "User 2",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = db.Create(&user2).Error
	if err == nil {
		t.Error("Expected unique constraint violation but insert succeeded")
	}
}

func TestMigration_HouseholdIdIndex(t *testing.T) {
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

	// Create multiple users with same household
	householdId := uuid.New()
	for i := 0; i < 3; i++ {
		user := Entity{
			Id:          uuid.New(),
			Email:       "user" + string(rune('0'+i)) + "@example.com",
			DisplayName: "User",
			HouseholdId: &householdId,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
	}

	// Query users by household (this should use the index)
	var users []Entity
	if err := db.Where("household_id = ?", householdId).Find(&users).Error; err != nil {
		t.Errorf("Failed to query users by household: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Found %d users, want 3", len(users))
	}
}
