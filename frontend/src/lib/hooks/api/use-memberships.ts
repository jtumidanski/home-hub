import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { accountService } from "@/services/api/account";
import { useTenant } from "@/context/tenant-context";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";

// --- Key factory ---

export const membershipKeys = {
  all: ["memberships"] as const,
  lists: () => [...membershipKeys.all, "list"] as const,
  byHousehold: (tenant: Tenant | null, householdId: string) =>
    [...membershipKeys.lists(), tenant?.id ?? "no-tenant", householdId] as const,
};

// --- Query hooks ---

export function useHouseholdMembers(householdId: string | undefined) {
  const { tenant } = useTenant();
  return useQuery({
    queryKey: membershipKeys.byHousehold(tenant, householdId ?? ""),
    queryFn: () => accountService.listMembershipsByHousehold(tenant!, householdId!),
    enabled: !!tenant?.id && !!householdId,
    staleTime: 2 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useUpdateMemberRole() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (args: { membershipId: string; role: string }) =>
      accountService.updateMembershipRole(tenant!, args.membershipId, args.role),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: membershipKeys.lists() });
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update member role"));
    },
  });
}

export function useRemoveMember() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (membershipId: string) =>
      accountService.deleteMembership(tenant!, membershipId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: membershipKeys.lists() });
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to remove member"));
    },
  });
}

export function useLeaveHousehold() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (membershipId: string) =>
      accountService.deleteMembership(tenant!, membershipId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: membershipKeys.all });
      qc.invalidateQueries({ queryKey: contextKeys.all });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to leave household"));
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateMemberships() {
  const qc = useQueryClient();
  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: membershipKeys.all }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: membershipKeys.lists() }),
  };
}
