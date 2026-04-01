import { z } from "zod";

export const shoppingListNameSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100, "Name must be 100 characters or fewer"),
});

export type ShoppingListNameFormData = z.infer<typeof shoppingListNameSchema>;
