package user

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	authjwt "github.com/jtumidanski/home-hub/services/auth-service/internal/jwt"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupResourceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	return db
}

func testResourceIssuer(t *testing.T) *authjwt.Issuer {
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

func setupRouter(t *testing.T, db *gorm.DB, issuer *authjwt.Issuer) *mux.Router {
	t.Helper()
	l, _ := test.NewNullLogger()
	router := mux.NewRouter()
	si := server.GetServerInformation()
	InitializeRoutes(db, issuer)(l, si, router)
	return router
}

func createUser(t *testing.T, db *gorm.DB, email, name string) Model {
	t.Helper()
	l, _ := test.NewNullLogger()
	proc := NewProcessor(l, context.Background(), db)
	m, err := proc.FindOrCreate(email, name, "Given", "Family", "")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return m
}

func addAuthCookie(t *testing.T, r *http.Request, issuer *authjwt.Issuer, userID uuid.UUID, email string) *http.Request {
	t.Helper()
	token, err := issuer.Issue(userID, email, uuid.Nil, uuid.Nil)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}
	r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	return r
}

func TestListUsersHandler(t *testing.T) {
	t.Run("returns users by IDs", func(t *testing.T) {
		db := setupResourceTestDB(t)
		issuer := testResourceIssuer(t)
		router := setupRouter(t, db, issuer)

		u1 := createUser(t, db, "user1@example.com", "User One")
		u2 := createUser(t, db, "user2@example.com", "User Two")

		ids := u1.Id().String() + "," + u2.Id().String()
		req := httptest.NewRequest(http.MethodGet, "/users?filter[ids]="+ids, nil)
		req = addAuthCookie(t, req, issuer, u1.Id(), u1.Email())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "user1@example.com") || !strings.Contains(body, "user2@example.com") {
			t.Error("expected both users in response")
		}
	})

	t.Run("missing filter returns 400", func(t *testing.T) {
		db := setupResourceTestDB(t)
		issuer := testResourceIssuer(t)
		router := setupRouter(t, db, issuer)

		u := createUser(t, db, "caller@example.com", "Caller")
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req = addAuthCookie(t, req, issuer, u.Id(), u.Email())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("too many IDs returns 400", func(t *testing.T) {
		db := setupResourceTestDB(t)
		issuer := testResourceIssuer(t)
		router := setupRouter(t, db, issuer)

		u := createUser(t, db, "caller2@example.com", "Caller")
		ids := make([]string, 51)
		for i := range ids {
			ids[i] = uuid.New().String()
		}
		req := httptest.NewRequest(http.MethodGet, "/users?filter[ids]="+strings.Join(ids, ","), nil)
		req = addAuthCookie(t, req, issuer, u.Id(), u.Email())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		db := setupResourceTestDB(t)
		issuer := testResourceIssuer(t)
		router := setupRouter(t, db, issuer)

		u := createUser(t, db, "caller3@example.com", "Caller")
		req := httptest.NewRequest(http.MethodGet, "/users?filter[ids]=not-a-uuid", nil)
		req = addAuthCookie(t, req, issuer, u.Id(), u.Email())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unknown IDs returns empty list", func(t *testing.T) {
		db := setupResourceTestDB(t)
		issuer := testResourceIssuer(t)
		router := setupRouter(t, db, issuer)

		u := createUser(t, db, "caller5@example.com", "Caller")
		ids := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/users?filter[ids]="+ids, nil)
		req = addAuthCookie(t, req, issuer, u.Id(), u.Email())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("unauthenticated returns 401", func(t *testing.T) {
		db := setupResourceTestDB(t)
		issuer := testResourceIssuer(t)
		router := setupRouter(t, db, issuer)

		req := httptest.NewRequest(http.MethodGet, "/users?filter[ids]="+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}
