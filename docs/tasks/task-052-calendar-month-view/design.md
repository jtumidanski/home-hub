# Calendar Month View — Design

Version: v1
Status: Approved for planning
Created: 2026-06-09
PRD: `docs/tasks/task-052-calendar-month-view/prd.md`

---

## 1. Summary

Add a **Month** view mode to the calendar page alongside the existing Week/3-day
time-grid view. A Week/Month toggle in the header switches presentation; Week stays the
default and is functionally unchanged. Month mode renders a standard 7-column,
5–6-row grid of complete weeks, color-codes each day's events by household member, and
drills into the Week view when a day is clicked.

This is a **frontend-only** change. It reuses the existing `useCalendarEvents(start, end)`
read path with a wider date range, the existing package overlay merge, and the existing
timezone-aware day-bucketing utilities. **No backend, schema, or endpoint changes.**

### Verified assumptions (against source)

- **`useCalendarEvents(start, end)`** (`frontend/src/lib/hooks/api/use-calendar.ts:42`) is
  range-driven and React-Query-cached by `[tenant, household, start, end]`. Calling it with
  a month-grid range is sufficient; no new hook or endpoint.
- **No result cap.** `services/calendar-service/internal/event/provider.go:17` (`getByHousehold­AndTimeRange`)
  filters `start_time < ? AND end_time > ?` with `Order("all_day DESC, start_time ASC")` and
  **no `LIMIT`**. A ~6-week window returns every matching event; the PRD's "endpoint range
  ceiling" open question is resolved — **no chunking needed**.
- Provider already orders **all-day first, then by start time ascending**, so month cells can
  render events in arrival order without re-sorting (we still sort defensively in the bucket).

---

## 2. Architecture & Key Decisions

### 2.1 Where the mode lives — component state in `CalendarPage`

`CalendarPage.tsx` gains a `viewMode: "week" | "month"` state, defaulting to `"week"`.
Not persisted to localStorage, not reflected in the URL (per PRD 4.1). The existing
`searchParams` usage is solely for the OAuth `connected`/`error` flash and is left alone.

**Decision:** keep two distinct time anchors rather than one shared cursor.

- `weekStart: Date` — **existing** state; the Sunday (desktop) / today−1 (mobile) window
  start. Sole driver of Week mode. Untouched semantics.
- `monthAnchor: Date` — **new** state; the **1st of the focused month at local midnight**.
  Sole driver of Month mode. Initialized to `getStartOfMonth(new Date())`.

**Alternatives considered:**
- *Single shared `focusDate` + derive both views.* Rejected: Week mode's anchor has bespoke
  mobile windowing (`today − 1 day`, see `CalendarPage.tsx:60-66, 149-152`) that doesn't map
  cleanly onto a month-aligned cursor; sharing one Date forces lossy conversions on every
  toggle and muddies the existing, tested week logic.
- *Store the focused month as a `weekStart` derivative.* Rejected: `weekStart` can straddle two
  months, so "which month am I viewing" becomes ambiguous. An explicit `monthAnchor` is
  unambiguous and trivially testable.

Two small, independent anchors keep each mode's date math isolated and leave the week path
byte-for-byte unchanged — satisfying the PRD's "no regression" requirement structurally, not
just by testing.

### 2.2 Transitions between modes (preserve position in time)

| Trigger | Effect |
|---|---|
| Toggle **Week → Month** | `setMonthAnchor(getStartOfMonth(weekStart))`, `setViewMode("month")`. (Month containing the focused week start, per PRD 4.1.) |
| Toggle **Month → Week** | Focus day = **today if today is within `monthAnchor`'s month, else the 1st of that month**; then apply the existing week-windowing (`focusDay`, below) and `setViewMode("week")`. |
| **Day-cell / chip click** (PRD 4.5) | `focusDay(clickedDay)` + `setViewMode("week")`. Distinct from the plain toggle. |
| **Date picker** in Month mode | `setMonthAnchor(getStartOfMonth(date))`. Does **not** switch to Week (PRD 4.6). |
| **Date picker** in Week mode | Existing `handleCalendarSelect` behavior, unchanged. |
| **prev / next** | Month mode: `monthAnchor ± 1 calendar month`. Week mode: existing `± dayCount` days. |

`focusDay(day: Date)` centralizes the "land on this day in the week grid" rule, reusing the
existing windowing exactly:
- **Desktop:** `setWeekStart(getStartOfWeek(day))` → day appears in the 7-day grid.
- **Mobile:** `setWeekStart(day − 1 day @ local midnight)` → day appears in the 3-day window,
  matching the current convention at `CalendarPage.tsx:149-152`.

`handleCalendarSelect` is refactored to call `focusDay` (Week mode) so the drill-down and the
date-picker share one implementation.

### 2.3 Data fetching — one query, range switches on mode

Keep the single `useCalendarEvents(startISO, endISO)` call. Compute the range from `viewMode`:

```
weekMode:  [weekStart, weekStart + dayCount)          // existing
monthMode: [gridStart, gridEnd)                        // visible grid bounds
           gridStart = getStartOfWeek(getStartOfMonth(monthAnchor))  // leading Sunday
           gridEnd   = gridStart + (rows * 7) days     // exclusive; Saturday + 1
```

Both feed the same `.toISOString()` → `startISO/endISO` → `useCalendarEvents`. React Query
caches each distinct range, so flipping back and forth between adjacent months/weeks is cheap.
The package query and the `events = [...calendarEvents, ...packageEvents]` merge
(`CalendarPage.tsx:99-106`) apply unchanged in both modes — package-overlay events appear in
month cells for free because they are already `CalendarEvent`s with a teal `userColor`.

### 2.4 Component breakdown

```
CalendarPage (owns viewMode, weekStart, monthAnchor, range, all CRUD/dialog state)
├─ ViewModeToggle          (header: Week | Month segmented control)
├─ [week mode] CalendarGrid   ← existing, untouched
└─ [month mode] MonthGrid
   ├─ weekday header row (SUN…SAT)
   └─ MonthDayCell × (35 or 42)
      ├─ day number + today/muted styling
      ├─ [desktop] MonthEventChip × N  (internally scrollable list)
      └─ [mobile]  colored dots (+ overflow indicator)
```

New files (all under `frontend/src/components/features/calendar/`):
- `month-grid.tsx` — `MonthGrid`
- `month-day-cell.tsx` — `MonthDayCell`
- `month-event-chip.tsx` — `MonthEventChip`
- `view-mode-toggle.tsx` — `ViewModeToggle`
- date-math + bucketing helpers **added to** existing `calendar-utils.ts`

`MonthGrid` props:
```ts
interface MonthGridProps {
  monthAnchor: Date;
  events: CalendarEvent[];
  isDesktop: boolean;
  onDayClick: (day: Date) => void;
}
```
It reads `household.attributes.timezone` from `useTenant()` (same as `CalendarGrid`), computes
the grid days and the per-day event buckets **once** (memoized), and renders the header + cells.

### 2.5 The day cell is the single interactive element (accessibility decision)

`MonthDayCell` renders as **one** keyboard-focusable, activatable control (a `<button>` /
`role="button"` with an accessible label like `"June 14, 3 events"`). Chips and dots inside it
are **presentational** (`aria-hidden` decorative content / plain text), not nested buttons.

**Rationale & deviation note:** PRD 8 says "day cells and chips are keyboard-focusable and
activatable," but nesting interactive `<button>`s inside an interactive cell is invalid HTML and
a known a11y anti-pattern. Since "clicking a chip behaves the same as clicking its day cell"
(PRD 4.5) and month chips never open a popover or edit dialog, a single cell-level activation
target fully satisfies the behavior while staying accessible. We therefore make the **cell** the
focus/activation target; chip text remains a visible, screen-reader-readable label. This is a
deliberate, behavior-preserving deviation from the literal wording.

---

## 3. Date Math & Bucketing (additions to `calendar-utils.ts`)

All new helpers are pure and unit-tested. They reuse the existing `getStartOfWeek`,
`getDateInZone`, `isToday`, and `getEventsForDay`.

- `getStartOfMonth(date: Date): Date` — 1st of `date`'s month at local midnight.
- `getMonthGridDays(monthAnchor: Date): Date[]` — the visible grid. Start at
  `getStartOfWeek(getStartOfMonth(monthAnchor))`; append days week-by-week until the grid both
  (a) covers the month's last day and (b) completes the final week → length **35 or 42**.
  Weeks start **Sunday**, matching the week view.
- `getMonthGridRange(monthAnchor): { start: Date; end: Date }` — `start = days[0]`,
  `end = days[last] + 1 day` (exclusive), for the events query.
- `isSameMonth(day: Date, monthAnchor: Date): boolean` — **local** month/year comparison, used
  to mute leading/trailing adjacent-month cells. Local (not tz-shifted) because the cell's
  day-number and grid membership are built from local `Date`s exactly like the week view's
  `day.getDate()` (`CalendarGrid` line 88) — keeping it local avoids the day-number and the
  muting disagreeing near midnight.
- `formatMonthYear(date: Date): string` — `"June 2026"` via
  `toLocaleDateString("en-US", { month: "long", year: "numeric" })`. Used for the header label
  in month mode.
- `formatChipTime(iso: string, timezone?: string): string` — compact timed-event prefix using
  the tz-aware `getTimeInZone`: drops `:00` minutes and uses single-letter meridiem, e.g.
  `"9a"`, `"9:30a"`, `"12p"`, `"2:30p"`. All-day events render with **no** time prefix.
- `bucketEventsByDay(gridDays, events, timezone): Map<string, { allDay, timed }>` — keyed by
  `YYYY-MM-DD` (local), one entry per grid day. Implemented by calling the **existing**
  `getEventsForDay` per grid day, so multi-day events, all-day date-string comparison, and
  tz bucketing all reuse already-tested logic. Within each bucket, `timed` is sorted by start
  ascending; cells render `allDay` first, then `timed` (consistent with provider ordering).

**Performance (PRD 8):** `bucketEventsByDay` runs **once per render**, memoized in `MonthGrid`
on `[gridDays, events, timezone]`. Cells read their bucket by key — O(1) per cell, no
O(days × events) churn on interaction. With ≤42 days and typical household event counts this is
negligible.

---

## 4. Rendering & Styling

### 4.1 Grid container

`MonthGrid` is the `flex-1 min-h-0` child (same slot as `CalendarGrid` today,
`CalendarPage.tsx:324-339`). Internally a CSS grid:
- columns: `grid-template-columns: repeat(7, minmax(0, 1fr))`
- rows: a fixed weekday-header row (`auto`) + `repeat(rows, minmax(0, 1fr))` so the 5–6 week
  rows split the available height evenly and the whole grid fills the page like the week view.
- `border rounded-lg overflow-hidden bg-background` to match `CalendarGrid`'s frame.

Weekday header: `SUN…SAT` via `toLocaleDateString("en-US", { weekday: "short" }).toUpperCase()`,
reusing the week view's `text-xs text-muted-foreground uppercase` treatment.

### 4.2 Day cell

- Day-of-month number top-aligned. **Today** highlighted with the week view's primary treatment
  (`bg-primary/10` cell tint + `text-primary` number, mirroring `CalendarGrid` lines 80-89),
  using tz-aware `isToday`.
- **Leading/trailing** cells (`!isSameMonth`) get muted styling (`text-muted-foreground`, subtle
  `bg-muted/30`) but still render events and remain clickable (PRD 4.2).
- Cell is `min-h-0 overflow-hidden flex flex-col`; the event list area is
  `flex-1 overflow-y-auto` → **internal scroll** when a day has more events than fit. **No
  "+N more" popover, no truncation** (PRD 4.3) — every event is reachable by scrolling the cell.

### 4.3 Desktop/tablet — chips (`MonthEventChip`)

Per event: a compact chip, full-width within the cell, `backgroundColor = userColor`, white
text, single line truncated. Timed events show `formatChipTime(startTime, tz)` as a small leading
label before the title; all-day events show title only — so the two are visually distinguishable
(PRD 4.3). "Busy" events from other members keep the week view's gray treatment
(`#9ca3af` when `title === "Busy" && !isOwner`, mirroring `EventBlock`/`AllDayEventRow`).
Chips are presentational (see 2.5); clicking anywhere in the cell drills into the day.

### 4.4 Mobile — dots

Below `md` (768px, the existing `MD_BREAKPOINT`), cells render **colored dots** instead of chips
(PRD 4.4). One dot per event in render order (all-day then timed), `backgroundColor = userColor`.
**Density cap:** show up to **4** dots; if a day has more than 4 events, show the first 3 colored
dots plus a muted **"+" overflow dot** (resolving PRD open question §9 "mobile dot density":
cap with an overflow indicator). Dots are presentational; tapping the cell drills into the day
(no per-dot tap target, per PRD 4.4).

`isDesktop` is already available in `CalendarPage` via `useSyncExternalStore` and is passed into
`MonthGrid` → `MonthDayCell` to choose chips vs dots. No new media-query plumbing.

---

## 5. Header & Navigation Wiring (`CalendarPage`)

- **Toggle:** `ViewModeToggle` (segmented Week | Month, `aria-pressed` per option) placed in the
  header control cluster next to prev/Today/next. Visible on mobile and desktop. "Add Event"
  (when `hasWriteAccess`) and the date-picker popover remain visible in both modes (PRD 4.1).
- **Label:** month mode shows `formatMonthYear(monthAnchor)` (e.g. "June 2026"); week mode keeps
  `formatDateRange(weekStart, weekEnd)` (`CalendarPage.tsx:244`).
- **prev/next** (`goPrev`/`goNext`) branch on `viewMode`: month → `monthAnchor ± 1 month`
  (via a `addMonths`-style helper using `setMonth`), week → existing `± dayCount`.
- **Date picker:** the `Calendar` popover's `selected`/`defaultMonth` reflect `monthAnchor` in
  month mode and `weekStart` in week mode; `onSelect` routes to the table in §2.2. The picker
  trigger continues to act as the "return to current" affordance (PRD 4.6).

The OAuth-flash effect, re-auth banner, connection status rows, household color legend, source
selection panel, and all event CRUD dialogs are **untouched** and render identically in both
modes (only the grid slot swaps).

---

## 6. Resolved PRD Open Questions (§9)

| Open question | Resolution |
|---|---|
| Mobile dot density | Cap at 4: up to 4 colored dots; >4 → 3 dots + muted "+" overflow dot. |
| Timed-chip time format | `formatChipTime` → compact `"9a"` / `"9:30a"` / `"2:30p"` leading label; all-day = no time. |
| Events endpoint range ceiling | **No cap exists** (provider has no `LIMIT`); single range query, no chunking. |
| First-of-row month label | **Omitted** (YAGNI). The header `formatMonthYear` label is the sole month orientation; muted adjacent-month cells already signal month boundaries. |

---

## 7. Testing Strategy

**Unit (`__tests__/calendar-utils.test.ts`, extend existing):**
- `getStartOfMonth` — 1st at midnight, arbitrary day.
- `getMonthGridDays` — 35-row month (e.g. a month starting Sunday) vs 42-row month; first day is
  the leading Sunday; last day is a Saturday; leading/trailing days belong to adjacent months.
- `getMonthGridRange` — exclusive end = last grid day + 1.
- `isSameMonth` — in-month true, leading/trailing false, year boundary (Dec/Jan).
- `formatMonthYear` — "June 2026".
- `formatChipTime` — on-the-hour drops minutes, half-hour keeps them, AM/PM, tz-aware via
  `getTimeInZone`.
- `bucketEventsByDay` — event lands in correct day key; tz bucketing near midnight; multi-day
  all-day event spans multiple keys; package overlay event buckets as all-day; timed sorted by
  start.

**Component (`__tests__/month-grid.test.tsx`, new):**
- Renders 7 weekday headers and 35 or 42 cells for representative months.
- Today cell highlighted; adjacent-month cells muted but populated.
- Desktop: events render as chips with time prefix for timed / none for all-day.
- Mobile (`isDesktop=false`): dots render; >4 events → 3 dots + overflow.
- Internal scroll: a dense day's cell is scrollable (overflow class / all events present in DOM,
  none dropped).
- `onDayClick` fires with the correct `Date` for an in-month cell and for a leading/trailing
  adjacent-month cell.

**Interaction (`__tests__/calendar-page` focused, or via MonthGrid + handler unit):**
- Toggle Week↔Month swaps the rendered grid and preserves position (Week→Month shows the month
  of `weekStart`; Month→Week focuses today-if-in-month-else-1st).
- Month-mode prev/next moves by one month and updates the "Month YYYY" label.
- Day click switches to Week mode focused on that day (desktop 7-day contains the day; mobile
  window start = day − 1).
- Date-picker selection in Month mode changes `monthAnchor` without leaving Month mode.

**Regression guard:** existing `calendar-utils.test.ts` and week-view behavior remain green;
Week mode renders the unchanged `CalendarGrid`.

`pnpm` build + the affected frontend tests must pass (PRD acceptance §10, final bullet).

---

## 8. File Manifest

| File | Change |
|---|---|
| `frontend/src/pages/CalendarPage.tsx` | Add `viewMode`/`monthAnchor` state; range branch; `focusDay`/`handleMonthSelect` helpers; mode-aware `goPrev`/`goNext`, label, date-picker; render `ViewModeToggle` + conditional `MonthGrid`/`CalendarGrid`. |
| `frontend/src/components/features/calendar/calendar-utils.ts` | Add `getStartOfMonth`, `getMonthGridDays`, `getMonthGridRange`, `isSameMonth`, `formatMonthYear`, `formatChipTime`, `bucketEventsByDay` (+ an `addMonths` helper). |
| `frontend/src/components/features/calendar/month-grid.tsx` | **New** `MonthGrid`. |
| `frontend/src/components/features/calendar/month-day-cell.tsx` | **New** `MonthDayCell`. |
| `frontend/src/components/features/calendar/month-event-chip.tsx` | **New** `MonthEventChip`. |
| `frontend/src/components/features/calendar/view-mode-toggle.tsx` | **New** `ViewModeToggle`. |
| `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts` | Extend with new helper tests. |
| `frontend/src/components/features/calendar/__tests__/month-grid.test.tsx` | **New** component tests. |

No changes to `calendar-service`, `useCalendarEvents`, the package overlay, or any shared type.

---

## 9. Risks & Mitigations

- **Toggle re-render cost / event churn.** Mitigated by memoizing `bucketEventsByDay` once per
  render and keying cells by date — O(1) cell reads.
- **Timezone edge near midnight.** Mitigated by reusing `getEventsForDay`/`getDateInZone` for
  bucketing and `isToday` for highlighting — the same guarantees the week view already ships.
- **Local vs tz month membership.** Deliberately local (matches the locally-derived day number);
  documented in §3 to avoid a future "why isn't this tz-aware" regression.
- **Accessibility deviation** (cell-level activation vs literal "chips focusable"). Documented in
  §2.5 with rationale; the behavior contract (chip click === day click) is fully met.

## 10. Out of Scope (per PRD non-goals)

No backend/endpoint/schema change; no single-day or agenda view; no drag interactions; no
create/edit from a month cell; no view-mode persistence; no print/export; no recurrence changes.
