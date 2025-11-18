'use client';

import { Reminder } from '@/lib/api/reminders';
import { ReminderListItem } from './ReminderListItem';

interface RemindersListProps {
  reminders: Reminder[];
  onSnooze: (id: string, minutes: number) => Promise<void>;
  onDismiss: (id: string) => Promise<void>;
  onEdit: (reminder: Reminder) => void;
  onDelete: (id: string) => Promise<void>;
}

export function RemindersList({
  reminders,
  onSnooze,
  onDismiss,
  onEdit,
  onDelete,
}: RemindersListProps) {
  // Group reminders
  const active = reminders
    .filter(r => r.status === 'active')
    .sort((a, b) => new Date(a.remindAt).getTime() - new Date(b.remindAt).getTime());

  const snoozed = reminders
    .filter(r => r.status === 'snoozed')
    .sort((a, b) => new Date(a.remindAt).getTime() - new Date(b.remindAt).getTime());

  const dismissed = reminders
    .filter(r => r.status === 'dismissed')
    .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());

  if (reminders.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 px-4">
        <div className="text-center">
          <div className="mb-4 text-4xl">🔔</div>
          <h3 className="text-lg font-semibold text-card-foreground mb-2">
            No reminders yet
          </h3>
          <p className="text-sm text-muted-foreground">
            Create your first reminder to get started
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Active */}
      {active.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-blue-600 dark:text-blue-500 mb-3 px-1">
            Active ({active.length})
          </h2>
          <div className="space-y-2">
            {active.map(reminder => (
              <ReminderListItem
                key={reminder.id}
                reminder={reminder}
                onSnooze={onSnooze}
                onDismiss={onDismiss}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Snoozed */}
      {snoozed.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-amber-600 dark:text-amber-500 mb-3 px-1">
            Snoozed ({snoozed.length})
          </h2>
          <div className="space-y-2">
            {snoozed.map(reminder => (
              <ReminderListItem
                key={reminder.id}
                reminder={reminder}
                onSnooze={onSnooze}
                onDismiss={onDismiss}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Dismissed */}
      {dismissed.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-muted-foreground mb-3 px-1">
            Dismissed ({dismissed.length})
          </h2>
          <div className="space-y-2">
            {dismissed.map(reminder => (
              <ReminderListItem
                key={reminder.id}
                reminder={reminder}
                onSnooze={onSnooze}
                onDismiss={onDismiss}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
