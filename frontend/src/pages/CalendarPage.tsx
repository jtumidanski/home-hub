import { useState, useEffect, useMemo, useCallback, useSyncExternalStore } from "react";
import { useSearchParams } from "react-router-dom";
import { ChevronLeft, ChevronRight, Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ErrorCard } from "@/components/common/error-card";
import { CalendarGrid } from "@/components/features/calendar/calendar-grid";
import { ConnectCalendarButton } from "@/components/features/calendar/connect-calendar-button";
import { ConnectionStatus } from "@/components/features/calendar/connection-status";
import { CalendarSelectionPanel } from "@/components/features/calendar/calendar-selection-panel";
import { EventFormDialog } from "@/components/features/calendar/event-form-dialog";
import { RecurringScopeDialog } from "@/components/features/calendar/recurring-scope-dialog";
import { ReauthorizeBanner } from "@/components/features/calendar/reauthorize-banner";
import { useCalendarConnections, useCalendarEvents, useCalendarSources, useDeleteEvent } from "@/lib/hooks/api/use-calendar";
import { usePackages } from "@/lib/hooks/api/use-packages";
import { packagesToCalendarEvents } from "@/components/features/packages/package-calendar-overlay";
import { getStartOfWeek, formatDateRange } from "@/components/features/calendar/calendar-utils";
import { getErrorMessage } from "@/lib/api/errors";
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

  // Event CRUD state
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [prefilledStart, setPrefilledStart] = useState<Date | undefined>();
  const [editEvent, setEditEvent] = useState<CalendarEvent | null>(null);
  const [editScope, setEditScope] = useState<"single" | "all">("single");
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<CalendarEvent | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleteScope, setDeleteScope] = useState<"single" | "all">("single");
  const [recurringAction, setRecurringAction] = useState<{ event: CalendarEvent; action: "edit" | "delete" } | null>(null);

  const deleteEvent = useDeleteEvent();

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

  const hasWriteAccess = useMemo(
    () => connections.some((c) => c.attributes.writeAccess && c.attributes.status === "connected"),
    [connections],
  );
  const needsReauth = useMemo(
    () => connections.some((c) => !c.attributes.writeAccess && c.attributes.status === "connected"),
    [connections],
  );

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

  const handleAddEvent = useCallback(() => {
    setPrefilledStart(undefined);
    setEditEvent(null);
    setShowCreateDialog(true);
  }, []);

  const handleSlotClick = useCallback((date: Date) => {
    setPrefilledStart(date);
    setEditEvent(null);
    setShowCreateDialog(true);
  }, []);

  const handleEditEvent = useCallback((evt: CalendarEvent) => {
    if (evt.attributes.isRecurring) {
      setRecurringAction({ event: evt, action: "edit" });
    } else {
      setEditEvent(evt);
      setEditScope("single");
      setShowEditDialog(true);
    }
  }, []);

  const handleDeleteEvent = useCallback((evt: CalendarEvent) => {
    if (evt.attributes.isRecurring) {
      setRecurringAction({ event: evt, action: "delete" });
    } else {
      setDeleteTarget(evt);
      setDeleteScope("single");
      setShowDeleteConfirm(true);
    }
  }, []);

  const handleRecurringScopeSelect = useCallback((scope: "single" | "all") => {
    if (!recurringAction) return;
    const { event, action } = recurringAction;
    setRecurringAction(null);

    if (action === "edit") {
      setEditEvent(event);
      setEditScope(scope);
      setShowEditDialog(true);
    } else {
      setDeleteTarget(event);
      setDeleteScope(scope);
      setShowDeleteConfirm(true);
    }
  }, [recurringAction]);

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteEvent.mutateAsync({
        connectionId: deleteTarget.attributes.connectionId,
        eventId: deleteTarget.id,
        scope: deleteScope,
      });
      toast.success("Event deleted");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to delete event"));
    }
    setShowDeleteConfirm(false);
    setDeleteTarget(null);
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

          {hasWriteAccess && (
            <Button size="sm" onClick={handleAddEvent}>
              <Plus className="h-4 w-4 mr-1" />
              Add Event
            </Button>
          )}

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

      {/* Re-authorization banner */}
      {needsReauth && <ReauthorizeBanner />}

      {/* Connection status rows and household color legend */}
      {(connections.length > 0 || eventUsers.length > 0) && (
        <div className="space-y-1">
          {connections.map((conn) => (
            <ConnectionStatus key={conn.id} connection={conn} />
          ))}
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
            <CalendarGrid
              weekStart={weekStart}
              events={events}
              dayCount={dayCount}
              hasWriteAccess={hasWriteAccess}
              onSlotClick={handleSlotClick}
              onEditEvent={handleEditEvent}
              onDeleteEvent={handleDeleteEvent}
            />
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

      {/* Create event dialog */}
      <EventFormDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        connections={connections}
        sources={sources}
        prefilledStart={prefilledStart}
      />

      {/* Edit event dialog */}
      <EventFormDialog
        open={showEditDialog}
        onOpenChange={setShowEditDialog}
        connections={connections}
        sources={sources}
        editEvent={editEvent}
        editScope={editScope}
      />

      {/* Recurring scope dialog */}
      <RecurringScopeDialog
        open={!!recurringAction}
        onOpenChange={() => setRecurringAction(null)}
        action={recurringAction?.action ?? "edit"}
        onSelect={handleRecurringScopeSelect}
      />

      {/* Delete confirmation dialog */}
      <Dialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete event</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Are you sure you want to delete &ldquo;{deleteTarget?.attributes.title}&rdquo;?
            This will also remove it from Google Calendar.
          </p>
          <div className="flex gap-2 justify-end">
            <Button variant="outline" size="sm" onClick={() => setShowDeleteConfirm(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={confirmDelete}
              disabled={deleteEvent.isPending}
            >
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
