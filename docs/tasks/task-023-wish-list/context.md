# Wish List — Context

Last Updated: 2026-04-08

Companion to `plan.md` and `prd.md`. Captures the concrete files, conventions, and decisions an implementer needs without re-reading the codebase.

---

## Key Files

### shopping-service (Go) — to create / modify

| File | Action | Purpose |
|------|--------|---------|
| `services/shopping-service/internal/wishlist/model.go` | create | Immutable domain model + urgency constants |
| `services/shopping-service/internal/wishlist/entity.go` | create | GORM entity, `TableName()`, `Migration(db)` |
| `services/shopping-service/internal/wishlist/builder.go` | create | Fluent builder + validation |
| `services/shopping-service/internal/wishlist/builder_test.go` | create | Builder validation tests |
| `services/shopping-service/internal/wishlist/processor.go` | create | Use cases: Create/Update/Delete/Vote/List |
| `services/shopping-service/internal/wishlist/processor_test.go` | create | Processor + concurrency tests |
| `services/shopping-service/internal/wishlist/provider.go` | create | DB queries, atomic vote increment |
| `services/shopping-service/internal/wishlist/rest.go` | create | JSON:API DTOs + Transform |
| `services/shopping-service/internal/wishlist/resource.go` | create | `InitializeRoutes`, HTTP handlers |
| `services/shopping-service/cmd/main.go` | modify | Register `wishlist.Migration` and `wishlist.InitializeRoutes` |

### Reference files (read before writing each layer)

| File | Why |
|------|-----|
| `services/shopping-service/internal/list/entity.go` | Migration pattern, GORM struct shape |
| `services/shopping-service/internal/list/model.go` | Immutable model conventions |
| `services/shopping-service/internal/list/builder.go` | Builder pattern + validation style |
| `services/shopping-service/internal/list/processor.go` | Use-case shape, ctx + tenant/household/user args |
| `services/shopping-service/internal/list/provider.go` | Query scoping, GORM patterns |
| `services/shopping-service/internal/list/resource.go` | Route registration, `tenantctx.MustFromContext` usage |
| `services/shopping-service/internal/list/rest.go` | JSON:API DTO + `Transform()` shape |
| `services/shopping-service/internal/item/*.go` | Second example with similar conventions |

### frontend — to create / modify

| File | Action | Purpose |
|------|--------|---------|
| `frontend/src/types/models/wish-list.ts` | create | `WishListItem`, `Urgency` types |
| `frontend/src/lib/schemas/wish-list.schema.ts` | create | Zod schemas for create/update |
| `frontend/src/services/api/wish-list.ts` | create | `WishListService` API client |
| `frontend/src/lib/hooks/api/use-wish-list.ts` | create | TanStack Query hooks |
| `frontend/src/pages/WishListPage.tsx` | create | The page itself (flat, not in `pages/shopping/` subdir) |
| `frontend/src/components/features/navigation/nav-config.ts` | modify | Remove `Shopping` from `lifestyle`; add new top-level `shopping` group with `Grocery Lists` → `/app/shopping/grocery` and `Wish List` → `/app/shopping/wish-list` |
| `frontend/src/App.tsx` | modify | Relocate grocery routes to `shopping/grocery` and `shopping/grocery/:id`; add redirect at `shopping` → `shopping/grocery`; register new `shopping/wish-list` route |

### Frontend reference files

| File | Why |
|------|-----|
| `frontend/src/pages/ShoppingListsPage.tsx` | Page-level conventions |
| `frontend/src/services/api/shopping.ts` | API client conventions, JSON:API parsing |
| `frontend/src/lib/hooks/api/use-shopping.ts` | Hook patterns, query keys, optimistic updates |
| `frontend/src/types/models/shopping.ts` | Model type conventions |
| `frontend/src/lib/schemas/shopping.schema.ts` | Zod schema conventions |

---

## Conventions to Follow

### Backend
- **Layer order:** model → entity → builder → processor → provider → rest → resource. Each layer only depends on layers below it.
- **Immutable models:** unexported fields + getters; no setters; mutation goes through the builder.
- **Tenant + household scoping:** every provider query filters by both `tenant_id` and `household_id`. Pull from `tenantctx.MustFromContext(r.Context())` in the resource layer; pass IDs explicitly into the processor.
- **Migrations:** an inline `Migration(db *gorm.DB) error` function on the entity, registered in `cmd/main.go` via `database.SetMigrations(...)`.
- **JSON:API:** type string is `"wish-items"`. DTOs in `rest.go` with `Transform()` helpers between entity ↔ model ↔ DTO.
- **Atomic vote increment:** raw `db.Exec("UPDATE shopping.wish_list_items SET vote_count = vote_count + 1, updated_at = NOW() WHERE id = ? AND tenant_id = ? AND household_id = ?", ...)` (or `gorm.Expr("vote_count + 1")` via `Updates`). NO read-modify-write.
- **PATCH must reject `vote_count`:** in the DTO transform layer, error out if the attribute is present rather than silently dropping.

### Frontend
- **Mobile UI rules** (per `feedback_ui-preferences.md`): tap-only, no swipe gestures, card-based mobile tables. Vote button must be a large tap target.
- **Form pattern:** shadcn `Dialog` + `react-hook-form` + Zod resolver.
- **Hooks:** TanStack Query with stable query keys; mutations invalidate the list query on success. Vote mutation uses optimistic update on `vote_count`, rolls back on error.
- **Routes:** registered in `App.tsx` under the `/app` parent route.
- **Sidebar groups:** array of `{ key, label, items: [{ to, icon, label }] }` in `nav-config.ts`.

---

## Decisions

| # | Decision | Resolution |
|---|----------|------------|
| D1 | Per-user vote attribution? | No. Anonymous shared counter only (PRD §2 non-goals). |
| D2 | Vote decrement / undo? | No. (PRD §2.) |
| D3 | Multiple wish lists per household? | No. One implicit list per `(tenant_id, household_id)`. |
| D4 | Separate `wish_lists` parent table? | No. Implicit list, scoped via composite key. |
| D5 | `wish_item_votes` table? | Not in v1 — no per-user tracking needed. |
| D6 | Default urgency on create? | `want`. |
| D7 | Sort order? | `vote_count DESC, created_at ASC`. Server-side. |
| D8 | Authorization model? | Any household member can do anything to any item. |
| D9 | Grocery list route after rename? | Relocate to `/app/shopping/grocery` with a redirect from `/app/shopping`. Detail route also moves to `/app/shopping/grocery/:id`. New wish list lives at `/app/shopping/wish-list`. |
| D10 | Page file location? | `frontend/src/pages/WishListPage.tsx` (flat, matches existing). |
| D11 | Migration approach? | Inline `entity.Migration` registered in `cmd/main.go`, consistent with `list` and `item` domains. |
| D12 | Empty state copy? | **"Nothing on the wish list yet. Add the first thing you've been eyeing."** |

---

## Dependencies

- **No new external libraries** (Go or npm).
- **No nginx changes** — `/api/v1/shopping/*` already routes to shopping-service.
- **No other services touched** — category-service, recipe-service, etc. are out of scope.
- **DB schema:** `shopping` schema must already exist (it does; `list` and `item` live there).

---

## Out of Scope (for v1)

Pulled forward from PRD §2 and §9 so they don't accidentally creep in:

- Multiple named wish lists per household.
- One-vote-per-user, vote caps, or vote undo.
- Vote downvoting / decrementing.
- Converting wish items into grocery items.
- Price tracking, links, images, product metadata.
- Notifications, comments, activity feed.
- Household-configurable voting rules (possible v2).
---

## Open Items Inherited from PRD §9

All resolved — see Decisions table (D9, D10, D12).

---

## Testing Notes

- **Concurrent vote test** is the most important new test. It catches the worst class of bug here (lost vote updates) and is cheap once a real test DB is wired.
- **Cross-household isolation test** in the processor: create items under household A, list under household B, expect zero results.
- **PATCH `vote_count` rejection test** to lock the invariant.
- **Manual smoke test** via `scripts/local-up.sh` covering every PRD §10 acceptance criterion.
