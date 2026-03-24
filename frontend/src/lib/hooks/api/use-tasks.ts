import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { productivityService } from "@/services/api/productivity";

export const taskKeys = {
  list: ["tasks"] as const,
  detail: (id: string) => ["tasks", id] as const,
  summary: ["tasks", "summary"] as const,
};

export function useTasks() {
  return useQuery({
    queryKey: taskKeys.list,
    queryFn: () => productivityService.listTasks(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: productivityService.createTask,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: taskKeys.list });
      qc.invalidateQueries({ queryKey: taskKeys.summary });
    },
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Record<string, unknown> }) =>
      productivityService.updateTask(id, attrs),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: taskKeys.list });
      qc.invalidateQueries({ queryKey: taskKeys.summary });
    },
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: productivityService.deleteTask,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: taskKeys.list });
      qc.invalidateQueries({ queryKey: taskKeys.summary });
    },
  });
}

export function useRestoreTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: productivityService.restoreTask,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: taskKeys.list });
      qc.invalidateQueries({ queryKey: taskKeys.summary });
    },
  });
}

export function useTaskSummary() {
  return useQuery({
    queryKey: taskKeys.summary,
    queryFn: () => productivityService.getTaskSummary(),
    staleTime: 60 * 1000,
  });
}
