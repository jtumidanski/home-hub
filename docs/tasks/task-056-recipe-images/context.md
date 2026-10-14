# Task 056 — Recipe Images: Implementation Context

Companion to [`plan.md`](plan.md). Read this first, then the task you are
implementing. Design: [`design.md`](design.md). PRD: [`prd.md`](prd.md).
Mockup: [`mockup.html`](mockup.html).

## What is being built

Each recipe can have one optional photo. recipe-service receives it as
multipart, decodes and validates it, and re-encodes it as two JPEG
renditions (full ≤ 1600 px, thumb ≤ 480 px). It stores them in MinIO and
records the metadata in `recipe.recipe_images`. Recipe list and detail
responses gain a nullable `image` attribute. The frontend adds a Photo
section to the recipe form, a hero with a lightbox to the detail page, and
thumbnails to the list cards.

## Task map and dependencies

| # | Task | Depends on | Area |
|---|---|---|---|
| 1 | objectstore package, fake, config | — | backend |
| 2 | EXIF orientation reader | — | backend |
| 3 | `imaging.Process` | 2 | backend |
| 4 | recipeimage domain (entities, outbox, queries) | — | backend |
| 5 | processor, metrics, outbox drainer | 1, 3, 4 | backend |
| 6 | HTTP handlers, `recipe.ImageVisibility`, `main.go` wiring | 5 | backend |
| 7 | `image` enrichment on recipe list/detail | 4, 5 | backend |
| 8 | retention cascade | 4 | backend |
| 9 | compose MinIO, nginx, k8s, `.env.example` | 1 (env contract) | deploy |
| 10 | client PUT upload, types, service, hooks, validation | REST shape only | frontend |
| 11 | `RecipeImage`, card, detail hero + lightbox | 10 | frontend |
| 12 | draft hook, Photo field, form save sequencing | 10 | frontend |
| 13 | docs, benchmark record, manual acceptance, gate | all | docs/verify |

Tasks 1, 2 and 4 are independent. Tasks 9 and 10 can run in parallel with
the backend once the REST shape (design §6/§8) is fixed. Tasks 11 and 12
both touch `frontend/src/components/features/recipes/__tests__/`, so run
them sequentially.

## Key files (existing code the plan touches)

| File | Why it matters |
|---|---|
| `services/recipe-service/cmd/main.go` | Migrations list; route wiring; `ctx` for background loops. Task 6 edits it once. |
| `services/recipe-service/internal/config/config.go` | `envOrDefault` / `parseDuration` helpers; Task 1 adds `parseBool` and `ObjectStore`. |
| `services/recipe-service/internal/planner/*` | Reference layout for a small child domain (entity/model/builder/provider/administrator/processor, batched `GetByRecipeIDs`). |
| `services/recipe-service/internal/recipe/processor.go` (`BuildDetailEnrichment`, `BuildListEnrichments`) | Where the `image` enrichment plugs in, following the planner/normalization degradation pattern (log and leave nil). |
| `services/recipe-service/internal/recipe/rest.go` | `RestModel`, `RestDetailModel`, `ListEnrichment`, `DetailEnrichment`, `Transform*`. |
| `services/recipe-service/internal/recipe/provider.go` | `getByID` filters `id` and `deleted_at IS NULL` only; tenant comes from the callback. |
| `services/recipe-service/internal/recipe/processor_integration_test.go` | `setupTestDB` is shared by all recipe package tests; Task 7 adds the recipeimage tables to it. |
| `services/recipe-service/internal/retention/handlers.go` | `cascadeDeleteRecipes`; Task 8 inserts the image purge before the final recipe delete. |
| `shared/go/database/tenant_callbacks.go` | Auto `WHERE tenant_id = ?` on Query/Update/Delete when the ctx carries a tenant. Not applied to Create. |
| `shared/go/server/handler.go`, `response.go` | `RegisterHandler`, `ParseID`, `WriteError`, `MarshalResponse`. Do not change (no `shared/` edits). |
| `shared/go/retention/loop.go` | `sr.Loop(ctx, l, base, fn func(context.Context))` is reused for the outbox drainer tick. |
| `frontend/src/lib/api/client.ts` | **Already has `upload()` (POST, FormData, no Content-Type, XHR progress path).** Task 10 adds a `method` option instead of a new method. |
| `frontend/src/services/api/base.ts` | `setTenant`, `remove` (DELETE). |
| `frontend/src/lib/hooks/api/use-recipes.ts` | `recipeKeys` factory; mutation invalidation pattern. |
| `frontend/src/components/ui/card.tsx` | `size="sm"` uses `px-3`/`py-3`; the card has `overflow-hidden`. |
| `frontend/src/components/ui/dialog.tsx` | base-ui Dialog: `Dialog` (`open`/`onOpenChange`), `DialogContent` (`showCloseButton` default true), `DialogTitle`. |

## Decisions made in planning (beyond the design)

1. **`api.upload` already exists.** Design §1 says the client has "no
   multipart path"; that is wrong. `upload()` posts FormData without a
   Content-Type and has an XHR progress variant. The plan adds
   `UploadOptions = RequestOptions & { method?: "POST" | "PUT" }` and
   threads it into both the fetch and XHR paths. Existing POST callers are
   unchanged.
2. **`Processor.WithStore(store)`** instead of a 4-arg `NewProcessor`. The
   recipe package's read-only enrichment calls
   `recipeimage.NewProcessor(l, ctx, db)` exactly like planner and
   normalization; only the handlers bind a store.
3. **The processor stays synchronous; handlers launch the drain goroutine**
   (`go reaper.DrainKeys(context.Background(), ids)`). This keeps processor
   tests deterministic.
4. **Test seam `beforeInsertHook`** (an unexported package var) makes the
   unique-violation retry path testable on SQLite.
5. **Test fixtures are generated in test code** (JPEG via `image/jpeg`,
   EXIF APP1 bytes built by hand, a header-only PNG for the 40 MP bomb).
   The design said "committed fixtures". The only binary fixture is a
   44-byte lossless WebP (Go has no WebP encoder), embedded as base64 in
   `imaging_test.go`. It was produced with `cwebp -lossless -exact` and
   verified to decode with `golang.org/x/image/webp` to 32×24 with a
   transparent left half.
6. **recipeimage tests use SQLite `:memory:` with `SetMaxOpenConns(1)`.**
   Each new `:memory:` connection is a fresh empty DB, and the handler
   drains in a goroutine.
7. **The upload handler also extends the write deadline** to 60 s (the
   design only mentions read). This is still per-handler; shared
   `server.go` is untouched.
8. **k8s `secretKeyRef` uses `optional: true`** for the `S3_*` keys, so the
   pod starts with storage disabled (503 on image endpoints) instead of
   crash-looping if `object-store-credentials` doesn't exist yet. This
   matches design §9.
9. **The hero renders nothing on image load error** (no placeholder on the
   detail page, consistent with FR-12). Cards fall back to the
   placeholder (FR-17).
10. **Form save state:** `persistedId` (set after a create whose image step
    failed) plus `lastSavedAttrs` (serialized attrs of the last successful
    recipe save). A retry skips the recipe step only when the attrs are
    unchanged. That covers planner-config edits, which are not part of
    react-hook-form's dirty tracking.
11. **`validateRecipeImageFile` lives in `frontend/src/lib/recipe-image.ts`**,
    not in the component file, to avoid the react-refresh
    only-export-components lint.
12. **An empty `file.type`** (some Android pickers) is rejected client-side.

## Open risks for the executor and reviewers

- **Household vs tenant visibility (Review Focus #1).** Existing recipe
  reads are tenant-scoped only, so a recipe can be listed in household B
  while its image row belongs to household A. By design (§6.1, PRD NFR),
  household B sees `image: null` and a 404 on the image endpoints.
  `TestImageEnrichmentOtherHousehold` pins this. If product intent is "a
  recipe's photo follows the recipe", that is a design change, not an
  implementation fix. Raise it; don't silently drop the household filter.
- **MinIO image availability.** MinIO's community distribution changed in
  late 2025, and Docker Hub tags may be stale or unpublished. Task 9 Step 1
  looks up and pull-tests the tags; if no usable tag exists, report BLOCKED
  rather than using `latest` or an unpinned image.
- **`clause.Returning{}` on SQLite** needs SQLite ≥ 3.35. The current
  `mattn/go-sqlite3 v1.14.22` bundles a new enough version. Task 4's tests
  exercise it directly.
- **`http.MaxBytesError` propagation through `multipart.Reader`.** The 413
  "body over cap" test depends on the error surfacing through
  `NextPart()`. If it is wrapped without `%w`, extend `isMaxBytes` to also
  check `strings.Contains(err.Error(), "request body too large")`, and add
  a comment explaining why.
- **Benchmark budget (PRD NFR < 2 s for 12 MP).** Task 3 runs it; the
  fallback is described there. Task 13 records the number.

## Dependencies added

- `github.com/minio/minio-go/v7` (latest at implementation time; v7.4.x
  when planned): Task 1.
- `golang.org/x/image` (v0.47.x when planned): Task 3.
- `github.com/prometheus/client_golang` moves from indirect to direct
  (`promauto`): Task 5.
- No new frontend packages.

## Verification

- Implementers run module-local checks only (`go build ./... && go test
  ./...` in `services/recipe-service`; `npx vitest run …` and
  `npx tsc -b` in `frontend`).
- The task-verifier runs `tools/verify.sh --quick` per task. Flagless
  `tools/verify.sh` is the final gate. No `shared/` changes, so the docker
  leg rebuilds only the recipe-service image.
- Manual compose acceptance (Task 13 Step 6) covers the PRD items that
  need real infrastructure, and is recorded in
  `docs/tasks/task-056-recipe-images/verification.md`.
- Code review (`plan-adherence-reviewer`, `backend-guidelines-reviewer`,
  `frontend-guidelines-reviewer`) is required before the PR.
