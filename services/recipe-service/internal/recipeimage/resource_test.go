package recipeimage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore/objectstoretest"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

type httpFixture struct {
	fixture
	router   *mux.Router
	hiddenID uuid.UUID
}

func newHTTPFixture(t *testing.T, store objectstore.Store) httpFixture {
	f := newFixture(t)
	hidden := uuid.New()
	visible := func(_ context.Context, id uuid.UUID) error {
		if id == hidden {
			return ErrRecipeNotFound
		}
		return nil
	}
	l, _ := test.NewNullLogger()
	if store == nil {
		store = f.store
	}
	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := tenantctx.WithContext(r.Context(), tenantctx.New(f.tenantID, f.householdID, uuid.New()))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	InitializeRoutes(f.db, store, visible, NewObjectReaper(l, f.db, store))(l, server.GetServerInformation(), api)
	return httpFixture{fixture: f, router: router, hiddenID: hidden}
}

func multipartBody(t *testing.T, field string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(field, "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func (h httpFixture) do(method, path string, body io.Reader, contentType string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	return rec
}

func imagePath(id uuid.UUID) string { return "/api/v1/recipes/" + id.String() + "/image" }

func TestPutImageSuccessShape(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()
	body, ct := multipartBody(t, "file", testJPEG(t, 2000, 1000))
	rec := h.do(http.MethodPut, imagePath(rid), body, ct, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/vnd.api+json" {
		t.Fatalf("content-type = %q", got)
	}
	var doc struct {
		Data struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				Width    int    `json:"width"`
				Height   int    `json:"height"`
				ThumbURL string `json:"thumbUrl"`
				FullURL  string `json:"fullUrl"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	a := doc.Data.Attributes
	if doc.Data.Type != "recipe-images" || a.Width != 1600 || a.Height != 800 {
		t.Fatalf("doc = %+v", doc)
	}
	if !strings.Contains(a.FullURL, "size=full&v="+doc.Data.ID) || !strings.Contains(a.ThumbURL, "size=thumb&v="+doc.Data.ID) {
		t.Fatalf("urls = %q %q", a.FullURL, a.ThumbURL)
	}
}

func TestPutImageErrors(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()

	// 400: not multipart
	if rec := h.do(http.MethodPut, imagePath(rid), strings.NewReader("{}"), "application/json", nil); rec.Code != http.StatusBadRequest {
		t.Errorf("not multipart: %d", rec.Code)
	}
	// 400: no file part
	body, ct := multipartBody(t, "other", testJPEG(t, 64, 64))
	if rec := h.do(http.MethodPut, imagePath(rid), body, ct, nil); rec.Code != http.StatusBadRequest {
		t.Errorf("no file part: %d", rec.Code)
	}
	// 404: invisible recipe
	body, ct = multipartBody(t, "file", testJPEG(t, 64, 64))
	if rec := h.do(http.MethodPut, imagePath(h.hiddenID), body, ct, nil); rec.Code != http.StatusNotFound {
		t.Errorf("hidden: %d", rec.Code)
	}
	// 413: file part > 10 MiB
	body, ct = multipartBody(t, "file", make([]byte, maxFileBytes+1))
	if rec := h.do(http.MethodPut, imagePath(rid), body, ct, nil); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("file too large: %d", rec.Code)
	}
	// 413: body over the cap via a large non-file part before the file
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	junk, _ := mw.CreateFormField("junk")
	_, _ = junk.Write(make([]byte, maxBodyBytes))
	fw, _ := mw.CreateFormFile("file", "a.jpg")
	_, _ = fw.Write(testJPEG(t, 64, 64))
	_ = mw.Close()
	if rec := h.do(http.MethodPut, imagePath(rid), &buf, mw.FormDataContentType(), nil); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("body too large: %d", rec.Code)
	}
	// 415: not an image
	body, ct = multipartBody(t, "file", []byte("hello"))
	if rec := h.do(http.MethodPut, imagePath(rid), body, ct, nil); rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("not an image: %d", rec.Code)
	}
	// 422: too small
	body, ct = multipartBody(t, "file", testJPEG(t, 8, 8))
	if rec := h.do(http.MethodPut, imagePath(rid), body, ct, nil); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("too small: %d", rec.Code)
	}
}

func TestPutImageStorageDisabled(t *testing.T) {
	l, _ := test.NewNullLogger()
	h := newHTTPFixture(t, objectstore.New(l, objectstore.Config{}))
	body, ct := multipartBody(t, "file", testJPEG(t, 64, 64))
	if rec := h.do(http.MethodPut, imagePath(uuid.New()), body, ct, nil); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func uploadVia(t *testing.T, h httpFixture, rid uuid.UUID) string {
	t.Helper()
	body, ct := multipartBody(t, "file", testJPEG(t, 64, 64))
	rec := h.do(http.MethodPut, imagePath(rid), body, ct, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status %d", rec.Code)
	}
	var doc struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &doc)
	return doc.Data.ID
}

func TestGetImageHeadersAndBody(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()
	imageID := uploadVia(t, h, rid)

	rec := h.do(http.MethodGet, imagePath(rid)+"?size=thumb&v="+imageID, nil, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	hdr := rec.Header()
	if hdr.Get("Content-Type") != "image/jpeg" || hdr.Get("X-Content-Type-Options") != "nosniff" ||
		hdr.Get("Cache-Control") != "private, max-age=31536000, immutable" ||
		hdr.Get("ETag") != `"`+imageID+`-thumb"` {
		t.Fatalf("headers = %v", hdr)
	}
	if rec.Body.Len() == 0 || hdr.Get("Content-Length") == "" {
		t.Fatal("empty body or missing length")
	}
	if rec := h.do(http.MethodGet, imagePath(rid), nil, "", nil); rec.Header().Get("ETag") != `"`+imageID+`-full"` {
		t.Fatalf("default size not full: %v", rec.Header())
	}
}

func TestGetImageNotModifiedWithoutStorage(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()
	imageID := uploadVia(t, h, rid)
	h.store.GetErr = objectstore.ErrUnavailable
	before := h.store.GetCalls

	rec := h.do(http.MethodGet, imagePath(rid)+"?size=full", nil, "", map[string]string{"If-None-Match": `"` + imageID + `-full"`})
	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d", rec.Code)
	}
	if h.store.GetCalls != before {
		t.Fatal("304 path must not touch storage")
	}
}

func TestGetImageNotModifiedWeakList(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()
	imageID := uploadVia(t, h, rid)
	rec := h.do(http.MethodGet, imagePath(rid), nil, "", map[string]string{"If-None-Match": `"zzz", W/"` + imageID + `-full"`})
	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d", rec.Code)
	}
	rec = h.do(http.MethodGet, imagePath(rid), nil, "", map[string]string{"If-None-Match": `"` + imageID + `-thumb"`})
	if rec.Code != http.StatusOK {
		t.Fatalf("mismatched etag status = %d", rec.Code)
	}
}

func TestGetImageErrors(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()
	if rec := h.do(http.MethodGet, imagePath(rid)+"?size=huge", nil, "", nil); rec.Code != http.StatusBadRequest {
		t.Errorf("bad size: %d", rec.Code)
	}
	if rec := h.do(http.MethodGet, imagePath(rid), nil, "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("no image: %d", rec.Code)
	}
	if rec := h.do(http.MethodGet, imagePath(h.hiddenID), nil, "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("hidden: %d", rec.Code)
	}
	uploadVia(t, h, rid)
	h.store.GetErr = objectstore.ErrUnavailable
	if rec := h.do(http.MethodGet, imagePath(rid), nil, "", nil); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("storage down: %d", rec.Code)
	}
	h.store.GetErr = nil
	for _, k := range h.store.Keys() {
		_ = h.store.Delete(context.Background(), k)
	}
	if rec := h.do(http.MethodGet, imagePath(rid), nil, "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("missing object: %d", rec.Code)
	}
}

func TestDeleteImage(t *testing.T) {
	h := newHTTPFixture(t, nil)
	rid := uuid.New()
	uploadVia(t, h, rid)
	if rec := h.do(http.MethodDelete, imagePath(rid), nil, "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", rec.Code)
	}
	if rec := h.do(http.MethodDelete, imagePath(rid), nil, "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("idempotent delete: %d", rec.Code)
	}
	if rec := h.do(http.MethodDelete, imagePath(h.hiddenID), nil, "", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("hidden delete: %d", rec.Code)
	}
}

// Compile-time check that the fake satisfies the interface.
var _ objectstore.Store = (*objectstoretest.Store)(nil)
