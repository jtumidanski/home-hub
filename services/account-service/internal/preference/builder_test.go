package preference

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
			name: "valid preference",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetUserID(uuid.New()).
					SetTheme("dark").
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "valid with active household",
			build: func() (Model, error) {
				hhID := uuid.New()
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetUserID(uuid.New()).
					SetTheme("light").
					SetActiveHouseholdID(&hhID).
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "missing user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetTheme("light").
					Build()
			},
			wantErr: ErrUserIDRequired,
		},
		{
			name: "nil user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserID(uuid.Nil).
					SetTheme("light").
					Build()
			},
			wantErr: ErrUserIDRequired,
		},
		{
			name: "missing theme",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserID(uuid.New()).
					Build()
			},
			wantErr: ErrThemeRequired,
		},
		{
			name: "empty theme",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserID(uuid.New()).
					SetTheme("").
					Build()
			},
			wantErr: ErrThemeRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.build()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Theme() == "" {
				t.Error("expected non-empty theme")
			}
		})
	}
}
