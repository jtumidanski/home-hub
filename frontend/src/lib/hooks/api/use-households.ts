import { useQuery } from "@tanstack/react-query";
import { accountService } from "@/services/api/account";
import { useTenant } from "@/context/tenant-context";

export const householdKeys = {
  all: (tenantId: string | null) => ["households", tenantId ?? "no-tenant"] as const,
  list: (tenantId: string | null) => [...householdKeys.all(tenantId), "list"] as const,
};

export function useHouseholds(enabled: boolean = true) {
  const { tenantId } = useTenant();
  return useQuery({
    queryKey: householdKeys.list(tenantId),
    queryFn: () => accountService.listHouseholds(tenantId!),
    enabled: enabled && !!tenantId,
    staleTime: 5 * 60 * 1000,
  });
}
