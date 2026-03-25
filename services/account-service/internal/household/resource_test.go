package household

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
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
	db.AutoMigrate(&membership.Entity{})

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
	t.Run("POST /households", func(t *testing.T) {
		router, _, tenantID, userID := setupHandlerTest(t)

		id := uuid.New()
		body := `{"data":{"type":"households","id":"` + id.String() + `","attributes":{"name":"My Home","timezone":"UTC","units":"metric"}}}`
		req := httptest.NewRequest(http.MethodPost, "/households", bytes.NewBufferString(body))
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
		if data["type"] != "households" {
			t.Errorf("expected type households, got %v", data["type"])
		}
	})

	t.Run("GET /households", func(t *testing.T) {
		router, db, tenantID, userID := setupHandlerTest(t)

		l, _ := test.NewNullLogger()
		ctx := tenantctx.WithContext(
			httptest.NewRequest(http.MethodGet, "/", nil).Context(),
			tenantctx.New(tenantID, uuid.Nil, userID),
		)
		p := NewProcessor(l, ctx, db)
		p.Create(tenantID, "Home 1", "UTC", "metric")

		req := httptest.NewRequest(http.MethodGet, "/households", nil)
		req = withTenant(req, tenantID, userID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("GET /households/{id}", func(t *testing.T) {
		tests := []struct {
			name       string
			setup      func(db *gorm.DB, tenantID, userID uuid.UUID) string
			wantStatus int
		}{
			{
				name: "existing household",
				setup: func(db *gorm.DB, tenantID, userID uuid.UUID) string {
					l, _ := test.NewNullLogger()
					ctx := tenantctx.WithContext(
						httptest.NewRequest(http.MethodGet, "/", nil).Context(),
						tenantctx.New(tenantID, uuid.Nil, userID),
					)
					p := NewProcessor(l, ctx, db)
					m, _ := p.Create(tenantID, "Test", "UTC", "metric")
					return m.Id().String()
				},
				wantStatus: http.StatusOK,
			},
			{
				name: "non-existent household",
				setup: func(_ *gorm.DB, _, _ uuid.UUID) string {
					return uuid.New().String()
				},
				wantStatus: http.StatusNotFound,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				router, db, tenantID, userID := setupHandlerTest(t)
				id := tt.setup(db, tenantID, userID)
				req := httptest.NewRequest(http.MethodGet, "/households/"+id, nil)
				req = withTenant(req, tenantID, userID)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("PATCH /households/{id}", func(t *testing.T) {
		router, db, tenantID, userID := setupHandlerTest(t)

		l, _ := test.NewNullLogger()
		ctx := tenantctx.WithContext(
			httptest.NewRequest(http.MethodGet, "/", nil).Context(),
			tenantctx.New(tenantID, uuid.Nil, userID),
		)
		p := NewProcessor(l, ctx, db)
		m, _ := p.Create(tenantID, "Old Name", "UTC", "metric")

		body := `{"data":{"type":"households","id":"` + m.Id().String() + `","attributes":{"name":"New Name","timezone":"America/Chicago","units":"imperial"}}}`
		req := httptest.NewRequest(http.MethodPatch, "/households/"+m.Id().String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/vnd.api+json")
		req = withTenant(req, tenantID, userID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}
