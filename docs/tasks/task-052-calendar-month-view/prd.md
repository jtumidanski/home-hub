# Calendar Month View — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-06-09
---

## 1. Overview

The Home Hub calendar page currently renders a single time-grid layout: a 7-day week
on desktop/tablet and a rolling 3-day window on mobile. This is excellent for seeing the
hour-by-hour shape of the next few days, but it gives no at-a-glance sense of a whole
month — users cannot scan ahead to spot a busy weekend three weeks out, find which days
of the month are empty, or get the "big picture" planning view that every consumer
calendar app offers.

This feature adds a **Month view mode** as an *addition alongside* the existing view, not
a replacement. A Week/Month toggle in the page header lets the user switch between the
current time-grid view (default) and a traditional month grid — a 7-column, 5–6 row
layout where each cell is a calendar day showing that day's events as compact chips
(desktop/tablet) or colored dots (mobile). The week view, its 3-day mobile variant, event
CRUD, household color legend, package overlay, and date picker all continue to work
exactly as today. Month view is purely a new way to *visualize and navigate to* the same
events.

This is a **frontend-only** feature. Events are already fetched from `calendar-service`
over an arbitrary `[start, end]` date range via `useCalendarEvents(startISO, endISO)`; the
month view simply requests a wider range (the visible month grid, including leading and
trailing days from adjacent months) and renders it differently. No new endpoints, schema
changes, or backend work are anticipated.

## 2. Goals

Primary goals:
- Add a Month view mode to the calendar page, toggleable with the existing Week view.
- Render a standard month grid (7 columns; 5–6 week rows) covering the full visible month,
  including leading/trailing days from adjacent months to complete the first and last rows.
- Show each day's events compactly in its cell, color-coded by household member, including
  package-overlay events, consistent with the week view's coloring.
- Let users navigate by month (prev/next) and jump to any month via the existing date picker.
- Let users drill from a month day into that day's detail by switching back to the week
  view focused on the clicked day.
- Keep the week/3-day view as the default and leave it functionally unchanged.

Non-goals:
- Any backend/API/schema change or new `calendar-service` endpoint.
- A separate single-day view or an agenda/list view.
- Drag-to-create, drag-to-move, or drag-to-resize events in the month grid.
- Creating or editing events directly from a month cell (creation still happens via the
  existing "Add Event" button and the week view's slot-click).
- Persisting the chosen view mode across page visits or sessions (defaults to Week each load).
- Printing or export of the month view.
- Recurrence-rule changes (recurring events already expand server-side and render as
  individual instances).

## 3. User Stories

- As a household member, I want to switch the calendar to a month view so that I can see
  the whole month at a glance and plan ahead.
- As a household member, I want each day in the month grid to show its events color-coded
  by who they belong to, so that I can tell whose commitments fall on which day.
- As a household member, I want to move forward and backward a month at a time, and jump
  to any month with the date picker, so that I can look ahead or back quickly.
- As a household member, I want to click a day in the month view and land on that day in
  the week view, so that I can see its hour-by-hour detail and act on individual events.
- As a household member on a busy day, I want to scroll within a day cell to see all of
  that day's events without leaving the month view.
- As a household member on mobile, I want a compact month overview using dots, and tapping
  a day takes me to that day's detail, so that the month grid stays legible on a small screen.
- As a household member, I want my expected default (the current week/time-grid view) to
  load every time, with Month being an opt-in I choose per visit.

## 4. Functional Requirements

### 4.1 View Mode Toggle

- A **Week / Month** toggle is added to the calendar page header, near the existing
  prev/Today/next navigation cluster.
- The toggle has two states: `week` (the existing time-grid view; default) and `month`
  (the new month grid).
- The page **defaults to `week`** on every load. The selected mode is held in component
  state only for the duration of the page session — it is **not** persisted to
  localStorage, and **not** reflected in the URL.
- Switching modes must not lose the user's current position in time: switching Week→Month
  shows the month containing the currently focused week start; switching Month→Week shows
  the week containing the currently focused day (see 4.5 for the day-click drill-down,
  which is a distinct path).
- The "Add Event" button (shown only when the user has write access) remains visible in
  both modes. The date-picker popover ("Today" button) remains visible in both modes.

### 4.2 Month Grid Layout

- The month grid is a 7-column layout, one column per weekday, with the same week-start
  convention as the week view (weeks start on **Sunday**, matching `getStartOfWeek`).
- The grid shows **complete weeks**: it begins on the Sunday on or before the first of the
  month and ends on the Saturday on or after the last of the month. This yields **5 or 6
  rows** depending on the month.
- Column headers show weekday abbreviations (e.g., SUN…SAT), consistent with the week
  view's `weekday: "short"` formatting.
- Each cell displays the **day-of-month number**. Today's cell is visually highlighted
  using the same primary-color treatment the week view uses for "today".
- Cells for **leading/trailing days** (days belonging to the previous or next month that
  fill out the first/last rows) are visually de-emphasized (muted) but still render their
  events and are still clickable.
- The grid fills the available page height (consistent with the week view filling
  `flex-1 min-h-0`).

### 4.3 Event Rendering Within a Day Cell (Desktop/Tablet)

- Each day cell lists that day's events as compact **chips**. A chip shows the event title
  (truncated) and is colored by the event owner's household color (`userColor`), matching
  the week view's color semantics.
- All-day events and timed events both appear as chips in the cell. Timed-event chips
  should include a compact start-time indicator (e.g., a small leading time label);
  all-day events render without a time. (Exact visual treatment is a design-phase concern;
  the requirement is that the two are distinguishable.)
- **Package-overlay events** (from `packagesToCalendarEvents`) appear in month cells using
  the same coloring/labeling they use in the week view.
- When a day has more events than fit in the cell's visible area, the cell becomes
  **internally scrollable** — the user scrolls within that day cell to reveal the
  remaining events. There is **no** "+N more" popover and no silent truncation; every
  event for the day is reachable by scrolling the cell.
- Event chips are read-only affordances in month view for the purpose of *viewing*; clicking
  a chip drills into the day (see 4.5). Editing/deleting events is done from the week view,
  not from month cells.

### 4.4 Event Rendering Within a Day Cell (Mobile)

- On mobile (below the existing `md` / 768px breakpoint), a full text-chip month grid is
  too cramped. Instead, each day cell renders **compact colored dots** — one dot per event
  (or, if a day is very dense, a representative set of dots keyed by household-member color)
  — so the month stays legible.
- Dots use the same household-member colors as chips so the "who" signal is preserved.
- Tapping a day cell on mobile drills into that day (see 4.5); the day's individual events
  are then seen in the (3-day) time-grid view. There is no per-dot tap target.
- The Week/Month toggle is available on mobile as well as desktop.

### 4.5 Day-Cell Click / Drill-Down

- Clicking (or tapping) a day cell — including leading/trailing days from adjacent months —
  **switches the page back to the week view focused on the clicked day**:
  - On desktop/tablet: switch to `week` mode and set the focused week to the week
    containing the clicked day (`getStartOfWeek(clickedDay)`), so the clicked day is visible
    in the 7-day grid.
  - On mobile: switch to `week` mode (which renders the 3-day window) positioned so the
    clicked day is in view, consistent with the existing mobile windowing logic
    (today − 1 day as the window start is the current convention; the equivalent is
    clickedDay − 1 day as the window start).
- Clicking an event chip behaves the same as clicking its day cell (drills into the day).
  Month view does not open the edit dialog directly.

### 4.6 Navigation

- In month mode, the **prev/next** controls move by **one calendar month** (not by
  `dayCount` days as in week mode).
- The header's date-range label shows the **month and year** (e.g., "June 2026") in month
  mode, replacing the week view's date-range string.
- The **date picker** ("Today" popover with the `Calendar` component) works in month mode:
  selecting a date sets the focused month to that date's month. Selecting a date does **not**
  by itself switch to week mode (only a day-cell click does).
- A clear way to return to the current month (analogous to the week view's "Today") is
  available — e.g., the date-picker trigger continues to act as "Today".

### 4.7 Data Fetching

- In month mode, `useCalendarEvents` is called with the start/end of the **visible grid**
  (the leading Sunday through the trailing Saturday, exclusive end), so every rendered cell
  has its events.
- Switching months issues a new range query for the newly visible grid. Loading and error
  states reuse the existing patterns (skeleton while loading, `ErrorCard` on failure).
- The existing package query and merge (`events = [...calendarEvents, ...packageEvents]`)
  applies in month mode as well.
- All-day vs. timed classification and timezone-aware day bucketing reuse the existing
  `getEventsForDay` / `getDateInZone` utilities (extended as needed to bucket events across
  a whole month rather than a fixed week).

## 5. API Surface

No new or modified endpoints. The feature reuses the existing read path:

- `GET` calendar events over a date range (consumed via
  `useCalendarEvents(startISO, endISO)`), called with a month-grid-sized range.
- Existing connections, sources, and package queries are unchanged.

If, during design/implementation, the events endpoint is found to cap or paginate results
in a way that a ~6-week range could exceed, that is called out as a risk (see §9) — but no
contract change is planned as part of this task.

## 6. Data Model

No data-model changes. No new entities, fields, migrations, or `tenant_id`-scoped tables.
Month view is a pure rendering/navigation layer over events that already carry their
owner's `userDisplayName`/`userColor`, `allDay`, `startTime`, `endTime`, and
`connectionId` attributes.

## 7. Service Impact

| Service | Change |
|---------|--------|
| `frontend` | New month-grid components and view-mode toggle on the calendar page; widen the events query range in month mode; reuse existing color/legend/package/date-picker logic. |
| `calendar-service` | **None anticipated.** Already serves arbitrary date ranges. |
| All other services | None. |

Anticipated frontend touch points (subject to the design phase):
- `frontend/src/pages/CalendarPage.tsx` — view-mode state, toggle, month-aware navigation
  and date-range label, conditional rendering of week vs. month grid, day-click handler that
  switches back to week mode.
- `frontend/src/components/features/calendar/` — new month-grid component(s) (e.g., a
  `month-grid.tsx` and day-cell/chip/dot subcomponents), plus a view-mode toggle component.
- `frontend/src/components/features/calendar/calendar-utils.ts` — new helpers for month-grid
  date math (visible-grid bounds, per-day bucketing across a month), reusing
  `getStartOfWeek`, `getDateInZone`, `isToday`, `getEventsForDay`.
- Tests under `frontend/src/components/features/calendar/__tests__/`.

## 8. Non-Functional Requirements

- **Performance:** Rendering up to ~42 day cells with their events must remain smooth.
  Per-day event bucketing should be computed once per render pass (e.g., memoized), not
  recomputed per cell, to avoid O(days × events) churn on every interaction.
- **Multi-tenancy:** No change to tenant scoping; events are already household-scoped by the
  existing query path and tenant context. Month view introduces no cross-tenant data access.
- **Timezone correctness:** Day bucketing and "today" highlighting must use the household
  timezone via the existing `getDateInZone`/`isToday` utilities, so an event near midnight
  lands in the correct day cell — the same correctness guarantee the week view provides.
- **Responsiveness:** Month grid must be legible from mobile (dot mode) through desktop
  (chip mode), switching presentation at the existing 768px breakpoint.
- **Accessibility:** Day cells and chips are keyboard-focusable and activatable; the
  view-mode toggle exposes its state to assistive tech; color is not the *only* signal where
  feasible (titles/labels accompany color in chip mode).
- **No regression:** The week/3-day view, event CRUD, re-auth banner, connection status,
  source selection, and package overlay behave exactly as before when in week mode.

## 9. Open Questions

- **Mobile dot density:** For a day with many events, do we cap the number of dots shown
  (e.g., show up to N dots) or render one per event? (Leaning: cap with a small overflow
  indicator dot; final call in design.)
- **Timed-chip time format:** Exact compact time presentation on desktop chips (e.g.,
  "9a Standup" vs. a leading dot + title). Design-phase detail.
- **Events endpoint range ceiling:** Confirm the events endpoint imposes no result cap that
  a ~6-week range could hit. If it does, decide whether to chunk requests (no contract change)
  — verify against `calendar-service` source during design.
- **First-of-row month label:** Whether to show the month name in the first cell of a new
  month within the grid (a common nicety) or rely solely on the header label. Design-phase.

## 10. Acceptance Criteria

- [ ] A Week/Month toggle appears in the calendar page header and switches between the
      existing time-grid view and the new month grid.
- [ ] The page loads in Week mode by default on every visit; the mode is not persisted to
      localStorage or the URL.
- [ ] Month mode renders a 7-column grid of complete weeks (5–6 rows) covering the visible
      month, with leading/trailing adjacent-month days muted but populated and clickable.
- [ ] Each day cell shows the day-of-month number; today's cell is highlighted using the
      same treatment as the week view.
- [ ] On desktop/tablet, each cell lists its events as color-coded chips (by household
      member), including package-overlay events; all-day and timed events are distinguishable.
- [ ] When a day has more events than fit, the day cell scrolls internally to reveal them
      all — no "+N more" popover, no silent truncation.
- [ ] On mobile, day cells render compact colored dots instead of text chips, and the
      toggle is available.
- [ ] Clicking/tapping a day cell (including adjacent-month days) switches to Week mode
      focused on that day (7-day grid on desktop, 3-day window on mobile). Clicking an event
      chip drills into the day the same way.
- [ ] In month mode, prev/next navigate by one calendar month and the header label shows
      "Month YYYY"; the date picker jumps to a selected month without switching to week mode.
- [ ] Day bucketing and "today" detection respect the household timezone.
- [ ] No backend/schema changes are introduced; only the events query range widens in month mode.
- [ ] The week/3-day view, event CRUD, color legend, connection status, re-auth banner,
      source selection, and package overlay are unchanged in week mode.
- [ ] Frontend builds and all affected frontend tests pass; new month-grid logic has tests
      (date-math/bucketing utilities and key interactions).
