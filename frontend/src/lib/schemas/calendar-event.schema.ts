import { z } from "zod";

export const eventFormSchema = z
  .object({
    title: z.string().min(1, "Title is required").max(1024, "Title must be 1024 characters or fewer"),
    allDay: z.boolean(),
    startDate: z.string().min(1, "Start date is required"),
    startTime: z.string(),
    endDate: z.string(),
    endTime: z.string(),
    recurrence: z.string(),
    location: z.string().max(1024, "Location must be 1024 characters or fewer"),
    description: z.string().max(8192, "Description must be 8192 characters or fewer"),
    calendarId: z.string().min(1, "Calendar is required"),
    connectionId: z.string(),
  })
  .refine(
    (data) => {
      if (data.allDay) {
        return data.endDate >= data.startDate;
      }
      const start = `${data.startDate}T${data.startTime}`;
      const end = `${data.endDate}T${data.endTime}`;
      return end >= start;
    },
    { message: "End must be after start", path: ["endDate"] },
  );

export type EventFormData = z.infer<typeof eventFormSchema>;

export const RECURRENCE_OPTIONS = [
  { value: "", label: "Does not repeat" },
  { value: "RRULE:FREQ=DAILY", label: "Daily" },
  { value: "RRULE:FREQ=WEEKLY", label: "Weekly" },
  { value: "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", label: "Weekdays (Mon-Fri)" },
  { value: "RRULE:FREQ=MONTHLY", label: "Monthly" },
  { value: "RRULE:FREQ=YEARLY", label: "Yearly" },
] as const;

function padTime(date: Date): string {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function createEventDefaults(prefilledStart?: Date): EventFormData {
  const now = prefilledStart ?? new Date();
  const roundedMinutes = Math.ceil(now.getMinutes() / 15) * 15;
  const start = new Date(now);
  start.setMinutes(roundedMinutes, 0, 0);

  const end = new Date(start);
  end.setHours(end.getHours() + 1);

  return {
    title: "",
    allDay: false,
    startDate: formatDate(start),
    startTime: padTime(start),
    endDate: formatDate(end),
    endTime: padTime(end),
    recurrence: "",
    location: "",
    description: "",
    calendarId: "",
    connectionId: "",
  };
}
