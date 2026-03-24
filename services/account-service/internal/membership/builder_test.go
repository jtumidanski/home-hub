package membership

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
			name: "valid membership",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetUserID(uuid.New()).
					SetRole("owner").
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "missing household ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserID(uuid.New()).
					SetRole("owner").
					Build()
			},
			wantErr: ErrHouseholdIDRequired,
		},
		{
			name: "missing user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetHouseholdID(uuid.New()).
					SetRole("owner").
					Build()
			},
			wantErr: ErrUserIDRequired,
		},
		{
			name: "missing role",
			build: func() (Model, error) {
				return NewBuilder().
					SetHouseholdID(uuid.New()).
					SetUserID(uuid.New()).
					Build()
			},
			wantErr: ErrRoleRequired,
		},
		{
			name: "nil household ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetHouseholdID(uuid.Nil).
					SetUserID(uuid.New()).
					SetRole("viewer").
					Build()
			},
			wantErr: ErrHouseholdIDRequired,
		},
		{
			name: "nil user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetHouseholdID(uuid.New()).
					SetUserID(uuid.Nil).
					SetRole("viewer").
					Build()
			},
			wantErr: ErrUserIDRequired,
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
			if m.Role() == "" {
				t.Error("expected non-empty role")
			}
		})
	}
}
