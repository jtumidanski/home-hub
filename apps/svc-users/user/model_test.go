package user

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestModel_Accessors(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now().Add(1 * time.Hour)

	model := Model{
		id:          id,
		email:       "test@example.com",
		displayName: "Test User",
		householdId: &householdId,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
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
			name:     "Email",
			accessor: func() interface{} { return model.Email() },
			want:     "test@example.com",
		},
		{
			name:     "DisplayName",
			accessor: func() interface{} { return model.DisplayName() },
			want:     "Test User",
		},
		{
			name:     "HouseholdId",
			accessor: func() interface{} { return model.HouseholdId() },
			want:     &householdId,
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

func TestModel_HasHousehold(t *testing.T) {
	householdId := uuid.New()

	tests := []struct {
		name        string
		householdId *uuid.UUID
		want        bool
	}{
		{
			name:        "with household",
			householdId: &householdId,
			want:        true,
		},
		{
			name:        "without household",
			householdId: nil,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model{
				id:          uuid.New(),
				email:       "test@example.com",
				displayName: "Test User",
				householdId: tt.householdId,
				createdAt:   time.Now(),
				updatedAt:   time.Now(),
			}

			if got := model.HasHousehold(); got != tt.want {
				t.Errorf("HasHousehold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_ValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "valid email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
			want:  true,
		},
		{
			name:  "valid email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "valid email with hyphen",
			email: "user-name@example.com",
			want:  true,
		},
		{
			name:  "valid email with numbers",
			email: "user123@example123.com",
			want:  true,
		},
		{
			name:  "invalid email - no @",
			email: "userexample.com",
			want:  false,
		},
		{
			name:  "invalid email - no domain",
			email: "user@",
			want:  false,
		},
		{
			name:  "invalid email - no TLD",
			email: "user@example",
			want:  false,
		},
		{
			name:  "invalid email - no local part",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "invalid email - empty",
			email: "",
			want:  false,
		},
		{
			name:  "invalid email - spaces",
			email: "user @example.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model{
				id:          uuid.New(),
				email:       tt.email,
				displayName: "Test User",
				createdAt:   time.Now(),
				updatedAt:   time.Now(),
			}

			if got := model.ValidEmail(); got != tt.want {
				t.Errorf("ValidEmail() for %q = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()

	tests := []struct {
		name        string
		model       Model
		wantContain []string
	}{
		{
			name: "with household",
			model: Model{
				id:          id,
				email:       "test@example.com",
				displayName: "Test User",
				householdId: &householdId,
				createdAt:   time.Now(),
				updatedAt:   time.Now(),
			},
			wantContain: []string{
				"User[",
				id.String(),
				"test@example.com",
				"Test User",
				householdId.String(),
			},
		},
		{
			name: "without household",
			model: Model{
				id:          id,
				email:       "test@example.com",
				displayName: "Test User",
				householdId: nil,
				createdAt:   time.Now(),
				updatedAt:   time.Now(),
			},
			wantContain: []string{
				"User[",
				id.String(),
				"test@example.com",
				"Test User",
				"none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.String()
			for _, want := range tt.wantContain {
				if !contains(got, want) {
					t.Errorf("String() = %q, want to contain %q", got, want)
				}
			}
		})
	}
}

func TestModel_MarshalJSON(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()
	createdAt := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		model   Model
		wantErr bool
	}{
		{
			name: "with household",
			model: Model{
				id:          id,
				email:       "test@example.com",
				displayName: "Test User",
				householdId: &householdId,
				createdAt:   createdAt,
				updatedAt:   updatedAt,
			},
			wantErr: false,
		},
		{
			name: "without household",
			model: Model{
				id:          id,
				email:       "test@example.com",
				displayName: "Test User",
				householdId: nil,
				createdAt:   createdAt,
				updatedAt:   updatedAt,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify JSON contains expected fields
				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("Failed to unmarshal JSON: %v", err)
				}

				// Check required fields
				if result["id"] != id.String() {
					t.Errorf("JSON id = %v, want %v", result["id"], id.String())
				}
				if result["email"] != tt.model.email {
					t.Errorf("JSON email = %v, want %v", result["email"], tt.model.email)
				}
				if result["displayName"] != tt.model.displayName {
					t.Errorf("JSON displayName = %v, want %v", result["displayName"], tt.model.displayName)
				}

				// Check optional household field
				if tt.model.householdId != nil {
					if result["householdId"] != householdId.String() {
						t.Errorf("JSON householdId = %v, want %v", result["householdId"], householdId.String())
					}
				}
			}
		})
	}
}

func TestModel_UnmarshalJSON(t *testing.T) {
	id := uuid.New()
	householdId := uuid.New()

	tests := []struct {
		name    string
		json    string
		wantErr bool
		verify  func(t *testing.T, model Model)
	}{
		{
			name: "valid JSON with household",
			json: `{
				"id": "` + id.String() + `",
				"email": "test@example.com",
				"displayName": "Test User",
				"householdId": "` + householdId.String() + `",
				"createdAt": "2025-01-01T12:00:00Z",
				"updatedAt": "2025-01-02T12:00:00Z"
			}`,
			wantErr: false,
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
					t.Error("Expected household to be present")
				}
				if model.HouseholdId() == nil || *model.HouseholdId() != householdId {
					t.Errorf("HouseholdId = %v, want %v", model.HouseholdId(), &householdId)
				}
			},
		},
		{
			name: "valid JSON without household",
			json: `{
				"id": "` + id.String() + `",
				"email": "test@example.com",
				"displayName": "Test User",
				"createdAt": "2025-01-01T12:00:00Z",
				"updatedAt": "2025-01-02T12:00:00Z"
			}`,
			wantErr: false,
			verify: func(t *testing.T, model Model) {
				if model.Id() != id {
					t.Errorf("Id = %v, want %v", model.Id(), id)
				}
				if model.HasHousehold() {
					t.Error("Expected no household")
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
		id:          id1,
		email:       "test1@example.com",
		displayName: "Test User 1",
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	model2 := Model{
		id:          id1, // Same ID
		email:       "test2@example.com",
		displayName: "Test User 2",
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	model3 := Model{
		id:          id2, // Different ID
		email:       "test1@example.com",
		displayName: "Test User 1",
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
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
	if restored.Email() != original.Email() {
		t.Errorf("Round-trip email = %v, want %v", restored.Email(), original.Email())
	}
	if restored.DisplayName() != original.DisplayName() {
		t.Errorf("Round-trip displayName = %v, want %v", restored.DisplayName(), original.DisplayName())
	}
	if !restored.HasHousehold() || *restored.HouseholdId() != *original.HouseholdId() {
		t.Errorf("Round-trip householdId = %v, want %v", restored.HouseholdId(), original.HouseholdId())
	}
	if !restored.CreatedAt().Equal(original.CreatedAt()) {
		t.Errorf("Round-trip createdAt = %v, want %v", restored.CreatedAt(), original.CreatedAt())
	}
	if !restored.UpdatedAt().Equal(original.UpdatedAt()) {
		t.Errorf("Round-trip updatedAt = %v, want %v", restored.UpdatedAt(), original.UpdatedAt())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
