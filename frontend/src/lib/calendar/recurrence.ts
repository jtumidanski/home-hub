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
