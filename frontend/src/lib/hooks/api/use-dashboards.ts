import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { dashboardService } from "@/services/api/dashboard";
import { useTenant } from "@/context/tenant-context";
import { createErrorFromUnknown, getErrorMessage } from "@/lib/api/errors";
import type { Layout } from "@/lib/dashboard/schema";
import type {
  DashboardCreateAttributes,
  DashboardOrderEntry,
  DashboardUpdateAttributes,
} from "@/types/models/dashboard";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const dashboardKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["dashboards", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  list: (tenant: Tenant | null, household: Household | null) =>
    [...dashboardKeys.all(tenant, household), "list"] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...dashboardKeys.all(tenant, household), "detail", id] as const,
};

// --- Query hooks ---

export function useDashboards() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: dashboardKeys.list(tenant, household),
    queryFn: () => dashboardService.listDashboards(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 30 * 1000,
    refetchOnWindowFocus: true,
  });
}

export function useDashboard(id: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: dashboardKeys.detail(tenant, household, id ?? ""),
    queryFn: () => dashboardService.getDashboard(tenant!, id!),
    enabled: !!tenant?.id && !!household?.id && !!id,
    staleTime: 30 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateDashboard() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: DashboardCreateAttributes) =>
      dashboardService.createDashboard(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Dashboard created");
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "validation") {
        toast.error(appError.message || "Invalid dashboard data");
      } else {
        toast.error(getErrorMessage(error, "Failed to create dashboard"));
      }
    },
  });
}

export function useUpdateDashboard() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: DashboardUpdateAttributes }) =>
      dashboardService.updateDashboard(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "validation") {
        toast.error(appError.message || "Invalid dashboard data");
      } else {
        toast.error(getErrorMessage(error, "Failed to update dashboard"));
      }
    },
  });
}

export function useDeleteDashboard() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => dashboardService.deleteDashboard(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Dashboard deleted");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete dashboard"));
    },
  });
}

export function useReorderDashboards() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (entries: DashboardOrderEntry[]) =>
      dashboardService.reorderDashboards(tenant!, entries),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to reorder dashboards"));
    },
  });
}

export function usePromoteDashboard() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => dashboardService.promoteDashboard(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Dashboard promoted to household");
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "conflict") {
        toast.error("Dashboard is already household-scoped");
      } else {
        toast.error(getErrorMessage(error, "Failed to promote dashboard"));
      }
    },
  });
}

export function useCopyDashboardToMine() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => dashboardService.copyDashboardToMine(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Dashboard copied");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to copy dashboard"));
    },
  });
}

export function useSeedDashboard() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ name, layout, key }: { name: string; layout: Layout; key?: string }) =>
      dashboardService.seedDashboard(tenant!, name, layout, key),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to seed dashboard"));
    },
  });
}
