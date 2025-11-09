package household

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestModel_Accessors(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now().Add(1 * time.Hour)

	model := Model{
		id:        id,
		name:      "Test Household",
		createdAt: createdAt,
		updatedAt: updatedAt,
	}

	tests := []struct {
		name     string
		accessor func() interface{}
		want     interface{}
	}{
		{
			name:     "Id",
			accessor: func() interface{} { return model.Id() },
			want:     id,
		},
		{
			name:     "Name",
			accessor: func() interface{} { return model.Name() },
			want:     "Test Household",
		},
		{
			name:     "CreatedAt",
			accessor: func() interface{} { return model.CreatedAt() },
			want:     createdAt,
		},
		{
			name:     "UpdatedAt",
			accessor: func() interface{} { return model.UpdatedAt() },
			want:     updatedAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.accessor()
			if got != tt.want {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	id := uuid.New()

	model := Model{
		id:        id,
		name:      "Smith Family",
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	str := model.String()

	// Verify string contains key information
	if !strings.Contains(str, "Household[") {
		t.Error("String() should start with 'Household['")
	}
	if !strings.Contains(str, id.String()) {
		t.Errorf("String() should contain ID: %s", id.String())
	}
	if !strings.Contains(str, "Smith Family") {
		t.Error("String() should contain name")
	}
}

func TestModel_MarshalJSON(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	model := Model{
		id:        id,
		name:      "Test Household",
		createdAt: createdAt,
		updatedAt: updatedAt,
	}

	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("MarshalJSON() failed: %v", err)
	}

	// Unmarshal to map to verify fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["id"] != id.String() {
		t.Errorf("JSON id = %v, want %v", result["id"], id.String())
	}
	if result["name"] != model.name {
		t.Errorf("JSON name = %v, want %v", result["name"], model.name)
	}
	if result["createdAt"] == nil {
		t.Error("JSON should contain createdAt")
	}
	if result["updatedAt"] == nil {
		t.Error("JSON should contain updatedAt")
	}
}

func TestModel_UnmarshalJSON(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name    string
		json    string
		wantErr bool
		verify  func(t *testing.T, model Model)
	}{
		{
			name: "valid JSON",
			json: `{
				"id": "` + id.String() + `",
				"name": "Test Household",
				"createdAt": "2025-01-01T12:00:00Z",
				"updatedAt": "2025-01-02T12:00:00Z"
			}`,
			wantErr: false,
			verify: func(t *testing.T, model Model) {
				if model.Id() != id {
					t.Errorf("Id = %v, want %v", model.Id(), id)
				}
				if model.Name() != "Test Household" {
					t.Errorf("Name = %v, want %v", model.Name(), "Test Household")
				}
				if model.CreatedAt().IsZero() {
					t.Error("CreatedAt should not be zero")
				}
				if model.UpdatedAt().IsZero() {
					t.Error("UpdatedAt should not be zero")
				}
			},
		},
		{
			name:    "invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
			verify:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var model Model
			err := json.Unmarshal([]byte(tt.json), &model)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.verify != nil {
				tt.verify(t, model)
			}
		})
	}
}

func TestModel_Is(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	model1 := Model{
		id:        id1,
		name:      "Household 1",
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	model2 := Model{
		id:        id1, // Same ID
		name:      "Household 2",
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	model3 := Model{
		id:        id2, // Different ID
		name:      "Household 1",
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	tests := []struct {
		name  string
		model Model
		other Model
		want  bool
	}{
		{
			name:  "same ID - should be equal",
			model: model1,
			other: model2,
			want:  true,
		},
		{
			name:  "different ID - should not be equal",
			model: model1,
			other: model3,
			want:  false,
		},
		{
			name:  "same model - should be equal",
			model: model1,
			other: model1,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.Is(tt.other); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_JSONRoundTrip(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	original := Model{
		id:        id,
		name:      "Test Household",
		createdAt: createdAt,
		updatedAt: updatedAt,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back to model
	var restored Model
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify all fields match
	if !restored.Is(original) {
		t.Error("Round-trip IDs don't match")
	}
	if restored.Name() != original.Name() {
		t.Errorf("Round-trip name = %v, want %v", restored.Name(), original.Name())
	}
	if !restored.CreatedAt().Equal(original.CreatedAt()) {
		t.Errorf("Round-trip createdAt = %v, want %v", restored.CreatedAt(), original.CreatedAt())
	}
	if !restored.UpdatedAt().Equal(original.UpdatedAt()) {
		t.Errorf("Round-trip updatedAt = %v, want %v", restored.UpdatedAt(), original.UpdatedAt())
	}
}
