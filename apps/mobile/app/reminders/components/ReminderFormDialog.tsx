'use client';

import { useState, useEffect } from 'react';
import { Reminder, CreateReminderInput, UpdateReminderInput } from '@/lib/api/reminders';
import { Button } from '@/components/ui/button';

interface ReminderFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  reminder?: Reminder;
  onSubmit: (data: CreateReminderInput | UpdateReminderInput) => Promise<void>;
}

export function ReminderFormDialog({ open, onOpenChange, reminder, onSubmit }: ReminderFormDialogProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [remindAt, setRemindAt] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditMode = !!reminder;

  // Initialize form when reminder changes or dialog opens
  useEffect(() => {
    if (open) {
      if (reminder) {
        setName(reminder.name);
        setDescription(reminder.description || '');
        // Convert ISO to datetime-local format
        const date = new Date(reminder.remindAt);
        const localDateTime = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
          .toISOString()
          .slice(0, 16);
        setRemindAt(localDateTime);
      } else {
        // Default to 1 hour from now
        const oneHourFromNow = new Date(Date.now() + 60 * 60 * 1000);
        const localDateTime = new Date(oneHourFromNow.getTime() - oneHourFromNow.getTimezoneOffset() * 60000)
          .toISOString()
          .slice(0, 16);
        setName('');
        setDescription('');
        setRemindAt(localDateTime);
      }
      setError(null);
    }
  }, [open, reminder]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Name is required');
      return;
    }

    if (!remindAt) {
      setError('Reminder time is required');
      return;
    }

    // Convert datetime-local to ISO
    const reminderDate = new Date(remindAt);
    const now = new Date();

    if (reminderDate <= now && !isEditMode) {
      setError('Reminder time must be in the future');
      return;
    }

    setIsSubmitting(true);

    try {
      const data: CreateReminderInput = {
        name: name.trim(),
        description: description.trim(),
        remindAt: reminderDate.toISOString(),
      };

      await onSubmit(data);
      onOpenChange(false);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to save reminder';
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm" onClick={() => onOpenChange(false)}>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="w-full max-w-lg bg-card border border-border rounded-lg shadow-lg" onClick={(e) => e.stopPropagation()}>
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-border">
            <h2 className="text-lg font-semibold text-card-foreground">
              {isEditMode ? 'Edit Reminder' : 'Create New Reminder'}
            </h2>
            <button
              onClick={() => onOpenChange(false)}
              className="text-muted-foreground hover:text-card-foreground"
              disabled={isSubmitting}
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {error && (
              <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            )}

            {/* Name */}
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-card-foreground mb-1.5">
                Name <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-3 py-2 bg-background border border-border rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="e.g., Team meeting"
                required
                disabled={isSubmitting}
              />
            </div>

            {/* Description */}
            <div>
              <label htmlFor="description" className="block text-sm font-medium text-card-foreground mb-1.5">
                Description
              </label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                className="w-full px-3 py-2 bg-background border border-border rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                placeholder="Add more details..."
                disabled={isSubmitting}
              />
            </div>

            {/* Remind At */}
            <div>
              <label htmlFor="remindAt" className="block text-sm font-medium text-card-foreground mb-1.5">
                Remind At <span className="text-destructive">*</span>
              </label>
              <input
                type="datetime-local"
                id="remindAt"
                value={remindAt}
                onChange={(e) => setRemindAt(e.target.value)}
                className="w-full px-3 py-2 bg-background border-2 border-border rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring [color-scheme:light] dark:[color-scheme:dark] cursor-pointer"
                required
                disabled={isSubmitting}
              />
            </div>

            {/* Actions */}
            <div className="flex items-center gap-3 pt-2">
              <Button type="submit" disabled={isSubmitting} className="flex-1">
                {isSubmitting ? 'Saving...' : isEditMode ? 'Update Reminder' : 'Create Reminder'}
              </Button>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting} className="flex-1">
                Cancel
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
