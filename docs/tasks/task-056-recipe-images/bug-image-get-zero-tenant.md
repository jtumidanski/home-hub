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

## Outcome

_Pending._
