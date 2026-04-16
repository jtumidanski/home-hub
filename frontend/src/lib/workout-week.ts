// Single source of truth for ISO-Monday math on the client. The server is the
// authoritative normalizer (per the data-model risk register), but the client
// still needs a Monday for the URL when no `:weekStart` parameter is supplied.

export function isoMondayOf(date: Date): string {
  // Use local calendar date so the week doesn't flip at UTC midnight.
  const d = new Date(date.getFullYear(), date.getMonth(), date.getDate());
  // Sunday=0..Saturday=6 in JS; convert to Mon=0..Sun=6 then subtract.
  const iso = (d.getDay() + 6) % 7;
  d.setDate(d.getDate() - iso);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function addDays(weekStart: string, days: number): string {
  const parts = weekStart.split("-").map(Number);
  const y = parts[0] ?? 1970;
  const m = parts[1] ?? 1;
  const dd = parts[2] ?? 1;
  const d = new Date(Date.UTC(y, m - 1, dd));
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

export function currentWeekStart(): string {
  return isoMondayOf(new Date());
}
