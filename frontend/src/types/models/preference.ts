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

// --- Update attributes (F14) ---

export type PreferenceUpdateAttributes = Partial<
  Pick<PreferenceAttributes, "theme">
>;

// --- Label map (F15) ---

export const themeLabelMap: Record<PreferenceAttributes["theme"], string> = {
  light: "Light",
  dark: "Dark",
};
