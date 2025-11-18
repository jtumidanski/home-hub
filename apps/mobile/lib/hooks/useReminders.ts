'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  getReminders,
  createReminder,
  updateReminder,
  snoozeReminder,
  dismissReminder,
  deleteReminder,
  type Reminder,
  type CreateReminderInput,
  type UpdateReminderInput,
} from '@/lib/api/reminders';

export function useReminders() {
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchReminders = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getReminders();
      setReminders(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch reminders';
      setError(message);
      console.error('Error fetching reminders:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReminders();
  }, [fetchReminders]);

  const handleCreateReminder = useCallback(async (input: CreateReminderInput): Promise<Reminder> => {
    try {
      const newReminder = await createReminder(input);
      setReminders(prev => [...prev, newReminder]);
      return newReminder;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create reminder';
      throw new Error(message);
    }
  }, []);

  const handleUpdateReminder = useCallback(async (id: string, input: UpdateReminderInput): Promise<Reminder> => {
    // Optimistic update
    const originalReminders = reminders;
    setReminders(prev =>
      prev.map(reminder =>
        reminder.id === id ? { ...reminder, ...input } : reminder
      )
    );

    try {
      const updatedReminder = await updateReminder(id, input);
      setReminders(prev =>
        prev.map(reminder => (reminder.id === id ? updatedReminder : reminder))
      );
      return updatedReminder;
    } catch (err) {
      // Rollback on error
      setReminders(originalReminders);
      const message = err instanceof Error ? err.message : 'Failed to update reminder';
      throw new Error(message);
    }
  }, [reminders]);

  const handleSnoozeReminder = useCallback(async (id: string, minutes: number): Promise<Reminder> => {
    // Optimistic update
    const originalReminders = reminders;
    const newRemindAt = new Date(Date.now() + minutes * 60 * 1000).toISOString();
    setReminders(prev =>
      prev.map(reminder =>
        reminder.id === id
          ? {
              ...reminder,
              remindAt: newRemindAt,
              status: 'snoozed' as const,
              snoozeCount: reminder.snoozeCount + 1,
            }
          : reminder
      )
    );

    try {
      const updatedReminder = await snoozeReminder(id, minutes);
      setReminders(prev =>
        prev.map(reminder => (reminder.id === id ? updatedReminder : reminder))
      );
      return updatedReminder;
    } catch (err) {
      // Rollback on error
      setReminders(originalReminders);
      const message = err instanceof Error ? err.message : 'Failed to snooze reminder';
      throw new Error(message);
    }
  }, [reminders]);

  const handleDismissReminder = useCallback(async (id: string): Promise<Reminder> => {
    // Optimistic update
    const originalReminders = reminders;
    const now = new Date().toISOString();
    setReminders(prev =>
      prev.map(reminder =>
        reminder.id === id
          ? { ...reminder, status: 'dismissed' as const, dismissedAt: now }
          : reminder
      )
    );

    try {
      const updatedReminder = await dismissReminder(id);
      setReminders(prev =>
        prev.map(reminder => (reminder.id === id ? updatedReminder : reminder))
      );
      return updatedReminder;
    } catch (err) {
      // Rollback on error
      setReminders(originalReminders);
      const message = err instanceof Error ? err.message : 'Failed to dismiss reminder';
      throw new Error(message);
    }
  }, [reminders]);

  const handleDeleteReminder = useCallback(async (id: string): Promise<void> => {
    // Optimistic update
    const originalReminders = reminders;
    setReminders(prev => prev.filter(reminder => reminder.id !== id));

    try {
      await deleteReminder(id);
    } catch (err) {
      // Rollback on error
      setReminders(originalReminders);
      const message = err instanceof Error ? err.message : 'Failed to delete reminder';
      throw new Error(message);
    }
  }, [reminders]);

  return {
    reminders,
    loading,
    error,
    refresh: fetchReminders,
    createReminder: handleCreateReminder,
    updateReminder: handleUpdateReminder,
    snoozeReminder: handleSnoozeReminder,
    dismissReminder: handleDismissReminder,
    deleteReminder: handleDeleteReminder,
  };
}
