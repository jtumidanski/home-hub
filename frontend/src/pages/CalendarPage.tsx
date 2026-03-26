import { useState, useEffect, useMemo, useSyncExternalStore } from "react";
import { useSearchParams } from "react-router-dom";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorCard } from "@/components/common/error-card";
import { CalendarGrid } from "@/components/features/calendar/calendar-grid";
import { ConnectCalendarButton } from "@/components/features/calendar/connect-calendar-button";
import { ConnectionStatus } from "@/components/features/calendar/connection-status";
import { CalendarSelectionPanel } from "@/components/features/calendar/calendar-selection-panel";
import { useCalendarConnections, useCalendarEvents, useCalendarSources } from "@/lib/hooks/api/use-calendar";
import { usePackages } from "@/lib/hooks/api/use-packages";
import { packagesToCalendarEvents } from "@/components/features/packages/package-calendar-overlay";
import { getStartOfWeek, formatDateRange } from "@/components/features/calendar/calendar-utils";
import type { CalendarConnection, CalendarEvent } from "@/types/models/calendar";
import type { Package } from "@/types/models/package";

const MD_BREAKPOINT = 768;
const mql = typeof window !== "undefined" ? window.matchMedia(`(min-width: ${MD_BREAKPOINT}px)`) : null;
function subscribeMedia(cb: () => void) {
  mql?.addEventListener("change", cb);
  return () => mql?.removeEventListener("change", cb);
}
function getIsDesktop() { return mql?.matches ?? true; }

export function CalendarPage() {
  const isDesktop = useSyncExternalStore(subscribeMedia, getIsDesktop, () => true);
  const dayCount = isDesktop ? 7 : 3;

  const [searchParams, setSearchParams] = useSearchParams();
  const [weekStart, setWeekStart] = useState(() => getStartOfWeek(new Date()));
  const [showSources, setShowSources] = useState(false);

  // Reset to a sensible start when switching between mobile and desktop
  useEffect(() => {
    if (isDesktop) {
      setWeekStart(getStartOfWeek(new Date()));
    } else {
      const today = new Date();
      const start = new Date(today);
      start.setDate(start.getDate() - 1);
      start.setHours(0, 0, 0, 0);
      setWeekStart(start);
    }
  }, [isDesktop]);

  useEffect(() => {
    if (searchParams.get("connected") === "true") {
      toast.success("Google Calendar connected! Your events will appear shortly.");
      setSearchParams({}, { replace: true });
    }
    const error = searchParams.get("error");
    if (error) {
      const messages: Record<string, string> = {
        auth_failed: "Google Calendar authorization failed. Please try again.",
        invalid_state: "Authorization session expired. Please try again.",
        already_connected: "You already have a Google Calendar connected.",
        missing_params: "Authorization failed. Please try again.",
        internal: "An internal error occurred. Please try again.",
      };
      toast.error(messages[error] ?? "An error occurred connecting your calendar.");
      setSearchParams({}, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const weekEnd = useMemo(() => {
    const end = new Date(weekStart);
    end.setDate(end.getDate() + dayCount);
    return end;
  }, [weekStart, dayCount]);

  const startISO = weekStart.toISOString();
  const endISO = weekEnd.toISOString();

  const connectionsQuery = useCalendarConnections();
  const eventsQuery = useCalendarEvents(startISO, endISO);
  const packagesQuery = usePackages("filter[status]=pre_transit,in_transit,out_for_delivery&filter[hasEta]=true");

  const connections = (connectionsQuery.data?.data ?? []) as CalendarConnection[];
  const calendarEvents = (eventsQuery.data?.data ?? []) as CalendarEvent[];
  const packages = (packagesQuery.data?.data ?? []) as Package[];
  const packageEvents = useMemo(() => packagesToCalendarEvents(packages), [packages]);
  const events = useMemo(() => [...calendarEvents, ...packageEvents], [calendarEvents, packageEvents]);
  const hasCalendar = connections.length > 0 || events.length > 0;

  // Derive unique users from events so all household members see the color legend,
  // even if they haven't connected their own calendar.
  const eventUsers = useMemo(() => {
    const seen = new Map<string, { displayName: string; color: string }>();
    for (const evt of events) {
      const { userDisplayName, userColor } = evt.attributes;
      if (!seen.has(userDisplayName)) {
        seen.set(userDisplayName, { displayName: userDisplayName, color: userColor });
      }
    }
    return Array.from(seen.values());
  }, [events]);
  const activeConnection = connections.find((c) => c.attributes.status === "connected");

  const sourcesQuery = useCalendarSources(activeConnection?.id ?? null);
  const sources = sourcesQuery.data?.data ?? [];

  const goToday = () => {
    const today = new Date();
    if (isDesktop) {
      setWeekStart(getStartOfWeek(today));
    } else {
      // On mobile, center today: show yesterday, today, tomorrow
      const start = new Date(today);
      start.setDate(start.getDate() - 1);
      start.setHours(0, 0, 0, 0);
      setWeekStart(start);
    }
  };
  const goPrev = () => {
    const prev = new Date(weekStart);
    prev.setDate(prev.getDate() - dayCount);
    setWeekStart(prev);
  };
  const goNext = () => {
    const next = new Date(weekStart);
    next.setDate(next.getDate() + dayCount);
    setWeekStart(next);
  };

  if (connectionsQuery.isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-[500px] w-full" />
      </div>
    );
  }

  if (connectionsQuery.isError) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load calendar. Try refreshing the page." />
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 flex flex-col h-full gap-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl md:text-2xl font-semibold">Calendar</h1>
          <p className="text-sm text-muted-foreground">{formatDateRange(weekStart, weekEnd)}</p>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center border rounded-md">
            <Button variant="ghost" size="sm" onClick={goPrev}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm" onClick={goToday} className="px-3">
              Today
            </Button>
            <Button variant="ghost" size="sm" onClick={goNext}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>

          {connections.length === 0 && <ConnectCalendarButton />}
          {activeConnection && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowSources(!showSources)}
            >
              Settings
            </Button>
          )}
        </div>
      </div>

      {/* Connection status rows (current user's connections) and household color legend */}
      {(connections.length > 0 || eventUsers.length > 0) && (
        <div className="space-y-1">
          {connections.map((conn) => (
            <ConnectionStatus key={conn.id} connection={conn} />
          ))}
          {/* Show legend for other household members' calendars (not already shown in ConnectionStatus) */}
          {eventUsers
            .filter((u) => !connections.some((c) => c.attributes.userDisplayName === u.displayName))
            .map((u) => (
              <div key={u.displayName} className="flex items-center gap-3 text-sm">
                <div
                  className="w-3 h-3 rounded-full flex-shrink-0"
                  style={{ backgroundColor: u.color }}
                />
                <span className="text-muted-foreground">{u.displayName}</span>
              </div>
            ))}
        </div>
      )}

      {/* Source selection panel */}
      {showSources && activeConnection && (
        <div className="border rounded-lg p-4">
          <CalendarSelectionPanel connectionId={activeConnection.id} sources={sources} />
        </div>
      )}

      {/* Calendar grid or empty state */}
      {hasCalendar ? (
        <div className="flex-1 min-h-0">
          {eventsQuery.isLoading ? (
            <Skeleton className="h-full w-full" />
          ) : (
            <CalendarGrid weekStart={weekStart} events={events} dayCount={dayCount} />
          )}
        </div>
      ) : (
        <div className="flex-1 flex flex-col items-center justify-center text-center py-16">
          <h2 className="text-lg font-medium mb-2">No calendars connected</h2>
          <p className="text-muted-foreground mb-6 max-w-md">
            Connect your Google Calendar to see your events on the household calendar. All household members can connect their own calendars for a merged view.
          </p>
          <ConnectCalendarButton />
        </div>
      )}
    </div>
  );
}
