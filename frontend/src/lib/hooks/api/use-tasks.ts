import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import type { TaskAttributes } from "@/types/models/task";
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
  });
}

export function useTaskSummary() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: taskKeys.summary(tenant, household),
    queryFn: () => productivityService.getTaskSummary(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 60 * 1000,
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
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Partial<TaskAttributes> }) =>
      productivityService.updateTask(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenant, household) });
    },
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteTask(tenant!, id),
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
