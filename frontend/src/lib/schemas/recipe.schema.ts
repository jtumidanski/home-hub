import { z } from "zod";

export const recipeFormSchema = z.object({
  title: z.string().min(1, "Title is required").max(255, "Title must be 255 characters or fewer"),
  description: z.string().max(2000, "Description must be 2000 characters or fewer").optional(),
  source: z.string().min(1, "Recipe source is required"),
});

export type RecipeFormData = z.infer<typeof recipeFormSchema>;

export const recipeFormDefaults: RecipeFormData = {
  title: "",
  description: "",
  source: "",
};
