import { api } from "@/lib/api/client";
import type { JsonApiResponse, JsonApiListResponse } from "@/types/api/responses";
import type { Task } from "@/types/models/task";
import type { Reminder } from "@/types/models/reminder";
import type { TaskSummary, ReminderSummary, DashboardSummary } from "@/types/models/summary";

export const productivityService = {
  // Tasks
  listTasks: () => api.get<JsonApiListResponse<Task>>("/tasks"),
  getTask: (id: string) => api.get<JsonApiResponse<Task>>(`/tasks/${id}`),
  createTask: (attrs: { title: string; notes?: string; dueOn?: string; rolloverEnabled?: boolean }) =>
    api.post<JsonApiResponse<Task>>("/tasks", {
      data: { type: "tasks", attributes: { status: "pending", ...attrs } },
    }),
  updateTask: (id: string, attrs: Record<string, unknown>) =>
    api.patch<JsonApiResponse<Task>>(`/tasks/${id}`, {
      data: { type: "tasks", id, attributes: attrs },
    }),
  deleteTask: (id: string) => api.delete(`/tasks/${id}`),
  restoreTask: (taskId: string) =>
    api.post("/tasks/restorations", {
      data: {
        type: "task-restorations",
        relationships: { task: { data: { type: "tasks", id: taskId } } },
      },
    }),

  // Reminders
  listReminders: () => api.get<JsonApiListResponse<Reminder>>("/reminders"),
  getReminder: (id: string) => api.get<JsonApiResponse<Reminder>>(`/reminders/${id}`),
  createReminder: (attrs: { title: string; notes?: string; scheduledFor: string }) =>
    api.post<JsonApiResponse<Reminder>>("/reminders", {
      data: { type: "reminders", attributes: attrs },
    }),
  updateReminder: (id: string, attrs: Record<string, unknown>) =>
    api.patch<JsonApiResponse<Reminder>>(`/reminders/${id}`, {
      data: { type: "reminders", id, attributes: attrs },
    }),
  deleteReminder: (id: string) => api.delete(`/reminders/${id}`),
  snoozeReminder: (reminderId: string, durationMinutes: number) =>
    api.post("/reminders/snoozes", {
      data: {
        type: "reminder-snoozes",
        attributes: { durationMinutes },
        relationships: { reminder: { data: { type: "reminders", id: reminderId } } },
      },
    }),
  dismissReminder: (reminderId: string) =>
    api.post("/reminders/dismissals", {
      data: {
        type: "reminder-dismissals",
        relationships: { reminder: { data: { type: "reminders", id: reminderId } } },
      },
    }),

  // Summaries
  getTaskSummary: () => api.get<JsonApiResponse<TaskSummary>>("/summary/tasks"),
  getReminderSummary: () => api.get<JsonApiResponse<ReminderSummary>>("/summary/reminders"),
  getDashboardSummary: () => api.get<JsonApiResponse<DashboardSummary>>("/summary/dashboard"),
};
