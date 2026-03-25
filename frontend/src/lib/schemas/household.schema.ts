import { z } from "zod";

export const createHouseholdSchema = z.object({
  name: z.string().min(1, "Name is required").max(100, "Name must be 100 characters or fewer"),
  timezone: z.string().min(1, "Timezone is required").max(100, "Timezone must be 100 characters or fewer"),
  units: z.enum(["imperial", "metric"]),
});

export type CreateHouseholdFormData = z.infer<typeof createHouseholdSchema>;

export const createHouseholdDefaults: CreateHouseholdFormData = {
  name: "",
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  units: "imperial",
};
