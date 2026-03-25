import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { TaskUpdateAttributes } from "@/types/models/task";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

// --- Key factory ---

export const taskKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["tasks", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  lists: (tenant: Tenant | null, household: Household | null) =>
    [...taskKeys.all(tenant, household), "list"] as const,
  details: (tenant: Tenant | null, household: Household | null) =>
    [...taskKeys.all(tenant, household), "detail"] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...taskKeys.details(tenant, household), id] as const,
  summary: (tenant: Tenant | null, household: Household | null) =>
    [...taskKeys.all(tenant, household), "summary"] as const,
};

// --- Query hooks ---

export function useTasks() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: taskKeys.lists(tenant, household),
    queryFn: () => productivityService.listTasks(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useTaskSummary() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: taskKeys.summary(tenant, household),
    queryFn: () => productivityService.getTaskSummary(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateTask() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: { title: string; notes?: string; dueOn?: string; rolloverEnabled?: boolean }) =>
      productivityService.createTask(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create task"));
    },
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: TaskUpdateAttributes }) =>
      productivityService.updateTask(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update task"));
    },
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteTask(tenant!, id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: taskKeys.lists(tenant, household) });
      const previous = qc.getQueryData(taskKeys.lists(tenant, household));
      if (previous) {
        qc.setQueryData(taskKeys.lists(tenant, household), {
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
        qc.setQueryData(taskKeys.lists(tenant, household), context.previous);
      }
      toast.error(getErrorMessage(error, "Failed to delete task"));
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenant, household) });
    },
  });
}

export function useRestoreTask() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (taskId: string) => productivityService.restoreTask(tenant!, taskId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to restore task"));
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateTasks() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: taskKeys.all(tenant, household) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenant, household) }),
    invalidateSummary: () =>
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenant, household) }),
    invalidateTask: (id: string) =>
      qc.invalidateQueries({ queryKey: taskKeys.detail(tenant, household, id) }),
  };
}

// --- Prefetch helper ---

export function usePrefetchTasks() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();

  return {
    prefetch: () => {
      if (!tenant || !household) return;
      qc.prefetchQuery({
        queryKey: taskKeys.lists(tenant, household),
        queryFn: () => productivityService.listTasks(tenant),
        staleTime: 5 * 60 * 1000,
      });
    },
    prefetchSummary: () => {
      if (!tenant || !household) return;
      qc.prefetchQuery({
        queryKey: taskKeys.summary(tenant, household),
        queryFn: () => productivityService.getTaskSummary(tenant),
        staleTime: 60 * 1000,
      });
    },
  };
}
