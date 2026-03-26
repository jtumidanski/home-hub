package authflow

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/externalidentity"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/oidc"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/refreshtoken"
	"github.com/jtumidanski/home-hub/services/auth-service/internal/user"
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
	if err := db.AutoMigrate(&user.Entity{}, &externalidentity.Entity{}, &refreshtoken.Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func testIssuer(t *testing.T) *authjwt.Issuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	pemKey := string(pem.EncodeToMemory(block))

	issuer, err := authjwt.NewIssuer(pemKey, "test-kid")
	if err != nil {
		t.Fatalf("failed to create issuer: %v", err)
	}
	return issuer
}

func TestHandleCallback(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	issuer := testIssuer(t)
	p := NewProcessor(l, context.Background(), db, issuer)

	userInfo := &oidc.UserInfo{
		Subject:     "google-sub-123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		GivenName:   "Test",
		FamilyName:  "User",
		AvatarURL:   "https://example.com/avatar.png",
	}

	result, err := p.HandleCallbackWithUserInfo(userInfo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestHandleCallback_IdempotentIdentityLink(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	issuer := testIssuer(t)
	p := NewProcessor(l, context.Background(), db, issuer)

	userInfo := &oidc.UserInfo{
		Subject:     "google-sub-idem",
		Email:       "idem@example.com",
		DisplayName: "Idem User",
		GivenName:   "Idem",
		FamilyName:  "User",
	}

	// First callback creates user and identity
	_, err := p.HandleCallbackWithUserInfo(userInfo)
	if err != nil {
		t.Fatalf("first callback: %v", err)
	}

	// Second callback should succeed (identity already linked)
	result, err := p.HandleCallbackWithUserInfo(userInfo)
	if err != nil {
		t.Fatalf("second callback: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token on second callback")
	}
}

func TestHandleRefresh(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	issuer := testIssuer(t)

	// Create a user and refresh token first
	userProc := user.NewProcessor(l, context.Background(), db)
	u, err := userProc.FindOrCreate("refresh@example.com", "Refresh User", "Refresh", "User", "")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	rtProc := refreshtoken.NewProcessor(l, context.Background(), db)
	oldRaw, err := rtProc.Create(u.Id())
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	p := NewProcessor(l, context.Background(), db, issuer)
	result, err := p.HandleRefresh(oldRaw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.RefreshToken == oldRaw {
		t.Error("expected rotated refresh token to differ from old")
	}
}

func TestHandleRefresh_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	issuer := testIssuer(t)
	p := NewProcessor(l, context.Background(), db, issuer)

	_, err := p.HandleRefresh("invalid-token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

func TestHandleLogout(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	issuer := testIssuer(t)

	userID := uuid.New()
	rtProc := refreshtoken.NewProcessor(l, context.Background(), db)
	raw, err := rtProc.Create(userID)
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	p := NewProcessor(l, context.Background(), db, issuer)
	if err := p.HandleLogout(userID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Token should be invalid after logout
	_, err = rtProc.Validate(raw)
	if err == nil {
		t.Error("expected token to be revoked after logout")
	}
}
