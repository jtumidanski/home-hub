package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const accessTokenTTL = 15 * time.Minute

// Claims represents the JWT claims issued by the auth service.
type Claims struct {
	jwtgo.RegisteredClaims
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	TenantID    uuid.UUID `json:"tenant_id"`
	HouseholdID uuid.UUID `json:"household_id"`
}

// Issuer creates signed JWTs using an RSA private key.
type Issuer struct {
	privateKey *rsa.PrivateKey
	kid        string
}

// NewIssuer creates a JWT issuer from a PEM-encoded RSA private key.
func NewIssuer(pemKey string, kid string) (*Issuer, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, errors.New("failed to parse private key")
		}
		rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
		key = rsaKey
	}

	return &Issuer{privateKey: key, kid: kid}, nil
}

// Issue creates a signed access token.
func (i *Issuer) Issue(userID uuid.UUID, email string, tenantID, householdID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwtgo.RegisteredClaims{
			Issuer:    "home-hub-auth",
			Subject:   userID.String(),
			IssuedAt:  jwtgo.NewNumericDate(now),
			ExpiresAt: jwtgo.NewNumericDate(now.Add(accessTokenTTL)),
			ID:        uuid.New().String(),
		},
		UserID:      userID,
		Email:       email,
		TenantID:    tenantID,
		HouseholdID: householdID,
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodRS256, claims)
	token.Header["kid"] = i.kid
	return token.SignedString(i.privateKey)
}

// PublicKey returns the public key for JWKS exposure.
func (i *Issuer) PublicKey() *rsa.PublicKey {
	return &i.privateKey.PublicKey
}

// Kid returns the key ID.
func (i *Issuer) Kid() string {
	return i.kid
}

// ExtractClaimsFromCookie validates the JWT from the access_token cookie
// using the given public key and returns the claims.
func ExtractClaimsFromCookie(r *http.Request, publicKey *rsa.PublicKey) (*Claims, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	token, err := jwtgo.ParseWithClaims(cookie.Value, claims, func(token *jwtgo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtgo.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
