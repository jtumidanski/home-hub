package user

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Builder
		verify  func(t *testing.T, model Model)
	}{
		{
			name: "valid user with all fields",
			setup: func() *Builder {
				id := uuid.New()
				householdId := uuid.New()
				createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

				return NewBuilder().
					SetId(id).
					SetEmail("test@example.com").
					SetDisplayName("Test User").
					SetHouseholdId(householdId).
					SetCreatedAt(createdAt).
					SetUpdatedAt(updatedAt)
			},
			verify: func(t *testing.T, model Model) {
				if model.Email() != "test@example.com" {
					t.Errorf("Email = %v, want %v", model.Email(), "test@example.com")
				}
				if model.DisplayName() != "Test User" {
					t.Errorf("DisplayName = %v, want %v", model.DisplayName(), "Test User")
				}
				if !model.HasHousehold() {
					t.Error("Expected household to be present")
				}
			},
		},
		{
			name: "valid user without household",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com").
					SetDisplayName("Test User")
			},
			verify: func(t *testing.T, model Model) {
				if model.Email() != "test@example.com" {
					t.Errorf("Email = %v, want %v", model.Email(), "test@example.com")
				}
				if model.HasHousehold() {
					t.Error("Expected no household")
				}
			},
		},
		{
			name: "email with whitespace is trimmed",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("  test@example.com  ").
					SetDisplayName("Test User")
			},
			verify: func(t *testing.T, model Model) {
				if model.Email() != "test@example.com" {
					t.Errorf("Email = %v, want %v (trimmed)", model.Email(), "test@example.com")
				}
			},
		},
		{
			name: "display name with whitespace is trimmed",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com").
					SetDisplayName("  Test User  ")
			},
			verify: func(t *testing.T, model Model) {
				if model.DisplayName() != "Test User" {
					t.Errorf("DisplayName = %v, want %v (trimmed)", model.DisplayName(), "Test User")
				}
			},
		},
		{
			name: "generates UUID if not provided",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com").
					SetDisplayName("Test User")
			},
			verify: func(t *testing.T, model Model) {
				if model.Id() == uuid.Nil {
					t.Error("Expected non-nil UUID to be generated")
				}
			},
		},
		{
			name: "generates timestamps if not provided",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com").
					SetDisplayName("Test User")
			},
			verify: func(t *testing.T, model Model) {
				if model.CreatedAt().IsZero() {
					t.Error("Expected CreatedAt to be generated")
				}
				if model.UpdatedAt().IsZero() {
					t.Error("Expected UpdatedAt to be generated")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			model, err := builder.Build()

			if err != nil {
				t.Errorf("Build() unexpected error = %v", err)
				return
			}

			tt.verify(t, model)
		})
	}
}

func TestBuilder_Build_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Builder
		wantErr error
	}{
		{
			name: "missing email",
			setup: func() *Builder {
				return NewBuilder().
					SetDisplayName("Test User")
			},
			wantErr: ErrEmailRequired,
		},
		{
			name: "email only whitespace",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("   ").
					SetDisplayName("Test User")
			},
			wantErr: ErrEmailRequired,
		},
		{
			name: "invalid email format - no @",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("testexample.com").
					SetDisplayName("Test User")
			},
			wantErr: ErrEmailInvalid,
		},
		{
			name: "invalid email format - no domain",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@").
					SetDisplayName("Test User")
			},
			wantErr: ErrEmailInvalid,
		},
		{
			name: "invalid email format - no TLD",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example").
					SetDisplayName("Test User")
			},
			wantErr: ErrEmailInvalid,
		},
		{
			name: "invalid email format - no local part",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("@example.com").
					SetDisplayName("Test User")
			},
			wantErr: ErrEmailInvalid,
		},
		{
			name: "missing display name",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com")
			},
			wantErr: ErrDisplayNameRequired,
		},
		{
			name: "display name only whitespace",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com").
					SetDisplayName("   ")
			},
			wantErr: ErrDisplayNameEmpty,
		},
		{
			name: "empty display name string",
			setup: func() *Builder {
				return NewBuilder().
					SetEmail("test@example.com").
					SetDisplayName("")
			},
			wantErr: ErrDisplayNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			_, err := builder.Build()

			if err == nil {
				t.Error("Build() expected error but got none")
				return
			}

			if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
				t.Errorf("Build() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuilder_FluentAPI(t *testing.T) {
	// Test that all methods return *Builder for chaining
	builder := NewBuilder()

	id := uuid.New()
	householdId := uuid.New()
	now := time.Now()

	result := builder.
		SetId(id).
		SetEmail("test@example.com").
		SetDisplayName("Test User").
		SetHouseholdId(householdId).
		SetCreatedAt(now).
		SetUpdatedAt(now)

	if result != builder {
		t.Error("Fluent API methods should return the same builder instance")
	}

	model, err := result.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if model.Id() != id {
		t.Errorf("Id = %v, want %v", model.Id(), id)
	}
	if model.Email() != "test@example.com" {
		t.Errorf("Email = %v, want %v", model.Email(), "test@example.com")
	}
}

func TestBuilder_ClearHouseholdId(t *testing.T) {
	householdId := uuid.New()

	// Start with household
	builder := NewBuilder().
		SetEmail("test@example.com").
		SetDisplayName("Test User").
		SetHouseholdId(householdId)

	model1, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if !model1.HasHousehold() {
		t.Error("Expected household to be set")
	}

	// Clear household
	builder.ClearHouseholdId()
	model2, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if model2.HasHousehold() {
		t.Error("Expected household to be cleared")
	}
}

func TestModel_Builder(t *testing.T) {
	// Create original model
	id := uuid.New()
	householdId := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Model{
		id:          id,
		email:       "original@example.com",
		displayName: "Original User",
		householdId: &householdId,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}

	// Create builder from model
	builder := original.Builder()

	// Verify builder has all original values
	model, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if model.Id() != original.Id() {
		t.Errorf("Id = %v, want %v", model.Id(), original.Id())
	}
	if model.Email() != original.Email() {
		t.Errorf("Email = %v, want %v", model.Email(), original.Email())
	}
	if model.DisplayName() != original.DisplayName() {
		t.Errorf("DisplayName = %v, want %v", model.DisplayName(), original.DisplayName())
	}
	if !model.HasHousehold() || *model.HouseholdId() != *original.HouseholdId() {
		t.Errorf("HouseholdId = %v, want %v", model.HouseholdId(), original.HouseholdId())
	}
}

func TestModel_Builder_ModificationFlow(t *testing.T) {
	// Create original model
	original := Model{
		id:          uuid.New(),
		email:       "original@example.com",
		displayName: "Original User",
		householdId: nil,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	// Modify via builder
	modified, err := original.Builder().
		SetEmail("modified@example.com").
		SetDisplayName("Modified User").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify modifications
	if modified.Email() != "modified@example.com" {
		t.Errorf("Email = %v, want %v", modified.Email(), "modified@example.com")
	}
	if modified.DisplayName() != "Modified User" {
		t.Errorf("DisplayName = %v, want %v", modified.DisplayName(), "Modified User")
	}

	// Verify ID is preserved
	if modified.Id() != original.Id() {
		t.Errorf("Id changed during modification: %v -> %v", original.Id(), modified.Id())
	}

	// Verify original is unchanged (immutability)
	if original.Email() != "original@example.com" {
		t.Error("Original model was mutated")
	}
}

func TestModel_Builder_AddHousehold(t *testing.T) {
	// Create user without household
	original := Model{
		id:          uuid.New(),
		email:       "test@example.com",
		displayName: "Test User",
		householdId: nil,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	if original.HasHousehold() {
		t.Fatal("Original should not have household")
	}

	// Add household via builder
	householdId := uuid.New()
	modified, err := original.Builder().
		SetHouseholdId(householdId).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify household was added
	if !modified.HasHousehold() {
		t.Error("Expected household to be added")
	}
	if *modified.HouseholdId() != householdId {
		t.Errorf("HouseholdId = %v, want %v", modified.HouseholdId(), householdId)
	}

	// Verify original is unchanged
	if original.HasHousehold() {
		t.Error("Original model was mutated")
	}
}

func TestModel_Builder_RemoveHousehold(t *testing.T) {
	// Create user with household
	householdId := uuid.New()
	original := Model{
		id:          uuid.New(),
		email:       "test@example.com",
		displayName: "Test User",
		householdId: &householdId,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	if !original.HasHousehold() {
		t.Fatal("Original should have household")
	}

	// Remove household via builder
	modified, err := original.Builder().
		ClearHouseholdId().
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify household was removed
	if modified.HasHousehold() {
		t.Error("Expected household to be removed")
	}

	// Verify original is unchanged
	if !original.HasHousehold() {
		t.Error("Original model was mutated")
	}
}

func TestBuilder_DefaultValues(t *testing.T) {
	model1, err := NewBuilder().
		SetEmail("test@example.com").
		SetDisplayName("Test User").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Small delay to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	model2, err := NewBuilder().
		SetEmail("test2@example.com").
		SetDisplayName("Test User 2").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify different UUIDs are generated
	if model1.Id() == model2.Id() {
		t.Error("Expected different UUIDs to be generated")
	}

	// Verify timestamps are generated and different
	if model1.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be generated for model1")
	}
	if model2.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be generated for model2")
	}
	if model1.CreatedAt().Equal(model2.CreatedAt()) {
		t.Error("Expected different timestamps (may fail if execution is very fast)")
	}
}

func TestBuilder_SetTimestamps(t *testing.T) {
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	model, err := NewBuilder().
		SetEmail("test@example.com").
		SetDisplayName("Test User").
		SetCreatedAt(createdAt).
		SetUpdatedAt(updatedAt).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if !model.CreatedAt().Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", model.CreatedAt(), createdAt)
	}
	if !model.UpdatedAt().Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", model.UpdatedAt(), updatedAt)
	}
}

func TestBuilder_ValidEmailFormats(t *testing.T) {
	validEmails := []string{
		"simple@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user_name@example.com",
		"user-name@example.com",
		"user123@example456.com",
		"a@example.com",
		"test@sub.domain.example.com",
	}

	for _, email := range validEmails {
		t.Run("valid: "+email, func(t *testing.T) {
			model, err := NewBuilder().
				SetEmail(email).
				SetDisplayName("Test User").
				Build()

			if err != nil {
				t.Errorf("Build() with email %q failed: %v", email, err)
				return
			}

			if model.Email() != email {
				t.Errorf("Email = %v, want %v", model.Email(), email)
			}
		})
	}
}
