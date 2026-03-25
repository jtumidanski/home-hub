import { api, type RequestOptions } from "@/lib/api/client";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";
import type { AppContext } from "@/types/models/context";
import type { Tenant } from "@/types/models/tenant";
import type { TenantCreateAttributes } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";
import type { HouseholdCreateAttributes } from "@/types/models/household";
import type { Preference } from "@/types/models/preference";

class AccountService {
  getContext(options?: RequestOptions) {
    return api.get<ApiResponse<AppContext>>(
      "/contexts/current?include=tenant,activeHousehold,preference,memberships",
      options,
    );
  }

  createTenant(attrs: TenantCreateAttributes) {
    return api.post<ApiResponse<Tenant>>("/tenants", {
      data: { type: "tenants", attributes: attrs },
    });
  }

  listHouseholds(tenant: Tenant) {
    api.setTenant(tenant);
    return api.get<ApiListResponse<Household>>("/households");
  }

  createHousehold(tenant: Tenant, attrs: HouseholdCreateAttributes) {
    api.setTenant(tenant);
    return api.post<ApiResponse<Household>>("/households", {
      data: { type: "households", attributes: attrs },
    });
  }

  // Merge not needed: "theme" is the only mutable field on PreferenceAttributes
  // (createdAt and updatedAt are server-managed), so the patch payload is already complete.
  updatePreferenceTheme(tenant: Tenant, preferenceId: string, theme: "light" | "dark") {
    api.setTenant(tenant);
    return api.patch<ApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: { type: "preferences", id: preferenceId, attributes: { theme } },
    });
  }

  setActiveHousehold(tenant: Tenant, preferenceId: string, householdId: string) {
    api.setTenant(tenant);
    return api.patch<ApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: {
        type: "preferences",
        id: preferenceId,
        relationships: {
          activeHousehold: {
            data: { type: "households", id: householdId },
          },
        },
      },
    });
  }
}

export const accountService = new AccountService();
