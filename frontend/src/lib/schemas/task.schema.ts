import { z } from "zod";

export const createTaskSchema = z.object({
  title: z.string().min(1, "Title is required"),
  notes: z.string().optional(),
  dueOn: z.string().optional(),
});

export type CreateTaskFormData = z.infer<typeof createTaskSchema>;

export const createTaskDefaults: CreateTaskFormData = {
  title: "",
  notes: "",
  dueOn: "",
};
