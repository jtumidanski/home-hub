package household

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Builder
		verify func(t *testing.T, model Model)
	}{
		{
			name: "valid household with all fields",
			setup: func() *Builder {
				id := uuid.New()
				createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

				return NewBuilder().
					SetId(id).
					SetName("Smith Family").
					SetCreatedAt(createdAt).
					SetUpdatedAt(updatedAt)
			},
			verify: func(t *testing.T, model Model) {
				if model.Name() != "Smith Family" {
					t.Errorf("Name = %v, want %v", model.Name(), "Smith Family")
				}
			},
		},
		{
			name: "valid household with minimal fields",
			setup: func() *Builder {
				return NewBuilder().SetName("Test Household")
			},
			verify: func(t *testing.T, model Model) {
				if model.Name() != "Test Household" {
					t.Errorf("Name = %v, want %v", model.Name(), "Test Household")
				}
			},
		},
		{
			name: "name with whitespace is trimmed",
			setup: func() *Builder {
				return NewBuilder().SetName("  Trimmed Household  ")
			},
			verify: func(t *testing.T, model Model) {
				if model.Name() != "Trimmed Household" {
					t.Errorf("Name = %v, want %v (trimmed)", model.Name(), "Trimmed Household")
				}
			},
		},
		{
			name: "generates UUID if not provided",
			setup: func() *Builder {
				return NewBuilder().SetName("Test Household")
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
				return NewBuilder().SetName("Test Household")
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
			name: "missing name",
			setup: func() *Builder {
				return NewBuilder()
			},
			wantErr: ErrNameRequired,
		},
		{
			name: "name only whitespace",
			setup: func() *Builder {
				return NewBuilder().SetName("   ")
			},
			wantErr: ErrNameEmpty,
		},
		{
			name: "empty name string",
			setup: func() *Builder {
				return NewBuilder().SetName("")
			},
			wantErr: ErrNameEmpty,
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
	now := time.Now()

	result := builder.
		SetId(id).
		SetName("Test Household").
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
	if model.Name() != "Test Household" {
		t.Errorf("Name = %v, want %v", model.Name(), "Test Household")
	}
}

func TestModel_Builder(t *testing.T) {
	// Create original model
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Model{
		id:        id,
		name:      "Original Household",
		createdAt: createdAt,
		updatedAt: updatedAt,
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
	if model.Name() != original.Name() {
		t.Errorf("Name = %v, want %v", model.Name(), original.Name())
	}
	if !model.CreatedAt().Equal(original.CreatedAt()) {
		t.Errorf("CreatedAt = %v, want %v", model.CreatedAt(), original.CreatedAt())
	}
	if !model.UpdatedAt().Equal(original.UpdatedAt()) {
		t.Errorf("UpdatedAt = %v, want %v", model.UpdatedAt(), original.UpdatedAt())
	}
}

func TestModel_Builder_ModificationFlow(t *testing.T) {
	// Create original model
	original := Model{
		id:        uuid.New(),
		name:      "Original Household",
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	// Modify via builder
	modified, err := original.Builder().
		SetName("Modified Household").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify modifications
	if modified.Name() != "Modified Household" {
		t.Errorf("Name = %v, want %v", modified.Name(), "Modified Household")
	}

	// Verify ID is preserved
	if modified.Id() != original.Id() {
		t.Errorf("Id changed during modification: %v -> %v", original.Id(), modified.Id())
	}

	// Verify original is unchanged (immutability)
	if original.Name() != "Original Household" {
		t.Error("Original model was mutated")
	}
}

func TestBuilder_DefaultValues(t *testing.T) {
	model1, err := NewBuilder().
		SetName("Household 1").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Small delay to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	model2, err := NewBuilder().
		SetName("Household 2").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify different UUIDs are generated
	if model1.Id() == model2.Id() {
		t.Error("Expected different UUIDs to be generated")
	}

	// Verify timestamps are generated
	if model1.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be generated for model1")
	}
	if model2.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be generated for model2")
	}
}

func TestBuilder_SetTimestamps(t *testing.T) {
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	model, err := NewBuilder().
		SetName("Test Household").
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

func TestBuilder_SetId(t *testing.T) {
	expectedId := uuid.New()

	model, err := NewBuilder().
		SetId(expectedId).
		SetName("Test Household").
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if model.Id() != expectedId {
		t.Errorf("Id = %v, want %v", model.Id(), expectedId)
	}
}

func TestBuilder_NameValidation(t *testing.T) {
	validNames := []string{
		"Smith Family",
		"A",
		"Household 123",
		"Test-Household",
		"Test_Household",
		"Test.Household",
	}

	for _, name := range validNames {
		t.Run("valid: "+name, func(t *testing.T) {
			model, err := NewBuilder().
				SetName(name).
				Build()

			if err != nil {
				t.Errorf("Build() with name %q failed: %v", name, err)
				return
			}

			if model.Name() != name {
				t.Errorf("Name = %v, want %v", model.Name(), name)
			}
		})
	}
}
