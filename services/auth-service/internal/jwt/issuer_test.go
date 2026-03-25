package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func generateTestKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	return string(pem.EncodeToMemory(block))
}

func TestNewIssuer_ValidKey(t *testing.T) {
	pemKey := generateTestKey(t)
	issuer, err := NewIssuer(pemKey, "test-kid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issuer.Kid() != "test-kid" {
		t.Errorf("expected kid test-kid, got %s", issuer.Kid())
	}
}

func TestNewIssuer_InvalidKey(t *testing.T) {
	_, err := NewIssuer("not-a-valid-pem", "test-kid")
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestIssue_ProducesValidJWT(t *testing.T) {
	pemKey := generateTestKey(t)
	issuer, err := NewIssuer(pemKey, "test-kid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()

	tokenStr, err := issuer.Issue(userID, tenantID, householdID)
	if err != nil {
		t.Fatalf("unexpected error issuing: %v", err)
	}

	// Parse and validate
	claims := &Claims{}
	token, err := jwtgo.ParseWithClaims(tokenStr, claims, func(token *jwtgo.Token) (interface{}, error) {
		return issuer.PublicKey(), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !token.Valid {
		t.Error("token is not valid")
	}
	if claims.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, claims.UserID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("expected tenant ID %s, got %s", tenantID, claims.TenantID)
	}
	if claims.HouseholdID != householdID {
		t.Errorf("expected household ID %s, got %s", householdID, claims.HouseholdID)
	}
	if claims.Issuer != "home-hub-auth" {
		t.Errorf("expected issuer home-hub-auth, got %s", claims.Issuer)
	}
}

func TestIssue_IncludesKid(t *testing.T) {
	pemKey := generateTestKey(t)
	issuer, err := NewIssuer(pemKey, "my-kid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokenStr, err := issuer.Issue(uuid.New(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parser := jwtgo.NewParser()
	token, _, err := parser.ParseUnverified(tokenStr, &Claims{})
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	kid, ok := token.Header["kid"].(string)
	if !ok || kid != "my-kid" {
		t.Errorf("expected kid my-kid, got %v", token.Header["kid"])
	}
}

func TestBuildJWKS(t *testing.T) {
	pemKey := generateTestKey(t)
	issuer, _ := NewIssuer(pemKey, "jwks-kid")

	jwks := BuildJWKS(issuer)
	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(jwks.Keys))
	}
	key := jwks.Keys[0]
	if key.Kid != "jwks-kid" {
		t.Errorf("expected kid jwks-kid, got %s", key.Kid)
	}
	if key.Kty != "RSA" {
		t.Errorf("expected kty RSA, got %s", key.Kty)
	}
	if key.Alg != "RS256" {
		t.Errorf("expected alg RS256, got %s", key.Alg)
	}
	if key.Use != "sig" {
		t.Errorf("expected use sig, got %s", key.Use)
	}
	if key.N == "" {
		t.Error("expected non-empty N")
	}
	if key.E == "" {
		t.Error("expected non-empty E")
	}
}
