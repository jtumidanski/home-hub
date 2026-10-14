package recipeimage

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage/imaging"
	"github.com/jtumidanski/home-hub/shared/go/server"
)

const (
	maxFileBytes   = 10 << 20
	maxBodyBytes   = maxFileBytes + 512<<10
	uploadDeadline = 60 * time.Second
)

// RecipeVisibility reports whether the recipe exists, is not soft-deleted
// and belongs to the caller's household. It is injected from the recipe
// package so recipeimage never imports recipe.
type RecipeVisibility func(ctx context.Context, recipeID uuid.UUID) error

func InitializeRoutes(db *gorm.DB, store objectstore.Store, visible RecipeVisibility, reaper *ObjectReaper) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/recipes/{id}/image", rh("UploadRecipeImage", uploadHandler(db, store, visible, reaper))).Methods(http.MethodPut)
		api.HandleFunc("/recipes/{id}/image", rh("GetRecipeImage", getHandler(db, store, visible))).Methods(http.MethodGet)
		api.HandleFunc("/recipes/{id}/image", rh("DeleteRecipeImage", deleteHandler(db, store, visible, reaper))).Methods(http.MethodDelete)
	}
}

// checkVisible writes 404/500 and returns false when the recipe is not
// visible to the caller.
func checkVisible(l logrus.FieldLogger, visible RecipeVisibility, w http.ResponseWriter, r *http.Request, id uuid.UUID) bool {
	err := visible(r.Context(), id)
	if err == nil {
		return true
	}
	if errors.Is(err, ErrRecipeNotFound) {
		server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe not found")
		return false
	}
	l.WithError(err).Error("Failed to check recipe visibility")
	server.WriteError(w, http.StatusInternalServerError, "Error", "")
	return false
}

func uploadHandler(db *gorm.DB, store objectstore.Store, visible RecipeVisibility, reaper *ObjectReaper) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if !checkVisible(d.Logger(), visible, w, r, id) {
					return
				}
				// Proxies buffer bodies, but guard against a slow direct
				// client exceeding the shared 15 s server timeouts. Errors
				// mean the writer does not support deadlines (tests).
				rc := http.NewResponseController(w)
				_ = rc.SetReadDeadline(time.Now().Add(uploadDeadline))
				_ = rc.SetWriteDeadline(time.Now().Add(uploadDeadline))

				data, status, detail := readFilePart(w, r)
				if status != 0 {
					if status == http.StatusRequestEntityTooLarge {
						recordUpload("rejected_size")
					}
					server.WriteError(w, status, http.StatusText(status), detail)
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db).WithStore(store)
				res, err := proc.Upload(id, data)
				if err != nil {
					writeUploadError(d.Logger(), w, err)
					return
				}
				if len(res.OutboxIDs) > 0 {
					go reaper.DrainKeys(context.Background(), res.OutboxIDs)
				}
				if res.Replaced {
					recordUpload("replaced")
				} else {
					recordUpload("created")
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(Transform(res.Model))
			}
		})
	}
}

// readFilePart streams the multipart body and returns the first part named
// "file". It never spills to temp files. A non-zero status means the caller
// must write that error.
func readFilePart(w http.ResponseWriter, r *http.Request) ([]byte, int, string) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" || params["boundary"] == "" {
		return nil, http.StatusBadRequest, "Request must be multipart/form-data with a file part"
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	mr := multipart.NewReader(r.Body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, http.StatusBadRequest, "Missing file part"
		}
		if err != nil {
			status, detail := statusForReadErr(err)
			return nil, status, detail
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, maxFileBytes+1))
		if err != nil {
			status, detail := statusForReadErr(err)
			return nil, status, detail
		}
		if len(data) > maxFileBytes {
			return nil, http.StatusRequestEntityTooLarge, "Image must be 10 MB or smaller"
		}
		return data, 0, ""
	}
}

// statusForReadErr maps a body read error to the HTTP status and detail.
func statusForReadErr(err error) (int, string) {
	if isMaxBytes(err) {
		return http.StatusRequestEntityTooLarge, "Image must be 10 MB or smaller"
	}
	return http.StatusBadRequest, "Malformed multipart body"
}

func isMaxBytes(err error) bool {
	var mbe *http.MaxBytesError
	return errors.As(err, &mbe)
}

func writeUploadError(l logrus.FieldLogger, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, imaging.ErrUnsupportedFormat):
		recordUpload("rejected_type")
		server.WriteError(w, http.StatusUnsupportedMediaType, "Unsupported Media Type", "Use a JPEG, PNG or WebP image")
	case errors.Is(err, imaging.ErrTooLarge), errors.Is(err, imaging.ErrTooSmall):
		recordUpload("rejected_dimensions")
		server.WriteError(w, http.StatusUnprocessableEntity, "Invalid Image Dimensions", "Image must be at most 40 megapixels and at least 16 px on each edge")
	case errors.Is(err, ErrStorage):
		recordUpload("storage_error")
		server.WriteError(w, http.StatusServiceUnavailable, "Service Unavailable", "Image storage is unavailable")
	case errors.Is(err, context.Canceled):
		// Client gone while queued for processing; nothing to write.
		recordUpload("db_error")
	default:
		recordUpload("db_error")
		l.WithError(err).Error("Failed to store recipe image")
		server.WriteError(w, http.StatusInternalServerError, "Upload Failed", "")
	}
}

func getHandler(db *gorm.DB, store objectstore.Store, visible RecipeVisibility) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				size, ok := ParseSize(r.URL.Query().Get("size"))
				if !ok {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "size must be thumb or full")
					return
				}
				if !checkVisible(d.Logger(), visible, w, r, id) {
					return
				}
				proc := NewProcessor(d.Logger(), r.Context(), db).WithStore(store)
				m, err := proc.GetByRecipeID(id)
				if errors.Is(err, ErrImageNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe has no image")
					return
				}
				if err != nil {
					d.Logger().WithError(err).Error("Failed to load recipe image")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				etag := `"` + m.Id().String() + "-" + string(size) + `"`
				if etagMatches(r.Header.Get("If-None-Match"), etag) {
					w.Header().Set("ETag", etag)
					w.WriteHeader(http.StatusNotModified)
					return
				}

				obj, err := proc.Open(m, size)
				if errors.Is(err, ErrObjectMissing) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe image not found")
					return
				}
				if err != nil {
					server.WriteError(w, http.StatusServiceUnavailable, "Service Unavailable", "Image storage is unavailable")
					return
				}
				defer obj.Body.Close()

				h := w.Header()
				h.Set("Content-Type", "image/jpeg")
				h.Set("Content-Length", strconv.FormatInt(obj.Size, 10))
				h.Set("X-Content-Type-Options", "nosniff")
				h.Set("Cache-Control", "private, max-age=31536000, immutable")
				h.Set("ETag", etag)
				w.WriteHeader(http.StatusOK)
				if _, err := io.Copy(w, obj.Body); err != nil {
					d.Logger().WithError(err).Warn("Failed to stream recipe image")
				}
			}
		})
	}
}

// etagMatches implements the weak comparison of If-None-Match.
func etagMatches(header, etag string) bool {
	if header == "" {
		return false
	}
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		if strings.TrimPrefix(candidate, "W/") == etag {
			return true
		}
	}
	return false
}

func deleteHandler(db *gorm.DB, store objectstore.Store, visible RecipeVisibility, reaper *ObjectReaper) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if !checkVisible(d.Logger(), visible, w, r, id) {
					return
				}
				proc := NewProcessor(d.Logger(), r.Context(), db).WithStore(store)
				ids, err := proc.Delete(id)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to delete recipe image")
					server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
					return
				}
				if len(ids) > 0 {
					go reaper.DrainKeys(context.Background(), ids)
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
