package sync

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
)

func TestClassifyTokenError_InvalidGrantHard(t *testing.T) {
	e := &Engine{}
	err := &googlecal.TokenRefreshError{StatusCode: http.StatusBadRequest, OAuthError: "invalid_grant", Body: "{}"}
	code, hard := e.classifyTokenError(err)
	if code != "token_revoked" || !hard {
		t.Fatalf("expected token_revoked/hard, got %s/%v", code, hard)
	}
}

func TestClassifyTokenError_Unauthorized401Hard(t *testing.T) {
	e := &Engine{}
	err := &googlecal.TokenRefreshError{StatusCode: http.StatusUnauthorized, OAuthError: "", Body: ""}
	code, hard := e.classifyTokenError(err)
	if code != "refresh_unauthorized" || !hard {
		t.Fatalf("expected refresh_unauthorized/hard, got %s/%v", code, hard)
	}
}

func TestClassifyTokenError_5xxTransient(t *testing.T) {
	e := &Engine{}
	err := &googlecal.TokenRefreshError{StatusCode: http.StatusInternalServerError, OAuthError: "", Body: "boom"}
	code, hard := e.classifyTokenError(err)
	if code != "refresh_http_error" || hard {
		t.Fatalf("expected refresh_http_error/transient, got %s/%v", code, hard)
	}
}

func TestClassifyTokenError_TransportTransient(t *testing.T) {
	e := &Engine{}
	code, hard := e.classifyTokenError(errors.New("dial tcp: connection reset"))
	if code != "unknown" || hard {
		t.Fatalf("expected unknown/transient, got %s/%v", code, hard)
	}
}

func TestClassifyTokenError_DecryptFailedHard(t *testing.T) {
	e := &Engine{}
	wrapped := errors.New("wrapper")
	// Build an error that wraps ErrDecryptFailed via fmt.Errorf — same pattern Decrypt uses.
	err := errWrap(crypto.ErrDecryptFailed, wrapped)
	code, hard := e.classifyTokenError(err)
	if code != "token_decrypt_failed" || !hard {
		t.Fatalf("expected token_decrypt_failed/hard, got %s/%v", code, hard)
	}
}

// errWrap mirrors the wrapping pattern in crypto.Decrypt so the test exercises errors.Is.
func errWrap(sentinel, inner error) error {
	return wrappedErr{sentinel: sentinel, inner: inner}
}

type wrappedErr struct {
	sentinel error
	inner    error
}

func (w wrappedErr) Error() string { return w.sentinel.Error() + ": " + w.inner.Error() }
func (w wrappedErr) Unwrap() error { return w.sentinel }
