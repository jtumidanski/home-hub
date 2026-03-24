import { api } from "@/lib/api/client";
import type { JsonApiResponse, JsonApiListResponse } from "@/types/api/responses";
import type { AppContext } from "@/types/models/context";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";
import type { Preference } from "@/types/models/preference";

export const accountService = {
  getContext: () =>
    api.get<JsonApiResponse<AppContext>>(
      "/contexts/current?include=tenant,activeHousehold,preference,memberships"
    ),

  createTenant: (name: string) =>
    api.post<JsonApiResponse<Tenant>>("/tenants", {
      data: { type: "tenants", attributes: { name } },
    }),

  listHouseholds: (tenantId: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiListResponse<Household>>("/households");
  },

  createHousehold: (tenantId: string, name: string, timezone: string, units: string) => {
    api.setTenant(tenantId);
    return api.post<JsonApiResponse<Household>>("/households", {
      data: { type: "households", attributes: { name, timezone, units } },
    });
  },

  updatePreferenceTheme: (tenantId: string, preferenceId: string, theme: "light" | "dark") => {
    api.setTenant(tenantId);
    return api.patch<JsonApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: { type: "preferences", id: preferenceId, attributes: { theme } },
    });
  },

  setActiveHousehold: (tenantId: string, preferenceId: string, householdId: string) => {
    api.setTenant(tenantId);
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
  },
};
