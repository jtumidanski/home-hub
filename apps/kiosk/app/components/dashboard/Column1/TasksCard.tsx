'use client';

import React, { useState } from 'react';
import { Card, CardSection } from '@/app/components/ui/Card';
import { Task } from '@/lib/api/tasks';
import { CheckCircle2, Circle, AlertCircle } from 'lucide-react';

interface TasksCardProps {
  tasks?: Task[] | null;
  loading?: boolean;
  error?: string | null;
}

export function TasksCard({ tasks, loading, error }: TasksCardProps) {
  const [completedTasks, setCompletedTasks] = useState<Set<string>>(new Set());

  const handleToggleTask = (taskId: string) => {
    setCompletedTasks(prev => {
      const next = new Set(prev);
      if (next.has(taskId)) {
        next.delete(taskId);
      } else {
        next.add(taskId);
      }
      return next;
    });
  };

  if (error) {
    return (
      <Card>
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-semibold text-gray-900 dark:text-white">Tasks</h3>
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

  if (loading || !tasks) {
    return <Card loading={true}>{null}</Card>;
  }

  const todayTasks = tasks.filter(t => t.status === 'pending' || completedTasks.has(t.id));
  const overdueTasks = tasks.filter(t => t.status === 'overdue' && !completedTasks.has(t.id));

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold text-gray-900 dark:text-white">Tasks</h3>
        <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
          Preview
        </span>
      </div>

      <div className="space-y-4">
        {/* Overdue Tasks */}
        {overdueTasks.length > 0 && (
          <CardSection>
            <div className="flex items-center gap-2 mb-2">
              <AlertCircle className="h-4 w-4 text-red-600 dark:text-red-400" />
              <span className="text-sm font-medium text-red-600 dark:text-red-400">
                Overdue ({overdueTasks.length})
              </span>
            </div>
            <div className="space-y-2">
              {overdueTasks.map(task => (
                <TaskItem
                  key={task.id}
                  task={task}
                  completed={completedTasks.has(task.id)}
                  onToggle={handleToggleTask}
                  overdue
                />
              ))}
            </div>
          </CardSection>
        )}

        {/* Today's Tasks */}
        <CardSection>
          <div className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Today
          </div>
          <div className="space-y-2">
            {todayTasks.length > 0 ? (
              todayTasks.map(task => (
                <TaskItem
                  key={task.id}
                  task={task}
                  completed={completedTasks.has(task.id) || task.status === 'completed'}
                  onToggle={handleToggleTask}
                />
              ))
            ) : (
              <p className="text-sm text-gray-500 dark:text-gray-400 italic">
                No tasks for today
              </p>
            )}
          </div>
        </CardSection>
      </div>
    </Card>
  );
}

interface TaskItemProps {
  task: Task;
  completed: boolean;
  onToggle: (taskId: string) => void;
  overdue?: boolean;
}

function TaskItem({ task, completed, onToggle, overdue = false }: TaskItemProps) {
  return (
    <div
      className={`flex items-start gap-3 p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors cursor-pointer ${
        overdue ? 'bg-red-50 dark:bg-red-900/10' : ''
      }`}
      onClick={() => onToggle(task.id)}
    >
      {completed ? (
        <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
      ) : (
        <Circle className="h-5 w-5 text-gray-400 dark:text-gray-600 flex-shrink-0 mt-0.5" />
      )}
      <div className="flex-1 min-w-0">
        <p
          className={`text-sm ${
            completed
              ? 'line-through text-gray-500 dark:text-gray-400'
              : 'text-gray-900 dark:text-white'
          }`}
        >
          {task.title}
        </p>
        {task.assignee && (
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {task.assignee}
          </p>
        )}
      </div>
    </div>
  );
}
