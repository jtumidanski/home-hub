import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import type { TaskAttributes } from "@/types/models/task";

// --- Key factory ---

export const taskKeys = {
  all: (tenantId: string | null, householdId: string | null) =>
    ["tasks", tenantId ?? "no-tenant", householdId ?? "no-household"] as const,
  lists: (tenantId: string | null, householdId: string | null) =>
    [...taskKeys.all(tenantId, householdId), "list"] as const,
  details: (tenantId: string | null, householdId: string | null) =>
    [...taskKeys.all(tenantId, householdId), "detail"] as const,
  detail: (tenantId: string | null, householdId: string | null, id: string) =>
    [...taskKeys.details(tenantId, householdId), id] as const,
  summary: (tenantId: string | null, householdId: string | null) =>
    [...taskKeys.all(tenantId, householdId), "summary"] as const,
};

// --- Query hooks ---

export function useTasks() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: taskKeys.lists(tenantId, householdId),
    queryFn: () => productivityService.listTasks(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useTaskSummary() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: taskKeys.summary(tenantId, householdId),
    queryFn: () => productivityService.getTaskSummary(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (attrs: { title: string; notes?: string; dueOn?: string; rolloverEnabled?: boolean }) =>
      productivityService.createTask(tenantId!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenantId, householdId) });
    },
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Partial<TaskAttributes> }) =>
      productivityService.updateTask(tenantId!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenantId, householdId) });
    },
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteTask(tenantId!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenantId, householdId) });
    },
  });
}

export function useRestoreTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (taskId: string) => productivityService.restoreTask(tenantId!, taskId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenantId, householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenantId, householdId) });
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateTasks() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: taskKeys.all(tenantId, householdId) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: taskKeys.lists(tenantId, householdId) }),
    invalidateSummary: () =>
      qc.invalidateQueries({ queryKey: taskKeys.summary(tenantId, householdId) }),
    invalidateTask: (id: string) =>
      qc.invalidateQueries({ queryKey: taskKeys.detail(tenantId, householdId, id) }),
  };
}

// --- Prefetch helper ---

export function usePrefetchTasks() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();

  return {
    prefetch: () => {
      if (!tenantId || !householdId) return;
      qc.prefetchQuery({
        queryKey: taskKeys.lists(tenantId, householdId),
        queryFn: () => productivityService.listTasks(tenantId),
        staleTime: 5 * 60 * 1000,
      });
    },
    prefetchSummary: () => {
      if (!tenantId || !householdId) return;
      qc.prefetchQuery({
        queryKey: taskKeys.summary(tenantId, householdId),
        queryFn: () => productivityService.getTaskSummary(tenantId),
        staleTime: 60 * 1000,
      });
    },
  };
}
