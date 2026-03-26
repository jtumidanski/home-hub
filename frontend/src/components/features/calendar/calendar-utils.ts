import type { CalendarEvent } from "@/types/models/calendar";

export const HOUR_HEIGHT = 60; // px per hour
export const START_HOUR = 6;
export const END_HOUR = 23;
export const TOTAL_HOURS = END_HOUR - START_HOUR;

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

export function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

export function isToday(date: Date): boolean {
  return isSameDay(date, new Date());
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

export function getEventPosition(event: CalendarEvent, _dayStart: Date): { top: number; height: number } {
  const start = new Date(event.attributes.startTime);
  const end = new Date(event.attributes.endTime);

  const startMinutes = (start.getHours() - START_HOUR) * 60 + start.getMinutes();
  const endMinutes = (end.getHours() - START_HOUR) * 60 + end.getMinutes();

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

export function getEventsForDay(events: CalendarEvent[], day: Date): { allDay: CalendarEvent[]; timed: CalendarEvent[] } {
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
