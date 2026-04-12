import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
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

export function CalendarWidget() {
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
    return (
      <Card className="border-destructive">
        <CardContent className="py-3">
          <p className="text-sm text-destructive">Failed to load calendar events</p>
        </CardContent>
      </Card>
    );
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
    <Link to="/app/calendar" className="block h-full transition-opacity hover:opacity-80">
      <Card className="h-full">
        <CardHeader>
          <CardTitle className="text-sm font-medium">Calendar</CardTitle>
          <CardAction>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          </CardAction>
        </CardHeader>
        <CardContent>
          {events.length === 0 ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <CalendarDays className="h-5 w-5" />
              <p className="text-sm">No events today</p>
            </div>
          ) : (
            <div className="space-y-2">
              {events.map((event) => (
                <div key={event.id} className="flex items-start gap-2">
                  <span className="text-xs font-medium text-muted-foreground w-16 shrink-0 pt-0.5">
                    {event.attributes.allDay ? "All Day" : formatEventTime(event.attributes.startTime)}
                  </span>
                  <div className="flex items-center gap-1.5 min-w-0">
                    <span
                      className="h-2 w-2 rounded-full shrink-0"
                      style={{ backgroundColor: event.attributes.userColor }}
                      title={event.attributes.userDisplayName}
                    />
                    <span className="text-sm truncate">{event.attributes.title}</span>
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
