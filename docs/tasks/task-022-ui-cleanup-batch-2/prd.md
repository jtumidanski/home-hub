# UI Cleanup Batch 2 — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-07
---

## 1. Overview

This task is a second batch of frontend polish items (companion to task-017) that addresses eight defects spread across the Tasks page, Calendar, Cook Mode, Ingredient pages, and shared UI primitives. Each item is small in isolation, but together they remove paper-cuts that erode confidence in the rest of the product.

Most items are pure frontend tweaks. Item 8 (stale "uncategorized" badge on the Ingredients list) requires a fix in `recipe-service` because the bug persists after a hard refresh — the backend itself is returning a stale `categoryName` value sourced from a legacy local table that is no longer the source of truth.

## 2. Goals

Primary goals:
- Stop classifying tasks due *today* as "overdue" on the Tasks page.
- Make every interactive control on grid pages, page headers, and dialogs show a pointer cursor on hover.
- Replace the static "check" complete-toggle icon on the Tasks page with a context-sensitive icon pair so the action is self-explanatory.
- Make all-day calendar events render as a full-width row in their day column instead of shrinking to text width.
- Show recipe **names** under "Used in Recipes" on the Ingredient detail page (not the raw usage name).
- Fix the Cook Mode single-step view so the top of long step text is reachable.
- Fix the Ingredients list page so an ingredient categorized after the initial load shows its category instead of "uncategorized".

Non-goals:
- A repo-wide cursor audit beyond the listed surfaces (grid row actions, "Clear Filters" buttons, page header action buttons, dialog `X` buttons, dialog action buttons).
- Redesigning the Tasks page, Cook Mode, or Calendar layout.
- Adding new task states (e.g., a separate "due today" status enum) — item 1 is purely a date-comparison fix.
- Restructuring how categories are stored across `recipe-service` and `category-service`. Item 8 is fixed by removing the stale enrichment, not by re-architecting the boundary.

## 3. User Stories

- As a user looking at the Tasks page on the morning a task is due, I see it as a normal pending task — not an alarming "overdue" state.
- As a user hovering over any clickable control in the app, I see a pointer cursor so I know it's interactive.
- As a user clicking the complete-toggle on a task, the icon I see on a *completed* task is visually distinct from the icon on an *incomplete* task, so I can tell at a glance which action the click will perform.
- As a user with all-day events on the calendar, each all-day event spans the full width of the day it belongs to (multi-day events span the full width of every covered day) instead of shrinking to fit its title.
- As a user looking at an ingredient's detail page, the "Used in Recipes" section lists the actual recipe titles so I can recognize them.
- As a user reading a long step in Cook Mode's single-step view, I can read the *whole* step from the top — not just the bottom two-thirds.
- As a user who categorizes an ingredient from its detail page, when I navigate back to the Ingredients list, that ingredient shows its new category instead of "uncategorized".

## 4. Functional Requirements

### 4.1 "Due today" is not overdue (Item 1)

- File: `frontend/src/types/models/task.ts:45` — `isTaskOverdue(task)`.
- Current implementation: `new Date(dueOn) < new Date()`. `dueOn` is a date-only string (e.g. `"2026-04-07"`); `new Date("2026-04-07")` parses to UTC midnight, which is *before* the current local instant for any user east of UTC, marking due-today tasks as overdue.
- Fix: compare on calendar-day boundaries in the user's local timezone. A task is overdue iff `status === "pending"` AND `dueOn` is strictly before *today* (local).
- Implementation must:
  - Parse `dueOn` as a local date (split `YYYY-MM-DD` and construct via `new Date(year, month - 1, day)`), not via the timezone-ambiguous `new Date(string)` constructor.
  - Build `today` from `new Date()` zeroed to local midnight.
  - Return `dueDate < today`.
- Add unit tests in `frontend/src/types/models/__tests__/task.test.ts` (create if missing) covering:
  - `dueOn === today` → not overdue.
  - `dueOn === yesterday` → overdue.
  - `dueOn === tomorrow` → not overdue.
  - Completed task with past `dueOn` → not overdue.
  - `dueOn === null` → not overdue.
- The Tasks page status filter dropdown's "Overdue" option (`frontend/src/pages/TasksPage.tsx:23-27`) and the status badge logic (`TasksPage.tsx:136-143`) require no changes — they consume `isTaskOverdue` and inherit the fix automatically.

### 4.2 Cursor affordances on interactive controls (Item 2 + 4)

The shadcn `Button` primitive does not currently set `cursor-pointer`. Native `<button>` elements default to `cursor: default` in browsers, so users see an arrow when hovering buttons.

Apply a **single global fix** in the relevant primitives:

- **`frontend/src/components/ui/button.tsx`** — add `cursor-pointer` to the base class string in the `buttonVariants` definition. This covers:
  - Tasks page complete-toggle and delete row actions.
  - Tasks page "New Task" header button.
  - All "Clear Filters" buttons across pages (they use `<Button variant="link">` or `<Button variant="ghost">`).
  - Top-right page action buttons (e.g., Ingredients page "Categories", "Bulk Edit", "New").
  - Dialog action buttons (e.g., "Create Reminder", "Create Task").
- **`frontend/src/components/ui/dialog.tsx`** — add `cursor-pointer` to the `DialogClose` / `X` button base class so the dialog close affordance hovers correctly.
- **`frontend/src/components/ui/sheet.tsx`** — apply the same fix to its close button if it has one (verify and patch if so).

Constraints:
- The fix must not affect `disabled` buttons. Use the existing disabled-state selector (`disabled:pointer-events-none` already neutralizes hover, but verify cursor doesn't appear pointer when disabled).
- Verify the change does not break any existing snapshot tests (`frontend/src/components/ui/__tests__/`). Update snapshots if necessary.
- Do **not** add per-call-site `className="cursor-pointer"` additions for any button covered by the global fix.
- Audit and patch any *non-`Button`* clickable elements on the listed surfaces:
  - Tasks page row actions if they use bare `<button>` instead of `<Button>` (none expected — verify).
  - The Ingredients page card-level "X" close on the search input (`IngredientsPage.tsx:185-193`) — currently a bare `<button>`; add `cursor-pointer`.

### 4.3 Context-sensitive complete-toggle icon on Tasks page (Item 3)

- File: `frontend/src/pages/TasksPage.tsx:101-118` (desktop table column) and `frontend/src/components/features/tasks/task-card.tsx` (mobile card).
- Replace the static `Check` icon with a pair from `lucide-react`:
  - **Incomplete (`status !== "completed"`)** → `Circle` (empty circle, muted color).
  - **Completed (`status === "completed"`)** → `CheckCircle2` (filled check circle, primary color).
- The button's `aria-label` must reflect the action it will perform: `"Mark complete"` when incomplete, `"Mark incomplete"` when complete.
- Apply identical behavior in `task-card.tsx` so mobile and desktop are consistent.
- Existing toast messages ("Task completed" / "Task reopened") remain unchanged.

### 4.4 All-day events span their full day column (Item 5)

- File: `frontend/src/components/features/calendar/all-day-event-row.tsx`.
- Current behavior: events render inside a `flex flex-wrap gap-1` container as inline-sized buttons that shrink to title width.
- Required behavior:
  - Each all-day event renders as a **block** that spans the full width of the day column (`w-full`).
  - When a day has multiple all-day events, they stack vertically (one per row), each spanning full width.
  - The container's `min-h-[28px]` and `px-1 py-1` paddings remain to preserve grid alignment.
- Multi-day all-day events: a single all-day event that covers multiple days must visually span the full width of *every* day column it covers. The current implementation passes per-day buckets to `AllDayEventRow` from `calendar-grid.tsx:100-104`, which renders each event inside its own day cell — preserve this per-day rendering, but each day's instance must be full-width.
- Truncation: titles longer than the day column width must `truncate` (existing `truncate` class is fine) without wrapping.
- The popover-on-click behavior is unchanged.

### 4.5 "Used in Recipes" shows recipe names (Item 6)

- File: `frontend/src/pages/IngredientDetailPage.tsx:241-261`.
- Current behavior: each entry renders `{ref.rawName} (recipe)` where `rawName` is the unparsed ingredient string from the recipe (e.g. "1 cup flour, sifted"). This is the *usage* of the ingredient inside the recipe, not the recipe's title.
- Required behavior: each entry renders the recipe's **title** (e.g. "Banana Bread") and remains a clickable button that navigates to `/app/recipes/{recipeId}`.
- Backend impact:
  - The endpoint `GET /ingredients/{id}/recipes` (handled by `recipe-service`) currently returns `RecipeRef { recipeId, rawName }` (see `services/recipe-service/internal/ingredient/provider.go:97-103`). Add a `recipeName` field to `RecipeRef` populated via a JOIN against the `recipes` table on `recipe_id`.
  - Update the JSON:API serializer in `recipe-service` to expose `recipeName` (preserve `rawName` for any other consumer).
  - Add a Go test verifying the join returns the recipe name.
- Frontend impact:
  - Add `recipeName: string` to the matching TypeScript type for `RecipeRef`.
  - Render `{ref.recipeName}` instead of `{ref.rawName} (recipe)`.
  - If `recipeName` is empty (defensive — recipe deleted), fall back to `"(unknown recipe)"`.

### 4.6 Cook Mode single-step view: long text reachable (Item 7)

- File: `frontend/src/components/features/recipes/cook-mode.tsx:191-210` (`SingleStepView`).
- Root cause: the content container is `flex-1 flex flex-col items-center justify-center ... overflow-y-auto`. When step content exceeds the viewport, `justify-center` centers the flex item *as if it could fit*, pushing the start of the content above the scroll origin and making it unreachable. This is the well-known flexbox+overflow centering trap.
- Required fix:
  - Remove `justify-center` from the scroll container.
  - Center the content vertically *only when it fits*. Recommended pattern: wrap the content in an inner `<div className="m-auto">` or use `min-h-full` with `flex` on the inner content. Concretely:
    ```tsx
    <div className="flex-1 overflow-y-auto px-6 md:px-12">
      <div className="min-h-full flex flex-col items-center justify-center py-6">
        {/* section header + step text */}
      </div>
    </div>
    ```
    With `min-h-full` on the inner wrapper, short content stays vertically centered; long content grows past viewport height and the outer container scrolls from the top of the inner wrapper.
- Behavior must be identical for short content (still centered) and fixed for long content (scrolls from the top of the step text, not from a vertical mid-point).
- Verify on both mobile and desktop breakpoints.
- The existing keyboard arrow nav, swipe handlers, and step counter must remain unchanged.

### 4.7 Ingredients list shows current category (Item 8)

- Root cause (verified): the list endpoint `GET /ingredients` in `recipe-service` calls `searchWithUsage` (`services/recipe-service/internal/ingredient/provider.go:126-170`), which joins `canonical_ingredients` against a **local** `ingredient_categories` table (line 149). Per `services/recipe-service/internal/ingredient/entity.go:38-42`, categories were moved to `category-service` and the FK constraint dropped — but the legacy local `ingredient_categories` table remains and is no longer kept in sync. When a user assigns a category (created in `category-service`) to an ingredient, recipe-service stores the `category_id` correctly, but the list query's JOIN against the stale local table returns an empty `category_name`, so the frontend renders "uncategorized". The Ingredient detail page is unaffected because it resolves the category name client-side from `useIngredientCategories` (which queries `category-service`).
- Required fix:
  - Remove the stale `ingredient_categories` JOIN and `category_name` SELECT from `searchWithUsage` in `services/recipe-service/internal/ingredient/provider.go`.
  - Drop the `CategoryName string` field from `entityWithUsage` and `entityWithCategory` (`provider.go:115-124`) — and from `Model` / `Builder` (`model.go:43,57,69`, `builder.go:22,38,59`) if no other call site needs it. If another call site does still need it, leave the type field in place but set it to empty in this code path; the frontend stops trusting it either way.
  - Drop `categoryName` from the JSON:API response in `services/recipe-service/internal/ingredient/rest.go:15` and `:32` (or set it to empty). The list response must continue to expose `categoryId`.
  - Update `frontend/src/pages/IngredientsPage.tsx:249-253` to resolve the category name client-side from the `categories` array already loaded via `useIngredientCategories`, using `ingredient.attributes.categoryId` as the lookup key. Same lookup pattern as `IngredientDetailPage.tsx:188-189`.
  - Update `frontend/src/types/models/ingredient.ts` to drop `categoryName` from `CanonicalIngredientListItem` (or keep it optional for compatibility — prefer dropping per the project guideline against backwards-compat shims).
  - The "uncategorized count" derivation in `IngredientsPage.tsx:64-71` continues to work because it relies on `total` and category counts, not `categoryName`.
- Cleanup (in scope): a code search confirmed the only references to `ingredient_categories` in `recipe-service` are the SELECT and JOIN at `provider.go:148-149` that this task removes. After the JOIN is removed, the table is fully unreferenced. Add a migration in `recipe-service` to `DROP TABLE IF EXISTS ingredient_categories` so the orphaned table is removed cleanly.

## 5. API Surface

| Endpoint | Service | Change |
|---|---|---|
| `GET /ingredients` (list) | recipe-service | Stop returning `categoryName` (or always return empty). `categoryId` continues to be returned. |
| `GET /ingredients/{id}/recipes` | recipe-service | Add `recipeName` field to each `RecipeRef` entry. `rawName` is preserved. |

No new endpoints. No URL or auth changes.

## 6. Data Model

One schema change in `recipe-service`: drop the orphaned `ingredient_categories` table via a migration (`DROP TABLE IF EXISTS ingredient_categories`). The table has no remaining readers after item 8's JOIN removal — verified via code search. Categories are managed exclusively by `category-service`.

## 7. Service Impact

| Service | Items | Notes |
|---|---|---|
| **frontend** | 1, 2, 3, 4, 5, 6, 7, 8 | All eight items touch the frontend; items 6 and 8 also need backend changes. |
| **recipe-service** | 6, 8 | Item 6: extend `RecipeRef` with recipe name via JOIN. Item 8: remove stale `ingredient_categories` JOIN from `searchWithUsage`. |

No other services are affected.

## 8. Non-Functional Requirements

- **Multi-tenancy:** the recipe-service queries already scope by `tenant_id`; preserve those filters in any modified query.
- **Performance:** the new JOIN against `recipes` for "Used in Recipes" must not regress query latency. Ingredient → recipes is a small list (paginated). No new index required.
- **Accessibility:**
  - Item 3 — `aria-label` on the complete-toggle button must reflect the action.
  - Item 7 — Cook Mode keyboard navigation must remain functional after the layout change.
- **Backwards compatibility:** per the project guideline against backwards-compat shims, drop `categoryName` from the list response cleanly rather than gating it behind a feature flag.
- **Tests:** unit tests required for the date-comparison fix (item 1) and the recipe-name JOIN (item 6). Other items are visual / behavioral fixes verified manually.

## 9. Open Questions

None — all questions resolved during spec review.

## 10. Acceptance Criteria

- [ ] A task with `dueOn === today` displays as "pending" (not "overdue") on the Tasks page, regardless of the user's local timezone.
- [ ] Unit tests for `isTaskOverdue` cover today / yesterday / tomorrow / completed / null cases and pass.
- [ ] Hovering any of the following shows a pointer cursor: Tasks page row action buttons, "Clear Filters" buttons across all list pages, top-right page action buttons (Ingredients, Tasks, etc.), dialog `X` close buttons, dialog action buttons (e.g. "Create Reminder").
- [ ] On the Tasks page, an incomplete task shows an empty `Circle` icon and a completed task shows a filled `CheckCircle2` icon; clicking either toggles state and the icon updates.
- [ ] The complete-toggle button's `aria-label` is "Mark complete" when incomplete and "Mark incomplete" when complete.
- [ ] All-day calendar events render as full-width blocks within their day column; multi-day events span every covered day at full width; long titles truncate without wrapping.
- [ ] On the Ingredient detail page, "Used in Recipes" entries display the recipe title (not the raw ingredient usage string) and clicking still navigates to the recipe detail page.
- [ ] In Cook Mode single-step view, a step with text long enough to overflow the viewport is scrollable from the top of the text; short steps remain vertically centered.
- [ ] After categorizing an ingredient on its detail page and navigating back to the Ingredients list, the ingredient displays its assigned category (no hard refresh required, and the category persists across hard refresh).
- [ ] `recipe-service` no longer JOINs the local `ingredient_categories` table in `searchWithUsage`; the list response either omits `categoryName` or always returns it empty.
- [ ] A `recipe-service` migration drops the orphaned `ingredient_categories` table.
- [ ] The Ingredients list page resolves category names client-side from `useIngredientCategories` keyed by `categoryId`.
- [ ] All affected services build and existing tests pass.
