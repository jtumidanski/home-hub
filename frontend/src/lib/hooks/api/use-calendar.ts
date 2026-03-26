import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { calendarService } from "@/services/api/calendar";
import { useTenant } from "@/context/tenant-context";
import { createErrorFromUnknown, getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const calendarKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["calendar", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  connections: (tenant: Tenant | null, household: Household | null) =>
    [...calendarKeys.all(tenant, household), "connections"] as const,
  sources: (tenant: Tenant | null, household: Household | null, connectionId: string) =>
    [...calendarKeys.all(tenant, household), "sources", connectionId] as const,
  events: (tenant: Tenant | null, household: Household | null, start: string, end: string) =>
    [...calendarKeys.all(tenant, household), "events", start, end] as const,
};

export function useCalendarConnections() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: calendarKeys.connections(tenant, household),
    queryFn: () => calendarService.getConnections(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 60 * 1000,
    refetchOnWindowFocus: true,
  });
}

export function useCalendarSources(connectionId: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: calendarKeys.sources(tenant, household, connectionId ?? ""),
    queryFn: () => calendarService.getCalendarSources(tenant!, connectionId!),
    enabled: !!tenant?.id && !!household?.id && !!connectionId,
    staleTime: 60 * 1000,
  });
}

export function useCalendarEvents(start: string, end: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: calendarKeys.events(tenant, household, start, end),
    queryFn: () => calendarService.getEvents(tenant!, start, end),
    enabled: !!tenant?.id && !!household?.id && !!start && !!end,
    staleTime: 60 * 1000,
    refetchOnWindowFocus: true,
  });
}

export function useConnectGoogleCalendar() {
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (redirectUri: string) => calendarService.authorizeGoogle(tenant!, redirectUri),
    onSuccess: (response) => {
      window.location.href = response.data.attributes.authorizeUrl;
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to initiate Google Calendar connection"));
    },
  });
}

export function useDisconnectCalendar() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => calendarService.deleteConnection(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: calendarKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to disconnect calendar"));
    },
  });
}

export function useToggleCalendarSource() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ connectionId, calId, visible }: { connectionId: string; calId: string; visible: boolean }) =>
      calendarService.toggleCalendarSource(tenant!, connectionId, calId, visible),
    onSettled: (_data, _error, variables) => {
      qc.invalidateQueries({ queryKey: calendarKeys.sources(tenant, household, variables.connectionId) });
      qc.invalidateQueries({ queryKey: calendarKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update calendar visibility"));
    },
  });
}

export function useTriggerSync() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (connectionId: string) => calendarService.triggerSync(tenant!, connectionId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: calendarKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Sync triggered");
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "rate-limited") {
        toast.error("Sync rate limited — try again in 5 minutes");
      } else {
        toast.error(getErrorMessage(error, "Sync failed"));
      }
    },
  });
}
