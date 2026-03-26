import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { packageService } from "@/services/api/package";
import { useTenant } from "@/context/tenant-context";
import { createErrorFromUnknown, getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const packageKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["packages", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  list: (tenant: Tenant | null, household: Household | null, params?: string) =>
    [...packageKeys.all(tenant, household), "list", params ?? ""] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...packageKeys.all(tenant, household), "detail", id] as const,
  summary: (tenant: Tenant | null, household: Household | null) =>
    [...packageKeys.all(tenant, household), "summary"] as const,
};

export function usePackages(params?: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: packageKeys.list(tenant, household, params),
    queryFn: () => packageService.getPackages(tenant!, params),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 30 * 1000,
    refetchOnWindowFocus: true,
  });
}

export function usePackage(id: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: packageKeys.detail(tenant, household, id ?? ""),
    queryFn: () => packageService.getPackage(tenant!, id!),
    enabled: !!tenant?.id && !!household?.id && !!id,
    staleTime: 30 * 1000,
  });
}

export function usePackageSummary() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: packageKeys.summary(tenant, household),
    queryFn: () => packageService.getSummary(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 60 * 1000,
    refetchOnWindowFocus: true,
  });
}

export function useCreatePackage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: {
      trackingNumber: string;
      carrier: string;
      label?: string;
      notes?: string;
      private?: boolean;
    }) => packageService.createPackage(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: packageKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Package added");
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "conflict") {
        toast.error("This tracking number already exists in your household");
      } else if (appError.type === "validation") {
        toast.error(appError.message || "Invalid package data");
      } else {
        toast.error(getErrorMessage(error, "Failed to add package"));
      }
    },
  });
}

export function useUpdatePackage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: {
      id: string;
      attrs: { label?: string; notes?: string; carrier?: string; private?: boolean };
    }) => packageService.updatePackage(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: packageKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update package"));
    },
  });
}

export function useDeletePackage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => packageService.deletePackage(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: packageKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Package deleted");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete package"));
    },
  });
}

export function useArchivePackage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => packageService.archivePackage(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: packageKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Package archived");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to archive package"));
    },
  });
}

export function useUnarchivePackage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => packageService.unarchivePackage(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: packageKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Package restored");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to unarchive package"));
    },
  });
}

export function useRefreshPackage() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => packageService.refreshPackage(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: packageKeys.all(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Tracking refreshed");
    },
    onError: (error) => {
      const appError = createErrorFromUnknown(error);
      if (appError.type === "rate-limited") {
        toast.error("Please wait a few minutes before refreshing again");
      } else {
        toast.error(getErrorMessage(error, "Failed to refresh tracking"));
      }
    },
  });
}

export function useDetectCarrier() {
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (trackingNumber: string) =>
      packageService.detectCarrier(tenant!, trackingNumber),
  });
}
