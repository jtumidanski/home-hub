# Recipe Images Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let each recipe carry one optional photo — uploaded from the recipe form, processed and stored by recipe-service in MinIO, shown as a hero on the detail page and a thumbnail on list cards.

**Architecture:** recipe-service gains three packages: `objectstore` (thin minio-go wrapper behind an interface), `recipeimage/imaging` (pure decode → limit → resize → orient → JPEG), and the `recipeimage` domain (entity, outbox, processor, handlers, drainer). Object deletions go through a transactional outbox table so replace, delete and retention purge all clean up objects after commit. The `recipe` package enriches list/detail responses with a nullable `image` attribute through one batched query. The frontend extends the existing `api.upload` with a `PUT` method, adds a staged-draft hook, a form field, and a `RecipeImage` renderer.

**Tech Stack:** Go 1.26 (GORM, gorilla/mux, api2go, logrus, prometheus `promauto`), `github.com/minio/minio-go/v7`, `golang.org/x/image` (`webp`, `draw`); React + TypeScript (TanStack Query, react-hook-form, Vitest + RTL, Tailwind, shadcn/base-ui Dialog, lucide); Docker Compose MinIO.

**Spec:** `docs/tasks/task-056-recipe-images/design.md` (PRD: `docs/tasks/task-056-recipe-images/prd.md`, mockup: `docs/tasks/task-056-recipe-images/mockup.html`). Read `context.md` in the same folder first.

## Global Constraints

- No changes under `shared/`. Everything lives in `services/recipe-service`, `frontend`, `deploy`, `.env.example`, and recipe-service docs.
- `internal/objectstore` is the only package that imports `github.com/minio/minio-go/v7`.
- Accepted input formats: JPEG, PNG, WebP — detected by decoding bytes only (never `Content-Type` or filename).
- File part limit: exactly 10 MiB (`10 << 20`). Request body cap: 10 MiB + 512 KiB. Both → 413.
- Pixel limit: `w*h > 40_000_000` → 422, checked with `image.DecodeConfig` before full decode. `min(w,h) < 16` → 422.
- Renditions: `full` longest edge ≤ 1600, `thumb` longest edge ≤ 480, never upscaled, JPEG quality 82, transparent pixels composited onto white, `thumb` scaled from `full`.
- Object keys: `{tenantId}/{householdId}/recipes/{recipeId}/{imageId}/full.jpg` and `…/thumb.jpg`, built only from `uuid.UUID` values.
- Image URLs: `/api/v1/recipes/{recipeId}/image?size=thumb|full&v={imageId}`, built only by `recipeimage.URLs`.
- GET image headers: `Content-Type: image/jpeg`, `X-Content-Type-Options: nosniff`, `Cache-Control: private, max-age=31536000, immutable`, `ETag: "<imageId>-<size>"`.
- Missing tenant/household or invisible recipe → 404, never 403.
- Config env: `S3_ENDPOINT` (default `""` = disabled), `S3_BUCKET` (`home-hub-recipes`), `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL` (`false`), `S3_REGION` (`us-east-1`).
- Metrics: `recipe_image_uploads_total{result}` (`created|replaced|rejected_size|rejected_type|rejected_dimensions|storage_error|db_error`), `recipe_image_processing_seconds`, `recipe_image_storage_errors_total{op}` (`put|get|delete|ping`).
- Frontend client validation message (verbatim): `Use a JPEG, PNG or WebP image under 10 MB.`
- Partial-failure toast (verbatim): `Recipe saved, but the photo failed to upload — try again.`
- Save button in-flight label (verbatim): `Uploading photo…`
- No `react-dropzone`, no new frontend dependencies.
- Never add `proxy_request_buffering off` to any nginx config.
- Implementers run module-local `go build ./... && go test ./...` (backend) or `npx vitest run <files>` + `npx tsc -b` (frontend). Repo-wide `tools/verify.sh` belongs to the task-verifier.
- Docs under `docs/` use repo-relative paths only.

## Review Focus

1. **Recipe shared across households in one tenant.** Existing recipe reads (`getByID`, `getAll`) filter only by tenant, but image reads also filter by household (design §6.1). A user in household B viewing household A's recipe must get `image: null` in list/detail and 404 from `GET …/image` — no 500, no error log spam. Pinned by the `TestImageEnrichmentOtherHousehold` test in Task 7.
2. **Portrait phone photo larger than the limit.** A 2000×1000 JPEG with EXIF orientation 6 must come out as an 800×1600 `full` and a 240×480 `thumb`. Pinned by `TestProcessOrientedDownscale` in Task 3.
3. **Paletted or grayscale PNG with transparency** (screenshots, exported graphics). It must decode, and its transparent pixels must turn white. Pinned by `TestProcessPalettedPNGWithTransparency` in Task 3.
4. **Browser revalidation with a weak or list-form `If-None-Match`** (for example `W/"<id>-full", "x"`). This must still return 304 without touching storage. Pinned by `TestGetImageNotModifiedWeakList` in Task 6.
5. **Android picker files with an empty MIME type** (`file.type === ""`). These must be rejected client-side with the standard message, not sent to the server. Pinned by the `rejects empty mime type` case in Task 10.

---

## File Structure

### Backend — `services/recipe-service`

| File | Responsibility |
|---|---|
| `internal/objectstore/objectstore.go` | `Store` interface, `Object`, `Config`, sentinel errors, `New` (picks minio or disabled) |
| `internal/objectstore/minio.go` | minio-go implementation (only minio-go importer) |
| `internal/objectstore/disabled.go` | Store that always returns `ErrUnavailable` |
| `internal/objectstore/objectstore_test.go` | `New` with empty endpoint → disabled; disabled returns `ErrUnavailable` |
| `internal/objectstore/objectstoretest/store.go` | In-memory fake with failure injection |
| `internal/objectstore/objectstoretest/store_test.go` | Fake contract tests |
| `internal/config/config.go` | add `ObjectStore objectstore.Config` |
| `internal/recipeimage/imaging/orientation.go` | `jpegOrientation` EXIF reader |
| `internal/recipeimage/imaging/orientation_test.go` | table + fuzz tests |
| `internal/recipeimage/imaging/imaging.go` | `Process`, `Result`, `Rendition`, errors |
| `internal/recipeimage/imaging/transform.go` | `fit`, `scaleOntoWhite`, `orient` |
| `internal/recipeimage/imaging/imaging_test.go` | table, fuzz, benchmark |
| `internal/recipeimage/imaging/testhelpers_test.go` | JPEG/PNG/EXIF fixture builders |
| `internal/recipeimage/entity.go` | `Entity`, `ObjectDeletionEntity`, `Migration`, `Make`, `ToEntity` |
| `internal/recipeimage/model.go` | immutable `Model`, `Size`, `ParseSize`, `Key` |
| `internal/recipeimage/builder.go` | `Builder` |
| `internal/recipeimage/keys.go` | `objectKeys`, `URLs` |
| `internal/recipeimage/provider.go` | read queries |
| `internal/recipeimage/administrator.go` | `deleteByRecipeIDReturning`, `createImage`, `enqueueObjectDeletions`, `PurgeByRecipeIDsTx` |
| `internal/recipeimage/errors.go` | sentinel errors, `isUniqueViolation` |
| `internal/recipeimage/metrics.go` | promauto collectors |
| `internal/recipeimage/processor.go` | `Processor`: Upload / Delete / Get* / Open |
| `internal/recipeimage/reaper.go` | `ObjectReaper` (outbox drainer) |
| `internal/recipeimage/rest.go` | `RestModel` (`recipe-images`), `Transform` |
| `internal/recipeimage/resource.go` | `InitializeRoutes`, PUT/GET/DELETE handlers, `RecipeVisibility` |
| `internal/recipeimage/*_test.go` | domain, processor, reaper, handler tests |
| `internal/recipe/visibility.go` | `ImageVisibility(db)` |
| `internal/recipe/rest.go`, `processor.go` | `image` enrichment |
| `internal/retention/handlers.go` | cascade purges images into outbox |
| `cmd/main.go` | wiring |
| `docs/rest.md`, `docs/storage.md`, `docs/domain.md`, `README.md` | docs |

### Frontend — `frontend/src`

| File | Responsibility |
|---|---|
| `lib/api/client.ts` | `upload` gains `method?: "POST" \| "PUT"` |
| `types/models/recipe.ts` | `RecipeImage`, `RecipeImageResource`, `image` attribute |
| `services/api/recipe.ts` | `uploadRecipeImage`, `deleteRecipeImage` |
| `lib/hooks/api/use-recipes.ts` | `useUploadRecipeImage`, `useDeleteRecipeImage` |
| `lib/recipe-image.ts` | constants + `validateRecipeImageFile` |
| `lib/hooks/use-recipe-image-draft.ts` | `useRecipeImageDraft` |
| `components/features/recipes/recipe-image.tsx` | `<RecipeImage>` hero/card + placeholder |
| `components/features/recipes/recipe-image-field.tsx` | drop zone / preview / replace / remove / undo |
| `components/features/recipes/recipe-card.tsx` | thumbnail |
| `pages/RecipeDetailPage.tsx` | hero + lightbox |
| `pages/RecipeFormPage.tsx` | Photo section, live-preview image, save sequencing |

### Deploy

`deploy/compose/docker-compose.yml`, `deploy/compose/nginx.conf`, `deploy/k8s/recipe-service.yaml`, `deploy/k8s/secrets.example.yaml`, `deploy/k8s/ingress.yaml`, `.env.example`.

---

## Task 1: Object store package, fake, config and startup wiring

**Files:**
- Create: `services/recipe-service/internal/objectstore/objectstore.go`
- Create: `services/recipe-service/internal/objectstore/disabled.go`
- Create: `services/recipe-service/internal/objectstore/minio.go`
- Create: `services/recipe-service/internal/objectstore/objectstore_test.go`
- Create: `services/recipe-service/internal/objectstore/objectstoretest/store.go`
- Create: `services/recipe-service/internal/objectstore/objectstoretest/store_test.go`
- Modify: `services/recipe-service/internal/config/config.go`
- Modify: `services/recipe-service/go.mod`, `go.sum`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `objectstore.Store` interface: `Put(ctx context.Context, key string, body []byte, contentType string) error`, `Get(ctx context.Context, key string) (objectstore.Object, error)`, `Delete(ctx context.Context, key string) error`, `Ping(ctx context.Context) error`
  - `objectstore.Object{Body io.ReadCloser; Size int64; ContentType string}`
  - `objectstore.ErrNotFound`, `objectstore.ErrUnavailable`
  - `objectstore.Config{Endpoint, Bucket, AccessKey, SecretKey, Region string; UseSSL bool}`
  - `objectstore.New(l logrus.FieldLogger, cfg objectstore.Config) objectstore.Store`
  - `objectstoretest.New() *objectstoretest.Store` with exported fields `PutHook func(key string) error`, `GetErr error`, `DeleteErr error`, `PingErr error`, counters `PutCalls`, `GetCalls`, `DeleteCalls int`, and methods `Has(key string) bool`, `Keys() []string`, `Seed(key string, body []byte)`
  - `config.Config.ObjectStore objectstore.Config`

`cmd/main.go` wiring for the store is done in Task 6, together with routes, so `main.go` changes only once.

- [ ] **Step 1: Add the minio-go dependency**

Run from `services/recipe-service`:
```bash
go get github.com/minio/minio-go/v7@latest
```
Expected: `go.mod` gains `github.com/minio/minio-go/v7`.

- [ ] **Step 2: Write the failing tests**

`internal/objectstore/objectstore_test.go`:
```go
package objectstore

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestNewWithEmptyEndpointIsDisabled(t *testing.T) {
	l, _ := test.NewNullLogger()
	s := New(l, Config{Bucket: "b"})
	ctx := context.Background()

	if err := s.Put(ctx, "k", []byte("x"), "image/jpeg"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Put err = %v, want ErrUnavailable", err)
	}
	if _, err := s.Get(ctx, "k"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Get err = %v, want ErrUnavailable", err)
	}
	if err := s.Delete(ctx, "k"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Delete err = %v, want ErrUnavailable", err)
	}
	if err := s.Ping(ctx); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Ping err = %v, want ErrUnavailable", err)
	}
}

func TestNewWithEndpointReturnsMinioStore(t *testing.T) {
	l, _ := test.NewNullLogger()
	s := New(l, Config{Endpoint: "localhost:9000", Bucket: "b", AccessKey: "a", SecretKey: "s", Region: "us-east-1"})
	if _, ok := s.(*minioStore); !ok {
		t.Fatalf("New returned %T, want *minioStore", s)
	}
}
```

`internal/objectstore/objectstoretest/store_test.go`:
```go
package objectstoretest

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

func TestFakeRoundTrip(t *testing.T) {
	s := New()
	ctx := context.Background()
	if err := s.Put(ctx, "a/b.jpg", []byte("hello"), "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	obj, err := s.Get(ctx, "a/b.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	got, _ := io.ReadAll(obj.Body)
	if string(got) != "hello" || obj.Size != 5 || obj.ContentType != "image/jpeg" {
		t.Fatalf("got %q size=%d ct=%q", got, obj.Size, obj.ContentType)
	}
	if err := s.Delete(ctx, "a/b.jpg"); err != nil {
		t.Fatal(err)
	}
	if s.Has("a/b.jpg") {
		t.Fatal("object still present after delete")
	}
	if _, err := s.Get(ctx, "a/b.jpg"); !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("Get after delete err = %v, want ErrNotFound", err)
	}
	// Deleting a missing key is a no-op, like S3/MinIO.
	if err := s.Delete(ctx, "missing"); err != nil {
		t.Fatalf("Delete missing err = %v, want nil", err)
	}
}

func TestFakeFailureInjection(t *testing.T) {
	s := New()
	ctx := context.Background()
	boom := errors.New("boom")
	s.PutHook = func(key string) error {
		if key == "bad" {
			return boom
		}
		return nil
	}
	if err := s.Put(ctx, "good", []byte("x"), "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, "bad", []byte("x"), "image/jpeg"); !errors.Is(err, boom) {
		t.Fatalf("Put bad err = %v", err)
	}
	if s.Has("bad") {
		t.Fatal("failed put must not store the object")
	}
	s.GetErr = boom
	if _, err := s.Get(ctx, "good"); !errors.Is(err, boom) {
		t.Fatalf("Get err = %v", err)
	}
	s.DeleteErr = boom
	if err := s.Delete(ctx, "good"); !errors.Is(err, boom) {
		t.Fatalf("Delete err = %v", err)
	}
	if s.PutCalls != 2 || s.GetCalls != 1 || s.DeleteCalls != 1 {
		t.Fatalf("calls put=%d get=%d delete=%d", s.PutCalls, s.GetCalls, s.DeleteCalls)
	}
}
```

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/objectstore/...`
Expected: FAIL (`undefined: New`, `undefined: Config`, …).

- [ ] **Step 4: Implement the package**

`internal/objectstore/objectstore.go`:
```go
// Package objectstore is recipe-service's thin abstraction over an
// S3-compatible object store. It is the only package that imports minio-go.
package objectstore

import (
	"context"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
)

var (
	// ErrNotFound is returned by Get when the key does not exist.
	ErrNotFound = errors.New("object not found")
	// ErrUnavailable is returned when the store is disabled or unreachable.
	ErrUnavailable = errors.New("object storage unavailable")
)

// Object is a streamed object body. Callers must close Body.
type Object struct {
	Body        io.ReadCloser
	Size        int64
	ContentType string
}

type Store interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	Get(ctx context.Context, key string) (Object, error)
	Delete(ctx context.Context, key string) error
	// Ping checks that the configured bucket exists and is reachable.
	Ping(ctx context.Context) error
}

type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

// New returns a minio-backed Store, or a disabled Store when no endpoint is
// configured or the client cannot be constructed. It never blocks on the
// network.
func New(l logrus.FieldLogger, cfg Config) Store {
	if cfg.Endpoint == "" {
		l.Warn("S3_ENDPOINT not set; object storage disabled, image endpoints will return 503")
		return disabled{}
	}
	s, err := newMinio(cfg)
	if err != nil {
		l.WithError(err).Error("invalid object storage configuration; object storage disabled")
		return disabled{}
	}
	return s
}
```

`internal/objectstore/disabled.go`:
```go
package objectstore

import "context"

type disabled struct{}

func (disabled) Put(context.Context, string, []byte, string) error { return ErrUnavailable }
func (disabled) Get(context.Context, string) (Object, error)       { return Object{}, ErrUnavailable }
func (disabled) Delete(context.Context, string) error              { return ErrUnavailable }
func (disabled) Ping(context.Context) error                        { return ErrUnavailable }
```

`internal/objectstore/minio.go`:
```go
package objectstore

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStore struct {
	client *minio.Client
	bucket string
}

func newMinio(cfg Config) (*minioStore, error) {
	c, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, err
	}
	return &minioStore{client: c, bucket: cfg.Bucket}, nil
}

func (s *minioStore) Put(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(body), int64(len(body)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("%w: put %s: %v", ErrUnavailable, key, err)
	}
	return nil
}

func (s *minioStore) Get(ctx context.Context, key string) (Object, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return Object{}, fmt.Errorf("%w: get %s: %v", ErrUnavailable, key, err)
	}
	// GetObject is lazy; Stat surfaces NoSuchKey and connection errors.
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return Object{}, ErrNotFound
		}
		return Object{}, fmt.Errorf("%w: stat %s: %v", ErrUnavailable, key, err)
	}
	return Object{Body: obj, Size: info.Size, ContentType: info.ContentType}, nil
}

func (s *minioStore) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return ErrNotFound
		}
		return fmt.Errorf("%w: delete %s: %v", ErrUnavailable, key, err)
	}
	return nil
}

func (s *minioStore) Ping(ctx context.Context) error {
	ok, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("%w: bucket check: %v", ErrUnavailable, err)
	}
	if !ok {
		return fmt.Errorf("%w: bucket %q does not exist", ErrUnavailable, s.bucket)
	}
	return nil
}
```

`internal/objectstore/objectstoretest/store.go`:
```go
// Package objectstoretest provides an in-memory objectstore.Store for tests.
package objectstoretest

import (
	"bytes"
	"context"
	"io"
	"sort"
	"sync"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

type stored struct {
	body        []byte
	contentType string
}

type Store struct {
	mu      sync.Mutex
	objects map[string]stored

	// PutHook, when set, is consulted before every Put; a non-nil return
	// fails that Put and nothing is stored.
	PutHook   func(key string) error
	GetErr    error
	DeleteErr error
	PingErr   error

	PutCalls    int
	GetCalls    int
	DeleteCalls int
}

func New() *Store { return &Store{objects: map[string]stored{}} }

func (s *Store) Put(_ context.Context, key string, body []byte, contentType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PutCalls++
	if s.PutHook != nil {
		if err := s.PutHook(key); err != nil {
			return err
		}
	}
	s.objects[key] = stored{body: append([]byte(nil), body...), contentType: contentType}
	return nil
}

func (s *Store) Get(_ context.Context, key string) (objectstore.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GetCalls++
	if s.GetErr != nil {
		return objectstore.Object{}, s.GetErr
	}
	o, ok := s.objects[key]
	if !ok {
		return objectstore.Object{}, objectstore.ErrNotFound
	}
	return objectstore.Object{
		Body:        io.NopCloser(bytes.NewReader(o.body)),
		Size:        int64(len(o.body)),
		ContentType: o.contentType,
	}, nil
}

func (s *Store) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DeleteCalls++
	if s.DeleteErr != nil {
		return s.DeleteErr
	}
	delete(s.objects, key)
	return nil
}

func (s *Store) Ping(context.Context) error { return s.PingErr }

func (s *Store) Has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.objects[key]
	return ok
}

func (s *Store) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.objects))
	for k := range s.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Seed stores an object without counting it as a Put call.
func (s *Store) Seed(key string, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = stored{body: append([]byte(nil), body...), contentType: "image/jpeg"}
}
```

- [ ] **Step 5: Add config**

In `internal/config/config.go`, add the import `"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"` and `"strconv"`, add the field `ObjectStore objectstore.Config` to `Config`, and add this to `Load()`'s returned struct:
```go
		ObjectStore: objectstore.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Bucket:    envOrDefault("S3_BUCKET", "home-hub-recipes"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			UseSSL:    parseBool(os.Getenv("S3_USE_SSL"), false),
			Region:    envOrDefault("S3_REGION", "us-east-1"),
		},
```
and the helper:
```go
func parseBool(v string, fallback bool) bool {
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
```

- [ ] **Step 6: Tidy, build and test**

Run: `cd services/recipe-service && go mod tidy && go build ./... && go test ./internal/objectstore/... ./internal/config/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add services/recipe-service/internal/objectstore services/recipe-service/internal/config services/recipe-service/go.mod services/recipe-service/go.sum
git commit -m "feat(recipe-service): add objectstore package with minio, disabled and fake stores"
```

---

## Task 2: EXIF orientation reader

**Files:**
- Create: `services/recipe-service/internal/recipeimage/imaging/orientation.go`
- Create: `services/recipe-service/internal/recipeimage/imaging/testhelpers_test.go`
- Create: `services/recipe-service/internal/recipeimage/imaging/orientation_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: unexported `jpegOrientation(src []byte) int` (1..8; 1 on absent/malformed). The test helpers `exifAPP1(orientation uint16, bigEndian bool) []byte`, `injectAPP1(jpg, app1 []byte) []byte`, `encodeJPEG(t testing.TB, img image.Image) []byte`, and `quadrantImage(w, h int) *image.RGBA` are used by Task 3's tests in the same package.

- [ ] **Step 1: Write the test helpers**

`internal/recipeimage/imaging/testhelpers_test.go`:
```go
package imaging

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// exifAPP1 builds a complete APP1 segment (marker, length, payload) carrying
// a single IFD0 Orientation entry.
func exifAPP1(orientation uint16, bigEndian bool) []byte {
	var bo binary.AppendByteOrder = binary.LittleEndian
	tiff := []byte("II*\x00")
	if bigEndian {
		bo = binary.BigEndian
		tiff = []byte("MM\x00*")
	}
	tiff = bo.AppendUint32(tiff, 8)      // IFD0 offset
	tiff = bo.AppendUint16(tiff, 1)      // entry count
	tiff = bo.AppendUint16(tiff, 0x0112) // Orientation tag
	tiff = bo.AppendUint16(tiff, 3)      // SHORT
	tiff = bo.AppendUint32(tiff, 1)      // count
	tiff = bo.AppendUint16(tiff, orientation)
	tiff = append(tiff, 0, 0)       // value padding to 4 bytes
	tiff = bo.AppendUint32(tiff, 0) // next IFD
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1}
	seg = binary.BigEndian.AppendUint16(seg, uint16(len(payload)+2))
	return append(seg, payload...)
}

// injectAPP1 inserts a segment right after SOI.
func injectAPP1(jpg, app1 []byte) []byte {
	out := make([]byte, 0, len(jpg)+len(app1))
	out = append(out, jpg[:2]...)
	out = append(out, app1...)
	return append(out, jpg[2:]...)
}

func encodeJPEG(t testing.TB, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// quadrantImage is blue with a solid red block in the top-left quarter, so a
// test can tell where the original top-left corner ended up.
func quadrantImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.RGBA{0, 0, 255, 255}
			if x < w/2 && y < h/2 {
				c = color.RGBA{255, 0, 0, 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// pngHeaderOnly returns a PNG signature plus a valid IHDR chunk for an 8-bit
// RGB image of the given size, with no pixel data. DecodeConfig accepts it;
// a full decode would fail.
func pngHeaderOnly(w, h uint32) []byte {
	b := []byte("\x89PNG\r\n\x1a\n")
	ihdr := binary.BigEndian.AppendUint32(nil, w)
	ihdr = binary.BigEndian.AppendUint32(ihdr, h)
	ihdr = append(ihdr, 8, 2, 0, 0, 0)
	b = binary.BigEndian.AppendUint32(b, uint32(len(ihdr)))
	chunk := append([]byte("IHDR"), ihdr...)
	b = append(b, chunk...)
	return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(chunk))
}
```

- [ ] **Step 2: Write the failing orientation tests**

`internal/recipeimage/imaging/orientation_test.go`:
```go
package imaging

import (
	"testing"
)

func TestJPEGOrientation(t *testing.T) {
	base := encodeJPEG(t, quadrantImage(32, 32))

	for _, be := range []bool{false, true} {
		for o := uint16(1); o <= 8; o++ {
			got := jpegOrientation(injectAPP1(base, exifAPP1(o, be)))
			if got != int(o) {
				t.Errorf("bigEndian=%v orientation=%d: got %d", be, o, got)
			}
		}
	}

	cases := []struct {
		name string
		src  []byte
	}{
		{"no APP1", base},
		{"empty", nil},
		{"not a jpeg", []byte("GIF89a....")},
		{"out of range 0", injectAPP1(base, exifAPP1(0, false))},
		{"out of range 9", injectAPP1(base, exifAPP1(9, false))},
		{"wrong type", injectAPP1(base, mutate(exifAPP1(6, false), 22, 4))},  // type LONG
		{"wrong count", injectAPP1(base, mutate(exifAPP1(6, false), 24, 2))}, // count 2
		{"bad byte order", injectAPP1(base, mutate(exifAPP1(6, false), 10, 'X'))},
		{"non-exif APP1", injectAPP1(base, mutate(exifAPP1(6, false), 4, 'X'))},
	}
	for _, c := range cases {
		if got := jpegOrientation(c.src); got != 1 {
			t.Errorf("%s: got %d, want 1", c.name, got)
		}
	}
}

func TestJPEGOrientationTruncatedAtEveryOffset(t *testing.T) {
	full := injectAPP1(encodeJPEG(t, quadrantImage(16, 16)), exifAPP1(6, false))
	for n := 0; n < 64 && n < len(full); n++ {
		got := jpegOrientation(full[:n])
		if got < 1 || got > 8 {
			t.Fatalf("truncated at %d: got %d", n, got)
		}
	}
}

// mutate returns a copy of b with byte i set to v. Offsets into an
// exifAPP1(…, false) segment: 0-1 marker, 2-3 length, 4-9 "Exif\0\0",
// 10-13 TIFF byte order + magic, 14-17 IFD0 offset, 18-19 entry count,
// 20-21 tag, 22-23 type, 24-27 count, 28-29 value (little-endian, so the
// low byte of each field comes first).
func mutate(b []byte, i int, v byte) []byte {
	out := append([]byte(nil), b...)
	out[i] = v
	return out
}

func FuzzJPEGOrientation(f *testing.F) {
	f.Add(injectAPP1(encodeJPEG(f, quadrantImage(16, 16)), exifAPP1(6, false)))
	f.Add(injectAPP1(encodeJPEG(f, quadrantImage(16, 16)), exifAPP1(3, true)))
	f.Add([]byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x02})
	f.Fuzz(func(t *testing.T, b []byte) {
		got := jpegOrientation(b)
		if got < 1 || got > 8 {
			t.Fatalf("got %d", got)
		}
	})
}
```

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/recipeimage/imaging/`
Expected: FAIL `undefined: jpegOrientation`.

- [ ] **Step 4: Implement**

`internal/recipeimage/imaging/orientation.go`:
```go
package imaging

import "encoding/binary"

const tagOrientation = 0x0112

// jpegOrientation returns the EXIF Orientation (1..8) of a JPEG, or 1 when it
// is absent or anything about the structure is malformed. It examines only
// the first APP1 segment whose payload starts with "Exif\0\0", and stops at
// SOS. Every slice access is bounds-checked: input is untrusted.
func jpegOrientation(b []byte) int {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return 1
	}
	i := 2
	for i+4 <= len(b) {
		if b[i] != 0xFF {
			return 1
		}
		marker := b[i+1]
		switch {
		case marker == 0xFF: // fill byte
			i++
			continue
		case marker == 0xDA || marker == 0xD9: // SOS / EOI
			return 1
		case marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7): // standalone
			i += 2
			continue
		}
		segLen := int(binary.BigEndian.Uint16(b[i+2 : i+4]))
		if segLen < 2 || i+2+segLen > len(b) {
			return 1
		}
		payload := b[i+4 : i+2+segLen]
		if marker == 0xE1 && len(payload) >= 6 && string(payload[:6]) == "Exif\x00\x00" {
			return tiffOrientation(payload[6:])
		}
		i += 2 + segLen
	}
	return 1
}

func tiffOrientation(t []byte) int {
	if len(t) < 8 {
		return 1
	}
	var bo binary.ByteOrder
	switch string(t[:4]) {
	case "II*\x00":
		bo = binary.LittleEndian
	case "MM\x00*":
		bo = binary.BigEndian
	default:
		return 1
	}
	off := uint64(bo.Uint32(t[4:8]))
	if off < 8 || off+2 > uint64(len(t)) {
		return 1
	}
	n := uint64(bo.Uint16(t[off : off+2]))
	for k := uint64(0); k < n; k++ {
		e := off + 2 + k*12
		if e+12 > uint64(len(t)) {
			return 1
		}
		if bo.Uint16(t[e:e+2]) != tagOrientation {
			continue
		}
		if bo.Uint16(t[e+2:e+4]) != 3 || bo.Uint32(t[e+4:e+8]) != 1 {
			return 1
		}
		v := int(bo.Uint16(t[e+8 : e+10]))
		if v < 1 || v > 8 {
			return 1
		}
		return v
	}
	return 1
}
```

- [ ] **Step 5: Run the tests and a short fuzz**

Run: `cd services/recipe-service && go test ./internal/recipeimage/imaging/ && go test ./internal/recipeimage/imaging/ -run=^$ -fuzz=FuzzJPEGOrientation -fuzztime=20s`
Expected: PASS; fuzz finds no failures.

- [ ] **Step 6: Commit**

```bash
git add services/recipe-service/internal/recipeimage/imaging
git commit -m "feat(recipe-service): add bounds-checked EXIF orientation reader"
```

---

## Task 3: Image processing pipeline (`imaging.Process`)

**Files:**
- Create: `services/recipe-service/internal/recipeimage/imaging/imaging.go`
- Create: `services/recipe-service/internal/recipeimage/imaging/transform.go`
- Create: `services/recipe-service/internal/recipeimage/imaging/imaging_test.go`
- Modify: `services/recipe-service/go.mod`, `go.sum`

**Interfaces:**
- Consumes: `jpegOrientation` (Task 2), the test helpers (Task 2).
- Produces:
  - `imaging.Process(src []byte) (imaging.Result, error)`
  - `imaging.Result{Full, Thumb imaging.Rendition}`, `imaging.Rendition{Bytes []byte; Width, Height int}`
  - `imaging.ErrUnsupportedFormat`, `imaging.ErrTooLarge`, `imaging.ErrTooSmall`
  - constants `imaging.MaxPixels = 40_000_000`, `imaging.MinEdge = 16`, `imaging.FullMaxEdge = 1600`, `imaging.ThumbMaxEdge = 480`, `imaging.JPEGQuality = 82`

- [ ] **Step 1: Add the dependency**

Run: `cd services/recipe-service && go get golang.org/x/image@latest`

- [ ] **Step 2: Write the failing tests**

`internal/recipeimage/imaging/imaging_test.go`:
```go
package imaging

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"
)

// 32x24 lossless WebP: left half fully transparent, right half opaque blue.
const webpFixture = "UklGRiQAAABXRUJQVlA4TBgAAAAvH8AFEA8QMf/zHwyBQJK/9AIDRPQ/qDc="

func decodeOut(t *testing.T, b []byte) image.Image {
	t.Helper()
	img, format, err := image.Decode(bytes.NewReader(b))
	if err != nil || format != "jpeg" {
		t.Fatalf("output not a JPEG: format=%q err=%v", format, err)
	}
	return img
}

func isRed(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r>>8 > 180 && g>>8 < 90 && b>>8 < 90
}

func isWhite(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r>>8 > 235 && g>>8 > 235 && b>>8 > 235
}

func hasAPP1(b []byte) bool {
	return bytes.Contains(b, []byte{0xFF, 0xE1})
}

func TestProcessOrientation(t *testing.T) {
	base := encodeJPEG(t, quadrantImage(64, 32))
	cases := []struct {
		o          uint16
		w, h       int
		redX, redY int // a point inside where the red quadrant must land
	}{
		{1, 64, 32, 6, 6},
		{3, 64, 32, 64 - 7, 32 - 7},
		{6, 32, 64, 32 - 7, 6},
		{8, 32, 64, 6, 64 - 7},
	}
	for _, c := range cases {
		res, err := Process(injectAPP1(base, exifAPP1(c.o, false)))
		if err != nil {
			t.Fatalf("o=%d: %v", c.o, err)
		}
		if res.Full.Width != c.w || res.Full.Height != c.h {
			t.Fatalf("o=%d: full %dx%d, want %dx%d", c.o, res.Full.Width, res.Full.Height, c.w, c.h)
		}
		img := decodeOut(t, res.Full.Bytes)
		if !isRed(img.At(c.redX, c.redY)) {
			t.Errorf("o=%d: pixel (%d,%d) = %v, want red", c.o, c.redX, c.redY, img.At(c.redX, c.redY))
		}
		if hasAPP1(res.Full.Bytes) || hasAPP1(res.Thumb.Bytes) {
			t.Errorf("o=%d: output carries an APP1 (EXIF) segment", c.o)
		}
	}
}

func TestProcessOrientedDownscale(t *testing.T) {
	src := injectAPP1(encodeJPEG(t, quadrantImage(2000, 1000)), exifAPP1(6, false))
	res, err := Process(src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Full.Width != 800 || res.Full.Height != 1600 {
		t.Fatalf("full %dx%d, want 800x1600", res.Full.Width, res.Full.Height)
	}
	if res.Thumb.Width != 240 || res.Thumb.Height != 480 {
		t.Fatalf("thumb %dx%d, want 240x480", res.Thumb.Width, res.Thumb.Height)
	}
	img := decodeOut(t, res.Thumb.Bytes)
	if b := img.Bounds(); b.Dx() != 240 || b.Dy() != 480 {
		t.Fatalf("decoded thumb %v", b)
	}
}

func TestProcessNeverUpscales(t *testing.T) {
	res, err := Process(encodeJPEG(t, quadrantImage(100, 50)))
	if err != nil {
		t.Fatal(err)
	}
	if res.Full.Width != 100 || res.Full.Height != 50 || res.Thumb.Width != 100 || res.Thumb.Height != 50 {
		t.Fatalf("full %dx%d thumb %dx%d", res.Full.Width, res.Full.Height, res.Thumb.Width, res.Thumb.Height)
	}
}

func TestProcessPNGAlphaCompositesWhite(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			if x >= 20 {
				img.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
			} // left half stays fully transparent
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	res, err := Process(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	out := decodeOut(t, res.Full.Bytes)
	if !isWhite(out.At(5, 10)) {
		t.Errorf("transparent area = %v, want white", out.At(5, 10))
	}
	if !isRed(out.At(35, 10)) {
		t.Errorf("opaque area = %v, want red", out.At(35, 10))
	}
}

func TestProcessPalettedPNGWithTransparency(t *testing.T) {
	pal := color.Palette{color.NRGBA{0, 0, 0, 0}, color.NRGBA{255, 0, 0, 255}}
	img := image.NewPaletted(image.Rect(0, 0, 32, 32), pal)
	for y := 0; y < 32; y++ {
		for x := 16; x < 32; x++ {
			img.SetColorIndex(x, y, 1)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	res, err := Process(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	out := decodeOut(t, res.Full.Bytes)
	if !isWhite(out.At(4, 16)) || !isRed(out.At(28, 16)) {
		t.Errorf("paletted composite wrong: left=%v right=%v", out.At(4, 16), out.At(28, 16))
	}

	gray := image.NewGray(image.Rect(0, 0, 20, 20))
	buf.Reset()
	if err := png.Encode(&buf, gray); err != nil {
		t.Fatal(err)
	}
	if _, err := Process(buf.Bytes()); err != nil {
		t.Fatalf("grayscale PNG: %v", err)
	}
}

func TestProcessWebP(t *testing.T) {
	src, _ := base64.StdEncoding.DecodeString(webpFixture)
	res, err := Process(src)
	if err != nil {
		t.Fatal(err)
	}
	if res.Full.Width != 32 || res.Full.Height != 24 {
		t.Fatalf("full %dx%d", res.Full.Width, res.Full.Height)
	}
	out := decodeOut(t, res.Full.Bytes)
	if !isWhite(out.At(4, 12)) {
		t.Errorf("transparent webp area = %v, want white", out.At(4, 12))
	}
}

func TestProcessRejections(t *testing.T) {
	var gifBuf bytes.Buffer
	if err := gif.Encode(&gifBuf, image.NewPaletted(image.Rect(0, 0, 32, 32), color.Palette{color.Black, color.White}), nil); err != nil {
		t.Fatal(err)
	}
	tiny := encodeJPEG(t, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	bomb := pngHeaderOnly(10000, 5000)

	cases := []struct {
		name string
		src  []byte
		want error
	}{
		{"gif", gifBuf.Bytes(), ErrUnsupportedFormat},
		{"garbage", []byte("definitely not an image"), ErrUnsupportedFormat},
		{"empty", nil, ErrUnsupportedFormat},
		{"too small", tiny, ErrTooSmall},
		{"too many pixels (header only)", bomb, ErrTooLarge},
	}
	for _, c := range cases {
		if _, err := Process(c.src); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
	if len(bomb) > 64 {
		t.Fatalf("bomb fixture is %d bytes; it must be header-only to prove no full decode", len(bomb))
	}
}

func TestProcessOutputIsBaselineJPEGQuality(t *testing.T) {
	res, err := Process(encodeJPEG(t, quadrantImage(64, 64)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.DecodeConfig(bytes.NewReader(res.Thumb.Bytes)); err != nil {
		t.Fatal(err)
	}
}

func FuzzProcess(f *testing.F) {
	f.Add(encodeJPEG(f, quadrantImage(32, 32)))
	src, _ := base64.StdEncoding.DecodeString(webpFixture)
	f.Add(src)
	f.Add(pngHeaderOnly(100, 100))
	f.Fuzz(func(t *testing.T, b []byte) {
		res, err := Process(b)
		if err != nil {
			return
		}
		if res.Full.Width < 1 || res.Full.Width > FullMaxEdge || res.Full.Height > FullMaxEdge {
			t.Fatalf("full %dx%d", res.Full.Width, res.Full.Height)
		}
	})
}

// BenchmarkProcess12MP measures the PRD budget (< 2 s for a 12 MP JPEG).
// Record the result in services/recipe-service/docs/storage.md (Task 13).
func BenchmarkProcess12MP(b *testing.B) {
	src := injectAPP1(encodeJPEG(b, quadrantImage(4000, 3000)), exifAPP1(6, false))
	b.SetBytes(int64(len(src)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Process(src); err != nil {
			b.Fatal(err)
		}
	}
}
```

Note: `encodeJPEG` takes `testing.TB`, so it can be called with `*testing.F` and `*testing.B`.

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/recipeimage/imaging/`
Expected: FAIL `undefined: Process`.

- [ ] **Step 4: Implement `imaging.go`**

```go
// Package imaging turns untrusted upload bytes into two JPEG renditions. It
// does no I/O. Decoders are registered here and nowhere else in the service.
package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const (
	MaxPixels    = 40_000_000
	MinEdge      = 16
	FullMaxEdge  = 1600
	ThumbMaxEdge = 480
	JPEGQuality  = 82
)

var (
	ErrUnsupportedFormat = errors.New("unsupported image format")
	ErrTooLarge          = errors.New("image dimensions exceed limit")
	ErrTooSmall          = errors.New("image dimensions below minimum")
)

type Rendition struct {
	Bytes         []byte
	Width, Height int
}

type Result struct {
	Full, Thumb Rendition
}

// Process validates and converts src. Format comes only from the bytes.
// Dimension limits are enforced from the header before the full decode.
func Process(src []byte) (Result, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return Result{}, ErrUnsupportedFormat
	}
	switch format {
	case "jpeg", "png", "webp":
	default:
		return Result{}, ErrUnsupportedFormat
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return Result{}, ErrTooLarge
	}
	if min(cfg.Width, cfg.Height) < MinEdge {
		return Result{}, ErrTooSmall
	}

	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return Result{}, ErrUnsupportedFormat
	}

	orientation := 1
	if format == "jpeg" {
		orientation = jpegOrientation(src)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	ow, oh := w, h
	if swapsAxes(orientation) {
		ow, oh = h, w
	}
	tw, th := fit(ow, oh, FullMaxEdge)
	sw, sh := tw, th
	if swapsAxes(orientation) {
		sw, sh = th, tw
	}

	// Resize first, then orient: rotating 1600 px is far cheaper than 4000 px,
	// and the two commute once the target axes are swapped.
	full := orient(scaleOntoWhite(img, sw, sh), orientation)

	thw, thh := fit(full.Bounds().Dx(), full.Bounds().Dy(), ThumbMaxEdge)
	thumb := scaleOntoWhite(full, thw, thh)

	fullBytes, err := encode(full)
	if err != nil {
		return Result{}, err
	}
	thumbBytes, err := encode(thumb)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Full:  Rendition{Bytes: fullBytes, Width: full.Bounds().Dx(), Height: full.Bounds().Dy()},
		Thumb: Rendition{Bytes: thumbBytes, Width: thumb.Bounds().Dx(), Height: thumb.Bounds().Dy()},
	}, nil
}

// encode re-encodes as baseline JPEG; the stdlib encoder writes no EXIF, XMP
// or ICC segments, which strips all metadata including GPS.
func encode(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 5: Implement `transform.go`**

```go
package imaging

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
)

func swapsAxes(orientation int) bool { return orientation >= 5 && orientation <= 8 }

// fit returns w,h scaled so the longest edge is at most maxEdge. It never
// upscales and never returns a zero edge.
func fit(w, h, maxEdge int) (int, int) {
	if w <= maxEdge && h <= maxEdge {
		return w, h
	}
	if w >= h {
		return maxEdge, max(1, int(math.Round(float64(h)*float64(maxEdge)/float64(w))))
	}
	return max(1, int(math.Round(float64(w)*float64(maxEdge)/float64(h)))), maxEdge
}

// scaleOntoWhite draws src onto an opaque white w×h canvas, compositing any
// transparency onto white in the same pass as the resize.
func scaleOntoWhite(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	sb := src.Bounds()
	if sb.Dx() == w && sb.Dy() == h {
		draw.Draw(dst, dst.Bounds(), src, sb.Min, draw.Over)
		return dst
	}
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, sb, draw.Over, nil)
	return dst
}

// orient applies an EXIF orientation transform. src must have Min = (0,0).
func orient(src *image.RGBA, o int) *image.RGBA {
	if o < 2 || o > 8 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dw, dh := w, h
	if swapsAxes(o) {
		dw, dh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			var sx, sy int
			switch o {
			case 2:
				sx, sy = w-1-x, y
			case 3:
				sx, sy = w-1-x, h-1-y
			case 4:
				sx, sy = x, h-1-y
			case 5:
				sx, sy = y, x
			case 6:
				sx, sy = y, h-1-x
			case 7:
				sx, sy = w-1-y, h-1-x
			case 8:
				sx, sy = w-1-y, x
			}
			dst.SetRGBA(x, y, src.RGBAAt(sx, sy))
		}
	}
	return dst
}
```

- [ ] **Step 6: Run the tests, the fuzz and the benchmark**

Run:
```bash
cd services/recipe-service
go mod tidy
go test ./internal/recipeimage/imaging/
go test ./internal/recipeimage/imaging/ -run=^$ -fuzz=FuzzProcess -fuzztime=30s
go test ./internal/recipeimage/imaging/ -run=^$ -bench=BenchmarkProcess12MP -benchtime=5x
```
Expected: tests PASS; fuzz finds nothing; the benchmark prints ns/op. **Write the ns/op figure into your task report.** Task 13 records it in `docs/storage.md`. If it is ≥ 2 s/op, implement the design §5.2 fallback in `scaleOntoWhite`: when the source is more than 2× the target on both axes, first scale with `draw.ApproxBiLinear` to 2× the target, then with `draw.CatmullRom` to the target. Re-run the benchmark.

- [ ] **Step 7: Commit**

```bash
git add services/recipe-service/internal/recipeimage/imaging services/recipe-service/go.mod services/recipe-service/go.sum
git commit -m "feat(recipe-service): add image processing pipeline with orientation and limits"
```

---

## Task 4: `recipeimage` domain — entities, model, builder, keys, provider, administrator

**Files:**
- Create: `services/recipe-service/internal/recipeimage/entity.go`
- Create: `services/recipe-service/internal/recipeimage/model.go`
- Create: `services/recipe-service/internal/recipeimage/builder.go`
- Create: `services/recipe-service/internal/recipeimage/keys.go`
- Create: `services/recipe-service/internal/recipeimage/errors.go`
- Create: `services/recipe-service/internal/recipeimage/provider.go`
- Create: `services/recipe-service/internal/recipeimage/administrator.go`
- Create: `services/recipe-service/internal/recipeimage/testdb_test.go`
- Create: `services/recipe-service/internal/recipeimage/domain_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces:
  - `recipeimage.Entity` (table `recipe_images`) and `recipeimage.ObjectDeletionEntity` (table `recipe_image_object_deletions`); `recipeimage.Migration(db *gorm.DB) error`
  - `recipeimage.Model` getters: `Id()`, `TenantID()`, `HouseholdID()`, `RecipeID() uuid.UUID`; `FullKey()`, `ThumbKey()`, `ContentType() string`; `Width()`, `Height()`, `FullBytes()`, `ThumbBytes() int`; `CreatedAt()`, `UpdatedAt() time.Time`; `Key(size Size) string`
  - `recipeimage.Make(e Entity) (Model, error)`, `recipeimage.NewBuilder() *Builder`
  - `recipeimage.Size` with `SizeFull = "full"` and `SizeThumb = "thumb"`; `recipeimage.ParseSize(s string) (Size, bool)` (`""` → full)
  - `recipeimage.URLs(recipeID, imageID uuid.UUID) (thumb, full string)`; unexported `objectKeys(tenantID, householdID, recipeID, imageID uuid.UUID) (full, thumb string)`
  - errors `ErrRecipeNotFound`, `ErrImageNotFound`, `ErrObjectMissing`, `ErrStorage`, `ErrNoTenant`; unexported `isUniqueViolation(err error) bool`
  - unexported `getByRecipeID(recipeID, householdID uuid.UUID) database.EntityProvider[Entity]`, `getByRecipeIDs(recipeIDs []uuid.UUID, householdID uuid.UUID) func(*gorm.DB) ([]Entity, error)`, `getPendingDeletions(limit int)`, `getDeletionsByIDs(ids)`
  - unexported `deleteByRecipeIDReturning(tx *gorm.DB, recipeID, householdID uuid.UUID) ([]Entity, error)`, `createImage(tx *gorm.DB, e *Entity) error`, `enqueueObjectDeletions(tx *gorm.DB, removed []Entity) ([]uuid.UUID, error)`
  - exported `recipeimage.PurgeByRecipeIDsTx(tx *gorm.DB, recipeIDs []uuid.UUID) (int, error)` (used by retention in Task 8)
  - test helper `newTestDB(t *testing.T) *gorm.DB` (sqlite memory, one connection, tenant callbacks, migrated) and `tenantCtx(tenantID, householdID uuid.UUID) context.Context`

- [ ] **Step 1: Write the test DB helper**

`internal/recipeimage/testdb_test.go`:
```go
package recipeimage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

// newTestDB returns an in-memory SQLite DB pinned to one connection so the
// schema survives across goroutines (each new :memory: connection would be
// a fresh, empty database).
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := Migration(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func tenantCtx(tenantID, householdID uuid.UUID) context.Context {
	return tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, uuid.New()))
}
```

- [ ] **Step 2: Write the failing domain tests**

`internal/recipeimage/domain_test.go`:
```go
package recipeimage

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedImage(t *testing.T, db *gorm.DB, tenantID, householdID, recipeID uuid.UUID) Entity {
	t.Helper()
	id := uuid.New()
	full, thumb := objectKeys(tenantID, householdID, recipeID, id)
	now := time.Now().UTC()
	e := Entity{
		Id: id, TenantId: tenantID, HouseholdId: householdID, RecipeId: recipeID,
		FullKey: full, ThumbKey: thumb, ContentType: "image/jpeg",
		Width: 1600, Height: 900, FullBytes: 1000, ThumbBytes: 100,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

func TestObjectKeysAndURLs(t *testing.T) {
	tid, hid, rid, iid := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	full, thumb := objectKeys(tid, hid, rid, iid)
	prefix := tid.String() + "/" + hid.String() + "/recipes/" + rid.String() + "/" + iid.String() + "/"
	if full != prefix+"full.jpg" || thumb != prefix+"thumb.jpg" {
		t.Fatalf("keys = %q, %q", full, thumb)
	}
	tu, fu := URLs(rid, iid)
	if tu != "/api/v1/recipes/"+rid.String()+"/image?size=thumb&v="+iid.String() {
		t.Fatalf("thumb url = %q", tu)
	}
	if fu != "/api/v1/recipes/"+rid.String()+"/image?size=full&v="+iid.String() {
		t.Fatalf("full url = %q", fu)
	}
}

func TestParseSize(t *testing.T) {
	for in, want := range map[string]Size{"": SizeFull, "full": SizeFull, "thumb": SizeThumb} {
		got, ok := ParseSize(in)
		if !ok || got != want {
			t.Errorf("ParseSize(%q) = %q,%v", in, got, ok)
		}
	}
	for _, bad := range []string{"FULL", "medium", "thumb "} {
		if _, ok := ParseSize(bad); ok {
			t.Errorf("ParseSize(%q) accepted", bad)
		}
	}
}

func TestBuilderValidation(t *testing.T) {
	if _, err := NewBuilder().SetFullKey("a").SetThumbKey("b").Build(); !errors.Is(err, ErrRecipeIDRequired) {
		t.Fatalf("err = %v", err)
	}
	if _, err := NewBuilder().SetRecipeID(uuid.New()).Build(); !errors.Is(err, ErrKeysRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestMakeAndKey(t *testing.T) {
	db := newTestDB(t)
	e := seedImage(t, db, uuid.New(), uuid.New(), uuid.New())
	m, err := Make(e)
	if err != nil {
		t.Fatal(err)
	}
	if m.Key(SizeFull) != e.FullKey || m.Key(SizeThumb) != e.ThumbKey || m.Width() != 1600 {
		t.Fatalf("model mismatch: %+v", m)
	}
}

func TestProviderScopesByHousehold(t *testing.T) {
	db := newTestDB(t)
	tid, hidA, hidB := uuid.New(), uuid.New(), uuid.New()
	r1, r2 := uuid.New(), uuid.New()
	seedImage(t, db, tid, hidA, r1)
	seedImage(t, db, tid, hidB, r2)

	ctxA := tenantCtx(tid, hidA)
	if _, err := getByRecipeID(r1, hidA)(db.WithContext(ctxA))(); err != nil {
		t.Fatalf("own household: %v", err)
	}
	if _, err := getByRecipeID(r2, hidA)(db.WithContext(ctxA))(); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other household err = %v, want not found", err)
	}
	rows, err := getByRecipeIDs([]uuid.UUID{r1, r2}, hidA)(db.WithContext(ctxA))
	if err != nil || len(rows) != 1 || rows[0].RecipeId != r1 {
		t.Fatalf("batched rows = %+v err=%v", rows, err)
	}
	// Tenant callback: another tenant sees nothing even with the right household.
	if _, err := getByRecipeID(r1, hidA)(db.WithContext(tenantCtx(uuid.New(), hidA)))(); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other tenant err = %v", err)
	}
}

func TestDeleteReturningEnqueuesKeys(t *testing.T) {
	db := newTestDB(t)
	tid, hid, rid := uuid.New(), uuid.New(), uuid.New()
	e := seedImage(t, db, tid, hid, rid)

	var ids []uuid.UUID
	err := db.WithContext(tenantCtx(tid, hid)).Transaction(func(tx *gorm.DB) error {
		removed, err := deleteByRecipeIDReturning(tx, rid, hid)
		if err != nil {
			return err
		}
		if len(removed) != 1 || removed[0].FullKey != e.FullKey {
			t.Fatalf("removed = %+v", removed)
		}
		ids, err = enqueueObjectDeletions(tx, removed)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("outbox ids = %d, want 2", len(ids))
	}
	var keys []string
	db.Model(&ObjectDeletionEntity{}).Order("object_key").Pluck("object_key", &keys)
	if len(keys) != 2 || !strings.HasSuffix(keys[0], "full.jpg") || !strings.HasSuffix(keys[1], "thumb.jpg") {
		t.Fatalf("outbox keys = %v", keys)
	}
	var count int64
	db.Model(&Entity{}).Count(&count)
	if count != 0 {
		t.Fatalf("image rows = %d, want 0", count)
	}
}

func TestPurgeByRecipeIDsTx(t *testing.T) {
	db := newTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	r1, r2, r3 := uuid.New(), uuid.New(), uuid.New()
	seedImage(t, db, tid, hid, r1)
	seedImage(t, db, tid, hid, r2)
	seedImage(t, db, tid, hid, r3)

	n, err := PurgeByRecipeIDsTx(db, []uuid.UUID{r1, r2})
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	var outbox, images int64
	db.Model(&ObjectDeletionEntity{}).Count(&outbox)
	db.Model(&Entity{}).Count(&images)
	if outbox != 4 || images != 1 {
		t.Fatalf("outbox=%d images=%d", outbox, images)
	}
	if n, err := PurgeByRecipeIDsTx(db, nil); err != nil || n != 0 {
		t.Fatalf("empty purge n=%d err=%v", n, err)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	db := newTestDB(t)
	tid, hid, rid := uuid.New(), uuid.New(), uuid.New()
	e := seedImage(t, db, tid, hid, rid)
	dup := e
	dup.Id = uuid.New()
	err := db.Create(&dup).Error
	if err == nil || !isUniqueViolation(err) {
		t.Fatalf("err = %v, want unique violation", err)
	}
	if isUniqueViolation(errors.New("other")) {
		t.Fatal("false positive")
	}
}
```

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/recipeimage/`
Expected: FAIL (undefined symbols).

- [ ] **Step 4: Implement**

`internal/recipeimage/entity.go`:
```go
package recipeimage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null;index:idx_recipe_image_tenant_household"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null;index:idx_recipe_image_tenant_household"`
	RecipeId    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_recipe_image_recipe"`
	FullKey     string    `gorm:"type:varchar(512);not null"`
	ThumbKey    string    `gorm:"type:varchar(512);not null"`
	ContentType string    `gorm:"type:varchar(50);not null"`
	Width       int       `gorm:"type:int;not null"`
	Height      int       `gorm:"type:int;not null"`
	FullBytes   int       `gorm:"type:int;not null"`
	ThumbBytes  int       `gorm:"type:int;not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "recipe_images" }

// ObjectDeletionEntity is the transactional outbox for object deletions. Rows
// are written in the same transaction that removes an image row and drained
// after commit by ObjectReaper.
type ObjectDeletionEntity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null"`
	ObjectKey   string    `gorm:"type:varchar(512);not null"`
	Attempts    int       `gorm:"type:int;not null;default:0"`
	LastError   *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;index:idx_recipe_image_deletion_created"`
}

func (ObjectDeletionEntity) TableName() string { return "recipe_image_object_deletions" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{}, &ObjectDeletionEntity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id: m.id, TenantId: m.tenantID, HouseholdId: m.householdID, RecipeId: m.recipeID,
		FullKey: m.fullKey, ThumbKey: m.thumbKey, ContentType: m.contentType,
		Width: m.width, Height: m.height, FullBytes: m.fullBytes, ThumbBytes: m.thumbBytes,
		CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetRecipeID(e.RecipeId).
		SetFullKey(e.FullKey).
		SetThumbKey(e.ThumbKey).
		SetContentType(e.ContentType).
		SetWidth(e.Width).
		SetHeight(e.Height).
		SetFullBytes(e.FullBytes).
		SetThumbBytes(e.ThumbBytes).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
```

`internal/recipeimage/model.go`:
```go
package recipeimage

import (
	"time"

	"github.com/google/uuid"
)

type Size string

const (
	SizeFull  Size = "full"
	SizeThumb Size = "thumb"
)

// ParseSize maps the ?size= query value. Empty means full.
func ParseSize(s string) (Size, bool) {
	switch s {
	case "", string(SizeFull):
		return SizeFull, true
	case string(SizeThumb):
		return SizeThumb, true
	default:
		return "", false
	}
}

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	recipeID    uuid.UUID
	fullKey     string
	thumbKey    string
	contentType string
	width       int
	height      int
	fullBytes   int
	thumbBytes  int
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) RecipeID() uuid.UUID    { return m.recipeID }
func (m Model) FullKey() string        { return m.fullKey }
func (m Model) ThumbKey() string       { return m.thumbKey }
func (m Model) ContentType() string    { return m.contentType }
func (m Model) Width() int             { return m.width }
func (m Model) Height() int            { return m.height }
func (m Model) FullBytes() int         { return m.fullBytes }
func (m Model) ThumbBytes() int        { return m.thumbBytes }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }

func (m Model) Key(size Size) string {
	if size == SizeThumb {
		return m.thumbKey
	}
	return m.fullKey
}
```

`internal/recipeimage/builder.go`:
```go
package recipeimage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRecipeIDRequired = errors.New("recipe ID is required")
	ErrKeysRequired     = errors.New("object keys are required")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	recipeID    uuid.UUID
	fullKey     string
	thumbKey    string
	contentType string
	width       int
	height      int
	fullBytes   int
	thumbBytes  int
	createdAt   time.Time
	updatedAt   time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder    { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder { b.householdID = id; return b }
func (b *Builder) SetRecipeID(id uuid.UUID) *Builder    { b.recipeID = id; return b }
func (b *Builder) SetFullKey(k string) *Builder         { b.fullKey = k; return b }
func (b *Builder) SetThumbKey(k string) *Builder        { b.thumbKey = k; return b }
func (b *Builder) SetContentType(ct string) *Builder    { b.contentType = ct; return b }
func (b *Builder) SetWidth(v int) *Builder              { b.width = v; return b }
func (b *Builder) SetHeight(v int) *Builder             { b.height = v; return b }
func (b *Builder) SetFullBytes(v int) *Builder          { b.fullBytes = v; return b }
func (b *Builder) SetThumbBytes(v int) *Builder         { b.thumbBytes = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder    { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder    { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.recipeID == uuid.Nil {
		return Model{}, ErrRecipeIDRequired
	}
	if b.fullKey == "" || b.thumbKey == "" {
		return Model{}, ErrKeysRequired
	}
	return Model{
		id: b.id, tenantID: b.tenantID, householdID: b.householdID, recipeID: b.recipeID,
		fullKey: b.fullKey, thumbKey: b.thumbKey, contentType: b.contentType,
		width: b.width, height: b.height, fullBytes: b.fullBytes, thumbBytes: b.thumbBytes,
		createdAt: b.createdAt, updatedAt: b.updatedAt,
	}, nil
}
```

`internal/recipeimage/keys.go`:
```go
package recipeimage

import "github.com/google/uuid"

// objectKeys builds the object keys for one image. Every input is a UUID, so
// user input can never reach a key.
func objectKeys(tenantID, householdID, recipeID, imageID uuid.UUID) (full, thumb string) {
	prefix := tenantID.String() + "/" + householdID.String() + "/recipes/" + recipeID.String() + "/" + imageID.String() + "/"
	return prefix + "full.jpg", prefix + "thumb.jpg"
}

// URLs returns the public, versioned image URLs. The /api/v1 prefix is the
// public path under both compose nginx and the k8s ingress. v changes on
// every replace, which busts browser caches of the immutable responses.
func URLs(recipeID, imageID uuid.UUID) (thumb, full string) {
	base := "/api/v1/recipes/" + recipeID.String() + "/image?size="
	v := "&v=" + imageID.String()
	return base + string(SizeThumb) + v, base + string(SizeFull) + v
}
```

`internal/recipeimage/errors.go`:
```go
package recipeimage

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrRecipeNotFound is returned by RecipeVisibility when the recipe does
	// not exist, is soft-deleted, or belongs to another household.
	ErrRecipeNotFound = errors.New("recipe not found")
	ErrImageNotFound  = errors.New("recipe image not found")
	// ErrObjectMissing means a DB row points at an object that is gone.
	ErrObjectMissing = errors.New("recipe image object missing")
	ErrStorage       = errors.New("object storage unavailable")
	ErrNoTenant      = errors.New("tenant missing from context")
)

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value") || // postgres 23505
		strings.Contains(msg, "UNIQUE constraint failed") // sqlite
}
```

`internal/recipeimage/provider.go`:
```go
package recipeimage

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
)

// Tenant scoping comes from the GORM tenant callback; household is filtered
// explicitly on every query.

func getByRecipeID(recipeID, householdID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("recipe_id = ? AND household_id = ?", recipeID, householdID)
	})
}

func getByRecipeIDs(recipeIDs []uuid.UUID, householdID uuid.UUID) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		if len(recipeIDs) == 0 {
			return []Entity{}, nil
		}
		var entities []Entity
		err := db.Where("recipe_id IN (?) AND household_id = ?", recipeIDs, householdID).Find(&entities).Error
		return entities, err
	}
}

func getPendingDeletions(limit int) func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
	return func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
		var rows []ObjectDeletionEntity
		err := db.Order("created_at ASC").Limit(limit).Find(&rows).Error
		return rows, err
	}
}

func getDeletionsByIDs(ids []uuid.UUID) func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
	return func(db *gorm.DB) ([]ObjectDeletionEntity, error) {
		if len(ids) == 0 {
			return []ObjectDeletionEntity{}, nil
		}
		var rows []ObjectDeletionEntity
		err := db.Where("id IN (?)", ids).Find(&rows).Error
		return rows, err
	}
}
```

`internal/recipeimage/administrator.go`:
```go
package recipeimage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// deleteByRecipeIDReturning removes the recipe's image row (if any) and
// returns it. In Postgres a concurrent second DELETE blocks on the row lock
// and then sees the winner's freshly inserted row, so no keys are lost.
func deleteByRecipeIDReturning(tx *gorm.DB, recipeID, householdID uuid.UUID) ([]Entity, error) {
	var removed []Entity
	err := tx.Clauses(clause.Returning{}).
		Where("recipe_id = ? AND household_id = ?", recipeID, householdID).
		Delete(&removed).Error
	return removed, err
}

func createImage(tx *gorm.DB, e *Entity) error {
	return tx.Create(e).Error
}

// enqueueObjectDeletions writes one outbox row per object key of the removed
// images and returns the outbox row ids.
func enqueueObjectDeletions(tx *gorm.DB, removed []Entity) ([]uuid.UUID, error) {
	if len(removed) == 0 {
		return nil, nil
	}
	now := time.Now().UTC()
	rows := make([]ObjectDeletionEntity, 0, len(removed)*2)
	for _, e := range removed {
		for _, key := range []string{e.FullKey, e.ThumbKey} {
			rows = append(rows, ObjectDeletionEntity{
				Id: uuid.New(), TenantId: e.TenantId, HouseholdId: e.HouseholdId,
				ObjectKey: key, CreatedAt: now,
			})
		}
	}
	if err := tx.Create(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.Id
	}
	return ids, nil
}

// PurgeByRecipeIDsTx deletes the image rows of hard-deleted recipes inside
// the caller's transaction and enqueues their objects for deletion. The
// caller (retention) has already scoped ids to one tenant + household. It
// returns the number of image rows deleted.
func PurgeByRecipeIDsTx(tx *gorm.DB, recipeIDs []uuid.UUID) (int, error) {
	if len(recipeIDs) == 0 {
		return 0, nil
	}
	var removed []Entity
	if err := tx.Clauses(clause.Returning{}).Where("recipe_id IN ?", recipeIDs).Delete(&removed).Error; err != nil {
		return 0, err
	}
	if _, err := enqueueObjectDeletions(tx, removed); err != nil {
		return 0, err
	}
	return len(removed), nil
}

func deleteOutboxRow(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&ObjectDeletionEntity{}).Error
}

func recordOutboxFailure(db *gorm.DB, id uuid.UUID, msg string) error {
	return db.Model(&ObjectDeletionEntity{}).Where("id = ?", id).
		Updates(map[string]interface{}{"attempts": gorm.Expr("attempts + 1"), "last_error": msg}).Error
}
```

- [ ] **Step 5: Run the tests**

Run: `cd services/recipe-service && go build ./... && go test ./internal/recipeimage/`
Expected: PASS. If `clause.Returning{}` returns no rows under SQLite, confirm `go list -m github.com/mattn/go-sqlite3` is ≥ v1.14.22 (it bundles SQLite ≥ 3.35, which supports `RETURNING`) before changing the approach.

- [ ] **Step 6: Commit**

```bash
git add services/recipe-service/internal/recipeimage
git commit -m "feat(recipe-service): add recipeimage domain entities, outbox and queries"
```

---

## Task 5: `recipeimage` processor, metrics and outbox drainer

**Files:**
- Create: `services/recipe-service/internal/recipeimage/metrics.go`
- Create: `services/recipe-service/internal/recipeimage/processor.go`
- Create: `services/recipe-service/internal/recipeimage/reaper.go`
- Create: `services/recipe-service/internal/recipeimage/processor_test.go`
- Create: `services/recipe-service/internal/recipeimage/reaper_test.go`
- Modify: `services/recipe-service/go.mod`, `go.sum` (prometheus becomes direct)

**Interfaces:**
- Consumes: Task 1 `objectstore.Store`, `objectstoretest`; Task 3 `imaging.Process` and its errors; Task 4 domain functions.
- Produces:
  - `recipeimage.NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor`
  - `(*Processor).WithStore(s objectstore.Store) *Processor`. This returns a copy; it is required before calling `Upload`, `Delete` or `Open`.
  - `(*Processor).Upload(recipeID uuid.UUID, src []byte) (UploadResult, error)` where `UploadResult{Model Model; Replaced bool; OutboxIDs []uuid.UUID; SourceBytes int; ProcessingTime time.Duration}`
  - `(*Processor).Delete(recipeID uuid.UUID) ([]uuid.UUID, error)` (returns outbox ids; idempotent)
  - `(*Processor).GetByRecipeID(recipeID uuid.UUID) (Model, error)` (→ `ErrImageNotFound`)
  - `(*Processor).GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error)`
  - `(*Processor).Open(m Model, size Size) (objectstore.Object, error)` (→ `ErrObjectMissing` | `ErrStorage`)
  - `recipeimage.NewObjectReaper(l logrus.FieldLogger, db *gorm.DB, store objectstore.Store) *ObjectReaper`; `(*ObjectReaper).Run(ctx context.Context)`; `(*ObjectReaper).DrainKeys(ctx context.Context, ids []uuid.UUID)`
  - `recipeimage.RecordStorageError(op string)`; unexported `recordUpload(result string)` and `processingSeconds`
  - package var `beforeInsertHook func(tx *gorm.DB, e *Entity) error` (a test seam; nil in production)

- [ ] **Step 1: Write the failing processor tests**

`internal/recipeimage/processor_test.go`:
```go
package recipeimage

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore/objectstoretest"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage/imaging"
)

func testJPEG(t testing.TB, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type fixture struct {
	db                     *gorm.DB
	store                  *objectstoretest.Store
	tenantID, householdID  uuid.UUID
	proc                   *Processor
}

func newFixture(t *testing.T) fixture {
	db := newTestDB(t)
	store := objectstoretest.New()
	tid, hid := uuid.New(), uuid.New()
	l, _ := test.NewNullLogger()
	return fixture{
		db: db, store: store, tenantID: tid, householdID: hid,
		proc: NewProcessor(l, tenantCtx(tid, hid), db).WithStore(store),
	}
}

func countRows[T any](t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(new(T)).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestUploadCreate(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	res, err := f.proc.Upload(rid, testJPEG(t, 2000, 1000))
	if err != nil {
		t.Fatal(err)
	}
	if res.Replaced || len(res.OutboxIDs) != 0 {
		t.Fatalf("first upload: replaced=%v outbox=%v", res.Replaced, res.OutboxIDs)
	}
	m := res.Model
	if m.Width() != 1600 || m.Height() != 800 || m.ContentType() != "image/jpeg" {
		t.Fatalf("model dims %dx%d ct=%s", m.Width(), m.Height(), m.ContentType())
	}
	if !f.store.Has(m.FullKey()) || !f.store.Has(m.ThumbKey()) {
		t.Fatalf("objects not stored: %v", f.store.Keys())
	}
	if countRows[Entity](t, f.db) != 1 {
		t.Fatal("want one image row")
	}
}

func TestUploadReplaceEnqueuesOldKeysAndKeepsCreatedAt(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	first, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	second, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	if !second.Replaced || len(second.OutboxIDs) != 2 {
		t.Fatalf("replace: replaced=%v outbox=%d", second.Replaced, len(second.OutboxIDs))
	}
	if second.Model.Id() == first.Model.Id() {
		t.Fatal("replace must mint a new image id")
	}
	tu1, fu1 := URLs(rid, first.Model.Id())
	tu2, fu2 := URLs(rid, second.Model.Id())
	if tu1 == tu2 || fu1 == fu2 {
		t.Fatal("replace must change the advertised URLs (cache busting)")
	}
	if !second.Model.CreatedAt().Equal(first.Model.CreatedAt()) {
		t.Fatalf("createdAt not carried forward: %v vs %v", second.Model.CreatedAt(), first.Model.CreatedAt())
	}
	if countRows[Entity](t, f.db) != 1 || countRows[ObjectDeletionEntity](t, f.db) != 2 {
		t.Fatal("want 1 image row and 2 outbox rows")
	}
}

func TestUploadRejectsBadInputWithoutSideEffects(t *testing.T) {
	f := newFixture(t)
	_, err := f.proc.Upload(uuid.New(), []byte("nope"))
	if !errors.Is(err, imaging.ErrUnsupportedFormat) {
		t.Fatalf("err = %v", err)
	}
	if f.store.PutCalls != 0 || countRows[Entity](t, f.db) != 0 {
		t.Fatal("rejected upload must not touch storage or DB")
	}
}

func TestUploadPutFailureCleansUp(t *testing.T) {
	f := newFixture(t)
	f.store.PutHook = func(key string) error {
		if bytes.HasSuffix([]byte(key), []byte("thumb.jpg")) {
			return errors.New("minio down")
		}
		return nil
	}
	_, err := f.proc.Upload(uuid.New(), testJPEG(t, 64, 64))
	if !errors.Is(err, ErrStorage) {
		t.Fatalf("err = %v, want ErrStorage", err)
	}
	if len(f.store.Keys()) != 0 {
		t.Fatalf("full rendition not cleaned up: %v", f.store.Keys())
	}
	if countRows[Entity](t, f.db) != 0 {
		t.Fatal("no row may be written on storage failure")
	}
}

func TestUploadDBFailureCleansUpNewObjects(t *testing.T) {
	f := newFixture(t)
	beforeInsertHook = func(*gorm.DB, *Entity) error { return errors.New("db exploded") }
	t.Cleanup(func() { beforeInsertHook = nil })

	_, err := f.proc.Upload(uuid.New(), testJPEG(t, 64, 64))
	if err == nil || errors.Is(err, ErrStorage) {
		t.Fatalf("err = %v, want a DB error", err)
	}
	if len(f.store.Keys()) != 0 {
		t.Fatalf("new objects not cleaned up: %v", f.store.Keys())
	}
}

func TestUploadRetriesOnceOnUniqueViolation(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	calls := 0
	beforeInsertHook = func(tx *gorm.DB, e *Entity) error {
		calls++
		if calls == 1 {
			// Simulate a concurrent first upload committing between our
			// DELETE and INSERT.
			competing := *e
			competing.Id = uuid.New()
			return tx.Create(&competing).Error
		}
		return nil
	}
	t.Cleanup(func() { beforeInsertHook = nil })

	res, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("hook calls = %d, want 2 (one retry)", calls)
	}
	var rows []Entity
	f.db.Find(&rows)
	if len(rows) != 1 || rows[0].Id != res.Model.Id() {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestDeleteIsIdempotentAndEnqueues(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	if _, err := f.proc.Upload(rid, testJPEG(t, 64, 64)); err != nil {
		t.Fatal(err)
	}
	ids, err := f.proc.Delete(rid)
	if err != nil || len(ids) != 2 {
		t.Fatalf("delete ids=%v err=%v", ids, err)
	}
	ids, err = f.proc.Delete(rid)
	if err != nil || len(ids) != 0 {
		t.Fatalf("second delete ids=%v err=%v", ids, err)
	}
	if _, err := f.proc.GetByRecipeID(rid); !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("get after delete err = %v", err)
	}
}

func TestGetIsHouseholdScoped(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	if _, err := f.proc.Upload(rid, testJPEG(t, 64, 64)); err != nil {
		t.Fatal(err)
	}
	l, _ := test.NewNullLogger()
	other := NewProcessor(l, tenantCtx(f.tenantID, uuid.New()), f.db)
	if _, err := other.GetByRecipeID(rid); !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("other household err = %v", err)
	}
	m, err := other.GetByRecipeIDs([]uuid.UUID{rid})
	if err != nil || len(m) != 0 {
		t.Fatalf("other household batch = %v err=%v", m, err)
	}
	own, err := f.proc.GetByRecipeIDs([]uuid.UUID{rid, uuid.New()})
	if err != nil || len(own) != 1 {
		t.Fatalf("own batch = %v err=%v", own, err)
	}
}

func TestOpenMapsErrors(t *testing.T) {
	f := newFixture(t)
	res, err := f.proc.Upload(uuid.New(), testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	obj, err := f.proc.Open(res.Model, SizeThumb)
	if err != nil {
		t.Fatal(err)
	}
	obj.Body.Close()

	f.store.GetErr = errors.New("down")
	if _, err := f.proc.Open(res.Model, SizeFull); !errors.Is(err, ErrStorage) {
		t.Fatalf("err = %v, want ErrStorage", err)
	}
	f.store.GetErr = nil
	_ = f.store.Delete(context.Background(), res.Model.FullKey())
	if _, err := f.proc.Open(res.Model, SizeFull); !errors.Is(err, ErrObjectMissing) {
		t.Fatalf("err = %v, want ErrObjectMissing", err)
	}
}

func TestUploadWithoutTenantFails(t *testing.T) {
	db := newTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, t.Context(), db).WithStore(objectstoretest.New())
	if _, err := p.Upload(uuid.New(), testJPEG(t, 64, 64)); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("err = %v", err)
	}
}
```

- [ ] **Step 2: Write the failing reaper tests**

`internal/recipeimage/reaper_test.go`:
```go
package recipeimage

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

func TestReaperDrainsOutbox(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	first, _ := f.proc.Upload(rid, testJPEG(t, 64, 64))
	second, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	l, _ := test.NewNullLogger()
	r := NewObjectReaper(l, f.db, f.store)

	r.DrainKeys(context.Background(), second.OutboxIDs)

	if f.store.Has(first.Model.FullKey()) || f.store.Has(first.Model.ThumbKey()) {
		t.Fatal("old objects not deleted")
	}
	if !f.store.Has(second.Model.FullKey()) {
		t.Fatal("current object deleted")
	}
	if countRows[ObjectDeletionEntity](t, f.db) != 0 {
		t.Fatal("outbox rows not removed")
	}
}

func TestReaperFailureIncrementsAttempts(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	_, _ = f.proc.Upload(rid, testJPEG(t, 64, 64))
	if _, err := f.proc.Delete(rid); err != nil {
		t.Fatal(err)
	}
	f.store.DeleteErr = errors.New("down")
	l, _ := test.NewNullLogger()
	r := NewObjectReaper(l, f.db, f.store)

	r.Run(context.Background())

	var rows []ObjectDeletionEntity
	f.db.Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	for _, row := range rows {
		if row.Attempts != 1 || row.LastError == nil || *row.LastError == "" {
			t.Fatalf("row not marked failed: %+v", row)
		}
	}

	f.store.DeleteErr = nil
	r.Run(context.Background())
	if countRows[ObjectDeletionEntity](t, f.db) != 0 {
		t.Fatal("retry did not drain")
	}
}

func TestReaperTreatsNotFoundAsSuccess(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	_, _ = f.proc.Upload(rid, testJPEG(t, 64, 64))
	_, _ = f.proc.Delete(rid)
	f.store.DeleteErr = objectstore.ErrNotFound
	l, _ := test.NewNullLogger()
	NewObjectReaper(l, f.db, f.store).Run(context.Background())
	if countRows[ObjectDeletionEntity](t, f.db) != 0 {
		t.Fatal("ErrNotFound must count as success")
	}
}
```

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/recipeimage/`
Expected: FAIL (undefined `NewProcessor`, `NewObjectReaper`, …).

- [ ] **Step 4: Implement `metrics.go`**

```go
package recipeimage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collectors live on the default registry, which recipe-service already
// exports at /metrics.
var (
	uploadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "recipe_image_uploads_total",
		Help: "Recipe image uploads by result.",
	}, []string{"result"})

	processingSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "recipe_image_processing_seconds",
		Help:    "Time spent decoding, resizing and encoding one upload.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	})

	storageErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "recipe_image_storage_errors_total",
		Help: "Object storage failures by operation.",
	}, []string{"op"})
)

func recordUpload(result string) { uploadsTotal.WithLabelValues(result).Inc() }

// RecordStorageError increments recipe_image_storage_errors_total{op}.
func RecordStorageError(op string) { storageErrorsTotal.WithLabelValues(op).Inc() }
```

- [ ] **Step 5: Implement `processor.go`**

```go
package recipeimage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage/imaging"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

const (
	contentTypeJPEG = "image/jpeg"
	putTimeout      = 30 * time.Second
	cleanupTimeout  = 10 * time.Second
)

// processSem bounds concurrent decodes: a 40 MP PNG decodes to ~160 MB.
var processSem = make(chan struct{}, 2)

// beforeInsertHook is a test seam that runs inside the swap transaction
// between the DELETE and the INSERT. Always nil in production.
var beforeInsertHook func(tx *gorm.DB, e *Entity) error

type Processor struct {
	l     logrus.FieldLogger
	ctx   context.Context
	db    *gorm.DB
	store objectstore.Store
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// WithStore returns a copy bound to an object store. Read-only callers
// (recipe enrichment) do not need one.
func (p *Processor) WithStore(s objectstore.Store) *Processor {
	c := *p
	c.store = s
	return &c
}

type UploadResult struct {
	Model          Model
	Replaced       bool
	OutboxIDs      []uuid.UUID
	SourceBytes    int
	ProcessingTime time.Duration
}

func (p *Processor) Upload(recipeID uuid.UUID, src []byte) (UploadResult, error) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return UploadResult{}, ErrNoTenant
	}

	start := time.Now()
	res, err := p.process(src)
	elapsed := time.Since(start)
	if err != nil {
		return UploadResult{}, err
	}
	processingSeconds.Observe(elapsed.Seconds())

	imageID := uuid.New()
	fullKey, thumbKey := objectKeys(t.Id(), t.HouseholdId(), recipeID, imageID)
	log := p.l.WithFields(logrus.Fields{"recipe_id": recipeID, "image_id": imageID})
	// Logged before Put so an operator can find orphans from a crash
	// between Put and commit (design §9).
	log.Info("storing recipe image renditions")

	storeStart := time.Now()
	if err := p.put(fullKey, res.Full.Bytes); err != nil {
		return UploadResult{}, err
	}
	if err := p.put(thumbKey, res.Thumb.Bytes); err != nil {
		p.deleteNow(fullKey)
		return UploadResult{}, err
	}
	storageElapsed := time.Since(storeStart)

	now := time.Now().UTC()
	e := Entity{
		Id: imageID, TenantId: t.Id(), HouseholdId: t.HouseholdId(), RecipeId: recipeID,
		FullKey: fullKey, ThumbKey: thumbKey, ContentType: contentTypeJPEG,
		Width: res.Full.Width, Height: res.Full.Height,
		FullBytes: len(res.Full.Bytes), ThumbBytes: len(res.Thumb.Bytes),
		CreatedAt: now, UpdatedAt: now,
	}
	replaced, outboxIDs, err := p.swap(&e)
	if err != nil {
		p.deleteNow(fullKey)
		p.deleteNow(thumbKey)
		return UploadResult{}, err
	}

	m, err := Make(e)
	if err != nil {
		return UploadResult{}, err
	}
	log.WithFields(logrus.Fields{
		"source_bytes": len(src), "full_bytes": e.FullBytes, "thumb_bytes": e.ThumbBytes,
		"processing_ms": elapsed.Milliseconds(), "storage_ms": storageElapsed.Milliseconds(),
		"replaced": replaced,
	}).Info("recipe image stored")
	return UploadResult{Model: m, Replaced: replaced, OutboxIDs: outboxIDs, SourceBytes: len(src), ProcessingTime: elapsed}, nil
}

func (p *Processor) process(src []byte) (imaging.Result, error) {
	select {
	case processSem <- struct{}{}:
	case <-p.ctx.Done():
		return imaging.Result{}, p.ctx.Err()
	}
	defer func() { <-processSem }()
	return imaging.Process(src)
}

func (p *Processor) put(key string, body []byte) error {
	ctx, cancel := context.WithTimeout(p.ctx, putTimeout)
	defer cancel()
	if err := p.store.Put(ctx, key, body, contentTypeJPEG); err != nil {
		RecordStorageError("put")
		p.l.WithError(err).WithField("object_key", key).Error("failed to store recipe image object")
		return fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return nil
}

// deleteNow is a best-effort, direct (non-outbox) delete used only to clean
// up objects that no committed row references.
func (p *Processor) deleteNow(key string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(p.ctx), cleanupTimeout)
	defer cancel()
	if err := p.store.Delete(ctx, key); err != nil && !errors.Is(err, objectstore.ErrNotFound) {
		RecordStorageError("delete")
		p.l.WithError(err).WithField("object_key", key).Warn("failed to clean up recipe image object")
	}
}

func (p *Processor) swap(e *Entity) (bool, []uuid.UUID, error) {
	replaced, ids, err := p.swapTx(e)
	if err != nil && isUniqueViolation(err) {
		p.l.WithField("recipe_id", e.RecipeId).Warn("concurrent first image upload; retrying swap once")
		replaced, ids, err = p.swapTx(e)
	}
	return replaced, ids, err
}

func (p *Processor) swapTx(e *Entity) (replaced bool, outboxIDs []uuid.UUID, err error) {
	err = p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		e.CreatedAt = e.UpdatedAt
		removed, err := deleteByRecipeIDReturning(tx, e.RecipeId, e.HouseholdId)
		if err != nil {
			return err
		}
		if len(removed) > 0 {
			replaced = true
			e.CreatedAt = removed[0].CreatedAt
		}
		if beforeInsertHook != nil {
			if err := beforeInsertHook(tx, e); err != nil {
				return err
			}
		}
		if err := createImage(tx, e); err != nil {
			return err
		}
		outboxIDs, err = enqueueObjectDeletions(tx, removed)
		return err
	})
	if err != nil {
		return false, nil, err
	}
	return replaced, outboxIDs, nil
}

// Delete removes the recipe's image row and enqueues its objects. Deleting a
// recipe with no image succeeds with no ids.
func (p *Processor) Delete(recipeID uuid.UUID) ([]uuid.UUID, error) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	var ids []uuid.UUID
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		removed, err := deleteByRecipeIDReturning(tx, recipeID, t.HouseholdId())
		if err != nil {
			return err
		}
		ids, err = enqueueObjectDeletions(tx, removed)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		p.l.WithField("recipe_id", recipeID).Info("recipe image deleted")
	}
	return ids, nil
}

func (p *Processor) GetByRecipeID(recipeID uuid.UUID) (Model, error) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return Model{}, ErrNoTenant
	}
	e, err := getByRecipeID(recipeID, t.HouseholdId())(p.db.WithContext(p.ctx))()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Model{}, ErrImageNotFound
	}
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// GetByRecipeIDs returns images keyed by recipe id in a single query. Empty
// input returns an empty map without hitting the database.
func (p *Processor) GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error) {
	result := make(map[uuid.UUID]Model)
	if len(recipeIDs) == 0 {
		return result, nil
	}
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	entities, err := getByRecipeIDs(recipeIDs, t.HouseholdId())(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		result[e.RecipeId] = m
	}
	return result, nil
}

// Open streams one rendition. The caller must close the body.
func (p *Processor) Open(m Model, size Size) (objectstore.Object, error) {
	obj, err := p.store.Get(p.ctx, m.Key(size))
	if errors.Is(err, objectstore.ErrNotFound) {
		p.l.WithFields(logrus.Fields{"recipe_id": m.RecipeID(), "image_id": m.Id(), "object_key": m.Key(size)}).
			Error("recipe image row points at a missing object")
		return objectstore.Object{}, ErrObjectMissing
	}
	if err != nil {
		RecordStorageError("get")
		return objectstore.Object{}, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	return obj, nil
}
```

- [ ] **Step 6: Implement `reaper.go`**

```go
package recipeimage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

const (
	drainBatch   = 100
	drainTimeout = 10 * time.Second
)

// ObjectReaper drains recipe_image_object_deletions. It runs with a
// tenant-less context so the tenant callback does not filter the outbox.
// Concurrent drainers are safe: a repeated object delete is a no-op or
// ErrNotFound (success), and a repeated row delete affects nothing.
type ObjectReaper struct {
	l     logrus.FieldLogger
	db    *gorm.DB
	store objectstore.Store
}

func NewObjectReaper(l logrus.FieldLogger, db *gorm.DB, store objectstore.Store) *ObjectReaper {
	return &ObjectReaper{l: l, db: db, store: store}
}

// Run drains up to one batch, oldest first. Its signature matches the
// shared retention Loop callback.
func (r *ObjectReaper) Run(ctx context.Context) {
	rows, err := getPendingDeletions(drainBatch)(r.db.WithContext(ctx))
	if err != nil {
		r.l.WithError(err).Error("failed to load recipe image deletion outbox")
		return
	}
	r.drain(ctx, rows)
}

// DrainKeys attempts the deletes for specific outbox rows right after a
// commit. Failures stay in the outbox for Run to retry.
func (r *ObjectReaper) DrainKeys(ctx context.Context, ids []uuid.UUID) {
	if len(ids) == 0 {
		return
	}
	rows, err := getDeletionsByIDs(ids)(r.db.WithContext(ctx))
	if err != nil {
		r.l.WithError(err).Error("failed to load recipe image deletion outbox rows")
		return
	}
	r.drain(ctx, rows)
}

func (r *ObjectReaper) drain(ctx context.Context, rows []ObjectDeletionEntity) {
	for _, row := range rows {
		dctx, cancel := context.WithTimeout(ctx, drainTimeout)
		err := r.store.Delete(dctx, row.ObjectKey)
		cancel()
		if err != nil && !errors.Is(err, objectstore.ErrNotFound) {
			RecordStorageError("delete")
			r.l.WithError(err).WithFields(logrus.Fields{
				"object_key": row.ObjectKey, "attempts": row.Attempts + 1,
			}).Warn("recipe image object delete failed; will retry")
			if uerr := recordOutboxFailure(r.db.WithContext(ctx), row.Id, err.Error()); uerr != nil {
				r.l.WithError(uerr).Error("failed to record outbox failure")
			}
			continue
		}
		if derr := deleteOutboxRow(r.db.WithContext(ctx), row.Id); derr != nil {
			r.l.WithError(derr).WithField("object_key", row.ObjectKey).Error("failed to remove drained outbox row")
		}
	}
}
```

- [ ] **Step 7: Run the tests**

Run: `cd services/recipe-service && go mod tidy && go build ./... && go test ./internal/recipeimage/...`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add services/recipe-service/internal/recipeimage services/recipe-service/go.mod services/recipe-service/go.sum
git commit -m "feat(recipe-service): add recipe image processor, metrics and outbox drainer"
```

---

## Task 6: Image HTTP handlers, visibility, routes and `main.go` wiring

**Files:**
- Create: `services/recipe-service/internal/recipeimage/rest.go`
- Create: `services/recipe-service/internal/recipeimage/resource.go`
- Create: `services/recipe-service/internal/recipeimage/resource_test.go`
- Create: `services/recipe-service/internal/recipe/visibility.go`
- Create: `services/recipe-service/internal/recipe/visibility_test.go`
- Modify: `services/recipe-service/cmd/main.go`

**Interfaces:**
- Consumes: Task 1 (`objectstore.New`, `config.ObjectStore`), Task 5 (`Processor`, `ObjectReaper`, metrics).
- Produces:
  - `recipeimage.RecipeVisibility func(ctx context.Context, recipeID uuid.UUID) error` (nil | `ErrRecipeNotFound` | other)
  - `recipeimage.InitializeRoutes(db *gorm.DB, store objectstore.Store, visible RecipeVisibility, reaper *ObjectReaper) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router)`
  - `recipeimage.RestModel` (`recipe-images`) and `recipeimage.Transform(m Model) RestModel`
  - `recipe.ImageVisibility(db *gorm.DB) recipeimage.RecipeVisibility`

- [ ] **Step 1: Write the failing visibility test**

`internal/recipe/visibility_test.go`:
```go
package recipe

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

func TestImageVisibility(t *testing.T) {
	db := setupTestDB(t)
	tid, hidA, hidB := uuid.New(), uuid.New(), uuid.New()
	ctxA := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidA, uuid.New()))
	ctxB := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidB, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctxA, db)
	m, _, err := p.Create(tid, hidA, CreateAttrs{Title: "Soup", Source: "Boil water."})
	if err != nil {
		t.Fatal(err)
	}
	visible := ImageVisibility(db)

	if err := visible(ctxA, m.Id()); err != nil {
		t.Fatalf("own household: %v", err)
	}
	if err := visible(ctxB, m.Id()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("other household err = %v", err)
	}
	if err := visible(ctxA, uuid.New()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("missing err = %v", err)
	}
	if err := visible(context.Background(), m.Id()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("no tenant err = %v", err)
	}
	if err := p.Delete(m.Id()); err != nil {
		t.Fatal(err)
	}
	if err := visible(ctxA, m.Id()); !errors.Is(err, recipeimage.ErrRecipeNotFound) {
		t.Fatalf("soft-deleted err = %v", err)
	}
}
```

- [ ] **Step 2: Write the failing handler tests**

`internal/recipeimage/resource_test.go`:
```go
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
```

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/recipeimage/ ./internal/recipe/`
Expected: FAIL (undefined `InitializeRoutes`, `ImageVisibility`, …).

- [ ] **Step 4: Implement `rest.go`**

```go
package recipeimage

import (
	"time"

	"github.com/google/uuid"
)

// RestModel is the JSON:API representation of a recipe image.
type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	ThumbURL  string    `json:"thumbUrl"`
	FullURL   string    `json:"fullUrl"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "recipe-images" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) RestModel {
	thumb, full := URLs(m.RecipeID(), m.Id())
	return RestModel{
		Id: m.Id(), Width: m.Width(), Height: m.Height(),
		ThumbURL: thumb, FullURL: full, UpdatedAt: m.UpdatedAt(),
	}
}
```

- [ ] **Step 5: Implement `resource.go`**

```go
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
			if isMaxBytes(err) {
				return nil, http.StatusRequestEntityTooLarge, "Image must be 10 MB or smaller"
			}
			return nil, http.StatusBadRequest, "Malformed multipart body"
		}
		if part.FormName() != "file" {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, maxFileBytes+1))
		if err != nil {
			if isMaxBytes(err) {
				return nil, http.StatusRequestEntityTooLarge, "Image must be 10 MB or smaller"
			}
			return nil, http.StatusBadRequest, "Malformed multipart body"
		}
		if len(data) > maxFileBytes {
			return nil, http.StatusRequestEntityTooLarge, "Image must be 10 MB or smaller"
		}
		return data, 0, ""
	}
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
		// Client went away while queued for processing; nothing to write.
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
```

- [ ] **Step 6: Implement `recipe/visibility.go`**

```go
package recipe

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

// ImageVisibility lets recipeimage check a recipe without importing this
// package: the recipe must exist, not be soft-deleted, and belong to the
// caller's household. Any miss is ErrRecipeNotFound (404, never 403).
func ImageVisibility(db *gorm.DB) recipeimage.RecipeVisibility {
	return func(ctx context.Context, recipeID uuid.UUID) error {
		t, ok := tenantctx.FromContext(ctx)
		if !ok {
			return recipeimage.ErrRecipeNotFound
		}
		e, err := getByID(recipeID)(db.WithContext(ctx))()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return recipeimage.ErrRecipeNotFound
		}
		if err != nil {
			return err
		}
		if e.HouseholdId != t.HouseholdId() {
			return recipeimage.ErrRecipeNotFound
		}
		return nil
	}
}
```

- [ ] **Step 7: Wire `cmd/main.go`**

Add the imports `"time"`, `".../internal/objectstore"`, `".../internal/recipeimage"`, and `sr "github.com/jtumidanski/home-hub/shared/go/retention"`. Add `recipeimage.Migration,` to `database.SetMigrations(...)` after `planitem.Migration`. After `catClient := …` and `ctx, cancel := …`, add:
```go
	store := objectstore.New(l, cfg.ObjectStore)
	go func() {
		pctx, pcancel := context.WithTimeout(ctx, 5*time.Second)
		defer pcancel()
		if err := store.Ping(pctx); err != nil {
			recipeimage.RecordStorageError("ping")
			l.WithError(err).Error("object storage unreachable at startup; image endpoints will return 503 until it recovers")
		}
	}()
	imageReaper := recipeimage.NewObjectReaper(l, db, store)
	go sr.Loop(ctx, l, 5*time.Minute, imageReaper.Run)
```
Inside the route initializer, after the `planitem.InitializeRoutes(...)` line, add:
```go
			recipeimage.InitializeRoutes(db, store, recipe.ImageVisibility(db), imageReaper)(l, si, api)
```

- [ ] **Step 8: Run the tests**

Run: `cd services/recipe-service && go build ./... && go vet ./... && go test ./internal/recipeimage/... ./internal/recipe/...`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add services/recipe-service/internal/recipeimage services/recipe-service/internal/recipe/visibility.go services/recipe-service/internal/recipe/visibility_test.go services/recipe-service/cmd/main.go
git commit -m "feat(recipe-service): add recipe image upload, serve and delete endpoints"
```

---

## Task 7: `image` attribute on recipe list and detail responses

**Files:**
- Modify: `services/recipe-service/internal/recipe/rest.go`
- Modify: `services/recipe-service/internal/recipe/processor.go:378-440` (`BuildDetailEnrichment`, `BuildListEnrichments`)
- Modify: `services/recipe-service/internal/recipe/processor_integration_test.go` (`setupTestDB` migrations)
- Modify: `services/recipe-service/internal/recipe/batch_test.go`
- Create: `services/recipe-service/internal/recipe/image_enrichment_test.go`

**Interfaces:**
- Consumes: Task 5 `recipeimage.NewProcessor(...).GetByRecipeID/GetByRecipeIDs`, `recipeimage.ErrImageNotFound`; Task 4 `recipeimage.URLs`, `recipeimage.Entity`, `recipeimage.ObjectDeletionEntity`.
- Produces: `recipe.RestImageModel{ThumbURL, FullURL string; Width, Height int}`; `RestModel.Image` and `RestDetailModel.Image *RestImageModel` (`json:"image"`, no omitempty); `ListEnrichment.Image`, `DetailEnrichment.Image`.

- [ ] **Step 1: Migrate the image tables in the recipe test DB**

In `processor_integration_test.go` `setupTestDB`, change the AutoMigrate call to:
```go
	if err := db.AutoMigrate(&Entity{}, &TagEntity{}, &RestorationEntity{}, &audit.Entity{}, &recipeimage.Entity{}, &recipeimage.ObjectDeletionEntity{}); err != nil {
```
and add the import `"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"`.

- [ ] **Step 2: Write the failing tests**

`internal/recipe/image_enrichment_test.go`:
```go
package recipe

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe/cooklang"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

func seedRecipeImage(t *testing.T, db *gorm.DB, tenantID, householdID, recipeID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	e := recipeimage.Entity{
		Id: id, TenantId: tenantID, HouseholdId: householdID, RecipeId: recipeID,
		FullKey: "f/" + id.String(), ThumbKey: "t/" + id.String(), ContentType: "image/jpeg",
		Width: 1600, Height: 1067, FullBytes: 1, ThumbBytes: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func TestBuildListEnrichmentsImageOneQuery(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hid, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)

	var models []Model
	for i := 0; i < 5; i++ {
		m, _, err := p.Create(tid, hid, CreateAttrs{Title: "R", Source: "Boil."})
		if err != nil {
			t.Fatal(err)
		}
		models = append(models, m)
	}
	imageID := seedRecipeImage(t, db, tid, hid, models[0].Id())

	imageQueries := 0
	if err := db.Callback().Query().After("gorm:query").Register("test:count_recipe_images", func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "recipe_images") {
			imageQueries++
		}
	}); err != nil {
		t.Fatal(err)
	}

	enrichments := p.BuildListEnrichments(models)

	if imageQueries != 1 {
		t.Fatalf("recipe_images queries = %d, want 1", imageQueries)
	}
	img := enrichments[0].Image
	if img == nil || img.Width != 1600 || img.Height != 1067 ||
		!strings.HasSuffix(img.FullURL, "size=full&v="+imageID.String()) ||
		!strings.HasSuffix(img.ThumbURL, "size=thumb&v="+imageID.String()) {
		t.Fatalf("image = %+v", img)
	}
	for i := 1; i < len(enrichments); i++ {
		if enrichments[i].Image != nil {
			t.Fatalf("recipe %d unexpectedly has an image", i)
		}
	}
}

func TestImageEnrichmentOtherHousehold(t *testing.T) {
	db := setupTestDB(t)
	tid, hidA, hidB := uuid.New(), uuid.New(), uuid.New()
	l, _ := test.NewNullLogger()
	ctxA := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidA, uuid.New()))
	ctxB := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidB, uuid.New()))
	m, _, err := NewProcessor(l, ctxA, db).Create(tid, hidA, CreateAttrs{Title: "R", Source: "Boil."})
	if err != nil {
		t.Fatal(err)
	}
	seedRecipeImage(t, db, tid, hidA, m.Id())

	pB := NewProcessor(l, ctxB, db)
	if e := pB.BuildListEnrichments([]Model{m}); e[0].Image != nil {
		t.Fatal("household B must not see household A's image in list")
	}
	if d := pB.BuildDetailEnrichment(m, nil); d.Image != nil {
		t.Fatal("household B must not see household A's image in detail")
	}
}

func TestDetailEnrichmentImage(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hid, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)
	m, _, err := p.Create(tid, hid, CreateAttrs{Title: "R", Source: "Boil."})
	if err != nil {
		t.Fatal(err)
	}
	if d := p.BuildDetailEnrichment(m, nil); d.Image != nil {
		t.Fatal("no image expected yet")
	}
	seedRecipeImage(t, db, tid, hid, m.Id())
	if d := p.BuildDetailEnrichment(m, nil); d.Image == nil || d.Image.Width != 1600 {
		t.Fatalf("detail image = %+v", d.Image)
	}
}

func TestImageSerializesAsNull(t *testing.T) {
	m, _ := NewBuilder().SetId(uuid.New()).SetTitle("T").SetSource("s").Build()
	list, _ := json.Marshal(Transform(m, ListEnrichment{}))
	if !strings.Contains(string(list), `"image":null`) {
		t.Fatalf("list json = %s", list)
	}
	detail, _ := json.Marshal(TransformDetail(m, cooklang.ParseResult{}, DetailEnrichment{}))
	if !strings.Contains(string(detail), `"image":null`) {
		t.Fatalf("detail json = %s", detail)
	}
}

func TestPatchWithImageAttributeDoesNotTouchImage(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hid, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)
	m, _, err := p.Create(tid, hid, CreateAttrs{Title: "R", Source: "Boil."})
	if err != nil {
		t.Fatal(err)
	}
	imageID := seedRecipeImage(t, db, tid, hid, m.Id())

	body := `{"data":{"type":"recipes","id":"` + m.Id().String() + `","attributes":{"title":"New","image":{"thumbUrl":"/evil","fullUrl":"/evil","width":1,"height":1}}}}`
	var req UpdateRequest
	if err := jsonapi.Unmarshal([]byte(body), &req); err == nil {
		title := req.Title
		if _, _, err := p.Update(m.Id(), UpdateAttrs{Title: &title}); err != nil {
			t.Fatal(err)
		}
	}
	var rows []recipeimage.Entity
	db.Find(&rows)
	if len(rows) != 1 || rows[0].Id != imageID || rows[0].FullKey != "f/"+imageID.String() {
		t.Fatalf("image row changed: %+v", rows)
	}
}
```

If `NewBuilder()` in the recipe package requires more fields to build, use the same builder calls as `builder_test.go` for a minimal valid model.

- [ ] **Step 3: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/recipe/ -run 'Image|PatchWithImage'`
Expected: FAIL (`enrichments[0].Image undefined`).

- [ ] **Step 4: Implement the REST changes in `recipe/rest.go`**

Add the type:
```go
// RestImageModel is the nullable `image` attribute on recipe resources.
type RestImageModel struct {
	ThumbURL string `json:"thumbUrl"`
	FullURL  string `json:"fullUrl"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}
```
Add `Image *RestImageModel `json:"image"`` to `RestModel` (after `UsageCount`) and to `RestDetailModel` (after `PlannerIssues`). Add `Image *RestImageModel` to `ListEnrichment` and `DetailEnrichment`. In `Transform` add `Image: enrichment.Image,`; in `TransformDetail` add `Image: enrichment.Image,`.

- [ ] **Step 5: Implement the enrichment in `recipe/processor.go`**

Add the import `"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"`. In `BuildDetailEnrichment`, before `return enrichment`:
```go
	imgProc := recipeimage.NewProcessor(p.l, p.ctx, p.db)
	if img, err := imgProc.GetByRecipeID(m.Id()); err == nil {
		enrichment.Image = toImageRestModel(img)
	} else if !errors.Is(err, recipeimage.ErrImageNotFound) {
		p.l.WithError(err).Error("Failed to fetch recipe image for detail enrichment")
	}
```
In `BuildListEnrichments`, change the doc comment's "exactly two extra queries" to "exactly three extra queries". After the planner batch fetch, add:
```go
	imgProc := recipeimage.NewProcessor(p.l, p.ctx, p.db)
	imagesByRecipe, err := imgProc.GetByRecipeIDs(recipeIDs)
	if err != nil {
		p.l.WithError(err).Error("Failed to batch-fetch recipe images for list enrichment")
		imagesByRecipe = map[uuid.UUID]recipeimage.Model{}
	}
```
and inside the loop, before `enrichments[i] = e`:
```go
		if img, ok := imagesByRecipe[m.Id()]; ok {
			e.Image = toImageRestModel(img)
		}
```
Add the helper next to `toPlannerRestModel`:
```go
func toImageRestModel(m recipeimage.Model) *RestImageModel {
	thumb, full := recipeimage.URLs(m.RecipeID(), m.Id())
	return &RestImageModel{ThumbURL: thumb, FullURL: full, Width: m.Width(), Height: m.Height()}
}
```

- [ ] **Step 6: Run the tests**

Run: `cd services/recipe-service && go build ./... && go test ./internal/recipe/...`
Expected: PASS, including the existing `batch_test.go` and `processor_test.go`.

- [ ] **Step 7: Commit**

```bash
git add services/recipe-service/internal/recipe
git commit -m "feat(recipe-service): expose image attribute on recipe list and detail"
```

---

## Task 8: Retention cascade purges image rows into the outbox

**Files:**
- Modify: `services/recipe-service/internal/retention/handlers.go`
- Modify: `services/recipe-service/internal/retention/handlers_test.go`

**Interfaces:**
- Consumes: Task 4 `recipeimage.PurgeByRecipeIDsTx`, `recipeimage.Entity`, `recipeimage.ObjectDeletionEntity`.
- Produces: nothing new.

- [ ] **Step 1: Write the failing tests**

In `handlers_test.go`, add `&recipeimage.Entity{}, &recipeimage.ObjectDeletionEntity{},` to the `newDB` AutoMigrate list, and add the import `"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"` plus `"errors"`. Then append:
```go
func seedPurgeable(t *testing.T, db *gorm.DB) (tenantID, householdID, recipeID uuid.UUID) {
	t.Helper()
	tenantID, householdID, recipeID = uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	past := now.Add(-31 * 24 * time.Hour)
	db.Create(&recipe.Entity{
		Id: recipeID, TenantId: tenantID, HouseholdId: householdID,
		Title: "Old", Source: "manual", DeletedAt: &past, CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&recipeimage.Entity{
		Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID, RecipeId: recipeID,
		FullKey: "k/full.jpg", ThumbKey: "k/thumb.jpg", ContentType: "image/jpeg",
		Width: 10, Height: 10, FullBytes: 1, ThumbBytes: 1, CreatedAt: now, UpdatedAt: now,
	})
	return
}

func TestDeletedRecipesRestoreWindowPurgesImages(t *testing.T) {
	db := newDB(t)
	tid, hid, _ := seedPurgeable(t, db)
	scope := sr.Scope{TenantId: tid, Kind: sr.ScopeHousehold, ScopeId: hid}

	res, err := DeletedRecipesRestoreWindow{}.Reap(context.Background(), db, scope, 30, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 2 { // 1 image row + 1 recipe row
		t.Fatalf("Deleted = %d, want 2", res.Deleted)
	}
	var images, outbox int64
	db.Model(&recipeimage.Entity{}).Count(&images)
	db.Model(&recipeimage.ObjectDeletionEntity{}).Count(&outbox)
	if images != 0 || outbox != 2 {
		t.Fatalf("images=%d outbox=%d, want 0 and 2", images, outbox)
	}
}

func TestDeletedRecipesRestoreWindowDryRunLeavesImages(t *testing.T) {
	db := newDB(t)
	tid, hid, _ := seedPurgeable(t, db)
	scope := sr.Scope{TenantId: tid, Kind: sr.ScopeHousehold, ScopeId: hid}
	rollback := errors.New("dry run")

	// The shared reaper implements dry-run by rolling back the tx.
	err := db.Transaction(func(tx *gorm.DB) error {
		res, err := DeletedRecipesRestoreWindow{}.Reap(context.Background(), tx, scope, 30, true)
		if err != nil {
			return err
		}
		if res.Deleted != 2 {
			t.Fatalf("dry-run Deleted = %d, want 2 (counts images)", res.Deleted)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	var images, outbox int64
	db.Model(&recipeimage.Entity{}).Count(&images)
	db.Model(&recipeimage.ObjectDeletionEntity{}).Count(&outbox)
	if images != 1 || outbox != 0 {
		t.Fatalf("images=%d outbox=%d, want 1 and 0", images, outbox)
	}
}
```

Update the expected `Deleted` count in the existing `TestDeletedRecipesRestoreWindowBoundary` only if it seeds no images. It does not seed any, so its count is unchanged.

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `cd services/recipe-service && go test ./internal/retention/`
Expected: FAIL (`images=1 outbox=0`).

- [ ] **Step 3: Implement**

In `handlers.go`, add the import `"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"`. Change the doc comment on `DeletedRecipesRestoreWindow` to:
```go
// DeletedRecipesRestoreWindow reaps soft-deleted recipes past their restore
// window. Cascade: recipes → tags → ingredients → recipe_restorations →
// plan_items → recipe_images (objects enqueued to the deletion outbox and
// removed from storage after commit).
```
In `cascadeDeleteRecipes`, before the final `r = tx.Where("id IN ?", ids).Delete(&recipe.Entity{})`, insert:
```go
	n, err := recipeimage.PurgeByRecipeIDsTx(tx, ids)
	if err != nil {
		return 0, err
	}
	total += n
```

- [ ] **Step 4: Run the tests**

Run: `cd services/recipe-service && go build ./... && go test ./internal/retention/ ./...`
Expected: PASS (the whole module).

- [ ] **Step 5: Commit**

```bash
git add services/recipe-service/internal/retention
git commit -m "feat(recipe-service): purge recipe images on retention cascade via outbox"
```

---

## Task 9: Deploy — compose MinIO, nginx, k8s, `.env.example`

**Files:**
- Modify: `deploy/compose/docker-compose.yml`
- Modify: `deploy/compose/nginx.conf:181-188`
- Modify: `deploy/k8s/recipe-service.yaml`
- Modify: `deploy/k8s/secrets.example.yaml`
- Modify: `deploy/k8s/ingress.yaml:8`
- Modify: `.env.example`

**Interfaces:**
- Consumes: the `S3_*` env contract (Task 1).
- Produces: a compose stack in which recipe-service reaches MinIO at `minio:9000` with bucket `home-hub-recipes`.

- [ ] **Step 1: Pick the image tags**

Look up the newest published tags:
```bash
curl -s "https://hub.docker.com/v2/repositories/minio/minio/tags?page_size=5&ordering=last_updated" | jq -r '.results[].name'
curl -s "https://hub.docker.com/v2/repositories/minio/mc/tags?page_size=5&ordering=last_updated" | jq -r '.results[].name'
```
Pick the newest `RELEASE.*` tag for each. Confirm both pull with `docker pull minio/minio:<tag>` and `docker pull minio/mc:<tag>`. Use those exact tags below, in place of `<MINIO_TAG>` and `<MC_TAG>`. Never use `latest`. If neither registry publishes a usable tag, stop and report BLOCKED with the output; see `context.md`.

- [ ] **Step 2: Compose services**

In `deploy/compose/docker-compose.yml`, add these services (place them before `recipe-service`):
```yaml
  minio:
    container_name: hh-minio
    image: minio/minio:<MINIO_TAG>
    command: server /data --console-address :9001
    expose:
      - "9000"
      - "9001"
    environment:
      - MINIO_ROOT_USER=${S3_ACCESS_KEY:-homehub}
      - MINIO_ROOT_PASSWORD=${S3_SECRET_KEY:-homehub-dev-secret}
    volumes:
      - minio-data:/data
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 5s
      retries: 12

  minio-init:
    container_name: hh-minio-init
    image: minio/mc:<MC_TAG>
    depends_on:
      minio:
        condition: service_healthy
    environment:
      - S3_ACCESS_KEY=${S3_ACCESS_KEY:-homehub}
      - S3_SECRET_KEY=${S3_SECRET_KEY:-homehub-dev-secret}
      - S3_BUCKET=${S3_BUCKET:-home-hub-recipes}
    entrypoint: ["/bin/sh", "-c"]
    command:
      - mc alias set local http://minio:9000 "$$S3_ACCESS_KEY" "$$S3_SECRET_KEY" && mc mb --ignore-existing "local/$$S3_BUCKET"
```
In `recipe-service`, append to `environment`:
```yaml
      - S3_ENDPOINT=minio:9000
      - S3_BUCKET=${S3_BUCKET:-home-hub-recipes}
      - S3_ACCESS_KEY=${S3_ACCESS_KEY:-homehub}
      - S3_SECRET_KEY=${S3_SECRET_KEY:-homehub-dev-secret}
      - S3_USE_SSL=false
      - S3_REGION=us-east-1
```
and add:
```yaml
    depends_on:
      minio-init:
        condition: service_completed_successfully
```
At the end of the file, add a top-level block. If a `volumes:` key already exists, add `minio-data:` under it instead:
```yaml
volumes:
  minio-data:
```

- [ ] **Step 3: Validate compose**

Run: `docker compose -f deploy/compose/docker-compose.yml config --quiet`
Expected: exit 0, no output.

- [ ] **Step 4: nginx**

In `deploy/compose/nginx.conf`, add `client_max_body_size 11m;` to the `location /api/v1/recipes {` block, right after the `proxy_pass $recipe_upstream;` line. Change no other location block, and do not add `proxy_request_buffering`.

- [ ] **Step 5: k8s**

In `deploy/k8s/ingress.yaml`, change `nginx.ingress.kubernetes.io/proxy-body-size: "10m"` to `"11m"`.

In `deploy/k8s/recipe-service.yaml`, add these env entries after `CATEGORY_SERVICE_URL`. `optional: true` keeps the pod starting with storage disabled (image endpoints return 503) if the secret has not been created yet:
```yaml
        - name: S3_ENDPOINT
          valueFrom:
            secretKeyRef:
              name: object-store-credentials
              key: S3_ENDPOINT
              optional: true
        - name: S3_BUCKET
          valueFrom:
            secretKeyRef:
              name: object-store-credentials
              key: S3_BUCKET
              optional: true
        - name: S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: object-store-credentials
              key: S3_ACCESS_KEY
              optional: true
        - name: S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: object-store-credentials
              key: S3_SECRET_KEY
              optional: true
        - name: S3_USE_SSL
          value: "false"
        - name: S3_REGION
          value: "us-east-1"
```

In `deploy/k8s/secrets.example.yaml`, append:
```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: object-store-credentials
  namespace: home-hub
type: Opaque
stringData:
  # S3-compatible endpoint host:port (no scheme). The bucket must already
  # exist; recipe-service only checks for it. Scope the key to
  # s3:GetObject, s3:PutObject, s3:DeleteObject and s3:ListBucket on this
  # bucket only.
  S3_ENDPOINT: "minio.home:9000"
  S3_BUCKET: "home-hub-recipes"
  S3_ACCESS_KEY: "changeme"
  S3_SECRET_KEY: "changeme"
```

- [ ] **Step 6: `.env.example`**

Append:
```
# Object storage (recipe-service images). Compose runs its own MinIO and
# falls back to these defaults when unset. S3_ENDPOINT empty disables images.
S3_ENDPOINT=minio:9000
S3_BUCKET=home-hub-recipes
S3_ACCESS_KEY=homehub
S3_SECRET_KEY=homehub-dev-secret
S3_USE_SSL=false
S3_REGION=us-east-1
```

- [ ] **Step 7: Validate YAML**

Run:
```bash
docker compose -f deploy/compose/docker-compose.yml config --quiet
python3 -c "import yaml,sys; [list(yaml.safe_load_all(open(f))) for f in sys.argv[1:]]" deploy/k8s/recipe-service.yaml deploy/k8s/secrets.example.yaml deploy/k8s/ingress.yaml
```
Expected: both exit 0.

- [ ] **Step 8: Commit**

```bash
git add deploy .env.example
git commit -m "chore(deploy): add MinIO for recipe images and raise recipe body limits"
```

---

## Task 10: Frontend transport — client `PUT` upload, types, service, hooks, validation helper

**Files:**
- Modify: `frontend/src/lib/api/client.ts` (`upload`, `uploadWithProgress`)
- Modify: `frontend/src/lib/api/__tests__/client.test.ts`
- Modify: `frontend/src/types/models/recipe.ts`
- Modify: `frontend/src/services/api/recipe.ts`
- Modify: `frontend/src/services/api/__tests__/recipe.test.ts`
- Modify: `frontend/src/lib/hooks/api/use-recipes.ts`
- Create: `frontend/src/lib/recipe-image.ts`
- Create: `frontend/src/lib/__tests__/recipe-image.test.ts`
- Modify: `frontend/src/components/features/recipes/__tests__/recipe-card.test.tsx` (fixture gets `image: null`)

**Interfaces:**
- Consumes: the backend REST contract (Tasks 6–7).
- Produces:
  - `export type UploadOptions = RequestOptions & { method?: "POST" | "PUT" }`; `api.upload<T>(path: string, formData: FormData, options?: UploadOptions): Promise<T>` (default `POST`)
  - `export interface RecipeImage { thumbUrl: string; fullUrl: string; width: number; height: number }`
  - `RecipeListAttributes.image: RecipeImage | null`. `RecipeDetailAttributes` inherits it.
  - `export interface RecipeImageResource { id: string; type: "recipe-images"; attributes: RecipeImage & { updatedAt: string } }`
  - `recipeService.uploadRecipeImage(tenant: Tenant, id: string, file: File): Promise<ApiResponse<RecipeImageResource>>`; `recipeService.deleteRecipeImage(tenant: Tenant, id: string): Promise<void>`
  - `useUploadRecipeImage()`: a mutation over `{ id: string; file: File }`. `useDeleteRecipeImage()`: a mutation over `id: string`. Both invalidate `recipeKeys.lists` and `recipeKeys.detail(id)` on success.
  - `lib/recipe-image.ts`: `RECIPE_IMAGE_MAX_BYTES`, `RECIPE_IMAGE_TYPES`, `RECIPE_IMAGE_ACCEPT`, `RECIPE_IMAGE_ERROR`, `validateRecipeImageFile(file: File): string | null`

- [ ] **Step 1: Write the failing tests**

`frontend/src/lib/__tests__/recipe-image.test.ts`:
```ts
import { describe, it, expect } from "vitest";
import { validateRecipeImageFile, RECIPE_IMAGE_ERROR, RECIPE_IMAGE_MAX_BYTES } from "../recipe-image";

function fileOf(type: string, size: number): File {
  const f = new File(["x"], "photo", { type });
  Object.defineProperty(f, "size", { value: size });
  return f;
}

describe("validateRecipeImageFile", () => {
  it.each(["image/jpeg", "image/png", "image/webp"])("accepts %s", (type) => {
    expect(validateRecipeImageFile(fileOf(type, 1000))).toBeNull();
  });

  it("accepts exactly 10 MiB", () => {
    expect(validateRecipeImageFile(fileOf("image/jpeg", RECIPE_IMAGE_MAX_BYTES))).toBeNull();
  });

  it("rejects one byte over 10 MiB", () => {
    expect(validateRecipeImageFile(fileOf("image/jpeg", RECIPE_IMAGE_MAX_BYTES + 1))).toBe(RECIPE_IMAGE_ERROR);
  });

  it.each(["image/gif", "image/heic", "application/pdf"])("rejects %s", (type) => {
    expect(validateRecipeImageFile(fileOf(type, 1000))).toBe(RECIPE_IMAGE_ERROR);
  });

  it("rejects empty mime type", () => {
    expect(validateRecipeImageFile(fileOf("", 1000))).toBe(RECIPE_IMAGE_ERROR);
  });

  it("uses the PRD message", () => {
    expect(RECIPE_IMAGE_ERROR).toBe("Use a JPEG, PNG or WebP image under 10 MB.");
  });
});
```

Append to `frontend/src/lib/api/__tests__/client.test.ts`, inside `describe("ApiClient", …)`:
```ts
  it("upload defaults to POST and sends FormData without a content type", async () => {
    const api = await getApi();
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ data: {} }),
    });
    const form = new FormData();
    await api.upload("/things", form);
    const [, init] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(init.method).toBe("POST");
    expect(init.body).toBe(form);
    expect(Object.keys(init.headers as Record<string, string>).map((k) => k.toLowerCase())).not.toContain("content-type");
  });

  it("upload honours method PUT", async () => {
    const api = await getApi();
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ data: {} }),
    });
    await api.upload("/recipes/r1/image", new FormData(), { method: "PUT" });
    const [url, init] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(url).toBe("/api/v1/recipes/r1/image");
    expect(init.method).toBe("PUT");
  });
```

In `frontend/src/services/api/__tests__/recipe.test.ts`, extend the `vi.mock("@/lib/api/client", …)` factory's `api` object with:
```ts
    upload: vi.fn().mockResolvedValue({ data: { id: "img-1", type: "recipe-images", attributes: {} } }),
    delete: vi.fn().mockResolvedValue(undefined),
```
and append:
```ts
describe("recipeService image methods", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uploadRecipeImage PUTs multipart FormData with the file part", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    const file = new File(["x"], "a.jpg", { type: "image/jpeg" });
    await recipeService.uploadRecipeImage(tenant, "r1", file);
    const [path, form, options] = (api.upload as ReturnType<typeof vi.fn>).mock.calls[0]!;
    expect(path).toBe("/recipes/r1/image");
    expect(form).toBeInstanceOf(FormData);
    expect((form as FormData).get("file")).toBe(file);
    expect(options).toEqual({ method: "PUT" });
  });

  it("deleteRecipeImage DELETEs the image path", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.deleteRecipeImage(tenant, "r1");
    expect(api.delete).toHaveBeenCalledWith("/recipes/r1/image");
  });
});
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `cd frontend && npx vitest run src/lib/__tests__/recipe-image.test.ts src/lib/api/__tests__/client.test.ts src/services/api/__tests__/recipe.test.ts`
Expected: FAIL (missing module, method not PUT, missing service methods).

- [ ] **Step 3: Implement `lib/recipe-image.ts`**

```ts
export const RECIPE_IMAGE_MAX_BYTES = 10 * 1024 * 1024;
export const RECIPE_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp"] as const;
export const RECIPE_IMAGE_ACCEPT = RECIPE_IMAGE_TYPES.join(",");
export const RECIPE_IMAGE_ERROR = "Use a JPEG, PNG or WebP image under 10 MB.";

/** Returns an error message for an unacceptable file, or null when valid. */
export function validateRecipeImageFile(file: File): string | null {
  const typeOk = (RECIPE_IMAGE_TYPES as readonly string[]).includes(file.type);
  if (!typeOk || file.size > RECIPE_IMAGE_MAX_BYTES) {
    return RECIPE_IMAGE_ERROR;
  }
  return null;
}
```

- [ ] **Step 4: Extend `api.upload`**

In `lib/api/client.ts`, add after the `RequestOptions` interface:
```ts
export type UploadOptions = RequestOptions & { method?: "POST" | "PUT" };
```
Change the `upload` signature to `async upload<T>(path: string, formData: FormData, options?: UploadOptions): Promise<T>`. In its `fetchWithRetry` init, replace `method: "POST",` with `method: options?.method ?? "POST",`. Change `uploadWithProgress`'s `options` parameter type to `UploadOptions`, and replace `xhr.open("POST", url, true);` with `xhr.open(options.method ?? "POST", url, true);`.

- [ ] **Step 5: Types**

In `types/models/recipe.ts`, add:
```ts
export interface RecipeImage {
  thumbUrl: string;
  fullUrl: string;
  width: number;
  height: number;
}

export interface RecipeImageResource {
  id: string;
  type: "recipe-images";
  attributes: RecipeImage & { updatedAt: string };
}
```
and add `image: RecipeImage | null;` to `RecipeListAttributes`, after `usageCount?`.

In `components/features/recipes/__tests__/recipe-card.test.tsx`, add `image: null,` to the `makeRecipe` attributes, before `...overrides`.

- [ ] **Step 6: Service methods**

In `services/api/recipe.ts`, add `RecipeImageResource` to the type import, `import type { ApiResponse } from "@/types/api/responses";`, and these methods to `RecipeService`:
```ts
  uploadRecipeImage(tenant: Tenant, id: string, file: File): Promise<ApiResponse<RecipeImageResource>> {
    this.setTenant(tenant);
    const form = new FormData();
    form.append("file", file);
    return api.upload<ApiResponse<RecipeImageResource>>(`/recipes/${id}/image`, form, { method: "PUT" });
  }

  deleteRecipeImage(tenant: Tenant, id: string) {
    return this.remove(tenant, `/recipes/${id}/image`);
  }
```

- [ ] **Step 7: Hooks**

In `lib/hooks/api/use-recipes.ts`, add after `useDeleteRecipe`:
```ts
export function useUploadRecipeImage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) =>
      recipeService.uploadRecipeImage(tenant!, id, file),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: recipeKeys.detail(tenant, household, variables.id) });
    },
  });
}

export function useDeleteRecipeImage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => recipeService.deleteRecipeImage(tenant!, id),
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: recipeKeys.detail(tenant, household, id) });
    },
  });
}
```
These hooks deliberately do not toast `onError`; the form owns the partial-failure toast (FR-6).

- [ ] **Step 8: Run the tests and the typecheck**

Run: `cd frontend && npx vitest run src/lib src/services src/components/features/recipes && npx tsc -b`
Expected: PASS, with no type errors. If `tsc` reports other fixtures that build `RecipeListAttributes` without `image`, add `image: null` to them.

- [ ] **Step 9: Commit**

```bash
git add frontend/src
git commit -m "feat(frontend): add recipe image upload/delete transport and types"
```

---

## Task 11: Frontend rendering — `RecipeImage`, card thumbnail, detail hero and lightbox

**Files:**
- Create: `frontend/src/components/features/recipes/recipe-image.tsx`
- Create: `frontend/src/components/features/recipes/__tests__/recipe-image.test.tsx`
- Modify: `frontend/src/components/features/recipes/recipe-card.tsx`
- Modify: `frontend/src/components/features/recipes/__tests__/recipe-card.test.tsx`
- Modify: `frontend/src/pages/RecipeDetailPage.tsx`
- Create: `frontend/src/pages/__tests__/RecipeDetailPage.test.tsx`

**Interfaces:**
- Consumes: Task 10 `RecipeImage` type, `attributes.image`.
- Produces: `<RecipeImage image={RecipeImage | null} variant="hero" | "card" alt={string} className?={string} />`.

- [ ] **Step 1: Write the failing component tests**

`components/features/recipes/__tests__/recipe-image.test.tsx`:
```tsx
import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { RecipeImage } from "../recipe-image";

const image = {
  thumbUrl: "/api/v1/recipes/r1/image?size=thumb&v=i1",
  fullUrl: "/api/v1/recipes/r1/image?size=full&v=i1",
  width: 1600,
  height: 900,
};

describe("RecipeImage", () => {
  it("card variant renders the lazy thumb rendition", () => {
    render(<RecipeImage image={image} variant="card" alt="Soup" />);
    const img = screen.getByRole("img", { name: "Soup" });
    expect(img).toHaveAttribute("src", image.thumbUrl);
    expect(img).toHaveAttribute("loading", "lazy");
  });

  it("hero variant renders the eager full rendition", () => {
    render(<RecipeImage image={image} variant="hero" alt="Soup" />);
    const img = screen.getByRole("img", { name: "Soup" });
    expect(img).toHaveAttribute("src", image.fullUrl);
    expect(img).toHaveAttribute("loading", "eager");
  });

  it("card variant shows the placeholder when there is no image", () => {
    render(<RecipeImage image={null} variant="card" alt="Soup" />);
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("hero variant renders nothing when there is no image", () => {
    const { container } = render(<RecipeImage image={null} variant="hero" alt="Soup" />);
    expect(container).toBeEmptyDOMElement();
  });

  it("card variant falls back to the placeholder on load error", () => {
    render(<RecipeImage image={image} variant="card" alt="Soup" />);
    fireEvent.error(screen.getByRole("img", { name: "Soup" }));
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
  });

  it("hero variant renders nothing on load error", () => {
    const { container } = render(<RecipeImage image={image} variant="hero" alt="Soup" />);
    fireEvent.error(screen.getByRole("img", { name: "Soup" }));
    expect(container).toBeEmptyDOMElement();
  });

  it("a new image after an error renders again", () => {
    const { rerender } = render(<RecipeImage image={image} variant="card" alt="Soup" />);
    fireEvent.error(screen.getByRole("img", { name: "Soup" }));
    const next = { ...image, thumbUrl: image.thumbUrl.replace("v=i1", "v=i2") };
    rerender(<RecipeImage image={next} variant="card" alt="Soup" />);
    expect(screen.getByRole("img", { name: "Soup" })).toHaveAttribute("src", next.thumbUrl);
  });
});
```

Append to `recipe-card.test.tsx`:
```tsx
  it("renders the thumbnail when the recipe has an image", () => {
    render(
      <RecipeCard
        recipe={makeRecipe({ image: { thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 10, height: 10 } })}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByRole("img", { name: "Pasta Carbonara" })).toHaveAttribute("src", "/t.jpg");
  });

  it("renders the placeholder when the recipe has no image", () => {
    render(<RecipeCard recipe={makeRecipe()} onDelete={vi.fn()} />);
    expect(screen.getByTestId("recipe-image-placeholder")).toBeInTheDocument();
  });
```

`pages/__tests__/RecipeDetailPage.test.tsx`:
```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { RecipeDetail } from "@/types/models/recipe";

const useRecipeMock = vi.fn();
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipe: () => useRecipeMock(),
  useDeleteRecipe: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("@/lib/hooks/api/use-ingredient-normalization", () => ({
  useRenormalize: () => ({ mutate: vi.fn(), isPending: false }),
}));
vi.mock("react-router-dom", () => ({
  useParams: () => ({ id: "r1" }),
  useNavigate: () => vi.fn(),
}));
vi.mock("@/components/features/recipes/recipe-ingredients", () => ({ RecipeIngredients: () => null }));
vi.mock("@/components/features/recipes/recipe-steps", () => ({ RecipeSteps: () => null }));
vi.mock("@/components/features/recipes/cook-mode", () => ({ CookMode: () => null }));

import { RecipeDetailPage } from "../RecipeDetailPage";

function detail(image: RecipeDetail["attributes"]["image"]): { data: RecipeDetail } {
  return {
    data: {
      id: "r1",
      type: "recipes",
      attributes: {
        title: "Soup",
        tags: [],
        plannerReady: false,
        resolvedIngredients: 0,
        totalIngredients: 0,
        createdAt: "",
        updatedAt: "",
        image,
        source: "",
        ingredients: [],
        steps: [],
        plannerIssues: [],
      },
    },
  };
}

describe("RecipeDetailPage hero", () => {
  beforeEach(() => vi.clearAllMocks());

  it("renders no hero when the recipe has no image", () => {
    useRecipeMock.mockReturnValue({ data: detail(null), isLoading: false });
    render(<RecipeDetailPage />);
    expect(screen.queryByRole("img", { name: "Soup" })).not.toBeInTheDocument();
    expect(screen.queryByTestId("recipe-image-placeholder")).not.toBeInTheDocument();
  });

  it("renders the hero and opens an uncropped lightbox on click", async () => {
    const user = userEvent.setup();
    useRecipeMock.mockReturnValue({
      data: detail({ thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 1600, height: 900 }),
      isLoading: false,
    });
    render(<RecipeDetailPage />);
    const hero = screen.getByRole("img", { name: "Soup" });
    expect(hero).toHaveAttribute("src", "/f.jpg");

    await user.click(screen.getByRole("button", { name: /view photo full size/i }));
    const imgs = await screen.findAllByRole("img", { name: "Soup" });
    expect(imgs).toHaveLength(2);
    expect(imgs[1]).toHaveClass("object-contain");
  });
});
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `cd frontend && npx vitest run src/components/features/recipes src/pages/__tests__/RecipeDetailPage.test.tsx`
Expected: FAIL (missing module, no hero).

- [ ] **Step 3: Implement `recipe-image.tsx`**

```tsx
import { useState } from "react";
import { ChefHat } from "lucide-react";
import { cn } from "@/lib/utils";
import type { RecipeImage as RecipeImageData } from "@/types/models/recipe";

interface RecipeImageProps {
  image: RecipeImageData | null;
  variant: "hero" | "card";
  alt: string;
  className?: string;
}

function RecipeImagePlaceholder({ className }: { className?: string | undefined }) {
  return (
    <div
      data-testid="recipe-image-placeholder"
      aria-hidden="true"
      className={cn("flex items-center justify-center bg-muted", className)}
    >
      <ChefHat className="h-6 w-6 text-muted-foreground/50" />
    </div>
  );
}

export function RecipeImage({ image, variant, alt, className }: RecipeImageProps) {
  // Keyed by src so a replaced image (new ?v=) gets a fresh attempt.
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const isHero = variant === "hero";
  const src = image ? (isHero ? image.fullUrl : image.thumbUrl) : null;

  if (!src || failedSrc === src) {
    // The detail page has no placeholder (FR-12); cards keep their footprint (FR-16/17).
    return isHero ? null : <RecipeImagePlaceholder className={className} />;
  }

  return (
    <img
      src={src}
      alt={alt}
      loading={isHero ? "eager" : "lazy"}
      className={cn("object-cover", className)}
      onError={() => setFailedSrc(src)}
    />
  );
}
```

- [ ] **Step 4: Update `recipe-card.tsx`**

Import `RecipeImage` from `"./recipe-image"`. Replace the JSX returned by `RecipeCard` with the following. The header/content markup is unchanged and only wrapped. Below `md` the image is a 64×64 square on the left; at `md+` it is a full-width 16:9 band on top:
```tsx
  return (
    <Card
      size="sm"
      className="cursor-pointer md:pt-0"
      onClick={() => navigate(`/app/recipes/${id}`)}
    >
      <div className="flex items-start gap-3 pl-3 md:flex-col md:items-stretch md:pl-0">
        <RecipeImage
          image={attributes.image}
          variant="card"
          alt={attributes.title}
          className="size-16 shrink-0 rounded-md md:aspect-video md:h-auto md:w-full md:rounded-none"
        />
        <div className="flex min-w-0 flex-1 flex-col gap-3">
          <CardHeader className="max-md:pl-0">
            {/* existing CardHeader children unchanged */}
          </CardHeader>
          <CardContent className="max-md:pl-0">
            {/* existing CardContent children unchanged */}
          </CardContent>
        </div>
      </div>
    </Card>
  );
```
Move the existing children of `CardHeader` (the title row and `CardAction` menu) and of `CardContent` (the description and the time/usage/tags row) into place unchanged. The visual check against `mockup.html` (desktop and 375 px widths) happens in Task 13's manual acceptance (item 3). Adjust only spacing classes there if the image doesn't sit flush with the card edges.

- [ ] **Step 5: Update `RecipeDetailPage.tsx`**

Add the imports:
```tsx
import { RecipeImage } from "@/components/features/recipes/recipe-image";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
```
Add `const [lightboxOpen, setLightboxOpen] = useState(false);` next to the existing `cookModeOpen` state, above any early return. In the header `<div className="space-y-3">`, right after the "Recipes" back `<Button>`, insert:
```tsx
        {attrs.image && (
          <button
            type="button"
            aria-label="View photo full size"
            className="-mx-4 block w-[calc(100%+2rem)] md:mx-0 md:w-full"
            onClick={() => setLightboxOpen(true)}
          >
            <RecipeImage
              image={attrs.image}
              variant="hero"
              alt={attrs.title}
              className="aspect-[4/3] w-full md:aspect-video md:rounded-lg"
            />
          </button>
        )}
```
Before `<CookMode …/>`, add:
```tsx
      {attrs.image && (
        <Dialog open={lightboxOpen} onOpenChange={setLightboxOpen}>
          <DialogContent className="max-w-[95vw] p-2 sm:max-w-[95vw]">
            <DialogTitle className="sr-only">{attrs.title}</DialogTitle>
            <img
              src={attrs.image.fullUrl}
              alt={attrs.title}
              className="max-h-[85vh] w-full object-contain"
            />
          </DialogContent>
        </Dialog>
      )}
```

- [ ] **Step 6: Run the tests, the typecheck and lint**

Run: `cd frontend && npx vitest run src/components/features/recipes src/pages && npx tsc -b && npx eslint src/components/features/recipes src/pages/RecipeDetailPage.tsx`
Expected: PASS, with no lint errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src
git commit -m "feat(frontend): show recipe thumbnails on cards and hero with lightbox on detail"
```

---

## Task 12: Frontend form — image draft hook, Photo field and save sequencing

**Files:**
- Create: `frontend/src/lib/hooks/use-recipe-image-draft.ts`
- Create: `frontend/src/lib/hooks/__tests__/use-recipe-image-draft.test.ts`
- Create: `frontend/src/components/features/recipes/recipe-image-field.tsx`
- Create: `frontend/src/components/features/recipes/__tests__/recipe-image-field.test.tsx`
- Modify: `frontend/src/pages/RecipeFormPage.tsx`
- Create: `frontend/src/pages/__tests__/RecipeFormPage.test.tsx`

**Interfaces:**
- Consumes: Task 10 (`useUploadRecipeImage`, `useDeleteRecipeImage`, `validateRecipeImageFile`, `RECIPE_IMAGE_ACCEPT`, `RecipeImage` type).
- Produces:
  - `export type ImageDraft = { kind: "unchanged" } | { kind: "staged"; file: File; previewUrl: string } | { kind: "removed" }`
  - `useRecipeImageDraft(existing: RecipeImage | null): { draft: ImageDraft; previewSrc: string | null; stage(file: File): void; remove(): void; undo(): void }`
  - `<RecipeImageField draft previewSrc onStage onRemove onUndo disabled? />`

- [ ] **Step 1: Write the failing hook tests**

`lib/hooks/__tests__/use-recipe-image-draft.test.ts`:
```ts
import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useRecipeImageDraft } from "../use-recipe-image-draft";

const existing = { thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 10, height: 10 };
const file = () => new File(["x"], "a.jpg", { type: "image/jpeg" });

describe("useRecipeImageDraft", () => {
  let n = 0;
  beforeEach(() => {
    n = 0;
    URL.createObjectURL = vi.fn(() => `blob:${++n}`);
    URL.revokeObjectURL = vi.fn();
  });

  it("starts unchanged and previews the existing image", () => {
    const { result } = renderHook(() => useRecipeImageDraft(existing));
    expect(result.current.draft.kind).toBe("unchanged");
    expect(result.current.previewSrc).toBe("/f.jpg");
  });

  it("stage creates an object URL; replacing revokes the previous one", () => {
    const { result } = renderHook(() => useRecipeImageDraft(null));
    act(() => result.current.stage(file()));
    expect(result.current.previewSrc).toBe("blob:1");
    act(() => result.current.stage(file()));
    expect(result.current.previewSrc).toBe("blob:2");
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });

  it("remove with an existing image stages removal; undo restores", () => {
    const { result } = renderHook(() => useRecipeImageDraft(existing));
    act(() => result.current.remove());
    expect(result.current.draft.kind).toBe("removed");
    expect(result.current.previewSrc).toBeNull();
    act(() => result.current.undo());
    expect(result.current.draft.kind).toBe("unchanged");
    expect(result.current.previewSrc).toBe("/f.jpg");
  });

  it("remove of a staged file without an existing image returns to unchanged", () => {
    const { result } = renderHook(() => useRecipeImageDraft(null));
    act(() => result.current.stage(file()));
    act(() => result.current.remove());
    expect(result.current.draft.kind).toBe("unchanged");
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });

  it("revokes the object URL on unmount", () => {
    const { result, unmount } = renderHook(() => useRecipeImageDraft(null));
    act(() => result.current.stage(file()));
    unmount();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:1");
  });
});
```

- [ ] **Step 2: Write the failing field tests**

`components/features/recipes/__tests__/recipe-image-field.test.tsx`:
```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RecipeImageField } from "../recipe-image-field";

const noop = () => {};

describe("RecipeImageField", () => {
  it("stages a valid picked file", async () => {
    const user = userEvent.setup();
    const onStage = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc={null} onStage={onStage} onRemove={noop} onUndo={noop} />);
    const file = new File(["x"], "a.jpg", { type: "image/jpeg" });
    await user.upload(screen.getByLabelText("Choose photo"), file);
    expect(onStage).toHaveBeenCalledWith(file);
  });

  it("shows the inline error and stages nothing for an invalid file", () => {
    const onStage = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc={null} onStage={onStage} onRemove={noop} onUndo={noop} />);
    const gif = new File(["x"], "a.gif", { type: "image/gif" });
    fireEvent.change(screen.getByLabelText("Choose photo"), { target: { files: [gif] } });
    expect(onStage).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toHaveTextContent("Use a JPEG, PNG or WebP image under 10 MB.");
  });

  it("stages a dropped file", () => {
    const onStage = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc={null} onStage={onStage} onRemove={noop} onUndo={noop} />);
    const file = new File(["x"], "a.png", { type: "image/png" });
    fireEvent.drop(screen.getByRole("button", { name: /add a photo/i }), { dataTransfer: { files: [file] } });
    expect(onStage).toHaveBeenCalledWith(file);
  });

  it("shows preview with Replace and Remove", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    render(<RecipeImageField draft={{ kind: "unchanged" }} previewSrc="/f.jpg" onStage={noop} onRemove={onRemove} onUndo={noop} />);
    expect(screen.getByRole("img", { name: /recipe photo preview/i })).toHaveAttribute("src", "/f.jpg");
    expect(screen.getByRole("button", { name: /replace/i })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /remove/i }));
    expect(onRemove).toHaveBeenCalled();
  });

  it("offers Undo after a staged removal", async () => {
    const user = userEvent.setup();
    const onUndo = vi.fn();
    render(<RecipeImageField draft={{ kind: "removed" }} previewSrc={null} onStage={noop} onRemove={noop} onUndo={onUndo} />);
    await user.click(screen.getByRole("button", { name: /undo/i }));
    expect(onUndo).toHaveBeenCalled();
  });
});
```

- [ ] **Step 3: Write the failing form tests**

`pages/__tests__/RecipeFormPage.test.tsx`:
```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const navigate = vi.fn();
let params: { id?: string } = {};
vi.mock("react-router-dom", () => ({
  useParams: () => params,
  useNavigate: () => navigate,
}));

const toastError = vi.fn();
const toastSuccess = vi.fn();
vi.mock("sonner", () => ({ toast: { error: (...a: unknown[]) => toastError(...a), success: (...a: unknown[]) => toastSuccess(...a) } }));

const createMutateAsync = vi.fn();
const updateMutateAsync = vi.fn();
const uploadMutateAsync = vi.fn();
const deleteImageMutateAsync = vi.fn();
let uploadPending = false;
let existing: unknown = undefined;
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipe: () => ({ data: existing }),
  useCreateRecipe: () => ({ mutateAsync: createMutateAsync }),
  useUpdateRecipe: () => ({ mutateAsync: updateMutateAsync }),
  useUploadRecipeImage: () => ({ mutateAsync: uploadMutateAsync, isPending: uploadPending }),
  useDeleteRecipeImage: () => ({ mutateAsync: deleteImageMutateAsync, isPending: false }),
}));
vi.mock("@/lib/hooks/use-cooklang-preview", () => ({
  useCooklangPreview: () => ({
    metadata: { title: "Soup" },
    ingredients: [],
    steps: [],
    errors: [],
    notes: [],
    normalization: [],
    isLoading: false,
  }),
}));
vi.mock("@/lib/hooks/use-mobile", () => ({ useMobile: () => false }));
vi.mock("@/components/features/recipes/cooklang-preview", () => ({ CooklangPreview: () => null }));
vi.mock("@/components/features/recipes/cooklang-help", () => ({ CooklangHelp: () => null }));
vi.mock("@/components/features/recipes/planner-config-form", () => ({ PlannerConfigForm: () => null }));
vi.mock("@/components/features/recipes/recipe-ingredients", () => ({ RecipeIngredients: () => null }));

import { RecipeFormPage } from "../RecipeFormPage";

const photo = () => new File(["x"], "a.jpg", { type: "image/jpeg" });
const existingImage = { thumbUrl: "/t.jpg", fullUrl: "/f.jpg", width: 10, height: 10 };

function existingRecipe(image: unknown) {
  return {
    data: {
      id: "r1",
      type: "recipes",
      attributes: { title: "Soup", source: "---\ntitle: Soup\n---\nBoil.", tags: [], ingredients: [], image },
    },
  };
}

async function typeSource(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText("Recipe (Cooklang)"), "Boil water.");
}

describe("RecipeFormPage image save sequencing", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    params = {};
    existing = undefined;
    uploadPending = false;
    URL.createObjectURL = vi.fn(() => "blob:preview");
    URL.revokeObjectURL = vi.fn();
  });

  it("create + upload, then navigates", async () => {
    const user = userEvent.setup();
    createMutateAsync.mockResolvedValue({ data: { id: "r1" } });
    uploadMutateAsync.mockResolvedValue({});
    render(<RecipeFormPage />);
    await typeSource(user);
    const file = photo();
    await user.upload(screen.getByLabelText("Choose photo"), file);
    await user.click(screen.getByRole("button", { name: /create recipe/i }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(uploadMutateAsync).toHaveBeenCalledWith({ id: "r1", file });
  });

  it("edit + remove deletes the image after the PATCH", async () => {
    const user = userEvent.setup();
    params = { id: "r1" };
    existing = existingRecipe(existingImage);
    updateMutateAsync.mockResolvedValue({});
    deleteImageMutateAsync.mockResolvedValue(undefined);
    render(<RecipeFormPage />);
    await user.click(screen.getByRole("button", { name: /^remove$/i }));
    await user.click(screen.getByRole("button", { name: /update recipe/i }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(updateMutateAsync).toHaveBeenCalled();
    expect(deleteImageMutateAsync).toHaveBeenCalledWith("r1");
    expect(uploadMutateAsync).not.toHaveBeenCalled();
  });

  it("remove then undo makes no image request", async () => {
    const user = userEvent.setup();
    params = { id: "r1" };
    existing = existingRecipe(existingImage);
    updateMutateAsync.mockResolvedValue({});
    render(<RecipeFormPage />);
    await user.click(screen.getByRole("button", { name: /^remove$/i }));
    await user.click(screen.getByRole("button", { name: /undo/i }));
    await user.click(screen.getByRole("button", { name: /update recipe/i }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(deleteImageMutateAsync).not.toHaveBeenCalled();
    expect(uploadMutateAsync).not.toHaveBeenCalled();
  });

  it("cancel makes no request", async () => {
    const user = userEvent.setup();
    render(<RecipeFormPage />);
    await user.upload(screen.getByLabelText("Choose photo"), photo());
    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(navigate).toHaveBeenCalledWith("/app/recipes");
    expect(createMutateAsync).not.toHaveBeenCalled();
    expect(uploadMutateAsync).not.toHaveBeenCalled();
  });

  it("partial failure keeps the user on the form and retries only the image", async () => {
    const user = userEvent.setup();
    createMutateAsync.mockResolvedValue({ data: { id: "r1" } });
    uploadMutateAsync.mockRejectedValueOnce(new Error("503")).mockResolvedValueOnce({});
    render(<RecipeFormPage />);
    await typeSource(user);
    const file = photo();
    await user.upload(screen.getByLabelText("Choose photo"), file);
    await user.click(screen.getByRole("button", { name: /create recipe/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("Recipe saved, but the photo failed to upload — try again."),
    );
    expect(navigate).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: /update recipe/i }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith("/app/recipes/r1"));
    expect(createMutateAsync).toHaveBeenCalledTimes(1);
    expect(updateMutateAsync).not.toHaveBeenCalled();
    expect(uploadMutateAsync).toHaveBeenCalledTimes(2);
    expect(uploadMutateAsync).toHaveBeenLastCalledWith({ id: "r1", file });
  });

  it("shows the uploading label while the image request is pending", () => {
    uploadPending = true;
    render(<RecipeFormPage />);
    const button = screen.getByRole("button", { name: "Uploading photo…" });
    expect(button).toBeDisabled();
  });
});
```

If `recipeFormSchema` requires more than a non-empty `source`, type whatever minimum it requires in `typeSource` (check `src/lib/schemas/recipe.schema.ts`).

- [ ] **Step 4: Run the tests and confirm they fail**

Run: `cd frontend && npx vitest run src/lib/hooks/__tests__/use-recipe-image-draft.test.ts src/components/features/recipes/__tests__/recipe-image-field.test.tsx src/pages/__tests__/RecipeFormPage.test.tsx`
Expected: FAIL (missing modules, no Photo section).

- [ ] **Step 5: Implement the draft hook**

`lib/hooks/use-recipe-image-draft.ts`:
```ts
import { useCallback, useEffect, useRef, useState } from "react";
import type { RecipeImage } from "@/types/models/recipe";

export type ImageDraft =
  | { kind: "unchanged" }
  | { kind: "staged"; file: File; previewUrl: string }
  | { kind: "removed" };

/**
 * Holds the not-yet-saved image change for the recipe form and owns the
 * object-URL lifecycle of the staged preview.
 */
export function useRecipeImageDraft(existing: RecipeImage | null) {
  const [draft, setDraft] = useState<ImageDraft>({ kind: "unchanged" });
  const urlRef = useRef<string | null>(null);

  const revoke = useCallback(() => {
    if (urlRef.current) {
      URL.revokeObjectURL(urlRef.current);
      urlRef.current = null;
    }
  }, []);

  useEffect(() => revoke, [revoke]);

  const stage = useCallback(
    (file: File) => {
      revoke();
      const previewUrl = URL.createObjectURL(file);
      urlRef.current = previewUrl;
      setDraft({ kind: "staged", file, previewUrl });
    },
    [revoke],
  );

  const remove = useCallback(() => {
    revoke();
    setDraft(existing ? { kind: "removed" } : { kind: "unchanged" });
  }, [revoke, existing]);

  const undo = useCallback(() => {
    revoke();
    setDraft({ kind: "unchanged" });
  }, [revoke]);

  const previewSrc =
    draft.kind === "staged" ? draft.previewUrl : draft.kind === "removed" ? null : (existing?.fullUrl ?? null);

  return { draft, previewSrc, stage, remove, undo };
}
```

- [ ] **Step 6: Implement the field**

`components/features/recipes/recipe-image-field.tsx`:
```tsx
import { useRef, useState, type DragEvent } from "react";
import { ImagePlus, RefreshCw, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { RECIPE_IMAGE_ACCEPT, validateRecipeImageFile } from "@/lib/recipe-image";
import type { ImageDraft } from "@/lib/hooks/use-recipe-image-draft";

interface RecipeImageFieldProps {
  draft: ImageDraft;
  previewSrc: string | null;
  onStage: (file: File) => void;
  onRemove: () => void;
  onUndo: () => void;
  disabled?: boolean;
}

export function RecipeImageField({ draft, previewSrc, onStage, onRemove, onUndo, disabled = false }: RecipeImageFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);

  const accept = (file: File | undefined) => {
    if (!file) return;
    const message = validateRecipeImageFile(file);
    setError(message);
    if (!message) onStage(file);
  };

  const openPicker = () => inputRef.current?.click();

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragging(false);
    if (!disabled) accept(e.dataTransfer.files[0]);
  };

  return (
    <div className="space-y-2">
      <span className="text-sm font-medium leading-none">Photo</span>
      <input
        ref={inputRef}
        type="file"
        accept={RECIPE_IMAGE_ACCEPT}
        aria-label="Choose photo"
        className="hidden"
        disabled={disabled}
        onChange={(e) => {
          accept(e.target.files?.[0]);
          e.target.value = "";
        }}
      />
      {previewSrc ? (
        <div className="flex items-start gap-4">
          <img src={previewSrc} alt="Recipe photo preview" className="h-32 w-48 rounded-md object-cover" />
          <div className="flex flex-col gap-2">
            <Button type="button" variant="outline" size="sm" onClick={openPicker} disabled={disabled}>
              <RefreshCw className="mr-1 h-4 w-4" /> Replace
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="text-destructive"
              onClick={() => {
                setError(null);
                onRemove();
              }}
              disabled={disabled}
            >
              <Trash2 className="mr-1 h-4 w-4" /> Remove
            </Button>
          </div>
        </div>
      ) : (
        <div
          role="button"
          tabIndex={0}
          aria-label="Add a photo"
          onClick={openPicker}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              openPicker();
            }
          }}
          onDragOver={(e) => {
            e.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={onDrop}
          className={cn(
            "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-md border border-dashed p-6 text-sm text-muted-foreground",
            dragging && "border-primary bg-muted/50",
            error && "border-destructive",
          )}
        >
          <ImagePlus className="h-6 w-6" />
          <span>Drop a photo here, or click to choose one</span>
          <span className="text-xs">JPEG, PNG or WebP, up to 10 MB</span>
        </div>
      )}
      {draft.kind === "removed" && (
        <p className="text-sm text-muted-foreground">
          Photo will be removed when you save.{" "}
          <button type="button" className="text-primary underline" onClick={onUndo}>
            Undo
          </button>
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
```

- [ ] **Step 7: Wire `RecipeFormPage.tsx`**

1. Imports: change `import { useEffect, useState } from "react";` to `import { useEffect, useRef, useState } from "react";`. Extend the use-recipes import with `useUploadRecipeImage, useDeleteRecipeImage`. Add `import { RecipeImageField } from "@/components/features/recipes/recipe-image-field";` and `import { useRecipeImageDraft } from "@/lib/hooks/use-recipe-image-draft";`.

2. After `const updateRecipe = useUpdateRecipe();`, add:
```tsx
  const uploadImage = useUploadRecipeImage();
  const deleteImage = useDeleteRecipeImage();

  // Set when a create succeeded but the image step failed (FR-6): the page
  // then behaves as edit mode for that recipe without remounting, so the
  // staged File survives.
  const [persistedId, setPersistedId] = useState<string | null>(null);
  // Serialized attrs of the last successful recipe save; lets a retry skip
  // the recipe step when nothing changed.
  const lastSavedAttrs = useRef<string | null>(null);
  const recipeId = id ?? persistedId;
  const editing = recipeId !== null;

  const existingImage = existingData?.data?.attributes.image ?? null;
  const imageDraft = useRecipeImageDraft(existingImage);
  const imageBusy = uploadImage.isPending || deleteImage.isPending;
```

3. Replace the `try { … } catch { … }` block at the end of `onSubmit` with:
```tsx
    let savedId = recipeId;
    const serialized = JSON.stringify(attrs);
    try {
      if (!savedId) {
        const result = await createRecipe.mutateAsync(attrs);
        savedId = result.data.id;
        setPersistedId(savedId);
        lastSavedAttrs.current = serialized;
      } else if (lastSavedAttrs.current !== serialized) {
        await updateRecipe.mutateAsync({ id: savedId, attrs });
        lastSavedAttrs.current = serialized;
      }
    } catch (error) {
      toast.error(createErrorFromUnknown(error, `Failed to ${editing ? "update" : "create"} recipe`).message);
      return;
    }
    if (!savedId) return; // unreachable; narrows the type for the image step

    try {
      if (imageDraft.draft.kind === "staged") {
        await uploadImage.mutateAsync({ id: savedId, file: imageDraft.draft.file });
      } else if (imageDraft.draft.kind === "removed" && existingImage) {
        await deleteImage.mutateAsync(savedId);
      }
    } catch {
      toast.error("Recipe saved, but the photo failed to upload — try again.");
      return;
    }

    toast.success(editing ? "Recipe updated" : "Recipe created");
    navigate(`/app/recipes/${savedId}`);
```
Here `editing` reflects the state at render time, so a first-time create shows "Recipe created", and a retry after partial failure shows "Recipe updated".

4. Insert the Photo section between `<PlannerConfigForm … />` and the editor/preview grid:
```tsx
          <RecipeImageField
            draft={imageDraft.draft}
            previewSrc={imageDraft.previewSrc}
            onStage={imageDraft.stage}
            onRemove={imageDraft.remove}
            onUndo={imageDraft.undo}
            disabled={form.formState.isSubmitting}
          />
```

5. In the Preview column, make this the first child of the `rounded-md border p-4 bg-muted/30` div, above the derived-title block:
```tsx
                {imageDraft.previewSrc && (
                  <img
                    src={imageDraft.previewSrc}
                    alt=""
                    className="aspect-video w-full rounded-md object-cover"
                  />
                )}
```

6. Replace the submit button with:
```tsx
            <Button type="submit" disabled={form.formState.isSubmitting || imageBusy}>
              {(form.formState.isSubmitting || imageBusy) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {imageBusy ? "Uploading photo…" : editing ? "Update Recipe" : "Create Recipe"}
            </Button>
```

7. In the title (`isEditing ? "Edit Recipe" : "New Recipe"`), keep `isEditing`; the URL has not changed. The back/cancel buttons also keep `isEditing`. Cancel still just navigates. The draft hook's unmount cleanup revokes the preview URL, and no request is made (FR-8).

- [ ] **Step 8: Run the tests, the typecheck and lint**

Run: `cd frontend && npx vitest run src/lib/hooks src/components/features/recipes src/pages && npx tsc -b && npx eslint src/pages/RecipeFormPage.tsx src/components/features/recipes src/lib/hooks/use-recipe-image-draft.ts src/lib/recipe-image.ts`
Expected: PASS, with no lint errors.

- [ ] **Step 9: Commit**

```bash
git add frontend/src
git commit -m "feat(frontend): add recipe photo field with staged save and partial-failure retry"
```

---

## Task 13: Docs, benchmark record, manual compose acceptance, full gate

**Files:**
- Modify: `services/recipe-service/docs/rest.md`
- Modify: `services/recipe-service/docs/storage.md`
- Modify: `services/recipe-service/docs/domain.md`
- Modify: `services/recipe-service/README.md`
- Create: `docs/tasks/task-056-recipe-images/verification.md`

**Interfaces:**
- Consumes: everything above.
- Produces: documentation and the verification record.

- [ ] **Step 1: `docs/rest.md`**

Add a "Recipe images" section that matches the existing endpoint format in that file:
- `PUT /api/v1/recipes/{id}/image`: a multipart part named `file`. This is an intentional deviation from JSON:API request bodies. The response is a `recipe-images` JSON:API document (show the PRD §5.1 example). Error table: 400 not multipart / no `file` part; 404 recipe not visible; 413 file > 10 MiB or body > 10.5 MiB; 415 not JPEG/PNG/WebP; 422 > 40 MP or an edge < 16 px; 503 storage unavailable.
- `GET /api/v1/recipes/{id}/image?size=thumb|full`: `size` defaults to `full`, and any other value is 400. Document the headers (`Content-Type: image/jpeg`, `X-Content-Type-Options: nosniff`, `Cache-Control: private, max-age=31536000, immutable`, `ETag: "<imageId>-<size>"`). `If-None-Match` returns 304 without touching storage. 404 when there is no image, the recipe is not visible, or the object is missing; 503 when storage is down. The `v` query parameter is ignored server-side.
- `DELETE /api/v1/recipes/{id}/image`: 204, idempotent; 404 when the recipe is not visible.
- In the existing recipe list/detail/create/update/restore sections: add the nullable `image` attribute (`thumbUrl`, `fullUrl`, `width`, `height`; `null` when absent). Note that `image` in POST/PATCH bodies is ignored.

- [ ] **Step 2: `docs/storage.md`**

Add:
- The `recipe_images` table: columns, the unique `recipe_id`, the `(tenant_id, household_id)` index, and replace semantics (new `id` per replace, `created_at` carried forward).
- The `recipe_image_object_deletions` outbox: columns and drain semantics (5-minute loop with batches of 100, plus immediate post-commit drain; `ErrNotFound` counts as success; retries are unbounded; `attempts`/`last_error` exist for diagnosis).
- The object key layout and bucket requirements: private bucket; it must pre-exist in production; key scoped to `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` and `s3:ListBucket`.
- Memory bounds: semaphore of 2, ~340 MB worst case.
- The benchmark result from Task 3. Re-run it now and record the number with the machine description: `cd services/recipe-service && go test ./internal/recipeimage/imaging/ -run=^$ -bench=BenchmarkProcess12MP -benchtime=5x`.
- Backups: MinIO data is outside the Postgres backups, and backing up the bucket is the operator's responsibility.
- Known orphan path: a crash between `Put` and commit. Find it from the `storing recipe image renditions` log line with its `image_id`.

- [ ] **Step 3: `docs/domain.md`**

Add a `recipeimage` domain section. Cover:
- One optional image per recipe.
- Processing rules (formats; 40 MP / 16 px limits; orientation; the 1600/480 JPEG q82 renditions; white composite; metadata stripped).
- Lifecycle: soft delete keeps the image; restore brings it back; the retention purge deletes the row and enqueues objects; replace and delete go through the outbox.
- Visibility: tenant callback plus explicit household filter; another household gets 404.
- Storage-down behaviour.

- [ ] **Step 4: `README.md`**

Add the `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL` and `S3_REGION` rows to the configuration table, with their defaults and the note "empty `S3_ENDPOINT` disables image endpoints (503)".

- [ ] **Step 5: Commit the docs**

```bash
git add services/recipe-service/docs services/recipe-service/README.md
git commit -m "docs(recipe-service): document recipe image endpoints, storage and domain"
```

- [ ] **Step 6: Manual compose acceptance**

Start the stack in the background with `scripts/local-up.sh` (see `docs/tooling-conventions.md` for long-running processes; do not poll with `sleep`). Then check each item and record the observed result in `docs/tasks/task-056-recipe-images/verification.md`, one line each with PASS/FAIL and evidence such as an HTTP status, a header or an `mc ls` line:
1. `minio-init` exits 0 and the bucket exists: `docker compose -f deploy/compose/docker-compose.yml run --rm --entrypoint sh minio-init -c 'mc alias set local http://minio:9000 "$S3_ACCESS_KEY" "$S3_SECRET_KEY" >/dev/null && mc ls local/'`.
2. Through nginx, upload a JPEG just under 10 MiB to an existing recipe: 200. Upload an 11 MiB file: 413.
3. Create a recipe with a photo in the UI: the detail hero renders, and the list card shows a thumbnail.
4. Replace the photo: the new image shows without a hard refresh. Within about 5 minutes, `mc ls --recursive` no longer lists the old image id.
5. Remove the photo: the hero disappears and the card shows the placeholder.
6. A rotated phone photo (EXIF orientation 6) displays upright.
7. Request `GET …/image` again with `If-None-Match` set to the returned ETag: 304.
8. `docker compose … stop minio`: recipe list and CRUD work, the image endpoints return 503, and cards show placeholders. Start it again, and images return with no service restart.
9. Soft-delete a recipe with a photo, restore it, and the photo is intact.

Commit `verification.md`:
```bash
git add docs/tasks/task-056-recipe-images/verification.md
git commit -m "docs(task-056): record manual compose acceptance results"
```

- [ ] **Step 7: Full gate**

This step is handled by the task-verifier, not the implementer: the flagless `tools/verify.sh` must exit 0. The docker leg rebuilds only the recipe-service image, because no `shared/` files changed.

---

## Self-review notes

- **Spec coverage:** design §2/§3.5 → T1; §3.3/§5.3 → T2; §3.4/§5.1/§5.2 → T3; §4 → T4; §3.1/§3.2/§6.2 steps 4–7/§7 drainer/§13 → T5; §3.6/§6 → T6; §8 → T7; §7 purge → T8; §10 → T9; §11.1/§11.2 → T10; §11.4 card/detail → T11; §11.3/§11.4 field → T12; §14 manual/§15 → T13. PRD acceptance: client limits → T10/T12; server 413/415/422 → T6; renditions/EXIF → T3; thumbnails/placeholder → T11; lightbox → T11; 404/304 → T6; `image` on all responses with one batched query → T7; soft delete/restore/purge → T8 plus T13 item 9; MinIO down → T6 plus T13 item 8; partial failure → T12; compose → T9/T13; docs → T13; verify → T13.
- **Type consistency:** `Processor.WithStore`, `UploadResult.OutboxIDs`, `ObjectReaper.DrainKeys/Run`, `RecipeVisibility`, `URLs(recipeID, imageID)` (thumb first), and `ImageDraft` kinds `unchanged|staged|removed` are used identically across tasks.
