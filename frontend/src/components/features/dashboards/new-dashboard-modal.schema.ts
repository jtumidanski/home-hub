import { z } from "zod";

export const dashboardNameSchema = z.string().trim().min(1, "Name is required").max(80, "Max 80 characters");

export const newDashboardFormSchema = z.object({
  name: dashboardNameSchema,
  scope: z.enum(["household", "user"]),
  copyOf: z.string().optional(),
});

export type NewDashboardFormData = z.infer<typeof newDashboardFormSchema>;
