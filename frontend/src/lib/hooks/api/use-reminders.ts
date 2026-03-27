import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { ReminderUpdateAttributes } from "@/types/models/reminder";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

// --- Key factory ---

export const reminderKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["reminders", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  lists: (tenant: Tenant | null, household: Household | null) =>
    [...reminderKeys.all(tenant, household), "list"] as const,
  details: (tenant: Tenant | null, household: Household | null) =>
    [...reminderKeys.all(tenant, household), "detail"] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...reminderKeys.details(tenant, household), id] as const,
  summary: (tenant: Tenant | null, household: Household | null) =>
    [...reminderKeys.all(tenant, household), "summary"] as const,
};

// --- Query hooks ---

export function useReminders() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: reminderKeys.lists(tenant, household),
    queryFn: () => productivityService.listReminders(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useReminderSummary() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: reminderKeys.summary(tenant, household),
    queryFn: () => productivityService.getReminderSummary(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateReminder() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: { title: string; notes?: string; scheduledFor: string; ownerUserId?: string | null }) =>
      productivityService.createReminder(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create reminder"));
    },
  });
}

export function useUpdateReminder() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: ReminderUpdateAttributes }) =>
      productivityService.updateReminder(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update reminder"));
    },
  });
}

export function useDeleteReminder() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteReminder(tenant!, id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: reminderKeys.lists(tenant, household) });
      const previous = qc.getQueryData(reminderKeys.lists(tenant, household));
      if (previous) {
        qc.setQueryData(reminderKeys.lists(tenant, household), {
          ...(previous as Record<string, unknown>),
          data: ((previous as { data: Array<{ id: string }> }).data ?? []).filter(
            (item) => item.id !== id,
          ),
        });
      }
      return { previous };
    },
    onError: (error, _id, context) => {
      if (context?.previous) {
        qc.setQueryData(reminderKeys.lists(tenant, household), context.previous);
      }
      toast.error(getErrorMessage(error, "Failed to delete reminder"));
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenant, household) });
    },
  });
}

export function useSnoozeReminder() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, minutes }: { id: string; minutes: number }) =>
      productivityService.snoozeReminder(tenant!, id, minutes),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to snooze reminder"));
    },
  });
}

export function useDismissReminder() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.dismissReminder(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to dismiss reminder"));
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateReminders() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.all(tenant, household) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.lists(tenant, household) }),
    invalidateSummary: () =>
      qc.invalidateQueries({ queryKey: reminderKeys.summary(tenant, household) }),
    invalidateReminder: (id: string) =>
      qc.invalidateQueries({ queryKey: reminderKeys.detail(tenant, household, id) }),
  };
}

// --- Prefetch helper ---

export function usePrefetchReminders() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();

  return {
    prefetch: () => {
      if (!tenant || !household) return;
      qc.prefetchQuery({
        queryKey: reminderKeys.lists(tenant, household),
        queryFn: () => productivityService.listReminders(tenant),
        staleTime: 5 * 60 * 1000,
      });
    },
    prefetchSummary: () => {
      if (!tenant || !household) return;
      qc.prefetchQuery({
        queryKey: reminderKeys.summary(tenant, household),
        queryFn: () => productivityService.getReminderSummary(tenant),
        staleTime: 60 * 1000,
      });
    },
  };
}
