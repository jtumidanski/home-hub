# bug: recipe photos upload but never display (image GET resolves the zero tenant)

**Reproduced:** live k8s deployment, namespace `home-hub`, pod
`recipe-service-6456cbd56c-974cx` (image `home-hub-recipe-service:main`, built
from merge `9911c58`). Tenant `ff97efd6-7737-4808-921e-09d475194c97`, household
`f887024d-3cdc-4e92-aa74-378be10eaa07`, recipe
`cbeec321-78a7-4d0c-8f21-0575997888ca`. Steps: edit the recipe, attach a photo,
save.

**Observed:** the upload succeeds. The `PUT .../image` handler logs
`"recipe image stored" ... "full_bytes":72089,"thumb_bytes":45518`. Every
following `GET /api/v1/recipes/cbeec321-.../image` then runs its visibility
query under the nil tenant and 404s:

```
SELECT * FROM "recipes" WHERE (id = 'cbeec321-78a7-4d0c-8f21-0575997888ca' AND deleted_at IS NULL) AND tenant_id = '00000000-0000-0000-0000-000000000000' ...
record not found
```

The `<img>` `onError` fires, so the hero is hidden and the card shows the
placeholder. To the user it looks like attaching the photo did nothing.

**Expected:** the stored image renders on the recipe detail hero, the card
thumbnail, the lightbox and the edit form preview (PRD FR-12/16/17).

**Root cause:** design.md:22 assumed tenant and household come from the
`access_token` JWT claims. They do not. auth-service always issues tokens with
nil tenant and household (`services/auth-service/internal/authflow/processor.go:91`
and `:124` pass `[16]byte{}`), so `shared/go/auth/auth.go:128-145` always takes
them from the `X-Tenant-ID` / `X-Household-ID` headers. The SPA's api client
adds those headers (`frontend/src/lib/api/client.ts:134-147`). A bare
`<img src={image.fullUrl}>` sends only cookies, so every image GET runs with
`uuid.Nil` for both.

Backend tests passed because they inject tenant context directly. The live
acceptance steps that would have caught this were NOT RUN (see
`verification.md` items 3-7).

## Fix

Ruling (user decision): **fix it in the frontend only.** Fetch the image bytes
through the api client, which sends the tenant headers, and render an object
URL. No backend, shared-lib or API contract change. `thumbUrl` / `fullUrl`
keep their current shape (`/api/v1/recipes/{id}/image?size=...&v=...`, built in
`services/recipe-service/internal/recipeimage/keys.go:16`).

- `frontend/src/lib/hooks/api/use-recipes.ts`: add a `useRecipeImageBlob(url: string | null)`
  hook (or put it in a new `frontend/src/lib/hooks/use-recipe-image-src.ts` if
  that fits the hooks layout better). It uses React Query, keyed on the tenant,
  the household and the full URL. The URL already carries `?v=`, so a replaced
  image gets a new key. It is enabled only when `url` is non-null, and it calls
  `api.download(path)` (`frontend/src/lib/api/client.ts:426`) with `path` = the
  URL minus the leading `/api/v1` (the client's `BASE_URL`). It returns an
  object URL made with `URL.createObjectURL(blob)` and revokes it on change and
  on unmount. Use a long `staleTime`: a versioned URL is immutable.
- `frontend/src/components/features/recipes/recipe-image.tsx`: resolve `src` through
  the hook. While loading, a card renders the placeholder (keeping its
  footprint) and a hero renders nothing. On a query error, behave exactly like
  today's `onError`: hero `null`, card placeholder. Keep `loading`, the
  `className` handling and `alt`.
- `frontend/src/pages/RecipeDetailPage.tsx:200-204`: the lightbox `<img src={attrs.image.fullUrl}>`
  must use the same resolved object URL. Use the hook, or reuse the hero's
  value.
- `frontend/src/lib/hooks/use-recipe-image-draft.ts:46-47`: `previewSrc` falls back to
  `existing?.fullUrl`, a server URL that has the same bug in
  `RecipeImageField` (`recipe-image-field.tsx:54`) and the form preview
  (`RecipeFormPage.tsx:201-207`). Resolve the existing image through the hook.
  The `staged` local blob preview URL must pass through unchanged, because it
  is already an object URL.
- Tests (vitest, colocated `__tests__` per the existing layout):
  - Hook test: mock `api.download`. Assert it is called with
    `/recipes/<id>/image?size=thumb&v=...` and the tenant-aware key, that an
    object URL is returned, and that it is revoked on unmount. **This test
    must fail before the fix**: today there is no fetch, so no request carries
    the tenant headers.
  - Update the existing `recipe-image` / `RecipeDetailPage` / `RecipeImageField` tests
    that assert `img.src === fullUrl/thumbUrl` so they assert the object URL instead.

## Not yet answered

- Whether `api.download` dedupes or caches. React Query is the cache layer
  here; do not add caching to the client.
- If the hook's placement (inside `use-recipes.ts` or its own file) conflicts with
  the frontend guidelines (`frontend-dev-guidelines` skill), follow the guidelines.
- Do not change `design.md`. A one-line erratum under its line 22 can be added
  in the outcome section below by the controller.

## Fix round 2 (review follow-ups)

Source: the frontend audit dated 2026-10-15 in `audit.md` (commit `82357d6`)
on round 1 (`3b42157`). The user approved fixing all of them on this branch.

1. **Restore lazy loading for cards.** `frontend/src/components/features/recipes/recipe-image.tsx`:
   for `variant="card"`, enable the query only once the placeholder element
   is near the viewport. Use an `IntersectionObserver` with a `rootMargin` of
   about `200px`, and stop observing once it is visible. Check
   `frontend/src/lib/hooks/` for an existing in-view hook before writing one.
   `variant="hero"` stays eager. Add `enabled` (default `true`) as an option on
   `useRecipeImageSrc` (`frontend/src/lib/hooks/use-recipe-image-src.ts`).
   jsdom has no `IntersectionObserver`, so stub it in the tests. Add a test
   that an off-screen card makes no download call until it intersects.
2. **No flicker on recipe invalidation.** Move the image query key out of
   `recipeKeys.all(...)`, so `useInvalidateRecipes().invalidateAll`
   (`frontend/src/lib/hooks/api/use-recipes.ts:179`) no longer matches it. Use
   something like `["recipe-images", tenantId, householdId, url]` and keep
   tenant and household in the key. Versioned URLs are immutable, so a replaced
   image arrives as a new `?v=` key. Add a test that `invalidateAll` does not
   refetch the image query.
3. **Lightbox before the image loads.** `frontend/src/pages/RecipeDetailPage.tsx`:
   the hero/lightbox trigger must not open, or must not render, until the
   resolved `src` exists, and the lightbox `<img>` must never get
   `src={undefined}`. Test it.
4. **FE-03.** Move the `api.download` call behind the recipe service in
   `frontend/src/services/api/recipe.ts` (e.g. `getImage(url)` / `downloadImage`,
   which strips the `/api/v1` prefix there). The hook calls the service, not
   `api` directly. Update the hook test's mock to match.

## Outcome

Branch `task-056-image-tenant-fix` (from `main` at `9911c58`; the original task
branch had already merged in PR #13). Fix PR: #14.

| Step | Agent / tool | Result |
|---|---|---|
| Round 1 fix | task-implementer (sonnet) | DONE, `3b42157`: `useRecipeImageSrc` fetches via the api client and renders an object URL; card, hero, lightbox and form preview wired |
| Round 1 review | frontend-guidelines-reviewer (sonnet) | PASS, `82357d6`: 4 non-blocking findings (lost lazy loading, flicker on invalidate, early lightbox open, FE-03) |
| Round 2 fix | task-implementer (sonnet) | DONE, `05a801e`: all 4 findings fixed (`use-in-view.ts`, image key outside `recipeKeys`, lightbox gated on src, `recipeService.downloadImage`) |
| Round 2 review | frontend-guidelines-reviewer (sonnet) | PASS, `ce0f35e`: 1 non-blocking finding (no `typeof IntersectionObserver` guard in `use-in-view.ts:13`; test-only exposure) |
| Gate | flagless `tools/verify.sh --base 9911c58` at `05a801e` | **PASSED**, all legs (pins, build, vet, test, lint, frontend-build, eslint, docker) |

An earlier flagless run reported 14 failures in `recipe-card` /
`recipe-image` tests. It was discarded because it overlapped round 2's
uncommitted edits (750 tests counted vs 746 at `3b42157`). The re-run on the
committed `05a801e` passed.

Not done here: the hook test was written before the hook existed, so it was
never shown to fail against the old components.

**Live re-test:** pending. Deploy, then attach a photo to a recipe. Confirm
`GET /api/v1/recipes/<id>/image` returns 200 in the recipe-service logs, with no
`tenant_id = '00000000-...'` visibility query, and that the photo shows on the
card, the detail page and the lightbox.

**Design erratum:** design.md:22 is wrong. auth-service issues JWTs with nil
tenant/household, so tenant scoping always rides on the `X-Tenant-ID` /
`X-Household-ID` headers. Any browser-native fetch (`<img>`, `<a download>`,
`EventSource`) will not carry them.
