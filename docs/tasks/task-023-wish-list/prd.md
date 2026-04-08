# Wish List — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-08

---

## 1. Overview

The Wish List feature gives a household a single shared collection of items its members would like to purchase eventually — things they can't currently afford, haven't had time to investigate, or are still deciding on. It is intentionally distinct from the grocery shopping list: wish-list items are aspirational, long-lived, and not tied to a specific shopping trip.

Any household member can add an item, delete an item, edit an item's fields, or upvote an item by tapping a vote button. Items are displayed sorted by vote count (descending) so the household's collective enthusiasm naturally surfaces what to act on next. Each item carries an optional purchase location and an urgency label (must / need / want) that is purely informational in v1.

This feature also reorganizes the sidebar: a new top-level **Shopping** group will contain both the existing grocery list page (renamed) and the new wish list page.

## 2. Goals

Primary goals:
- Provide a single shared wish list per household for tracking aspirational purchases
- Allow any household member to add, edit, delete, and upvote items
- Sort items by vote count so the most-wanted items rise to the top
- Reorganize the frontend sidebar to group grocery lists and wish list under a single "Shopping" category

Non-goals:
- Multiple named wish lists per household
- Per-user vote attribution, one-vote-per-user rules, or vote undo (raw shared counter only)
- Decrementing votes or "downvoting"
- Converting wish list items into grocery list items or purchase orders
- Price tracking, links, images, or product metadata beyond name and location
- Notifications, comments, or activity feeds
- Household-configurable voting rules (noted as a possible v2 enhancement; see §9)

## 3. User Stories

- As a household member, I want to add an item I'd like to buy someday so that I don't forget about it.
- As a household member, I want to record where the item can be purchased so I know where to look when I'm ready to buy it.
- As a household member, I want to mark an item with an urgency label (must have / need to have / want) so we can communicate how badly we want it.
- As a household member, I want to tap a vote button as many times as I like to express enthusiasm for an item, so the things we collectively care about float to the top.
- As a household member, I want the list sorted by vote count so I can see at a glance what we are most excited about.
- As a household member, I want to edit or delete any item on the list (not just my own) so the list stays accurate without coordination overhead.
- As a household member, I want the sidebar to group all shopping-related pages together so the navigation reflects how I think about these features.

## 4. Functional Requirements

### 4.1 Wish List Lifecycle

- Each household has exactly one implicit wish list. There is no "create list" action — the list is materialized on first read or first item insert.
- The wish list cannot be deleted, archived, or renamed. There is no list-level UI; only the items are surfaced.
- The wish list is scoped to `tenant_id` + `household_id`, consistent with other household-scoped features.

### 4.2 Wish List Items

Each item has the following fields:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | string | yes | Free text, up to 255 chars |
| `purchase_location` | string | no | Free text, up to 255 chars (e.g., "Amazon", "Costco", a store name, or a URL) |
| `urgency` | enum | yes | One of `must_have`, `need_to_have`, `want`. Defaults to `want` if not provided on create. |
| `vote_count` | int | system | Non-negative integer, defaults to 0 on create. Modified only via the vote endpoint. |
| `created_by` | UUID | system | User who created the item |
| `created_at` / `updated_at` | timestamp | system | |

Operations:

- **Create**: any household member can add an item. `vote_count` initializes to 0.
- **Edit**: any household member can edit `name`, `purchase_location`, or `urgency` of any item, regardless of who created it. `vote_count` cannot be modified via the edit endpoint.
- **Delete**: any household member can delete any item (hard delete).
- **Vote**: any household member can call the vote endpoint, which atomically increments `vote_count` by 1. There is no decrement, no per-user limit, and no per-user vote tracking. Votes are anonymous and uncapped.

### 4.3 Display & Sorting

- The default and only sort order is `vote_count DESC`, with a tie-breaker of `created_at ASC` (older items rank above newer items at the same vote count, so a newcomer doesn't leapfrog).
- Urgency is displayed as a label/badge but does **not** influence sort order.
- The list is rendered as a single flat list (no grouping by urgency or location).

### 4.4 Sidebar Reorganization

The frontend sidebar currently has the grocery list page directly under the "Lifestyle" group as `Shopping` (route `/app/shopping`). This reorganization:

- Adds a new top-level navigation group with key `shopping`, label **Shopping**.
- Moves the existing grocery list page out of "Lifestyle" and into the new "Shopping" group, renamed to **Grocery Lists**. The route stays at `/app/shopping` to avoid breaking bookmarks. (See §9 — open question on whether to relocate the route.)
- Adds a new entry **Wish List** under the "Shopping" group at route `/app/shopping/wish-list`.
- The "Lifestyle" group keeps Recipes, Meal Planner, Ingredients, and Weather.

Resulting "Shopping" group:

| Label | Route | Notes |
|-------|-------|-------|
| Grocery Lists | `/app/shopping` | Renamed from "Shopping" |
| Wish List | `/app/shopping/wish-list` | New |

## 5. API Surface

All endpoints live in **shopping-service** under a new resource path. Multi-tenancy and household scoping follow the same JWT-context conventions used elsewhere in shopping-service.

Base: `/api/v1/shopping/wish-list`

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/items`                  | List all wish-list items for the caller's household, sorted by `vote_count DESC, created_at ASC` |
| POST   | `/items`                  | Create a new wish-list item |
| PATCH  | `/items/{id}`             | Update an item's `name`, `purchase_location`, and/or `urgency` |
| DELETE | `/items/{id}`             | Delete an item |
| POST   | `/items/{id}/vote`        | Atomically increment the item's `vote_count` by 1 |

JSON:API type: `"wish-items"`

**Create request attributes:**
- `name` (string, required)
- `purchase_location` (string, optional)
- `urgency` (enum, optional, default `want`)

**Update request attributes:** any subset of `name`, `purchase_location`, `urgency`.

**Vote request:** no body required. The endpoint is a single-action POST that returns the updated item.

**Response attributes:** `name`, `purchase_location`, `urgency`, `vote_count`, `created_by`, `created_at`, `updated_at`.

Validation:
- `name`: non-empty after trim, ≤ 255 chars
- `purchase_location`: ≤ 255 chars when present
- `urgency`: must be one of the three enum values
- The PATCH endpoint MUST reject any attempt to set `vote_count`

Concurrency:
- The vote endpoint MUST use a SQL `UPDATE ... SET vote_count = vote_count + 1` (or equivalent atomic increment) rather than a read-modify-write, so concurrent taps from multiple devices do not lose votes.

## 6. Data Model

Schema: `shopping` (existing)

### `wish_list_items`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `tenant_id` | UUID | NOT NULL |
| `household_id` | UUID | NOT NULL |
| `name` | VARCHAR(255) | NOT NULL |
| `purchase_location` | VARCHAR(255) | NULL |
| `urgency` | VARCHAR(20) | NOT NULL, CHECK in (`must_have`, `need_to_have`, `want`), DEFAULT `want` |
| `vote_count` | INT | NOT NULL, DEFAULT 0, CHECK >= 0 |
| `created_by` | UUID | NOT NULL |
| `created_at` | TIMESTAMP | NOT NULL |
| `updated_at` | TIMESTAMP | NOT NULL |

Indexes:
- `(tenant_id, household_id, vote_count DESC, created_at ASC)` — primary listing index

Notes:
- No separate `wish_lists` parent table. The "list" is implicit per `(tenant_id, household_id)`.
- No `wish_item_votes` table in v1, since votes are anonymous and uncapped.

## 7. Service Impact

### 7.1 Modified: shopping-service

- New domain package `wishlist/` (model, entity, builder, processor, provider, resource, rest), following existing shopping-service domain conventions.
- New GORM entity for `wish_list_items` with tenant + household scoping callbacks.
- New migration adding the `wish_list_items` table to the `shopping` schema.
- New REST routes mounted under `/api/v1/shopping/wish-list`.
- Atomic vote-increment SQL on the vote endpoint.

### 7.2 Modified: frontend

- New page: `pages/shopping/WishListPage.tsx` (or equivalent under the existing shopping pages directory).
- New components: wish-list item row, urgency badge, vote button, create/edit dialog.
- New TanStack Query hooks: `useWishListItems`, `useCreateWishListItem`, `useUpdateWishListItem`, `useDeleteWishListItem`, `useVoteWishListItem`.
- New API client module under the existing shopping-service client area.
- Update `frontend/src/components/features/navigation/nav-config.ts`:
  - Remove the existing `Shopping` entry from the `lifestyle` group.
  - Add a new `shopping` group containing `Grocery Lists` (route `/app/shopping`, existing page) and `Wish List` (route `/app/shopping/wish-list`, new page).
- Update `App.tsx` route registration to add `/app/shopping/wish-list`.

### 7.3 Modified: nginx

- No new routing rules required; the new endpoints sit under the existing `/api/v1/shopping` prefix already routed to shopping-service.

### 7.4 Not modified

- category-service, recipe-service, and other services are untouched by this feature.

## 8. Non-Functional Requirements

- **Multi-tenancy:** Tenant scoping via existing JWT-context GORM callbacks. Household scoping enforced in the processor layer using the household ID resolved from the request context.
- **Authorization:** Any authenticated member of the household can perform any operation on any item in that household's wish list.
- **Concurrency:** The vote endpoint MUST be safe under concurrent calls. Use an atomic SQL increment, not a read-modify-write transaction.
- **Performance:** A wish list is expected to hold tens of items, not thousands. The listing index `(tenant_id, household_id, vote_count DESC, created_at ASC)` supports the only query pattern.
- **Mobile UI:** Per project mobile UI guidelines, the page must be tap-only (no swipe gestures), card-based on mobile, and the vote button must be large enough to tap repeatedly without misfires.
- **Observability:** Standard structured logging and tracing consistent with other shopping-service endpoints.

## 9. Open Questions

- **Household-configurable voting rules (v2):** The user noted that voting semantics may eventually need to vary per household (e.g., one-vote-per-user, capped votes). v1 ships with anonymous uncapped voting only. A future task would introduce a household-level wish-list configuration and possibly a `wish_item_votes` table for per-user attribution. Out of scope for v1.
- **Route relocation:** Should the existing grocery list route move from `/app/shopping` to `/app/shopping/grocery` for path consistency with the new sub-page? Current spec keeps `/app/shopping` to avoid breaking bookmarks. Decide before implementation if a redirect is preferred.
- **Empty state copy:** Wording for the empty wish list state is left to implementation.

## 10. Acceptance Criteria

- [ ] A new `wish_list_items` table exists in the `shopping` schema with the columns and constraints defined in §6.
- [ ] Shopping-service exposes the five endpoints in §5 under `/api/v1/shopping/wish-list`.
- [ ] All wish-list endpoints enforce tenant + household scoping from the JWT context.
- [ ] Creating an item without an `urgency` value defaults to `want`.
- [ ] Creating an item initializes `vote_count` to 0.
- [ ] PATCH on an item rejects any attempt to set `vote_count`.
- [ ] POST to `/items/{id}/vote` increments `vote_count` by exactly 1 atomically and returns the updated item.
- [ ] Concurrent vote requests from multiple clients result in the correct final count (no lost updates).
- [ ] Any household member can edit or delete any item on the household's wish list, regardless of `created_by`.
- [ ] GET `/items` returns items sorted by `vote_count DESC` with `created_at ASC` as tie-breaker.
- [ ] The frontend sidebar contains a new top-level "Shopping" group with two entries: "Grocery Lists" (existing page, route `/app/shopping`) and "Wish List" (new page, route `/app/shopping/wish-list`).
- [ ] The "Lifestyle" group no longer contains the "Shopping" entry.
- [ ] The wish list page displays items in vote-sorted order with name, purchase location, urgency badge, vote count, a tap-to-vote button, and edit/delete affordances.
- [ ] The vote button can be tapped repeatedly and each tap increments the count by 1, with optimistic UI feedback.
- [ ] An empty household wish list renders an empty state rather than an error.
