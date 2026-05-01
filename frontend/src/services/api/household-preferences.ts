import { api } from "@/lib/api/client";
import { BaseService } from "./base";
import type { ApiResponse } from "@/types/api/responses";
import type {
  HouseholdPreferences,
  HouseholdPreferencesUpdateAttributes,
} from "@/types/models/dashboard";

class HouseholdPreferencesService extends BaseService {
  constructor() {
    super("/household-preferences");
  }

  /**
   * Returns a single-row list (the backend FindOrCreate's the row if
   * missing). The hook normalises this to the first element.
   */
  getPreferences(tenant: { id: string }) {
    return this.getList<HouseholdPreferences>(tenant, "/household-preferences");
  }

  updatePreferences(
    tenant: { id: string },
    id: string,
    attrs: HouseholdPreferencesUpdateAttributes,
  ) {
    return this.update<HouseholdPreferences>(tenant, `/household-preferences/${id}`, {
      data: {
        type: "householdPreferences",
        id,
        attributes: attrs,
      },
    });
  }

  /**
   * Sets the write-once-true `kioskDashboardSeeded` flag via the dedicated
   * sub-route. Body is plain JSON (not JSON:API) — see Go resource.go for
   * rationale. Frontend never sends false; only true is accepted.
   */
  markKioskSeeded(tenant: { id: string }, id: string) {
    this.setTenant(tenant);
    return api.patch<ApiResponse<HouseholdPreferences>>(
      `/household-preferences/${id}/kiosk-seeded`,
      { value: true },
    );
  }
}

export const householdPreferencesService = new HouseholdPreferencesService();
