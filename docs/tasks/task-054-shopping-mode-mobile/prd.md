# Mobile Shopping Mode — "Big Tap Rows" — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-07-01

---

## 1. Overview

The grocery list detail page (`frontend/src/pages/ShoppingListDetailPage.tsx`) has a "shopping mode" that turns each list item into a checkable row. On desktop this works well. On mobile it does not: rows are `py-2` with `text-sm` text and a 20px (`h-5 w-5`) checkbox, so checking items off one-handed while pushing a cart means aiming at small targets and mis-tapping. The primary actions (Uncheck All, Finish Shopping) and the progress count live in a single wrapping button bar at the top of the page, which scrolls away as you work down a long list.

This feature makes shopping mode comfortable to use one-handed on a phone by adopting the **"Big Tap Rows"** direction chosen from an HTML/CSS prototype comparison. On mobile viewports, each shopping-mode item becomes a full-width ~64px tap target with a ~30px checkbox and larger text; the whole row remains tappable to toggle checked. A sticky progress header keeps "how much is left" visible while scrolling, and a sticky bottom action bar keeps the primary actions within thumb reach.

Every change is gated to mobile viewports (below Tailwind's `md` / 768px breakpoint). Desktop shopping mode and both mobile and desktop **edit** mode remain visually and behaviorally unchanged. There are no backend, API, data-model, or React Query changes — this is a presentational, responsive-layout change to a single page (plus, optionally, one small extracted row subcomponent).

Reference prototype (Option A of four): a standalone mock produced during brainstorming demonstrating the enlarged rows, sticky progress header, and sticky bottom bar.

## 2. Goals

Primary goals:
- On mobile, make each shopping-mode item an easy, forgiving one-handed tap target (≈64px row height, ≈30px checkbox, larger item text).
- Keep progress ("X of Y items") and the primary shopping actions reachable at all times on mobile without scrolling back up.
- Preserve every existing shopping-mode behavior: whole-row tap toggles checked; checked items dim, strike through, and sink to the bottom of their category; category grouping and sort order unchanged.
- Zero visual or behavioral change on desktop (`≥ md`).

Non-goals:
- Edit-mode layout changes (quick-add input, per-item trash targets) — left exactly as-is.
- The other prototyped directions: swipe-to-check gestures, one-item-at-a-time focus mode, aisle-by-aisle accordions.
- Any backend, API, JSON:API contract, data-model, or migration change.
- New React Query hooks or changes to mutation/query behavior.
- Custom Tailwind breakpoints — use the existing `md` default.
- Reordering or restructuring categories/aisles relative to the user's store.

## 3. User Stories

- As a shopper using my phone in a store, I want each item to be a large, full-width tap target so that I can check items off one-handed without mis-tapping.
- As a shopper working down a long list, I want the progress indicator to stay visible so that I always know how many items remain.
- As a shopper, I want "Finish Shopping" and "Uncheck All" to stay within thumb reach so that I don't have to scroll back to the top to use them.
- As a desktop user, I want shopping mode to look and behave exactly as it does today so that nothing I rely on regresses.

## 4. Functional Requirements

Requirements are grouped by capability area. "Mobile" means viewport width `< md` (768px); "desktop" means `≥ md`. All requirements apply to the **active** (non-archived) shopping-mode view unless stated otherwise.

### 4.1 Enlarged item rows (mobile, shopping mode)

- FR-1: On mobile, each shopping-mode item row MUST present a tap target of at least 56px height, targeting ~64px, spanning the full width of the list container.
- FR-2: On mobile, the checkbox indicator MUST be ~30px (e.g. `h-7 w-7`), with the check glyph scaled proportionally, replacing today's 20px checkbox.
- FR-3: On mobile, the item name MUST render at a larger size than today's `text-sm` (e.g. `text-base`/`text-[17px]`), remaining single-line-friendly with truncation/wrapping consistent with current behavior.
- FR-4: On mobile, tapping anywhere on the row MUST toggle the item's checked state (same mutation as today: `useCheckShoppingItem` with `{ itemId, checked: !item.checked }`). No separate hit area is required for the checkbox.
- FR-5: On desktop, the row layout MUST remain identical to the current implementation (`py-2`, `text-sm`, 20px checkbox). The enlargements MUST be applied via responsive utilities (base = mobile sizing, `md:` = current desktop sizing) or an equivalent viewport-gated mechanism.

### 4.2 Preserved shopping-mode behaviors

- FR-6: Checked items MUST continue to dim (current `opacity-60`) and show `line-through` on the name in shopping mode.
- FR-7: Within each category, unchecked items MUST continue to display before checked items in shopping mode (current `[...unchecked, ...checked]` ordering).
- FR-8: Category grouping and category sort order MUST be unchanged (same `groupItemsByCategory` logic and card grouping).
- FR-9: The quantity badge MUST continue to render when present; on mobile it MAY be repositioned (e.g. right-aligned) for readability but MUST remain visible.

### 4.3 Sticky progress header (mobile, shopping mode)

- FR-10: On mobile in shopping mode, a progress indicator MUST be sticky to the top of the scroll area so it remains visible while the item list scrolls.
- FR-11: The progress indicator MUST show a progress bar and the "X of Y items" count derived from the existing `checkedCount` / `totalCount` values.
- FR-12: The sticky header MUST include a way to exit shopping mode ("Back to Edit"), consistent with the confirmed design (back affordance lives in the top area).
- FR-13: On desktop, the progress count MUST remain in its current location/format (the top action bar); the sticky mobile header MUST NOT appear on desktop.

### 4.4 Sticky bottom action bar (mobile, shopping mode)

- FR-14: On mobile in shopping mode, a sticky bottom bar MUST keep the primary actions reachable: "Finish Shopping" (primary) and "Uncheck All".
- FR-15: The bottom bar MUST use the established in-container `sticky bottom-0` pattern (as in `bulk-categorize.tsx`), sitting at the bottom of the `<main>` scroll container — NOT `position: fixed`.
- FR-16: The bottom bar buttons MUST retain their current actions/handlers: "Finish Shopping" opens the existing finish-confirmation dialog; "Uncheck All" calls `uncheckAll.mutate()`.
- FR-17: The bottom bar MUST NOT overlap the last list item's content — the scroll area MUST allow the final item to be fully visible above the bar (e.g. bottom padding on the list).
- FR-18: On desktop, no sticky bottom bar appears; the top action bar remains as today.

### 4.5 Edit mode and archived view

- FR-19: Edit mode (non-shopping, non-archived) MUST be completely unchanged on all viewports: quick-add input, per-item trash buttons, item ordering, and row sizing stay as today.
- FR-20: The archived read-only view (which also renders checkboxes) MAY inherit the same enlarged mobile row sizing for visual consistency, since it shares the render path. It MUST remain read-only (no toggling, no sticky action bar with mutating actions). Progress/bottom-bar chrome specific to active shopping is NOT required in the archived view.

### 4.6 States

- FR-21: The empty-list state ("No items yet") MUST be unchanged.
- FR-22: The loading skeleton and "list not found" states MUST be unchanged.
- FR-23: When `totalCount` is 0, the sticky progress header MUST degrade gracefully (no divide-by-zero in the progress bar; e.g. 0% width).

## 5. API Surface

No API changes. This feature is presentational only. Existing hooks are reused unchanged:

- `useShoppingList(id)` — read list + nested items.
- `useCheckShoppingItem(id)` — toggle an item's `checked` state.
- `useUncheckAllItems(id)` — bulk uncheck.
- `useArchiveShoppingList()` — invoked by "Finish Shopping" (via confirm dialog).

No new endpoints, request/response shapes, or error cases.

## 6. Data Model

No data-model changes. The existing `NestedShoppingItem` shape is consumed as-is:

- `name: string`
- `quantity: string | null`
- `category_name: string | null`, `category_sort_order: number | null`
- `checked: boolean`
- `position: number`

No new entities, fields, relationships, constraints, or migrations. No `tenant_id` considerations beyond what the existing hooks already enforce.

## 7. Service Impact

- **frontend** — the only affected service.
  - `frontend/src/pages/ShoppingListDetailPage.tsx` — apply responsive row sizing in the shopping-mode item render; add the mobile-only sticky progress header and sticky bottom action bar; gate the current top action bar / progress count to desktop where appropriate.
  - Optional: extract the item row into a small subcomponent (e.g. `ShoppingItemRow`) if it improves clarity/testability; not required by scope.
  - Possible new/updated unit tests under `frontend/src/pages/__tests__/` (or a colocated `__tests__` folder) for the mobile layout logic (e.g. progress rendering, empty/zero-count handling).
- **No Go services are affected.** No shared library changes, no Docker build impact for backend services.

## 8. Non-Functional Requirements

- **Responsiveness:** All new mobile chrome MUST be gated below `md` and MUST NOT affect desktop rendering. Verify at a representative mobile width (e.g. 375–414px) and at `≥ md`.
- **Performance:** No additional network requests, no added re-render cost beyond trivial layout. Item list rendering complexity is unchanged (same grouping/sorting).
- **Accessibility / touch:** Tap targets meet a ~44px minimum (the ~64px rows exceed this comfortably). The whole-row tap affordance MUST remain keyboard/pointer accessible as today; do not remove existing semantics.
- **Multi-tenancy:** Unaffected — no data access changes; tenancy remains enforced by existing hooks/services.
- **Observability:** No new logging/metrics required.
- **Testing:** Follow frontend-dev-guidelines (FE-*). Any date/timezone-sensitive tests (none expected here) run under `TZ=UTC`. Run the frontend test suite and lint before completion.
- **No regressions:** Desktop shopping mode, and edit mode on all viewports, MUST be byte-for-byte equivalent in behavior to `main`.

## 9. Open Questions

None blocking. Confirmed decisions from the brainstorming interview:

- Mobile-only, responsive via the existing `md` breakpoint. (Confirmed)
- Include both the sticky progress header and the sticky bottom action bar. (Confirmed)
- "Back to Edit" moves into the sticky top header as a back affordance; bottom bar holds Uncheck All + Finish Shopping. (Confirmed default)
- Archived read-only view inherits the enlarged mobile row sizing for consistency, staying read-only. (Confirmed default)

Deferred to the design phase (non-blocking): exact quantity-badge placement on mobile; whether to extract a `ShoppingItemRow` subcomponent; precise sticky-header visual treatment (background/blur/border).

## 10. Acceptance Criteria

- [ ] On a mobile viewport (`< md`) in shopping mode, each item row is a full-width tap target of ≥56px (≈64px) with a ~30px checkbox and larger-than-`text-sm` name text.
- [ ] Tapping anywhere on a mobile row toggles its checked state via the existing check mutation.
- [ ] Checked items still dim, strike through, and sort after unchecked items within their category; category grouping/sort is unchanged.
- [ ] On mobile in shopping mode, a sticky progress header (bar + "X of Y items") stays visible while the list scrolls, and includes a "Back to Edit" affordance.
- [ ] On mobile in shopping mode, a sticky bottom bar (using in-container `sticky bottom-0`, not `fixed`) keeps "Finish Shopping" (opens the existing confirm dialog) and "Uncheck All" reachable; the last item is not obscured by the bar.
- [ ] On desktop (`≥ md`), shopping mode renders and behaves exactly as on `main` (same rows, same top action bar, same progress location) — no sticky mobile chrome appears.
- [ ] Edit mode (quick-add, trash targets, ordering, sizing) is unchanged on all viewports.
- [ ] The archived read-only view remains read-only; if it adopts the enlarged mobile rows, it does so without mutating controls.
- [ ] Empty-list, loading, "not found", and zero-count progress states render without errors.
- [ ] `frontend` lint and unit tests pass; no backend services are touched.
