import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { accountService } from "@/services/api/account";
import { useTenant } from "@/context/tenant-context";
import type { Member } from "@/types/models/member";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

export const memberKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["members", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
};

export function useHouseholdMembers() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: memberKeys.all(tenant, household),
    queryFn: () => accountService.listHouseholdMembers(tenant!, household!.id),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 10 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
  });
}

export function useMemberMap(): Map<string, string> {
  const { data } = useHouseholdMembers();
  return useMemo(() => {
    const map = new Map<string, string>();
    const members = (data?.data ?? []) as Member[];
    for (const m of members) {
      const userId = m.relationships.user.data.id;
      map.set(userId, m.attributes.displayName);
    }
    return map;
  }, [data]);
}
