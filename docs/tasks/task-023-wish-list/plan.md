# Wish List — Implementation Plan

Last Updated: 2026-04-08
Status: Draft
Source PRD: `docs/tasks/task-023-wish-list/prd.md`

---

## 1. Executive Summary

Add a household-shared **Wish List** feature to Home Hub. Members can add aspirational purchase items, edit/delete any item, and tap a vote button (anonymous, uncapped) to express enthusiasm. Items are sorted by vote count, with `created_at ASC` as the tie-breaker. The frontend sidebar is reorganized so a new top-level **Shopping** group contains the existing grocery list page (renamed **Grocery Lists**) and the new **Wish List** page.

The work is contained to two components:

- **shopping-service** (Go): new `wishlist` domain (model/entity/builder/processor/provider/rest), GORM migration for `wish_list_items`, five REST endpoints under `/api/v1/shopping/wish-list`, atomic vote-increment SQL.
- **frontend** (React/TS): new page, hooks, API client module, schema, sidebar restructure, route registration.

No nginx, category-service, or recipe-service changes are required.

## 2. Current State Analysis

### shopping-service (`services/shopping-service/`)
- DDD layout with two existing domains: `internal/list/` and `internal/item/`. Each has `model.go`, `entity.go`, `builder.go`, `processor.go`, `provider.go`, `resource.go`, `rest.go`, plus tests.
- Migrations are inline `Migration(db)` functions on each entity, registered in `cmd/main.go` via `database.SetMigrations(list.Migration, item.Migration, ...)`.
- REST routes are mounted via `<domain>.InitializeRoutes(db, ...)(l, si, api)` in `cmd/main.go`. Existing routes live under `/shopping/lists`, `/shopping/items`, etc.
- Tenant + household + user are pulled from request context via `tenantctx.MustFromContext(r.Context())` inside resource handlers and passed into processor calls. GORM callbacks scope queries by tenant.
- No existing concept of an aspirational wish list — only grocery shopping lists tied to a trip.

### frontend (`frontend/src/`)
- Sidebar config is a flat array of groups in `components/features/navigation/nav-config.ts`. The current `Shopping` entry sits inside the `lifestyle` group at `/app/shopping`.
- Existing shopping UI: `pages/ShoppingListsPage.tsx`, `pages/ShoppingListDetailPage.tsx`, with hooks in `lib/hooks/api/use-shopping.ts`, API client in `services/api/shopping.ts`, models in `types/models/shopping.ts`, and Zod schemas in `lib/schemas/shopping.schema.ts`.
- Routes are registered in `App.tsx` lines 62–63 (`shopping` and `shopping/:id`).
- Mobile rules from `feedback_ui-preferences.md`: tap-only, no swipe gestures, card-based mobile tables.

### nginx
- The `/api/v1/shopping/*` prefix already routes to shopping-service. No ingress changes needed.

## 3. Proposed Future State

### Backend
- New `internal/wishlist/` domain in shopping-service with the standard six layers and tests, mirroring `internal/list/` conventions.
- New table `shopping.wish_list_items` (schema columns and indexes per PRD §6).
- Five new endpoints under `/api/v1/shopping/wish-list/items`:
  - `GET /items`
  - `POST /items`
  - `PATCH /items/{id}`
  - `DELETE /items/{id}`
  - `POST /items/{id}/vote`
- Vote endpoint uses atomic `UPDATE ... SET vote_count = vote_count + 1 WHERE id = ? AND tenant_id = ? AND household_id = ?` (no read-modify-write).
- PATCH explicitly drops/rejects any `vote_count` field on input.

### Frontend
- New page `pages/shopping/WishListPage.tsx` (new `pages/shopping/` subdir, OR collocated with existing shopping pages — see Risk R3).
- New hooks: `useWishListItems`, `useCreateWishListItem`, `useUpdateWishListItem`, `useDeleteWishListItem`, `useVoteWishListItem` in `lib/hooks/api/use-wish-list.ts`.
- New API client `services/api/wish-list.ts`, model types in `types/models/wish-list.ts`, Zod schemas in `lib/schemas/wish-list.schema.ts`.
- Sidebar restructure: remove `Shopping` from `lifestyle` group; add new top-level group `shopping` with two items: `Grocery Lists` (`/app/shopping/grocery`) and `Wish List` (`/app/shopping/wish-list`).
- Grocery list route relocated from `/app/shopping` to `/app/shopping/grocery`, with a redirect at `/app/shopping` → `/app/shopping/grocery` so existing bookmarks keep working.
- New route `shopping/wish-list` registered in `App.tsx`.
- Card-based mobile layout, large tap target for vote button, optimistic vote updates.

## 4. Implementation Phases

The phases are ordered so the backend stabilizes before the frontend consumes it. Phase 4 (sidebar) is small and can land independently if convenient.

### Phase 1 — Backend Domain & Migration
Build the `wishlist` domain end-to-end inside shopping-service so the API is callable via curl/Postman before any UI exists.

### Phase 2 — Backend Tests & Hardening
Add unit tests for the processor and provider, plus a focused test for the atomic vote increment under concurrency. Verify multi-tenant isolation.

### Phase 3 — Frontend Data Layer
Add types, Zod schema, API client, and TanStack Query hooks. No UI yet.

### Phase 4 — Frontend Sidebar Reorg
Update `nav-config.ts` to introduce the `shopping` top-level group and rename the entry. Ship-able independently.

### Phase 5 — Frontend Wish List Page
Build the page, item card, urgency badge, vote button, and create/edit dialog. Wire to hooks. Add route to `App.tsx`.

### Phase 6 — Verification & Polish
Build all affected services, run lint/tests, smoke-test in local env via `scripts/local-up.sh`.

---

## 5. Detailed Tasks

Effort scale: **S** (≤1h focused work), **M** (half day), **L** (full day), **XL** (multi-day).

### Phase 1 — Backend Domain & Migration

**1.1 Create `wishlist` package skeleton** — S
- Create `services/shopping-service/internal/wishlist/` with empty `model.go`, `entity.go`, `builder.go`, `processor.go`, `provider.go`, `resource.go`, `rest.go`.
- **Acceptance:** package compiles (empty stubs OK).

**1.2 Define entity + migration** — S
- `entity.go`: GORM struct `WishListItemEntity` with columns from PRD §6, `TableName() = "shopping.wish_list_items"`, and `Migration(db)` function that runs `AutoMigrate` plus creates the listing index `(tenant_id, household_id, vote_count DESC, created_at ASC)`.
- Register `wishlist.Migration` in `cmd/main.go` `database.SetMigrations(...)`.
- **Acceptance:** service starts, table appears in `shopping` schema with all columns + check constraints + index.

**1.3 Define immutable model** — S
- `model.go`: `Model` struct with private fields and getters; `urgency` constants (`UrgencyMustHave`, `UrgencyNeedToHave`, `UrgencyWant`).
- **Acceptance:** model has no setters; getters cover all PRD response attributes.

**1.4 Builder + validation** — S
- `builder.go`: fluent builder with `SetName`, `SetPurchaseLocation`, `SetUrgency`, `Build() (Model, error)`.
- Validation: name non-empty after trim & ≤255; purchase_location ≤255; urgency ∈ enum; default urgency = `want`.
- **Acceptance:** unit tests in `builder_test.go` cover happy path + each validation failure.

**1.5 Provider (data access)** — M
- `provider.go`: `GetById`, `ListByHousehold` (sorted vote_count DESC, created_at ASC), `Create`, `Update`, `Delete`, `IncrementVote`.
- `IncrementVote` MUST issue a single `UPDATE shopping.wish_list_items SET vote_count = vote_count + 1, updated_at = NOW() WHERE id = ? AND tenant_id = ? AND household_id = ? RETURNING ...` (or equivalent two-statement pattern that re-reads).
- All queries scoped by tenant + household.
- **Acceptance:** no read-modify-write in vote path; queries use parameterized SQL.

**1.6 Processor (business logic)** — M
- `processor.go`: `Create`, `Update`, `Delete`, `Vote`, `List` taking `tenantId`, `householdId`, `userId` and returning models or errors.
- `Update` MUST silently ignore (or explicitly drop) any incoming `vote_count`.
- **Acceptance:** processor functions take ctx + tenant/household and return immutable models; updates never touch `vote_count`.

**1.7 REST resource + DTOs** — M
- `rest.go`: JSON:API request/response DTOs with type `"wish-items"`, plus `Transform()` between entity/model and DTO.
- `resource.go`: `InitializeRoutes` registering the five endpoints under `/shopping/wish-list/items`.
- Vote endpoint POST `/items/{id}/vote` with no body, returns updated item.
- PATCH endpoint MUST reject (400) if `vote_count` is present in attributes.
- **Acceptance:** all five endpoints respond, return JSON:API shapes, and enforce tenant + household via `tenantctx`.

**1.8 Wire routes in main.go** — S
- Call `wishlist.InitializeRoutes(db, ...)(l, si, api)` in `cmd/main.go`.
- **Acceptance:** routes appear under `/api/v1/shopping/wish-list/...` and respond after restart.

### Phase 2 — Backend Tests & Hardening

**2.1 Processor unit tests** — M
- Cover create defaults urgency to `want`, vote_count defaults to 0; update rejects vote_count; delete is hard delete; vote increments by exactly 1; cross-household isolation.
- **Acceptance:** `go test ./internal/wishlist/...` passes.

**2.2 Concurrent vote test** — M
- Spin N goroutines hitting `IncrementVote` against a real test DB; assert final count == N (no lost updates).
- **Acceptance:** test consistently passes locally with N ≥ 50.

**2.3 PATCH vote_count rejection test** — S
- Integration test ensures PATCH with `vote_count` in attributes returns 400 and does not mutate the row.
- **Acceptance:** test passes.

### Phase 3 — Frontend Data Layer

**3.1 Types + schemas** — S
- `frontend/src/types/models/wish-list.ts` — `WishListItem`, `Urgency` union.
- `frontend/src/lib/schemas/wish-list.schema.ts` — Zod schemas for create + update inputs.
- **Acceptance:** TS compiles, schemas validate sample inputs.

**3.2 API client** — M
- `frontend/src/services/api/wish-list.ts` — `WishListService` class with `list`, `create`, `update`, `remove`, `vote` methods, JSON:API shapes mirroring shopping service.
- **Acceptance:** methods compile; manually callable from devtools.

**3.3 TanStack Query hooks** — M
- `frontend/src/lib/hooks/api/use-wish-list.ts` — `useWishListItems`, `useCreateWishListItem`, `useUpdateWishListItem`, `useDeleteWishListItem`, `useVoteWishListItem`.
- Vote hook uses optimistic update on `vote_count`; rolls back on error.
- **Acceptance:** hooks invalidate the list query on mutate; optimistic vote bumps the visible count instantly.

### Phase 4 — Frontend Sidebar Reorg

**4.1 Update `nav-config.ts`** — S
- Remove existing `Shopping` entry from the `lifestyle` group.
- Add a new top-level group `{ key: "shopping", label: "Shopping", items: [...] }` with two items: `Grocery Lists` → `/app/shopping/grocery` and `Wish List` → `/app/shopping/wish-list`.
- **Acceptance:** sidebar renders the new group with both entries pointing at the new routes.

**4.2 Relocate grocery route + redirect** — S
- In `App.tsx`, change the existing grocery route from `shopping` to `shopping/grocery`.
- Add a redirect from `shopping` → `shopping/grocery` (e.g. `<Route path="shopping" element={<Navigate to="/app/shopping/grocery" replace />} />`) so old bookmarks keep working.
- Keep the existing `shopping/:id` detail route working — relocate it to `shopping/grocery/:id` to stay consistent.
- **Acceptance:** `/app/shopping` redirects to `/app/shopping/grocery`; existing list view + detail view still load.

### Phase 5 — Frontend Wish List Page

**5.1 Page scaffold + route** — S
- New file `frontend/src/pages/WishListPage.tsx` (collocated with existing shopping pages — see R3).
- Register `<Route path="shopping/wish-list" element={<WishListPage />} />` in `App.tsx`.
- **Acceptance:** navigating to the route renders an empty placeholder page.

**5.2 Item list + card** — M
- Render items via `useWishListItems`, sorted server-side. Each row shows name, purchase location (if any), urgency badge, vote count, vote button, edit/delete actions.
- Card-based layout on mobile (per `feedback_ui-preferences.md`); tap-only, no swipe.
- **Acceptance:** list shows items in correct order; mobile renders cards.

**5.3 Vote button with optimistic update** — M
- Large tap target. Each tap fires `useVoteWishListItem` with optimistic +1. No debounce — every tap is a request, per PRD.
- **Acceptance:** rapid taps from a phone show counts incrementing immediately and converge with the server count.

**5.4 Create / edit dialog** — M
- shadcn `Dialog` + `react-hook-form` with the wish-list Zod schema. Fields: name (required), purchase_location, urgency (radio or select).
- **Acceptance:** create + edit flows work; validation errors render inline.

**5.5 Delete confirmation** — S
- Confirm before delete. Hard delete via hook.
- **Acceptance:** delete removes the item and refreshes the list.

**5.6 Empty state** — S
- Render an empty-state component when the list is empty rather than an error.
- Copy: **"Nothing on the wish list yet. Add the first thing you've been eyeing."**
- **Acceptance:** new household sees the empty state with the locked copy, not a spinner or error.

### Phase 6 — Verification & Polish

**6.1 Build all affected services** — S
- `go build ./...` in shopping-service; `npm run build` (or equivalent) in frontend.
- **Acceptance:** clean build for both.

**6.2 Lint + tests** — S
- Run shopping-service tests and frontend lint/typecheck.
- **Acceptance:** all green.

**6.3 Local smoke test** — M
- `scripts/local-up.sh`, log in to local, verify: sidebar reorg, create item, edit, delete, vote, sort order, multi-tenant isolation (switch household if practical).
- **Acceptance:** all PRD §10 acceptance criteria pass manually.

**6.4 Tear down** — S
- `scripts/local-down.sh`.

---

## 6. Risk Assessment & Mitigation

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|------------|--------|------------|
| R1 | Vote race condition causes lost updates if implemented as read-modify-write | M | H | Phase 1.5 mandates atomic SQL `UPDATE ... SET vote_count = vote_count + 1`; Phase 2.2 adds a concurrent test that would catch regressions. |
| R2 | PATCH allows clients to silently set `vote_count`, breaking the "votes only via /vote" invariant | M | M | Reject in DTO transform (Phase 1.7) + integration test in Phase 2.3. |
| R3 | Frontend page placement decision (`pages/shopping/WishListPage.tsx` subdir vs. flat `pages/WishListPage.tsx`) inconsistent with existing structure | L | L | Existing shopping pages live flat under `pages/`. Default to flat (`pages/WishListPage.tsx`) for consistency unless reviewer prefers a subdir. |
| R4 | Grocery route relocation `/app/shopping` → `/app/shopping/grocery` breaks bookmarks | M | M | Phase 4.2 adds a redirect at the old path. Detail route also relocated to `shopping/grocery/:id`. |
| R5 | Optimistic vote UI drifts from server count under flaky network | M | L | On error, roll back optimistic update and refetch the list. Final list query is the source of truth. |
| R6 | Tenant/household isolation regression — wish items leak across households | L | H | Provider scopes every query by `tenant_id` + `household_id`; processor tests assert cross-household isolation. |
| R7 | Migration adds an unindexed table, causing slow listing once items grow | L | L | Phase 1.2 creates the `(tenant_id, household_id, vote_count DESC, created_at ASC)` index in the same migration. |
| R8 | Atomic increment via GORM ORM semantics may fall back to read-modify-write if implemented carelessly | M | H | Use raw SQL via `db.Exec` or `gorm.Expr("vote_count + 1")` in `Updates`; verify the executed SQL in tests. |

## 7. Success Metrics

- All PRD §10 acceptance criteria checked off after Phase 6 smoke test.
- Concurrent vote test (Phase 2.2) passes with N=50 with zero lost increments.
- Shopping-service build, tests, and frontend build all green.
- Manual UX check: a wish list with 5 items renders, sorts, votes, edits, and deletes correctly on both desktop and mobile viewport.
- Sidebar shows the new "Shopping" group at the top level; "Lifestyle" no longer contains a "Shopping" entry.

## 8. Required Resources & Dependencies

**Code areas touched:**
- `services/shopping-service/internal/wishlist/` (new package)
- `services/shopping-service/cmd/main.go` (migration + route registration)
- `frontend/src/pages/WishListPage.tsx` (new)
- `frontend/src/lib/hooks/api/use-wish-list.ts` (new)
- `frontend/src/services/api/wish-list.ts` (new)
- `frontend/src/types/models/wish-list.ts` (new)
- `frontend/src/lib/schemas/wish-list.schema.ts` (new)
- `frontend/src/components/features/navigation/nav-config.ts` (modify)
- `frontend/src/App.tsx` (modify)

**External dependencies:** none new. Reuses existing tenant/JWT middleware, GORM, JSON:API response helpers, shadcn/ui components, TanStack Query.

**Tooling:**
- `scripts/local-up.sh` / `scripts/local-down.sh` for local env (per memory).
- Standard `go build`, `go test`, frontend `npm run build` / lint.

## 9. Timeline Estimate

Effort, not wall clock:

| Phase | Effort |
|-------|--------|
| Phase 1 — Backend Domain & Migration | M–L |
| Phase 2 — Backend Tests & Hardening | M |
| Phase 3 — Frontend Data Layer | M |
| Phase 4 — Sidebar Reorg | S |
| Phase 5 — Wish List Page | L |
| Phase 6 — Verification & Polish | S–M |

## 10. Resolved Decisions

All PRD §9 open questions are resolved:

1. **Route relocation** — RESOLVED: relocate grocery list from `/app/shopping` to `/app/shopping/grocery`. Add a redirect at `/app/shopping` so old bookmarks keep working. Detail route moves to `/app/shopping/grocery/:id`.
2. **Page file location** — RESOLVED: flat at `frontend/src/pages/WishListPage.tsx`, consistent with existing `ShoppingListsPage.tsx`.
3. **Empty state copy** — RESOLVED: **"Nothing on the wish list yet. Add the first thing you've been eyeing."**
