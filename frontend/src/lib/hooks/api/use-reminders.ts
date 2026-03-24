import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";

export const reminderKeys = {
  list: ["reminders"] as const,
  summary: ["reminders", "summary"] as const,
};

export function useReminders() {
  return useQuery({
    queryKey: reminderKeys.list,
    queryFn: () => productivityService.listReminders(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateReminder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: productivityService.createReminder,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list });
      qc.invalidateQueries({ queryKey: reminderKeys.summary });
    },
  });
}

export function useUpdateReminder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Record<string, unknown> }) =>
      productivityService.updateReminder(id, attrs),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list });
      qc.invalidateQueries({ queryKey: reminderKeys.summary });
    },
  });
}

export function useDeleteReminder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: productivityService.deleteReminder,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list });
      qc.invalidateQueries({ queryKey: reminderKeys.summary });
    },
  });
}

export function useSnoozeReminder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, minutes }: { id: string; minutes: number }) =>
      productivityService.snoozeReminder(id, minutes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list });
      qc.invalidateQueries({ queryKey: reminderKeys.summary });
    },
  });
}

export function useDismissReminder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: productivityService.dismissReminder,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.list });
      qc.invalidateQueries({ queryKey: reminderKeys.summary });
    },
  });
}

export function useReminderSummary() {
  return useQuery({
    queryKey: reminderKeys.summary,
    queryFn: () => productivityService.getReminderSummary(),
    staleTime: 60 * 1000,
  });
}
