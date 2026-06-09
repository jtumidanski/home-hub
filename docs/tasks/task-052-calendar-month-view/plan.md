# Calendar Month View Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Month view mode to the calendar page — a 7-column, 5–6-row grid of complete weeks, color-coded by household member — toggleable with the existing Week/3-day time-grid view, drilling into the Week view on a day click.

**Architecture:** Frontend-only. `CalendarPage` gains `viewMode: "week" | "month"` and a `monthAnchor` (1st of the focused month) state. Month mode widens the existing `useCalendarEvents(start, end)` range to the visible grid and renders a new `MonthGrid` instead of `CalendarGrid`. All date math lives in pure, unit-tested helpers added to `calendar-utils.ts`; bucketing reuses the existing tz-aware `getEventsForDay`. The Week path is left byte-for-byte unchanged. No backend, schema, or endpoint changes.

**Tech Stack:** React 19, TypeScript, Vite, Tailwind CSS, shadcn/ui, TanStack React Query, Vitest + React Testing Library.

---

## File Structure

| File | Responsibility |
|---|---|
| `frontend/src/components/features/calendar/calendar-utils.ts` | **Modify.** Add pure date-math + bucketing helpers: `getStartOfMonth`, `addMonths`, `getMonthGridDays`, `getMonthGridRange`, `isSameMonth`, `formatMonthYear`, `formatChipTime`, `toDayKey`, `bucketEventsByDay`. |
| `frontend/src/components/features/calendar/view-mode-toggle.tsx` | **New.** `ViewModeToggle` — segmented Week \| Month control. |
| `frontend/src/components/features/calendar/month-event-chip.tsx` | **New.** `MonthEventChip` — presentational colored chip for one event (desktop). |
| `frontend/src/components/features/calendar/month-day-cell.tsx` | **New.** `MonthDayCell` — one focusable day cell; chips (desktop) or dots (mobile). |
| `frontend/src/components/features/calendar/month-grid.tsx` | **New.** `MonthGrid` — weekday header + grid of cells; memoized bucketing. |
| `frontend/src/pages/CalendarPage.tsx` | **Modify.** `viewMode`/`monthAnchor` state, range branch, `focusDay`/transition handlers, mode-aware nav/label/date-picker, render toggle + conditional grid. |
| `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts` | **Modify.** Extend with tests for the new helpers. |
| `frontend/src/components/features/calendar/__tests__/view-mode-toggle.test.tsx` | **New.** Toggle component tests. |
| `frontend/src/components/features/calendar/__tests__/month-grid.test.tsx` | **New.** MonthGrid + cell render/interaction tests. |

No changes to `calendar-service`, `useCalendarEvents`, the package overlay, or any shared type.

---

## Conventions used in this plan

- Test runner: `npm test` (`vitest run`) from `frontend/`. Run a single file with `npm test -- <path>`. Run a single test with `npx vitest run <path> -t "<name>"`.
- Type/build check: `npm run build` (`tsc -b && vite build`) from `frontend/`.
- Lint: `npm run lint` from `frontend/`.
- All `git` commands run from the worktree root (`.worktrees/task-052-calendar-month-view`); test/build commands run from `frontend/`.
- The reference month dates below are real and verified: **March 1, 2026 is a Sunday** (per the existing `calendar-utils.test.ts`), so **March 2026 → a 35-cell (5-row) grid** (grid Mar 1 → Apr 4). **August 1, 2026 is a Saturday**, so **August 2026 → a 42-cell (6-row) grid** (grid Jul 26 → Sep 5, with leading July and trailing September days).

---

## Task 1: `getStartOfMonth` and `addMonths` helpers

**Files:**
- Modify: `frontend/src/components/features/calendar/calendar-utils.ts`
- Test: `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts`

- [ ] **Step 1: Write the failing tests**

Add to the imports at the top of `__tests__/calendar-utils.test.ts` (extend the existing import block from `../calendar-utils`):

```ts
import {
  // ...existing imports stay...
  getStartOfMonth,
  addMonths,
} from "../calendar-utils";
```

Append these describe blocks to the end of the test file:

```ts
describe("getStartOfMonth", () => {
  it("returns the 1st of the month at local midnight", () => {
    const d = new Date(2026, 7, 14, 15, 30); // Aug 14, 2026 3:30pm
    const start = getStartOfMonth(d);
    expect(start.getFullYear()).toBe(2026);
    expect(start.getMonth()).toBe(7); // August
    expect(start.getDate()).toBe(1);
    expect(start.getHours()).toBe(0);
    expect(start.getMinutes()).toBe(0);
    expect(start.getSeconds()).toBe(0);
  });

  it("returns the same month when already on the 1st", () => {
    const start = getStartOfMonth(new Date(2026, 0, 1));
    expect(start.getMonth()).toBe(0);
    expect(start.getDate()).toBe(1);
  });
});

describe("addMonths", () => {
  it("advances by one calendar month", () => {
    const next = addMonths(new Date(2026, 7, 1), 1);
    expect(next.getFullYear()).toBe(2026);
    expect(next.getMonth()).toBe(8); // September
    expect(next.getDate()).toBe(1);
  });

  it("rolls over the year going forward (Dec -> Jan)", () => {
    const next = addMonths(new Date(2026, 11, 1), 1);
    expect(next.getFullYear()).toBe(2027);
    expect(next.getMonth()).toBe(0); // January
  });

  it("rolls back the year going backward (Jan -> Dec)", () => {
    const prev = addMonths(new Date(2026, 0, 1), -1);
    expect(prev.getFullYear()).toBe(2025);
    expect(prev.getMonth()).toBe(11); // December
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: FAIL — `getStartOfMonth is not a function` / `addMonths is not a function`.

- [ ] **Step 3: Implement the helpers**

Add to `calendar-utils.ts` (place after `getStartOfWeek`, around line 70):

```ts
export function getStartOfMonth(date: Date): Date {
  const d = new Date(date.getFullYear(), date.getMonth(), 1);
  d.setHours(0, 0, 0, 0);
  return d;
}

/**
 * Add `delta` calendar months to `date`. Intended for the month anchor
 * (always the 1st of a month), so no end-of-month day clamping is needed.
 */
export function addMonths(date: Date, delta: number): Date {
  const d = new Date(date);
  d.setMonth(d.getMonth() + delta);
  return d;
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/calendar-utils.ts frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts
git commit -m "feat(frontend): add getStartOfMonth and addMonths calendar helpers"
```

---

## Task 2: `getMonthGridDays` and `getMonthGridRange`

**Files:**
- Modify: `frontend/src/components/features/calendar/calendar-utils.ts`
- Test: `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts`

- [ ] **Step 1: Write the failing tests**

Add `getMonthGridDays, getMonthGridRange` to the import block. Append:

```ts
describe("getMonthGridDays", () => {
  it("returns a 35-cell grid for a month that starts on Sunday (March 2026)", () => {
    // March 1, 2026 is a Sunday; March has 31 days.
    const days = getMonthGridDays(new Date(2026, 2, 1));
    expect(days).toHaveLength(35);
    expect(days[0]!.getDay()).toBe(0); // first cell is Sunday
    expect(days[days.length - 1]!.getDay()).toBe(6); // last cell is Saturday
    // First cell is March 1 (in-month), last cell is April 4 (trailing).
    expect(days[0]!.getMonth()).toBe(2);
    expect(days[0]!.getDate()).toBe(1);
    expect(days[34]!.getMonth()).toBe(3); // April
    expect(days[34]!.getDate()).toBe(4);
  });

  it("returns a 42-cell grid with leading and trailing days (August 2026)", () => {
    // August 1, 2026 is a Saturday; grid starts on the prior Sunday (Jul 26).
    const days = getMonthGridDays(new Date(2026, 7, 1));
    expect(days).toHaveLength(42);
    expect(days[0]!.getDay()).toBe(0);
    expect(days[days.length - 1]!.getDay()).toBe(6);
    // Leading day belongs to July, trailing day belongs to September.
    expect(days[0]!.getMonth()).toBe(6); // July
    expect(days[0]!.getDate()).toBe(26);
    expect(days[41]!.getMonth()).toBe(8); // September
    expect(days[41]!.getDate()).toBe(5);
  });

  it("produces cells at local midnight", () => {
    const days = getMonthGridDays(new Date(2026, 7, 1));
    for (const d of days) {
      expect(d.getHours()).toBe(0);
      expect(d.getMinutes()).toBe(0);
    }
  });
});

describe("getMonthGridRange", () => {
  it("spans the first grid day to one day past the last (exclusive end)", () => {
    const { start, end } = getMonthGridRange(new Date(2026, 7, 1));
    // start = Jul 26 2026, end = Sep 6 2026 (last grid day Sep 5 + 1).
    expect(start.getMonth()).toBe(6);
    expect(start.getDate()).toBe(26);
    expect(end.getMonth()).toBe(8);
    expect(end.getDate()).toBe(6);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: FAIL — `getMonthGridDays is not a function`.

- [ ] **Step 3: Implement the helpers**

Add to `calendar-utils.ts` (after `addMonths`):

```ts
/**
 * The visible month grid: complete weeks (Sunday-start) covering the whole
 * month, including leading/trailing days from adjacent months. Length is a
 * multiple of 7 (typically 35 or 42).
 */
export function getMonthGridDays(monthAnchor: Date): Date[] {
  const firstOfMonth = getStartOfMonth(monthAnchor);
  const lastOfMonth = new Date(firstOfMonth.getFullYear(), firstOfMonth.getMonth() + 1, 0);
  const cursor = getStartOfWeek(firstOfMonth);
  const days: Date[] = [];
  // Keep appending until we have covered the last day of the month AND
  // completed the final week (length is a multiple of 7).
  while (days.length % 7 !== 0 || cursor <= lastOfMonth) {
    days.push(new Date(cursor));
    cursor.setDate(cursor.getDate() + 1);
  }
  return days;
}

/**
 * Half-open [start, end) date range covering the visible grid, for the
 * events query. `end` is the day after the last grid cell.
 */
export function getMonthGridRange(monthAnchor: Date): { start: Date; end: Date } {
  const days = getMonthGridDays(monthAnchor);
  const start = days[0]!;
  const last = days[days.length - 1]!;
  const end = new Date(last);
  end.setDate(end.getDate() + 1);
  return { start: new Date(start), end };
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/calendar-utils.ts frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts
git commit -m "feat(frontend): add month grid day and range helpers"
```

---

## Task 3: `isSameMonth` and `formatMonthYear`

**Files:**
- Modify: `frontend/src/components/features/calendar/calendar-utils.ts`
- Test: `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts`

- [ ] **Step 1: Write the failing tests**

Add `isSameMonth, formatMonthYear` to the import block. Append:

```ts
describe("isSameMonth", () => {
  const anchor = new Date(2026, 7, 1); // August 2026

  it("is true for a day within the anchor's month", () => {
    expect(isSameMonth(new Date(2026, 7, 15), anchor)).toBe(true);
  });

  it("is false for a leading adjacent-month day", () => {
    expect(isSameMonth(new Date(2026, 6, 26), anchor)).toBe(false); // July 26
  });

  it("is false for a trailing adjacent-month day", () => {
    expect(isSameMonth(new Date(2026, 8, 5), anchor)).toBe(false); // Sept 5
  });

  it("respects the year (Dec vs next Jan)", () => {
    expect(isSameMonth(new Date(2026, 11, 31), new Date(2027, 0, 1))).toBe(false);
  });
});

describe("formatMonthYear", () => {
  it("formats as full month name and year", () => {
    expect(formatMonthYear(new Date(2026, 5, 1))).toBe("June 2026");
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: FAIL — `isSameMonth is not a function`.

- [ ] **Step 3: Implement the helpers**

Add to `calendar-utils.ts`:

```ts
/**
 * Local (not tz-shifted) month/year comparison — matches the locally-derived
 * day number on each grid cell, so muting and the day number never disagree.
 */
export function isSameMonth(day: Date, monthAnchor: Date): boolean {
  return day.getFullYear() === monthAnchor.getFullYear() && day.getMonth() === monthAnchor.getMonth();
}

export function formatMonthYear(date: Date): string {
  return date.toLocaleDateString("en-US", { month: "long", year: "numeric" });
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/calendar-utils.ts frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts
git commit -m "feat(frontend): add isSameMonth and formatMonthYear helpers"
```

---

## Task 4: `formatChipTime`

**Files:**
- Modify: `frontend/src/components/features/calendar/calendar-utils.ts`
- Test: `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts`

- [ ] **Step 1: Write the failing tests**

Add `formatChipTime` to the import block. Append:

```ts
describe("formatChipTime", () => {
  it("drops :00 minutes on the hour (AM)", () => {
    expect(formatChipTime("2026-03-26T09:00:00Z", "UTC")).toBe("9a");
  });

  it("keeps non-zero minutes", () => {
    expect(formatChipTime("2026-03-26T09:30:00Z", "UTC")).toBe("9:30a");
  });

  it("formats noon as 12p", () => {
    expect(formatChipTime("2026-03-26T12:00:00Z", "UTC")).toBe("12p");
  });

  it("formats afternoon times as PM", () => {
    expect(formatChipTime("2026-03-26T14:30:00Z", "UTC")).toBe("2:30p");
  });

  it("formats midnight as 12a", () => {
    expect(formatChipTime("2026-03-26T00:00:00Z", "UTC")).toBe("12a");
  });

  it("is timezone-aware", () => {
    // 14:00 UTC is 10:00 in America/New_York (EDT, UTC-4) on this date.
    expect(formatChipTime("2026-03-26T14:00:00Z", "America/New_York")).toBe("10a");
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: FAIL — `formatChipTime is not a function`.

- [ ] **Step 3: Implement the helper**

Add to `calendar-utils.ts` (it reuses the existing `getTimeInZone`):

```ts
/**
 * Compact timed-event prefix for month chips: single-letter meridiem,
 * minutes dropped on the hour. e.g. "9a", "9:30a", "12p", "2:30p".
 * Timezone-aware via getTimeInZone.
 */
export function formatChipTime(iso: string, timezone?: string): string {
  const { hours, minutes } = getTimeInZone(new Date(iso), timezone);
  const meridiem = hours < 12 ? "a" : "p";
  const h12 = hours % 12 === 0 ? 12 : hours % 12;
  return minutes === 0 ? `${h12}${meridiem}` : `${h12}:${String(minutes).padStart(2, "0")}${meridiem}`;
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/calendar-utils.ts frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts
git commit -m "feat(frontend): add formatChipTime helper for month chips"
```

---

## Task 5: `toDayKey` and `bucketEventsByDay`

**Files:**
- Modify: `frontend/src/components/features/calendar/calendar-utils.ts`
- Test: `frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts`

- [ ] **Step 1: Write the failing tests**

Add `toDayKey, bucketEventsByDay` to the import block. Append (uses the file's existing `makeEvent` helper):

```ts
describe("toDayKey", () => {
  it("formats a local date as YYYY-MM-DD with zero padding", () => {
    expect(toDayKey(new Date(2026, 7, 5))).toBe("2026-08-05");
  });
});

describe("bucketEventsByDay", () => {
  it("places a timed event in the correct day's bucket", () => {
    const gridDays = [new Date(2026, 2, 25), new Date(2026, 2, 26), new Date(2026, 2, 27)];
    const evt = makeEvent({ startTime: "2026-03-26T10:00:00Z", endTime: "2026-03-26T11:00:00Z" });
    const map = bucketEventsByDay(gridDays, [evt], "UTC");
    expect(map.get("2026-03-26")!.timed).toHaveLength(1);
    expect(map.get("2026-03-25")!.timed).toHaveLength(0);
    expect(map.get("2026-03-27")!.timed).toHaveLength(0);
  });

  it("sorts timed events within a day by start time ascending", () => {
    const gridDays = [new Date(2026, 2, 26)];
    const later = makeEvent({ id: "later", startTime: "2026-03-26T14:00:00Z", endTime: "2026-03-26T15:00:00Z" });
    const earlier = makeEvent({ id: "earlier", startTime: "2026-03-26T09:00:00Z", endTime: "2026-03-26T10:00:00Z" });
    const map = bucketEventsByDay(gridDays, [later, earlier], "UTC");
    expect(map.get("2026-03-26")!.timed.map((e) => e.id)).toEqual(["earlier", "later"]);
  });

  it("spans a multi-day all-day event across every covered day", () => {
    const gridDays = [new Date(2026, 3, 1), new Date(2026, 3, 2), new Date(2026, 3, 3), new Date(2026, 3, 4)];
    const evt = makeEvent({ allDay: true, startTime: "2026-04-01", endTime: "2026-04-03" });
    const map = bucketEventsByDay(gridDays, [evt], "UTC");
    expect(map.get("2026-04-01")!.allDay).toHaveLength(1);
    expect(map.get("2026-04-02")!.allDay).toHaveLength(1);
    expect(map.get("2026-04-03")!.allDay).toHaveLength(1);
    expect(map.get("2026-04-04")!.allDay).toHaveLength(0);
  });

  it("returns an entry for every grid day", () => {
    const gridDays = getMonthGridDays(new Date(2026, 7, 1));
    const map = bucketEventsByDay(gridDays, [], "UTC");
    expect(map.size).toBe(gridDays.length);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: FAIL — `bucketEventsByDay is not a function`.

- [ ] **Step 3: Implement the helpers**

Add to `calendar-utils.ts` (after `getEventsForDay`, reusing it):

```ts
/** Local YYYY-MM-DD key for a grid day. */
export function toDayKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

/**
 * Bucket events into each grid day, keyed by local YYYY-MM-DD. Delegates to
 * the existing tz-aware getEventsForDay so multi-day/all-day/midnight bucketing
 * reuse already-tested logic. Within a day, timed events are sorted by start.
 * Computed once per render in MonthGrid.
 */
export function bucketEventsByDay(
  gridDays: Date[],
  events: CalendarEvent[],
  timezone?: string,
): Map<string, { allDay: CalendarEvent[]; timed: CalendarEvent[] }> {
  const map = new Map<string, { allDay: CalendarEvent[]; timed: CalendarEvent[] }>();
  for (const day of gridDays) {
    const { allDay, timed } = getEventsForDay(events, day, timezone);
    const sortedTimed = [...timed].sort(
      (a, b) => new Date(a.attributes.startTime).getTime() - new Date(b.attributes.startTime).getTime(),
    );
    map.set(toDayKey(day), { allDay, timed: sortedTimed });
  }
  return map;
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `npm test -- src/components/features/calendar/__tests__/calendar-utils.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/calendar-utils.ts frontend/src/components/features/calendar/__tests__/calendar-utils.test.ts
git commit -m "feat(frontend): add toDayKey and bucketEventsByDay helpers"
```

---

## Task 6: `ViewModeToggle` component

**Files:**
- Create: `frontend/src/components/features/calendar/view-mode-toggle.tsx`
- Test: `frontend/src/components/features/calendar/__tests__/view-mode-toggle.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `__tests__/view-mode-toggle.test.tsx`:

```tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ViewModeToggle } from "../view-mode-toggle";

describe("ViewModeToggle", () => {
  it("renders Week and Month options", () => {
    render(<ViewModeToggle mode="week" onChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Week" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Month" })).toBeInTheDocument();
  });

  it("marks the active mode with aria-pressed", () => {
    render(<ViewModeToggle mode="month" onChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Month" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Week" })).toHaveAttribute("aria-pressed", "false");
  });

  it("calls onChange with the clicked mode", async () => {
    const onChange = vi.fn();
    render(<ViewModeToggle mode="week" onChange={onChange} />);
    await userEvent.click(screen.getByRole("button", { name: "Month" }));
    expect(onChange).toHaveBeenCalledWith("month");
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `npm test -- src/components/features/calendar/__tests__/view-mode-toggle.test.tsx`
Expected: FAIL — cannot resolve `../view-mode-toggle`.

- [ ] **Step 3: Implement the component**

Create `view-mode-toggle.tsx`:

```tsx
import { Button } from "@/components/ui/button";

export type CalendarViewMode = "week" | "month";

interface ViewModeToggleProps {
  mode: CalendarViewMode;
  onChange: (mode: CalendarViewMode) => void;
}

export function ViewModeToggle({ mode, onChange }: ViewModeToggleProps) {
  return (
    <div className="flex items-center border rounded-md" role="group" aria-label="Calendar view mode">
      <Button
        variant="ghost"
        size="sm"
        aria-pressed={mode === "week"}
        className={mode === "week" ? "bg-accent text-accent-foreground" : ""}
        onClick={() => onChange("week")}
      >
        Week
      </Button>
      <Button
        variant="ghost"
        size="sm"
        aria-pressed={mode === "month"}
        className={mode === "month" ? "bg-accent text-accent-foreground" : ""}
        onClick={() => onChange("month")}
      >
        Month
      </Button>
    </div>
  );
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `npm test -- src/components/features/calendar/__tests__/view-mode-toggle.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/view-mode-toggle.tsx frontend/src/components/features/calendar/__tests__/view-mode-toggle.test.tsx
git commit -m "feat(frontend): add ViewModeToggle for calendar week/month switch"
```

---

## Task 7: `MonthEventChip` component

**Files:**
- Create: `frontend/src/components/features/calendar/month-event-chip.tsx`
- Test: covered in `__tests__/month-grid.test.tsx` (Task 9). No standalone test file — the chip is exercised through `MonthDayCell`/`MonthGrid`.

- [ ] **Step 1: Implement the component**

Create `month-event-chip.tsx`:

```tsx
import type { CalendarEvent } from "@/types/models/calendar";
import { formatChipTime } from "./calendar-utils";

interface MonthEventChipProps {
  event: CalendarEvent;
  timezone?: string | undefined;
}

/**
 * Presentational chip for one event in a month-grid day cell (desktop/tablet).
 * Not interactive — the enclosing day cell owns the click/focus (see design 2.5).
 */
export function MonthEventChip({ event, timezone }: MonthEventChipProps) {
  const { attributes: attrs } = event;
  const isBusy = attrs.title === "Busy" && !attrs.isOwner;
  const timeLabel = attrs.allDay ? "" : formatChipTime(attrs.startTime, timezone);

  return (
    <div
      className="rounded px-1 py-0.5 text-[10px] leading-tight truncate text-white"
      style={{ backgroundColor: isBusy ? "#9ca3af" : attrs.userColor }}
      title={attrs.title}
    >
      {timeLabel && <span className="font-semibold mr-1">{timeLabel}</span>}
      <span>{attrs.title}</span>
    </div>
  );
}
```

- [ ] **Step 2: Verify it type-checks**

Run: `npx tsc -b`
Expected: PASS (no type errors). Behavior is verified by the MonthGrid tests in Task 9.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/features/calendar/month-event-chip.tsx
git commit -m "feat(frontend): add MonthEventChip presentational component"
```

---

## Task 8: `MonthDayCell` component

**Files:**
- Create: `frontend/src/components/features/calendar/month-day-cell.tsx`
- Test: covered in `__tests__/month-grid.test.tsx` (Task 9) via the full grid.

- [ ] **Step 1: Implement the component**

Create `month-day-cell.tsx`:

```tsx
import type { CalendarEvent } from "@/types/models/calendar";
import { MonthEventChip } from "./month-event-chip";

interface DayBucket {
  allDay: CalendarEvent[];
  timed: CalendarEvent[];
}

interface MonthDayCellProps {
  day: Date;
  bucket: DayBucket;
  inMonth: boolean;
  isToday: boolean;
  isDesktop: boolean;
  timezone?: string | undefined;
  onDayClick: (day: Date) => void;
}

const MAX_DOTS = 4;

function eventColor(event: CalendarEvent): string {
  const { attributes: attrs } = event;
  return attrs.title === "Busy" && !attrs.isOwner ? "#9ca3af" : attrs.userColor;
}

function MonthDayDots({ events }: { events: CalendarEvent[] }) {
  const overflow = events.length > MAX_DOTS;
  const shown = overflow ? events.slice(0, MAX_DOTS - 1) : events;
  return (
    <div className="flex flex-wrap gap-0.5 mt-1" aria-hidden="true">
      {shown.map((evt) => (
        <span
          key={evt.id}
          className="w-1.5 h-1.5 rounded-full"
          style={{ backgroundColor: eventColor(evt) }}
        />
      ))}
      {overflow && (
        <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground" data-testid="dot-overflow" />
      )}
    </div>
  );
}

/**
 * One month-grid day cell. The cell is the single keyboard-focusable,
 * activatable control (design 2.5); chips/dots inside are presentational.
 */
export function MonthDayCell({
  day,
  bucket,
  inMonth,
  isToday,
  isDesktop,
  timezone,
  onDayClick,
}: MonthDayCellProps) {
  const ordered = [...bucket.allDay, ...bucket.timed];
  const dayLabel = day.toLocaleDateString("en-US", { month: "long", day: "numeric" });
  const label = `${dayLabel}, ${ordered.length} event${ordered.length === 1 ? "" : "s"}`;

  const cellTone = isToday ? "bg-primary/10" : !inMonth ? "bg-muted/30" : "";
  const numberTone = isToday ? "text-primary" : !inMonth ? "text-muted-foreground" : "";

  return (
    <button
      type="button"
      aria-label={label}
      onClick={() => onDayClick(day)}
      className={`flex flex-col min-h-0 overflow-hidden border-r border-b last:border-r-0 text-left p-1 cursor-pointer hover:bg-accent/40 ${cellTone}`}
    >
      <span className={`text-xs font-medium px-0.5 ${numberTone}`}>{day.getDate()}</span>
      {isDesktop ? (
        <div className="flex-1 overflow-y-auto flex flex-col gap-0.5 mt-0.5">
          {ordered.map((evt) => (
            <MonthEventChip key={evt.id} event={evt} timezone={timezone} />
          ))}
        </div>
      ) : (
        <MonthDayDots events={ordered} />
      )}
    </button>
  );
}
```

- [ ] **Step 2: Verify it type-checks**

Run: `npx tsc -b`
Expected: PASS. Behavior verified by MonthGrid tests in Task 9.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/features/calendar/month-day-cell.tsx
git commit -m "feat(frontend): add MonthDayCell with chips (desktop) and dots (mobile)"
```

---

## Task 9: `MonthGrid` component

**Files:**
- Create: `frontend/src/components/features/calendar/month-grid.tsx`
- Test: `frontend/src/components/features/calendar/__tests__/month-grid.test.tsx`

- [ ] **Step 1: Write the failing test**

Create `__tests__/month-grid.test.tsx`:

```tsx
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { CalendarEvent } from "@/types/models/calendar";
import { MonthGrid } from "../month-grid";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({ household: { attributes: { timezone: "UTC" } } }),
}));

function makeEvent(overrides: Partial<CalendarEvent["attributes"]> & { id?: string } = {}): CalendarEvent {
  const { id = "evt-1", ...attrs } = overrides;
  return {
    id,
    type: "calendar-events",
    attributes: {
      title: "Test Event",
      description: null,
      startTime: "2026-08-14T10:00:00Z",
      endTime: "2026-08-14T11:00:00Z",
      allDay: false,
      location: null,
      visibility: "default",
      isOwner: true,
      userDisplayName: "Test User",
      userColor: "#4285F4",
      sourceId: "",
      connectionId: "",
      isRecurring: false,
      ...attrs,
    },
  };
}

afterEach(() => {
  vi.useRealTimers();
});

describe("MonthGrid", () => {
  it("renders seven weekday headers", () => {
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    for (const label of ["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"]) {
      expect(screen.getByText(label)).toBeInTheDocument();
    }
  });

  it("renders 42 day cells for August 2026 (6-row month)", () => {
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    // Each cell is a button with an aria-label; weekday headers are not buttons.
    expect(screen.getAllByRole("button")).toHaveLength(42);
  });

  it("renders 35 day cells for March 2026 (5-row month)", () => {
    render(<MonthGrid monthAnchor={new Date(2026, 2, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    expect(screen.getAllByRole("button")).toHaveLength(35);
  });

  it("highlights today's cell", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-08-14T12:00:00Z"));
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={vi.fn()} />);
    const todayCell = screen.getByRole("button", { name: /August 14, / });
    expect(todayCell.className).toContain("bg-primary/10");
  });

  it("renders a timed event as a chip with a time prefix on desktop", () => {
    const evt = makeEvent({ startTime: "2026-08-14T09:00:00Z", endTime: "2026-08-14T10:00:00Z", title: "Standup" });
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[evt]} isDesktop onDayClick={vi.fn()} />);
    expect(screen.getByText("Standup")).toBeInTheDocument();
    expect(screen.getByText("9a")).toBeInTheDocument();
  });

  it("renders all events of a dense day in the DOM (no truncation)", () => {
    const events = Array.from({ length: 8 }, (_, i) =>
      makeEvent({
        id: `e${i}`,
        title: `Event ${i}`,
        startTime: `2026-08-14T0${i}:00:00Z`,
        endTime: `2026-08-14T0${i}:30:00Z`,
      }),
    );
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={events} isDesktop onDayClick={vi.fn()} />);
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`Event ${i}`)).toBeInTheDocument();
    }
  });

  it("renders dots (not chips) on mobile and caps at 4 with an overflow indicator", () => {
    const events = Array.from({ length: 6 }, (_, i) =>
      makeEvent({ id: `e${i}`, title: `Event ${i}`, startTime: `2026-08-14T0${i}:00:00Z`, endTime: `2026-08-14T0${i}:30:00Z` }),
    );
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={events} isDesktop={false} onDayClick={vi.fn()} />);
    // Chips are not rendered in mobile mode.
    expect(screen.queryByText("Event 0")).not.toBeInTheDocument();
    // 6 events > 4 -> 3 colored dots + 1 overflow dot.
    expect(screen.getByTestId("dot-overflow")).toBeInTheDocument();
  });

  it("fires onDayClick with the cell's date for an in-month day", async () => {
    const onDayClick = vi.fn();
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={onDayClick} />);
    await userEvent.click(screen.getByRole("button", { name: /August 14, / }));
    expect(onDayClick).toHaveBeenCalledTimes(1);
    const arg = onDayClick.mock.calls[0][0] as Date;
    expect(arg.getMonth()).toBe(7);
    expect(arg.getDate()).toBe(14);
  });

  it("fires onDayClick for a trailing adjacent-month day", async () => {
    const onDayClick = vi.fn();
    render(<MonthGrid monthAnchor={new Date(2026, 7, 1)} events={[]} isDesktop onDayClick={onDayClick} />);
    // Sept 5, 2026 is the last (trailing) grid cell.
    await userEvent.click(screen.getByRole("button", { name: /September 5, / }));
    const arg = onDayClick.mock.calls[0][0] as Date;
    expect(arg.getMonth()).toBe(8);
    expect(arg.getDate()).toBe(5);
  });
});
```

> Note on the count assertions: the weekday header cells are plain `<div>`s, not buttons, so `getAllByRole("button")` counts only day cells. If two grid cells ever share a day number across months (e.g. "September 5" vs nothing), the `name` regex includes the full-month label (`/September 5, /`) which is unique within a single grid.

- [ ] **Step 2: Run the test to verify it fails**

Run: `npm test -- src/components/features/calendar/__tests__/month-grid.test.tsx`
Expected: FAIL — cannot resolve `../month-grid`.

- [ ] **Step 3: Implement the component**

Create `month-grid.tsx`:

```tsx
import { useMemo } from "react";
import type { CalendarEvent } from "@/types/models/calendar";
import { useTenant } from "@/context/tenant-context";
import { MonthDayCell } from "./month-day-cell";
import {
  bucketEventsByDay,
  getMonthGridDays,
  isSameMonth,
  isToday,
  toDayKey,
} from "./calendar-utils";

interface MonthGridProps {
  monthAnchor: Date;
  events: CalendarEvent[];
  isDesktop: boolean;
  onDayClick: (day: Date) => void;
}

const EMPTY_BUCKET = { allDay: [] as CalendarEvent[], timed: [] as CalendarEvent[] };

export function MonthGrid({ monthAnchor, events, isDesktop, onDayClick }: MonthGridProps) {
  const { household } = useTenant();
  const timezone = household?.attributes.timezone;

  const gridDays = useMemo(() => getMonthGridDays(monthAnchor), [monthAnchor]);
  const buckets = useMemo(
    () => bucketEventsByDay(gridDays, events, timezone),
    [gridDays, events, timezone],
  );

  const rows = gridDays.length / 7;
  const weekdayLabels = gridDays
    .slice(0, 7)
    .map((d) => d.toLocaleDateString("en-US", { weekday: "short" }).toUpperCase());

  return (
    <div
      className="border rounded-lg overflow-hidden bg-background grid h-full"
      style={{
        gridTemplateColumns: "repeat(7, minmax(0, 1fr))",
        gridTemplateRows: `auto repeat(${rows}, minmax(0, 1fr))`,
      }}
    >
      {weekdayLabels.map((label) => (
        <div
          key={label}
          className="text-xs text-muted-foreground uppercase text-center py-1 border-b border-r last:border-r-0"
        >
          {label}
        </div>
      ))}
      {gridDays.map((day) => {
        const key = toDayKey(day);
        return (
          <MonthDayCell
            key={key}
            day={day}
            bucket={buckets.get(key) ?? EMPTY_BUCKET}
            inMonth={isSameMonth(day, monthAnchor)}
            isToday={isToday(day, timezone)}
            isDesktop={isDesktop}
            timezone={timezone}
            onDayClick={onDayClick}
          />
        );
      })}
    </div>
  );
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `npm test -- src/components/features/calendar/__tests__/month-grid.test.tsx`
Expected: PASS (all cases).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/calendar/month-grid.tsx frontend/src/components/features/calendar/__tests__/month-grid.test.tsx
git commit -m "feat(frontend): add MonthGrid month-view calendar grid"
```

---

## Task 10: Wire Month view into `CalendarPage`

**Files:**
- Modify: `frontend/src/pages/CalendarPage.tsx`

> This task is wiring. The pure logic it depends on (range, transition math, day-key bucketing) is unit-tested in Tasks 1–5, and the rendered grid + day-click interaction are component-tested in Task 9. A full `CalendarPage` render test would require mocking ~6 query hooks plus tenant/auth context for marginal added coverage, so it is intentionally omitted (YAGNI); verification here is type-check + full suite + the existing tests staying green.

- [ ] **Step 1: Update the imports**

In `CalendarPage.tsx`, change the calendar-utils import (line 21) to add the new helpers and the new components:

```tsx
import { CalendarGrid } from "@/components/features/calendar/calendar-grid";
import { MonthGrid } from "@/components/features/calendar/month-grid";
import { ViewModeToggle, type CalendarViewMode } from "@/components/features/calendar/view-mode-toggle";
```

and replace the existing `getStartOfWeek, formatDateRange` import line with:

```tsx
import {
  getStartOfWeek,
  formatDateRange,
  getStartOfMonth,
  getMonthGridRange,
  formatMonthYear,
  isSameMonth,
  addMonths,
} from "@/components/features/calendar/calendar-utils";
```

(Keep the `CalendarGrid` import; add the two new component imports near it.)

- [ ] **Step 2: Add view-mode and month-anchor state**

After the `weekStart` state declaration (line 39), add:

```tsx
  const [viewMode, setViewMode] = useState<CalendarViewMode>("week");
  const [monthAnchor, setMonthAnchor] = useState(() => getStartOfMonth(new Date()));
```

- [ ] **Step 3: Branch the events query range on view mode**

Replace the two lines computing `startISO`/`endISO` (currently lines 95–96):

```tsx
  const startISO = weekStart.toISOString();
  const endISO = weekEnd.toISOString();
```

with a mode-aware memo:

```tsx
  const { startISO, endISO } = useMemo(() => {
    if (viewMode === "month") {
      const { start, end } = getMonthGridRange(monthAnchor);
      return { startISO: start.toISOString(), endISO: end.toISOString() };
    }
    return { startISO: weekStart.toISOString(), endISO: weekEnd.toISOString() };
  }, [viewMode, monthAnchor, weekStart, weekEnd]);
```

- [ ] **Step 4: Add `focusDay` and refactor navigation/selection handlers**

Add a `focusDay` callback (near the other handlers, after `weekEnd`/before `goPrev`):

```tsx
  const focusDay = useCallback((day: Date) => {
    if (isDesktop) {
      setWeekStart(getStartOfWeek(day));
    } else {
      const start = new Date(day);
      start.setDate(start.getDate() - 1);
      start.setHours(0, 0, 0, 0);
      setWeekStart(start);
    }
  }, [isDesktop]);
```

Replace `goPrev`/`goNext` (lines 133–142) with mode-aware versions:

```tsx
  const goPrev = () => {
    if (viewMode === "month") {
      setMonthAnchor((a) => addMonths(a, -1));
      return;
    }
    const prev = new Date(weekStart);
    prev.setDate(prev.getDate() - dayCount);
    setWeekStart(prev);
  };
  const goNext = () => {
    if (viewMode === "month") {
      setMonthAnchor((a) => addMonths(a, 1));
      return;
    }
    const next = new Date(weekStart);
    next.setDate(next.getDate() + dayCount);
    setWeekStart(next);
  };
```

Replace `handleCalendarSelect` (lines 144–155) so it routes by mode and reuses `focusDay`:

```tsx
  const handleCalendarSelect = (date: Date | undefined) => {
    if (!date) return;
    if (viewMode === "month") {
      setMonthAnchor(getStartOfMonth(date));
    } else {
      focusDay(date);
    }
    setCalPickerOpen(false);
  };
```

Add the day-click and toggle handlers (after `handleCalendarSelect`):

```tsx
  const handleDayClick = useCallback((day: Date) => {
    focusDay(day);
    setViewMode("week");
  }, [focusDay]);

  const handleViewModeChange = (mode: CalendarViewMode) => {
    if (mode === viewMode) return;
    if (mode === "month") {
      setMonthAnchor(getStartOfMonth(weekStart));
    } else {
      const today = new Date();
      focusDay(isSameMonth(today, monthAnchor) ? today : monthAnchor);
    }
    setViewMode(mode);
  };
```

- [ ] **Step 5: Update the header label, toggle placement, and date picker**

Change the date-range label (line 244) to be mode-aware:

```tsx
          <p className="text-sm text-muted-foreground">
            {viewMode === "month" ? formatMonthYear(monthAnchor) : formatDateRange(weekStart, weekEnd)}
          </p>
```

Add the toggle inside the header control cluster — place it as the first child of `<div className="flex items-center gap-2">` (line 247), before the prev/Today/next group:

```tsx
        <div className="flex items-center gap-2">
          <ViewModeToggle mode={viewMode} onChange={handleViewModeChange} />
          <div className="flex items-center border rounded-md">
            {/* ...existing prev / Today popover / next... */}
```

Update the `Calendar` popover (lines 260–265) so `selected`/`defaultMonth` follow the active mode:

```tsx
                <Calendar
                  mode="single"
                  selected={viewMode === "month" ? monthAnchor : weekStart}
                  onSelect={handleCalendarSelect}
                  defaultMonth={viewMode === "month" ? monthAnchor : weekStart}
                />
```

- [ ] **Step 6: Render the correct grid for the active mode**

In the `hasCalendar` block (lines 324–339), replace the grid render so month mode shows `MonthGrid`:

```tsx
      {hasCalendar ? (
        <div className="flex-1 min-h-0">
          {eventsQuery.isLoading ? (
            <Skeleton className="h-full w-full" />
          ) : viewMode === "month" ? (
            <MonthGrid
              monthAnchor={monthAnchor}
              events={events}
              isDesktop={isDesktop}
              onDayClick={handleDayClick}
            />
          ) : (
            <CalendarGrid
              weekStart={weekStart}
              events={events}
              dayCount={dayCount}
              hasWriteAccess={hasWriteAccess}
              onSlotClick={handleSlotClick}
              onEditEvent={handleEditEvent}
              onDeleteEvent={handleDeleteEvent}
            />
          )}
        </div>
      ) : (
```

- [ ] **Step 7: Type-check and lint**

Run: `npm run build` and `npm run lint`
Expected: PASS — no type errors, no lint errors. (If lint flags an unused import, ensure every newly imported helper/component is actually referenced.)

- [ ] **Step 8: Run the full calendar test suite**

Run: `npm test -- src/components/features/calendar`
Expected: PASS — new tests green, existing `calendar-utils.test.ts` and `event-form-dialog.test.tsx` still green (Week path unchanged).

- [ ] **Step 9: Commit**

```bash
git add frontend/src/pages/CalendarPage.tsx
git commit -m "feat(frontend): add month view mode toggle and wiring to CalendarPage"
```

---

## Task 11: Full verification

**Files:** none (verification only).

- [ ] **Step 1: Run the full frontend test suite**

Run (from `frontend/`): `npm test`
Expected: PASS — entire suite green.

- [ ] **Step 2: Run the production build**

Run (from `frontend/`): `npm run build`
Expected: PASS — `tsc -b` clean, `vite build` succeeds.

- [ ] **Step 3: Lint**

Run (from `frontend/`): `npm run lint`
Expected: PASS — no errors.

- [ ] **Step 4: Manual acceptance smoke (optional but recommended)**

If running the app locally (`scripts/local-up.sh`), verify against the PRD §10 acceptance list:
- Week/Month toggle switches grids; page loads in Week by default.
- Month grid shows complete weeks (5–6 rows), muted adjacent-month cells, today highlighted.
- Desktop chips show time prefix for timed events; dense day cell scrolls internally.
- Mobile (narrow viewport) shows dots, capped at 4 with overflow dot.
- Clicking a day (incl. adjacent-month) drops into Week mode focused on that day.
- Month prev/next moves a month; header reads "Month YYYY"; date picker jumps month without leaving Month mode.

- [ ] **Step 5: Final commit (if any uncommitted verification fixups)**

```bash
git add -A
git commit -m "chore(frontend): month view verification fixups" || echo "nothing to commit"
```

---

## Self-Review Notes (author)

**Spec coverage** — every PRD §4 requirement and design §8 file maps to a task:
- 4.1 toggle / default Week / non-persisted → Tasks 6, 10 (state defaults `"week"`, in component state only).
- 4.2 grid layout / complete weeks / muted adjacent / today highlight → Tasks 2, 8, 9.
- 4.3 desktop chips / time prefix / internal scroll / no "+N more" → Tasks 4, 7, 8, 9 (scroll test asserts all events in DOM).
- 4.4 mobile dots / cap-4 + overflow → Task 8, 9.
- 4.5 day-cell + chip click drill-down → Tasks 8, 9, 10 (`handleDayClick`).
- 4.6 month prev/next, "Month YYYY" label, date picker jumps month → Tasks 3, 10.
- 4.7 data fetching widened range, reuse package merge + `getEventsForDay` → Tasks 2, 5, 10.
- §8 perf (memoized bucket once per render) → Task 9 (`useMemo`).
- §8 tz correctness → Tasks 4, 5, 9 (reuse `getEventsForDay`/`getDateInZone`/`isToday`).
- §8 a11y (cell-level activation, aria-pressed toggle) → Tasks 6, 8.

**Placeholder scan** — no TBD/TODO/"handle edge cases"/"similar to Task N"; all code blocks are complete.

**Type consistency** — `CalendarViewMode` exported from `view-mode-toggle.tsx` and reused in `CalendarPage`; `toDayKey`/`bucketEventsByDay`/`getMonthGridDays`/`isSameMonth`/`isToday` names match across utils, `MonthGrid`, and `CalendarPage`; `DayBucket` shape (`{ allDay, timed }`) is consistent between `bucketEventsByDay`, `MonthGrid` (`EMPTY_BUCKET`), and `MonthDayCell`.
