package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestServer wires a dashboard router with a test-only middleware that
// injects a fake tenant into the context (replaces real JWT auth for tests).
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

func doRequest(t *testing.T, h http.Handler, method, url string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		buf, err := json.Marshal(body)
		require.NoError(t, err)
		r = httptest.NewRequest(method, url, bytes.NewReader(buf))
		r.Header.Set("Content-Type", "application/vnd.api+json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// jsonapiDoc is a minimal shape for parsing single-resource responses.
type jsonapiDoc struct {
	Data struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Attributes RestModel `json:"attributes"`
	} `json:"data"`
}

// jsonapiSlice is for list responses.
type jsonapiSlice struct {
	Data []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Attributes RestModel `json:"attributes"`
	} `json:"data"`
}

// jsonapiErrors captures a JSON:API error envelope for assertion.
type jsonapiErrorDoc struct {
	Errors []struct {
		Status string `json:"status"`
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Source *struct {
			Pointer string `json:"pointer"`
		} `json:"source,omitempty"`
	} `json:"errors"`
}

func TestListHandlerScopesVisibility(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	callerUID := uuid.New()
	otherUID := uuid.New()

	proc := newTestProcessor(t, db)
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)
	_, err := proc.Create(tid, hid, callerUID, CreateAttrs{Name: "Shared", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)
	_, err = proc.Create(tid, hid, callerUID, CreateAttrs{Name: "Mine", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)
	_, err = proc.Create(tid, hid, otherUID, CreateAttrs{Name: "Theirs", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, callerUID))
	rec := doRequest(t, h, http.MethodGet, "/api/v1/dashboards", nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var doc jsonapiSlice
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Len(t, doc.Data, 2)

	names := map[string]bool{}
	for _, d := range doc.Data {
		names[d.Attributes.Name] = true
	}
	require.True(t, names["Shared"])
	require.True(t, names["Mine"])
	require.False(t, names["Theirs"])
}

// jsonapiBody wraps an attribute payload into the {"data":{"type":"dashboards","attributes":...}}
// shape that api2go's Unmarshal expects.
func jsonapiBody(resourceType string, attrs any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"type":       resourceType,
			"attributes": attrs,
		},
	}
}

func TestCreateHouseholdReturns201(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	body := jsonapiBody("dashboards", map[string]any{
		"name":   "Home",
		"scope":  "household",
		"layout": map[string]any{"version": 1, "widgets": []any{}},
	})
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var doc jsonapiDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "Home", doc.Data.Attributes.Name)
	require.Equal(t, "household", doc.Data.Attributes.Scope)
}

func TestCreateUserScopeReturnsUserScope(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	body := jsonapiBody("dashboards", map[string]any{
		"name":   "Mine",
		"scope":  "user",
		"layout": map[string]any{"version": 1, "widgets": []any{}},
	})
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var doc jsonapiDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "user", doc.Data.Attributes.Scope)
}

func TestCreateWithInvalidLayoutReturns422(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	// Widget missing an id → layout.widget_bad_id.
	body := jsonapiBody("dashboards", map[string]any{
		"name":  "Bad",
		"scope": "household",
		"layout": map[string]any{
			"version": 1,
			"widgets": []any{
				map[string]any{
					"type":   "weather",
					"x":      0,
					"y":      0,
					"w":      1,
					"h":      1,
					"config": map[string]any{},
				},
			},
		},
	})
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards", body)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	var doc jsonapiErrorDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Len(t, doc.Errors, 1)
	require.Equal(t, "layout.widget_bad_id", doc.Errors[0].Code)
	require.NotNil(t, doc.Errors[0].Source)
	require.Contains(t, doc.Errors[0].Source.Pointer, "/data/attributes/layout/widgets/0/id")
}

func TestUpdateNonOwnerReturns403(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()

	proc := newTestProcessor(t, db)
	m, err := proc.Create(tid, hid, userA, CreateAttrs{
		Name:   "A's Board",
		Scope:  "user",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	require.NoError(t, err)

	// userB attempts to rename userA's user-scoped dashboard.
	h := newTestServer(t, db, tenantctx.New(tid, hid, userB))
	body := jsonapiBody("dashboards", map[string]any{"name": "Hacked"})
	rec := doRequest(t, h, http.MethodPatch, "/api/v1/dashboards/"+m.Id().String(), body)
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestReorderSingleScopeReturns200(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	proc := newTestProcessor(t, db)
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)
	a, err := proc.Create(tid, hid, uid, CreateAttrs{Name: "A", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)
	b, err := proc.Create(tid, hid, uid, CreateAttrs{Name: "B", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))
	body := ReorderRequest{Entries: []ReorderEntry{
		{ID: a.Id().String(), SortOrder: 5},
		{ID: b.Id().String(), SortOrder: 3},
	}}
	rec := doRequest(t, h, http.MethodPatch, "/api/v1/dashboards/order", body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var doc jsonapiSlice
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Len(t, doc.Data, 2)
}

func TestReorderMixedScopeReturns400(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	proc := newTestProcessor(t, db)
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)
	hh, err := proc.Create(tid, hid, uid, CreateAttrs{Name: "H", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)
	uu, err := proc.Create(tid, hid, uid, CreateAttrs{Name: "U", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))
	body := ReorderRequest{Entries: []ReorderEntry{
		{ID: hh.Id().String(), SortOrder: 1},
		{ID: uu.Id().String(), SortOrder: 2},
	}}
	rec := doRequest(t, h, http.MethodPatch, "/api/v1/dashboards/order", body)
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())

	var doc jsonapiErrorDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Len(t, doc.Errors, 1)
	require.Equal(t, "dashboard.mixed_scope", doc.Errors[0].Code)
}

func TestPromoteSucceeds(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	proc := newTestProcessor(t, db)
	m, err := proc.Create(tid, hid, uid, CreateAttrs{
		Name:   "Mine",
		Scope:  "user",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/"+m.Id().String()+"/promote", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var doc jsonapiDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "household", doc.Data.Attributes.Scope)
}

func TestPromoteAlreadyHouseholdReturns409(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	proc := newTestProcessor(t, db)
	m, err := proc.Create(tid, hid, uid, CreateAttrs{
		Name:   "Shared",
		Scope:  "household",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/"+m.Id().String()+"/promote", nil)
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())

	var doc jsonapiErrorDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "dashboard.already_household", doc.Errors[0].Code)
}

func TestCopyToMineReturns201WithMaxSortOrder(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	proc := newTestProcessor(t, db)
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	// Source: a household dashboard to copy.
	src, err := proc.Create(tid, hid, uid, CreateAttrs{Name: "Home", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)

	// An existing user-scoped row (sort_order=0 by default auto-max+1 of empty=0)
	// establishes a baseline max in the user scope so the copy gets max+1.
	existing, err := proc.Create(tid, hid, uid, CreateAttrs{Name: "Pre", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/"+src.Id().String()+"/copy-to-mine", nil)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var doc jsonapiDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "user", doc.Data.Attributes.Scope)
	require.Equal(t, existing.SortOrder()+1, doc.Data.Attributes.SortOrder)
}

func TestSeedFirstCallReturns201(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	body := jsonapiBody("dashboards", map[string]any{
		"name":   "Home",
		"layout": map[string]any{"version": 1, "widgets": []any{}},
	})
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/seed", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var doc jsonapiDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Equal(t, "Home", doc.Data.Attributes.Name)
	require.Equal(t, "household", doc.Data.Attributes.Scope)
}

func TestSeedSecondCallReturns200Slice(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	body := jsonapiBody("dashboards", map[string]any{
		"name":   "Home",
		"layout": map[string]any{"version": 1, "widgets": []any{}},
	})
	rec1 := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/seed", body)
	require.Equal(t, http.StatusCreated, rec1.Code, rec1.Body.String())

	rec2 := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/seed", body)
	require.Equal(t, http.StatusOK, rec2.Code, rec2.Body.String())

	var doc jsonapiSlice
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &doc))
	require.GreaterOrEqual(t, len(doc.Data), 1)
}

func TestSeedHandlerWithKioskKey(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	body := jsonapiBody("dashboards", map[string]any{
		"name":   "Kiosk",
		"key":    "kiosk",
		"layout": map[string]any{"version": 1, "widgets": []any{}},
	})
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/seed", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	// Verify the row was persisted with seed_key = "kiosk".
	var ent Entity
	require.NoError(t, db.Where("tenant_id = ? AND household_id = ? AND seed_key = ?", tid, hid, "kiosk").First(&ent).Error)
	require.NotNil(t, ent.SeedKey)
	require.Equal(t, "kiosk", *ent.SeedKey)
}

func TestSeedHandlerRejectsMalformedKey(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	h := newTestServer(t, db, tenantctx.New(tid, hid, uid))

	body := jsonapiBody("dashboards", map[string]any{
		"name":   "Kiosk",
		"key":    "Has Spaces",
		"layout": map[string]any{"version": 1, "widgets": []any{}},
	})
	rec := doRequest(t, h, http.MethodPost, "/api/v1/dashboards/seed", body)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	var doc jsonapiErrorDoc
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	require.Len(t, doc.Errors, 1)
	require.NotNil(t, doc.Errors[0].Source)
	require.Equal(t, "/data/attributes/key", doc.Errors[0].Source.Pointer)
}

func TestDeleteHouseholdAnyMemberReturns204(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	creator, other := uuid.New(), uuid.New()

	proc := newTestProcessor(t, db)
	m, err := proc.Create(tid, hid, creator, CreateAttrs{
		Name:   "Family",
		Scope:  "household",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	require.NoError(t, err)

	h := newTestServer(t, db, tenantctx.New(tid, hid, other))
	rec := doRequest(t, h, http.MethodDelete, "/api/v1/dashboards/"+m.Id().String(), nil)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
}
