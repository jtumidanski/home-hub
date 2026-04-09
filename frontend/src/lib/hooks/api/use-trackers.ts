import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { trackerService } from "@/services/api/tracker";
import { useTenant } from "@/context/tenant-context";
import { createErrorFromUnknown, getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const trackerKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["trackers", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  list: (tenant: Tenant | null, household: Household | null) =>
    [...trackerKeys.all(tenant, household), "list"] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...trackerKeys.all(tenant, household), "detail", id] as const,
  today: (tenant: Tenant | null, household: Household | null) =>
    [...trackerKeys.all(tenant, household), "today"] as const,
  month: (tenant: Tenant | null, household: Household | null, month: string) =>
    [...trackerKeys.all(tenant, household), "month", month] as const,
  report: (tenant: Tenant | null, household: Household | null, month: string) =>
    [...trackerKeys.all(tenant, household), "report", month] as const,
  entries: (tenant: Tenant | null, household: Household | null, month: string) =>
    [...trackerKeys.all(tenant, household), "entries", month] as const,
};

export function useTrackers() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: trackerKeys.list(tenant, household),
    queryFn: () => trackerService.getTrackers(tenant!),
    enabled: !!tenant?.id,
    staleTime: 30 * 1000,
  });
}

export function useTracker(id: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: trackerKeys.detail(tenant, household, id ?? ""),
    queryFn: () => trackerService.getTracker(tenant!, id!),
    enabled: !!tenant?.id && !!id,
    staleTime: 30 * 1000,
  });
}

export function useTrackerToday() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: trackerKeys.today(tenant, household),
    queryFn: () => trackerService.getToday(tenant!),
    enabled: !!tenant?.id,
    staleTime: 30 * 1000,
  });
}

export function useMonthSummary(month: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: trackerKeys.month(tenant, household, month),
    queryFn: () => trackerService.getMonthSummary(tenant!, month),
    enabled: !!tenant?.id && !!month,
    staleTime: 30 * 1000,
  });
}

export function useMonthReport(month: string, enabled: boolean) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: trackerKeys.report(tenant, household, month),
    queryFn: () => trackerService.getMonthReport(tenant!, month),
    enabled: !!tenant?.id && !!month && enabled,
    staleTime: 60 * 1000,
  });
}

export function useCreateTracker() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: {
      name: string;
      scale_type: string;
      scale_config: { min: number; max: number } | null;
      schedule: number[];
      color: string;
      sort_order?: number;
    }) => trackerService.createTracker(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Tracking item created");
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "conflict") {
        toast.error("A tracking item with this name already exists");
      } else {
        toast.error(getErrorMessage(error, "Failed to create tracking item"));
      }
    },
  });
}

export function useUpdateTracker() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: {
      id: string;
      attrs: {
        name?: string;
        color?: string;
        schedule?: number[];
        sort_order?: number;
        scale_config?: { min: number; max: number };
      };
    }) => trackerService.updateTracker(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update tracking item"));
    },
  });
}

export function useDeleteTracker() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => trackerService.deleteTracker(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Tracking item deleted");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete tracking item"));
    },
  });
}

export function usePutEntry() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ itemId, date, value, note }: {
      itemId: string;
      date: string;
      value: unknown;
      note?: string | null;
    }) => trackerService.putEntry(tenant!, itemId, date, value, note),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to save entry"));
    },
  });
}

export function useDeleteEntry() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ itemId, date }: { itemId: string; date: string }) =>
      trackerService.deleteEntry(tenant!, itemId, date),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
  });
}

export function useSkipEntry() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ itemId, date }: { itemId: string; date: string }) =>
      trackerService.skipEntry(tenant!, itemId, date),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to skip entry"));
    },
  });
}

export function useRemoveSkip() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ itemId, date }: { itemId: string; date: string }) =>
      trackerService.removeSkip(tenant!, itemId, date),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: trackerKeys.all(tenant, household) });
    },
  });
}
