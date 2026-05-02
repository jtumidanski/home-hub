package householdpreference

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newTestServer(t *testing.T, db *gorm.DB, tenant tenantctx.Tenant) http.Handler {
	t.Helper()
	l, _ := test.NewNullLogger()
	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := tenantctx.WithContext(r.Context(), tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	si := server.GetServerInformation()
	InitializeRoutes(db)(l, si, api)
	return router
}

// jsonapiSliceDoc parses a list-of-one response.
type jsonapiSliceDoc struct {
	Data []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Attributes RestModel `json:"attributes"`
	} `json:"data"`
}

type jsonapiSingleDoc struct {
	Data struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Attributes RestModel `json:"attributes"`
	} `json:"data"`
}

func TestGetHouseholdPreferencesAutoCreates(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/household-preferences", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d: %s", w.Code, w.Body.String())
	}
	var doc jsonapiSliceDoc
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, w.Body.String())
	}
	if len(doc.Data) != 1 {
		t.Fatalf("expected 1 row got %d", len(doc.Data))
	}
	if doc.Data[0].Type != "householdPreferences" {
		t.Fatalf("type: got %q", doc.Data[0].Type)
	}
	if doc.Data[0].Attributes.DefaultDashboardId != nil {
		t.Fatalf("expected null defaultDashboardId, got %v", doc.Data[0].Attributes.DefaultDashboardId)
	}
}

func TestGetHouseholdPreferencesIncludesKioskFlag(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/household-preferences", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"kioskDashboardSeeded":false`)) {
		t.Fatalf("expected kioskDashboardSeeded:false in body, got: %s", w.Body.String())
	}
}

func TestPatchSetsDefaultDashboardId(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	// GET to auto-create.
	getResp := doGet(t, h)
	rowID := getResp.Data[0].ID

	dashID := uuid.New()
	body := map[string]any{
		"data": map[string]any{
			"type": "householdPreferences",
			"id":   rowID,
			"attributes": map[string]any{
				"defaultDashboardId": dashID.String(),
			},
		},
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/household-preferences/"+rowID, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d: %s", w.Code, w.Body.String())
	}
	var single jsonapiSingleDoc
	if err := json.Unmarshal(w.Body.Bytes(), &single); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, w.Body.String())
	}
	if single.Data.Attributes.DefaultDashboardId == nil || *single.Data.Attributes.DefaultDashboardId != dashID {
		t.Fatalf("expected default dashboard %s, got %v", dashID, single.Data.Attributes.DefaultDashboardId)
	}
}

func TestPatchClearsDefaultDashboardId(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	getResp := doGet(t, h)
	rowID := getResp.Data[0].ID

	// First, set a value.
	dashID := uuid.New()
	body := map[string]any{
		"data": map[string]any{
			"type": "householdPreferences",
			"id":   rowID,
			"attributes": map[string]any{
				"defaultDashboardId": dashID.String(),
			},
		},
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/household-preferences/"+rowID, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("set status: got %d: %s", w.Code, w.Body.String())
	}

	// Now clear via explicit null.
	clearBody := map[string]any{
		"data": map[string]any{
			"type": "householdPreferences",
			"id":   rowID,
			"attributes": map[string]any{
				"defaultDashboardId": nil,
			},
		},
	}
	buf2, _ := json.Marshal(clearBody)
	req2 := httptest.NewRequest(http.MethodPatch, "/api/v1/household-preferences/"+rowID, bytes.NewReader(buf2))
	req2.Header.Set("Content-Type", "application/vnd.api+json")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("clear status: got %d: %s", w2.Code, w2.Body.String())
	}

	// Verify via GET.
	getResp2 := doGet(t, h)
	if getResp2.Data[0].Attributes.DefaultDashboardId != nil {
		t.Fatalf("expected cleared defaultDashboardId, got %v", getResp2.Data[0].Attributes.DefaultDashboardId)
	}
}

func TestMarkKioskSeededFlipsFlag(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	// Auto-create the row via GET.
	getResp := doGet(t, h)
	if len(getResp.Data) != 1 {
		t.Fatalf("expected 1 row got %d", len(getResp.Data))
	}
	rowID := getResp.Data[0].ID
	if getResp.Data[0].Attributes.KioskDashboardSeeded {
		t.Fatalf("expected initial kioskDashboardSeeded to be false")
	}

	// PATCH /api/v1/household-preferences/{id}/kiosk-seeded with {"value": true}.
	patchReq := httptest.NewRequest(http.MethodPatch,
		"/api/v1/household-preferences/"+rowID+"/kiosk-seeded",
		bytes.NewReader([]byte(`{"value":true}`)))
	patchReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, patchReq)
	if w.Code != http.StatusOK {
		t.Fatalf("patch status: got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"kioskDashboardSeeded":true`)) {
		t.Fatalf("expected kioskDashboardSeeded:true in body, got: %s", w.Body.String())
	}

	// Idempotency: second PATCH stays 200 + true.
	patchReq2 := httptest.NewRequest(http.MethodPatch,
		"/api/v1/household-preferences/"+rowID+"/kiosk-seeded",
		bytes.NewReader([]byte(`{"value":true}`)))
	patchReq2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, patchReq2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second patch status: got %d: %s", w2.Code, w2.Body.String())
	}
	if !bytes.Contains(w2.Body.Bytes(), []byte(`"kioskDashboardSeeded":true`)) {
		t.Fatalf("expected kioskDashboardSeeded:true on second patch, got: %s", w2.Body.String())
	}
}

func TestMarkKioskSeededRejectsFalse(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	getResp := doGet(t, h)
	rowID := getResp.Data[0].ID

	patchReq := httptest.NewRequest(http.MethodPatch,
		"/api/v1/household-preferences/"+rowID+"/kiosk-seeded",
		bytes.NewReader([]byte(`{"value":false}`)))
	patchReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, patchReq)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for value:false, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMarkKioskSeededReturnsNotFoundForUnknownID(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, hid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	unknownID := uuid.New().String()
	patchReq := httptest.NewRequest(http.MethodPatch,
		"/api/v1/household-preferences/"+unknownID+"/kiosk-seeded",
		bytes.NewReader([]byte(`{"value":true}`)))
	patchReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, patchReq)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown id, got %d: %s", w.Code, w.Body.String())
	}
}

func doGet(t *testing.T, h http.Handler) jsonapiSliceDoc {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/household-preferences", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get status: got %d: %s", w.Code, w.Body.String())
	}
	var doc jsonapiSliceDoc
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("get unmarshal: %v body=%s", err, w.Body.String())
	}
	return doc
}
