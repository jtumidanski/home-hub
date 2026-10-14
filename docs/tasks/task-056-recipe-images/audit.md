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
