import { api } from "@/lib/api/client";
import type { JsonApiResponse, JsonApiListResponse } from "@/types/api/responses";
import type { AppContext } from "@/types/models/context";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";
import type { Preference } from "@/types/models/preference";

class AccountService {
  getContext() {
    return api.get<JsonApiResponse<AppContext>>(
      "/contexts/current?include=tenant,activeHousehold,preference,memberships",
    );
  }

  createTenant(name: string) {
    return api.post<JsonApiResponse<Tenant>>("/tenants", {
      data: { type: "tenants", attributes: { name } },
    });
  }

  listHouseholds(tenant: Tenant) {
    api.setTenant(tenant);
    return api.get<JsonApiListResponse<Household>>("/households");
  }

  createHousehold(tenant: Tenant, name: string, timezone: string, units: string) {
    api.setTenant(tenant);
    return api.post<JsonApiResponse<Household>>("/households", {
      data: { type: "households", attributes: { name, timezone, units } },
    });
  }

  updatePreferenceTheme(tenant: Tenant, preferenceId: string, theme: "light" | "dark") {
    api.setTenant(tenant);
    return api.patch<JsonApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: { type: "preferences", id: preferenceId, attributes: { theme } },
    });
  }

  setActiveHousehold(tenant: Tenant, preferenceId: string, householdId: string) {
    api.setTenant(tenant);
    return api.patch<JsonApiResponse<Preference>>(`/preferences/${preferenceId}`, {
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
