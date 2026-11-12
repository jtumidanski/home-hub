import type { QueryParamCodec } from "@/lib/hooks/use-query-param-state";
import { formatLocalDateStr, parseLocalDateStr } from "@/lib/date-utils";
import { getStartOfMonth, getStartOfWeek } from "./calendar-utils";
import type { CalendarViewMode } from "./view-mode-toggle";

/** First displayed day of a week view that focuses `day`. */
export function weekAnchorFor(day: Date, isDesktop: boolean): Date {
  if (isDesktop) return getStartOfWeek(day);
  const start = new Date(day);
  start.setDate(start.getDate() - 1);
  start.setHours(0, 0, 0, 0);
  return start;
}

export function calendarDateCodec(view: CalendarViewMode, isDesktop: boolean): QueryParamCodec<Date> {
  const normalize = (d: Date): Date => {
    if (view === "month") return getStartOfMonth(d);
    return isDesktop ? getStartOfWeek(d) : d;
  };
  return {
    key: "date",
    parse: (raw) => {
      const d = parseLocalDateStr(raw);
      return d ? normalize(d) : null;
    },
    serialize: formatLocalDateStr,
    current: () => (view === "month" ? getStartOfMonth(new Date()) : weekAnchorFor(new Date(), isDesktop)),
    equals: (a, b) => formatLocalDateStr(a) === formatLocalDateStr(b),
  };
}
