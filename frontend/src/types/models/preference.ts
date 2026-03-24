export interface PreferenceAttributes {
  theme: "light" | "dark";
  createdAt: string;
  updatedAt: string;
}

export interface Preference {
  id: string;
  type: "preferences";
  attributes: PreferenceAttributes;
}
