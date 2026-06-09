# Calendar Month View — Context

Companion to `plan.md`. Key files, decisions, and dependencies an implementer needs before starting. Source of truth is the code; this summarizes what was verified during planning.

## Scope

Frontend-only addition: a **Month** view mode for the calendar page, toggleable with the existing Week/3-day time-grid view. Week stays the default and is functionally unchanged. **No backend, schema, or endpoint changes.** PRD: `prd.md`. Design: `design.md`.

## Key existing files

| File | Why it matters |
|---|---|
| `frontend/src/pages/CalendarPage.tsx` | Owns all calendar state (week start, queries, CRUD dialogs). Gains `viewMode` + `monthAnchor`; only the grid slot and header label/nav branch on mode. The week path must remain byte-for-byte unchanged. |
| `frontend/src/components/features/calendar/calendar-utils.ts` | Pure date/event helpers. New month helpers are added here; bucketing **reuses** the existing tz-aware `getEventsForDay`, `getDateInZone`, `getTimeInZone`, `getStartOfWeek`, `isToday`. |
| `frontend/src/components/features/calendar/calendar-grid.tsx` | The existing Week grid. **Not modified.** Source of the styling idioms the month grid mirrors: today = `bg-primary/10` cell + `text-primary` number (lines 80–89); frame = `border rounded-lg overflow-hidden bg-background`; weekday header = `text-xs text-muted-foreground uppercase`. |
| `frontend/src/components/features/calendar/event-block.tsx` | Source of the "Busy" gray treatment: `title === "Busy" && !isOwner` → `#9ca3af`, else `userColor` (line 19, 31). Month chips/dots replicate this. |
| `frontend/src/lib/hooks/api/use-calendar.ts` | `useCalendarEvents(start, end)` (line ~42) — range-driven, React-Query cached by `[tenant, household, start, end]`, `staleTime: 60s`. Month mode just passes a wider range; no new hook. |
| `frontend/src/components/features/packages/package-calendar-overlay.tsx` | `packagesToCalendarEvents` emits `allDay: true` events with a teal `userColor` (verified lines 27–29). They bucket as all-day in month cells for free via the existing merge `events = [...calendarEvents, ...packageEvents]` (`CalendarPage.tsx:106`). |
| `frontend/src/context/tenant-context.tsx` | `useTenant()` (line 123) exposes `household.attributes.timezone`. `MonthGrid` reads it the same way `CalendarGrid` does. |

## Verified facts (against source)

- **No result cap on the events endpoint.** `services/calendar-service/internal/event/provider.go:17` filters `start_time < ? AND end_time > ?`, orders `all_day DESC, start_time ASC`, **no `LIMIT`**. A ~6-week range returns every event → no chunking. (Resolves PRD §9 "endpoint range ceiling".)
- **Package overlay events are all-day** (see file table above) → they render in month cells without special casing.
- **Week-start convention is Sunday**, via `getStartOfWeek` (`calendar-utils.ts:64`) which uses `getDay()`/`setDate`. Month grid reuses it.
- **Reference dates** (used in tests): March 1, 2026 is a **Sunday** (the existing `calendar-utils.test.ts` asserts March 22/25 are Sun/Wed), so March 2026 → **35-cell** grid. August 1, 2026 is a **Saturday**, so August 2026 → **42-cell** grid with leading July / trailing September days. These give deterministic 5-row and 6-row coverage.

## Key decisions (from design)

1. **Two independent time anchors**, not one shared cursor. `weekStart` (existing) drives Week mode untouched; new `monthAnchor` (1st of focused month, local midnight) drives Month mode. Keeps each mode's date math isolated and the week path unchanged. (design §2.1)
2. **One events query, range switches on mode.** Week → `[weekStart, weekStart+dayCount)`; Month → `[gridStart, gridEnd)` from `getMonthGridRange`. Same `.toISOString()` → `useCalendarEvents`. React Query caches each distinct range. (design §2.3)
3. **The day cell is the single interactive element** (one `<button>` with an aria-label like "August 14, 3 events"). Chips and dots inside are presentational (`aria-hidden`/plain text). This is a deliberate, documented deviation from the PRD's literal "chips focusable" wording — nested interactive buttons are invalid HTML, and chip-click === day-click anyway. (design §2.5)
4. **Month membership comparison is local** (`isSameMonth`), not tz-shifted, so the muted-cell signal never disagrees with the locally-derived day number. (design §3)
5. **Bucket once per render**, memoized in `MonthGrid` on `[gridDays, events, timezone]`; cells read their bucket by `toDayKey` — O(1) per cell. (design §3 / PRD §8 perf)
6. **Resolved PRD open questions:** mobile dots capped at 4 (3 colored + "+" overflow when >4); timed-chip time format = compact `formatChipTime` ("9a"/"9:30a"/"2:30p"); no first-of-row month label (YAGNI). (design §6)

## Transition rules (CalendarPage handlers)

| Trigger | Effect |
|---|---|
| Toggle Week → Month | `setMonthAnchor(getStartOfMonth(weekStart))` |
| Toggle Month → Week | focus = today if `isSameMonth(today, monthAnchor)` else `monthAnchor`; `focusDay(focus)` |
| Day-cell / chip click | `focusDay(clickedDay)` + `setViewMode("week")` |
| Date picker (Month) | `setMonthAnchor(getStartOfMonth(date))` — does **not** switch to Week |
| Date picker (Week) | `focusDay(date)` (refactored to share with drill-down) |
| prev / next | Month: `monthAnchor ± 1 month` via `addMonths`; Week: `± dayCount` days (existing) |

`focusDay(day)`: desktop → `setWeekStart(getStartOfWeek(day))`; mobile → `setWeekStart(day − 1 day @ midnight)` (mirrors existing windowing at `CalendarPage.tsx:149-152`).

## New helpers added to `calendar-utils.ts`

`getStartOfMonth`, `addMonths`, `getMonthGridDays`, `getMonthGridRange`, `isSameMonth`, `formatMonthYear`, `formatChipTime`, `toDayKey`, `bucketEventsByDay`. All pure and unit-tested. Signatures and bodies are in `plan.md` Tasks 1–5.

## New components (`frontend/src/components/features/calendar/`)

- `view-mode-toggle.tsx` — `ViewModeToggle` (exports `CalendarViewMode` type). Tested.
- `month-event-chip.tsx` — `MonthEventChip` (presentational; tested via MonthGrid).
- `month-day-cell.tsx` — `MonthDayCell` + internal `MonthDayDots` (tested via MonthGrid).
- `month-grid.tsx` — `MonthGrid` (header + cells, memoized bucketing). Tested.

## Tooling / commands

- Package manager: **npm** (`frontend/package-lock.json`; CI runs `npm run build` in `.github/workflows/pr.yml`).
- Tests: `npm test` (= `vitest run`) from `frontend/`. Single file: `npm test -- <path>`.
- Build/type-check: `npm run build` (= `tsc -b && vite build`). Lint: `npm run lint`.
- Component tests mock `@/context/tenant-context` (`useTenant: () => ({ household: { attributes: { timezone: "UTC" } } })`) — pattern from `meal-plan-today-detail.test.tsx`.
- "Today" tests use `vi.useFakeTimers()` + `vi.setSystemTime(...)` to make `isToday` deterministic.

## Out of scope (PRD non-goals)

No backend/endpoint/schema change; no single-day or agenda view; no drag interactions; no create/edit from a month cell; no view-mode persistence (defaults to Week each load, not in URL/localStorage); no print/export; no recurrence-rule changes.

## Testing strategy summary

- **Unit** (`calendar-utils.test.ts`, extended): all new helpers — grid sizing (35/42), exclusive range end, local month membership incl. year boundary, compact tz-aware time format, bucketing (correct day, sort, multi-day all-day span, entry-per-grid-day).
- **Component** (`view-mode-toggle.test.tsx`, `month-grid.test.tsx`, new): toggle aria-pressed + onChange; 7 headers; 35/42 cells; today highlight (faked time); desktop chips with time prefix; dense-day no-truncation (all events in DOM); mobile dots + overflow; `onDayClick` fires correct Date for in-month and trailing cells.
- **Regression:** existing calendar tests stay green; Week mode renders the unchanged `CalendarGrid`. Full `npm test` + `npm run build` + `npm run lint` must pass.
