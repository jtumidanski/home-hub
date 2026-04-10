# Task 033: Progress Checklist

Last Updated: 2026-04-10

---

## Phase 1: Habits Rename (S)

- [x] 1.1 Update nav-config.ts — label "Habits", route `/app/habits`
- [x] 1.2 Update TrackerPage.tsx — page title "Habits"
- [x] 1.3 Update App.tsx — route path + redirect from old `/app/tracker`
- [x] 1.4 Update user-facing "tracking" text in calendar-grid, tracker-setup, create-tracker-dialog

## Phase 2: Input State Fixes (M)

- [x] 2.1 NumericInput (today-view.tsx) — muted styling when unset
- [x] 2.2 RangeInput (today-view.tsx) — "Not set" label when unset
- [x] 2.3 Verify SentimentInput already works correctly (no changes expected)
- [x] 2.4 CellEditor numeric (calendar-grid.tsx) — "–" placeholder when no entry
- [x] 2.5 RangeEditor (calendar-grid.tsx) — "Not set" indicator when no entry

## Phase 3: Casing Audit (S)

- [x] 3.1 Exercise casing — kind/weight type display, dropdown labels, "Weight Type" label
- [x] 3.2 Workout page casing — "Add Exercise", "Rest Day"
- [x] 3.3 Calendar recurring dialog — "Edit Recurring Event", "Delete Recurring Event"
- [x] 3.4 Wish list dialogs — "Add/Edit/Delete Wish List Item"
- [x] 3.5 Full grep audit for any remaining casing inconsistencies

## Phase 4: Cleanup (S)

- [x] 4.1 Remove redundant Calendar button from today-view.tsx + prop cleanup
- [x] 4.2 Build verification (`npm run build`)
- [x] 4.3 Test verification (`npm test`)

## Guidelines (already done)

- [x] Frontend dev guidelines updated — casing rules in patterns-components.md
- [x] Frontend dev guidelines updated — cursor:pointer rule in patterns-components.md
- [x] AI guidance updated — rules 13-15 and validation checklist in ai-guidance.md
