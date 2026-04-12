import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { retentionService } from "@/services/api/retention";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";

export const retentionKeys = {
  all: (tenant: Tenant | null) => ["retention", tenant?.id ?? "no-tenant"] as const,
  policy: (tenant: Tenant | null) => [...retentionKeys.all(tenant), "policy"] as const,
  runs: (tenant: Tenant | null, params: { category?: string; trigger?: string; limit?: number }) =>
    [...retentionKeys.all(tenant), "runs", params] as const,
};

export function useRetentionPolicies(enabled: boolean = true) {
  const { tenant } = useTenant();
  return useQuery({
    queryKey: retentionKeys.policy(tenant),
    queryFn: () => retentionService.getPolicies(tenant!),
    enabled: enabled && !!tenant?.id,
    staleTime: 5 * 60 * 1000,
  });
}

export function useRetentionRuns(params: { category?: string; trigger?: string; limit?: number } = {}) {
  const { tenant } = useTenant();
  return useQuery({
    queryKey: retentionKeys.runs(tenant, params),
    queryFn: () => retentionService.listRuns(tenant!, params),
    enabled: !!tenant?.id,
    staleTime: 30 * 1000,
  });
}

export function usePatchHouseholdRetention() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (args: { householdId: string; categories: Record<string, number | null> }) =>
      retentionService.patchHousehold(tenant!, args.householdId, args.categories),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: retentionKeys.policy(tenant) });
    },
    onError: (e) => toast.error(getErrorMessage(e, "Failed to update retention policy")),
    onSuccess: () => toast.success("Retention policy updated"),
  });
}

export function usePatchUserRetention() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (categories: Record<string, number | null>) =>
      retentionService.patchUser(tenant!, categories),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: retentionKeys.policy(tenant) });
    },
    onError: (e) => toast.error(getErrorMessage(e, "Failed to update retention policy")),
    onSuccess: () => toast.success("Retention policy updated"),
  });
}

export function usePurgeRetention() {
  const qc = useQueryClient();
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (args: { category: string; scope: "household" | "user"; dryRun?: boolean }) =>
      retentionService.purge(tenant!, args.category, args.scope, args.dryRun ?? false),
    onSettled: (_data, _err, vars) => {
      if (!vars?.dryRun) {
        qc.invalidateQueries({ queryKey: retentionKeys.all(tenant) });
      }
    },
    onError: (e) => toast.error(getErrorMessage(e, "Purge failed")),
  });
}
