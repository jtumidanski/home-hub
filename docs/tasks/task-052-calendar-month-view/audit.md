# Audit — task-052-calendar-month-view

## Plan Adherence Audit

**Plan Path:** docs/tasks/task-052-calendar-month-view/plan.md
**Audit Date:** 2026-06-09
**Branch:** task-052-calendar-month-view
**Base (plan) commit:** 712180a
**Head (final) commit:** 396513d

### Executive Summary

All 11 plan tasks were faithfully implemented. The 11 planned feature commits (23693bc → 396513d) are all present, plus one additional in-scope fix commit (8a5b8c0) that made timezone bucketing DST-correct — a deviation that is documented in the task brief and is an improvement, not a gap. The full frontend test suite passes (661/661), `npm run build` is clean (tsc + vite), and `npm run lint` reports zero errors in any file this branch touched (all 9 lint errors are pre-existing, in unrelated files). The Week-view code path is functionally unchanged. **Verdict: READY_TO_MERGE.**

### Task Completion

| # | Task | Status | Evidence |
|---|------|--------|----------|
| 1 | `getStartOfMonth` + `addMonths` | IMPLEMENTED | calendar-utils.ts:72-86; tests calendar-utils.test.ts:299, :318; commit 23693bc |
| 2 | `getMonthGridDays` + `getMonthGridRange` | IMPLEMENTED | calendar-utils.ts:93-118; tests :339, :375; commit d5be891 |
| 3 | `isSameMonth` + `formatMonthYear` | IMPLEMENTED | calendar-utils.ts:124-130; tests :386, :406; commit dd9b424 |
| 4 | `formatChipTime` | IMPLEMENTED | calendar-utils.ts:252-257; tests :412; commit 650fd8f |
| 5 | `toDayKey` + `bucketEventsByDay` | IMPLEMENTED | calendar-utils.ts:329-353; tests :439, :445; commit 37c0c28 |
| 6 | `ViewModeToggle` component | IMPLEMENTED | view-mode-toggle.tsx:1-33; test view-mode-toggle.test.tsx; commit e9b024f |
| 7 | `MonthEventChip` component | IMPLEMENTED | month-event-chip.tsx:1-28; commit 61b7189 |
| 8 | `MonthDayCell` component | IMPLEMENTED | month-day-cell.tsx:1-84 (incl. `MonthDayDots`, MAX_DOTS=4 overflow); commit 2f3fdbe |
| 9 | `MonthGrid` component (+ test) | IMPLEMENTED | month-grid.tsx:1-70; test month-grid.test.tsx; commits bae30f0, 8a5b8c0 |
| 10 | `CalendarPage` wiring | IMPLEMENTED | CalendarPage.tsx:50-51, :107-113, :150-204, :294, :299, :314-316, :380-386; commit 396513d |
| 11 | Full verification (tests/build/lint) | IMPLEMENTED | See build/test results below |

**Completion Rate:** 11/11 (100%)
**Skipped without approval:** 0
**Partial implementations:** 0

### Per-task detail

- **Tasks 1–5 (helpers):** All nine pure helpers exist with bodies matching the plan's specified implementations. All corresponding `describe` blocks are present in `calendar-utils.test.ts`.
- **Task 6 `ViewModeToggle`:** Renders Week|Month buttons with `aria-pressed`, exports `CalendarViewMode`, calls `onChange`. Matches plan verbatim.
- **Task 7 `MonthEventChip`:** Presentational, non-interactive; replicates the "Busy" gray (`#9ca3af` when `title === "Busy" && !isOwner`) else `userColor`; drops time for all-day.
- **Task 8 `MonthDayCell`:** Single `<button>` is the only interactive element (design §2.5); chips on desktop, dots on mobile capped at 4 with `data-testid="dot-overflow"`; today/adjacent-month tones match `calendar-grid` idioms.
- **Task 9 `MonthGrid`:** 7 weekday headers, memoized `getMonthGridDays` + `bucketEventsByDay` on `[gridDays, events, timezone]`, `isToday`/`isSameMonth` per cell. Test file present and green.
- **Task 10 `CalendarPage`:** `viewMode`/`monthAnchor` state (:50-51); mode-aware range memo (:107-113); `focusDay` extracted + mode-aware `goPrev`/`goNext`/`handleCalendarSelect`/`handleDayClick`/`handleViewModeChange` (:150-204); mode-aware header label (:294), date picker `selected`/`defaultMonth` (:314-316), and conditional `MonthGrid` vs `CalendarGrid` render (:380-386).

### Additional in-scope DST fix (commit 8a5b8c0)

The brief flagged that during Task 9 an extra fix was made to `getEventsForDay`/`midnightInZone`. **Confirmed.** At base commit 712180a, `getEventsForDay` computed the day window with `new Date(year, month-1, d)` (host-local midnight, not tz-aware). The branch introduces `midnightInZone` (calendar-utils.ts:269-295), which derives the correct UTC instant for midnight in the household timezone using a measure-and-correct loop (two passes to handle DST transitions), and `getEventsForDay` now uses it (calendar-utils.ts:299). This corrects the plan's incorrect assumption that `getEventsForDay` was already fully tz-aware.

- **Did not break the Week path:** `getEventsForDay` is consumed by `CalendarGrid` (Week) as well. The full suite (661 tests), including the existing calendar-utils and calendar-grid tests, passes. The fallback branch (`if (!timezone)` → local midnight) preserves back-compat for tests passing tz-less ISO strings.

### Week path unchanged (plan requirement)

Functionally preserved. Base-vs-head comparison of CalendarPage:
- Week branch of the range memo still computes `weekStart.toISOString()` / `weekEnd.toISOString()`.
- `goPrev`/`goNext` Week branches still shift by `± dayCount` days.
- The previously-inline desktop/mobile windowing in `handleCalendarSelect` (base lines 147/152) is extracted verbatim into the new `focusDay` callback and reused — same behavior.
- `CalendarGrid` render props (`weekStart`, `events`, `dayCount`, `hasWriteAccess`, `onSlotClick`, `onEditEvent`, `onDeleteEvent`) are unchanged.

### Build & Test Results

| Gate | Command | Result | Notes |
|------|---------|--------|-------|
| Tests | `node_modules/.bin/vitest run` (Linux node v22) | PASS | 98 files, 661/661 tests. `npm test` shim avoided per brief. |
| Build | `npm run build` | PASS | `tsc -b` clean; `vite build` succeeds. Chunk-size >500kB is a pre-existing informational warning, not an error. |
| Lint | `npm run lint` | PASS (this branch) | 9 errors total, ALL in untouched files: use-cooklang-preview.ts, DashboardDesigner.tsx, WorkoutReviewPage.test.tsx, event-form-dialog.tsx, new-dashboard-modal.tsx, tracker/calendar-grid.tsx, recurrence.ts. None are in the 9 files this branch modified. Pre-existing, not this branch's responsibility. |

### Overall Assessment

- **Plan Adherence:** FULL
- **Recommendation:** READY_TO_MERGE

### Action Items

None required for this branch. (Optional, out-of-scope: the 9 pre-existing lint errors in unrelated files could be addressed in a separate cleanup task.)

---

## Frontend Guidelines Audit (FE-* checklist)

- **Audit Scope:** `git diff 712180a 396513d -- 'frontend/**'` (Month view feature)
- **Guidelines Source:** frontend-dev-guidelines skill
- **Date:** 2026-06-09
- **Build:** PASS (`tsc -b` exit 0, `vite build` exit 0)
- **Tests:** 86 passed, 0 failed (5 calendar test files, run on Linux node v22)
- **Overall:** NEEDS-WORK — build + tests green; one blocking FE-02 style violation in new files.

### File Inventory

- `components/features/calendar/calendar-utils.ts` — Other (pure date/event helpers)
- `components/features/calendar/view-mode-toggle.tsx` — Component (presentational)
- `components/features/calendar/month-event-chip.tsx` — Component (presentational)
- `components/features/calendar/month-day-cell.tsx` — Component (presentational)
- `components/features/calendar/month-grid.tsx` — Component (feature container, reads `useTenant`)
- `pages/CalendarPage.tsx` — Page
- `__tests__/{calendar-utils,month-grid,view-mode-toggle}.{test.ts,test.tsx}` — Tests

### Anti-Pattern Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-01 | No `any` type | PASS | grep `: any` / `as any` over all 6 source files → 0 matches |
| FE-02 | No manual class concatenation (`cn()` required) | FAIL | `month-day-cell.tsx:70` and `:72` use template-literal `className={\`...${cellTone}\`}` / `${numberTone}`; `view-mode-toggle.tsx:17` and `:26` use bare ternary `className={mode === "week" ? "..." : ""}`. None use `cn()`. |
| FE-03 | No direct API client calls in components | PASS | No `@/lib/api/client` import in any changed file; `CalendarPage.tsx:20` uses `use-calendar` hooks |
| FE-04 | No inline Zod schemas | PASS | No `z.object/z.string` in scope (no forms added) |
| FE-05 | No spinners for content loading | PASS | Loading uses `<Skeleton>` (`CalendarPage.tsx:273-274,378`); no `animate-spin` in changed files |
| FE-06 | No hardcoded colors | PASS-with-note | `month-event-chip.tsx:20` uses `text-white`; `#9ca3af` busy fallback at `month-event-chip.tsx:21` & `month-day-cell.tsx:23`. Mirrors pre-existing week-view convention (`event-block.tsx:31`, `all-day-event-row.tsx:22-23`); chip backgrounds are dynamic per-user `userColor` that cannot use semantic tokens. (Minor) |
| FE-07 | No state mutation | PASS | New arrays/Maps built; sorts on copies (`calendar-utils.ts:347` `[...timed].sort`); `CalendarPage` uses functional `setMonthAnchor((a) => addMonths(...))` |
| FE-08 | No default exports for components | PASS | grep `export default` → 0 matches; all named exports |
| FE-09 | Tenant guard in hooks | PASS (N/A) | No new hooks. `month-grid.tsx:23` reads `useTenant()` for `household.timezone`; reused `useCalendarEvents` (`use-calendar.ts:42-48`, unchanged) has `enabled: !!tenant?.id && !!household?.id` |
| FE-10 | Tenant ID in query keys | PASS (N/A) | Events query reuses `calendarKeys.events(...)` with `tenant?.id` + `household?.id` (`use-calendar.ts:17-18`); no new keys |
| FE-11 | `createErrorFromUnknown` in catches | PASS (N/A) | No new async catch blocks in changed files |

### Architecture Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-12 | JSON:API model shape | PASS | `CalendarEvent` consumed as `{ id, type, attributes }` (`types/models/calendar.ts:53-57`); helpers read `event.attributes.*` only |
| FE-13 | Service extends BaseService | PASS (N/A) | No service files changed |
| FE-14 | Query key factory `as const` | PASS (N/A) | Reuses existing `calendarKeys` (`use-calendar.ts:10-18`, all `as const`) |
| FE-15 | Forms use RHF + zodResolver | PASS (N/A) | No forms added |
| FE-16 | Schema in `lib/schemas/` with inferred type | PASS (N/A) | No schemas added |
| FE-19 | Interactive elements show `cursor-pointer` | PASS | Single interactive surface is the day-cell `<button>` (`month-day-cell.tsx:66-71`) which includes `cursor-pointer`; chips/dots are presentational (`aria-hidden`, `month-day-cell.tsx:30`). `ViewModeToggle` uses native shadcn `<Button>`. |

### Testing Checklist

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| FE-17 | Tests exist for changed components | PASS | `calendar-utils.test.ts` (helpers incl. DST/tz cases :480-496), `month-grid.test.tsx`, `view-mode-toggle.test.tsx`. `MonthDayCell`/`MonthEventChip` covered via `MonthGrid` integration. `CalendarPage` mode-switching has no direct test (Minor). |
| FE-18 | Mocks updated when services changed | PASS (N/A) | No services changed; `month-grid.test.tsx:7-9` mocks `useTenant` |

### FE-02 detail and recommended fix

The entire `components/features/calendar/` directory does **not** import `cn()` — the pre-existing week view (`calendar-grid.tsx:80,87,138`) already interpolates conditional classes via template literals. The new month-view files follow that local precedent, but the guideline is absolute ("Never concatenate class strings manually — always use `cn()`", patterns-styling.md:31 / anti-patterns.md #4), and template-literal interpolation does no tailwind-merge dedup. Because these are **new** lines, they count as FE-02 failures.

- `month-day-cell.tsx:70` → `cn("flex flex-col min-h-0 overflow-hidden border-r border-b last:border-r-0 text-left p-1 cursor-pointer hover:bg-accent/40", cellTone)`
- `month-day-cell.tsx:72` → `cn("text-xs font-medium px-0.5", numberTone)`
- `view-mode-toggle.tsx:17,26` → `cn(mode === "week" && "bg-accent text-accent-foreground")`

### Summary

**Blocking (must fix):**
- **FE-02** — Manual class concatenation instead of `cn()`: `month-day-cell.tsx:70`, `month-day-cell.tsx:72`, `view-mode-toggle.tsx:17`, `view-mode-toggle.tsx:26`.

**Non-Blocking (should fix):**
- **FE-06 (Minor)** — `text-white` + hardcoded `#9ca3af` busy fallback in `month-event-chip.tsx:20-21`, `month-day-cell.tsx:23`. Matches existing convention; consider a shared constant for the busy color.
- **FE-17 (Minor)** — No direct test for `CalendarPage` mode-switching handlers (`handleViewModeChange`, month branches of `goPrev/goNext`/`handleCalendarSelect`).

**Strengths:**
- Single reused `useCalendarEvents(start, end)`; range switches by mode in a memo (`CalendarPage.tsx:107-116`) — no duplicate hook or query key.
- Household timezone sourced via `useTenant()` (`month-grid.tsx:23-24`) and threaded through tz-aware bucketing.
- Strong DST/tz test coverage (`calendar-utils.test.ts:480-496`).
- Clean presentational/container split; single interactive `<button>` per cell with `aria-label` + `cursor-pointer`; memoized grid/bucket computation.
