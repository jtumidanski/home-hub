import { z } from "zod";

export const createInvitationSchema = z.object({
  email: z
    .string()
    .min(1, "Email is required")
    .email("Must be a valid email address"),
  role: z.enum(["viewer", "editor", "admin"]),
});

export type CreateInvitationFormData = z.infer<typeof createInvitationSchema>;

export const createInvitationDefaults: CreateInvitationFormData = {
  email: "",
  role: "viewer",
};
