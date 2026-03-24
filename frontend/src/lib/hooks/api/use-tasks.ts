import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";
import type { TaskAttributes } from "@/types/models/task";

// --- Key factory ---

export const taskKeys = {
  all: (householdId: string | null) =>
    ["tasks", householdId ?? "no-household"] as const,
  lists: (householdId: string | null) =>
    [...taskKeys.all(householdId), "list"] as const,
  list: (householdId: string | null) =>
    [...taskKeys.lists(householdId)] as const,
  details: (householdId: string | null) =>
    [...taskKeys.all(householdId), "detail"] as const,
  detail: (householdId: string | null, id: string) =>
    [...taskKeys.details(householdId), id] as const,
  summary: (householdId: string | null) =>
    [...taskKeys.all(householdId), "summary"] as const,
};

// --- Query hooks ---

export function useTasks() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: taskKeys.list(householdId),
    queryFn: () => productivityService.listTasks(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useTaskSummary() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: taskKeys.summary(householdId),
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
      qc.invalidateQueries({ queryKey: taskKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) });
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
      qc.invalidateQueries({ queryKey: taskKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) });
    },
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (id: string) => productivityService.deleteTask(tenantId!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) });
    },
  });
}

export function useRestoreTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (taskId: string) => productivityService.restoreTask(tenantId!, taskId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.lists(householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) });
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateTasks() {
  const qc = useQueryClient();
  const { householdId } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: taskKeys.all(householdId) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: taskKeys.lists(householdId) }),
    invalidateSummary: () =>
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) }),
    invalidateTask: (id: string) =>
      qc.invalidateQueries({ queryKey: taskKeys.detail(householdId, id) }),
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
        queryKey: taskKeys.list(householdId),
        queryFn: () => productivityService.listTasks(tenantId),
        staleTime: 5 * 60 * 1000,
      });
    },
    prefetchSummary: () => {
      if (!tenantId || !householdId) return;
      qc.prefetchQuery({
        queryKey: taskKeys.summary(householdId),
        queryFn: () => productivityService.getTaskSummary(tenantId),
        staleTime: 60 * 1000,
      });
    },
  };
}
