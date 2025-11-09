package user

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

// householdEntity is a minimal household entity for testing
type householdEntity struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (householdEntity) TableName() string {
	return "households"
}

func setupProcessorTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations for both users and households
	if err := db.AutoMigrate(&householdEntity{}); err != nil {
		t.Fatalf("Failed to migrate households: %v", err)
	}
	if err := Migration()(db); err != nil {
		t.Fatalf("Failed to migrate users: %v", err)
	}

	return db
}

func createTestHousehold(t *testing.T, db *gorm.DB, name string) householdEntity {
	t.Helper()

	household := householdEntity{
		Id:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&household).Error; err != nil {
		t.Fatalf("Failed to create test household: %v", err)
	}

	return household
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
		Email:       "test@example.com",
		DisplayName: "Test User",
		HouseholdId: nil,
	}

	model, err := processor.Create(input)()

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if model.Email() != input.Email {
		t.Errorf("Email = %v, want %v", model.Email(), input.Email)
	}
	if model.DisplayName() != input.DisplayName {
		t.Errorf("DisplayName = %v, want %v", model.DisplayName(), input.DisplayName)
	}
	if model.HasHousehold() {
		t.Error("Expected no household")
	}
	if model.Id() == uuid.Nil {
		t.Error("Expected non-nil UUID to be generated")
	}
}

func TestProcessor_Create_WithHousehold(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household first
	household := createTestHousehold(t, db, "Test Household")

	input := CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		HouseholdId: &household.Id,
	}

	model, err := processor.Create(input)()

	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !model.HasHousehold() {
		t.Error("Expected household to be set")
	}
	if *model.HouseholdId() != household.Id {
		t.Errorf("HouseholdId = %v, want %v", *model.HouseholdId(), household.Id)
	}
}

func TestProcessor_Create_InvalidEmail(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	input := CreateInput{
		Email:       "invalid-email",
		DisplayName: "Test User",
	}

	_, err := processor.Create(input)()

	if err == nil {
		t.Error("Expected error for invalid email")
	}
	if !errors.Is(err, ErrEmailInvalid) {
		t.Errorf("Expected ErrEmailInvalid, got %v", err)
	}
}

func TestProcessor_Create_MissingDisplayName(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	input := CreateInput{
		Email:       "test@example.com",
		DisplayName: "",
	}

	_, err := processor.Create(input)()

	if err == nil {
		t.Error("Expected error for missing display name")
	}
}

func TestProcessor_Create_NonExistingHousehold(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingHouseholdId := uuid.New()

	input := CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		HouseholdId: &nonExistingHouseholdId,
	}

	_, err := processor.Create(input)()

	if err == nil {
		t.Error("Expected error for non-existing household")
	}
	if !errors.Is(err, ErrHouseholdNotFound) {
		t.Errorf("Expected ErrHouseholdNotFound, got %v", err)
	}
}

func TestProcessor_Create_DuplicateEmail(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create first user
	input1 := CreateInput{
		Email:       "duplicate@example.com",
		DisplayName: "User 1",
	}

	_, err := processor.Create(input1)()
	if err != nil {
		t.Fatalf("First Create() failed: %v", err)
	}

	// Try to create second user with same email
	input2 := CreateInput{
		Email:       "duplicate@example.com",
		DisplayName: "User 2",
	}

	_, err = processor.Create(input2)()
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("Expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestProcessor_GetById_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user
	createInput := CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
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
	if model.Email() != created.Email() {
		t.Errorf("Email = %v, want %v", model.Email(), created.Email())
	}
}

func TestProcessor_GetById_NotFound(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingId := uuid.New()

	_, err := processor.GetById(nonExistingId)()

	if err == nil {
		t.Error("Expected error for non-existing user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestProcessor_Update_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user
	createInput := CreateInput{
		Email:       "original@example.com",
		DisplayName: "Original Name",
	}
	created, err := processor.Create(createInput)()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Update user
	newEmail := "updated@example.com"
	newDisplayName := "Updated Name"
	updateInput := UpdateInput{
		Email:       &newEmail,
		DisplayName: &newDisplayName,
	}

	updated, err := processor.Update(created.Id(), updateInput)()

	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	if updated.Email() != newEmail {
		t.Errorf("Email = %v, want %v", updated.Email(), newEmail)
	}
	if updated.DisplayName() != newDisplayName {
		t.Errorf("DisplayName = %v, want %v", updated.DisplayName(), newDisplayName)
	}
	if updated.Id() != created.Id() {
		t.Error("ID should not change during update")
	}
}

func TestProcessor_Update_PartialUpdate(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user
	createInput := CreateInput{
		Email:       "test@example.com",
		DisplayName: "Original Name",
	}
	created, err := processor.Create(createInput)()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Update only email
	newEmail := "updated@example.com"
	updateInput := UpdateInput{
		Email:       &newEmail,
		DisplayName: nil,
	}

	updated, err := processor.Update(created.Id(), updateInput)()

	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	if updated.Email() != newEmail {
		t.Errorf("Email = %v, want %v", updated.Email(), newEmail)
	}
	if updated.DisplayName() != created.DisplayName() {
		t.Errorf("DisplayName changed when it shouldn't: %v -> %v", created.DisplayName(), updated.DisplayName())
	}
}

func TestProcessor_Update_NotFound(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingId := uuid.New()
	newEmail := "updated@example.com"
	updateInput := UpdateInput{
		Email: &newEmail,
	}

	_, err := processor.Update(nonExistingId, updateInput)()

	if err == nil {
		t.Error("Expected error for non-existing user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestProcessor_Update_DuplicateEmail(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create first user
	user1, err := processor.Create(CreateInput{
		Email:       "user1@example.com",
		DisplayName: "User 1",
	})()
	if err != nil {
		t.Fatalf("Create user1 failed: %v", err)
	}

	// Create second user
	user2, err := processor.Create(CreateInput{
		Email:       "user2@example.com",
		DisplayName: "User 2",
	})()
	if err != nil {
		t.Fatalf("Create user2 failed: %v", err)
	}

	// Try to update user2 with user1's email
	user1Email := user1.Email()
	updateInput := UpdateInput{
		Email: &user1Email, // This should fail
	}

	_, err = processor.Update(user2.Id(), updateInput)()

	if err == nil {
		t.Error("Expected error for duplicate email")
	}
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("Expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestProcessor_Update_SameEmail(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user
	created, err := processor.Create(CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Update with same email (should succeed)
	sameEmail := created.Email()
	newDisplayName := "Updated Name"
	updateInput := UpdateInput{
		Email:       &sameEmail,
		DisplayName: &newDisplayName,
	}

	updated, err := processor.Update(created.Id(), updateInput)()

	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	if updated.Email() != sameEmail {
		t.Errorf("Email = %v, want %v", updated.Email(), sameEmail)
	}
	if updated.DisplayName() != newDisplayName {
		t.Errorf("DisplayName = %v, want %v", updated.DisplayName(), newDisplayName)
	}
}

func TestProcessor_Delete_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user
	created, err := processor.Create(CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Delete user
	err = processor.Delete(created.Id())

	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify user is gone
	_, err = processor.GetById(created.Id())()
	if err == nil {
		t.Error("Expected user to be deleted")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound after delete, got %v", err)
	}
}

func TestProcessor_Delete_NotFound(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingId := uuid.New()

	err := processor.Delete(nonExistingId)

	if err == nil {
		t.Error("Expected error for non-existing user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestProcessor_AssociateHousehold_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	household := createTestHousehold(t, db, "Test Household")

	// Create user without household
	created, err := processor.Create(CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if created.HasHousehold() {
		t.Fatal("User should not have household initially")
	}

	// Associate household
	updated, err := processor.AssociateHousehold(created.Id(), household.Id)()

	if err != nil {
		t.Fatalf("AssociateHousehold() failed: %v", err)
	}

	if !updated.HasHousehold() {
		t.Error("Expected household to be associated")
	}
	if *updated.HouseholdId() != household.Id {
		t.Errorf("HouseholdId = %v, want %v", *updated.HouseholdId(), household.Id)
	}
}

func TestProcessor_AssociateHousehold_NonExistingHousehold(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user
	created, err := processor.Create(CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Try to associate non-existing household
	nonExistingHouseholdId := uuid.New()
	_, err = processor.AssociateHousehold(created.Id(), nonExistingHouseholdId)()

	if err == nil {
		t.Error("Expected error for non-existing household")
	}
	if !errors.Is(err, ErrHouseholdNotFound) {
		t.Errorf("Expected ErrHouseholdNotFound, got %v", err)
	}
}

func TestProcessor_AssociateHousehold_NonExistingUser(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household
	household := createTestHousehold(t, db, "Test Household")

	// Try to associate household with non-existing user
	nonExistingUserId := uuid.New()
	_, err := processor.AssociateHousehold(nonExistingUserId, household.Id)()

	if err == nil {
		t.Error("Expected error for non-existing user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestProcessor_DisassociateHousehold_Success(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create household and user with household
	household := createTestHousehold(t, db, "Test Household")
	created, err := processor.Create(CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		HouseholdId: &household.Id,
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !created.HasHousehold() {
		t.Fatal("User should have household initially")
	}

	// Disassociate household
	updated, err := processor.DisassociateHousehold(created.Id())()

	if err != nil {
		t.Fatalf("DisassociateHousehold() failed: %v", err)
	}

	if updated.HasHousehold() {
		t.Error("Expected household to be disassociated")
	}
}

func TestProcessor_DisassociateHousehold_NonExistingUser(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	nonExistingUserId := uuid.New()

	_, err := processor.DisassociateHousehold(nonExistingUserId)()

	if err == nil {
		t.Error("Expected error for non-existing user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestProcessor_DisassociateHousehold_AlreadyNoHousehold(t *testing.T) {
	db := setupProcessorTestDB(t)
	processor := newTestProcessor(db)

	// Create user without household
	created, err := processor.Create(CreateInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
	})()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if created.HasHousehold() {
		t.Fatal("User should not have household initially")
	}

	// Disassociate household (should succeed even though there's no household)
	updated, err := processor.DisassociateHousehold(created.Id())()

	if err != nil {
		t.Fatalf("DisassociateHousehold() failed: %v", err)
	}

	if updated.HasHousehold() {
		t.Error("Expected no household")
	}
}
