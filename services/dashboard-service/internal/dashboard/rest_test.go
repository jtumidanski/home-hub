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
