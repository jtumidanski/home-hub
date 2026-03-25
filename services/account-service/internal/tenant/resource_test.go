package tenant

import (
	"bytes"
	"context"
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

func setupHandlerTest(t *testing.T) (*mux.Router, *gorm.DB) {
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

	return router, db
}

func withTenant(r *http.Request, tenantID, userID uuid.UUID) *http.Request {
	ctx := tenantctx.WithContext(r.Context(), tenantctx.New(tenantID, uuid.Nil, userID))
	return r.WithContext(ctx)
}

func createTenantInDB(t *testing.T, db *gorm.DB, name string) Model {
	t.Helper()
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)
	m, err := p.Create(name)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	return m
}

func TestHandlers(t *testing.T) {
	t.Run("POST /tenants", func(t *testing.T) {
		router, _ := setupHandlerTest(t)

		id := uuid.New()
		body := `{"data":{"type":"tenants","id":"` + id.String() + `","attributes":{"name":"Test Tenant"}}}`
		req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/vnd.api+json")
		req = withTenant(req, uuid.New(), uuid.New())
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
		if data["type"] != "tenants" {
			t.Errorf("expected type tenants, got %v", data["type"])
		}
	})

	t.Run("GET /tenants/{id}", func(t *testing.T) {
		tests := []struct {
			name       string
			setup      func(db *gorm.DB) uuid.UUID
			wantStatus int
		}{
			{
				name: "existing tenant",
				setup: func(db *gorm.DB) uuid.UUID {
					l, _ := test.NewNullLogger()
					p := NewProcessor(l, context.Background(), db)
					m, _ := p.Create("Lookup")
					return m.Id()
				},
				wantStatus: http.StatusOK,
			},
			{
				name: "non-existent tenant",
				setup: func(_ *gorm.DB) uuid.UUID {
					return uuid.New()
				},
				wantStatus: http.StatusNotFound,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				router, db := setupHandlerTest(t)
				id := tt.setup(db)
				req := httptest.NewRequest(http.MethodGet, "/tenants/"+id.String(), nil)
				req = withTenant(req, uuid.New(), uuid.New())
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("GET /tenants", func(t *testing.T) {
		router, db := setupHandlerTest(t)

		// Create a tenant and use its ID in the context
		ten := createTenantInDB(t, db, "List Tenant")

		req := httptest.NewRequest(http.MethodGet, "/tenants", nil)
		req = withTenant(req, ten.Id(), uuid.New())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}
