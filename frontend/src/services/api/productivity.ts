import { BaseService } from "./base";

import type { Task, TaskCreateAttributes, TaskUpdateAttributes } from "@/types/models/task";
import type { Reminder, ReminderCreateAttributes, ReminderUpdateAttributes } from "@/types/models/reminder";
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

  createTask(tenant: Tenant, attrs: TaskCreateAttributes) {
    return this.create<Task>(tenant, "/tasks", {
      data: { type: "tasks", attributes: { status: "pending", ...attrs } },
    });
  }

  async updateTask(tenant: Tenant, id: string, attrs: TaskUpdateAttributes) {
    const existing = await this.getOne<Task>(tenant, `/tasks/${id}`);
    return this.update<Task>(tenant, `/tasks/${id}`, {
      data: { type: "tasks", id, attributes: { ...existing.data.attributes, ...attrs } },
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

  createReminder(tenant: Tenant, attrs: ReminderCreateAttributes) {
    return this.create<Reminder>(tenant, "/reminders", {
      data: { type: "reminders", attributes: attrs },
    });
  }

  async updateReminder(tenant: Tenant, id: string, attrs: ReminderUpdateAttributes) {
    const existing = await this.getOne<Reminder>(tenant, `/reminders/${id}`);
    return this.update<Reminder>(tenant, `/reminders/${id}`, {
      data: { type: "reminders", id, attributes: { ...existing.data.attributes, ...attrs } },
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

  getTaskSummary(tenant: Tenant, date: string) {
    return this.getOne<TaskSummary>(tenant, `/summary/tasks?date=${encodeURIComponent(date)}`);
  }

  getReminderSummary(tenant: Tenant) {
    return this.getOne<ReminderSummary>(tenant, "/summary/reminders");
  }

  getDashboardSummary(tenant: Tenant, date: string) {
    return this.getOne<DashboardSummary>(tenant, `/summary/dashboard?date=${encodeURIComponent(date)}`);
  }
}

export const productivityService = new ProductivityService();
