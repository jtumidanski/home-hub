import { z } from "zod";

export const planFormSchema = z.object({
  starts_on: z.string().min(1, "Start date is required"),
  name: z.string().max(255).optional(),
});

export type PlanFormData = z.infer<typeof planFormSchema>;

export const planFormDefaults: PlanFormData = {
  starts_on: "",
  name: "",
};

export const planItemFormSchema = z.object({
  day: z.string().min(1, "Day is required"),
  slot: z.enum(["breakfast", "lunch", "dinner", "snack", "side"] as const, { message: "Slot is required" }),
  recipe_id: z.string().min(1, "Recipe is required"),
  serving_multiplier: z.number().positive().nullable().optional(),
  planned_servings: z.number().int().positive().nullable().optional(),
  notes: z.string().nullable().optional(),
});

export type PlanItemFormData = z.infer<typeof planItemFormSchema>;

const servingsRefinement = (
  data: { serving_multiplier?: number | null | undefined; planned_servings?: number | null | undefined },
  ctx: z.RefinementCtx,
  mode: "multiplier" | "planned",
) => {
  if (mode === "planned" && !data.planned_servings) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Planned servings is required",
      path: ["planned_servings"],
    });
  }
  if (mode === "multiplier" && !data.serving_multiplier) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Multiplier is required",
      path: ["serving_multiplier"],
    });
  }
};

export { servingsRefinement };

export const planItemPopoverSchema = planItemFormSchema.omit({ recipe_id: true });

export type PlanItemPopoverFormData = z.infer<typeof planItemPopoverSchema>;

export const planItemPopoverDefaults: PlanItemPopoverFormData = {
  day: "",
  slot: "dinner",
  serving_multiplier: null,
  planned_servings: null,
  notes: null,
};

export const planItemAddSchema = planItemFormSchema.omit({ recipe_id: true, day: true }).extend({
  days: z.array(z.string()).min(1, "Select at least one day"),
});

export type PlanItemAddFormData = z.infer<typeof planItemAddSchema>;

export const planItemAddDefaults: PlanItemAddFormData = {
  days: [],
  slot: "dinner",
  serving_multiplier: null,
  planned_servings: null,
  notes: null,
};
