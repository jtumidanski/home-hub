'use client';

import { useState, useEffect } from 'react';
import { Task, CreateTaskInput, UpdateTaskInput } from '@/lib/api/tasks';
import { Button } from '@/components/ui/button';

interface TaskFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task?: Task;
  onSubmit: (data: CreateTaskInput | UpdateTaskInput) => Promise<void>;
}

export function TaskFormDialog({ open, onOpenChange, task, onSubmit }: TaskFormDialogProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [day, setDay] = useState('');
  const [status, setStatus] = useState<'incomplete' | 'complete'>('incomplete');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditMode = !!task;

  // Initialize form when task changes or dialog opens
  useEffect(() => {
    if (open) {
      if (task) {
        setTitle(task.title);
        setDescription(task.description || '');
        setDay(task.day);
        setStatus(task.status);
      } else {
        // Default to today for new tasks
        const today = new Date().toISOString().split('T')[0];
        setTitle('');
        setDescription('');
        setDay(today);
        setStatus('incomplete');
      }
      setError(null);
    }
  }, [open, task]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!title.trim()) {
      setError('Title is required');
      return;
    }

    if (!day) {
      setError('Date is required');
      return;
    }

    setIsSubmitting(true);

    try {
      const data: CreateTaskInput = {
        title: title.trim(),
        description: description.trim(),
        day,
        status,
      };

      await onSubmit(data);
      onOpenChange(false);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to save task';
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
              {isEditMode ? 'Edit Task' : 'Create New Task'}
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

            {/* Title */}
            <div>
              <label htmlFor="title" className="block text-sm font-medium text-card-foreground mb-1.5">
                Title <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full px-3 py-2 bg-background border border-border rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="e.g., Buy groceries"
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

            {/* Date */}
            <div>
              <label htmlFor="day" className="block text-sm font-medium text-card-foreground mb-1.5">
                Date <span className="text-destructive">*</span>
              </label>
              <input
                type="date"
                id="day"
                value={day}
                onChange={(e) => setDay(e.target.value)}
                className="w-full px-3 py-2 bg-background border-2 border-border rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring [color-scheme:light] dark:[color-scheme:dark] cursor-pointer"
                required
                disabled={isSubmitting}
              />
            </div>

            {/* Status (only in edit mode) */}
            {isEditMode && (
              <div>
                <label htmlFor="status" className="block text-sm font-medium text-card-foreground mb-1.5">
                  Status
                </label>
                <select
                  id="status"
                  value={status}
                  onChange={(e) => setStatus(e.target.value as 'incomplete' | 'complete')}
                  className="w-full px-3 py-2 bg-background border border-border rounded-md text-card-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                  disabled={isSubmitting}
                >
                  <option value="incomplete">Incomplete</option>
                  <option value="complete">Complete</option>
                </select>
              </div>
            )}

            {/* Actions */}
            <div className="flex items-center gap-3 pt-2">
              <Button type="submit" disabled={isSubmitting} className="flex-1">
                {isSubmitting ? 'Saving...' : isEditMode ? 'Update Task' : 'Create Task'}
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
