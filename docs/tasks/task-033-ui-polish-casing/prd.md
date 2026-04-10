# UI Polish: Habits Rename, Input State, and Casing Consistency — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-10
---

## 1. Overview

A batch of UI polish fixes spanning the tracker (now "Habits") feature, workout/exercise pages, and a project-wide button/dialog label audit. The changes address user confusion around input state visibility, a zero-vs-unset ambiguity in the calendar view, inconsistent text casing across the UI, and a redundant navigation element.

All changes are frontend-only — no backend or API modifications are required.

## 2. Goals

Primary goals:
- Rename "Tracker" to "Habits" across navigation, routes, page titles, and component references
- Remove the redundant inline "Calendar" button from the Habits today view
- Ensure the today view inputs reflect previously saved values (sentiment highlights, numeric counts, range positions)
- Distinguish intentional zero/midpoint values from unset state in the calendar cell editor
- Capitalize display values for exercise Kind and Weight Type in dropdowns, cards, and labels
- Audit and fix all button labels and dialog titles to follow title-case capitalization

Non-goals:
- Backend API changes (display casing handled in UI layer)
- New features or behavioral changes beyond what's described
- Changing stored enum values — only display formatting

## 3. User Stories

- As a user, I want the "Tracker" feature to be called "Habits" so the name reflects what it actually tracks
- As a user, I want to see my previously saved values when I open the today view, so I know what I already logged
- As a user, I want to intentionally set a zero or minimum value in the calendar editor without it looking like "unset"
- As a user, I want consistent title-case capitalization on all buttons and dialog titles so the UI feels polished
- As a user, I want exercise attributes displayed with proper casing (e.g., "Strength" not "strength")

## 4. Functional Requirements

### 4.1 Rename "Tracker" to "Habits"

- Update navigation label in `nav-config.ts` from "Tracker" to "Habits"
- Update page title in `TrackerPage.tsx` from "Tracker" to "Habits"
- Update route path from `/app/tracker` to `/app/habits` (add redirect from old path for bookmarks)
- Update view switcher button labels if they reference "Tracker"
- Update any toast messages, empty states, or helper text that say "tracker" or "tracking"
- File and component names may remain as-is (internal naming) unless the rename is trivial

### 4.2 Remove Redundant Calendar Button

- Remove the inline `<Button>Calendar</Button>` next to the "Today — [date]" header in `today-view.tsx`
- The top-right view switcher "Calendar" button already provides this navigation

### 4.3 Today View: Show Current Values

- **Sentiment inputs**: Highlight the emoji button matching the saved rating when an entry exists. When no entry exists, no emoji should be highlighted.
- **Numeric inputs**: Display the saved count value from the entry, not a default of `0`. When no entry exists, display `0` but visually distinguish it (e.g., muted styling or a placeholder state).
- **Range inputs**: Position the slider at the saved value from the entry. When no entry exists, show midpoint but visually indicate "not yet logged" (e.g., muted track color or a label like "not set").
- **Notes**: Already loads saved values — no change needed.

### 4.4 Calendar Cell Editor: Zero vs Unset

- **Numeric scale**: When no entry exists, show a "not set" placeholder instead of `0`. Once the user interacts (clicks +/-), treat the value as intentionally set. A saved value of `0` must display as `0` with normal styling.
- **Range scale**: When no entry exists, show slider at midpoint with a "not set" indicator. Once the user adjusts the slider, treat the value as set. A saved midpoint value must display normally.
- **Sentiment scale**: Already distinguishes unset (no highlight) from set. No change needed.
- The key distinction: `entry === undefined/null` means unset; `entry.value` of any number (including 0) means intentionally set.

### 4.5 Exercise Display Casing

- Capitalize the first letter of Kind values in display contexts: "strength" -> "Strength", "isometric" -> "Isometric", "cardio" -> "Cardio"
- Capitalize the first letter of Weight Type values: "free" -> "Free", "bodyweight" -> "Bodyweight"
- Apply in: dropdown `<SelectItem>` labels, exercise list/card displays, any other visible text
- Fix the "Weight type" label to "Weight Type"
- Keep underlying form/API values unchanged (lowercase enums)

### 4.6 Button and Dialog Title Casing Audit

All interactive text labels must use **title case** (capitalize the first letter of each significant word). Known fixes:

| Location | Current | Fixed |
|----------|---------|-------|
| WorkoutWeekPage button | "Add exercise" | "Add Exercise" |
| WorkoutWeekPage dialog | "Add exercise" | "Add Exercise" |
| WorkoutExercisesPage dialog | "New exercise" | "New Exercise" |
| WorkoutExercisesPage label | "Weight type" | "Weight Type" |
| CalendarPage recurring edit dialog | "Edit recurring event" | "Edit Recurring Event" |
| CalendarPage recurring delete dialog | "Delete recurring event" | "Delete Recurring Event" |
| WishListPage dialog (add) | "Add wish list item" | "Add Wish List Item" |
| WishListPage dialog (edit) | "Edit wish list item" | "Edit Wish List Item" |
| WishListPage dialog (delete) | "Delete wish list item?" | "Delete Wish List Item?" |
| WorkoutTodayPage badge | "Rest day" | "Rest Day" |

Perform a full audit of all `<Button>`, `<DialogTitle>`, and `<Badge>` text across the frontend to catch any additional inconsistencies not listed above.

## 5. API Surface

No API changes. All modifications are display-layer only.

## 6. Data Model

No data model changes.

## 7. Service Impact

| Service | Impact |
|---------|--------|
| frontend | All changes — navigation, components, display formatting, casing |
| No backend services | No changes needed |

## 8. Non-Functional Requirements

- **Performance**: No impact — changes are purely presentational
- **Accessibility**: Title-case labels improve screen reader clarity
- **Multi-tenancy**: No impact — no data or API changes

## 9. Open Questions

None — all scope decisions confirmed.

## 10. Acceptance Criteria

- [ ] Navigation shows "Habits" instead of "Tracker"
- [ ] Route is `/app/habits` with redirect from `/app/tracker`
- [ ] Page title reads "Habits"
- [ ] No inline "Calendar" button in the today view header
- [ ] Sentiment inputs in today view highlight the saved emoji (or none if unset)
- [ ] Numeric inputs in today view display the saved count (not hardcoded 0)
- [ ] Range inputs in today view show the saved position (not always midpoint)
- [ ] Calendar numeric editor shows "not set" placeholder when no entry exists, and displays `0` normally when saved as `0`
- [ ] Calendar range editor shows "not set" indicator when no entry exists, and displays midpoint normally when saved as midpoint
- [ ] Exercise Kind values display with first letter capitalized
- [ ] Exercise Weight Type values display with first letter capitalized
- [ ] "Weight Type" label is properly capitalized
- [ ] All button labels use title case
- [ ] All dialog titles use title case
- [ ] All badge text uses title case
- [ ] Frontend dev guidelines updated with casing rules and cursor requirements
