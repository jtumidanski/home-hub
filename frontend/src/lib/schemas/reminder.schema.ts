import { z } from "zod";

export const createReminderSchema = z.object({
  title: z.string().min(1, "Title is required").max(200, "Title must be 200 characters or fewer"),
  notes: z.string().max(1000, "Notes must be 1000 characters or fewer").optional(),
  scheduledFor: z.string().min(1, "Scheduled time is required"),
  ownerUserId: z.string().optional(),
});

export type CreateReminderFormData = z.infer<typeof createReminderSchema>;

export const createReminderDefaults: CreateReminderFormData = {
  title: "",
  notes: "",
  scheduledFor: "",
  ownerUserId: "",
};
