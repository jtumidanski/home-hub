import { z } from "zod";

export const createReminderSchema = z.object({
  title: z.string().min(1, "Title is required"),
  notes: z.string().optional(),
  scheduledFor: z.string().min(1, "Scheduled time is required"),
});

export type CreateReminderFormData = z.infer<typeof createReminderSchema>;

export const createReminderDefaults: CreateReminderFormData = {
  title: "",
  notes: "",
  scheduledFor: "",
};
