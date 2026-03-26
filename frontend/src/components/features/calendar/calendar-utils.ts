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

export function getEventsForDay(events: CalendarEvent[], day: Date, _timezone?: string): { allDay: CalendarEvent[]; timed: CalendarEvent[] } {
  const dayStart = new Date(day);
  dayStart.setHours(0, 0, 0, 0);
  const dayEnd = new Date(day);
  dayEnd.setHours(23, 59, 59, 999);

  const allDay: CalendarEvent[] = [];
  const timed: CalendarEvent[] = [];

  for (const evt of events) {
    const start = new Date(evt.attributes.startTime);
    const end = new Date(evt.attributes.endTime);

    if (end <= dayStart || start > dayEnd) continue;

    if (evt.attributes.allDay) {
      allDay.push(evt);
    } else {
      timed.push(evt);
    }
  }

  return { allDay, timed };
}
