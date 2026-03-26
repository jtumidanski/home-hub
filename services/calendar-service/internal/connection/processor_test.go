package connection

import (
	"testing"
	"time"
)

func TestCheckManualSyncAllowed_NeverSynced(t *testing.T) {
	m := Model{lastSyncAt: nil}
	p := &Processor{}
	if err := p.CheckManualSyncAllowed(m); err != nil {
		t.Fatalf("expected nil error for never-synced connection, got %v", err)
	}
}

func TestCheckManualSyncAllowed_CooldownExpired(t *testing.T) {
	past := time.Now().Add(-6 * time.Minute)
	m := Model{lastSyncAt: &past}
	p := &Processor{}
	if err := p.CheckManualSyncAllowed(m); err != nil {
		t.Fatalf("expected nil error when cooldown expired, got %v", err)
	}
}

func TestCheckManualSyncAllowed_WithinCooldown(t *testing.T) {
	recent := time.Now().Add(-2 * time.Minute)
	m := Model{lastSyncAt: &recent}
	p := &Processor{}
	err := p.CheckManualSyncAllowed(m)
	if err == nil {
		t.Fatal("expected ErrSyncRateLimited, got nil")
	}
	if err != ErrSyncRateLimited {
		t.Fatalf("expected ErrSyncRateLimited, got %v", err)
	}
}

func TestCheckManualSyncAllowed_ExactlyAtBoundary(t *testing.T) {
	boundary := time.Now().Add(-manualSyncCooldown)
	m := Model{lastSyncAt: &boundary}
	p := &Processor{}
	// At exactly 5 minutes ago, Since() >= cooldown so should be allowed
	if err := p.CheckManualSyncAllowed(m); err != nil {
		t.Fatalf("expected nil error at cooldown boundary, got %v", err)
	}
}
