import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";
import { useTenant } from "@/context/tenant-context";

export const taskKeys = {
  all: (householdId: string | null) => ["tasks", householdId ?? "no-household"] as const,
  list: (householdId: string | null) => [...taskKeys.all(householdId), "list"] as const,
  detail: (householdId: string | null, id: string) => [...taskKeys.all(householdId), id] as const,
  summary: (householdId: string | null) => [...taskKeys.all(householdId), "summary"] as const,
};

export function useTasks() {
  const { tenantId, householdId } = useTenant();
  return useQuery({
    queryKey: taskKeys.list(householdId),
    queryFn: () => productivityService.listTasks(tenantId!),
    enabled: !!tenantId && !!householdId,
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: (attrs: { title: string; notes?: string; dueOn?: string; rolloverEnabled?: boolean }) =>
      productivityService.createTask(tenantId!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.list(householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) });
    },
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  const { tenantId, householdId } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Record<string, unknown> }) =>
      productivityService.updateTask(tenantId!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: taskKeys.list(householdId) });
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
      qc.invalidateQueries({ queryKey: taskKeys.list(householdId) });
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
      qc.invalidateQueries({ queryKey: taskKeys.list(householdId) });
      qc.invalidateQueries({ queryKey: taskKeys.summary(householdId) });
    },
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
