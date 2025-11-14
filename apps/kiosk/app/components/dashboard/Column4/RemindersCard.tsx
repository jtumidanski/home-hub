'use client';

import React, { useState } from 'react';
import { Card } from '@/app/components/ui/Card';
import { Reminder, dismissReminder, snoozeReminder } from '@/lib/api/reminders';
import { Bell, Check, Clock } from 'lucide-react';

interface RemindersCardProps {
  reminders?: Reminder[] | null;
  loading?: boolean;
  error?: string | null;
  onUpdate?: () => void;
}

export function RemindersCard({ reminders, loading, error, onUpdate }: RemindersCardProps) {
  const [dismissedIds, setDismissedIds] = useState<Set<string>>(new Set());
  const [snoozedIds, setSnoozedIds] = useState<Set<string>>(new Set());

  const handleDismiss = async (id: string) => {
    setDismissedIds(prev => new Set([...prev, id]));
    await dismissReminder(id);
    if (onUpdate) onUpdate();
  };

  const handleSnooze = async (id: string) => {
    setSnoozedIds(prev => new Set([...prev, id]));
    await snoozeReminder(id, 900000); // 15 minutes
    if (onUpdate) onUpdate();
  };

  if (error) {
    return (
      <Card>
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-semibold text-gray-900 dark:text-white">Reminders</h3>
          <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
            Preview
          </span>
        </div>
        <div className="text-center py-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      </Card>
    );
  }

  if (loading || !reminders) {
    return <Card loading={true}>{null}</Card>;
  }

  const activeReminders = reminders.filter(
    r => r.status === 'active' && !dismissedIds.has(r.id) && !snoozedIds.has(r.id)
  );

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold text-gray-900 dark:text-white">Reminders</h3>
        <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
          Preview
        </span>
      </div>

      {activeReminders.length === 0 ? (
        <div className="text-center py-8">
          <Bell className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600 mb-2" />
          <p className="text-sm text-gray-500 dark:text-gray-400">No active reminders</p>
        </div>
      ) : (
        <div className="space-y-3">
          {activeReminders.map(reminder => (
            <ReminderItem
              key={reminder.id}
              reminder={reminder}
              onDismiss={handleDismiss}
              onSnooze={handleSnooze}
            />
          ))}
        </div>
      )}
    </Card>
  );
}

interface ReminderItemProps {
  reminder: Reminder;
  onDismiss: (id: string) => void;
  onSnooze: (id: string) => void;
}

function ReminderItem({ reminder, onDismiss, onSnooze }: ReminderItemProps) {
  const triggerTime = new Date(reminder.triggerAt);
  const isPast = triggerTime < new Date();

  return (
    <div className="p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
      <div className="flex items-start gap-3 mb-3">
        <Bell className="h-5 w-5 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
        <div className="flex-1 min-w-0">
          <p className="text-sm text-gray-900 dark:text-white">{reminder.text}</p>
          <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
            {isPast ? 'Now' : formatRelativeTime(triggerTime)}
          </p>
        </div>
      </div>

      <div className="flex gap-2">
        <button
          onClick={() => onSnooze(reminder.id)}
          className="flex-1 flex items-center justify-center gap-2 px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
        >
          <Clock className="h-3 w-3" />
          Snooze
        </button>
        <button
          onClick={() => onDismiss(reminder.id)}
          className="flex-1 flex items-center justify-center gap-2 px-3 py-2 text-xs font-medium text-white bg-blue-600 dark:bg-blue-500 rounded hover:bg-blue-700 dark:hover:bg-blue-600 transition-colors"
        >
          <Check className="h-3 w-3" />
          Dismiss
        </button>
      </div>
    </div>
  );
}

function formatRelativeTime(date: Date): string {
  const now = new Date();
  const seconds = Math.floor((date.getTime() - now.getTime()) / 1000);

  if (seconds < 0) return 'Now';
  if (seconds < 60) return 'in less than a minute';
  if (seconds < 3600) return `in ${Math.floor(seconds / 60)} minutes`;
  if (seconds < 86400) return `in ${Math.floor(seconds / 3600)} hours`;
  return `in ${Math.floor(seconds / 86400)} days`;
}
