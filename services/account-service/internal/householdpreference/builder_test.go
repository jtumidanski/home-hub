package householdpreference

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
			name: "valid household preference without default dashboard",
			build: func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetUserID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "valid with default dashboard",
			build: func() (Model, error) {
				ddID := uuid.New()
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetUserID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetDefaultDashboardID(&ddID).
					SetCreatedAt(time.Now()).
					SetUpdatedAt(time.Now()).
					Build()
			},
		},
		{
			name: "missing tenant ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetUserID(uuid.New()).
					SetHouseholdID(uuid.New()).
					Build()
			},
			wantErr: ErrTenantRequired,
		},
		{
			name: "missing user ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetTenantID(uuid.New()).
					SetHouseholdID(uuid.New()).
					Build()
			},
			wantErr: ErrUserRequired,
		},
		{
			name: "missing household ID",
			build: func() (Model, error) {
				return NewBuilder().
					SetTenantID(uuid.New()).
					SetUserID(uuid.New()).
					Build()
			},
			wantErr: ErrHouseholdRequired,
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
			if m.TenantID() == uuid.Nil {
				t.Error("expected non-nil tenant id")
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	ddID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	orig, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(uuid.New()).
		SetUserID(uuid.New()).
		SetHouseholdID(uuid.New()).
		SetDefaultDashboardID(&ddID).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	e := orig.ToEntity()
	got, err := Make(e)
	if err != nil {
		t.Fatal(err)
	}
	if got.Id() != orig.Id() || got.TenantID() != orig.TenantID() || got.UserID() != orig.UserID() || got.HouseholdID() != orig.HouseholdID() {
		t.Fatal("round-trip mismatch on ids")
	}
	if got.DefaultDashboardID() == nil || *got.DefaultDashboardID() != ddID {
		t.Fatal("round-trip mismatch on default dashboard id")
	}
	if !got.CreatedAt().Equal(orig.CreatedAt()) || !got.UpdatedAt().Equal(orig.UpdatedAt()) {
		t.Fatal("round-trip mismatch on timestamps")
	}
}
