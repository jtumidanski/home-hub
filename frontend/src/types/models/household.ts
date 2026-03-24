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
