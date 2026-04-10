# Task 033: UI Polish — Habits Rename, Input State, and Casing Consistency

Last Updated: 2026-04-10

---

## Executive Summary

This task addresses six UI polish issues across the tracker (renamed to "Habits"), workout, and general frontend surfaces. All changes are frontend-only with zero backend impact. The work naturally splits into three phases: the "Habits" rename (navigation + routing), input state fixes (today view + calendar editor), and a casing consistency audit across all buttons, dialogs, badges, and enum display values.

---

## Current State Analysis

### Tracker Naming
- Navigation label is "Tracker" (`nav-config.ts:49`)
- Page title is "Tracker" (`TrackerPage.tsx:28`)
- Route is `/app/tracker` (`App.tsx:63`)
- Various user-facing strings say "tracking items" in setup and calendar views

### Today View Input State
- **Sentiment**: Correctly highlights selected emoji via `currentRating` prop (`today-view.tsx:99,145`)
- **Numeric**: Falls back to `0` when `currentCount` is `undefined` (`today-view.tsx:154`). The user sees `0` even when nothing has been logged, making it indistinguishable from a saved `0`.
- **Range**: Falls back to midpoint `Math.round((min + max) / 2)` when `currentValue` is `undefined` (`today-view.tsx:167`). Same ambiguity problem as numeric.

### Calendar Cell Editor Zero vs Unset
- **Numeric**: Displays `(entry?.value as NumericValue)?.count ?? 0` — no distinction between unset and zero (`calendar-grid.tsx:399-401`)
- **Range**: `RangeEditor` defaults `initial ?? Math.round((min + max) / 2)` — no distinction between unset and midpoint (`calendar-grid.tsx:357`)
- **Sentiment**: Already distinguishes via button variant (`calendar-grid.tsx:391`)

### Exercise Display Casing
- Kind dropdown shows raw enum values: `"strength"`, `"isometric"`, `"cardio"` (`WorkoutExercisesPage.tsx:164-166`)
- Weight Type dropdown shows: `"free (barbell, dumbbell, etc.)"`, `"bodyweight (dips, pull-ups, etc.)"` (`WorkoutExercisesPage.tsx:178-179`)
- Label reads `"Weight type"` not `"Weight Type"` (`WorkoutExercisesPage.tsx:172`)
- Exercise list shows raw `e.attributes.kind` with no casing (`WorkoutExercisesPage.tsx:66`)

### Button/Dialog Casing Inconsistencies
| File | Line | Current Text |
|------|------|-------------|
| `WorkoutWeekPage.tsx` | 474 | `"+ Add exercise"` |
| `WorkoutWeekPage.tsx` | 479 | `"Add exercise"` (dialog) |
| `WorkoutExercisesPage.tsx` | 149 | `"New exercise"` (dialog) |
| `WorkoutExercisesPage.tsx` | 172 | `"Weight type"` (label) |
| `recurring-scope-dialog.tsx` | 16 | `"Edit recurring event"` / `"Delete recurring event"` |
| `WishListPage.tsx` | 219 | `"Edit wish list item"` / `"Add wish list item"` |
| `WishListPage.tsx` | 301 | `"Delete wish list item?"` |
| `WorkoutTodayPage.tsx` | 35 | `"Rest day"` (badge) |

### Redundant Calendar Button
- `today-view.tsx:63`: Inline `<Button>Calendar</Button>` duplicates the top-right view switcher in `TrackerPage.tsx:33-34`

---

## Proposed Future State

1. Navigation, page title, and route all say "Habits" with a redirect from the old `/app/tracker` path
2. Today view inputs visually distinguish "not yet logged" from "logged as zero/midpoint"
3. Calendar cell editor shows "not set" placeholders for unlogged entries
4. All buttons, dialog titles, badges, and labels use title case
5. Exercise enum values display with first-letter capitalization
6. Frontend dev guidelines document the casing and cursor rules (already done in spec phase)

---

## Implementation Phases

### Phase 1: Habits Rename (Effort: S)

Straightforward string replacements and a route addition.

**Task 1.1** — Update navigation config
- File: `frontend/src/components/features/navigation/nav-config.ts`
- Line 49: Change `label: "Tracker"` to `label: "Habits"`
- Line 49: Change `to: "/app/tracker"` to `to: "/app/habits"`

**Task 1.2** — Update page title
- File: `frontend/src/pages/TrackerPage.tsx`
- Line 28: Change `"Tracker"` to `"Habits"`

**Task 1.3** — Update route and add redirect
- File: `frontend/src/App.tsx`
- Line 63: Change `path="tracker"` to `path="habits"`
- Add `<Route path="tracker" element={<Navigate to="/app/habits" replace />} />` for bookmark compatibility

**Task 1.4** — Update user-facing text referencing "tracking"
- `calendar-grid.tsx:195`: `"No tracking items. Add items in Settings to get started."` → `"No habits yet. Add items in Setup to get started."`
- `tracker-setup.tsx:41`: `"My Tracking Items"` → `"My Habits"`
- `tracker-setup.tsx:46`: `"No tracking items yet..."` → `"No habits yet..."`
- `tracker-setup.tsx:82`: `"Delete tracking item"` → `"Delete Habit"`
- `create-tracker-dialog.tsx:54`: `"Create Tracking Item"` → `"Create Habit"`
- `edit-tracker-dialog.tsx:73`: `"Scale type cannot be changed"` — leave as-is (technical term)

### Phase 2: Input State Fixes (Effort: M)

#### 2A — Today View Current Values

**Task 2.1** — Numeric input: distinguish unset from zero
- File: `frontend/src/components/features/tracker/today-view.tsx`
- `NumericInput` (lines 153-161): When `currentCount` is `undefined`, render the count display with muted styling and show `0` as a placeholder. When `currentCount` is a number (including `0`), display normally.
- Change: Add `const isSet = currentCount !== undefined;` and apply `text-muted-foreground` class when `!isSet`.

**Task 2.2** — Range input: distinguish unset from midpoint
- File: `frontend/src/components/features/tracker/today-view.tsx`
- `RangeInput` (lines 164-182): When `currentValue` is `undefined`, show "Not set" label next to the value display and mute the value text. When set, show the value normally.
- Change: Add `const isSet = currentValue !== undefined;` and conditionally render the value display with muted text or "Not set" label.

**Task 2.3** — Sentiment input: already works correctly
- `SentimentInput` passes `currentRating` and the variant toggles between `"default"` and `"outline"` (line 145). No change needed — verify only.

#### 2B — Calendar Cell Editor Zero vs Unset

**Task 2.4** — Numeric editor: show "not set" placeholder
- File: `frontend/src/components/features/tracker/calendar-grid.tsx`
- `CellEditor` numeric block (lines 397-402): When `entry?.value` is undefined/null, show "–" or "not set" instead of `0`. Track interaction state with a local `touched` flag so first click on +/- initializes to 0 then increments.
- Implementation: Add `const hasEntry = entry?.value != null;` and `const count = hasEntry ? ((entry.value as NumericValue)?.count ?? 0) : null;`. Display `count ?? "–"`. On button click, if `count === null`, save `{ count: 0 }` (for minus) or `{ count: 1 }` (for plus).

**Task 2.5** — Range editor: show "not set" indicator
- File: `frontend/src/components/features/tracker/calendar-grid.tsx`
- `RangeEditor` (lines 356-371): When `initial` is undefined, show "Not set" below the slider and use muted styling. Once the user interacts, remove the indicator. A saved value at midpoint should display normally.
- Implementation: Track `const [touched, setTouched] = useState(initial !== undefined);`. Show "Not set" text when `!touched`. On `onChange`, set `touched = true`.

### Phase 3: Casing Audit (Effort: S)

**Task 3.1** — Exercise display casing
- File: `frontend/src/pages/WorkoutExercisesPage.tsx`
  - Line 66: `{e.attributes.kind}` → capitalize first letter for display
  - Line 149: `"New exercise"` → `"New Exercise"`
  - Line 164-166: `"strength"` → `"Strength"`, `"isometric"` → `"Isometric"`, `"cardio"` → `"Cardio"` (SelectItem display text, keep `value` unchanged)
  - Line 172: `"Weight type"` → `"Weight Type"`
  - Lines 178-179: `"free (barbell, dumbbell, etc.)"` → `"Free (barbell, dumbbell, etc.)"`, `"bodyweight (dips, pull-ups, etc.)"` → `"Bodyweight (dips, pull-ups, etc.)"`

**Task 3.2** — Workout page casing
- File: `frontend/src/pages/WorkoutWeekPage.tsx`
  - Line 474: `"+ Add exercise"` → `"+ Add Exercise"`
  - Line 479: `"Add exercise"` → `"Add Exercise"`
- File: `frontend/src/pages/WorkoutTodayPage.tsx`
  - Line 35: `"Rest day"` → `"Rest Day"`

**Task 3.3** — Calendar recurring dialog casing
- File: `frontend/src/components/features/calendar/recurring-scope-dialog.tsx`
  - Line 16: `"Edit recurring event"` → `"Edit Recurring Event"` and `"Delete recurring event"` → `"Delete Recurring Event"`

**Task 3.4** — Wish list dialog casing
- File: `frontend/src/pages/WishListPage.tsx`
  - Line 219: `"Edit wish list item"` → `"Edit Wish List Item"` and `"Add wish list item"` → `"Add Wish List Item"`
  - Line 301: `"Delete wish list item?"` → `"Delete Wish List Item?"`

### Phase 4: Cleanup (Effort: S)

**Task 4.1** — Remove redundant calendar button
- File: `frontend/src/components/features/tracker/today-view.tsx`
- Line 63: Remove the `<Button variant="outline" size="sm" onClick={onNavigateToCalendar}>Calendar</Button>`
- Remove the `onNavigateToCalendar` prop from the component interface since it's no longer needed
- Update `TrackerPage.tsx:44` which passes this prop

**Task 4.2** — Build verification
- Run `npm run build` from `frontend/` to verify no TypeScript errors
- Run `npm test` if tests exist for affected components
- Visual spot-check in browser if possible

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Old `/app/tracker` bookmarks break | Medium | Low | Redirect route handles this |
| Calendar cell editor interaction feels different | Low | Medium | Preserve existing behavior for set values; only change unset display |
| Missed casing inconsistencies | Low | Low | Full grep audit in Phase 3; can fix incrementally |

---

## Success Metrics

- Zero TypeScript build errors after all changes
- All user-facing text follows title-case convention
- Today view inputs correctly reflect saved state
- Calendar editor clearly distinguishes set from unset values
- Navigation says "Habits" and old route redirects properly

---

## Dependencies

- No external dependencies — all changes are self-contained frontend work
- No backend API changes required
- Frontend dev guidelines already updated (done in spec phase)
