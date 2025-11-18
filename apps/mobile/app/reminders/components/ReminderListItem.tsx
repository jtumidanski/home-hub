'use client';

import { useState } from 'react';
import { Clock, Check, Pencil, Trash2 } from 'lucide-react';
import { Reminder } from '@/lib/api/reminders';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

interface ReminderListItemProps {
  reminder: Reminder;
  onSnooze: (id: string, minutes: number) => Promise<void>;
  onDismiss: (id: string) => Promise<void>;
  onEdit: (reminder: Reminder) => void;
  onDelete: (id: string) => Promise<void>;
}

export function ReminderListItem({
  reminder,
  onSnooze,
  onDismiss,
  onEdit,
  onDelete,
}: ReminderListItemProps) {
  const [isActing, setIsActing] = useState(false);
  const [showSnoozeMenu, setShowSnoozeMenu] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showDismissConfirm, setShowDismissConfirm] = useState(false);

  const handleSnooze = async (minutes: number) => {
    setIsActing(true);
    setShowSnoozeMenu(false);
    try {
      await onSnooze(reminder.id, minutes);
    } catch (error) {
      console.error('Failed to snooze reminder:', error);
    } finally {
      setIsActing(false);
    }
  };

  const handleDismiss = async () => {
    setIsActing(true);
    try {
      await onDismiss(reminder.id);
    } catch (error) {
      console.error('Failed to dismiss reminder:', error);
      setIsActing(false);
    }
  };

  const handleDelete = async () => {
    setIsActing(true);
    try {
      await onDelete(reminder.id);
    } catch (error) {
      console.error('Failed to delete reminder:', error);
      setIsActing(false);
    }
  };

  const formatRelativeTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffMins = Math.round(diffMs / 60000);
    const diffHours = Math.round(diffMins / 60);
    const diffDays = Math.round(diffHours / 24);

    if (diffMins < 1) return 'Now';
    if (diffMins < 60) return `In ${diffMins}m`;
    if (diffHours < 24) return `In ${diffHours}h`;
    if (diffDays < 7) return `In ${diffDays}d`;

    if (diffMins < 0 && diffMins > -60) return `${Math.abs(diffMins)}m ago`;
    if (diffHours < 0 && diffHours > -24) return `${Math.abs(diffHours)}h ago`;
    if (diffDays < 0 && diffDays > -7) return `${Math.abs(diffDays)}d ago`;

    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
  };

  const isPast = new Date(reminder.remindAt) < new Date();
  const isDismissed = reminder.status === 'dismissed';

  return (
    <div
      className={`bg-card border rounded-lg p-4 transition-all ${
        isDismissed ? 'opacity-60' : ''
      } ${isPast && !isDismissed ? 'border-blue-500/50' : 'border-border'}`}
    >
      <div className="space-y-3">
        {/* Header */}
        <div className="flex items-start justify-between gap-2">
          <h3
            className={`font-medium text-card-foreground ${
              isDismissed ? 'line-through' : ''
            }`}
          >
            {reminder.name}
          </h3>
          <div className="flex items-center gap-1 shrink-0">
            {reminder.snoozeCount > 0 && (
              <Badge variant="secondary" className="text-xs">
                {reminder.snoozeCount}x snoozed
              </Badge>
            )}
            <Badge
              variant={
                reminder.status === 'active'
                  ? 'default'
                  : reminder.status === 'snoozed'
                  ? 'secondary'
                  : 'outline'
              }
              className="text-xs"
            >
              {formatRelativeTime(reminder.remindAt)}
            </Badge>
          </div>
        </div>

        {/* Description */}
        {reminder.description && (
          <p className="text-sm text-muted-foreground line-clamp-2">
            {reminder.description}
          </p>
        )}

        {/* Actions */}
        <div className="flex items-center gap-1 flex-wrap">
          {!isDismissed && (
            <>
              {/* Snooze Button with Menu */}
              <div className="relative">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowSnoozeMenu(!showSnoozeMenu)}
                  disabled={isActing}
                  className="h-7 text-xs"
                >
                  <Clock className="h-3.5 w-3.5 mr-1" />
                  Snooze
                </Button>

                {/* Snooze Menu */}
                {showSnoozeMenu && (
                  <div className="absolute top-full left-0 mt-1 bg-card border border-border rounded-lg shadow-lg z-10 overflow-hidden min-w-[120px]">
                    {[
                      { label: '15 min', minutes: 15 },
                      { label: '30 min', minutes: 30 },
                      { label: '1 hour', minutes: 60 },
                      { label: '2 hours', minutes: 120 },
                    ].map(({ label, minutes }) => (
                      <button
                        key={minutes}
                        onClick={() => handleSnooze(minutes)}
                        className="w-full px-3 py-2 text-left text-sm hover:bg-accent transition-colors"
                      >
                        {label}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              {/* Dismiss Button */}
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowDismissConfirm(true)}
                disabled={isActing}
                className="h-7 text-xs text-green-600 hover:text-green-600"
              >
                <Check className="h-3.5 w-3.5 mr-1" />
                Dismiss
              </Button>
            </>
          )}

          {/* Edit Button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(reminder)}
            disabled={isActing}
            className="h-7 text-xs"
          >
            <Pencil className="h-3.5 w-3.5 mr-1" />
            Edit
          </Button>

          {/* Delete Button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowDeleteConfirm(true)}
            disabled={isActing}
            className="h-7 text-xs text-destructive hover:text-destructive"
          >
            <Trash2 className="h-3.5 w-3.5 mr-1" />
            Delete
          </Button>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
        title="Delete Reminder"
        description={`Are you sure you want to delete "${reminder.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        cancelLabel="Cancel"
        variant="destructive"
        onConfirm={handleDelete}
      />

      {/* Dismiss Confirmation Dialog */}
      <ConfirmDialog
        open={showDismissConfirm}
        onOpenChange={setShowDismissConfirm}
        title="Dismiss Reminder"
        description={`Are you sure you want to dismiss "${reminder.name}"?`}
        confirmLabel="Dismiss"
        cancelLabel="Cancel"
        variant="default"
        onConfirm={handleDismiss}
      />
    </div>
  );
}
