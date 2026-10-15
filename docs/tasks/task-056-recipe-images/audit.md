## Plan adherence

**Verdict: MOSTLY_COMPLETE / READY_TO_MERGE after the user runs the deferred manual acceptance (items 2-9). No code gaps found.**

Base `main` c55e085, HEAD 0d09f97, 79 files changed. Plan has 13 tasks and 96 step checkboxes. All 96 are still `- [ ]` in plan.md; no step was ticked, so progress cannot be read from the plan. I verified against the diff instead.

| # | Task | Status | Evidence |
|---|------|--------|----------|
| 1 | objectstore pkg, fake, config, wiring | DONE | `internal/objectstore/{objectstore,minio,disabled}.go`, `objectstoretest/store.go`, tests; `config.go` +22; minio-go imported only by `objectstore.go` and `minio.go`; `main.go:55` |
| 2 | EXIF orientation reader | DONE | `imaging/orientation.go` + test |
| 3 | `imaging.Process` | DONE | `imaging.go:16-20` (40M px, 1600, 480, q82); `TestProcessOrientedDownscale` :71, `TestProcessPalettedPNGWithTransparency` :125 |
| 4 | recipeimage domain | DONE | entity, model, builder, keys, errors, provider, administrator, domain_test, testdb_test all present |
| 5 | processor, metrics, drainer | DONE | `processor.go`, `metrics.go` (3 collectors), `reaper.go`; `main.go:64-65` runs the reaper on a 5 min loop |
| 6 | HTTP handlers, visibility, routes | DONE | `resource.go:26-27` (10<<20, +512<<10), nosniff and immutable cache headers, weak-list ETag at :239; `TestGetImageNotModifiedWeakList` :228; `recipe/visibility.go`; `main.go:82` |
| 7 | `image` on list and detail | DONE | `recipe/rest.go`, `processor.go`, `image_enrichment_test.go`. `batch_test.go` is listed in the plan but not touched; the build and tests pass, so this is harmless. |
| 8 | Retention purges into outbox | DONE | `retention/handlers.go:97` `PurgeByRecipeIDsTx`; `handlers_test.go` +90 |
| 9 | Deploy | DONE (deviation, ruled) | compose `rustfs` + `rustfs-init`, nginx `client_max_body_size 11m` (:184), no `proxy_request_buffering off`, ingress 10m to 11m, k8s manifest, secrets example, `.env.example` S3_* vars. Consistently RustFS; k8s secrets comment cites the RustFS runbook. The .gitignore change `recipes/` to `/recipes/` is needed so the frontend `recipes/` dir is tracked; it is not in the plan. |
| 10 | Frontend transport | DONE | `client.ts`, `recipe.ts`, `use-recipes.ts`, `lib/recipe-image.ts:4` verbatim error string, "rejects empty mime type" test :27 |
| 11 | Frontend rendering | DONE | `recipe-image.tsx`, `recipe-card.tsx`, `RecipeDetailPage.tsx` + tests |
| 12 | Frontend form | DONE | `use-recipe-image-draft.ts`, `recipe-image-field.tsx`, `RecipeFormPage.tsx:145` toast and :305 "Uploading photo…" verbatim |
| 13 | Docs, benchmark, acceptance, gate | PARTIAL (ruled) | `rest.md`, `storage.md`, `domain.md`, README, `verification.md` present. Benchmark recorded: 0.54 s/op against the 2 s budget. Acceptance item 1 PASS; **items 2-9 NOT RUN and open**, with UI visual check pending the user. Gate: flagless `tools/verify.sh` passed at 0d09f97 (reported by the controller, not re-run here). |

Completion: 12 DONE and 1 PARTIAL (the deferred manual items). Skipped without approval: 0.

**Global constraints spot-checked:**
- No changes under `shared/`.
- Metric label values `created`, `replaced`, `rejected_size`, `rejected_type`, `rejected_dimensions`, `storage_error` and `db_error` all appear in `resource.go`. The `put|get|delete|ping` ops were not verified.
- Frontend strings match the plan verbatim.
- No new frontend dependencies were checked against package.json (`go.mod` and `go.sum` changed as planned).

**Build and test (module-local):**

| Module | Build | Tests |
|---|---|---|
| services/recipe-service | PASS (`go build`, `go vet`) | PASS (all packages, `-count=1`) |
| frontend | PASS (`tsc -b`) | PASS (vitest, 112 files, 739 tests) |

**Action items:**
1. The user runs manual acceptance items 2-9 and the UI visual check (login via `scripts/local-up.sh`, then curl with the `access_token` cookie), and records the results in `verification.md`.
2. Optional: tick the plan.md checkboxes, or accept that they are unticked.
3. Optional: note the `.gitignore` change in the PR description.

## Backend guidelines

- **Scope:** services/recipe-service, changed packages (c55e085..0d09f97). **Date:** 2026-10-14
- **Build:** PASS (`go build ./...`). **Tests:** all 15 packages ok, 0 failed (`go test ./... -count=1`).
- **Overall: NEEDS-WORK** (2 FAIL, 4 WARN, no security-class findings; no shared/ changes)

Paths below are relative to `services/recipe-service/internal/`.

### Classification
`recipeimage` = domain (has model.go; no recipe-level CRUD, singleton resource per recipe). `recipeimage/imaging` and `objectstore` (+`objectstoretest`) = support. `recipe` changes = enrichment + `visibility.go` only. `retention` = cascade hook. Auth/SEC checks: not applicable (not an auth service); no `os.Getenv` outside config, no secrets in source.

### recipeimage (domain)

| ID | Status | Evidence |
|----|--------|----------|
| DOM-01 | PASS | recipeimage/builder.go:27 NewBuilder, setters :29-41, Build() validates :43 |
| DOM-02 | PASS | entity.go `func (m Model) ToEntity()` |
| DOM-03 | PASS | entity.go `func Make(e Entity) (Model, error)` |
| DOM-04 | PASS | rest.go:27 Transform |
| DOM-05 | FAIL | No `TransformSlice` in rest.go. No list endpoint exists (singleton), so the symbol is unused; recommend an explicit waiver rather than dead code. recipe.TransformSlice (recipe/rest.go:111) is the only list path and is used. |
| DOM-06 | PASS | processor.go:38 `l logrus.FieldLogger` |
| DOM-07 | PASS | resource.go:84,184,253 all `d.Logger()`; no StandardLogger |
| DOM-08 | WARN | Only write route is PUT (resource.go:39) on plain `RegisterHandler` (:38) with manual multipart parsing (readFilePart). Documented intentional deviation (services/recipe-service/docs/rest.md); check literally targets POST/PATCH. |
| DOM-09 | PASS (vacuous) | Transform returns no error (rest.go:27), so nothing to discard. Guideline signature is `(RestModel, error)`; WARN-level deviation, cannot fail. |
| DOM-10 | PASS | testdb_test.go:30 `database.RegisterTenantCallbacks(l, db)` |
| DOM-11 | WARN | provider.go:13 uses `database.Query`; provider.go:19,30,38 are hand-rolled `func(db)([]T, error)` closures instead of `database.SliceQuery`. Deferred (not FixedProvider-eager), so not a hard fail. |
| DOM-12 | PASS | zero `os.Getenv` in resource.go |
| DOM-13 | WARN | Handlers call the injected `RecipeVisibility` (resource.go:47-60, called at :65,:181,:250) and spawn `go reaper.DrainKeys` (:91,:261). Injection avoids a package cycle and no recipe internals are touched, but the visibility check and post-commit drain are orchestration in the handler layer, not the processor. |
| DOM-14 | PASS | no provider calls in resource.go (grep) |
| DOM-15 | PASS | no db.Create/Save/Delete in resource.go; writes in administrator.go:27,34,52,70,76 via processor.go swapTx/Delete |
| DOM-16 | PASS | administrator.go exists |
| DOM-17 | PASS | resource.go:47-58 (404 / 500), :136-157 (415, 422, 503, 500), :69-79 (400, 413), :189 (400), :196 (404) |
| DOM-18 | PASS | rest.go GetName/GetID/SetID |
| DOM-19 | PASS (N/A) | no request model; multipart body, documented |
| DOM-20 | FAIL | No `t.Run` and no `[]struct` table anywhere in recipeimage/*_test.go, objectstore tests, recipe/visibility_test.go, recipe/image_enrichment_test.go or retention tests; they are one-func-per-case (e.g. resource_test.go:83,118,164,189,212,228,242; processor_test.go:62-255). Only imaging/imaging_test.go:43,179 and orientation_test.go:19 use tables. Several are natural table candidates (TestPutImageErrors, TestGetImageErrors, TestParseSize domain_test.go:46). |

### Cross-cutting
- Tenant scoping: no manual `tenant_id` WHERE; household filtered explicitly (provider.go:15,22; administrator.go:23) as design requires. Reaper runs tenant-less by design (reaper.go struct comment; shared callback no-ops without tenant, shared/go/database tenant callback). PASS.
- Config read once: config/config.go ObjectStore block; passed via cmd/main.go:55 `objectstore.New`. PASS. `RegisterTenantCallbacks` not added to main.go. PASS.
- minio isolation: only objectstore/minio.go imports minio-go. PASS.
- Layering: recipe -> recipeimage dependency one-way (recipe/processor.go enrichment, visibility.go); recipeimage does not import recipe. Retention uses exported `PurgeByRecipeIDsTx` in same tx (retention/handlers.go cascade). PASS.
- SUB-04-style check: zero `json.NewDecoder/Unmarshal` in recipeimage; `io.ReadAll` at resource.go bounded by `LimitReader` for the multipart part (intentional, documented).
- Enrichment failure isolation: list degrades to no images with an error log (recipe/processor.go list enrichment), detail ignores ErrImageNotFound. PASS.

### Blocking (must fix)
- DOM-20: convert case-per-function tests to table-driven with `t.Run` (at minimum TestPutImageErrors, TestGetImageErrors, TestParseSize, TestOpenMapsErrors), or record an explicit waiver.
- DOM-05: add `TransformSlice` or record an explicit waiver (singleton resource, no list endpoint).

### Non-Blocking
- DOM-08: PUT multipart on RegisterHandler is a documented deviation; no action unless the reviewer rejects it.
- DOM-09: Transform has no error return (guideline signature).
- DOM-11: provider.go:19,30,38 hand-rolled closures vs `database.SliceQuery`.
- DOM-13: visibility check and `go reaper.DrainKeys` live in handlers.

### Re-audit after DOM-20 fix (2026-10-14)

- **Scope:** `services/recipe-service`, fix range c4d5504..3cb1afd (test-only, commit 3cb1afd) against the current tree. Test files from `git diff --name-only c55e085..HEAD`.
- **Build:** PASS. **Tests:** `go build ./...` and `go test ./... -count=1` ok, 15 packages, 0 failed.
- **Overall: PASS** (0 FAIL, 5 WARN, 1 WAIVED). Production code is untouched by the fix, so DOM-08/09/11/13 are carried forward unchanged.

| ID | Status | Evidence |
|----|--------|----------|
| DOM-05 | WAIVED | `recipeimage/rest.go` has no `TransformSlice`. By ruling, recipe-images is a singleton per recipe with no list endpoint, so it would be dead code. |
| DOM-20 | WARN (was FAIL) | Tables with `t.Run` now exist where cases share a shape: `recipeimage/resource_test.go:129,153` (`TestPutImageErrors`), `:246,276` (`TestGetImageErrors`), `recipeimage/processor_test.go:233,242` (`TestOpenMapsErrors`), `recipe/visibility_test.go:28,51`, `objectstore/objectstore_test.go:16,26`, `recipe/processor_integration_test.go:38,111`, and `imaging/*`. Review-focus pins remain top-level and named: `image_enrichment_test.go:82`, `resource_test.go:231`. |
| DOM-08 | WARN | Unchanged. PUT on `RegisterHandler` (documented multipart deviation). |
| DOM-09 | WARN | Unchanged. `Transform` has no error return. |
| DOM-11 | WARN | Unchanged. `provider.go:19,30,38` hand-rolled closures. |
| DOM-13 | WARN | Unchanged. Visibility check and `go reaper.DrainKeys` in handlers. |

DOM-20 residual (non-blocking). Two tests in `recipeimage/domain_test.go` still repeat one setup and assertion shape as loops or sequential `if`s:
- `TestParseSize` (`:46-58`), which was a named candidate in the fix brief. It uses a map loop plus a reject loop instead of one table.
- `TestBuilderValidation` (`:60-67`), two builder-input to expected-error cases.

The remaining untabled tests (processor upload, replace, failure and retry paths, reaper, retention, objectstoretest, enrichment) have distinct setup and assertions, which the brief allows. The two above are the only same-shape cases left, so DOM-20 is WARN, not PASS. Converting them is a trivial test-only change.

**Blocking (must fix):** None.

## Frontend guidelines

- **Scope:** 20 changed files under `frontend/` (c55e085..HEAD), 12 source, 8 test.
- **Date:** 2026-10-14
- **Build:** PASS (`npm run build`, tsc + vite; chunk-size warning is pre-existing)
- **Tests:** 739 passed, 0 failed (112 files). `npm test -- --watchAll=false` is rejected by vitest (`Unknown option --watchAll`), so I ran `npx vitest run`.
- **Overall:** PASS, with 3 non-blocking warnings.

### Anti-pattern checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 | PASS | No `: any` or `as any` in any changed source file (grep, zero matches). |
| FE-02 | PASS | `cn()` used at `recipe-image.tsx:47`, `recipe-image-field.tsx:96`. No `className={"`/template concatenation. |
| FE-03 | PASS | `@/lib/api/client` imported only by `services/api/recipe.ts:15`, not by components or pages. |
| FE-04 | PASS | No `z.` in components or pages. No new schema needed (client-side file validation is a plain helper, `lib/recipe-image.ts:6-13`). |
| FE-05 | PASS | The only `animate-spin` is `RecipeFormPage.tsx:304`, on the submit button. |
| FE-06 | PASS | Semantic classes only (`bg-muted`, `text-muted-foreground`, `text-destructive`, `border-primary`). Grep for raw palette classes found none. |
| FE-07 | PASS | The `.push(` hits in `recipe-card.tsx:23-28` are on a local array and predate this change. The draft hook uses immutable `setDraft` (`use-recipe-image-draft.ts:35,41,46`). |
| FE-08 | PASS | No `export default` in any changed file. |
| FE-09 | PASS | `useUploadRecipeImage` and `useDeleteRecipeImage` (`use-recipes.ts:137-160`) use `useTenant()` with `tenant!`, matching the existing `useDeleteRecipe` (`:109`). They are mutations, so `enabled` does not apply. No new query hooks. |
| FE-10 | PASS | No new query keys. Invalidation uses the existing `recipeKeys.lists/detail(tenant, household, ...)` (`use-recipes.ts:143-144,155-156`). |
| FE-11 | WARN | `RecipeFormPage.tsx:144` is a bare `catch {}` with no `createErrorFromUnknown`. The plan mandates the verbatim toast (`:145`), so no server message is lost on purpose. The recipe-save catch (`:132-133`) does use `createErrorFromUnknown`. The new mutations have no `onError` toast, unlike the sibling hooks. This is acceptable only because the page owns the toast. |

### Architecture checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-12 | PASS | `RecipeImageResource` is `{ id, type, attributes }` (`types/models/recipe.ts:~171-175`). `RecipeImage` is an attribute object nested in `RecipeListAttributes.image: RecipeImage \| null`. |
| FE-13 | PASS | `RecipeService extends BaseService`. `deleteRecipeImage` uses `this.remove` (`recipe.ts:~46`). `uploadRecipeImage` calls `api.upload` directly (`:~41`), the same direct-client pattern as other upload callers, with `setTenant` first. |
| FE-14 | PASS | No key factory changes. `RECIPE_IMAGE_TYPES` is `as const` (`lib/recipe-image.ts:2`). |
| FE-15 | PASS | `RecipeFormPage` keeps `useForm` + `zodResolver`. Image state sits outside the form schema (the file is not form data), which is reasonable. |
| FE-16 | PASS | No new schema. `recipe.schema.ts:3,7` is unchanged and still paired with `z.infer`. |

### Styling

| ID | Status | Evidence |
|----|--------|----------|
| FE-19 | PASS | The drop zone (`role="button"`) has `cursor-pointer` (`recipe-image-field.tsx:97`). Recipe card keeps `cursor-pointer` (`recipe-card.tsx:~57`). The hero trigger and Undo are native `<button>`s (`RecipeDetailPage.tsx:~87`, `recipe-image-field.tsx:~110`), so they are exempt. |

### Testing

| ID | Status | Evidence |
|----|--------|----------|
| FE-17 | PASS | Tests exist for every new or changed unit: `recipe-image.test.tsx`, `recipe-image-field.test.tsx`, `recipe-card.test.tsx`, `use-recipe-image-draft.test.ts`, `lib/__tests__/recipe-image.test.ts`, `client.test.ts`, `recipe.test.ts`, `RecipeFormPage.test.tsx`, `RecipeDetailPage.test.tsx`. All 739 tests pass. |
| FE-18 | PASS (n/a) | The repo has no `__mocks__` directory for services. Tests mock inline. |

### Task constraints

- No new dependencies: `package.json` is not in the frontend diff. No `react-dropzone` in the diff.
- Verbatim strings: `lib/recipe-image.ts:4` (validation message), `RecipeFormPage.tsx:145` (partial-failure toast), `RecipeFormPage.tsx:305` (`Uploading photo…`). All three match the plan.
- `PUT` upload support: `client.ts` `UploadOptions` with `method?: "POST" | "PUT"`, used by both the fetch path and the XHR path (`xhr.open(options.method ?? "POST", ...)`).
- Save sequencing (`RecipeFormPage.tsx:~119-152`): create or update, then image step, then success toast and navigate. A failed image step returns before navigating. `persistedId` plus `lastSavedAttrs` stop a retry from creating a duplicate recipe.

### Non-blocking (should fix)

1. FE-11: the bare `catch {}` at `RecipeFormPage.tsx:144` discards the error. Consider logging it, or passing it through `createErrorFromUnknown`, while keeping the verbatim toast text.
2. `recipe-image-field.tsx:~91-96`: the drop zone stays focusable and clickable while `disabled` is true, and has no `aria-disabled`. The picker click is a no-op because the input is disabled, and the drop handler is guarded (`:30`), so there is no functional bug. `onDragOver` still shows the dragging highlight while disabled.
3. `RecipeFormPage.tsx:~149`: `if (!savedId) return; // unreachable` is a dead branch kept to narrow the type. It is harmless.

### Blocking (must fix)

None.

---

# Frontend Audit — image tenant-header fix (2026-10-15)

- **Audit Scope:** `git diff --stat 10145f2..3b42157` (10 frontend files, branch `task-056-image-tenant-fix`)
- **Requirement:** `docs/tasks/task-056-recipe-images/bug-image-get-zero-tenant.md`
- **Build:** PASS (`npm run build`)
- **Tests:** 746 passed, 0 failed (`npx vitest run`, 113 files). `npm test -- --watchAll=false` errors because vitest has no `--watchAll` flag, so the run used `npx vitest run`.
- **Lint:** 0 errors. The 5 warnings are pre-existing and none are in the changed files.
- **Overall:** PASS (0 blocking, 4 non-blocking)

## File inventory

- Hook: `lib/hooks/use-recipe-image-src.ts` (new), `lib/hooks/use-recipe-image-draft.ts`
- Component: `components/features/recipes/recipe-image.tsx`
- Page: `pages/RecipeDetailPage.tsx`
- Tests: `use-recipe-image-src.test.tsx` (new), `use-recipe-image-draft.test.ts`, `recipe-image.test.tsx`, `recipe-card.test.tsx`, `RecipeDetailPage.test.tsx`, `RecipeFormPage.test.tsx`

## Checklist

| ID | Status | Evidence |
|----|--------|----------|
| FE-01 | PASS | No `any` in the changed sources. `url!` at `use-recipe-image-src.ts:21` is a non-null assertion guarded by `enabled: !!url` (`:22`). |
| FE-02 | PASS | `recipe-image.tsx` uses `cn("object-cover", className)`. |
| FE-03 | WARN | The component and page do not import the client. The hook imports `api` directly (`use-recipe-image-src.ts:3`) rather than going through `services/api`. The requirement doc prescribes `api.download`, and no service method exists for binary fetches. |
| FE-04 | PASS | No Zod in the changed files. |
| FE-05 | PASS | No `animate-spin` in the diff. Loading shows the placeholder (card) or nothing (hero), per the requirement. |
| FE-06 | PASS | No color classes added. |
| FE-07 | PASS | No mutation. |
| FE-08 | PASS | Named exports only. |
| FE-09 | PASS | The hook uses `useTenant()` and has `enabled: !!url && !!tenant?.id && !!household?.id` (`:22`). |
| FE-10 | PASS | The key is `[...recipeKeys.all(tenant, household), "image", url] as const` (`:20`). `recipeKeys.all` expands to `["recipes", tenant?.id ?? "no-tenant", household?.id ?? "no-household"]` (`use-recipes.ts:13-14`). The test asserts the concrete key (`use-recipe-image-src.test.tsx:44`). |
| FE-11 | PASS (n/a) | No `.catch`. A query failure surfaces as `isError`, which the component maps to hero `null` or card placeholder (`recipe-image.tsx:34`). |
| FE-14 | PASS | `as const` at `use-recipe-image-src.ts:20`. |
| FE-19 | PASS | The lightbox `<img>` is not interactive. The hero trigger is a native `<button>` (`RecipeDetailPage.tsx:~91`). |
| FE-17 | PASS | New hook test covers: fetch via `api.download` with the `/api/v1` prefix stripped, tenant-aware key, object URL returned, revoke on unmount, no fetch when `url` is null, error state. Component, page and draft tests were updated. |
| FE-18 | PASS (n/a) | No `__mocks__` directory. Tests mock inline. |

## Correctness review

- **Shared key, hero and lightbox: PASS.** Both call `useRecipeImageSrc(image.fullUrl)` and share one cached `Blob`, but each hook instance creates its own object URL in its own effect (`use-recipe-image-src.ts:32-37`) and revokes only that one (`:36`). A revoke by one consumer cannot invalidate the other's `src`. The hero is unmounted only with the page. The lightbox hook is always mounted in the page (`RecipeDetailPage.tsx:~31`).
- **Leaks: PASS.** Create and revoke are paired in one effect keyed on `blob`. Unmount and blob change both revoke. `resolved.blob === blob` (`:39`) means a revoked URL is never returned for a new blob. Under StrictMode the effect runs, cleans up and re-runs, so the first URL is revoked and no URL leaks.
- **Tenant and household in key: PASS** (FE-10). A tenant switch changes the key, so cached images do not leak across tenants. The `enabled` guard prevents the nil-tenant fetch.
- **The bug is fixed.** Image bytes now go through `api.download`, which sends `X-Tenant-ID` and `X-Household-ID` (`client.ts:426`). The draft hook resolves only the `unchanged` existing image and passes a staged preview through unchanged (`use-recipe-image-draft.ts:47-49`).

## Non-blocking

1. **Lazy loading is defeated.** `loading="lazy"` on card thumbnails (`recipe-image.tsx:~44`) no longer defers anything. The blob fetch fires on mount for every card, off-screen or not, so a large recipe list issues one request per card at once. Consider an IntersectionObserver gate on `enabled`, or accept this for list sizes in use.
2. **Refetch flicker.** `useInvalidateRecipes().invalidateAll` (`use-recipes.ts:179`) matches the image keys, which sit under `recipeKeys.all`. Active image queries refetch and return a new `Blob`, which gives a new object URL. `src` is `null` for one render in between (`:39`), so the hero briefly vanishes and the card shows the placeholder. Placing the image key outside `recipeKeys.all`, or setting `refetchOnMount`/`structuralSharing` options, would avoid this. Replacement images already get a new `?v=` key, so they do not need the invalidation.
3. **Hero button during load.** While loading, the hero `RecipeImage` renders `null` but the `View photo full size` button (`RecipeDetailPage.tsx:~91`) is still present and focusable. If the lightbox is opened before the blob resolves, the `<img>` has `src={undefined}` (`:204`). The window is small but real.
4. **FE-03 (WARN above).** If binary fetches become common, add a service method rather than calling `api` from a hook.

### Blocking (must fix)

None.

---

# Frontend Audit — Fix round 2 (2026-10-15)

- **Scope:** commit range `8cdb39a..05a801e` (9 files), against `## Fix round 2` items 1-4 in `bug-image-get-zero-tenant.md`.
- **Verdict:** PASS. All four items are resolved. No blocking findings.
- **Tests:** targeted only (the full gate runs separately). `recipe-image`, `recipe-card`, `use-recipe-image-src`, `RecipeDetailPage` and `RecipesPage` suites: 34 passed, 0 failed.

## Items

| # | Item | Status | Evidence |
|---|------|--------|----------|
| 1 | Lazy loading for cards | PASS | `recipe-image.tsx:40-41` calls `useInView("200px")` and passes `enabled: isHero \|\| inView`. The ref is attached to the card placeholder (`:46`). `use-in-view.ts:13-24` uses `rootMargin` and disconnects once visible (`:17`). `use-recipe-image-src.ts:15-16,21` adds `enabled` (default `true`). No in-view hook existed before (`lib/hooks` has only the new file). Test: `recipe-image.test.tsx:39-52` shows the card is disabled until `isIntersecting: true`, the hero is eager, and `disconnect` is called. The IntersectionObserver stub is at `:13-20`. |
| 2 | No flicker on invalidation | PASS | The key is `["recipe-images", tenant?.id, household?.id, url]` (`use-recipe-image-src.ts:19`), outside `recipeKeys.all`. Test: `use-recipe-image-src.test.tsx:60-66` invalidates `["recipes", ...]` and asserts one download call. The comment at `:11-13` documents the reason. |
| 3 | Lightbox before load | PASS | `RecipeDetailPage.tsx:89` (trigger) and `:199` (Dialog) are both gated on `attrs.image && lightboxSrc`. The `<img>` uses `src={lightboxSrc}` with no `?? undefined`. Test: `RecipeDetailPage.test.tsx:82` (no trigger until resolved) and `:64-80` (opens with the blob src). |
| 4 | FE-03 | PASS | `recipe.ts:131-139` adds `downloadImage(url)`, which strips `/api/v1` and calls `api.download`. The hook calls `recipeService.downloadImage` (`use-recipe-image-src.ts:20`) and no longer imports the api client. The hook test mocks `recipeService.downloadImage` (`use-recipe-image-src.test.tsx:8`). |

## Regression checks

- **Tenant and household in the key: PASS.** Both are in the key (`use-recipe-image-src.ts:19`) and both gate `enabled` (`:21`). A tenant switch changes the key, and the nil-tenant fetch cannot fire.
- **Observer cleanup: PASS.** The effect returns `observer.disconnect()` (`use-in-view.ts:23`). It also disconnects on the first intersection (`:17`). It is a no-op once `inView` is latched (`:12`). The ref is a state setter, so a node that unmounts when the placeholder is replaced by the `<img>` only re-runs a no-op effect.
- **Card that never loads if the observer never fires: acceptable.** The card stays on the placeholder, which is its designed loading and error footprint. It cannot be stuck on a spinner or blank. Hero is unaffected because it is eager.
- **Missing IntersectionObserver: WARN (non-blocking).** `use-in-view.ts:13` constructs `IntersectionObserver` with no `typeof` guard. In a runtime without it (jsdom, or a very old browser), a card render throws and unmounts the tree. All current browsers support it, and the card and image tests stub it (`recipe-card.test.tsx:11-13`, `recipe-image.test.tsx:20`). `RecipesPage.test.tsx` only renders an empty list, so it does not hit this. Any future test that renders a populated list without the stub will throw. Suggested fix: `if (typeof IntersectionObserver === "undefined") { setInView(true); return; }` in the effect, which also gives a safe fallback of loading eagerly. There is no SSR in this Vite SPA, and the constructor runs in an effect, so there is no SSR hazard.

## Checklist (changed files)

FE-01 PASS (no `any` in the diff). FE-02 PASS (`cn()` used, `recipe-image.tsx:27,54`). FE-03 PASS (item 4). FE-05 PASS (no `animate-spin`). FE-06 PASS (no color classes added). FE-08 PASS (named exports). FE-09 PASS (`useTenant()` plus `enabled` guard, `use-recipe-image-src.ts:17,21`). FE-10 PASS (tenant and household in the key). FE-14 PASS (`as const`, `:19`). FE-17 PASS (tests added or updated for the hook, component, card and page). FE-18 PASS (n/a; inline mocks updated for `downloadImage`). FE-19 PASS (the observer ref sits on a non-interactive placeholder; the hero trigger is a native `<button>`).

### Blocking (must fix)

None.

### Non-Blocking (should fix)

- Guard `typeof IntersectionObserver === "undefined"` in `use-in-view.ts:13` (see above).
