import { useState, useEffect, useMemo } from "react";
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
import { UserLegend } from "@/components/features/calendar/user-legend";
import { useCalendarConnections, useCalendarEvents, useCalendarSources } from "@/lib/hooks/api/use-calendar";
import { getStartOfWeek, formatDateRange } from "@/components/features/calendar/calendar-utils";
import type { CalendarConnection, CalendarEvent } from "@/types/models/calendar";

export function CalendarPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [weekStart, setWeekStart] = useState(() => getStartOfWeek(new Date()));
  const [showSources, setShowSources] = useState(false);

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
    end.setDate(end.getDate() + 7);
    return end;
  }, [weekStart]);

  const startISO = weekStart.toISOString();
  const endISO = weekEnd.toISOString();

  const connectionsQuery = useCalendarConnections();
  const eventsQuery = useCalendarEvents(startISO, endISO);

  const connections = (connectionsQuery.data?.data ?? []) as CalendarConnection[];
  const events = (eventsQuery.data?.data ?? []) as CalendarEvent[];
  const activeConnection = connections.find((c) => c.attributes.status === "connected");

  const sourcesQuery = useCalendarSources(activeConnection?.id ?? null);
  const sources = sourcesQuery.data?.data ?? [];

  const goToday = () => setWeekStart(getStartOfWeek(new Date()));
  const goPrev = () => {
    const prev = new Date(weekStart);
    prev.setDate(prev.getDate() - 7);
    setWeekStart(prev);
  };
  const goNext = () => {
    const next = new Date(weekStart);
    next.setDate(next.getDate() + 7);
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

          {connections.length === 0 ? (
            <ConnectCalendarButton />
          ) : (
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

      {/* Connection status and legend */}
      {connections.length > 0 && (
        <div className="space-y-2">
          {connections.map((conn) => (
            <ConnectionStatus key={conn.id} connection={conn} />
          ))}
          <UserLegend connections={connections} />
        </div>
      )}

      {/* Source selection panel */}
      {showSources && activeConnection && (
        <div className="border rounded-lg p-4">
          <CalendarSelectionPanel connectionId={activeConnection.id} sources={sources} />
          {connections.length === 0 && <ConnectCalendarButton />}
        </div>
      )}

      {/* Empty state */}
      {connections.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center text-center py-16">
          <h2 className="text-lg font-medium mb-2">No calendars connected</h2>
          <p className="text-muted-foreground mb-6 max-w-md">
            Connect your Google Calendar to see your events on the household calendar. All household members can connect their own calendars for a merged view.
          </p>
          <ConnectCalendarButton />
        </div>
      ) : (
        /* Calendar grid */
        <div className="flex-1 min-h-0">
          {eventsQuery.isLoading ? (
            <Skeleton className="h-full w-full" />
          ) : (
            <CalendarGrid weekStart={weekStart} events={events} />
          )}
        </div>
      )}
    </div>
  );
}
