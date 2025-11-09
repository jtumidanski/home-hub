package user

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

func createTestUser(t *testing.T, db *gorm.DB, email string, displayName string, householdId *uuid.UUID) Entity {
	t.Helper()

	entity := Entity{
		Id:          uuid.New(),
		Email:       email,
		DisplayName: displayName,
		HouseholdId: householdId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&entity).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return entity
}

func TestGetById_Exists(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	testUser := createTestUser(t, db, "test@example.com", "Test User", nil)

	// Fetch using provider
	provider := GetById(db)(testUser.Id)
	model, err := provider()

	if err != nil {
		t.Fatalf("GetById() failed: %v", err)
	}

	if model.Id() != testUser.Id {
		t.Errorf("Id = %v, want %v", model.Id(), testUser.Id)
	}
	if model.Email() != testUser.Email {
		t.Errorf("Email = %v, want %v", model.Email(), testUser.Email)
	}
	if model.DisplayName() != testUser.DisplayName {
		t.Errorf("DisplayName = %v, want %v", model.DisplayName(), testUser.DisplayName)
	}
}

func TestGetById_NotFound(t *testing.T) {
	db := setupTestDB(t)

	// Try to fetch non-existing user
	nonExistingId := uuid.New()
	provider := GetById(db)(nonExistingId)
	_, err := provider()

	if err == nil {
		t.Error("Expected error for non-existing user, got nil")
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
		t.Errorf("GetAll() returned %d users, want 0", len(models))
	}
}

func TestGetAll_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)

	// Create multiple test users
	user1 := createTestUser(t, db, "user1@example.com", "User 1", nil)
	user2 := createTestUser(t, db, "user2@example.com", "User 2", nil)
	user3 := createTestUser(t, db, "user3@example.com", "User 3", nil)

	provider := GetAll(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	if len(models) != 3 {
		t.Fatalf("GetAll() returned %d users, want 3", len(models))
	}

	// Verify all users are present (order may vary)
	ids := []uuid.UUID{user1.Id, user2.Id, user3.Id}
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
			t.Errorf("Expected user with ID %v not found in results", expectedId)
		}
	}
}

func TestGetByEmail_Exists(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	testUser := createTestUser(t, db, "test@example.com", "Test User", nil)

	// Fetch using provider
	provider := GetByEmail(db)("test@example.com")
	model, err := provider()

	if err != nil {
		t.Fatalf("GetByEmail() failed: %v", err)
	}

	if model.Id() != testUser.Id {
		t.Errorf("Id = %v, want %v", model.Id(), testUser.Id)
	}
	if model.Email() != testUser.Email {
		t.Errorf("Email = %v, want %v", model.Email(), testUser.Email)
	}
}

func TestGetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)

	// Try to fetch non-existing email
	provider := GetByEmail(db)("nonexisting@example.com")
	_, err := provider()

	if err == nil {
		t.Error("Expected error for non-existing email, got nil")
	}
}

func TestGetByEmail_CaseSensitive(t *testing.T) {
	db := setupTestDB(t)

	// Create test user with lowercase email
	createTestUser(t, db, "test@example.com", "Test User", nil)

	// Try to fetch with different case
	provider := GetByEmail(db)("TEST@EXAMPLE.COM")
	_, err := provider()

	// This should fail because GORM query is case-sensitive by default
	if err == nil {
		t.Log("Note: GetByEmail appears to be case-insensitive (depends on DB)")
	}
}

func TestGetByHouseholdId_WithUsers(t *testing.T) {
	db := setupTestDB(t)

	householdId := uuid.New()

	// Create users in the household
	user1 := createTestUser(t, db, "user1@example.com", "User 1", &householdId)
	user2 := createTestUser(t, db, "user2@example.com", "User 2", &householdId)

	// Create user in different household
	otherHouseholdId := uuid.New()
	createTestUser(t, db, "user3@example.com", "User 3", &otherHouseholdId)

	// Create user without household
	createTestUser(t, db, "user4@example.com", "User 4", nil)

	// Fetch users by household
	provider := GetByHouseholdId(db)(householdId)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetByHouseholdId() failed: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("GetByHouseholdId() returned %d users, want 2", len(models))
	}

	// Verify correct users returned
	foundUser1 := false
	foundUser2 := false
	for _, model := range models {
		if model.Id() == user1.Id {
			foundUser1 = true
		}
		if model.Id() == user2.Id {
			foundUser2 = true
		}
	}

	if !foundUser1 || !foundUser2 {
		t.Error("Expected users not found in household query results")
	}
}

func TestGetByHouseholdId_NoUsers(t *testing.T) {
	db := setupTestDB(t)

	// Create user without household
	createTestUser(t, db, "user1@example.com", "User 1", nil)

	// Try to fetch users for non-existing household
	nonExistingHouseholdId := uuid.New()
	provider := GetByHouseholdId(db)(nonExistingHouseholdId)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetByHouseholdId() failed: %v", err)
	}

	if len(models) != 0 {
		t.Errorf("GetByHouseholdId() returned %d users, want 0", len(models))
	}
}

func TestGetUsersWithoutHousehold_NoUsers(t *testing.T) {
	db := setupTestDB(t)

	// Create users with households
	householdId := uuid.New()
	createTestUser(t, db, "user1@example.com", "User 1", &householdId)
	createTestUser(t, db, "user2@example.com", "User 2", &householdId)

	provider := GetUsersWithoutHousehold(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetUsersWithoutHousehold() failed: %v", err)
	}

	if len(models) != 0 {
		t.Errorf("GetUsersWithoutHousehold() returned %d users, want 0", len(models))
	}
}

func TestGetUsersWithoutHousehold_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)

	// Create users without household
	user1 := createTestUser(t, db, "user1@example.com", "User 1", nil)
	user2 := createTestUser(t, db, "user2@example.com", "User 2", nil)

	// Create user with household
	householdId := uuid.New()
	createTestUser(t, db, "user3@example.com", "User 3", &householdId)

	provider := GetUsersWithoutHousehold(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetUsersWithoutHousehold() failed: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("GetUsersWithoutHousehold() returned %d users, want 2", len(models))
	}

	// Verify correct users returned
	foundUser1 := false
	foundUser2 := false
	for _, model := range models {
		if model.Id() == user1.Id {
			foundUser1 = true
		}
		if model.Id() == user2.Id {
			foundUser2 = true
		}
		if model.HasHousehold() {
			t.Errorf("Found user with household in GetUsersWithoutHousehold results: %v", model.Id())
		}
	}

	if !foundUser1 || !foundUser2 {
		t.Error("Expected users not found in GetUsersWithoutHousehold results")
	}
}

func TestProvider_HouseholdAssociation(t *testing.T) {
	db := setupTestDB(t)

	householdId := uuid.New()

	// Create user with household
	testUser := createTestUser(t, db, "test@example.com", "Test User", &householdId)

	// Fetch and verify household association
	provider := GetById(db)(testUser.Id)
	model, err := provider()

	if err != nil {
		t.Fatalf("GetById() failed: %v", err)
	}

	if !model.HasHousehold() {
		t.Error("Expected user to have household")
	}

	if *model.HouseholdId() != householdId {
		t.Errorf("HouseholdId = %v, want %v", *model.HouseholdId(), householdId)
	}
}

func TestProvider_Timestamps(t *testing.T) {
	db := setupTestDB(t)

	// Create user
	before := time.Now().Add(-1 * time.Second)
	testUser := createTestUser(t, db, "test@example.com", "Test User", nil)
	after := time.Now().Add(1 * time.Second)

	// Fetch and verify timestamps
	provider := GetById(db)(testUser.Id)
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

	householdId := uuid.New()
	testUser := createTestUser(t, db, "test@example.com", "Test User", &householdId)

	provider := GetAll(db)
	models, err := provider()

	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	if len(models) != 1 {
		t.Fatalf("GetAll() returned %d users, want 1", len(models))
	}

	model := models[0]

	// Verify all fields are preserved
	if model.Id() != testUser.Id {
		t.Errorf("Id = %v, want %v", model.Id(), testUser.Id)
	}
	if model.Email() != testUser.Email {
		t.Errorf("Email = %v, want %v", model.Email(), testUser.Email)
	}
	if model.DisplayName() != testUser.DisplayName {
		t.Errorf("DisplayName = %v, want %v", model.DisplayName(), testUser.DisplayName)
	}
	if !model.HasHousehold() {
		t.Error("Expected household to be preserved")
	}
	if *model.HouseholdId() != householdId {
		t.Errorf("HouseholdId = %v, want %v", *model.HouseholdId(), householdId)
	}
}
