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
