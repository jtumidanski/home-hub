# Recipe Images — Design

Task: `task-056-recipe-images`
PRD: [`prd.md`](prd.md) (approved) · Mockup: [`mockup.html`](mockup.html)
Status: Draft for review
Created: 2026-10-14

This document decides *how* the PRD gets built. It does not restate
requirements; FR/§ numbers refer to the PRD. Where this design deviates from
or tightens PRD wording, the deviation is called out explicitly in §12.

---

## 1. Grounding: what the code looks like today

Facts this design depends on, verified against source in this worktree:

| Fact | Where |
|---|---|
| recipe-service package layout uses `rest.go` for REST models/transforms and `resource.go` for route registration + handlers (inverse of the names some services use). New packages follow the recipe-service convention. | `services/recipe-service/internal/recipe/{rest,resource}.go` |
| Tenant scoping is automatic: a GORM callback injects `WHERE tenant_id = ?` on Query/Update/Delete for any model with a `TenantId` field **when the ctx carries a tenant**. Household is **not** auto-filtered; `recipe.getByID` filters only `id` + `deleted_at IS NULL`. | `shared/go/database/tenant_callbacks.go:31-60`, `recipe/provider.go:12` |
| Tenant + household come from the `access_token` cookie's JWT claims; `X-Tenant-ID` / `X-Household-ID` headers are only a fallback when the claim is nil. So a bare `<img src>` (cookie only, no custom headers) resolves the right tenant/household. | `shared/go/auth/auth.go:112-151` |
| Cross-domain enrichment is done by the recipe processor constructing sibling processors (`planner`, `normalization`) and calling batched `GetByRecipeIDs`. | `recipe/processor.go:380-420` |
| Where a dependency would cycle, it is injected at wiring time in `cmd/main.go` (e.g. `plan.NewRecipeValidator(db)` → `planitem.InitializeRoutes`). | `cmd/main.go` |
| Retention `Reap()` runs inside a reaper-owned transaction; dry-run is implemented by the reaper returning a sentinel to roll the tx back. There is **no post-commit hook**. | `shared/go/retention/reaper.go:103-152` |
| `/metrics` is already mounted (default Prometheus registry via `promhttp.Handler()`), so `promauto` collectors in recipe-service are exported with no extra wiring. | `recipe-service/internal/retention/wire.go`, `shared/go/retention/metrics.go:72` |
| Shared server sets `ReadTimeout` and `WriteTimeout` to 15 s. | `shared/go/server/server.go:66-67` |
| Runtime image is `alpine` with a `CGO`-free `go build`. | `services/recipe-service/Dockerfile` |
| No k8s resource requests/limits on recipe-service. | `deploy/k8s/recipe-service.yaml` |
| Frontend `api.put` always JSON-stringifies with `Content-Type: application/vnd.api+json`; there is no multipart path. PUT/PATCH use `maxRetries: 0`. | `frontend/src/lib/api/client.ts:330-349` |
| `recipeKeys.lists` / `recipeKeys.detail` exist; mutations invalidate them in `onSuccess`. Toasts are `sonner`. | `frontend/src/lib/hooks/api/use-recipes.ts`, `pages/RecipeFormPage.tsx` |
| nginx has a single prefix `location /api/v1/recipes` block (covers `/recipes/{id}/image`). | `deploy/compose/nginx.conf:181` |

---

## 2. Architecture overview

```
                 ┌──────────────────────── recipe-service ─────────────────────────┐
 browser ──PUT──▶│ recipeimage.rest (resource.go)                                  │
  <img> ──GET──▶ │   ├─ RecipeVisibility (injected from recipe pkg)                │
                 │   ├─ imaging  (decode → limit → resize → orient → JPEG)         │
                 │   ├─ processor (orchestrates; owns tx + outbox)                 │
                 │   └─ objectstore.Store (interface) ──▶ minio-go ──▶ MinIO       │
                 │                                                                  │
                 │ recipe.rest  ──enrich──▶ recipeimage.Processor.GetByRecipeIDs    │
                 │ retention    ──tx────▶  recipeimage.PurgeByRecipeIDsTx (DB only)│
                 │ recipeimage.ObjectReaper (loop) ── drains outbox ──▶ Store.Delete│
                 └──────────────────────────────────────────────────────────────────┘
```

New units, each with one job:

| Unit | Job | Depends on |
|---|---|---|
| `internal/objectstore` | Thin S3 abstraction: `Put`, `Get`, `Delete`, `Ping`. Only importer of `minio-go`. | minio-go |
| `internal/objectstore/objectstoretest` | In-memory `Store` fake with failure injection, for tests in other packages. | — |
| `internal/recipeimage` | Domain package: entity, model, builder, provider, administrator, processor, REST models (`rest.go`), handlers (`resource.go`), outbox reaper. | `objectstore.Store`, `RecipeVisibility` (injected) |
| `internal/recipeimage/imaging` | Pure functions: bytes in → two JPEG renditions out, or a typed error. No I/O, no DB. | stdlib, `golang.org/x/image` |
| `internal/recipeimage/imaging/orientation.go` | Minimal EXIF orientation reader (§5.3). | stdlib |

Dependency direction: `recipe → recipeimage → objectstore`; `retention →
recipeimage`; `recipeimage` never imports `recipe` (the visibility check is
injected, mirroring the `plan.NewRecipeValidator` → `planitem` pattern).

**No `shared/` changes.** Every piece lives in recipe-service, so the flagless
`verify.sh` docker leg rebuilds only the recipe-service image. This is also
why retention cleanup is solved inside recipe-service (§7) rather than by
adding a post-commit hook to `shared/go/retention`.

---

## 3. Key decisions and alternatives

### 3.1 How object deletions survive transactions → transactional outbox

The constraint: retention deletes happen inside a reaper-owned tx with no
post-commit hook, and dry-run works by rollback. Replace/delete also want
"DB first, objects after commit" (FR-20/21).

| Option | Verdict |
|---|---|
| **A. Outbox table `recipe_image_object_deletions`** written in the same tx as the row delete; a drainer deletes objects after commit. | **Chosen.** Dry-run is free (rollback discards outbox rows). Retention stays DB-only and needs no `Store`. Failed deletes are retried instead of silently orphaned. One code path for FR-19/20/21. |
| B. Delete objects inside `Reap()` before commit | Rejected. Must special-case dry-run; a later tx failure leaves rows pointing at deleted objects. |
| C. Add an after-commit callback to `shared/go/retention` | Rejected. Shared change → all 12 images rebuild; widens a shared contract for one consumer. |
| D. Periodic full sweep comparing bucket keys to `recipe_images` | Rejected for v1. Needs bucket listing per tenant, is O(objects), and still needs a grace window for in-flight uploads. Outbox makes it unnecessary for every case except the crash window (§9). |

The outbox is strictly stronger than the PRD's "best effort, logged on
failure": request paths still try the delete immediately after commit (so the
common case is synchronous-ish), and the drainer retries what failed.

### 3.2 Replace semantics → `DELETE … RETURNING` + `INSERT` in one tx

PRD §6.1 says "updates the row in place with a new id … single upsert". We
need the *old* keys inside the tx to enqueue them, and we need correctness
under two concurrent replaces of the same recipe.

| Option | Verdict |
|---|---|
| `INSERT … ON CONFLICT (recipe_id) DO UPDATE SET id = excluded.id, …` | Loses the old keys (no reliable `RETURNING` of pre-update values). Would need a prior `SELECT … FOR UPDATE`, which the SQLite test driver can't render. |
| `SELECT` then upsert, no lock | Race: two replacers both read key K0, one set of new keys is orphaned. |
| **`DELETE FROM recipe_images WHERE recipe_id = ? AND tenant/household … RETURNING full_key, thumb_key`, then `INSERT` the new row, then insert outbox rows for any returned keys** | **Chosen.** In Postgres the second concurrent DELETE blocks on the row lock, then deletes the winner's freshly inserted row and returns *its* keys → they go to the outbox. Nothing is orphaned. Works on SQLite ≥ 3.35 (GORM `clause.Returning{}`) for tests. |

The only remaining race is "no prior row, two concurrent first uploads":
both DELETEs return nothing, one INSERT hits the `recipe_id` unique index.
The processor retries the whole tx **once** on a unique violation (the retry's
DELETE now returns the winner's keys). Observable behaviour matches the PRD:
one row, new `id` per replace.

### 3.3 EXIF orientation → minimal in-repo parser (resolves PRD OQ3)

| Option | Verdict |
|---|---|
| `github.com/rwcarlsen/goexif` | Rejected. Unmaintained since 2019, parses every tag (much larger attack surface on untrusted input) to read one SHORT. |
| `github.com/dsoprea/go-exif/v3` | Rejected. Heavy dependency tree for one tag. |
| **In-repo `orientation.go` (~120 lines)** | **Chosen.** Walks JPEG markers to the first `APP1` with `Exif\0\0`, reads the TIFF byte order, scans IFD0 for tag `0x0112`. Every offset is bounds-checked; any malformation returns orientation `1`. Covered by table tests and a native Go fuzz target. |

### 3.4 Image processing → pure Go, `golang.org/x/image`

- Decoders: stdlib `image/jpeg`, `image/png`, `golang.org/x/image/webp`
  (registered via blank imports **only** inside `imaging`, so no other format,
  such as GIF or BMP, is accepted by accident).
- Resampler: `golang.org/x/image/draw.CatmullRom` for `full`;
  `thumb` is scaled **from the `full` rendition**, not from the original
  (≈ 11× fewer source pixels for a 12 MP photo, visually identical at 480 px).
- Encoder: stdlib `image/jpeg`, quality 82. JPEG-only output confirmed (PRD
  OQ4): no cgo, `alpine` runtime unchanged.
- Rejected: `disintegration/imaging` (unmaintained, adds nothing over
  `x/image/draw`); libvips/`bimg` (cgo, Dockerfile changes, PRD non-goal).

### 3.5 Object store client → `minio-go/v7` behind an interface

`objectstore.Store` is an interface so `recipeimage` and `retention` tests use
the in-memory fake. The minio implementation is the only importer of
`minio-go`. Not shared (PRD non-goal). Rejected: AWS SDK v2 (larger, needs
path-style + endpoint resolver gymnastics for MinIO).

### 3.6 Upload transport → custom handler, not `RegisterInputHandler`

`server.RegisterInputHandler[T]` unmarshals JSON:API bodies. The upload
handler is registered with `server.RegisterHandler` (as GET/DELETE handlers
are) and parses multipart itself. The response is still JSON:API, via
`server.MarshalResponse[RestModel]`.

---

## 4. Data model

### 4.1 `recipe.recipe_images` (PRD §6.1, unchanged columns)

`recipeimage.Entity` with GORM tags:

- `Id uuid` PK
- `TenantId`, `HouseholdId` → `index:idx_recipe_image_tenant_household`
- `RecipeId` → `uniqueIndex:idx_recipe_image_recipe`
- `FullKey`, `ThumbKey` `varchar(512) not null`
- `ContentType varchar(50) not null`, `Width`, `Height`, `FullBytes`, `ThumbBytes int not null`
- `CreatedAt`, `UpdatedAt not null`

`CreatedAt` is carried forward from the deleted row on replace, so it means
"first image added", and `UpdatedAt` means "last replaced". `updatedAt` in the
REST model is the row's `UpdatedAt`.

No FK constraint to `recipes`: the existing recipe tables don't use DB-level FKs
for child rows except tags (`OnDelete:CASCADE`), and retention deletes children
explicitly. Matches the `planner`/`normalization` children.

### 4.2 `recipe.recipe_image_object_deletions` (new: outbox)

| Column | Type | Notes |
|---|---|---|
| id | UUID | PK |
| tenant_id | UUID | NOT NULL (audit and log context; not used for filtering by the drainer) |
| household_id | UUID | NOT NULL |
| object_key | VARCHAR(512) | NOT NULL |
| attempts | INT | NOT NULL, default 0 |
| last_error | TEXT | NULL |
| created_at | TIMESTAMP | NOT NULL, indexed (drain order) |

Both tables are migrated by `recipeimage.Migration`, which is added to
`database.SetMigrations(...)` in `cmd/main.go`.

### 4.3 Object keys (PRD §6.2)

`{tenantId}/{householdId}/recipes/{recipeId}/{imageId}/full.jpg` and
`…/thumb.jpg`, built by one function `objectKeys(tenantID, householdID,
recipeID, imageID uuid.UUID) (full, thumb string)` in `recipeimage`. Every input
is a `uuid.UUID`, so user input cannot reach a key by construction.

---

## 5. Processing pipeline (`imaging`)

Single entry point:

```go
type Rendition struct { Bytes []byte; Width, Height int }
type Result struct { Full, Thumb Rendition }

func Process(src []byte) (Result, error)
// errors: ErrUnsupportedFormat (→415), ErrTooLarge (→422), ErrTooSmall (→422)
```

### 5.1 Steps

1. `image.DecodeConfig(bytes.NewReader(src))`. A decode error, or a format
   not in {jpeg, png, webp}, returns `ErrUnsupportedFormat`. Type comes only
   from the bytes (FR-22).
2. `w*h > 40_000_000` → `ErrTooLarge`; `min(w,h) < 16` → `ErrTooSmall`. This
   runs **before** the full decode (FR-23).
3. `image.Decode`.
4. Compute `full` target size: fit longest edge ≤ 1600, never upscale. The
   orientation (step 6) may swap axes, so the fit uses the *oriented*
   dimensions and the scale step uses the matching un-oriented target.
5. Scale onto a white-filled `*image.RGBA` with `draw.CatmullRom.Scale(…,
   draw.Over, nil)`. This composites transparency onto white in the same pass
   as the resize (FR-25). If no downscale is needed, use `draw.Draw` onto
   white instead (one copy, no resample).
6. If the format is JPEG, read the orientation (§5.3) and apply the 8-case
   transform to the **already-downscaled** image. Rotating 1600 px instead of
   4000 px is the main CPU saving; the result is identical because the
   resize and the orientation commute once the target axes are swapped.
7. `thumb` = scale of the oriented `full` to longest edge ≤ 480 (never
   upscale), `draw.CatmullRom`.
8. `jpeg.Encode(…, &jpeg.Options{Quality: 82})` for both. The re-encode
   carries no EXIF, XMP or ICC data (FR-25, NFR Security).

### 5.2 Memory and CPU bounds

- A 40 MP PNG decodes to an NRGBA buffer of about 160 MB, and the 12 MP JPEG
  case to a YCbCr buffer of about 18 MB. recipe-service has no k8s memory
  limit, but unbounded parallel uploads could still exhaust the node.
- `recipeimage` holds a package-level **semaphore of 2** around `Process`.
  Acquisition honours the request ctx: if the client goes away it returns
  without processing. No 503 on saturation, because a household has a
  handful of users, and queueing is the better UX.
- The worst-case resident size is about 2 × (160 MB + 1600² × 4 B ≈ 10 MB),
  which is roughly 340 MB. This is documented in `services/recipe-service/docs/storage.md`. Adding
  k8s requests/limits is out of scope (no other service sets them).
- **Budget check (PRD NFR: 12 MP < 2 s):** `imaging` ships a Go benchmark on a
  committed 4000×3000 JPEG fixture (orientation 6). The plan runs it once
  during implementation and records the number in `services/recipe-service/docs/storage.md`. If it
  is over budget, the fallback is a two-stage resize (`draw.ApproxBiLinear`
  to 2× the target, then `CatmullRom`). Use the fallback only if needed; it
  is not built speculatively.

### 5.3 Orientation reader contract

```go
func jpegOrientation(src []byte) int // 1..8; 1 on absent/malformed
```

- It scans markers from SOI and stops at SOS (`0xFFDA`). Only the first
  `APP1` (`0xFFE1`) whose payload starts with `Exif\x00\x00` is examined.
- TIFF header: `II*\0` or `MM\0*`. It follows the IFD0 offset and reads the
  `uint16` entry count. For each 12-byte entry it matches tag `0x0112`,
  type `SHORT(3)`, count `1`, and returns the value if it is in 1..8.
- Every slice access is guarded by an explicit length check. The fuzz
  target (`FuzzJPEGOrientation`) asserts no panic and that the result is in
  1..8.

---

## 6. Request flows (`recipeimage` processor + handlers)

### 6.1 Injected visibility check

```go
// in recipeimage
type RecipeVisibility func(ctx context.Context, recipeID uuid.UUID) error // nil | ErrRecipeNotFound

// in recipe (exported, constructed in main.go)
func ImageVisibility(db *gorm.DB) recipeimage.RecipeVisibility
```

`recipe.ImageVisibility` uses `getByID` (tenant callback + `deleted_at IS
NULL`) **and** compares `HouseholdId` with `tenant.MustFromContext(ctx)`.
A mismatch returns `ErrRecipeNotFound`, which gives the 404-not-403
requirement (PRD §5). It returns only an error, never the recipe model, so
`recipeimage` stays decoupled from `recipe.Model`.

Every `recipeimage` provider query also adds an explicit
`household_id = ?` (the tenant part comes from the callback). This meets the
PRD NFR ("every DB query on `recipe_images` filters by `tenant_id` and
`household_id`") independently of the visibility check.

### 6.2 `PUT /recipes/{id}/image`

1. `ParseID("id")`, `visibility(ctx, id)` → 404.
2. Extend the per-request read deadline to 60 s with
   `http.NewResponseController(w).SetReadDeadline` (§10.3).
3. `r.Body = http.MaxBytesReader(w, r.Body, 10 MiB + 512 KiB)`. Content-Type
   must be `multipart/form-data` with a boundary, otherwise 400. A
   `multipart.Reader` walks the parts. It reads the first part named `file`
   through `io.LimitReader(part, 10 MiB + 1)`. More than 10 MiB, or a
   `*http.MaxBytesError` → 413. No `file` part → 400. `ParseMultipartForm` is
   not used, so nothing spills to temp files.
4. `imaging.Process` under the semaphore → 415 / 422. Records
   `recipe_image_processing_seconds`.
5. `imageID := uuid.New()`, build the keys. `store.Put(full)`, then
   `store.Put(thumb)`, each with a 30 s ctx timeout. If either fails, it
   deletes whatever was put (best effort, direct, not outbox) and returns 503.
6. Tx (§3.2): `DELETE … RETURNING`, `INSERT` the new row (carrying forward
   `created_at` if a row was returned), and insert outbox rows for the
   returned keys. On a unique violation, retry the tx once. On any other
   failure, best-effort delete the new objects and return 500.
7. After commit, `go reaper.DrainKeys(context.Background(), outboxIDs)`
   attempts the deletes now (10 s timeout). Failures stay in the outbox.
8. 200 with a `recipe-images` JSON:API document (PRD §5.1). Increments
   `recipe_image_uploads_total{result="created"|"replaced"}`.

Each failure path increments `recipe_image_uploads_total{result=…}` with one
of: `rejected_size`, `rejected_type`, `rejected_dimensions`,
`storage_error` or `db_error`.

### 6.3 `GET /recipes/{id}/image?size=thumb|full`

1. `size` must be `""`, `full` or `thumb`; any other value is 400.
2. `visibility` → 404. The provider loads the row by `recipe_id` (scoped) →
   404 if there is none.
3. `etag := "\"<imageId>-<size>\""`. If `If-None-Match` matches (a simple
   comma-split compare, which also accepts `*`), return **304 without
   touching storage**. Repeat views therefore work during a MinIO outage.
4. `store.Get(ctx, key)`. `objectstore.ErrNotFound` means the DB row points
   at a missing object: log it at error and return 404. Any other error →
   503 and `recipe_image_storage_errors_total{op="get"}`.
5. Headers: `Content-Type: image/jpeg`, `Content-Length` from the object
   info, `X-Content-Type-Options: nosniff`, `Cache-Control: private,
   max-age=31536000, immutable`, `ETag`. Then `io.Copy(w, obj)`.

The `v` query parameter is ignored server-side. It exists only to make the
URL unique per image id. A stale `v` gets the current bytes, which is
harmless because the frontend never renders a stale URL after its queries
are invalidated.

### 6.4 `DELETE /recipes/{id}/image`

`visibility` → 404. A tx does `DELETE … RETURNING` plus outbox inserts.
After commit it runs `DrainKeys` in a goroutine and returns 204, including
when no row existed (idempotent, PRD §5.3).

### 6.5 Error document shape

Errors use the existing `server.WriteError(w, status, title, detail)` helper,
so they match the rest of recipe-service.

---

## 7. Retention and lifecycle

- **Soft delete / restore (FR-18):** these paths are untouched. The image
  row persists. `restoreHandler` already builds the detail enrichment, which
  now includes `image`.
- **Purge (FR-19):** `cascadeDeleteRecipes(tx, ids)` gains one call before
  the final `recipes` delete:

  ```go
  n, err := recipeimage.PurgeByRecipeIDsTx(tx, ids) // DELETE … RETURNING keys; INSERT outbox rows
  total += n // image rows count toward Deleted, like tags/ingredients
  ```

  The reaper passes `tx` already scoped by explicit `tenant_id`/`household_id`
  predicates on the recipe ids, so `PurgeByRecipeIDsTx` filters by
  `recipe_id IN ?` only, like its siblings. In dry-run mode the reaper rolls
  back the whole tx, outbox rows included, so dry-run counts images and
  deletes nothing (PRD §7). The cascade comment
  (`recipes → tags → …`) is updated.
- **Outbox drainer:** `recipeimage.ObjectReaper` with `Run(ctx)` for
  periodic draining and `DrainKeys(ctx, ids)` for targeted draining.
  - It is started in `main.go` with
    `go sr.Loop(ctx, l, 5*time.Minute, reaper.Run)`, reusing the shared
    loop helper, which is a call and not a contract change.
  - Each tick it selects up to 100 rows ordered by `created_at`, calls
    `store.Delete` for each, and treats `ErrNotFound` as success. On
    success it deletes the outbox row. On failure it sets
    `attempts++, last_error`, logs at warn, and increments
    `recipe_image_storage_errors_total{op="delete"}`.
  - The drainer runs with a tenant-less ctx, so the tenant callback is a
    no-op. Retries are unbounded, but the 5-minute tick keeps load trivial;
    `attempts` exists for diagnosis.
  - Concurrent drainers (the targeted goroutine and the tick) are safe:
    deleting an already-deleted object is `ErrNotFound`, which counts as
    success, and deleting an already-deleted outbox row is a no-op.

---

## 8. Recipe resource changes (PRD §5.4)

- New REST type in `recipe/rest.go`:

  ```go
  type RestImageModel struct {
      ThumbURL string `json:"thumbUrl"`; FullURL string `json:"fullUrl"`
      Width int `json:"width"`; Height int `json:"height"`
  }
  ```

  `RestModel` and `RestDetailModel` gain `Image *RestImageModel
  \`json:"image"\``, with **no `omitempty`**, so a missing image serializes
  as `null`.
- `ListEnrichment` / `DetailEnrichment` gain `Image *RestImageModel`.
  `BuildListEnrichments` adds one batched
  `recipeimage.NewProcessor(...).GetByRecipeIDs(ids)` call, which is the
  third batched query next to normalization and planner. Its doc comment is
  updated from "exactly two extra queries" to "three". `BuildDetailEnrichment`
  adds a single `GetByRecipeID`. A lookup error is logged and leaves
  `Image: nil`, the same degradation policy as planner and normalization.
- The URL builder lives in `recipeimage`:
  `func URLs(recipeID, imageID uuid.UUID) (thumb, full string)`, which
  returns `/api/v1/recipes/{rid}/image?size=…&v={iid}`. Both the
  `recipe-images` response and the `recipe` enrichment use it, so the two
  can't drift. The `/api/v1` prefix is hard-coded on purpose: it is the
  public path under both compose nginx and the k8s ingress, and recipe-service
  already mounts there.
- `CreateRequest` / `UpdateRequest` are unchanged. They have no `image`
  field, so a client-supplied `image` attribute can never be bound or
  persisted. The plan adds a test that a PATCH carrying `image` doesn't
  change the stored image, whatever api2go does with unknown attributes.
- Create, update, get and restore all flow through `TransformDetail`, so
  `image` appears on every response listed in §5.4 with no per-handler code.

---

## 9. Failure modes and resilience

| Situation | Behaviour |
|---|---|
| `S3_ENDPOINT` unset | `objectstore.New` returns a `disabled` Store whose every method returns `ErrUnavailable`. Image endpoints → 503. Everything else works. A startup log line says the store is disabled. |
| MinIO down at startup | `main.go` calls `store.Ping(ctx)` (bucket exists, 5 s timeout) in a goroutine and logs an error. Startup is never blocked, and the status is **not sticky**: each request just tries the op, so the service recovers on its own when MinIO comes back. `/readyz` is unaffected. |
| MinIO down at runtime | PUT → 503, after the objects are cleaned up. GET → 503, but 304 still works. DELETE → 204, because the DB row is removed and the outbox retries the objects later. List/detail JSON is unchanged, still from Postgres, and the frontend `onError` shows the placeholder. |
| Crash between `Put` and commit | Up to two orphaned objects under an image id that no row references. **Accepted for v1** (PRD OQ2). This is the only orphan path left, and it is logged with the image id at `Put` time so an operator can find it. |
| DB row exists but object missing | GET returns 404 and logs at error. The frontend shows the placeholder, and a replace or delete repairs it. |
| Two concurrent replaces | Last commit wins. The loser's objects are enqueued for deletion (§3.2). |

---

## 10. Configuration and deploy

### 10.1 Config

`config.Config.ObjectStore objectstore.Config` with fields `Endpoint`
(`S3_ENDPOINT`, default `""` = disabled), `Bucket` (`S3_BUCKET`,
`home-hub-recipes`), `AccessKey`, `SecretKey`, `UseSSL` (`S3_USE_SSL`,
`false`) and `Region` (`S3_REGION`, `us-east-1`). minio-go is configured
with path-style lookup (`BucketLookupPath`) so it works with MinIO behind
a plain hostname.

### 10.2 Compose (`deploy/compose/docker-compose.yml`)

- `minio`: `minio/minio:<pinned tag>` (Renovate manages the pin),
  `command: server /data --console-address :9001`, a named volume
  `minio-data`, a healthcheck on `/minio/health/live`, `expose: 9000`, and
  no host ports by default.
- `minio-init`: `minio/mc:<pinned tag>`, one-shot
  `mc alias set … && mc mb --ignore-existing local/$S3_BUCKET`, with
  `depends_on: minio: condition: service_healthy`.
- recipe-service: `S3_*` env vars, with compose defaults
  (`${S3_ACCESS_KEY:-homehub}`, `${S3_SECRET_KEY:-homehub-dev-secret}`) so
  `scripts/local-up.sh` works with an unchanged `.env`. Also
  `depends_on: minio-init: condition: service_completed_successfully`.
- The MinIO root user and password reuse the same `S3_ACCESS_KEY` /
  `S3_SECRET_KEY` values. This is acceptable for local dev; production uses
  a scoped key (§10.4).
- `.env.example`: the `S3_*` entries, documented.

### 10.3 Proxies and timeouts

- `nginx.conf`: add `client_max_body_size 11m;` to the
  `location /api/v1/recipes` block only.
- `ingress.yaml`: change `proxy-body-size` from `"10m"` to `"11m"`. This
  annotation is ingress-wide, so every route gets the slightly larger
  limit. That is harmless, and splitting the ingress for one path isn't
  worth it.
- **Slow uplinks vs the 15 s `ReadTimeout`:** both nginx and ingress-nginx
  buffer request bodies by default (`proxy_request_buffering on`), so the
  proxy absorbs a slow phone upload and forwards it to the service at LAN
  speed. As a second guard, the upload handler raises its own read
  deadline to 60 s via `http.ResponseController` (§6.2 step 2). That is a
  per-handler change, so shared `server.go` is untouched. Neither nginx
  config disables request buffering today, and the plan must not add
  `proxy_request_buffering off`.

### 10.4 k8s

- `recipe-service.yaml`: `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY` and
  `S3_SECRET_KEY` from a new secret `object-store-credentials`. `S3_USE_SSL`
  and `S3_REGION` are plain values.
- `secrets.example.yaml`: an `object-store-credentials` stanza.
- **PRD OQ1:** this design follows the PRD assumption. Production uses an
  existing S3/MinIO endpoint configured by env, and no MinIO Deployment/PVC
  is added to `deploy/k8s/`. The bucket must exist beforehand; the service
  only checks for it and never creates it, so the production key can be
  scoped to `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` and
  `s3:ListBucket` on that one bucket. `services/recipe-service/docs/storage.md` records the required
  policy and the backup note from PRD §8.

---

## 11. Frontend design

### 11.1 Transport

- `lib/api/client.ts`: a new `upload<T>(path, form: FormData, options)` →
  `PUT`, `credentials: "include"`. Headers come from
  `buildHeaders(undefined, options)`; no `Content-Type` is set, so the
  browser adds the multipart boundary. Tenant headers are still added for
  parity. `maxRetries: 0` and the same 401 → refresh → retry path as `put`.
  Errors go through `handleError`, so `createErrorFromUnknown` works
  unchanged.
- `services/api/recipe.ts`: `uploadRecipeImage(tenant, id, file)` builds a
  `FormData` with `file` and calls `api.upload`; it returns
  `ApiResponse<RecipeImageResource>`. `deleteRecipeImage(tenant, id)` calls
  `api.delete`.
- `use-recipes.ts`: `useUploadRecipeImage()` and `useDeleteRecipeImage()`.
  `onSuccess` invalidates `recipeKeys.lists` and
  `recipeKeys.detail(tenant, household, id)`. The new `image.*Url` values
  carry a new `v`, so the `<img>` re-fetches without a hard refresh
  (acceptance criterion 2).

### 11.2 Types

`types/models/recipe.ts`: `export interface RecipeImage { thumbUrl: string;
fullUrl: string; width: number; height: number }`, plus `image: RecipeImage |
null` on the list and detail attribute interfaces, and a
`RecipeImageResource` for the PUT response.

### 11.3 Form state and save sequencing (FR-4 to FR-9)

The image draft is a small discriminated union in a hook
`useRecipeImageDraft(existing: RecipeImage | null)`:

```ts
type ImageDraft =
  | { kind: "unchanged" }                      // keep server state
  | { kind: "staged"; file: File; previewUrl: string }
  | { kind: "removed" };                       // staged removal (Undo → unchanged)
```

- The hook owns the `URL.createObjectURL` / `revokeObjectURL` lifecycle,
  revoking on replace, on clear and on unmount.
- `previewSrc` is derived: `staged` → `previewUrl`; `unchanged` →
  `existing?.fullUrl`; `removed` → `null`. The Photo section and the
  live-preview column both read it.

Save is a single async function in `RecipeFormPage`:

```
savedId = persistedId ?? id
if !savedId || formDirtySinceLastSave:  save recipe (POST or PATCH) → savedId; persistedId = savedId
if draft.kind === "staged":  await upload(savedId, file)
if draft.kind === "removed" && hadExisting:  await deleteImage(savedId)
navigate(`/app/recipes/${savedId}`)
```

- **Partial failure (FR-6):** if the recipe step succeeded and the image
  step throws, set `persistedId`, keep the draft, show the PRD toast, and
  stay on the page. The page then treats `persistedId` like `id` (edit
  mode: PATCH, not a second POST). Pressing Save again sees a clean form
  and retries only the image step. If the user edited fields in between,
  the form is dirty and Save PATCHes first.
- The route is deliberately **not** changed to `/edit` on partial failure,
  because remounting would drop the staged `File`.
- Button label/state (FR-7): it shows "Uploading photo…" and is disabled
  while the image mutation is pending. It keeps the existing label/pending
  behaviour for the recipe step.
- Cancel (FR-8) navigates away as today. The draft is discarded on unmount,
  and no request is made.

### 11.4 Components

- `components/features/recipes/recipe-image-field.tsx`: a presentational
  component driven by the draft hook. It has a drop zone (native
  `onDragOver`/`onDrop` plus a hidden `<input type="file" accept=…>`;
  **no `react-dropzone` dependency**), a preview, Replace / Remove / Undo,
  and the inline validation error. Validation is a pure
  `validateRecipeImageFile(file): string | null` (type set + `≤ 10 * 1024 *
  1024`), exported for tests.
- `components/features/recipes/recipe-image.tsx`:
  `<RecipeImage image variant="hero" | "card" alt />`. It renders
  `<img>` with the right rendition, `loading`, and aspect classes, and keeps
  local `failed` state for the `onError` → placeholder fallback (FR-17).
  The placeholder is a muted box with a centred lucide `ChefHat`. With
  `variant="hero"` and `image == null` it renders nothing (FR-12).
- `recipe-card.tsx`: **one** `<img>` with responsive Tailwind classes
  (`flex-row` with a `size-16` rounded thumb below `md`; `flex-col` with an
  `aspect-video` top image at `md+`). That avoids two elements and two
  requests per card (FR-14/15/16).
- `RecipeDetailPage.tsx`: the hero above the title row (`aspect-[4/3]
  -mx-4` edge-to-edge on mobile, `md:aspect-video md:rounded-lg` at
  max-w-5xl), `loading="eager"`. Clicking it opens the existing shadcn
  `dialog.tsx` with `object-contain` on the `full` rendition (FR-10/11).
  Esc, backdrop and close come from the Dialog primitive.

### 11.5 Tests (Vitest + RTL, matching `__tests__` placement)

- `validateRecipeImageFile`: type and size boundaries.
- `useRecipeImageDraft`: transitions, plus revoke calls with
  `URL.createObjectURL` mocked.
- `RecipeFormPage` save sequencing, with the service mocked: create+upload;
  edit+replace; edit+remove; remove→undo (no request); cancel (no request);
  **partial failure** (toast shown, stays on page, the second Save calls
  only upload, and no second POST).
- `RecipeImage`/`recipe-card`: with an image, with `null` (placeholder), and
  the `onError` fallback.
- `RecipeDetailPage`: no hero when `null`; the hero opens the lightbox.
- `services/api/__tests__/recipe.test.ts`: `uploadRecipeImage` sends
  `FormData`, sets no JSON content type, and uses PUT.

---

## 12. Deviations from PRD wording (explicit)

1. **Replace mechanics (§6.1, "updated in place … single statement"):** this
   design uses `DELETE … RETURNING` plus `INSERT` in one tx (§3.2). The
   external behaviour is identical: one row per recipe, a new `id` per
   replace, `created_at` carried forward. The change is what makes the old
   keys recoverable and concurrent replaces orphan-free.
2. **Object deletion durability (FR-19/20/21, "best effort, logged"):** this
   design uses an outbox with retries (§3.1). It is a superset of the PRD
   behaviour, and failures are still logged.
3. **Request limit interpretation (§5.1/FR-23 "10 MiB"):** the *file part*
   limit is exactly 10 MiB. The *request body* allows 10 MiB + 512 KiB of
   multipart overhead, and both produce 413. This matches the 11m proxy
   limit's intent.
4. **Missing object behind a row** returns 404 (not 503). It is a data
   inconsistency, not an outage, and retrying won't help.

---

## 13. Observability

All collectors are `promauto` in `recipeimage/metrics.go` on the default
registry, already exported at `/metrics`:

- `recipe_image_uploads_total{result}`. `result` is one of `created`,
  `replaced`, `rejected_size`, `rejected_type`, `rejected_dimensions`,
  `storage_error` or `db_error`.
- `recipe_image_processing_seconds`: a histogram with buckets
  `0.05 … 5`.
- `recipe_image_storage_errors_total{op}`: `op` is `put`, `get`, `delete`
  or `ping`.

Structured logs (logrus fields) carry `recipe_id`, `image_id`, `full_bytes`,
`thumb_bytes`, `source_bytes`, `processing_ms` and `storage_ms` on upload,
replace and delete; `object_key` and `attempts` on drainer failures; and
`tenant_id`/`household_id`, which are already added to the request ctx by
the auth middleware.

---

## 14. Testing strategy (backend)

| Layer | Tests |
|---|---|
| `imaging` | Table tests over committed small fixtures: JPEG orientations 1/3/6/8 (asserts output dims swap and a corner-pixel colour lands in the right place); PNG with alpha (asserts a white composite); WebP; a GIF → `ErrUnsupportedFormat`; garbage bytes; a crafted header claiming 10000×5000 (→ `ErrTooLarge`, with a tiny file size proving no full decode); 8×8 → `ErrTooSmall`; small images are not upscaled; output decodes as JPEG with no APP1 segment. A fuzz target for `Process`, plus a benchmark (§5.2). |
| `orientation.go` | Table tests for II/MM, a missing APP1, truncation at every offset, wrong type/count, and values out of range; `FuzzJPEGOrientation`. |
| `objectstore` | Fake-backed contract tests for `objectstoretest`. The minio implementation is thin and is exercised by the compose acceptance run, not by unit tests (no testcontainers in this repo). |
| `recipeimage` processor | SQLite + fake store: create; replace (old keys in the outbox, new id, created_at kept); delete (idempotent); a Put failure cleans up and creates no row; a commit failure cleans up new objects; the unique-violation retry path; household isolation (another household's recipe → `ErrRecipeNotFound`); the outbox drainer (success, failure → attempts++, NotFound counts as success). |
| `recipeimage` handlers | `httptest`: 400 (not multipart / no `file` / bad `size`), 413 (part > 10 MiB, body > cap), 415, 422, 503 (disabled store), 200 shape, GET headers, 304 on `If-None-Match` with a store that would fail (proves no storage call), 404 cross-household. |
| `recipe` | `BuildListEnrichments` issues exactly one `recipe_images` query for N recipes (asserted by a test-local GORM `Before("gorm:query")` callback that counts statements against the `recipe_images` table; extend `recipe/batch_test.go`); `image: null` vs populated in list/detail JSON. |
| `retention` | Extend `handlers_test.go`: a purge removes the `recipe_images` rows and writes outbox rows; dry-run leaves both untouched. |

The PRD acceptance items that need real infrastructure (a 10 MiB upload
through compose nginx, MinIO stopped → 503, purge removes objects from
MinIO) are verified manually against `scripts/local-up.sh` and recorded in
the task's verification notes. Flagless `tools/verify.sh` remains the gate.

---

## 15. Documentation

- `services/recipe-service/docs/rest.md`: the three image endpoints, the
  `image` attribute, the error table and the caching headers.
- `services/recipe-service/docs/storage.md`: both new tables, the object key layout, bucket policy,
  the outbox semantics, memory bounds and the benchmark result, and the
  backup note (PRD §8).
- `services/recipe-service/docs/domain.md`: the `recipeimage` domain, its lifecycle relative to
  soft delete/restore/purge, and the processing rules.
- `services/recipe-service/README.md`: the `S3_*` configuration table.

---

## 16. Rough implementation order (input to `/plan-task`)

1. `objectstore` package plus the fake; config; `main.go` wiring with the
   disabled store and the startup ping.
2. `imaging` (orientation reader, processing, fixtures, fuzz, benchmark).
3. `recipeimage` domain: entity/model/builder/provider/administrator with
   both tables and the migration.
4. `recipeimage` processor (upload/replace/delete tx, outbox) and the
   `ObjectReaper`.
5. `recipeimage` handlers, metrics and routes; `recipe.ImageVisibility`.
6. `recipe` enrichment (`image` attribute, batched).
7. Retention cascade plus tests.
8. Deploy: compose MinIO and init, nginx, k8s, `.env.example`.
9. Frontend: client `upload`, service, hooks, types.
10. Frontend: `RecipeImage`, card, detail hero and lightbox.
11. Frontend: `recipe-image-field`, draft hook, form save sequencing.
12. Docs; benchmark recorded; manual compose acceptance; flagless
    `verify.sh`.

Steps 1–2 and 9 have no dependencies on each other, and steps 1–7 (backend)
and 9–11 (frontend) can proceed in parallel once the REST shape in §6/§8 is
fixed.
