package connection

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func seedConnection(t *testing.T, db *gorm.DB, mut func(*Entity)) Entity {
	t.Helper()
	now := time.Now().UTC()
	e := Entity{
		Id:              uuid.New(),
		TenantId:        uuid.New(),
		HouseholdId:     uuid.New(),
		UserId:          uuid.New(),
		Provider:        "google",
		Status:          "connected",
		Email:           "user@example.com",
		AccessToken:     "encrypted-access",
		RefreshToken:    "encrypted-refresh",
		TokenExpiry:     now.Add(time.Hour),
		UserDisplayName: "Test User",
		UserColor:       "#4285F4",
		WriteAccess:     true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if mut != nil {
		mut(&e)
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("failed to seed connection: %v", err)
	}
	return e
}

func reload(t *testing.T, db *gorm.DB, id uuid.UUID) Entity {
	t.Helper()
	var e Entity
	if err := db.WithContext(database.WithoutTenantFilter(context.Background())).First(&e, "id = ?", id).Error; err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	return e
}

func strPtr(s string) *string { return &s }

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

func TestRecordSyncAttempt_SetsTimestamp(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	e := seedConnection(t, db, nil)

	at := time.Now().UTC().Truncate(time.Second)
	if err := p.RecordSyncAttempt(e.Id, at); err != nil {
		t.Fatalf("RecordSyncAttempt: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.LastSyncAttemptAt == nil {
		t.Fatal("expected LastSyncAttemptAt to be set")
	}
	if !got.LastSyncAttemptAt.Equal(at) {
		t.Errorf("expected %v, got %v", at, *got.LastSyncAttemptAt)
	}
}

func TestRecordSyncSuccess_ClearsErrorAndResetsCounter(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	prior := time.Now().UTC().Add(-time.Hour)
	e := seedConnection(t, db, func(e *Entity) {
		e.Status = "error"
		e.ErrorCode = strPtr("refresh_http_error")
		e.ErrorMessage = strPtr("boom")
		e.LastErrorAt = &prior
		e.ConsecutiveFailures = 4
	})

	at := time.Now().UTC().Truncate(time.Second)
	if err := p.RecordSyncSuccess(e.Id, 17, at); err != nil {
		t.Fatalf("RecordSyncSuccess: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "connected" {
		t.Errorf("expected status connected, got %s", got.Status)
	}
	if got.ErrorCode != nil || got.ErrorMessage != nil || got.LastErrorAt != nil {
		t.Errorf("expected error fields cleared, got code=%v msg=%v at=%v", got.ErrorCode, got.ErrorMessage, got.LastErrorAt)
	}
	if got.ConsecutiveFailures != 0 {
		t.Errorf("expected counter reset, got %d", got.ConsecutiveFailures)
	}
	if got.LastSyncAt == nil || !got.LastSyncAt.Equal(at) {
		t.Errorf("expected LastSyncAt=%v, got %v", at, got.LastSyncAt)
	}
	if got.LastSyncAttemptAt == nil || !got.LastSyncAttemptAt.Equal(at) {
		t.Errorf("expected LastSyncAttemptAt=%v, got %v", at, got.LastSyncAttemptAt)
	}
	if got.LastSyncEventCount != 17 {
		t.Errorf("expected event count 17, got %d", got.LastSyncEventCount)
	}
}

func TestRecordSyncFailure_TransientUnderThresholdLeavesStatus(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	e := seedConnection(t, db, nil)

	at := time.Now().UTC().Truncate(time.Second)
	if err := p.RecordSyncFailure(e.Id, "refresh_http_error", "503 Service Unavailable", at); err != nil {
		t.Fatalf("RecordSyncFailure: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "connected" {
		t.Errorf("expected status connected (under threshold), got %s", got.Status)
	}
	if got.ConsecutiveFailures != 1 {
		t.Errorf("expected counter 1, got %d", got.ConsecutiveFailures)
	}
	if got.ErrorCode == nil || *got.ErrorCode != "refresh_http_error" {
		t.Errorf("expected error_code refresh_http_error, got %v", got.ErrorCode)
	}
}

func TestRecordSyncFailure_TransientReachesThresholdSetsError(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	e := seedConnection(t, db, func(e *Entity) {
		e.ConsecutiveFailures = 2
	})

	at := time.Now().UTC().Truncate(time.Second)
	if err := p.RecordSyncFailure(e.Id, "refresh_http_error", "boom", at); err != nil {
		t.Fatalf("RecordSyncFailure: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "error" {
		t.Errorf("expected status error at threshold, got %s", got.Status)
	}
	if got.ConsecutiveFailures != FailureEscalationThreshold {
		t.Errorf("expected counter %d, got %d", FailureEscalationThreshold, got.ConsecutiveFailures)
	}
}

func TestRecordSyncFailure_HardFailureSetsDisconnectedAndForcesCounter(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	e := seedConnection(t, db, nil)

	at := time.Now().UTC().Truncate(time.Second)
	if err := p.RecordSyncFailure(e.Id, "token_revoked", "invalid_grant", at); err != nil {
		t.Fatalf("RecordSyncFailure: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "disconnected" {
		t.Errorf("expected status disconnected, got %s", got.Status)
	}
	if got.ConsecutiveFailures != FailureEscalationThreshold {
		t.Errorf("expected counter forced to %d, got %d", FailureEscalationThreshold, got.ConsecutiveFailures)
	}
}

func TestRecordSyncFailure_HardFailurePreservesHigherCounter(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	e := seedConnection(t, db, func(e *Entity) {
		e.ConsecutiveFailures = 5
	})

	at := time.Now().UTC().Truncate(time.Second)
	if err := p.RecordSyncFailure(e.Id, "token_revoked", "invalid_grant", at); err != nil {
		t.Fatalf("RecordSyncFailure: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "disconnected" {
		t.Errorf("expected status disconnected, got %s", got.Status)
	}
	if got.ConsecutiveFailures != 6 {
		t.Errorf("expected counter incremented to 6, got %d", got.ConsecutiveFailures)
	}
}

func TestRecordSyncFailure_TruncatesLongMessage(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	e := seedConnection(t, db, nil)

	long := make([]byte, 800)
	for i := range long {
		long[i] = 'x'
	}
	if err := p.RecordSyncFailure(e.Id, "unknown", string(long), time.Now().UTC()); err != nil {
		t.Fatalf("RecordSyncFailure: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.ErrorMessage == nil || len(*got.ErrorMessage) != errorMessageMaxLen {
		t.Errorf("expected message truncated to %d chars, got %v", errorMessageMaxLen, got.ErrorMessage)
	}
}

func TestClearErrorState_ResetsAllFields(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	prior := time.Now().UTC().Add(-time.Hour)
	e := seedConnection(t, db, func(e *Entity) {
		e.Status = "disconnected"
		e.ErrorCode = strPtr("token_revoked")
		e.ErrorMessage = strPtr("revoked")
		e.LastErrorAt = &prior
		e.ConsecutiveFailures = 7
	})

	if err := p.ClearErrorState(e.Id); err != nil {
		t.Fatalf("ClearErrorState: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "connected" {
		t.Errorf("expected status connected, got %s", got.Status)
	}
	if got.ErrorCode != nil || got.ErrorMessage != nil || got.LastErrorAt != nil {
		t.Errorf("expected error fields cleared, got code=%v msg=%v at=%v", got.ErrorCode, got.ErrorMessage, got.LastErrorAt)
	}
	if got.ConsecutiveFailures != 0 {
		t.Errorf("expected counter reset, got %d", got.ConsecutiveFailures)
	}
}

func TestUpdateTokensAndWriteAccess_ClearsErrorState(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	prior := time.Now().UTC().Add(-time.Hour)
	e := seedConnection(t, db, func(e *Entity) {
		e.Status = "disconnected"
		e.ErrorCode = strPtr("token_revoked")
		e.ErrorMessage = strPtr("revoked")
		e.LastErrorAt = &prior
		e.ConsecutiveFailures = 4
	})

	if err := p.UpdateTokensAndWriteAccess(e.Id, "new-access", "new-refresh", time.Now().UTC().Add(time.Hour), true); err != nil {
		t.Fatalf("UpdateTokensAndWriteAccess: %v", err)
	}

	got := reload(t, db, e.Id)
	if got.Status != "connected" {
		t.Errorf("expected status connected, got %s", got.Status)
	}
	if got.ErrorCode != nil || got.ErrorMessage != nil || got.LastErrorAt != nil {
		t.Errorf("expected error fields cleared, got code=%v msg=%v at=%v", got.ErrorCode, got.ErrorMessage, got.LastErrorAt)
	}
	if got.ConsecutiveFailures != 0 {
		t.Errorf("expected counter reset, got %d", got.ConsecutiveFailures)
	}
	if got.AccessToken != "new-access" || got.RefreshToken != "new-refresh" {
		t.Errorf("expected tokens updated, got access=%s refresh=%s", got.AccessToken, got.RefreshToken)
	}
}
