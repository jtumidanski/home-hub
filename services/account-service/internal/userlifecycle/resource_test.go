package userlifecycle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/account-service/internal/householdpreference"
	"github.com/jtumidanski/home-hub/shared/go/database"
	sharedevents "github.com/jtumidanski/home-hub/shared/go/events"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testToken = "test-internal-token"
const testTopic = "home-hub.user.lifecycle"

type stubProducer struct {
	mu    sync.Mutex
	calls []producedCall
}

type producedCall struct {
	topic   string
	key     []byte
	value   []byte
	headers map[string]string
}

func (s *stubProducer) Produce(_ context.Context, topic string, key, value []byte, headers map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, producedCall{topic: topic, key: key, value: value, headers: headers})
	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&householdpreference.Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newTestRouter(t *testing.T, db *gorm.DB, prod Producer) http.Handler {
	t.Helper()
	l, _ := test.NewNullLogger()
	r := mux.NewRouter()
	InitializeRoutes(db, prod, Config{Topic: testTopic, InternalToken: testToken})(l, r)
	return r
}

func insertRow(t *testing.T, db *gorm.DB, tenantID, userID, householdID uuid.UUID) {
	t.Helper()
	e, err := householdpreference.NewBuilder().
		SetId(uuid.New()).
		SetTenantID(tenantID).
		SetUserID(userID).
		SetHouseholdID(householdID).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	ent := e.ToEntity()
	if err := db.Create(&ent).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func TestDeletedHandlerRemovesRowsAndProduces(t *testing.T) {
	db := setupTestDB(t)
	prod := &stubProducer{}
	h := newTestRouter(t, db, prod)

	tenantID := uuid.New()
	userID := uuid.New()
	hid := uuid.New()
	// Row to delete.
	insertRow(t, db, tenantID, userID, hid)
	// Row for a different user in the same tenant — must survive.
	insertRow(t, db, tenantID, uuid.New(), hid)
	// Row for the same user in a different tenant — must survive.
	insertRow(t, db, uuid.New(), userID, hid)

	req := httptest.NewRequest(http.MethodPost, "/internal/users/"+userID.String()+"/deleted", nil)
	req.Header.Set("X-Internal-Token", testToken)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status: got %d: %s", w.Code, w.Body.String())
	}

	var remaining []householdpreference.Entity
	if err := db.Find(&remaining).Error; err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 surviving rows, got %d", len(remaining))
	}
	for _, row := range remaining {
		if row.TenantId == tenantID && row.UserId == userID {
			t.Fatalf("target row not deleted: %+v", row)
		}
	}

	prod.mu.Lock()
	calls := append([]producedCall(nil), prod.calls...)
	prod.mu.Unlock()
	if len(calls) != 1 {
		t.Fatalf("expected 1 producer call, got %d", len(calls))
	}
	if calls[0].topic != testTopic {
		t.Fatalf("topic: got %q", calls[0].topic)
	}
	var env sharedevents.Envelope
	if err := json.Unmarshal(calls[0].value, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Type != sharedevents.TypeUserDeleted {
		t.Fatalf("event type: got %s", env.Type)
	}
	var payload sharedevents.UserDeletedEvent
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.TenantID != tenantID || payload.UserID != userID {
		t.Fatalf("payload ids mismatch: %+v", payload)
	}
}

func TestDeletedHandlerMissingTokenRejects(t *testing.T) {
	db := setupTestDB(t)
	prod := &stubProducer{}
	h := newTestRouter(t, db, prod)

	req := httptest.NewRequest(http.MethodPost, "/internal/users/"+uuid.New().String()+"/deleted", nil)
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want 401", w.Code)
	}
	if len(prod.calls) != 0 {
		t.Fatalf("producer must not be invoked on auth failure")
	}
}

func TestDeletedHandlerMalformedUUIDReturns400(t *testing.T) {
	db := setupTestDB(t)
	prod := &stubProducer{}
	h := newTestRouter(t, db, prod)

	req := httptest.NewRequest(http.MethodPost, "/internal/users/not-a-uuid/deleted", nil)
	req.Header.Set("X-Internal-Token", testToken)
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", w.Code)
	}
}

func TestDeletedHandlerMissingTenantHeaderReturns400(t *testing.T) {
	db := setupTestDB(t)
	prod := &stubProducer{}
	h := newTestRouter(t, db, prod)

	req := httptest.NewRequest(http.MethodPost, "/internal/users/"+uuid.New().String()+"/deleted", nil)
	req.Header.Set("X-Internal-Token", testToken)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", w.Code)
	}
}
