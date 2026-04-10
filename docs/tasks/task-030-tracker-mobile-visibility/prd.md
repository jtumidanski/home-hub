# Tracker Mobile Visibility — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-09
---

## 1. Overview

The tracker page's mobile views (Today and Calendar day view) lack clear visual indicators for whether tracker items have been logged, skipped, or still need attention. The current Today view uses a barely-perceptible border opacity change for unset items and small green text for logged ones. The Calendar mobile day view has no visual differentiation at all for unset cards.

This makes it difficult for users to glance at the tracker and quickly understand their progress. The goal is to make item states immediately obvious on mobile through improved visual treatments, consistent across both views.

## 2. Goals

Primary goals:
- Make unset (needs attention) tracker items visually distinct and prominent on mobile
- Make logged items feel "done" with a clear success indicator
- Make skipped items visually distinct from both logged and unset items
- Provide consistent state treatment across Today view and Calendar mobile day view
- Add a progress bar to the Today view matching the Calendar view's existing pattern

Non-goals:
- Changing the desktop calendar grid styling (colored fills and dashed borders work well at that scale)
- Adding new features, API endpoints, or data model changes
- Modifying the Report view
- Changing the Setup/Settings view

## 3. User Stories

- As a user checking my tracker on my phone, I want to immediately see which items I haven't logged yet so I can quickly complete my daily tracking
- As a user glancing at my tracker, I want logged items to feel "done" so I get a sense of accomplishment and can focus on remaining items
- As a user who skipped an item, I want it to look different from unlogged items so I know I made a conscious decision about it
- As a user opening the Calendar day view, I want the same clear visual states as the Today view so the experience is consistent

## 4. Functional Requirements

### 4.1 Unset Item Treatment (Both Views)

- Cards for items that have not been logged must have a colored left border (3px) using the item's assigned color, creating a "to-do" visual cue
- The card background should remain default (no tint) to contrast with logged items
- No badge is needed for unset items — the left border accent and absence of a "logged" indicator is sufficient

### 4.2 Logged Item Treatment (Both Views)

- Replace the current small green "logged" text with a Badge component containing a check icon and "logged" text
- Badge should use a subtle success style (green tint) that is clearly visible but not overwhelming
- The card's left border accent should be removed (or replaced with a subtle success color) to reduce visual weight compared to unset items

### 4.3 Skipped Item Treatment (Both Views)

- Show a "skipped" Badge (muted/gray style) consistently in both Today and Calendar day views
- The card should have reduced opacity or a muted background to visually de-emphasize it
- The left border accent should not be present on skipped items
- Today view currently shows no indicator for skipped items — this must be fixed

### 4.4 Today View Progress Bar

- Add a progress bar above or below the item list, matching the Calendar view's existing pattern
- Show filled count and total: "{filled + skipped}/{expected} entries"
- Use the same `bg-primary` bar style as the Calendar view's progress bar

### 4.5 State Summary Table

| State | Left Border | Badge | Card Style |
|-------|------------|-------|------------|
| Unset (needs attention) | 3px item color | None | Default background |
| Logged | None | Green check "logged" | Default or subtle success tint |
| Skipped | None | Gray "skipped" | Muted/reduced opacity |

## 5. API Surface

No API changes required. This is a frontend-only change.

## 6. Data Model

No data model changes required.

## 7. Service Impact

### Frontend (`frontend/`)

- `src/components/features/tracker/today-view.tsx` — Update card rendering for all three states (unset, logged, skipped), add progress bar, use Badge component for logged/skipped indicators
- `src/components/features/tracker/calendar-grid.tsx` — Update `MobileDayView` card rendering to match Today view's state treatments

No other services are affected.

## 8. Non-Functional Requirements

- All visual changes must work in both light and dark mode
- Changes must not affect the desktop calendar grid layout or cell styling
- Transitions/animations should be minimal — no jarring visual effects
- Badge and border colors must use existing Tailwind color classes (no custom CSS)

## 9. Open Questions

None — scope is well-defined and contained.

## 10. Acceptance Criteria

- [ ] Unset items in Today view have a visible colored left border using the item's color
- [ ] Unset items in Calendar mobile day view have the same colored left border
- [ ] Logged items in both views show a Badge with check icon and "logged" text instead of plain green text
- [ ] Skipped items in Today view show a "skipped" Badge (currently missing)
- [ ] Skipped items in Calendar mobile day view show a "skipped" Badge (already exists, verify consistency)
- [ ] Skipped items have muted/reduced visual weight in both views
- [ ] Today view includes a progress bar showing completion count
- [ ] All states render correctly in dark mode
- [ ] Desktop calendar grid is unchanged
