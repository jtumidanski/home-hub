# Wish List — Task Checklist

Last Updated: 2026-04-08

Companion to `plan.md`. Check items off as you complete them. Each task references its plan section for full acceptance criteria.

---

## Pre-flight

- [x] PRD §9 open decisions resolved — see context.md D9, D10, D12 (relocate grocery to `/app/shopping/grocery` with redirect; flat page file; empty state copy locked)
- [ ] Read reference files in `context.md` for the layer being implemented

## Phase 1 — Backend Domain & Migration

- [ ] **1.1** Create empty `internal/wishlist/` package skeleton (S)
- [ ] **1.2** Define `WishListItemEntity` + `Migration(db)`; register in `cmd/main.go` (S)
- [ ] **1.3** Define immutable `Model` + urgency constants (S)
- [ ] **1.4** Implement builder + validation; add `builder_test.go` (S)
- [ ] **1.5** Implement provider including atomic `IncrementVote` SQL (M)
- [ ] **1.6** Implement processor (Create/Update/Delete/Vote/List); reject `vote_count` in Update (M)
- [ ] **1.7** Implement REST DTOs + handlers; PATCH rejects `vote_count`; type `"wish-items"` (M)
- [ ] **1.8** Wire `wishlist.InitializeRoutes` in `cmd/main.go` (S)

**Phase 1 done when:** all five endpoints respond under `/api/v1/shopping/wish-list/items` against a freshly migrated DB.

## Phase 2 — Backend Tests & Hardening

- [ ] **2.1** Processor unit tests (defaults, isolation, hard delete) (M)
- [ ] **2.2** Concurrent vote test with N≥50 goroutines, asserts zero lost updates (M)
- [ ] **2.3** Integration test: PATCH with `vote_count` returns 400, row unchanged (S)

**Phase 2 done when:** `go test ./internal/wishlist/...` is green and concurrent vote test consistently passes.

## Phase 3 — Frontend Data Layer

- [ ] **3.1** Add `types/models/wish-list.ts` and `lib/schemas/wish-list.schema.ts` (S)
- [ ] **3.2** Add `services/api/wish-list.ts` `WishListService` (M)
- [ ] **3.3** Add `lib/hooks/api/use-wish-list.ts` with all five hooks; vote hook is optimistic (M)

**Phase 3 done when:** hooks compile, query keys are stable, and a manual call from devtools returns data.

## Phase 4 — Frontend Sidebar Reorg & Route Relocation

- [ ] **4.1** Update `nav-config.ts`: remove `Shopping` from `lifestyle`; add new top-level `shopping` group with `Grocery Lists` → `/app/shopping/grocery` and `Wish List` → `/app/shopping/wish-list` (S)
- [ ] **4.2** In `App.tsx`, relocate grocery routes to `shopping/grocery` and `shopping/grocery/:id`; add a `<Navigate>` redirect at `shopping` → `shopping/grocery` (S)

**Phase 4 done when:** sidebar shows the new group; `/app/shopping` redirects to `/app/shopping/grocery`; both list and detail views still load.

## Phase 5 — Frontend Wish List Page

- [ ] **5.1** Scaffold `pages/WishListPage.tsx`; register `shopping/wish-list` route in `App.tsx` (S)
- [ ] **5.2** Item list + card layout; mobile cards; tap-only (M)
- [ ] **5.3** Vote button with optimistic +1 per tap; large tap target (M)
- [ ] **5.4** Create / edit dialog (Dialog + react-hook-form + Zod) (M)
- [ ] **5.5** Delete with confirmation (S)
- [ ] **5.6** Empty state component with copy: "Nothing on the wish list yet. Add the first thing you've been eyeing." (S)

**Phase 5 done when:** all PRD §10 UI-related criteria are visually verifiable on desktop and mobile widths.

## Phase 6 — Verification & Polish

- [ ] **6.1** `go build ./...` in shopping-service is clean (S)
- [ ] **6.2** Frontend `npm run build` + lint + typecheck clean (S)
- [ ] **6.3** Local smoke test via `scripts/local-up.sh` covering every PRD §10 acceptance criterion (M)
- [ ] **6.4** `scripts/local-down.sh` to tear down (S)

## PRD §10 Acceptance Criteria (final gate)

- [ ] `wish_list_items` table exists in `shopping` schema with PRD §6 columns and constraints
- [ ] Five endpoints respond under `/api/v1/shopping/wish-list`
- [ ] All endpoints enforce tenant + household scoping from JWT context
- [ ] Create without `urgency` defaults to `want`
- [ ] Create initializes `vote_count` to 0
- [ ] PATCH rejects any attempt to set `vote_count`
- [ ] POST `/items/{id}/vote` increments by exactly 1 atomically and returns updated item
- [ ] Concurrent votes from multiple clients produce correct final count (no lost updates)
- [ ] Any household member can edit / delete any item regardless of `created_by`
- [ ] GET `/items` returns items sorted `vote_count DESC, created_at ASC`
- [ ] Sidebar has new top-level `Shopping` group with `Grocery Lists` and `Wish List`
- [ ] `Lifestyle` group no longer contains a `Shopping` entry
- [ ] Wish list page shows name, location, urgency badge, vote count, vote button, edit/delete
- [ ] Vote button can be tapped repeatedly with optimistic feedback
- [ ] Empty wish list renders an empty state, not an error
