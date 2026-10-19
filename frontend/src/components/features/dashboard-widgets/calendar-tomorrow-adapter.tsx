// Read-only widget — no mutations. See PRD §4.4.
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";
import { getEventsForDateStr } from "@/components/features/calendar/calendar-utils";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "./widget-error";
import { CalendarDays, ChevronRight } from "lucide-react";
import type { CalendarEvent } from "@/types/models/calendar";

export interface CalendarTomorrowConfig {
  includeAllDay: boolean;
  limit: number;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

// Backend query window. We fetch a generous ±1 day around the target UTC date
// rather than the exact local-midnight span, because all-day events are stored
// at UTC midnight and a tz-shifted window would miss the target day's own
// all-day events while catching the next day's. The precise day selection is
// done client-side by getEventsForDateStr, which handles all-day (date-string)
// and timed (tz-aware) events correctly.
function tomorrowRange(tomorrow: string): { start: string; end: string } {
  const parts = tomorrow.split("-").map(Number);
  const y = parts[0] ?? 1970;
  const m = parts[1] ?? 1;
  const d = parts[2] ?? 1;
  const dayMs = Date.UTC(y, m - 1, d);
  const start = new Date(dayMs - 24 * 60 * 60 * 1000).toISOString();
  const end = new Date(dayMs + 2 * 24 * 60 * 60 * 1000).toISOString();
  return { start, end };
}

export function CalendarTomorrowAdapter({ config }: { config: CalendarTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { start, end } = useMemo(() => tomorrowRange(tomorrow), [tomorrow]);
  const { data, isLoading, isError } = useCalendarEvents(start, end);

  if (isLoading) {
    return (
      <Card className="@container h-full">
        <CardHeader><Skeleton className="h-5 w-32" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <WidgetError message="Failed to load calendar" />
    );
  }

  const all = (data?.data ?? []) as CalendarEvent[];
  const { allDay, timed } = getEventsForDateStr(all, tomorrow, household?.attributes.timezone);
  const filtered = config.includeAllDay ? [...allDay, ...timed] : timed;
  const sorted = filtered.sort((a, b) => {
    if (a.attributes.allDay && !b.attributes.allDay) return -1;
    if (!a.attributes.allDay && b.attributes.allDay) return 1;
    return a.attributes.startTime.localeCompare(b.attributes.startTime);
  });
  const visible = sorted.slice(0, config.limit);
  const remainder = sorted.length - visible.length;

  return (
    <Card className="@container h-full">
      <CardHeader>
        <CardTitle className="font-heading text-lg @xl:text-2xl">
          <Link to="/app/calendar" className="hover:underline">Tomorrow</Link>
        </CardTitle>
        <CardAction><Link to="/app/calendar" aria-label="Open calendar" className="inline-flex size-6 items-center justify-center rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring coarse:size-11"><ChevronRight aria-hidden="true" className="size-4 @xl:size-6 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {visible.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CalendarDays className="size-5 @xl:size-6" />
            <p className="text-sm @xl:text-xl">No events tomorrow</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {visible.map((e) => (
              <li key={e.id} className="flex items-start gap-2 text-sm @xl:text-xl">
                <span className="text-sm @xl:text-base font-medium text-muted-foreground w-20 shrink-0 pt-0.5">
                  {e.attributes.allDay ? "All Day" : formatTime(e.attributes.startTime)}
                </span>
                <span className="flex items-center gap-1.5 min-w-0">
                  <span aria-hidden="true" className="size-3 rounded-full shrink-0 border border-border ring-1 ring-background" style={{ backgroundColor: e.attributes.userColor }} />
                  <span className="truncate">{e.attributes.title}</span>
                  <span className="shrink-0 text-sm @xl:text-base text-muted-foreground">({e.attributes.userDisplayName})</span>
                </span>
              </li>
            ))}
            {remainder > 0 && (
              <li className="text-sm @xl:text-base text-muted-foreground">+{remainder} more</li>
            )}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
