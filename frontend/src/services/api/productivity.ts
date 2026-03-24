import { api } from "@/lib/api/client";
import type { JsonApiResponse, JsonApiListResponse } from "@/types/api/responses";
import type { Task, TaskAttributes } from "@/types/models/task";
import type { Reminder, ReminderAttributes } from "@/types/models/reminder";
import type { TaskSummary, ReminderSummary, DashboardSummary } from "@/types/models/summary";

export const productivityService = {
  // Tasks
  listTasks: (tenantId: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiListResponse<Task>>("/tasks");
  },
  getTask: (tenantId: string, id: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiResponse<Task>>(`/tasks/${id}`);
  },
  createTask: (tenantId: string, attrs: { title: string; notes?: string; dueOn?: string; rolloverEnabled?: boolean }) => {
    api.setTenant(tenantId);
    return api.post<JsonApiResponse<Task>>("/tasks", {
      data: { type: "tasks", attributes: { status: "pending", ...attrs } },
    });
  },
  updateTask: (tenantId: string, id: string, attrs: Partial<TaskAttributes>) => {
    api.setTenant(tenantId);
    return api.patch<JsonApiResponse<Task>>(`/tasks/${id}`, {
      data: { type: "tasks", id, attributes: attrs },
    });
  },
  deleteTask: (tenantId: string, id: string) => {
    api.setTenant(tenantId);
    return api.delete(`/tasks/${id}`);
  },
  restoreTask: (tenantId: string, taskId: string) => {
    api.setTenant(tenantId);
    return api.post<JsonApiResponse<Task>>("/tasks/restorations", {
      data: {
        type: "task-restorations",
        relationships: { task: { data: { type: "tasks", id: taskId } } },
      },
    });
  },

  // Reminders
  listReminders: (tenantId: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiListResponse<Reminder>>("/reminders");
  },
  getReminder: (tenantId: string, id: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiResponse<Reminder>>(`/reminders/${id}`);
  },
  createReminder: (tenantId: string, attrs: { title: string; notes?: string; scheduledFor: string }) => {
    api.setTenant(tenantId);
    return api.post<JsonApiResponse<Reminder>>("/reminders", {
      data: { type: "reminders", attributes: attrs },
    });
  },
  updateReminder: (tenantId: string, id: string, attrs: Partial<ReminderAttributes>) => {
    api.setTenant(tenantId);
    return api.patch<JsonApiResponse<Reminder>>(`/reminders/${id}`, {
      data: { type: "reminders", id, attributes: attrs },
    });
  },
  deleteReminder: (tenantId: string, id: string) => {
    api.setTenant(tenantId);
    return api.delete(`/reminders/${id}`);
  },
  snoozeReminder: (tenantId: string, reminderId: string, durationMinutes: number) => {
    api.setTenant(tenantId);
    return api.post<JsonApiResponse<Reminder>>("/reminders/snoozes", {
      data: {
        type: "reminder-snoozes",
        attributes: { durationMinutes },
        relationships: { reminder: { data: { type: "reminders", id: reminderId } } },
      },
    });
  },
  dismissReminder: (tenantId: string, reminderId: string) => {
    api.setTenant(tenantId);
    return api.post<JsonApiResponse<Reminder>>("/reminders/dismissals", {
      data: {
        type: "reminder-dismissals",
        relationships: { reminder: { data: { type: "reminders", id: reminderId } } },
      },
    });
  },

  // Summaries
  getTaskSummary: (tenantId: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiResponse<TaskSummary>>("/summary/tasks");
  },
  getReminderSummary: (tenantId: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiResponse<ReminderSummary>>("/summary/reminders");
  },
  getDashboardSummary: (tenantId: string) => {
    api.setTenant(tenantId);
    return api.get<JsonApiResponse<DashboardSummary>>("/summary/dashboard");
  },
};
