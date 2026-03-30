import { z } from "zod";

export const planFormSchema = z.object({
  starts_on: z.string().min(1, "Start date is required"),
  name: z.string().max(255).optional(),
});

export type PlanFormData = z.infer<typeof planFormSchema>;

export const planItemFormSchema = z.object({
  day: z.string().min(1, "Day is required"),
  slot: z.enum(["breakfast", "lunch", "dinner", "snack", "side"] as const, { message: "Slot is required" }),
  recipe_id: z.string().min(1, "Recipe is required"),
  serving_multiplier: z.number().positive().nullable().optional(),
  planned_servings: z.number().int().positive().nullable().optional(),
  notes: z.string().nullable().optional(),
});

export type PlanItemFormData = z.infer<typeof planItemFormSchema>;
