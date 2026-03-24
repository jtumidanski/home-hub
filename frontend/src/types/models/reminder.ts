export interface ReminderAttributes {
  title: string;
  notes?: string;
  scheduledFor: string;
  active: boolean;
  lastDismissedAt?: string;
  lastSnoozedUntil?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Reminder {
  id: string;
  type: "reminders";
  attributes: ReminderAttributes;
}
