export interface TaskSummaryAttributes {
  pendingCount: number;
  completedTodayCount: number;
  overdueCount: number;
}

export interface ReminderSummaryAttributes {
  dueNowCount: number;
  upcomingCount: number;
  snoozedCount: number;
}

export interface DashboardSummaryAttributes {
  householdName: string;
  timezone: string;
  pendingTaskCount: number;
  dueReminderCount: number;
  generatedAt: string;
}

export interface TaskSummary {
  id: "current";
  type: "task-summaries";
  attributes: TaskSummaryAttributes;
}

export interface ReminderSummary {
  id: "current";
  type: "reminder-summaries";
  attributes: ReminderSummaryAttributes;
}

export interface DashboardSummary {
  id: "current";
  type: "dashboard-summaries";
  attributes: DashboardSummaryAttributes;
}
