// Read-only widget — no mutations. See PRD §4.4.
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useCalendarEvents } from "@/lib/hooks/api/use-calendar";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CalendarDays, ChevronRight } from "lucide-react";
import type { CalendarEvent } from "@/types/models/calendar";

export interface CalendarTomorrowConfig {
  includeAllDay: boolean;
  limit: number;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function tomorrowRange(tomorrow: string): { start: string; end: string } {
  const parts = tomorrow.split("-").map(Number);
  const y = parts[0] ?? 1970;
  const m = parts[1] ?? 1;
  const d = parts[2] ?? 1;
  const start = new Date(y, m - 1, d, 0, 0, 0, 0).toISOString();
  const end = new Date(y, m - 1, d, 23, 59, 59, 999).toISOString();
  return { start, end };
}

export function CalendarTomorrowAdapter({ config }: { config: CalendarTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { start, end } = useMemo(() => tomorrowRange(tomorrow), [tomorrow]);
  const { data, isLoading, isError } = useCalendarEvents(start, end);

  if (isLoading) {
    return (
      <Card className="h-full">
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
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load calendar</p></CardContent>
      </Card>
    );
  }

  const all = (data?.data ?? []) as CalendarEvent[];
  const filtered = all.filter((e) => config.includeAllDay || !e.attributes.allDay);
  const sorted = filtered.sort((a, b) => {
    if (a.attributes.allDay && !b.attributes.allDay) return -1;
    if (!a.attributes.allDay && b.attributes.allDay) return 1;
    return a.attributes.startTime.localeCompare(b.attributes.startTime);
  });
  const visible = sorted.slice(0, config.limit);
  const remainder = sorted.length - visible.length;

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/calendar" className="hover:underline">Tomorrow</Link>
        </CardTitle>
        <CardAction><Link to="/app/calendar"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {visible.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CalendarDays className="h-5 w-5" />
            <p className="text-sm">No events tomorrow</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {visible.map((e) => (
              <li key={e.id} className="flex items-start gap-2 text-sm">
                <span className="text-xs font-medium text-muted-foreground w-16 shrink-0 pt-0.5">
                  {e.attributes.allDay ? "All Day" : formatTime(e.attributes.startTime)}
                </span>
                <span className="flex items-center gap-1.5 min-w-0">
                  <span className="h-2 w-2 rounded-full shrink-0" style={{ backgroundColor: e.attributes.userColor }} />
                  <span className="truncate">{e.attributes.title}</span>
                </span>
              </li>
            ))}
            {remainder > 0 && (
              <li className="text-xs text-muted-foreground">+{remainder} more</li>
            )}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
