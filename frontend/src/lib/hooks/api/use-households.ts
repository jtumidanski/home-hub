import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { accountService } from "@/services/api/account";
import { useTenant } from "@/context/tenant-context";
import { contextKeys } from "@/lib/hooks/api/use-context";

// --- Key factory ---

export const householdKeys = {
  all: (tenantId: string | null) =>
    ["households", tenantId ?? "no-tenant"] as const,
  lists: (tenantId: string | null) =>
    [...householdKeys.all(tenantId), "list"] as const,
  details: (tenantId: string | null) =>
    [...householdKeys.all(tenantId), "detail"] as const,
  detail: (tenantId: string | null, id: string) =>
    [...householdKeys.details(tenantId), id] as const,
};

// --- Query hooks ---

export function useHouseholds(enabled: boolean = true) {
  const { tenantId } = useTenant();
  return useQuery({
    queryKey: householdKeys.lists(tenantId),
    queryFn: () => accountService.listHouseholds(tenantId!),
    enabled: enabled && !!tenantId,
    staleTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateHousehold() {
  const qc = useQueryClient();
  const { tenantId } = useTenant();
  return useMutation({
    mutationFn: (attrs: { name: string; timezone: string; units: string }) =>
      accountService.createHousehold(tenantId!, attrs.name, attrs.timezone, attrs.units),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: householdKeys.lists(tenantId) });
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateHouseholds() {
  const qc = useQueryClient();
  const { tenantId } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: householdKeys.all(tenantId) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: householdKeys.lists(tenantId) }),
    invalidateHousehold: (id: string) =>
      qc.invalidateQueries({ queryKey: householdKeys.detail(tenantId, id) }),
  };
}

// --- Prefetch helper ---

export function usePrefetchHouseholds() {
  const qc = useQueryClient();
  const { tenantId } = useTenant();

  return {
    prefetch: () => {
      if (!tenantId) return;
      qc.prefetchQuery({
        queryKey: householdKeys.lists(tenantId),
        queryFn: () => accountService.listHouseholds(tenantId),
        staleTime: 5 * 60 * 1000,
      });
    },
  };
}
