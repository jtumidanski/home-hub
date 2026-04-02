package oauthstate

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(10 * time.Minute)

	tests := []struct {
		name      string
		build     func() (Model, error)
		expectErr error
		validate  func(t *testing.T, m Model)
	}{
		{
			"requires redirect URI",
			func() (Model, error) {
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetUserID(uuid.New()).
					SetExpiresAt(time.Now().Add(10 * time.Minute)).
					Build()
			},
			ErrRedirectURIRequired,
			nil,
		},
		{
			"succeeds with all fields",
			func() (Model, error) {
				return NewBuilder().
					SetId(id).
					SetTenantID(tenantID).
					SetHouseholdID(householdID).
					SetUserID(userID).
					SetRedirectURI("https://example.com/callback").
					SetExpiresAt(expiresAt).
					Build()
			},
			nil,
			func(t *testing.T, m Model) {
				if m.Id() != id {
					t.Fatalf("expected id %v, got %v", id, m.Id())
				}
				if m.TenantID() != tenantID {
					t.Fatalf("expected tenantID %v, got %v", tenantID, m.TenantID())
				}
				if m.HouseholdID() != householdID {
					t.Fatalf("expected householdID %v, got %v", householdID, m.HouseholdID())
				}
				if m.UserID() != userID {
					t.Fatalf("expected userID %v, got %v", userID, m.UserID())
				}
				if m.RedirectURI() != "https://example.com/callback" {
					t.Fatalf("expected redirect URI, got %q", m.RedirectURI())
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.build()
			if tc.expectErr != nil {
				if err != tc.expectErr {
					t.Fatalf("expected %v, got %v", tc.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.validate != nil {
				tc.validate(t, m)
			}
		})
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		offset   time.Duration
		expected bool
	}{
		{"not expired", 10 * time.Minute, false},
		{"expired", -1 * time.Minute, true},
		{"just expired", -1 * time.Second, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := NewBuilder().
				SetRedirectURI("https://example.com/callback").
				SetExpiresAt(time.Now().Add(tc.offset)).
				Build()

			if m.IsExpired() != tc.expected {
				t.Fatalf("expected IsExpired()=%v for %s", tc.expected, tc.name)
			}
		})
	}
}
