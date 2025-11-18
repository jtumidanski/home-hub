'use client';

import { Task } from '@/lib/api/tasks';
import { TaskListItem } from './TaskListItem';

interface TasksListProps {
  tasks: Task[];
  onComplete: (id: string) => Promise<void>;
  onEdit: (task: Task) => void;
  onDelete: (id: string) => Promise<void>;
}

export function TasksList({ tasks, onComplete, onEdit, onDelete }: TasksListProps) {
  // Get today's date
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const todayStr = today.toISOString().split('T')[0];

  // Group tasks
  const overdue = tasks.filter(
    task => task.day < todayStr && task.status === 'incomplete'
  );
  const todayTasks = tasks.filter(
    task => task.day === todayStr && task.status === 'incomplete'
  );
  const future = tasks.filter(
    task => task.day > todayStr && task.status === 'incomplete'
  );
  const completed = tasks.filter(task => task.status === 'complete');

  if (tasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 px-4">
        <div className="text-center">
          <div className="mb-4 text-4xl">📝</div>
          <h3 className="text-lg font-semibold text-card-foreground mb-2">
            No tasks yet
          </h3>
          <p className="text-sm text-muted-foreground">
            Create your first task to get started
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Overdue */}
      {overdue.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-destructive mb-3 px-1">
            Overdue ({overdue.length})
          </h2>
          <div className="space-y-2">
            {overdue.map(task => (
              <TaskListItem
                key={task.id}
                task={task}
                onComplete={onComplete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Today */}
      {todayTasks.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-blue-600 dark:text-blue-500 mb-3 px-1">
            Today ({todayTasks.length})
          </h2>
          <div className="space-y-2">
            {todayTasks.map(task => (
              <TaskListItem
                key={task.id}
                task={task}
                onComplete={onComplete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Future */}
      {future.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-muted-foreground mb-3 px-1">
            Upcoming ({future.length})
          </h2>
          <div className="space-y-2">
            {future.map(task => (
              <TaskListItem
                key={task.id}
                task={task}
                onComplete={onComplete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Completed */}
      {completed.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-green-600 dark:text-green-500 mb-3 px-1">
            Completed ({completed.length})
          </h2>
          <div className="space-y-2">
            {completed.map(task => (
              <TaskListItem
                key={task.id}
                task={task}
                onComplete={onComplete}
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
