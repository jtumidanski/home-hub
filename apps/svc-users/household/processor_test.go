package household

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// userEntity is a minimal user entity for testing household deletion business rule
type userEntity struct {
	Id          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email       string     `gorm:"type:varchar(255);not null"`
	DisplayName string     `gorm:"type:varchar(255);not null"`
	HouseholdId *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

func (userEntity) TableName() string {
	return "users"
}

func setupProcessorTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations for both households and users
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("Failed to migrate households: %v", err)
	}
	if err := db.AutoMigrate(&userEntity{}); err != nil {
		t.Fatalf("Failed to migrate users: %v", err)
	}

	return db
}

func createTestUserInHousehold(t *testing.T, db *gorm.DB, email string, householdId uuid.UUID) userEntity {
	t.Helper()

	user := userEntity{
		Id:          uuid.New(),
		Email:       email,
		DisplayName: "Test User",
		HouseholdId: &householdId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func newTestProcessor(db *gorm.DB) Processor {
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel) // Reduce noise in tests
	return NewProcessor(log, context.Background(), db)
}

func TestProcessor_Create_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	input := CreateInput{
		Name: "Smith Family",
	}

	model, err := processor.Create(input)()

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if model.Name() != input.Name {
		t.Errorf("Name = %v, want %v", model.Name(), input.Name)
	}
	if model.Id() == uuid.Nil {
		t.Error("Expected non-nil UUID to be generated")
	}
	if model.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if model.UpdatedAt().IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestProcessor_Create_MissingName(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	input := CreateInput{
		Name: "",
	}

	_, err := processor.Create(input)()

	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestProcessor_Create_WhitespaceName(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	input := CreateInput{
		Name: "   ",
	}

	_, err := processor.Create(input)()

	if err == nil {
		t.Error("Expected error for whitespace-only name")
	}
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("Expected ErrNameEmpty, got %v", err)
	}
}

func TestProcessor_GetById_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	createInput := CreateInput{
		Name: "Test Household",
	}
	created, err := processor.Create(createInput)()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Get by ID
	model, err := processor.GetById(created.Id())()

	if err != nil {
		t.Fatalf("GetById() failed: %v", err)
	}

	if model.Id() != created.Id() {
		t.Errorf("Id = %v, want %v", model.Id(), created.Id())
	}
	if model.Name() != created.Name() {
		t.Errorf("Name = %v, want %v", model.Name(), created.Name())
	}
}

func TestProcessor_GetById_NotFound(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingId := uuid.New()

	_, err := processor.GetById(nonExistingId)()

	if err == nil {
		t.Error("Expected error for non-existing household")
	}
	if !errors.Is(err, ErrHouseholdNotFound) {
		t.Errorf("Expected ErrHouseholdNotFound, got %v", err)
	}
}

func TestProcessor_Update_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	createInput := CreateInput{
		Name: "Original Household",
	}
	created, err := processor.Create(createInput)()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Update household
	newName := "Updated Household"
	updateInput := UpdateInput{
		Name: &newName,
	}

	updated, err := processor.Update(created.Id(), updateInput)()

	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	if updated.Name() != newName {
		t.Errorf("Name = %v, want %v", updated.Name(), newName)
	}
	if updated.Id() != created.Id() {
		t.Error("ID should not change during update")
	}
}

func TestProcessor_Update_NotFound(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingId := uuid.New()
	newName := "Updated Household"
	updateInput := UpdateInput{
		Name: &newName,
	}

	_, err := processor.Update(nonExistingId, updateInput)()

	if err == nil {
		t.Error("Expected error for non-existing household")
	}
	if !errors.Is(err, ErrHouseholdNotFound) {
		t.Errorf("Expected ErrHouseholdNotFound, got %v", err)
	}
}

func TestProcessor_Update_EmptyInput(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	createInput := CreateInput{
		Name: "Original Household",
	}
	created, err := processor.Create(createInput)()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Update with empty input (should preserve existing values)
	updateInput := UpdateInput{
		Name: nil,
	}

	updated, err := processor.Update(created.Id(), updateInput)()

	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	if updated.Name() != created.Name() {
		t.Errorf("Name changed when it shouldn't: %v -> %v", created.Name(), updated.Name())
	}
}

func TestProcessor_Delete_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	created, err := processor.Create(CreateInput{
		Name: "Test Household",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Delete household
	err = processor.Delete(created.Id())

	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify household is gone
	_, err = processor.GetById(created.Id())()
	if err == nil {
		t.Error("Expected household to be deleted")
	}
	if !errors.Is(err, ErrHouseholdNotFound) {
		t.Errorf("Expected ErrHouseholdNotFound after delete, got %v", err)
	}
}

func TestProcessor_Delete_NotFound(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingId := uuid.New()

	err := processor.Delete(nonExistingId)

	if err == nil {
		t.Error("Expected error for non-existing household")
	}
	if !errors.Is(err, ErrHouseholdNotFound) {
		t.Errorf("Expected ErrHouseholdNotFound, got %v", err)
	}
}

func TestProcessor_Delete_WithUsers(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	household, err := processor.Create(CreateInput{
		Name: "Test Household",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Create user in household
	createTestUserInHousehold(t, db, "user@example.com", household.Id())

	// Try to delete household (should fail)
	err = processor.Delete(household.Id())

	if err == nil {
		t.Error("Expected error when deleting household with users")
	}
	if !errors.Is(err, ErrHouseholdHasUsers) {
		t.Errorf("Expected ErrHouseholdHasUsers, got %v", err)
	}

	// Verify household still exists
	retrieved, err := processor.GetById(household.Id())()
	if err != nil {
		t.Error("Household should still exist after failed delete")
	}
	if retrieved.Id() != household.Id() {
		t.Error("Household ID mismatch after failed delete")
	}
}

func TestProcessor_Delete_WithMultipleUsers(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	household, err := processor.Create(CreateInput{
		Name: "Test Household",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Create multiple users in household
	createTestUserInHousehold(t, db, "user1@example.com", household.Id())
	createTestUserInHousehold(t, db, "user2@example.com", household.Id())
	createTestUserInHousehold(t, db, "user3@example.com", household.Id())

	// Try to delete household (should fail)
	err = processor.Delete(household.Id())

	if err == nil {
		t.Error("Expected error when deleting household with multiple users")
	}
	if !errors.Is(err, ErrHouseholdHasUsers) {
		t.Errorf("Expected ErrHouseholdHasUsers, got %v", err)
	}
}

func TestProcessor_Delete_NoUsers(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create two households
	household1, err := processor.Create(CreateInput{
		Name: "Household 1",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	household2, err := processor.Create(CreateInput{
		Name: "Household 2",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Create user in household2
	createTestUserInHousehold(t, db, "user@example.com", household2.Id())

	// Delete household1 (should succeed, no users)
	err = processor.Delete(household1.Id())

	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify household1 is gone
	_, err = processor.GetById(household1.Id())()
	if err == nil {
		t.Error("Expected household1 to be deleted")
	}

	// Verify household2 still exists
	retrieved, err := processor.GetById(household2.Id())()
	if err != nil {
		t.Error("Household2 should still exist")
	}
	if retrieved.Id() != household2.Id() {
		t.Error("Household2 ID mismatch")
	}
}

func TestProcessor_CRUD_Lifecycle(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create
	created, err := processor.Create(CreateInput{
		Name: "Test Household",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Read
	retrieved, err := processor.GetById(created.Id())()
	if err != nil {
		t.Fatalf("GetById() failed: %v", err)
	}
	if retrieved.Name() != "Test Household" {
		t.Errorf("Retrieved name = %v, want %v", retrieved.Name(), "Test Household")
	}

	// Update
	newName := "Updated Household"
	updated, err := processor.Update(created.Id(), UpdateInput{
		Name: &newName,
	})()
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if updated.Name() != newName {
		t.Errorf("Updated name = %v, want %v", updated.Name(), newName)
	}

	// Delete
	err = processor.Delete(created.Id())
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	_, err = processor.GetById(created.Id())()
	if err == nil {
		t.Error("Expected household to be deleted")
	}
}
