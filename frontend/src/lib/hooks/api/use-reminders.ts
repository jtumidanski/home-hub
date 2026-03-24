import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";

export const reminderKeys = {
  all: (householdId: string | null) => ["reminders", householdId ?? "no-household"] as const,
  list: (householdId: string | null) => [...reminderKeys.all(householdId), "list"] as const,
  summary: (householdId: string | null) => [...reminderKeys.all(householdId), "summary"] as const,
};

export function useReminders() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: reminderKeys.list(householdId),
    queryFn: () => productivityService.listReminders(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (attrs: { title: string; notes?: string; scheduledFor: string }) =>
      productivityService.createReminder(tenantId!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
    },
  });
}

export function useUpdateReminder() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Record<string, unknown> }) =>
      productivityService.updateReminder(tenantId!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list(householdId) });
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
      qc.invalidateQueries({ queryKey: reminderKeys.list(householdId) });
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
      qc.invalidateQueries({ queryKey: reminderKeys.list(householdId) });
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
      qc.invalidateQueries({ queryKey: reminderKeys.list(householdId) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(householdId) });
    },
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
