import { z } from "zod";

export const createTenantSchema = z.object({
  name: z.string().min(1, "Name is required").max(100, "Name must be 100 characters or fewer"),
});

export type CreateTenantFormData = z.infer<typeof createTenantSchema>;

export const createTenantDefaults: CreateTenantFormData = {
  name: "",
};
