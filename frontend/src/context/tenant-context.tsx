import { createContext, useCallback, useContext, useEffect, useMemo, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/components/providers/auth-provider";
import { api } from "@/lib/api/client";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

interface TenantContextValue {
  tenantId: string | null;
  householdId: string | null;
  tenant: Tenant | null;
  household: Household | null;
  setActiveHousehold: (householdId: string) => Promise<void>;
}

const TenantContext = createContext<TenantContextValue | undefined>(undefined);

export function TenantProvider({ children }: { children: ReactNode }) {
  const { appContext } = useAuth();
  const queryClient = useQueryClient();

  const tenantId = appContext?.relationships?.tenant?.data?.id ?? null;
  const householdId = appContext?.relationships?.activeHousehold?.data?.id ?? null;
  const preferenceId = appContext?.relationships?.preference?.data?.id ?? null;

  const tenant: Tenant | null = useMemo(() => {
    if (!tenantId) return null;
    const included = (queryClient.getQueryData(contextKeys.current) as { included?: Array<{ type: string; id: string; attributes: Record<string, unknown> }> })?.included;
    const tenantResource = included?.find((r) => r.type === "tenants" && r.id === tenantId);
    if (tenantResource) {
      return tenantResource as unknown as Tenant;
    }
    return { id: tenantId, type: "tenants" as const, attributes: { name: "", createdAt: "", updatedAt: "" } };
  }, [tenantId, queryClient]);

  const household: Household | null = useMemo(() => {
    if (!householdId) return null;
    const included = (queryClient.getQueryData(contextKeys.current) as { included?: Array<{ type: string; id: string; attributes: Record<string, unknown> }> })?.included;
    const householdResource = included?.find((r) => r.type === "households" && r.id === householdId);
    if (householdResource) {
      return householdResource as unknown as Household;
    }
    return { id: householdId, type: "households" as const, attributes: { name: "", timezone: "", units: "imperial" as const, createdAt: "", updatedAt: "" } };
  }, [householdId, queryClient]);

  const setActiveHousehold = useCallback(
    async (newHouseholdId: string) => {
      if (!tenantId || !preferenceId) return;
      await accountService.setActiveHousehold(tenantId, preferenceId, newHouseholdId);
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
    },
    [tenantId, preferenceId, queryClient],
  );

  useEffect(() => {
    if (tenantId) {
      api.setTenant(tenantId);
    } else {
      api.clearTenant();
    }
  }, [tenantId]);

  return (
    <TenantContext.Provider value={{ tenantId, householdId, tenant, household, setActiveHousehold }}>
      {children}
    </TenantContext.Provider>
  );
}

export function useTenant() {
  const context = useContext(TenantContext);
  if (!context) {
    throw new Error("useTenant must be used within TenantProvider");
  }
  return context;
}
