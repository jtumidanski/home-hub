/**
 * Timezone-aware date utilities.
 *
 * All functions accept an optional IANA timezone string (e.g. "America/New_York").
 * When omitted, the browser's timezone is used as a fallback.
 *
 * Implementation relies on Intl.DateTimeFormat — no external date library needed.
 */

function resolveTimezone(tz?: string): string {
  return tz || Intl.DateTimeFormat().resolvedOptions().timeZone;
}

function getDateParts(tz?: string): { year: number; month: number; day: number } {
  const resolved = resolveTimezone(tz);
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: resolved,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());

  const year = Number(parts.find((p) => p.type === "year")!.value);
  const month = Number(parts.find((p) => p.type === "month")!.value);
  const day = Number(parts.find((p) => p.type === "day")!.value);
  return { year, month, day };
}

/** Returns a Date representing the start of today (midnight) in the given timezone. */
export function getLocalToday(tz?: string): Date {
  const { year, month, day } = getDateParts(tz);
  return new Date(year, month - 1, day, 0, 0, 0, 0);
}

/** Returns today's date as YYYY-MM-DD in the given timezone. */
export function getLocalTodayStr(tz?: string): string {
  const { year, month, day } = getDateParts(tz);
  return `${year}-${String(month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
}

/** Returns ISO 8601 start/end timestamps spanning the full day of "today" in the given timezone. */
export function getLocalTodayRange(tz?: string): { start: string; end: string } {
  const { year, month, day } = getDateParts(tz);
  const start = new Date(year, month - 1, day, 0, 0, 0, 0);
  const end = new Date(year, month - 1, day, 23, 59, 59, 999);
  return { start: start.toISOString(), end: end.toISOString() };
}

/** Returns a Date representing Monday of the current week in the given timezone. */
export function getLocalWeekStart(tz?: string): Date {
  const { year, month, day } = getDateParts(tz);
  const date = new Date(year, month - 1, day);
  const dow = date.getDay(); // 0=Sun
  const diff = dow === 0 ? -6 : 1 - dow;
  date.setDate(date.getDate() + diff);
  date.setHours(0, 0, 0, 0);
  return date;
}

/** Returns the current month as YYYY-MM in the given timezone. */
export function getLocalMonth(tz?: string): string {
  const { year, month } = getDateParts(tz);
  return `${year}-${String(month).padStart(2, "0")}`;
}
