# Mobile Shopping Mode — "Big Tap Rows" Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make grocery shopping mode comfortable one-handed on a phone — enlarged tap rows, a sticky progress header, and a sticky bottom action bar — with zero change to desktop or edit mode.

**Architecture:** A single-file, presentational, responsive change to `frontend/src/pages/ShoppingListDetailPage.tsx`. Both layouts live in one JSX tree: base Tailwind utilities = mobile sizing, `md:` variants = today's desktop sizing. Two new mobile-only (`md:hidden`) sticky chrome blocks are added and gated on `!isArchived && shoppingMode`; the desktop shopping-mode top action bar is gated `hidden md:flex`. One pure exported helper (`progressPercent`) is added for the progress-bar math. No backend, API, JSON:API, data-model, migration, or React Query changes; all existing hooks are reused verbatim.

**Tech Stack:** React + TypeScript, Vite, Tailwind CSS, shadcn/ui, lucide-react, Vitest + jsdom + @testing-library/react + userEvent.

## Global Constraints

- **Single file of production change:** all production edits are in `frontend/src/pages/ShoppingListDetailPage.tsx`. No other production files change.
- **No API/data/query changes:** reuse `useShoppingList`, `useCheckShoppingItem`, `useUncheckAllItems`, `useArchiveShoppingList` (and existing edit-mode hooks) with identical call signatures. No new hooks, endpoints, or mutations.
- **Desktop byte-for-byte unchanged (`≥ md`, 768px):** every mobile utility MUST be reset at `md:` to today's exact value; every new chrome block MUST be `md:hidden`; the desktop shopping-mode action bar MUST be gated `hidden md:flex`.
- **Edit mode unchanged on all viewports:** quick-add input, per-item trash buttons, ordering, and row sizing stay as today.
- **Mobile breakpoint = Tailwind default `md` (768px).** No custom breakpoints. "Mobile" = `< md`; "desktop" = `≥ md`.
- **Sticky, not fixed:** mobile chrome uses in-flow `sticky` inside the `<main className="overflow-auto">` scroll container (mirroring `bulk-categorize.tsx:170`), never `position: fixed`.
- **"Active" gate** for new chrome = `!isArchived && shoppingMode`.
- **Zero-count safety:** the progress bar MUST render `0%` (never `NaN`) when `totalCount === 0`.
- **Test/lint gate:** `frontend` lint and unit tests MUST pass before completion. No date/timezone-sensitive logic is introduced, so `TZ=UTC` is not required here; still run the suite and lint.

### Node/npm invocation (this machine)

`node`/`npx` are not on the default shell PATH. **Every command below is prefixed** with:

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
```

All commands run from `frontend/` inside the worktree:
`/home/tumidanski/source/home-hub/.worktrees/task-054-shopping-mode-mobile/frontend`.

---

## File Structure

- **Modify:** `frontend/src/pages/ShoppingListDetailPage.tsx`
  - Add exported pure helper `progressPercent(checked, total)` at module level.
  - Enlarge shopping/archived item rows on mobile; reset at `md:`.
  - Gate the shopping-mode top action bar to `hidden md:flex`.
  - Add sticky mobile progress header (`data-testid="mobile-shopping-progress"`, `md:hidden`).
  - Add sticky mobile bottom action bar (`data-testid="mobile-shopping-actions"`, `md:hidden`).
- **Create:** `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx`
  - Pure `progressPercent` unit tests (Task 1).
  - Render harness + row-behavior guard tests (Task 2).
  - Sticky progress-header render tests (Task 3).
  - Sticky bottom-bar render tests (Task 4).

---

## Task 1: `progressPercent` pure helper

Extract the only new exported symbol: a pure, module-level function that drives the progress-bar width and guarantees the zero-count case is `0%`, not `NaN`. Testable without rendering.

**Files:**
- Modify: `frontend/src/pages/ShoppingListDetailPage.tsx` (add exported function near the top, after imports / before `groupItemsByCategory`)
- Test: `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` (create)

**Interfaces:**
- Consumes: nothing.
- Produces: `export function progressPercent(checked: number, total: number): number` — returns a clamped percentage in `[0, 100]`; returns `0` when `total <= 0`. Consumed by the sticky progress header in Task 3.

- [ ] **Step 1: Write the failing test**

Create `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx`:

```tsx
import { describe, it, expect } from "vitest";
import { progressPercent } from "../ShoppingListDetailPage";

describe("progressPercent", () => {
  it("returns 0 when total is 0 (no divide-by-zero)", () => {
    expect(progressPercent(0, 0)).toBe(0);
  });
  it("returns 0 when nothing is checked", () => {
    expect(progressPercent(0, 4)).toBe(0);
  });
  it("returns 50 at the halfway point", () => {
    expect(progressPercent(2, 4)).toBe(50);
  });
  it("returns 100 when everything is checked", () => {
    expect(progressPercent(4, 4)).toBe(100);
  });
  it("clamps to 100 when checked exceeds total", () => {
    expect(progressPercent(5, 4)).toBe(100);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: FAIL — `progressPercent` is not exported from `../ShoppingListDetailPage` (import resolves to `undefined`, tests error/fail).

- [ ] **Step 3: Write minimal implementation**

In `frontend/src/pages/ShoppingListDetailPage.tsx`, add this exported function immediately after the imports block and before `interface GroupedItems` (line ~40):

```tsx
export function progressPercent(checked: number, total: number): number {
  if (total <= 0) return 0;
  return Math.min(100, Math.max(0, (checked / total) * 100));
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: PASS — all 5 `progressPercent` cases green.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/ShoppingListDetailPage.tsx frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx
git commit -m "feat(shopping): add progressPercent helper for mobile shopping mode"
```

---

## Task 2: Enlarged mobile rows + render harness + row-behavior guard tests

Enlarge the shopping/archived item rows on mobile (reset at `md:`), and stand up the render-test harness that Tasks 3–4 reuse. The row-behavior guard tests characterize the existing tap-toggle contract (FR-4) and the archived read-only contract (FR-20) so the responsive-class edits provably don't regress behavior.

**Files:**
- Modify: `frontend/src/pages/ShoppingListDetailPage.tsx` (row container, checkbox, check glyph, item-name classes — lines ~253–287)
- Test: `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` (add mock harness + guard `describe`)

**Interfaces:**
- Consumes: `progressPercent` (Task 1) already exists; not used here.
- Produces: the shared render harness — `mockCheckMutate`, `mockUncheckMutate`, `renderPage()`, and the `mockList(status)` factory — reused by Tasks 3 and 4. Also produces enlarged mobile row markup (base `min-h-[64px]`, `h-7 w-7` checkbox, `h-4 w-4` glyph, `text-[17px]` name) with `md:` resets.

- [ ] **Step 1: Write the failing tests (harness + guards)**

Prepend the mock harness and add a guard `describe` to `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx`. The file's final shape after this step (keep the existing `progressPercent` describe at the bottom):

```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockCheckMutate = vi.fn();
const mockUncheckMutate = vi.fn();
const mockUseShoppingList = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useParams: () => ({ id: "list-1" }), useNavigate: () => vi.fn() };
});

vi.mock("@/lib/hooks/api/use-shopping", () => ({
  useShoppingList: () => mockUseShoppingList(),
  useAddShoppingItem: () => ({ mutate: vi.fn() }),
  useRemoveShoppingItem: () => ({ mutate: vi.fn() }),
  useCheckShoppingItem: () => ({ mutate: mockCheckMutate }),
  useUncheckAllItems: () => ({ mutate: mockUncheckMutate }),
  useArchiveShoppingList: () => ({ mutate: vi.fn(), isPending: false }),
  useUnarchiveShoppingList: () => ({ mutate: vi.fn() }),
  useDeleteShoppingList: () => ({ mutate: vi.fn() }),
  useImportMealPlan: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("@/lib/hooks/api/use-meals", () => ({
  usePlans: () => ({ data: { data: [] } }),
}));

import { ShoppingListDetailPage, progressPercent } from "../ShoppingListDetailPage";

const ITEMS = [
  {
    id: "i1",
    name: "Milk",
    quantity: null,
    category_name: "Dairy",
    category_sort_order: 1,
    checked: false,
    position: 0,
  },
  {
    id: "i2",
    name: "Eggs",
    quantity: "12",
    category_name: "Dairy",
    category_sort_order: 1,
    checked: true,
    position: 1,
  },
];

function mockList(status: "active" | "archived" = "active") {
  return {
    data: {
      data: {
        id: "list-1",
        attributes: {
          name: "Groceries",
          status,
          archived_at: status === "archived" ? "2026-07-01T00:00:00Z" : null,
          items: ITEMS,
        },
      },
    },
    isLoading: false,
  };
}

function renderPage() {
  return render(
    <MemoryRouter>
      <ShoppingListDetailPage />
    </MemoryRouter>,
  );
}

describe("ShoppingListDetailPage — row behavior", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseShoppingList.mockReturnValue(mockList("active"));
  });

  it("toggles an item's checked state when a shopping-mode row is tapped", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    await userEvent.click(screen.getByText("Milk"));
    expect(mockCheckMutate).toHaveBeenCalledWith({ itemId: "i1", checked: true });
  });

  it("does not toggle items in the archived read-only view", async () => {
    mockUseShoppingList.mockReturnValue(mockList("archived"));
    renderPage();
    await userEvent.click(screen.getByText("Milk"));
    expect(mockCheckMutate).not.toHaveBeenCalled();
  });
});

describe("progressPercent", () => {
  it("returns 0 when total is 0 (no divide-by-zero)", () => {
    expect(progressPercent(0, 0)).toBe(0);
  });
  it("returns 0 when nothing is checked", () => {
    expect(progressPercent(0, 4)).toBe(0);
  });
  it("returns 50 at the halfway point", () => {
    expect(progressPercent(2, 4)).toBe(50);
  });
  it("returns 100 when everything is checked", () => {
    expect(progressPercent(4, 4)).toBe(100);
  });
  it("clamps to 100 when checked exceeds total", () => {
    expect(progressPercent(5, 4)).toBe(100);
  });
});
```

Note: the `progressPercent` import moved onto the shared `import { ShoppingListDetailPage, progressPercent } from "../ShoppingListDetailPage";` line — delete the standalone `import { progressPercent }` line from Task 1.

- [ ] **Step 2: Run tests to verify the guard tests pass (harness works)**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: PASS — all 7 tests green. The two guard tests exercise the *existing* tap-toggle and archived-read-only behavior, confirming the harness renders the page and the current row contract holds before any class change.

- [ ] **Step 3: Apply responsive row sizing**

In `frontend/src/pages/ShoppingListDetailPage.tsx`, apply exactly these four class changes inside the per-item row (the `displayItems.map` block, lines ~253–287). Do **not** restructure the DOM — only add responsive classes.

Row container `className` — add `min-h-[64px] md:min-h-0`:

```tsx
                  <div
                    key={item.id}
                    className={cn(
                      "flex items-center gap-3 px-2 py-2 rounded hover:bg-accent/30 transition-colors min-h-[64px] md:min-h-0",
                      shoppingMode && "cursor-pointer",
                      item.checked && "opacity-60",
                    )}
```

Checkbox box — `h-5 w-5` → `h-7 w-7 md:h-5 md:w-5`:

```tsx
                      <div
                        className={cn(
                          "h-7 w-7 md:h-5 md:w-5 rounded border flex items-center justify-center flex-shrink-0",
                          item.checked
                            ? "bg-primary border-primary text-primary-foreground"
                            : "border-muted-foreground",
                        )}
                      >
                        {item.checked && <Check className="h-4 w-4 md:h-3 md:w-3" />}
                      </div>
```

Item name span — `text-sm` → `text-[17px] md:text-sm`:

```tsx
                      <span
                        className={cn(
                          "text-[17px] md:text-sm",
                          item.checked && shoppingMode && "line-through",
                        )}
                      >
                        {item.name}
                      </span>
```

Leave the quantity `<Badge>` (`ml-2 text-xs`) inline and unchanged, per design §4.1.

- [ ] **Step 4: Run tests to verify behavior is preserved**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: PASS — all 7 tests still green. The tap-toggle and archived-read-only contracts are unchanged by the responsive-class edits.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/ShoppingListDetailPage.tsx frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx
git commit -m "feat(shopping): enlarge mobile shopping-mode rows with md: resets"
```

---

## Task 3: Sticky mobile progress header + desktop top-bar gate

Add the mobile-only sticky progress header (bar + "X of Y items" + "Back to Edit") and gate the desktop shopping-mode top action bar to `hidden md:flex`. On mobile the top bar disappears and its progress/back affordances live in the sticky header.

**Files:**
- Modify: `frontend/src/pages/ShoppingListDetailPage.tsx` (gate top action bar ~line 164; insert sticky header before the items block ~line 229)
- Test: `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` (add progress-header `describe`)

**Interfaces:**
- Consumes: `progressPercent` (Task 1); `checkedCount`, `totalCount` (existing, lines 94–95); `setShoppingMode` (existing state); the render harness from Task 2 (`renderPage`, `mockList`).
- Produces: a sticky element with `data-testid="mobile-shopping-progress"` (consumed by no later task, but establishes the "X of Y items"/"Back to Edit" mobile chrome). Task 4 adds the sibling bottom bar.

- [ ] **Step 1: Write the failing tests**

Add this `describe` to `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` (after the "row behavior" describe, before the `progressPercent` describe). It reuses the harness defined in Task 2:

```tsx
describe("ShoppingListDetailPage — sticky mobile progress header", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseShoppingList.mockReturnValue(mockList("active"));
  });

  it("does not render the mobile progress header outside shopping mode", () => {
    renderPage();
    expect(screen.queryByTestId("mobile-shopping-progress")).not.toBeInTheDocument();
  });

  it("renders the sticky progress header with count and Back to Edit in shopping mode", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    const header = screen.getByTestId("mobile-shopping-progress");
    expect(within(header).getByText("1 of 2 items")).toBeInTheDocument();
    expect(within(header).getByText("Back to Edit")).toBeInTheDocument();
  });

  it("does not render the mobile progress header in the archived view", () => {
    mockUseShoppingList.mockReturnValue(mockList("archived"));
    renderPage();
    expect(screen.queryByTestId("mobile-shopping-progress")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: FAIL — the "renders the sticky progress header…" test fails because `getByTestId("mobile-shopping-progress")` finds nothing. (The two negative tests already pass; that's fine.)

- [ ] **Step 3: Gate the desktop top action bar**

In `frontend/src/pages/ShoppingListDetailPage.tsx`, change the shopping-mode action-bar wrapper (line ~164) so it is hidden on mobile only when in shopping mode. The non-shopping branch (Start Shopping / Import) stays visible on all viewports because the extra classes only apply when `shoppingMode` is true:

```tsx
      {!isArchived && (
        <div className={cn("flex gap-2 flex-wrap", shoppingMode && "hidden md:flex")}>
```

Leave the inner `shoppingMode ? ( … ) : ( … )` branches exactly as they are.

- [ ] **Step 4: Insert the sticky mobile progress header**

In `frontend/src/pages/ShoppingListDetailPage.tsx`, insert this block immediately before the `{/* Items grouped by category */}` comment (line ~229), so it sits after the action/quick-add area and pins to the top of `<main>` as the list scrolls:

```tsx
      {/* Sticky mobile progress header (shopping mode, active) */}
      {!isArchived && shoppingMode && (
        <div
          data-testid="mobile-shopping-progress"
          className="md:hidden sticky top-0 z-20 -mx-4 px-4 py-3 bg-background border-b space-y-2"
        >
          <div className="h-2 rounded bg-muted">
            <div
              className="h-2 rounded bg-primary transition-all"
              style={{ width: `${progressPercent(checkedCount, totalCount)}%` }}
            />
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">
              {checkedCount} of {totalCount} items
            </span>
            <Button variant="outline" size="sm" onClick={() => setShoppingMode(false)}>
              <ArrowLeft className="h-4 w-4 mr-1" /> Back to Edit
            </Button>
          </div>
        </div>
      )}
```

`ArrowLeft`, `Button`, `progressPercent`, `checkedCount`, `totalCount`, `setShoppingMode` are all already in scope — no new imports.

- [ ] **Step 5: Run tests to verify they pass**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: PASS — all header tests green; the row-behavior and `progressPercent` suites remain green.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/ShoppingListDetailPage.tsx frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx
git commit -m "feat(shopping): add sticky mobile progress header; gate desktop top bar"
```

---

## Task 4: Sticky mobile bottom action bar

Add the mobile-only sticky bottom bar keeping "Uncheck All" and "Finish Shopping" in thumb reach, mirroring the in-flow `sticky bottom-0` pattern from `bulk-categorize.tsx:170`. Reuses the exact handlers today's desktop buttons use — no new state or mutations.

**Files:**
- Modify: `frontend/src/pages/ShoppingListDetailPage.tsx` (insert bottom bar after the items block, before the Finish-confirmation dialog, ~line 308)
- Test: `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` (add bottom-bar `describe`)

**Interfaces:**
- Consumes: `uncheckAll` (existing hook result), `setShowFinishConfirm` (existing state), the render harness from Task 2 (`renderPage`, `mockList`, `mockUncheckMutate`).
- Produces: a sticky element with `data-testid="mobile-shopping-actions"`. Final task — no downstream consumers.

- [ ] **Step 1: Write the failing tests**

Add this `describe` to `frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx` (after the progress-header describe, before the `progressPercent` describe):

```tsx
describe("ShoppingListDetailPage — sticky mobile bottom bar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseShoppingList.mockReturnValue(mockList("active"));
  });

  it("does not render the bottom action bar outside shopping mode", () => {
    renderPage();
    expect(screen.queryByTestId("mobile-shopping-actions")).not.toBeInTheDocument();
  });

  it("renders Uncheck All and Finish Shopping in shopping mode", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    const bar = screen.getByTestId("mobile-shopping-actions");
    expect(within(bar).getByText("Uncheck All")).toBeInTheDocument();
    expect(within(bar).getByText("Finish Shopping")).toBeInTheDocument();
  });

  it("calls the uncheck mutation when mobile Uncheck All is tapped", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    const bar = screen.getByTestId("mobile-shopping-actions");
    await userEvent.click(within(bar).getByText("Uncheck All"));
    expect(mockUncheckMutate).toHaveBeenCalled();
  });

  it("opens the finish-confirmation dialog when mobile Finish Shopping is tapped", async () => {
    renderPage();
    await userEvent.click(screen.getByText("Start Shopping"));
    const bar = screen.getByTestId("mobile-shopping-actions");
    await userEvent.click(within(bar).getByText("Finish Shopping"));
    expect(screen.getByText("Finish Shopping?")).toBeInTheDocument();
  });

  it("does not render the bottom action bar in the archived view", () => {
    mockUseShoppingList.mockReturnValue(mockList("archived"));
    renderPage();
    expect(screen.queryByTestId("mobile-shopping-actions")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: FAIL — the three positive tests fail because `getByTestId("mobile-shopping-actions")` finds nothing. (The two negative tests already pass.)

- [ ] **Step 3: Insert the sticky mobile bottom action bar**

In `frontend/src/pages/ShoppingListDetailPage.tsx`, insert this block immediately after the items block's closing `)}` (right after the `grouped.map(...)` ternary ends, ~line 307) and immediately before the `{/* Finish Shopping Confirmation */}` comment:

```tsx
      {/* Sticky mobile bottom action bar (shopping mode, active) */}
      {!isArchived && shoppingMode && (
        <div
          data-testid="mobile-shopping-actions"
          className="md:hidden sticky bottom-0 z-20 -mx-4 px-4 py-3 bg-background border-t flex items-center gap-2"
        >
          <Button variant="outline" onClick={() => uncheckAll.mutate()}>
            <RotateCcw className="h-4 w-4 mr-1" /> Uncheck All
          </Button>
          <Button className="flex-1" onClick={() => setShowFinishConfirm(true)}>
            <Archive className="h-4 w-4 mr-1" /> Finish Shopping
          </Button>
        </div>
      )}
```

`RotateCcw`, `Archive`, `Button`, `uncheckAll`, `setShowFinishConfirm` are all already in scope — no new imports.

- [ ] **Step 4: Run tests to verify they pass**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run src/pages/__tests__/ShoppingListDetailPage.test.tsx
```

Expected: PASS — all bottom-bar tests green; every prior suite remains green.

- [ ] **Step 5: Full suite + lint gate**

```bash
export PATH="$HOME/.nvm/versions/node/v22.22.2/bin:$PATH"
npx vitest run && npm run lint
```

Expected: entire frontend test suite PASS and lint clean (no new warnings/errors from the changed file).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/ShoppingListDetailPage.tsx frontend/src/pages/__tests__/ShoppingListDetailPage.test.tsx
git commit -m "feat(shopping): add sticky mobile bottom action bar for shopping mode"
```

---

## Manual verification (after Task 4)

jsdom cannot evaluate Tailwind media queries, so mobile-vs-desktop visual gating is verified manually, not in tests:

- [ ] At a mobile width (375–414px), enter shopping mode on a list with items and confirm: ~64px rows, ~30px checkbox, larger name text; sticky progress header pins under the mobile header while scrolling; sticky bottom bar stays reachable and does **not** obscure the last item; "Back to Edit" returns to edit mode.
- [ ] At `≥ md` (desktop), confirm shopping mode is visually and behaviorally identical to `main`: today's row sizing, the top action bar with inline count, no sticky header, no sticky bottom bar.
- [ ] Confirm edit mode (quick-add, trash targets, ordering) is unchanged at both widths, and the archived read-only view shows enlarged mobile rows without any mutating sticky chrome.

---

## Traceability (FR → task)

- FR-1/2/3/5 (enlarged rows, `md:` reset) → Task 2, Step 3.
- FR-4 (row-tap toggle) → Task 2 guard test + preserved handler.
- FR-6/7/8 (dim, ordering, grouping) → preserved verbatim (no edits to `groupItemsByCategory` / display ordering / `opacity-60` / `line-through`).
- FR-9 (quantity badge visible, inline) → Task 2 (badge left unchanged).
- FR-10/11/12/13 (sticky progress header, count, Back to Edit, desktop count preserved) → Task 3.
- FR-14/15/16/17/18 (sticky in-flow bottom bar, handlers, no overlap, desktop unchanged) → Task 4.
- FR-19 (edit mode unchanged) → no edits to edit-mode chrome; Global Constraints.
- FR-20 (archived read-only, enlarged rows, no mutating chrome) → Task 2 (rows) + Tasks 3/4 negative archived tests.
- FR-21/22 (empty/loading/not-found unchanged) → untouched early-return branches.
- FR-23 (zero-count `0%`) → Task 1 `progressPercent`.
