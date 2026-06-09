import type { CalendarEvent } from "@/types/models/calendar";

export const HOUR_HEIGHT = 60; // px per hour
export const START_HOUR = 6;
export const END_HOUR = 23;
export const TOTAL_HOURS = END_HOUR - START_HOUR;

/**
 * Extract the hour and minute of a Date in a given IANA timezone.
 * Falls back to local time if the timezone is not provided or invalid.
 */
export function getTimeInZone(date: Date, timezone?: string): { hours: number; minutes: number } {
  if (!timezone) {
    return { hours: date.getHours(), minutes: date.getMinutes() };
  }
  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      timeZone: timezone,
      hour: "numeric",
      minute: "numeric",
      hour12: false,
    }).formatToParts(date);
    const hour = Number(parts.find((p) => p.type === "hour")?.value ?? date.getHours());
    const minute = Number(parts.find((p) => p.type === "minute")?.value ?? date.getMinutes());
    return { hours: hour === 24 ? 0 : hour, minutes: minute };
  } catch {
    return { hours: date.getHours(), minutes: date.getMinutes() };
  }
}

/**
 * Get the calendar date parts (year, month, day) for a Date in a given timezone.
 */
export function getDateInZone(date: Date, timezone?: string): { year: number; month: number; day: number } {
  if (!timezone) {
    return { year: date.getFullYear(), month: date.getMonth() + 1, day: date.getDate() };
  }
  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      timeZone: timezone,
      year: "numeric",
      month: "numeric",
      day: "numeric",
    }).formatToParts(date);
    const year = Number(parts.find((p) => p.type === "year")?.value ?? date.getFullYear());
    const month = Number(parts.find((p) => p.type === "month")?.value ?? date.getMonth() + 1);
    const day = Number(parts.find((p) => p.type === "day")?.value ?? date.getDate());
    return { year, month, day };
  } catch {
    return { year: date.getFullYear(), month: date.getMonth() + 1, day: date.getDate() };
  }
}

export function getWeekDays(startDate: Date): Date[] {
  const days: Date[] = [];
  for (let i = 0; i < 7; i++) {
    const d = new Date(startDate);
    d.setDate(d.getDate() + i);
    days.push(d);
  }
  return days;
}

export function getStartOfWeek(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  d.setDate(d.getDate() - day);
  d.setHours(0, 0, 0, 0);
  return d;
}

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

export function isSameDay(a: Date, b: Date, timezone?: string): boolean {
  const aDate = getDateInZone(a, timezone);
  const bDate = getDateInZone(b, timezone);
  return aDate.year === bDate.year && aDate.month === bDate.month && aDate.day === bDate.day;
}

export function isToday(date: Date, timezone?: string): boolean {
  return isSameDay(date, new Date(), timezone);
}

export function formatDateRange(start: Date, end: Date): string {
  const opts: Intl.DateTimeFormatOptions = { month: "long", day: "numeric" };
  const yearOpts: Intl.DateTimeFormatOptions = { ...opts, year: "numeric" };
  const endDate = new Date(end);
  endDate.setDate(endDate.getDate() - 1);

  if (start.getFullYear() !== endDate.getFullYear()) {
    return `${start.toLocaleDateString("en-US", yearOpts)} – ${endDate.toLocaleDateString("en-US", yearOpts)}`;
  }
  return `${start.toLocaleDateString("en-US", opts)} – ${endDate.toLocaleDateString("en-US", opts)}, ${start.getFullYear()}`;
}

export function formatHour(hour: number): string {
  if (hour === 0) return "12 AM";
  if (hour < 12) return `${hour} AM`;
  if (hour === 12) return "12 PM";
  return `${hour - 12} PM`;
}

export function getEventPosition(event: CalendarEvent, _dayStart: Date, timezone?: string): { top: number; height: number } {
  const start = new Date(event.attributes.startTime);
  const end = new Date(event.attributes.endTime);

  const startTz = getTimeInZone(start, timezone);
  const endTz = getTimeInZone(end, timezone);

  const startMinutes = (startTz.hours - START_HOUR) * 60 + startTz.minutes;
  const endMinutes = (endTz.hours - START_HOUR) * 60 + endTz.minutes;

  const top = Math.max(0, (startMinutes / 60) * HOUR_HEIGHT);
  const height = Math.max(HOUR_HEIGHT / 4, ((endMinutes - startMinutes) / 60) * HOUR_HEIGHT);

  return { top, height };
}

export interface PositionedEvent {
  event: CalendarEvent;
  column: number;
  totalColumns: number;
}

export function layoutOverlappingEvents(events: CalendarEvent[]): PositionedEvent[] {
  if (events.length === 0) return [];

  const sorted = [...events].sort((a, b) =>
    new Date(a.attributes.startTime).getTime() - new Date(b.attributes.startTime).getTime()
  );

  const positioned: PositionedEvent[] = [];
  const groups: CalendarEvent[][] = [];
  let currentGroup: CalendarEvent[] = [];
  let groupEnd = 0;

  for (const evt of sorted) {
    const start = new Date(evt.attributes.startTime).getTime();
    const end = new Date(evt.attributes.endTime).getTime();

    if (currentGroup.length === 0 || start < groupEnd) {
      currentGroup.push(evt);
      groupEnd = Math.max(groupEnd, end);
    } else {
      groups.push(currentGroup);
      currentGroup = [evt];
      groupEnd = end;
    }
  }
  if (currentGroup.length > 0) {
    groups.push(currentGroup);
  }

  for (const group of groups) {
    const columns: CalendarEvent[][] = [];

    for (const evt of group) {
      const evtStart = new Date(evt.attributes.startTime).getTime();
      let placed = false;

      for (let col = 0; col < columns.length; col++) {
        const colEvents = columns[col]!;
        const lastInCol = colEvents[colEvents.length - 1];
        if (!lastInCol) continue;
        const lastEnd = new Date(lastInCol.attributes.endTime).getTime();
        if (evtStart >= lastEnd) {
          colEvents.push(evt);
          placed = true;
          break;
        }
      }

      if (!placed) {
        columns.push([evt]);
      }
    }

    const totalColumns = columns.length;
    for (let col = 0; col < columns.length; col++) {
      for (const evt of columns[col]!) {
        positioned.push({ event: evt, column: col, totalColumns });
      }
    }
  }

  return positioned;
}

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

export function getEventsForDay(events: CalendarEvent[], day: Date, timezone?: string): { allDay: CalendarEvent[]; timed: CalendarEvent[] } {
  const { year, month, day: d } = getDateInZone(day, timezone);
  const dayStart = new Date(year, month - 1, d, 0, 0, 0, 0);
  const dayEnd = new Date(year, month - 1, d, 23, 59, 59, 999);

  // For date-only comparison of all-day events (YYYY-MM-DD)
  const dayDateStr = `${year}-${String(month).padStart(2, "0")}-${String(d).padStart(2, "0")}`;

  const allDay: CalendarEvent[] = [];
  const timed: CalendarEvent[] = [];

  for (const evt of events) {
    if (evt.attributes.allDay) {
      // All-day events: compare by date string only, ignoring timezone.
      // Both start and end are inclusive (e.g., Apr 1 to Apr 3 = three days).
      const startDate = evt.attributes.startTime.slice(0, 10);
      const endDate = evt.attributes.endTime.slice(0, 10);
      if (dayDateStr >= startDate && dayDateStr <= endDate) {
        allDay.push(evt);
      }
    } else {
      const start = new Date(evt.attributes.startTime);
      const end = new Date(evt.attributes.endTime);
      if (end <= dayStart || start > dayEnd) continue;
      timed.push(evt);
    }
  }

  return { allDay, timed };
}
