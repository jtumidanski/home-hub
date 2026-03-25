import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { accountService } from "@/services/api/account";
import { useTenant } from "@/context/tenant-context";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import type { HouseholdUpdateAttributes } from "@/types/models/household";
import { weatherKeys } from "@/lib/hooks/api/use-weather";

// --- Key factory ---

export const householdKeys = {
  all: (tenant: Tenant | null) =>
    ["households", tenant?.id ?? "no-tenant"] as const,
  lists: (tenant: Tenant | null) =>
    [...householdKeys.all(tenant), "list"] as const,
  details: (tenant: Tenant | null) =>
    [...householdKeys.all(tenant), "detail"] as const,
  detail: (tenant: Tenant | null, id: string) =>
    [...householdKeys.details(tenant), id] as const,
};

// --- Query hooks ---

export function useHouseholds(enabled: boolean = true) {
  const { tenant } = useTenant();
  return useQuery({
    queryKey: householdKeys.lists(tenant),
    queryFn: () => accountService.listHouseholds(tenant!),
    enabled: enabled && !!tenant?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateHousehold() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (attrs: { name: string; timezone: string; units: "imperial" | "metric" }) =>
      accountService.createHousehold(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: householdKeys.lists(tenant) });
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create household"));
    },
  });
}

export function useUpdateHousehold() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (args: { householdId: string; attrs: HouseholdUpdateAttributes }) =>
      accountService.updateHousehold(tenant!, args.householdId, args.attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: householdKeys.lists(tenant) });
      qc.invalidateQueries({ queryKey: contextKeys.current() });
      qc.invalidateQueries({ queryKey: weatherKeys.all });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update household"));
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateHouseholds() {
  const qc = useQueryClient();
  const { tenant } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: householdKeys.all(tenant) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: householdKeys.lists(tenant) }),
    invalidateHousehold: (id: string) =>
      qc.invalidateQueries({ queryKey: householdKeys.detail(tenant, id) }),
  };
}

// --- Prefetch helper ---

export function usePrefetchHouseholds() {
  const qc = useQueryClient();
  const { tenant } = useTenant();

  return {
    prefetch: () => {
      if (!tenant) return;
      qc.prefetchQuery({
        queryKey: householdKeys.lists(tenant),
        queryFn: () => accountService.listHouseholds(tenant),
        staleTime: 5 * 60 * 1000,
      });
    },
  };
}
