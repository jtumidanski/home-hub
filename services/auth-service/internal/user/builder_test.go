package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		displayName string
		wantErr     error
	}{
		{
			name:        "valid user",
			email:       "test@example.com",
			displayName: "Test User",
			wantErr:     nil,
		},
		{
			name:        "missing email",
			email:       "",
			displayName: "Test User",
			wantErr:     ErrEmailRequired,
		},
		{
			name:        "missing display name",
			email:       "test@example.com",
			displayName: "",
			wantErr:     ErrDisplayNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewBuilder().
				SetId(uuid.New()).
				SetEmail(tt.email).
				SetDisplayName(tt.displayName).
				Build()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Email() != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, m.Email())
			}
			if m.DisplayName() != tt.displayName {
				t.Errorf("expected display name %s, got %s", tt.displayName, m.DisplayName())
			}
		})
	}
}
