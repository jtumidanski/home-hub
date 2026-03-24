import { createContext, useContext, useEffect, type ReactNode } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { api } from "@/lib/api/client";

interface TenantContextValue {
  tenantId: string | null;
  householdId: string | null;
}

const TenantContext = createContext<TenantContextValue | undefined>(undefined);

export function TenantProvider({ children }: { children: ReactNode }) {
  const { appContext } = useAuth();

  const tenantId = appContext?.relationships?.tenant?.data?.id ?? null;
  const householdId = appContext?.relationships?.activeHousehold?.data?.id ?? null;

  useEffect(() => {
    if (tenantId) {
      api.setTenant(tenantId);
    } else {
      api.clearTenant();
    }
  }, [tenantId]);

  return (
    <TenantContext.Provider value={{ tenantId, householdId }}>
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
