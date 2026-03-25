import { BaseService } from "./base";

import type { Task, TaskAttributes } from "@/types/models/task";
import type { Reminder, ReminderAttributes } from "@/types/models/reminder";
import type { TaskSummary, ReminderSummary, DashboardSummary } from "@/types/models/summary";
import type { Tenant } from "@/types/models/tenant";

class ProductivityService extends BaseService {
  constructor() {
    super("/tasks");
  }

  // Tasks

  listTasks(tenant: Tenant) {
    return this.getList<Task>(tenant, "/tasks");
  }

  getTask(tenant: Tenant, id: string) {
    return this.getOne<Task>(tenant, `/tasks/${id}`);
  }

  createTask(tenant: Tenant, attrs: { title: string; notes?: string; dueOn?: string; rolloverEnabled?: boolean }) {
    return this.create<Task>(tenant, "/tasks", {
      data: { type: "tasks", attributes: { status: "pending", ...attrs } },
    });
  }

  updateTask(tenant: Tenant, id: string, attrs: Partial<TaskAttributes>) {
    return this.update<Task>(tenant, `/tasks/${id}`, {
      data: { type: "tasks", id, attributes: attrs },
    });
  }

  deleteTask(tenant: Tenant, id: string) {
    return this.remove(tenant, `/tasks/${id}`);
  }

  restoreTask(tenant: Tenant, taskId: string) {
    return this.create<Task>(tenant, "/tasks/restorations", {
      data: {
        type: "task-restorations",
        relationships: { task: { data: { type: "tasks", id: taskId } } },
      },
    });
  }

  // Reminders

  listReminders(tenant: Tenant) {
    return this.getList<Reminder>(tenant, "/reminders");
  }

  getReminder(tenant: Tenant, id: string) {
    return this.getOne<Reminder>(tenant, `/reminders/${id}`);
  }

  createReminder(tenant: Tenant, attrs: { title: string; notes?: string; scheduledFor: string }) {
    return this.create<Reminder>(tenant, "/reminders", {
      data: { type: "reminders", attributes: attrs },
    });
  }

  updateReminder(tenant: Tenant, id: string, attrs: Partial<ReminderAttributes>) {
    return this.update<Reminder>(tenant, `/reminders/${id}`, {
      data: { type: "reminders", id, attributes: attrs },
    });
  }

  deleteReminder(tenant: Tenant, id: string) {
    return this.remove(tenant, `/reminders/${id}`);
  }

  snoozeReminder(tenant: Tenant, reminderId: string, durationMinutes: number) {
    return this.create<Reminder>(tenant, "/reminders/snoozes", {
      data: {
        type: "reminder-snoozes",
        attributes: { durationMinutes },
        relationships: { reminder: { data: { type: "reminders", id: reminderId } } },
      },
    });
  }

  dismissReminder(tenant: Tenant, reminderId: string) {
    return this.create<Reminder>(tenant, "/reminders/dismissals", {
      data: {
        type: "reminder-dismissals",
        relationships: { reminder: { data: { type: "reminders", id: reminderId } } },
      },
    });
  }

  // Summaries

  getTaskSummary(tenant: Tenant) {
    return this.getOne<TaskSummary>(tenant, "/summary/tasks");
  }

  getReminderSummary(tenant: Tenant) {
    return this.getOne<ReminderSummary>(tenant, "/summary/reminders");
  }

  getDashboardSummary(tenant: Tenant) {
    return this.getOne<DashboardSummary>(tenant, "/summary/dashboard");
  }
}

export const productivityService = new ProductivityService();
