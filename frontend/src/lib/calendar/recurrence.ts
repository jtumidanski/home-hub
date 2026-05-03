export type EndsMode = "on" | "after" | "never";

export function eventStartInstant(
  startDate: string,
  startTime: string,
  allDay: boolean,
  _timeZone: string,
): Date {
  if (allDay) {
    return new Date(`${startDate}T00:00:00`);
  }
  return new Date(`${startDate}T${startTime}`);
}

// Returns the UTC instant corresponding to `endsOnDate` at 23:59:59 in
// `timeZone`, formatted as YYYYMMDDTHHMMSSZ. Derives the offset for the
// chosen date specifically via Intl.DateTimeFormat parts so DST transitions
// produce the right offset.
export function formatUntilUTC(endsOnDate: string, timeZone: string): string {
  const [y, m, d] = endsOnDate.split("-").map(Number);
  // Construct the wall-clock end-of-day as if it were UTC so we can compute
  // the offset induced by the named zone.
  const asUTC = Date.UTC(y, m - 1, d, 23, 59, 59);

  const fmt = new Intl.DateTimeFormat("en-US", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });

  // What does that UTC instant LOOK LIKE in the named zone?
  const parts = Object.fromEntries(
    fmt.formatToParts(new Date(asUTC)).map((p) => [p.type, p.value]),
  );
  const localAsUTC = Date.UTC(
    Number(parts.year),
    Number(parts.month) - 1,
    Number(parts.day),
    parts.hour === "24" ? 0 : Number(parts.hour),
    Number(parts.minute),
    Number(parts.second),
  );
  // Offset = how far the named zone's wall clock is from UTC at that moment.
  const offsetMs = localAsUTC - asUTC;
  const utcInstant = new Date(asUTC - offsetMs);

  const pad = (n: number, w = 2) => String(n).padStart(w, "0");
  return (
    `${utcInstant.getUTCFullYear()}` +
    `${pad(utcInstant.getUTCMonth() + 1)}` +
    `${pad(utcInstant.getUTCDate())}` +
    `T${pad(utcInstant.getUTCHours())}` +
    `${pad(utcInstant.getUTCMinutes())}` +
    `${pad(utcInstant.getUTCSeconds())}Z`
  );
}
