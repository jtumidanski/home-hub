import { z } from "zod";

export const createTaskSchema = z.object({
  title: z.string().min(1, "Title is required").max(200, "Title must be 200 characters or fewer"),
  notes: z.string().max(1000, "Notes must be 1000 characters or fewer").optional(),
  dueOn: z.string().optional(),
});

export type CreateTaskFormData = z.infer<typeof createTaskSchema>;

export const createTaskDefaults: CreateTaskFormData = {
  title: "",
  notes: "",
  dueOn: "",
};
