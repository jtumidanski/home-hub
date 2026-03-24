import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import type { ReminderAttributes } from "@/types/models/reminder";

// --- Key factory ---

export const reminderKeys = {
  all: (householdId: string | null) =>
    ["reminders", householdId ?? "no-household"] as const,
  lists: (householdId: string | null) =>
    [...reminderKeys.all(householdId), "list"] as const,
  list: (householdId: string | null) =>
    [...reminderKeys.lists(householdId)] as const,
  details: (householdId: string | null) =>
    [...reminderKeys.all(householdId), "detail"] as const,
  detail: (householdId: string | null, id: string) =>
    [...reminderKeys.details(householdId), id] as const,
  summary: (householdId: string | null) =>
    [...reminderKeys.all(householdId), "summary"] as const,
};

// --- Query hooks ---

export function useReminders() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: reminderKeys.list(householdId),
    queryFn: () => productivityService.listReminders(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useReminderSummary() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: reminderKeys.summary(householdId),
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
      qc.invalidateQueries({ queryKey: reminderKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
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
      qc.invalidateQueries({ queryKey: reminderKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
    },
  });
}

export function useDeleteReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteReminder(tenantId!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
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
      qc.invalidateQueries({ queryKey: reminderKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
    },
  });
}

export function useDismissReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.dismissReminder(tenantId!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateReminders() {
  const qc = useQueryClient();
  const { householdId } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.all(householdId) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.lists(householdId) }),
    invalidateSummary: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) }),
    invalidateReminder: (id: string) =>
      qc.invalidateQueries({ queryKey: reminderKeys.detail(householdId, id) }),
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
        queryKey: reminderKeys.list(householdId),
        queryFn: () => productivityService.listReminders(tenantId),
        staleTime: 5 * 60 * 1000,
      });
    },
    prefetchSummary: () => {
      if (!tenantId || !householdId) return;
      qc.prefetchQuery({
        queryKey: reminderKeys.summary(householdId),
        queryFn: () => productivityService.getReminderSummary(tenantId),
        staleTime: 60 * 1000,
      });
    },
  };
}
