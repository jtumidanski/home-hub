import { useRef, useEffect } from "react";
import type { CalendarEvent } from "@/types/models/calendar";
import { EventBlock } from "./event-block";
import { AllDayEventRow } from "./all-day-event-row";
import {
  HOUR_HEIGHT,
  START_HOUR,
  END_HOUR,
  getWeekDays,
  isToday,
  formatHour,
  getEventPosition,
  getEventsForDay,
  layoutOverlappingEvents,
} from "./calendar-utils";

interface CalendarGridProps {
  weekStart: Date;
  events: CalendarEvent[];
}

export function CalendarGrid({ weekStart, events }: CalendarGridProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const days = getWeekDays(weekStart);
  const hours = Array.from({ length: END_HOUR - START_HOUR }, (_, i) => START_HOUR + i);

  useEffect(() => {
    if (scrollRef.current) {
      const now = new Date();
      const currentHour = now.getHours();
      const scrollTo = Math.max(0, (currentHour - START_HOUR - 1) * HOUR_HEIGHT);
      scrollRef.current.scrollTop = scrollTo;
    }
  }, []);

  const allDayByDay = days.map((day) => getEventsForDay(events, day).allDay);
  const hasAllDay = allDayByDay.some((d) => d.length > 0);

  return (
    <div className="border rounded-lg overflow-hidden bg-background flex flex-col h-full">
      {/* Header row: day names + dates */}
      <div className="flex border-b flex-shrink-0">
        <div className="w-16 flex-shrink-0 border-r" />
        {days.map((day, i) => (
          <div
            key={i}
            className={`flex-1 text-center py-2 border-r last:border-r-0 ${
              isToday(day) ? "bg-primary/10" : ""
            }`}
          >
            <div className="text-xs text-muted-foreground uppercase">
              {day.toLocaleDateString("en-US", { weekday: "short" })}
            </div>
            <div className={`text-lg font-semibold ${isToday(day) ? "text-primary" : ""}`}>
              {day.getDate()}
            </div>
          </div>
        ))}
      </div>

      {/* All-day events section */}
      {hasAllDay && (
        <div className="flex border-b flex-shrink-0">
          <div className="w-16 flex-shrink-0 border-r flex items-center justify-center">
            <span className="text-xs text-muted-foreground">All day</span>
          </div>
          {days.map((_day, i) => (
            <div key={i} className="flex-1 border-r last:border-r-0">
              <AllDayEventRow events={allDayByDay[i] ?? []} />
            </div>
          ))}
        </div>
      )}

      {/* Scrollable hour grid */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto">
        <div className="flex" style={{ height: `${hours.length * HOUR_HEIGHT}px` }}>
          {/* Hour labels */}
          <div className="w-16 flex-shrink-0 border-r relative">
            {hours.map((hour) => (
              <div
                key={hour}
                className="absolute w-full text-right pr-2 text-xs text-muted-foreground"
                style={{ top: `${(hour - START_HOUR) * HOUR_HEIGHT}px`, height: `${HOUR_HEIGHT}px` }}
              >
                <span className="relative -top-2">{formatHour(hour)}</span>
              </div>
            ))}
          </div>

          {/* Day columns */}
          {days.map((day, dayIdx) => {
            const { timed } = getEventsForDay(events, day);
            const positioned = layoutOverlappingEvents(timed);
            const today = isToday(day);
            const now = new Date();
            const currentTimeTop = today
              ? ((now.getHours() - START_HOUR) * 60 + now.getMinutes()) / 60 * HOUR_HEIGHT
              : -1;

            return (
              <div
                key={dayIdx}
                className={`flex-1 border-r last:border-r-0 relative ${today ? "bg-primary/5" : ""}`}
              >
                {/* Hour grid lines */}
                {hours.map((hour) => (
                  <div
                    key={hour}
                    className="absolute w-full border-t border-dashed border-muted"
                    style={{ top: `${(hour - START_HOUR) * HOUR_HEIGHT}px` }}
                  />
                ))}

                {/* Current time indicator */}
                {today && currentTimeTop >= 0 && currentTimeTop <= hours.length * HOUR_HEIGHT && (
                  <div
                    className="absolute w-full border-t-2 border-red-500 z-10"
                    style={{ top: `${currentTimeTop}px` }}
                  >
                    <div className="absolute -left-1 -top-1.5 w-3 h-3 rounded-full bg-red-500" />
                  </div>
                )}

                {/* Events */}
                {positioned.map(({ event, column, totalColumns }) => {
                  const { top, height } = getEventPosition(event, day);
                  const colWidth = 100 / totalColumns;
                  return (
                    <EventBlock
                      key={event.id}
                      event={event}
                      top={top}
                      height={height}
                      left={`${column * colWidth}%`}
                      width={`${colWidth}%`}
                    />
                  );
                })}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
