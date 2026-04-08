import { z } from "zod";

export const URGENCY_VALUES = ["must_have", "need_to_have", "want"] as const;

export const wishListItemSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required")
    .max(255, "Name must be 255 characters or fewer"),
  purchase_location: z
    .string()
    .trim()
    .max(255, "Purchase location must be 255 characters or fewer")
    .optional()
    .or(z.literal("")),
  urgency: z.enum(URGENCY_VALUES),
});

export type WishListItemFormData = z.infer<typeof wishListItemSchema>;
