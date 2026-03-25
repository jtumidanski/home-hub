package preference

import (
	"bytes"
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
	t.Run("GET /preferences", func(t *testing.T) {
		tests := []struct {
			name       string
			wantStatus int
		}{
			{"auto-creates default preference", http.StatusOK},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				router, _, tenantID, userID := setupHandlerTest(t)
				req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
				req = withTenant(req, tenantID, userID)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("PATCH /preferences/{id}", func(t *testing.T) {
		tests := []struct {
			name       string
			body       func(id string) string
			wantStatus int
		}{
			{
				name: "update theme",
				body: func(id string) string {
					return `{"data":{"type":"preferences","id":"` + id + `","attributes":{"theme":"dark"}}}`
				},
				wantStatus: http.StatusOK,
			},
			{
				name: "set active household",
				body: func(id string) string {
					return `{"data":{"type":"preferences","id":"` + id + `","attributes":{"active_household_id":"` + uuid.New().String() + `"}}}`
				},
				wantStatus: http.StatusOK,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				router, db, tenantID, userID := setupHandlerTest(t)

				// Create preference first
				l, _ := test.NewNullLogger()
				ctx := tenantctx.WithContext(
					httptest.NewRequest(http.MethodGet, "/", nil).Context(),
					tenantctx.New(tenantID, uuid.Nil, userID),
				)
				p := NewProcessor(l, ctx, db)
				pref, _ := p.FindOrCreate(tenantID, userID)

				body := tt.body(pref.Id().String())
				req := httptest.NewRequest(http.MethodPatch, "/preferences/"+pref.Id().String(), bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/vnd.api+json")
				req = withTenant(req, tenantID, userID)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
				}
			})
		}
	})
}
