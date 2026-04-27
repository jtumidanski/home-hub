import { BaseService } from "./base";
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
}

export const householdPreferencesService = new HouseholdPreferencesService();
