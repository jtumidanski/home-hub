# Recipe Images — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-10-14
---

## 1. Overview

Recipes in Home Hub are text-only today. The list page is a grid of cards
showing a title, a description, time and tags. The detail page is a header
followed by ingredients and instructions. With dozens of recipes, the list is
hard to scan, and nothing visual shows what a dish looks like.

This feature lets each recipe have **one optional image**. A user uploads it
from the recipe form (create or edit), and can replace or remove it. It
appears as a hero image on the recipe detail page and as a thumbnail on the
recipe list cards. recipe-service resizes uploads into a "full" rendition and
a "thumb" rendition, strips EXIF metadata (including GPS), and stores the
bytes in an S3-compatible object store (MinIO). Image metadata lives in
Postgres next to the recipe.

This is the first feature in Home Hub that stores user-uploaded binary data.
The object-storage integration is therefore scoped to recipe-service and is
deliberately not a shared library. A visual proposal for the UI is in
[`mockup.html`](mockup.html). Open it directly in a browser; it has no
dependencies.

## 2. Goals

Primary goals:
- A user can add, replace, view and remove a single image per recipe.
- The recipe detail page shows the image as a responsive hero image.
- The recipe list cards show a thumbnail. Cards without an image show a
  neutral placeholder so the grid stays aligned.
- Uploads are validated, resized and stripped of metadata on the server, so
  list pages never load full-size phone photos.
- Image bytes live in MinIO and are served only through recipe-service. They
  are protected by the existing cookie auth and the tenant/household scoping.

Non-goals:
- Multiple images, galleries, or per-step photos.
- Adding an image by URL, from the recipe's `sourceUrl` (og:image), or by
  clipboard paste.
- In-app cropping, rotation or filters. The only change is automatic EXIF
  orientation correction.
- Images in cook mode, the meal planner, plan items, shopping lists or
  dashboard widgets.
- Presigned/direct-to-MinIO URLs, a CDN, or public image links.
- A shared `shared/go` object-storage library. Extract one later if a second
  service needs it.
- HEIC/HEIF and AVIF input.

## 3. User Stories

- As a household member, I want to attach a photo when I create a recipe so
  the dish is recognisable at a glance.
- As a household member, I want to replace a recipe's photo with a better one
  without recreating the recipe.
- As a household member, I want to remove a recipe's photo.
- As a household member browsing recipes, I want thumbnails on the recipe
  cards so I can find a dish visually.
- As a household member viewing a recipe, I want to see the photo at a useful
  size above the ingredients and instructions.
- As a mobile user, I want to pick a photo from my camera roll, or take one,
  from the recipe form.
- As a privacy-conscious user, I want location data removed from photos I
  upload.

## 4. Functional Requirements

### 4.1 Upload (create and edit)

- FR-1: The recipe form (`RecipeFormPage`, used for both create and edit) has
  a **Photo** section between the planner-config block and the Cooklang
  editor/preview columns.
- FR-2: With no image, the Photo section shows a drop zone that accepts:
  - drag-and-drop of one file (desktop), and
  - click/tap to open the native file picker with
    `accept="image/jpeg,image/png,image/webp"`. On mobile this offers the
    camera roll and the camera.
- FR-3: The client validates before staging. The type must be JPEG, PNG or
  WebP, and the size must be ≤ 10 MiB. If either check fails, show an inline
  error in the drop zone ("Use a JPEG, PNG or WebP image under 10 MB.") and
  stage nothing.
- FR-4: A valid file is **staged**, not uploaded immediately. The section
  shows a local preview (object URL) and **Replace** and **Remove** actions.
  The Cooklang live-preview column also shows the staged image above the
  derived title.
- FR-5: On **Save**, the frontend:
  1. saves the recipe (POST for create, PATCH for edit), as today, then
  2. if a new file is staged, `PUT /recipes/{id}/image`, or
  3. if the user removed an existing image, `DELETE /recipes/{id}/image`.
  It navigates to the detail page only after every step succeeds.
- FR-6: If the recipe saves but the image step fails, the user stays on the
  form, now in edit mode for the saved recipe, with the staged image kept.
  An error toast says "Recipe saved, but the photo failed to upload — try
  again." Pressing Save again retries only the image step.
- FR-7: While the image request is in flight, the Save button shows a
  progress state ("Uploading photo…") and is disabled.
- FR-8: **Cancel** discards staged changes. No image request is made.
- FR-9: In edit mode, an existing image is shown with Replace and Remove.
  Remove is staged; the image stays on the server until Save. An **Undo**
  link restores it before saving.

### 4.2 Detail page

- FR-10: When the recipe has an image, `RecipeDetailPage` renders it at the
  top of the page, above the title row:
  - Desktop (≥ md): full content width (max-w-5xl), fixed aspect 16:9,
    `object-fit: cover`, rounded corners.
  - Mobile: edge-to-edge, aspect 4:3, `object-fit: cover`.
  - Source is the `full` rendition. `loading="eager"`. `alt` is the recipe
    title.
- FR-11: Clicking or tapping the hero opens a lightbox (dialog) with the
  uncropped `full` rendition (`object-fit: contain`). Esc, a click on the
  backdrop, or the close button dismisses it.
- FR-12: A recipe with no image looks exactly like today's page. There is no
  placeholder on the detail page.
- FR-13: The Edit button is unchanged. Image management happens only in the
  form.

### 4.3 List page

- FR-14: Desktop grid: each `RecipeCard` has an image area at the top of the
  card, aspect 16:9, using the `thumb` rendition, `loading="lazy"`,
  `object-fit: cover`.
- FR-15: Mobile list (single column): the thumbnail is a 64×64 rounded square
  to the left of the title and description, so rows stay compact.
- FR-16: Cards without an image show a placeholder in the same footprint: a
  muted background with a centred `ChefHat` (lucide) icon. The grid row
  heights then stay consistent.
- FR-17: If an image fails to load, the placeholder is shown (`onError`
  fallback).

### 4.4 Lifecycle

- FR-18: Soft-deleting a recipe leaves its image in place. Restoring the
  recipe (`POST /recipes/restorations`) brings the image back unchanged.
- FR-19: When retention purges a recipe (`DeletedRecipesRestoreWindow`), its
  `recipe_images` row is deleted in the same cascade, and its objects are
  removed from MinIO.
- FR-20: Replacing an image writes the new objects, switches the DB row to
  them, then deletes the old objects (best effort, logged on failure).
- FR-21: Deleting an image removes the DB row, then the objects (best effort,
  logged on failure).

### 4.5 Processing (server)

- FR-22: Accepted input types are JPEG, PNG and WebP. The server finds the
  type by decoding the bytes. It never trusts the `Content-Type` header or
  the filename.
- FR-23: Maximum request size is 10 MiB. Maximum decoded dimensions are
  40 megapixels; this is checked with `image.DecodeConfig` **before** the
  full decode, to guard against decompression bombs. Anything larger gets a
  422.
- FR-24: The EXIF orientation of JPEG input is applied to the pixels before
  resizing.
- FR-25: Two renditions are produced, both re-encoded as JPEG (quality 82).
  Re-encoding drops all EXIF/XMP/ICC metadata:
  - `full`: longest edge ≤ 1600 px, aspect preserved, never upscaled.
  - `thumb`: longest edge ≤ 480 px, aspect preserved, never upscaled.
  Transparent PNG/WebP pixels are composited onto white.
- FR-26: The stored width and height are those of the `full` rendition.

## 5. API Surface

All paths are under `/api/v1`, routed to recipe-service by the existing
`location /api/v1/recipes` blocks. Every endpoint resolves the recipe through
the existing tenant- and household-scoped lookup. Another household's recipe
is a 404, never a 403.

### 5.1 `PUT /recipes/{id}/image` — upload or replace

- Request: `multipart/form-data` with a single part named `file`. This is an
  intentional deviation from JSON:API request bodies, because JSON:API has no
  binary upload convention. The response is still JSON:API.
- Response `200 OK`, `Content-Type: application/vnd.api+json`:

```json
{
  "data": {
    "type": "recipe-images",
    "id": "4b1e…",
    "attributes": {
      "width": 1600,
      "height": 1067,
      "thumbUrl": "/api/v1/recipes/9a0c…/image?size=thumb&v=4b1e…",
      "fullUrl": "/api/v1/recipes/9a0c…/image?size=full&v=4b1e…",
      "updatedAt": "2026-10-14T18:03:11Z"
    }
  }
}
```

- Errors (JSON:API error objects):

| Status | Condition |
|---|---|
| 400 | Not multipart, or no `file` part |
| 404 | Recipe not found, soft-deleted, or not in the caller's household |
| 413 | Body > 10 MiB |
| 415 | Bytes are not decodable JPEG/PNG/WebP |
| 422 | Decoded image > 40 MP, or an edge < 16 px |
| 503 | Object storage unavailable |

### 5.2 `GET /recipes/{id}/image?size=thumb|full` — serve bytes

- `size` defaults to `full`. Any other value is a 400.
- `200 OK` with `Content-Type: image/jpeg` and
  `X-Content-Type-Options: nosniff`. The body streams from MinIO through
  recipe-service.
- Caching: `Cache-Control: private, max-age=31536000, immutable` and
  `ETag: "<imageId>-<size>"`. The `v=<imageId>` query parameter in the
  advertised URLs changes on every replace, which busts the cache. A
  matching `If-None-Match` returns `304`.
- `404` if the recipe has no image or is not visible to the caller. `503` if
  storage is unavailable.
- Auth is the existing `access_token` cookie, so a plain `<img src>` works.

### 5.3 `DELETE /recipes/{id}/image` — remove

- `204 No Content`. Deleting a recipe that already has no image is also 204
  (idempotent). `404` if the recipe is not visible.

### 5.4 Changes to existing recipe resources

`GET /recipes` (list items), `GET /recipes/{id}`, `POST /recipes`,
`PATCH /recipes/{id}` and `POST /recipes/restorations` responses gain one
nullable attribute:

```json
"image": {
  "thumbUrl": "/api/v1/recipes/9a0c…/image?size=thumb&v=4b1e…",
  "fullUrl": "/api/v1/recipes/9a0c…/image?size=full&v=4b1e…",
  "width": 1600,
  "height": 1067
}
```

The value is `null` when there is no image. The list endpoint loads image
rows for the page in one batched query, with no N+1. The recipe POST and
PATCH request bodies do **not** accept `image`; it is managed only through
the endpoints above.

### 5.5 Proxy limits

- `deploy/compose/nginx.conf`: the `/api/v1/recipes` locations currently
  inherit nginx's default `client_max_body_size` of 1 MiB. Raise it to `11m`
  on the recipes locations.
- `deploy/k8s/ingress.yaml`: already `proxy-body-size: "10m"`. Raise it to
  `11m` so a 10 MiB file plus multipart overhead fits.

## 6. Data Model

### 6.1 New table `recipe.recipe_images`

| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK. Doubles as the cache-busting version; a new id on every replace |
| tenant_id | UUID | NOT NULL, indexed with household_id |
| household_id | UUID | NOT NULL |
| recipe_id | UUID | NOT NULL, **UNIQUE** (one image per recipe) |
| full_key | VARCHAR(512) | NOT NULL, object key of the full rendition |
| thumb_key | VARCHAR(512) | NOT NULL, object key of the thumb rendition |
| content_type | VARCHAR(50) | NOT NULL, `image/jpeg` |
| width | INT | NOT NULL (full rendition) |
| height | INT | NOT NULL (full rendition) |
| full_bytes | INT | NOT NULL |
| thumb_bytes | INT | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

- A new domain package `internal/recipeimage` follows the standard layout
  (entity, model, builder, processor, provider, administrator, resource,
  rest). GORM AutoMigrate runs from `Migration()`, as in the other packages.
- Replacing an image updates the row in place with a new `id`. Because
  `recipe_id` is unique, the upsert is a single statement in a transaction.
- No column is added to `recipes`.

### 6.2 Object layout (MinIO)

- Bucket: configured, default `home-hub-recipes`. Private, with no anonymous
  policy.
- Keys: `{tenantId}/{householdId}/recipes/{recipeId}/{imageId}/full.jpg` and
  `…/thumb.jpg`. The image id in the path means a replace never overwrites
  objects that may still be in browser or proxy caches.

### 6.3 Migration notes

- Additive only. Existing recipes have `image: null`. No backfill.

## 7. Service Impact

### recipe-service
- New `internal/recipeimage` domain package: upload/serve/delete handlers,
  image processing, and the DB entity.
- New `internal/objectstore` package: a thin wrapper around
  `github.com/minio/minio-go/v7` exposing `Put`, `Get` (stream), `Delete`
  and `Ping`. It is the only code that imports minio-go.
- `config.Config` gains `ObjectStore` with `S3_ENDPOINT`, `S3_BUCKET`
  (default `home-hub-recipes`), `S3_ACCESS_KEY`, `S3_SECRET_KEY`,
  `S3_USE_SSL` (default `false`) and `S3_REGION` (default `us-east-1`).
- Startup: the service checks that the bucket exists. If storage is
  unreachable or the bucket is missing, it logs an error and **continues to
  start**. Image endpoints return 503, and every other recipe endpoint
  works.
- `recipe` REST resources (list and detail) gain the `image` attribute,
  filled in by a batched provider call into `recipeimage`.
- `retention.cascadeDeleteRecipes` also deletes `recipe_images` rows and
  collects their keys. The objects are deleted after the transaction
  commits. The reaper's dry-run mode counts images and deletes nothing.
- Image processing needs `golang.org/x/image` (WebP decode and high-quality
  resampling), plus a small EXIF-orientation reader, either a
  minimal in-repo parser or `github.com/rwcarlsen/goexif`; design phase
  decides.
- Docs: `docs/rest.md`, `docs/storage.md`, `docs/domain.md` and the README
  are updated.

### frontend
- `types/models/recipe.ts`: `image: RecipeImage | null` on the list and
  detail attributes.
- New service methods `uploadRecipeImage(id, file)` (multipart, through the
  existing API client's base URL and credentials) and
  `deleteRecipeImage(id)`, plus React Query mutations that invalidate the
  recipe detail and list keys.
- New component `components/features/recipes/recipe-image-field.tsx`: drop
  zone, staged preview, replace/remove/undo, client validation.
- New component `components/features/recipes/recipe-image.tsx`: thumbnail and
  hero rendering, placeholder, and onError fallback.
- `RecipeFormPage.tsx`: Photo section; staged-image state; save sequencing
  (FR-5 – FR-8); image above the derived title in the live preview.
- `RecipeDetailPage.tsx`: hero image and lightbox.
- `recipe-card.tsx`: top thumbnail (grid) or left thumbnail (mobile), and
  the placeholder.
- Tests: the field's validation and staging, the save sequencing (including
  the partial-failure path), and card/detail rendering with and without an
  image.

### deploy
- `deploy/compose/docker-compose.yml`: add a `minio` service (with a named
  volume) and a one-shot `minio-init` that creates the bucket. Add `S3_*`
  environment variables to recipe-service. `scripts/local-up.sh` must keep
  working without manual steps.
- `deploy/compose/nginx.conf`: `client_max_body_size 11m` on the recipes
  locations.
- `deploy/k8s/recipe-service.yaml`: `S3_*` environment variables from a
  secret. `deploy/k8s/secrets.example.yaml`: the new keys.
- `deploy/k8s/ingress.yaml`: body size `11m`.
- `.env.example`: the `S3_*` entries.

### Other services
- None. No other service reads recipe images.

## 8. Non-Functional Requirements

- **Security**
  - Image bytes are reachable only through recipe-service with a valid
    session. The bucket is private, and MinIO credentials never reach the
    browser.
  - Server-side type detection by decoding, a pixel-count limit before the
    full decode, and re-encoding (which strips all metadata, including GPS)
    are mandatory.
  - Served images send `X-Content-Type-Options: nosniff` and
    `Content-Type: image/jpeg`.
  - Object keys come only from server-generated UUIDs. User input never
    enters a key.
- **Multi-tenancy**: every DB query on `recipe_images` filters by
  `tenant_id` and `household_id` through the existing context helpers.
  Object keys are prefixed by tenant and household.
- **Performance**
  - A thumb is ≤ ~60 KB for typical photos. A recipe list page of 20 cards
    adds at most 20 lazy-loaded thumb requests, and the list JSON adds one
    batched query.
  - Upload processing (decode, resize ×2, encode ×2) for a 12 MP JPEG
    finishes in < 2 s on the current recipe-service resources. If it
    doesn't, design phase revisits the resource limits.
  - Repeat views are served from the browser cache (immutable, versioned
    URLs).
- **Resilience**: a MinIO outage degrades only image features. Recipe CRUD,
  the planner and other endpoints are unaffected. List and detail responses
  still include `image` metadata from Postgres; the browser falls back to
  the placeholder when the bytes 503.
- **Observability**
  - Structured logs for upload, replace, delete and storage failures, with
    recipe id, image id, byte sizes and durations.
  - Prometheus counters/histograms: `recipe_image_uploads_total{result}`,
    `recipe_image_processing_seconds` and
    `recipe_image_storage_errors_total{op}`.
- **Backups**: MinIO data is outside the Postgres backups. Backups of the
  production bucket are the operator's responsibility (see Open Questions).

## 9. Open Questions

1. **Production MinIO**: is there an existing MinIO/S3 endpoint next to
   `postgres.home`, or should this task add a MinIO Deployment and PVC to
   `deploy/k8s/`? The PRD assumes an existing endpoint, configured by
   environment variables. The compose stack gets its own MinIO either way.
2. **Orphaned objects**: if the process crashes between `Put` and the DB
   commit, objects can be orphaned. Accept this for v1 (it is rare and
   small), or add a periodic sweep comparing keys to `recipe_images`? The
   PRD assumes accept-and-log.
3. **EXIF parsing dependency**: a minimal in-repo orientation parser or
   `rwcarlsen/goexif`? Left to the design phase.
4. **Output format**: JPEG only, because the Go stdlib has no WebP encoder
   and a cgo dependency would complicate the Dockerfile. Revisit if thumb
   sizes are a problem.

## 10. Acceptance Criteria

- [ ] On the create form, a user can drop or pick a JPEG/PNG/WebP ≤ 10 MiB,
      see a preview, save, and land on a detail page with a hero image.
- [ ] On the edit form, a user can replace the image. After saving, the
      detail page and the list show the new image without a hard refresh.
- [ ] On the edit form, a user can remove the image, and can undo before
      saving. After saving, the detail page has no hero and the card shows
      the placeholder.
- [ ] A file > 10 MiB, or one that is not an image, is rejected client-side
      with an inline error. The server enforces the same limits (413/415).
- [ ] An image > 40 MP returns 422 without being fully decoded.
- [ ] Stored renditions are JPEG, ≤ 1600 px and ≤ 480 px on the longest
      edge, have no EXIF data, and are oriented correctly for a rotated
      phone photo.
- [ ] Recipe list cards show thumbnails (grid and mobile layouts), with the
      `ChefHat` placeholder for recipes without an image.
- [ ] Clicking the detail hero opens a lightbox with the uncropped image.
- [ ] `GET /recipes/{id}/image` returns 404 for another household's recipe,
      and 304 on a matching `If-None-Match`.
- [ ] `image` is `null` or populated correctly on list, detail, create,
      update and restore responses, and the list uses one batched image
      query.
- [ ] A soft-deleted and then restored recipe keeps its image. A recipe
      purged by retention has no `recipe_images` row, and its objects are
      gone from MinIO.
- [ ] With MinIO stopped, recipe-service starts, recipe CRUD works, and the
      image endpoints return 503.
- [ ] If the recipe saves but the image upload fails, the user sees the
      partial-failure toast, and Save retries only the image.
- [ ] `scripts/local-up.sh` brings up MinIO and the bucket with no manual
      steps, and a 10 MiB upload succeeds through the compose nginx.
- [ ] recipe-service docs (`rest.md`, `storage.md`, `domain.md`) are
      updated.
- [ ] Flagless `tools/verify.sh` exits 0.
