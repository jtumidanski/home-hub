export interface HouseholdAttributes {
  name: string;
  timezone: string;
  units: "imperial" | "metric";
  createdAt: string;
  updatedAt: string;
}

export interface Household {
  id: string;
  type: "households";
  attributes: HouseholdAttributes;
}

// --- Create attributes (F14) ---

export interface HouseholdCreateAttributes {
  name: string;
  timezone: string;
  units: "imperial" | "metric";
}

// --- Update attributes (F14) ---

export type HouseholdUpdateAttributes = Partial<
  Pick<HouseholdAttributes, "name" | "timezone" | "units">
>;

// --- Label map (F15) ---

export const unitsLabelMap: Record<HouseholdAttributes["units"], string> = {
  imperial: "Imperial",
  metric: "Metric",
};

// --- Helpers (F16) ---

export function isHouseholdMetric(household: Household): boolean {
  return household.attributes.units === "metric";
}
