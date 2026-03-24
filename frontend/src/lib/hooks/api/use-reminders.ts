import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import type { ReminderAttributes } from "@/types/models/reminder";

// --- Key factory ---

export const reminderKeys = {
  all: (tenantId: string | null, householdId: string | null) =>
    ["reminders", tenantId ?? "no-tenant", householdId ?? "no-household"] as const,
  lists: (tenantId: string | null, householdId: string | null) =>
    [...reminderKeys.all(tenantId, householdId), "list"] as const,
  details: (tenantId: string | null, householdId: string | null) =>
    [...reminderKeys.all(tenantId, householdId), "detail"] as const,
  detail: (tenantId: string | null, householdId: string | null, id: string) =>
    [...reminderKeys.details(tenantId, householdId), id] as const,
  summary: (tenantId: string | null, householdId: string | null) =>
    [...reminderKeys.all(tenantId, householdId), "summary"] as const,
};

// --- Query hooks ---

export function useReminders() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: reminderKeys.lists(tenantId, householdId),
    queryFn: () => productivityService.listReminders(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useReminderSummary() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: reminderKeys.summary(tenantId, householdId),
    queryFn: () => productivityService.getReminderSummary(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (attrs: { title: string; notes?: string; scheduledFor: string }) =>
      productivityService.createReminder(tenantId!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenantId, householdId) });
    },
  });
}

export function useUpdateReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Partial<ReminderAttributes> }) =>
      productivityService.updateReminder(tenantId!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenantId, householdId) });
    },
  });
}

export function useDeleteReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteReminder(tenantId!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenantId, householdId) });
    },
  });
}

export function useSnoozeReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: ({ id, minutes }: { id: string; minutes: number }) =>
      productivityService.snoozeReminder(tenantId!, id, minutes),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenantId, householdId) });
    },
  });
}

export function useDismissReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.dismissReminder(tenantId!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenantId, householdId) });
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateReminders() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.all(tenantId, householdId) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenantId, householdId) }),
    invalidateSummary: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenantId, householdId) }),
    invalidateReminder: (id: string) =>
      qc.invalidateQueries({ queryKey: reminderKeys.detail(tenantId, householdId, id) }),
  };
}

// --- Prefetch helper ---

export function usePrefetchReminders() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();

  return {
    prefetch: () => {
      if (!tenantId || !householdId) return;
      qc.prefetchQuery({
        queryKey: reminderKeys.lists(tenantId, householdId),
        queryFn: () => productivityService.listReminders(tenantId),
        staleTime: 5 * 60 * 1000,
      });
    },
    prefetchSummary: () => {
      if (!tenantId || !householdId) return;
      qc.prefetchQuery({
        queryKey: reminderKeys.summary(tenantId, householdId),
        queryFn: () => productivityService.getReminderSummary(tenantId),
        staleTime: 60 * 1000,
      });
    },
  };
}
