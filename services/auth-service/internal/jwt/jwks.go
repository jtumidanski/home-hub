package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

// JWK represents a single JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// BuildJWKS creates a JWKS from the issuer's public key.
func BuildJWKS(issuer *Issuer) JWKS {
	pub := issuer.PublicKey()
	return JWKS{
		Keys: []JWK{
			rsaPublicKeyToJWK(pub, issuer.Kid()),
		},
	}
}

func rsaPublicKeyToJWK(pub *rsa.PublicKey, kid string) JWK {
	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}
