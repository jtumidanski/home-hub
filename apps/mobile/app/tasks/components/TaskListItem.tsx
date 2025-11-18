'use client';

import { useState } from 'react';
import { Check, Pencil, Trash2 } from 'lucide-react';
import { Task } from '@/lib/api/tasks';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

interface TaskListItemProps {
  task: Task;
  onComplete: (id: string) => Promise<void>;
  onEdit: (task: Task) => void;
  onDelete: (id: string) => Promise<void>;
}

export function TaskListItem({ task, onComplete, onEdit, onDelete }: TaskListItemProps) {
  const [isActing, setIsActing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const handleComplete = async () => {
    if (task.status === 'complete') return;

    setIsActing(true);
    try {
      await onComplete(task.id);
    } catch (error) {
      console.error('Failed to complete task:', error);
    } finally {
      setIsActing(false);
    }
  };

  const handleDelete = async () => {
    setIsActing(true);
    try {
      await onDelete(task.id);
    } catch (error) {
      console.error('Failed to delete task:', error);
      setIsActing(false);
    }
  };

  const formatDate = (dateStr: string) => {
    // Parse date string as local date (YYYY-MM-DD format)
    const [year, month, day] = dateStr.split('-').map(Number);
    const taskDate = new Date(year, month - 1, day);

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    taskDate.setHours(0, 0, 0, 0);

    const diffTime = taskDate.getTime() - today.getTime();
    const diffDays = Math.round(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Tomorrow';
    if (diffDays === -1) return 'Yesterday';
    if (diffDays < -1) return `${Math.abs(diffDays)} days ago`;
    if (diffDays > 1) return `In ${diffDays} days`;

    return taskDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const isOverdue = task.day < new Date().toISOString().split('T')[0] && task.status === 'incomplete';

  return (
    <div
      className={`bg-card border rounded-lg p-4 transition-all ${
        task.status === 'complete' ? 'opacity-60' : ''
      } ${isOverdue ? 'border-destructive/50' : 'border-border'}`}
    >
      <div className="space-y-3">
        {/* Header */}
        <div className="flex items-start justify-between gap-2">
          <h3
            className={`font-medium text-card-foreground ${
              task.status === 'complete' ? 'line-through' : ''
            }`}
          >
            {task.title}
          </h3>
          <div className="flex items-center gap-1 shrink-0">
            <Badge
              variant={task.status === 'complete' ? 'secondary' : 'default'}
              className="text-xs"
            >
              {formatDate(task.day)}
            </Badge>
          </div>
        </div>

        {/* Description */}
        {task.description && (
          <p className="text-sm text-muted-foreground line-clamp-2">
            {task.description}
          </p>
        )}

        {/* Actions */}
        <div className="flex items-center gap-1 flex-wrap">
          {task.status === 'incomplete' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleComplete}
              disabled={isActing}
              className="h-7 text-xs text-green-600 hover:text-green-600"
            >
              <Check className="h-3.5 w-3.5 mr-1" />
              Complete
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(task)}
            disabled={isActing}
            className="h-7 text-xs"
          >
            <Pencil className="h-3.5 w-3.5 mr-1" />
            Edit
          </Button>
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
        title="Delete Task"
        description={`Are you sure you want to delete "${task.title}"? This action cannot be undone.`}
        confirmLabel="Delete"
        cancelLabel="Cancel"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  );
}
