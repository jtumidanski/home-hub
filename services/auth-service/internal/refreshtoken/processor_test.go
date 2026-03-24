package refreshtoken

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestCreate_ReturnsToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	raw, err := p.Create(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == "" {
		t.Error("expected non-empty token")
	}
	if len(raw) != tokenLength*2 { // hex encoded
		t.Errorf("expected token length %d, got %d", tokenLength*2, len(raw))
	}
}

func TestValidate_ValidToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	raw, err := p.Create(userID)
	if err != nil {
		t.Fatalf("unexpected error creating: %v", err)
	}

	gotUserID, err := p.Validate(raw)
	if err != nil {
		t.Fatalf("unexpected error validating: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, gotUserID)
	}
}

func TestValidate_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	_, err := p.Validate("nonexistent-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidate_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	raw, _ := generateToken()
	now := time.Now().UTC()

	e := &Entity{
		Id:        uuid.New(),
		UserId:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: now.Add(-1 * time.Hour), // expired
		Revoked:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	db.Create(e)

	_, err := p.Validate(raw)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidate_RevokedToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	raw, _ := generateToken()
	now := time.Now().UTC()

	e := &Entity{
		Id:        uuid.New(),
		UserId:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: now.Add(tokenTTL),
		Revoked:   true, // revoked
		CreatedAt: now,
		UpdatedAt: now,
	}
	db.Create(e)

	_, err := p.Validate(raw)
	if err == nil {
		t.Error("expected error for revoked token")
	}
}

func TestRotate_IssuesNewToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	oldRaw, err := p.Create(userID)
	if err != nil {
		t.Fatalf("unexpected error creating: %v", err)
	}

	newRaw, gotUserID, err := p.Rotate(oldRaw)
	if err != nil {
		t.Fatalf("unexpected error rotating: %v", err)
	}
	if newRaw == oldRaw {
		t.Error("expected different token after rotation")
	}
	if gotUserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, gotUserID)
	}

	// Old token should now be invalid
	_, err = p.Validate(oldRaw)
	if err == nil {
		t.Error("expected old token to be invalid after rotation")
	}

	// New token should be valid
	_, err = p.Validate(newRaw)
	if err != nil {
		t.Fatalf("expected new token to be valid: %v", err)
	}
}

func TestRevokeAllForUser(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	raw1, _ := p.Create(userID)
	raw2, _ := p.Create(userID)

	if err := p.RevokeAllForUser(userID); err != nil {
		t.Fatalf("unexpected error revoking: %v", err)
	}

	_, err := p.Validate(raw1)
	if err == nil {
		t.Error("expected token 1 to be invalid after revoke all")
	}
	_, err = p.Validate(raw2)
	if err == nil {
		t.Error("expected token 2 to be invalid after revoke all")
	}
}
