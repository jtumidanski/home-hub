package connection

import (
	"testing"
	"time"
)

func TestCheckManualSyncAllowed(t *testing.T) {
	boundary := time.Now().Add(-manualSyncCooldown)
	past := time.Now().Add(-6 * time.Minute)
	recent := time.Now().Add(-2 * time.Minute)

	tests := []struct {
		name      string
		lastSync  *time.Time
		expectErr error
	}{
		{"never synced", nil, nil},
		{"cooldown expired", &past, nil},
		{"within cooldown", &recent, ErrSyncRateLimited},
		{"at boundary", &boundary, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{lastSyncAt: tc.lastSync}
			p := &Processor{}
			err := p.CheckManualSyncAllowed(m)
			if err != tc.expectErr {
				t.Fatalf("expected %v, got %v", tc.expectErr, err)
			}
		})
	}
}
