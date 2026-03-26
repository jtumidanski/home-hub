import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { accountService } from "@/services/api/account";
import { useTenant } from "@/context/tenant-context";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { householdKeys } from "@/lib/hooks/api/use-households";
import { getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";
import type { InvitationCreateAttributes } from "@/types/models/invitation";

// --- Key factory ---

export const invitationKeys = {
  all: ["invitations"] as const,
  lists: () => [...invitationKeys.all, "list"] as const,
  byHousehold: (tenant: Tenant | null, householdId: string) =>
    [...invitationKeys.lists(), tenant?.id ?? "no-tenant", householdId] as const,
  mine: () => [...invitationKeys.all, "mine"] as const,
};

// --- Query hooks ---

export function useHouseholdInvitations(householdId: string | undefined) {
  const { tenant } = useTenant();
  return useQuery({
    queryKey: invitationKeys.byHousehold(tenant, householdId ?? ""),
    queryFn: () => accountService.listInvitationsByHousehold(tenant!, householdId!),
    enabled: !!tenant?.id && !!householdId,
    staleTime: 2 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useMyInvitations(enabled: boolean = true) {
  return useQuery({
    queryKey: invitationKeys.mine(),
    queryFn: () => accountService.listMyInvitations(),
    enabled,
    staleTime: 2 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateInvitation() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (args: { householdId: string; attrs: InvitationCreateAttributes }) =>
      accountService.createInvitation(tenant!, args.householdId, args.attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: invitationKeys.lists() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create invitation"));
    },
  });
}

export function useRevokeInvitation() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (id: string) => accountService.revokeInvitation(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: invitationKeys.lists() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to revoke invitation"));
    },
  });
}

export function useAcceptInvitation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => accountService.acceptInvitation(id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: contextKeys.all });
      qc.invalidateQueries({ queryKey: householdKeys.all(null) });
      qc.invalidateQueries({ queryKey: invitationKeys.all });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to accept invitation"));
    },
  });
}

export function useDeclineInvitation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => accountService.declineInvitation(id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: contextKeys.all });
      qc.invalidateQueries({ queryKey: invitationKeys.all });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to decline invitation"));
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateInvitations() {
  const qc = useQueryClient();
  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: invitationKeys.all }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: invitationKeys.lists() }),
    invalidateMine: () =>
      qc.invalidateQueries({ queryKey: invitationKeys.mine() }),
  };
}
