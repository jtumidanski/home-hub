import type { Layout } from "@/lib/dashboard/schema";

export type DashboardScope = "household" | "user";

export interface DashboardAttributes {
  name: string;
  scope: DashboardScope;
  sortOrder: number;
  layout: Layout;
  schemaVersion: number;
  createdAt: string;
  updatedAt: string;
}

export interface Dashboard {
  id: string;
  type: "dashboards";
  attributes: DashboardAttributes;
}

export interface DashboardCreateAttributes {
  name: string;
  scope: DashboardScope;
  layout?: Layout;
  sortOrder?: number;
}

export interface DashboardUpdateAttributes {
  name?: string;
  layout?: Layout;
  sortOrder?: number;
}

export interface DashboardOrderEntry {
  id: string;
  sortOrder: number;
}

export interface HouseholdPreferencesAttributes {
  defaultDashboardId: string | null;
  kioskDashboardSeeded: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface HouseholdPreferences {
  id: string;
  type: "householdPreferences";
  attributes: HouseholdPreferencesAttributes;
}

export interface HouseholdPreferencesUpdateAttributes {
  /**
   * Literal `null` clears the preference; omitting the field leaves it
   * unchanged (handled by the caller — see PATCH semantics).
   */
  defaultDashboardId?: string | null;
}
