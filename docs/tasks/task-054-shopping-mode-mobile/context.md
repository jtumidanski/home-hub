# Task 054 — Mobile Shopping Mode — Context

Companion to `plan.md`. Everything an implementer needs to orient without re-reading the whole codebase. All paths are relative to the worktree root: `/home/tumidanski/source/home-hub/.worktrees/task-054-shopping-mode-mobile`.

## What this is

A presentational, responsive-layout change to a **single** React page so grocery "shopping mode" is comfortable one-handed on a phone. No backend, API, JSON:API, data-model, migration, or React Query change. Confirmed direction: Option A ("Big Tap Rows") from `prototype-options.html`.

## Key files

| File | Role |
|---|---|
| `frontend/src/pages/ShoppingListDetailPage.tsx` | **The only production file changed.** Holds all shopping-list detail UI: header, action bar, quick-add, category cards, item rows, finish/import dialogs. |
| `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` | **New.** Vitest + jsdom render tests + `progressPercent` unit tests. |
| `frontend/src/lib/hooks/api/use-shopping.ts` | Existing hooks — reused verbatim. See signatures below. |
| `frontend/src/lib/hooks/api/use-meals.ts` | `usePlans()` — used by the Import dialog; mocked in tests. |
| `frontend/src/components/features/navigation/app-shell.tsx` | `<main className="flex-1 overflow-auto">` at line 62 is the **scroll container** `sticky` resolves against. |
| `frontend/src/components/features/tasks/bulk-categorize.tsx` | Line 170 is the reference in-flow `sticky bottom-0` pattern the bottom bar mirrors. |
| `frontend/src/pages/__tests__/RemindersPage.test.tsx` | Reference for the page-test mocking style (module mocks + `MemoryRouter`). |

## Verified hook signatures (do not change)

From `use-shopping.ts`:

- `useShoppingList(id: string | null)` → `{ data: { data: { id, attributes: { name, status, archived_at, items } } }, isLoading }`. Page reads `listData?.data`, `list.attributes.status` (`"archived"` ⇒ archived view), `listData?.data?.attributes?.items ?? []`.
- `useCheckShoppingItem(listId)` → `{ mutate }`; `mutate({ itemId, checked })`.
- `useUncheckAllItems(listId)` → `{ mutate }`; `mutate()`.
- `useArchiveShoppingList()` → `{ mutate, isPending }`; invoked by "Finish Shopping" via the confirm dialog (`handleFinishShopping`).
- Also present/reused: `useAddShoppingItem`, `useRemoveShoppingItem`, `useUnarchiveShoppingList`, `useDeleteShoppingList`, `useImportMealPlan`.

`NestedShoppingItem` (from `@/types/models/shopping`): `id`, `name: string`, `quantity: string | null`, `category_name: string | null`, `category_sort_order: number | null`, `checked: boolean`, `position: number`.

## Current-state anchors (line numbers as of planning)

- `progressPercent` insertion point: after imports, before `interface GroupedItems` (~line 40).
- Derived counts `checkedCount` / `totalCount`: lines 94–95.
- Shopping-mode top action bar wrapper: line ~164 (`<div className="flex gap-2 flex-wrap">`), inner `shoppingMode ? (…) : (…)` branches; the shopping branch holds Back to Edit / Uncheck All / Finish Shopping / count.
- Items block start (`{/* Items grouped by category */}`): line ~229 → progress header goes just before it.
- Per-item row markup (`displayItems.map`): lines ~252–302 → responsive class edits here (row container, checkbox `h-5 w-5`, glyph `h-3 w-3`, name `text-sm`).
- Items block closes (~line 307) → bottom bar goes right after, before `{/* Finish Shopping Confirmation */}` (line ~309).

Line numbers may drift as edits land; anchor by the marker comments and class strings, not raw numbers.

## Central design decisions

1. **Pure responsive Tailwind, no JS viewport detection.** Base class = mobile; `md:` = today's desktop. New chrome is `md:hidden`; desktop shopping-mode top bar is `hidden md:flex`. This is the lowest-risk path to "desktop byte-for-byte unchanged" (every mobile utility is explicitly reset at `md:`). Rejected: a `useMediaQuery` hook (flicker + logic duplication) and full row/chrome component extraction (wide prop interface for a single consumer — YAGNI).
2. **One extraction only:** the pure `progressPercent(checked, total)` helper — the sole new exported symbol — so the zero-count guard (FR-23) is unit-testable without rendering.
3. **Quantity badge stays inline** after the name on both viewports (FR-9 permits but does not require right-aligning). Avoids row DOM restructuring that could regress desktop.
4. **Sticky, in-flow (not fixed).** Because `sticky` participates in normal flow, the bottom bar reserves its own slot at the end of scroll content and cannot overlap the last item (FR-17) — no manual bottom padding needed. Mirrors `bulk-categorize.tsx:170`.
5. **Archived view** shares the row render path, so it inherits enlarged mobile rows (FR-20) but renders **no** mutating sticky chrome (both sticky blocks gate on `!isArchived && shoppingMode`).

## Testing notes (important)

- **jsdom does not evaluate CSS media queries.** Both the mobile chrome (`md:hidden`) and the desktop top bar (`hidden md:flex`) render into the DOM simultaneously in tests, so text like "Back to Edit", "Uncheck All", "Finish Shopping", and the count appear **twice**. Disambiguate with `data-testid` on the mobile chrome (`mobile-shopping-progress`, `mobile-shopping-actions`) and scope assertions with `within(...)`. `data-testid` + `getAllBy*`/`within` are established patterns in this repo.
- Entering shopping mode in a test: render, then `await userEvent.click(screen.getByText("Start Shopping"))` (the only un-duplicated entry point; `shoppingMode` defaults to `false`).
- `useParams`/`useNavigate` are mocked via a partial `react-router-dom` mock (`importActual` + override) since the page reads the `:id` route param.
- The finish dialog title is **"Finish Shopping?"** (with `?`), distinct from the "Finish Shopping" button — safe to assert directly.
- Test framework: Vitest (`npx vitest run`), lint: `npm run lint`. No `TZ=UTC` needed (no date/timezone logic introduced).

## Environment gotcha

`node`/`npx`/`npm` are **not** on the default shell PATH. Prefix every frontend command with:

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
```

Run from `frontend/` inside the worktree.

## Dependencies / ordering

Tasks are strictly sequential: Task 1 (`progressPercent`) → Task 2 (rows + render harness) → Task 3 (progress header, consumes helper + harness) → Task 4 (bottom bar, consumes harness). The render harness (`renderPage`, `mockList`, `mockCheckMutate`, `mockUncheckMutate`) is created in Task 2 and reused by Tasks 3–4. Task 3 depends on Task 1's `progressPercent`.

## Out of scope (reaffirmed)

Edit-mode layout; swipe/focus/accordion prototype directions; any backend/API/JSON:API/data-model/migration change; new React Query hooks; custom Tailwind breakpoints; aisle reordering; component extraction beyond `progressPercent`.
