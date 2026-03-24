import { z } from "zod";

export const createHouseholdSchema = z.object({
  name: z.string().min(1, "Name is required"),
  timezone: z.string().min(1, "Timezone is required"),
  units: z.enum(["imperial", "metric"]),
});

export type CreateHouseholdFormData = z.infer<typeof createHouseholdSchema>;

export const createHouseholdDefaults: CreateHouseholdFormData = {
  name: "",
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  units: "imperial",
};
