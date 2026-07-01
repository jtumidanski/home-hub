# Mobile Shopping Mode — "Big Tap Rows" — Design

Task: task-054-shopping-mode-mobile
Status: Approved for planning
Created: 2026-07-01
Companion PRD: `docs/tasks/task-054-shopping-mode-mobile/prd.md`
Confirmed direction: Option A ("Big Tap Rows") from `prototype-options.html`

---

## 1. Summary

This is a presentational, responsive-layout change confined to a single React page,
`frontend/src/pages/ShoppingListDetailPage.tsx`. On mobile viewports (`< md`, 768px) the
shopping-mode item list becomes a set of ~64px full-width tap rows with a ~30px checkbox and
larger text, a sticky top progress header (bar + "X of Y" + "Back to Edit"), and a sticky bottom
action bar ("Uncheck All" + "Finish Shopping"). Desktop (`≥ md`) shopping mode and edit mode on
all viewports are unchanged.

There are **no** backend, API, JSON:API, data-model, migration, or React Query changes. All
existing hooks (`useShoppingList`, `useCheckShoppingItem`, `useUncheckAllItems`,
`useArchiveShoppingList`) are reused verbatim, with the same call signatures.

The entire feature is achieved with **pure responsive Tailwind utilities** — base classes = mobile
sizing, `md:` classes = today's desktop sizing — plus two new mobile-only (`md:hidden`) chrome
blocks. There is **no** JavaScript viewport detection, no `matchMedia`, no resize listeners, no new
custom breakpoint. This is the central architectural decision (see §3).

## 2. Current-State Grounding

Everything below is verified against the worktree source, not assumed.

**Scroll container.** `AppShell` (`src/components/features/navigation/app-shell.tsx:62`) renders the
routed page inside `<main className="flex-1 overflow-auto">`. On mobile the layout is a flex
**column**: `MobileHeader` (`h-14`, `md:hidden`) sits above `<main>`, and `<main>` is its own
independent scroll area. This is the scroll ancestor that CSS `sticky` resolves against. A
`sticky top-0` child of the page pins flush to the top of `<main>` (just under the mobile header);
a `sticky bottom-0` child pins to the bottom of `<main>`. No `position: fixed` is needed or wanted.

**Established sticky pattern.** `bulk-categorize.tsx:170` is the reference:
`className="sticky bottom-0 flex items-center gap-2 rounded-md border bg-background p-3"`. It is an
**in-flow** sticky element (opaque `bg-background`, a `border`) that reserves its own layout slot at
the end of the scroll content. Because sticky (unlike fixed) participates in normal flow, the
element occupies space at the bottom of the document and therefore **cannot overlap** the last item
above it. We mirror this pattern exactly for the bottom action bar — which is why FR-17 ("last item
not obscured") needs no manual bottom padding.

**Today's row markup** (`ShoppingListDetailPage.tsx:252-302`), per category card:
- Row container: `flex items-center gap-3 px-2 py-2 rounded hover:bg-accent/30 transition-colors`,
  plus `cursor-pointer` in shopping mode and `opacity-60` when checked. `onClick` fires
  `checkItem.mutate({ itemId, checked: !item.checked })` only when `shoppingMode && !isArchived`.
- Checkbox (rendered when `shoppingMode || isArchived`): `h-5 w-5 rounded border ...`, glyph
  `<Check className="h-3 w-3" />`.
- Name + quantity live inside `<div className="flex-1 min-w-0">`: the name `<span>` is
  `text-sm` (with `line-through` when checked in shopping mode), and the quantity `<Badge>` is
  inline right after the name with `ml-2 text-xs`.
- Trash button renders only when `!shoppingMode && !isArchived` (edit mode).

**Today's top action bar** (`ShoppingListDetailPage.tsx:163-191`), shopping-mode branch: a wrapping
`flex gap-2 flex-wrap` row containing "Back to Edit", "Uncheck All", "Finish Shopping", and the
`{checkedCount} of {totalCount} items` count (`ml-auto`).

**Derived counts** already exist: `checkedCount` and `totalCount` (`ShoppingListDetailPage.tsx:94-95`).

**Test conventions.** Vitest + jsdom; page tests live in `src/pages/__tests__/*.test.tsx` and are
render-based (e.g. `TasksPage.test.tsx`, `RemindersPage.test.tsx`). No existing shopping-list test.

## 3. Approaches Considered

**A. Pure responsive Tailwind (recommended, chosen).** Express both layouts in one JSX tree: base
utility = mobile, `md:` variant = current desktop. New mobile chrome is gated with `md:hidden`; the
desktop shopping-mode action bar is gated `hidden md:flex`. No JS runs to decide layout.

- *Pros:* Zero runtime cost; no hydration/SSR/first-paint flicker; desktop output is provably
  unchanged because every desktop-affecting class is reasserted at `md:`; matches how the rest of
  the app already does responsiveness (`md:hidden`, `md:flex-row`, `p-4 md:p-6`). Diff is minimal
  and reviewable class-by-class.
- *Cons:* Both variants coexist in one JSX block; a few `class` strings get longer. Two mobile-only
  chrome blocks are rendered on desktop but hidden via CSS (negligible DOM cost).

**B. JS viewport hook (`useMediaQuery`) branching two render paths.** Detect `< md` in JS and render
a distinct mobile tree.

- *Pros:* Cleaner separation of the two layouts in source.
- *Cons:* Introduces a hook the codebase doesn't currently need; first-render/SSR flicker risk;
  duplicates the item-render logic (drift risk between mobile/desktop behavior — directly threatens
  the "byte-for-byte desktop" requirement); more code to test. Rejected.

**C. Full `ShoppingItemRow` + chrome subcomponent extraction.** Extract the row, the progress
header, and the bottom bar into separate components.

- *Pros:* Smaller units, independently testable.
- *Cons:* The row is deeply coupled to page-local state/handlers (`shoppingMode`, `isArchived`,
  `checkItem`, `removeItem`) — extraction means threading a wide prop interface for a single
  consumer, which is churn without payoff (YAGNI). Rejected as the default; see §5 for the one small
  extraction we *do* make (a pure helper).

**Decision:** Approach A, with a single pure helper extracted for the progress math (testability),
and no component extraction. This is the lowest-risk path to "desktop unchanged" and keeps the whole
change in one file plus one small helper.

## 4. Architecture

All edits are in `ShoppingListDetailPage.tsx`. Four coordinated changes, each independently
verifiable:

### 4.1 Responsive row sizing (in-place, both shopping mode and archived)

Applied to the existing row markup so mobile enlarges while `md:` restores today's exact values.

| Element | Base (mobile) | `md:` (desktop, = today) |
|---|---|---|
| Row container height | `min-h-[64px]` | `md:min-h-0` |
| Row vertical padding | `py-2` (unchanged) | `py-2` |
| Checkbox box | `h-7 w-7` | `md:h-5 md:w-5` |
| Check glyph | `h-4 w-4` | `md:h-3 md:w-3` |
| Item name text | `text-[17px]` | `md:text-sm` |

Rationale: `py-2` already reads fine on desktop; the mobile height is driven by `min-h-[64px]`
(exceeds the 56px floor and the ~44px a11y minimum) rather than by changing padding, which keeps the
desktop box model identical. Every enlarged property is explicitly reset at `md:`, so desktop
rendering is unchanged.

**Quantity badge placement — resolved deferred question.** Keep the badge **inline** after the name
(exactly as today) on both viewports. FR-9 *permits* right-aligning it on mobile but does not require
it; keeping it inline (a) guarantees zero desktop regression, (b) avoids restructuring the row's DOM
(which would risk desktop layout), and (c) is perfectly legible inside a 64px row. Right-alignment is
explicitly declined as unnecessary DOM churn. The badge stays `ml-2 text-xs` and remains visible in
all cases.

The archived read-only view shares this render path and therefore inherits the enlarged mobile rows
(FR-20). It stays read-only: `onClick` already no-ops when `isArchived`, and no sticky mutating chrome
is rendered for it (§4.3/§4.4 are gated on `shoppingMode && !isArchived`).

### 4.2 Desktop-only top action bar (gate existing chrome)

The shopping-mode branch of the top action bar (Back to Edit / Uncheck All / Finish Shopping / count)
is wrapped so it shows only at `md+`: `hidden md:flex`. On mobile these move into the sticky chrome
below. The **non**-shopping-mode branch (Start Shopping / Import from Meal Plan) and the archived
action bar are **untouched** and render on all viewports as today. Edit-mode quick-add is untouched.

### 4.3 Sticky mobile progress header (new, `md:hidden`, shopping mode only)

Rendered only when `!isArchived && shoppingMode`, as a child of the page root so it sticks against
`<main>`:

```
sticky top-0 z-20 md:hidden -mx-4 px-4 py-3 bg-background border-b
  ├─ progress bar:  <div class="h-2 rounded bg-muted"><div style={{width: `${pct}%`}} class="h-2 rounded bg-primary"/></div>
  ├─ "X of Y items" text (from checkedCount / totalCount)
  └─ "Back to Edit" button (ArrowLeft) → setShoppingMode(false)
```

- `bg-background` makes it opaque so category cards scroll cleanly beneath it; `border-b` separates it.
- `-mx-4 px-4` lets the bar span the page's horizontal padding (full-bleed) while keeping inner
  content aligned to the page gutter.
- `z-20` keeps it above card content (cards have no competing sticky/z context).
- `pct` is computed via the pure helper (§5) so the zero-count case (FR-23) is `0%`, never `NaN`.
- Placed after the page title header and before the category cards, so on scroll the title scrolls
  away and this header pins to the top of `<main>`.

### 4.4 Sticky mobile bottom action bar (new, `md:hidden`, shopping mode only)

Rendered only when `!isArchived && shoppingMode`, as the last in-flow child of the page root before
the dialogs, mirroring `bulk-categorize.tsx`:

```
sticky bottom-0 z-20 md:hidden -mx-4 px-4 py-3 bg-background border-t
  ├─ "Uncheck All" (variant outline) → uncheckAll.mutate()
  └─ "Finish Shopping" (primary, flex-1) → setShowFinishConfirm(true)  // opens existing dialog
```

Handlers are the exact ones used by today's desktop buttons — no new state, no new mutation. Because
the bar is `sticky` (in-flow), it reserves its own slot at the end of the scroll content, so the last
category card is never obscured (FR-17) with no extra padding. The existing Finish-confirmation
`Dialog` and its `handleFinishShopping` are reused unchanged.

### 4.5 Rendering / data flow (unchanged)

`groupItemsByCategory` grouping, the `shoppingMode ? [...unchecked, ...checked] : group.items`
per-category ordering, `opacity-60` dimming, and `line-through` on checked names are all preserved
verbatim (FR-6/7/8). The only behavioral surface is the same `checkItem.mutate` toggle on row tap
(FR-4). No memoization or query behavior changes.

## 5. The one extraction: `progressPercent` helper

Extract a single pure, exported, module-level function for the progress-bar width:

```ts
export function progressPercent(checked: number, total: number): number {
  if (total <= 0) return 0;
  return Math.min(100, Math.max(0, (checked / total) * 100));
}
```

Purpose: make the zero-count guard (FR-23) and the ratio trivially unit-testable without rendering,
and give the sticky header a single source of truth for its width. This is the *only* new exported
symbol. No component extraction (Approach C is declined).

## 6. Responsive gating matrix (single source of truth)

| Element | Mobile `< md` | Desktop `≥ md` |
|---|---|---|
| Item row (shopping/archived) | ~64px, 30px checkbox, `text-[17px]` | today's `py-2`, 20px checkbox, `text-sm` |
| Quantity badge | inline, visible | inline, visible (unchanged) |
| Top action bar (shopping-mode branch) | hidden | shown (unchanged) |
| Top action bar (Start Shopping / Import) | shown (unchanged) | shown (unchanged) |
| Sticky progress header | shown (shopping, active) | hidden |
| Sticky bottom bar | shown (shopping, active) | hidden |
| Edit mode (quick-add, trash, sizing) | unchanged | unchanged |
| Archived action bar (Reopen/Delete) | unchanged | unchanged |

"Active" = `!isArchived && shoppingMode`.

## 7. Edge cases & states

- **Zero items / totalCount 0:** progress bar width `0%` via `progressPercent` (no `NaN`, FR-23).
  The empty-list card ("No items yet") is unchanged (FR-21). Note: the sticky chrome is gated on
  `shoppingMode`; Start Shopping is only reachable with items present, but the helper makes the header
  safe regardless.
- **Loading / not-found:** early-return branches (`isLoading`, `!list`) are untouched (FR-22).
- **Archived view:** enlarged mobile rows inherited; read-only preserved; no sticky mutating chrome
  (FR-20).
- **Desktop:** every mobile class is reset at `md:`, and new chrome is `md:hidden` → desktop output
  is unchanged (FR-5/13/18, no-regression clause).
- **Sticky-under-header sliver:** the progress header pins at `top-0` of `<main>` with an opaque
  `bg-background`, so content scrolling beneath is fully covered; the page's `p-4` top gutter is
  above the pin point and shows nothing once pinned.

## 8. Testing strategy

Add `src/pages/__tests__/ShoppingListDetailPage.test.tsx` (render-based, matching existing page-test
style) plus direct helper tests:

1. **`progressPercent` (pure):** `(0, 0) → 0`, `(0, 4) → 0`, `(2, 4) → 50`, `(4, 4) → 100`,
   `(5, 4) → 100` (clamped). No render needed; fast and deterministic.
2. **Mobile chrome presence (render):** with a mocked `useShoppingList` returning a small item set and
   in shopping mode, assert the sticky progress header ("X of Y items" text + "Back to Edit") and the
   sticky bottom bar ("Uncheck All", "Finish Shopping") are in the DOM. (Presence, not viewport
   media evaluation — jsdom doesn't do CSS media queries; the `md:hidden` gating is a
   compile-time class assertion, verified by class inspection if needed.)
3. **Row tap toggles:** clicking a shopping-mode row calls the mocked `useCheckShoppingItem` mutate
   with `{ itemId, checked: !checked }` (guards FR-4 during refactor).
4. **Archived read-only:** in archived mode, clicking a row does **not** call the check mutation.

No date/timezone-sensitive logic is introduced, so the `TZ=UTC` concern does not apply here; still,
run the suite and lint before completion per project rules. jsdom cannot evaluate Tailwind media
queries, so mobile-vs-desktop *visual* gating is verified by class presence and by manual check at
375–414px and `≥ md`, not by asserting computed layout in tests.

## 9. Risks & mitigations

- **Desktop regression (highest-priority risk).** Mitigation: every mobile utility has an explicit
  `md:` reset to today's value; new chrome is `md:hidden`; desktop action bar gated `hidden md:flex`.
  Reviewer can diff class-by-class. Manual desktop check at `≥ md`.
- **Sticky positioning vs. an unexpected intermediate scroll/overflow ancestor.** Verified: the page
  root under `<main className="overflow-auto">` has no intervening `overflow`/`transform` ancestor
  that would re-scope sticky. Mitigation: manual scroll test on device-width.
- **Row DOM restructuring breaking desktop.** Mitigation: we deliberately do **not** restructure the
  row (badge stays inline) — only add responsive size classes.

## 10. Out of scope (from PRD, reaffirmed)

Edit-mode layout; swipe/focus/accordion prototype directions; any backend/API/JSON:API/data-model/
migration change; new React Query hooks; custom Tailwind breakpoints; aisle reordering. Component
extraction beyond the single `progressPercent` helper.

## 11. Traceability (FR → design)

FR-1/2/3/5 → §4.1 row matrix. FR-4 → §4.5 tap toggle. FR-6/7/8 → §4.5 preserved. FR-9 → §4.1 badge
inline (deferred question resolved). FR-10/11/12/13 → §4.3 sticky header (+§4.2 desktop count).
FR-14/15/16/17/18 → §4.4 sticky bottom bar (in-flow sticky, reused handlers). FR-19 → untouched edit
mode. FR-20 → §4.1/§4.3/§4.4 archived read-only. FR-21/22/23 → §7 states + §5 helper.
