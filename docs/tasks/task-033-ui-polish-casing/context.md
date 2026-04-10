# Task 033: Context & Key Files

Last Updated: 2026-04-10

---

## Key Files

### Phase 1 — Habits Rename
| File | Purpose | Lines of Interest |
|------|---------|-------------------|
| `frontend/src/components/features/navigation/nav-config.ts` | Nav label + route | Line 49: `"Tracker"` label and `/app/tracker` path |
| `frontend/src/pages/TrackerPage.tsx` | Page title | Line 28: `"Tracker"` heading |
| `frontend/src/App.tsx` | Route definition | Line 63: `path="tracker"` |
| `frontend/src/components/features/tracker/calendar-grid.tsx` | Empty state text | Line 195: `"No tracking items..."` |
| `frontend/src/components/features/tracker/tracker-setup.tsx` | Setup heading + empty state | Lines 41, 46, 82: `"My Tracking Items"`, `"No tracking items yet..."`, `"Delete tracking item"` |
| `frontend/src/components/features/tracker/create-tracker-dialog.tsx` | Dialog title | Line 54: `"Create Tracking Item"` |

### Phase 2 — Input State Fixes
| File | Purpose | Lines of Interest |
|------|---------|-------------------|
| `frontend/src/components/features/tracker/today-view.tsx` | Today view inputs | Lines 153-161: `NumericInput` (count ?? 0), Lines 164-182: `RangeInput` (midpoint default), Lines 136-151: `SentimentInput` (already correct) |
| `frontend/src/components/features/tracker/calendar-grid.tsx` | Calendar cell editor | Lines 397-402: Numeric editor (count ?? 0), Lines 356-371: `RangeEditor` (initial ?? midpoint), Lines 388-395: Sentiment (already correct) |

### Phase 3 — Casing Audit
| File | Purpose | Lines of Interest |
|------|---------|-------------------|
| `frontend/src/pages/WorkoutExercisesPage.tsx` | Exercise casing | Lines 66, 149, 164-166, 172, 178-179 |
| `frontend/src/pages/WorkoutWeekPage.tsx` | Button/dialog casing | Lines 474, 479 |
| `frontend/src/pages/WorkoutTodayPage.tsx` | Badge casing | Line 35: `"Rest day"` |
| `frontend/src/components/features/calendar/recurring-scope-dialog.tsx` | Dialog title casing | Line 16 |
| `frontend/src/pages/WishListPage.tsx` | Dialog title casing | Lines 219, 301 |

### Phase 4 — Cleanup
| File | Purpose | Lines of Interest |
|------|---------|-------------------|
| `frontend/src/components/features/tracker/today-view.tsx` | Redundant button removal | Line 63: Calendar button, Line 35: `onNavigateToCalendar` prop |
| `frontend/src/pages/TrackerPage.tsx` | Prop removal | Line 44: passes `onNavigateToCalendar` |

---

## Key Decisions

1. **Rename to "Habits"** — chosen over "Daily Log", "Check-In", "Pulse", "Journal" per user preference
2. **File/component names stay as-is** — internal naming (TrackerPage, TrackerSetup, etc.) unchanged; only user-facing text changes
3. **Backend unchanged** — casing is display-only; API enum values remain lowercase
4. **Redirect from old route** — `/app/tracker` → `/app/habits` via `<Navigate>` to avoid broken bookmarks
5. **Unset vs zero distinction** — use muted styling and "Not set" / "–" placeholder text, not a separate modal or toggle
6. **Frontend guidelines already updated** — casing rules and cursor:pointer documented in patterns-components.md and ai-guidance.md

---

## Dependencies Between Phases

```
Phase 1 (Rename) ─────────┐
Phase 2 (Input State) ────┤── independent, can be done in any order
Phase 3 (Casing Audit) ───┤
Phase 4 (Cleanup) ────────┘── depends on Phase 1 (remove button from renamed component)
```

All phases modify different code locations. Phase 4 touches `today-view.tsx` which is also modified in Phase 2, so they should be done sequentially (Phase 2 first, then Phase 4).
