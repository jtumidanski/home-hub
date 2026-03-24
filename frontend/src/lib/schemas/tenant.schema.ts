import { z } from "zod";

export const createTenantSchema = z.object({
  name: z.string().min(1, "Name is required"),
});

export type CreateTenantFormData = z.infer<typeof createTenantSchema>;

export const createTenantDefaults: CreateTenantFormData = {
  name: "",
};
