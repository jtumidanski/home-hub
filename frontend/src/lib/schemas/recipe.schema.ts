import { z } from "zod";

export const recipeFormSchema = z.object({
  source: z.string().min(1, "Recipe source is required"),
});

export type RecipeFormData = z.infer<typeof recipeFormSchema>;

export const recipeFormDefaults: RecipeFormData = {
  source: "",
};
