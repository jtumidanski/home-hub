import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { householdPreferencesService } from "@/services/api/household-preferences";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { HouseholdPreferencesUpdateAttributes } from "@/types/models/dashboard";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const householdPreferencesKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["household-preferences", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
};

export function useHouseholdPreferences() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: householdPreferencesKeys.all(tenant, household),
    queryFn: () => householdPreferencesService.getPreferences(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 60 * 1000,
  });
}

export function useUpdateHouseholdPreferences() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: HouseholdPreferencesUpdateAttributes }) =>
      householdPreferencesService.updatePreferences(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: householdPreferencesKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update household preferences"));
    },
  });
}
