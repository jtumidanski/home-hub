package refreshtoken

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name    string
		build   func() (Model, error)
		wantErr error
	}{
		{
			name: "valid token",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetUserId(uuid.New()).
					SetTokenHash("abc123hash").
					SetExpiresAt(time.Now().Add(24 * time.Hour)).
					Build()
			},
			wantErr: nil,
		},
		{
			name: "missing user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetTokenHash("abc123hash").
					SetExpiresAt(time.Now().Add(24 * time.Hour)).
					Build()
			},
			wantErr: ErrUserIDRequired,
		},
		{
			name: "missing token hash",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserId(uuid.New()).
					SetExpiresAt(time.Now().Add(24 * time.Hour)).
					Build()
			},
			wantErr: ErrTokenHashRequired,
		},
		{
			name: "missing expiration",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserId(uuid.New()).
					SetTokenHash("abc123hash").
					Build()
			},
			wantErr: ErrExpiresAtRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.build()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.TokenHash() != "abc123hash" {
				t.Errorf("expected token hash abc123hash, got %s", m.TokenHash())
			}
		})
	}
}
