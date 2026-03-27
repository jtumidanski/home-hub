export interface ReminderAttributes {
  title: string;
  notes?: string;
  scheduledFor: string;
  ownerUserId?: string | null;
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

// --- Create attributes (F14) ---

export interface ReminderCreateAttributes {
  title: string;
  notes?: string;
  scheduledFor: string;
  ownerUserId?: string | null;
}

// --- Update attributes (F14) ---

export type ReminderUpdateAttributes = Partial<
  Pick<ReminderAttributes, "title" | "notes" | "scheduledFor" | "ownerUserId">
>;

// --- Helpers (F16) ---

export function isReminderDismissed(reminder: Reminder): boolean {
  return !reminder.attributes.active && reminder.attributes.lastDismissedAt != null;
}

export function isReminderSnoozed(reminder: Reminder): boolean {
  if (!reminder.attributes.lastSnoozedUntil) {
    return false;
  }
  return new Date(reminder.attributes.lastSnoozedUntil) > new Date();
}
