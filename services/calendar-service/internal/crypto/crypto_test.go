package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"
)

func testKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	enc, err := NewEncryptor(testKey(t))
	if err != nil {
		t.Fatal(err)
	}

	plaintext := "my-secret-oauth-token"
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	if ciphertext == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	enc, err := NewEncryptor(testKey(t))
	if err != nil {
		t.Fatal(err)
	}

	c1, _ := enc.Encrypt("same-input")
	c2, _ := enc.Encrypt("same-input")

	if c1 == c2 {
		t.Fatal("encrypting the same plaintext should produce different ciphertexts due to random nonces")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	enc1, _ := NewEncryptor(testKey(t))
	enc2, _ := NewEncryptor(testKey(t))

	ciphertext, _ := enc1.Encrypt("secret")
	_, err := enc2.Decrypt(ciphertext)
	if err == nil {
		t.Fatal("decryption with wrong key should fail")
	}
}

func TestNewEncryptorEmptyKey(t *testing.T) {
	_, err := NewEncryptor("")
	if err != ErrKeyRequired {
		t.Fatalf("expected ErrKeyRequired, got %v", err)
	}
}

func TestNewEncryptorInvalidKeyLength(t *testing.T) {
	shortKey := base64.StdEncoding.EncodeToString([]byte("too-short"))
	_, err := NewEncryptor(shortKey)
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	enc, _ := NewEncryptor(testKey(t))
	_, err := enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Fatal("should fail on invalid base64")
	}
}

func TestDecryptTooShort(t *testing.T) {
	enc, _ := NewEncryptor(testKey(t))
	short := base64.StdEncoding.EncodeToString([]byte("x"))
	_, err := enc.Decrypt(short)
	if err == nil {
		t.Fatal("should fail on ciphertext shorter than nonce")
	}
	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed sentinel, got %v", err)
	}
}

func TestDecryptCorruptedCiphertextWrapsErrDecryptFailed(t *testing.T) {
	enc, _ := NewEncryptor(testKey(t))
	_, err := enc.Decrypt("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err == nil {
		t.Fatal("expected decrypt failure")
	}
	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed sentinel, got %v", err)
	}
}
