package oauthstate

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilderRequiresRedirectURI(t *testing.T) {
	_, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(uuid.New()).
		SetHouseholdID(uuid.New()).
		SetUserID(uuid.New()).
		SetExpiresAt(time.Now().Add(10 * time.Minute)).
		Build()
	if err != ErrRedirectURIRequired {
		t.Fatalf("expected ErrRedirectURIRequired, got %v", err)
	}
}

func TestBuilderSucceeds(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(10 * time.Minute)

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetHouseholdID(householdID).
		SetUserID(userID).
		SetRedirectURI("https://example.com/callback").
		SetExpiresAt(expiresAt).
		Build()
	if err != nil {
		t.Fatal(err)
	}

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
}

func TestIsExpired_NotExpired(t *testing.T) {
	m, _ := NewBuilder().
		SetRedirectURI("https://example.com/callback").
		SetExpiresAt(time.Now().Add(10 * time.Minute)).
		Build()

	if m.IsExpired() {
		t.Fatal("expected model to not be expired")
	}
}

func TestIsExpired_Expired(t *testing.T) {
	m, _ := NewBuilder().
		SetRedirectURI("https://example.com/callback").
		SetExpiresAt(time.Now().Add(-1 * time.Minute)).
		Build()

	if !m.IsExpired() {
		t.Fatal("expected model to be expired")
	}
}

func TestIsExpired_JustExpired(t *testing.T) {
	m, _ := NewBuilder().
		SetRedirectURI("https://example.com/callback").
		SetExpiresAt(time.Now().Add(-1 * time.Second)).
		Build()

	if !m.IsExpired() {
		t.Fatal("expected model to be expired when 1 second past expiry")
	}
}
