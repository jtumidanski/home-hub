import { api } from "@/lib/api/client";
import type { JsonApiResponse } from "@/types/api/responses";
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

  createHousehold: (name: string, timezone: string, units: string) =>
    api.post<JsonApiResponse<Household>>("/households", {
      data: { type: "households", attributes: { name, timezone, units } },
    }),

  updatePreferenceTheme: (preferenceId: string, theme: "light" | "dark") =>
    api.patch<JsonApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: { type: "preferences", id: preferenceId, attributes: { theme } },
    }),

  setActiveHousehold: (preferenceId: string, householdId: string) =>
    api.patch<JsonApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: {
        type: "preferences",
        id: preferenceId,
        relationships: {
          activeHousehold: {
            data: { type: "households", id: householdId },
          },
        },
      },
    }),
};
