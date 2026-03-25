package membership

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupHandlerTest(t *testing.T) (*mux.Router, *gorm.DB, uuid.UUID, uuid.UUID) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	db.AutoMigrate(&Entity{})

	router := mux.NewRouter()
	si := server.GetServerInformation()
	InitializeRoutes(db)(l, si, router)

	tenantID := uuid.New()
	userID := uuid.New()
	return router, db, tenantID, userID
}

func withTenant(r *http.Request, tenantID, userID uuid.UUID) *http.Request {
	ctx := tenantctx.WithContext(r.Context(), tenantctx.New(tenantID, uuid.Nil, userID))
	return r.WithContext(ctx)
}

func TestHandlers(t *testing.T) {
	t.Run("POST /memberships", func(t *testing.T) {
		router, _, tenantID, userID := setupHandlerTest(t)

		id := uuid.New()
		householdID := uuid.New()
		body := `{"data":{"type":"memberships","id":"` + id.String() + `","attributes":{"household_id":"` + householdID.String() + `","user_id":"` + userID.String() + `","role":"owner"}}}`
		req := httptest.NewRequest(http.MethodPost, "/memberships", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/vnd.api+json")
		req = withTenant(req, tenantID, userID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			t.Fatal("expected data object in response")
		}
		if data["type"] != "memberships" {
			t.Errorf("expected type memberships, got %v", data["type"])
		}
	})

	t.Run("GET /memberships", func(t *testing.T) {
		router, db, tenantID, userID := setupHandlerTest(t)

		l, _ := test.NewNullLogger()
		ctx := tenantctx.WithContext(
			httptest.NewRequest(http.MethodGet, "/", nil).Context(),
			tenantctx.New(tenantID, uuid.Nil, userID),
		)
		p := NewProcessor(l, ctx, db)
		p.Create(tenantID, uuid.New(), userID, "owner")

		req := httptest.NewRequest(http.MethodGet, "/memberships", nil)
		req = withTenant(req, tenantID, userID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("PATCH /memberships/{id}", func(t *testing.T) {
		router, db, tenantID, userID := setupHandlerTest(t)

		l, _ := test.NewNullLogger()
		ctx := tenantctx.WithContext(
			httptest.NewRequest(http.MethodGet, "/", nil).Context(),
			tenantctx.New(tenantID, uuid.Nil, userID),
		)
		p := NewProcessor(l, ctx, db)
		m, _ := p.Create(tenantID, uuid.New(), userID, "viewer")

		body := `{"data":{"type":"memberships","id":"` + m.Id().String() + `","attributes":{"role":"admin"}}}`
		req := httptest.NewRequest(http.MethodPatch, "/memberships/"+m.Id().String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/vnd.api+json")
		req = withTenant(req, tenantID, userID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("DELETE /memberships/{id}", func(t *testing.T) {
		router, db, tenantID, userID := setupHandlerTest(t)

		l, _ := test.NewNullLogger()
		ctx := tenantctx.WithContext(
			httptest.NewRequest(http.MethodGet, "/", nil).Context(),
			tenantctx.New(tenantID, uuid.Nil, userID),
		)
		p := NewProcessor(l, ctx, db)
		m, _ := p.Create(tenantID, uuid.New(), userID, "editor")

		req := httptest.NewRequest(http.MethodDelete, "/memberships/"+m.Id().String(), nil)
		req = withTenant(req, tenantID, userID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d: %s", w.Code, w.Body.String())
		}
	})
}
