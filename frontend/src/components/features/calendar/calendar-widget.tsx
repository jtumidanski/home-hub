import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "@/components/features/dashboard-widgets/widget-error";
import { CalendarDays, ChevronRight } from "lucide-react";
import { getLocalTodayRange, getLocalTodayStr } from "@/lib/date-utils";
import { useTenant } from "@/context/tenant-context";
import type { CalendarEvent } from "@/types/models/calendar";

function formatEventTime(startTime: string): string {
  const date = new Date(startTime);
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function sortEvents(events: CalendarEvent[]): CalendarEvent[] {
  return [...events].sort((a, b) => {
    if (a.attributes.allDay && !b.attributes.allDay) return -1;
    if (!a.attributes.allDay && b.attributes.allDay) return 1;
    return new Date(a.attributes.startTime).getTime() - new Date(b.attributes.startTime).getTime();
  });
}

export interface CalendarWidgetProps {
  /**
   * Optional planning horizon for future dashboard configurability. Accepted
   * on the prop contract so the widget registry can pass config through,
   * but not yet wired through the existing today-only query range.
   */
  horizonDays?: 1 | 3 | 7;
  /**
   * Optional all-day event filter. Accepted but not yet wired — all-day
   * events are always included in the current rendering.
   */
  includeAllDay?: boolean;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function CalendarWidget(_props: CalendarWidgetProps = {}) {
  const { household } = useTenant();
  const timezone = household?.attributes.timezone;
  const { start, end } = useMemo(() => getLocalTodayRange(timezone), [timezone]);
  const todayStr = useMemo(() => getLocalTodayStr(timezone), [timezone]);
  const { data, isLoading, isError } = useCalendarEvents(start, end);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-32" />
        </CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-5/6" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return <WidgetError message="Failed to load calendar events" />;
  }

  const events = sortEvents(
    (data?.data ?? []).filter((evt) => {
      if (!evt.attributes.allDay) return true;
      const startDate = evt.attributes.startTime.slice(0, 10);
      const endDate = evt.attributes.endTime.slice(0, 10);
      return todayStr >= startDate && todayStr <= endDate;
    }),
  );

  return (
    <Link to="/app/calendar" className="@container group/calendar block h-full rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card className="@container h-full motion-safe:transition-colors group-hover/calendar:bg-accent">
        <CardHeader>
          <CardTitle className="font-heading text-lg @xl:text-2xl">Calendar</CardTitle>
          <CardAction>
            <ChevronRight className="size-4 text-muted-foreground" aria-hidden="true" />
          </CardAction>
        </CardHeader>
        <CardContent>
          {events.length === 0 ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <CalendarDays className="size-5 @xl:size-6 shrink-0" />
              <p className="text-sm">No events today</p>
            </div>
          ) : (
            <div className="space-y-2">
              {events.map((event) => (
                <div key={event.id} className="flex items-start gap-2">
                  <span className="text-sm @xl:text-xl font-medium tabular-nums text-muted-foreground w-20 @xl:w-28 shrink-0">
                    {event.attributes.allDay ? "All Day" : formatEventTime(event.attributes.startTime)}
                  </span>
                  <div className="flex items-center gap-1.5 min-w-0">
                    <span
                      className="size-3 rounded-full shrink-0 ring-1 ring-border"
                      style={{ backgroundColor: event.attributes.userColor }}
                    />
                    <span className="text-sm @xl:text-xl truncate">{event.attributes.title}</span>
                    <span className="text-xs @xl:text-base text-muted-foreground truncate max-w-[30%] shrink-0">
                      {event.attributes.userDisplayName}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
