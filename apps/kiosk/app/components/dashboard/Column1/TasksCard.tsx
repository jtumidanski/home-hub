'use client';

import React, { useState } from 'react';
import { Card, CardSection } from '@/app/components/ui/Card';
import { Task, completeTask, isTaskOverdue, isTaskToday } from '@/lib/api/tasks';
import { CheckCircle2, Circle, AlertCircle } from 'lucide-react';

interface TasksCardProps {
  tasks?: Task[] | null;
  loading?: boolean;
  error?: string | null;
  userNames?: Map<string, string>; // userId -> displayName
}

export function TasksCard({ tasks, loading, error, userNames }: TasksCardProps) {
  const [completingTasks, setCompletingTasks] = useState<Set<string>>(new Set());
  const [completionErrors, setCompletionErrors] = useState<Map<string, string>>(new Map());

  const handleToggleTask = async (task: Task) => {
    // Don't allow toggling already completed tasks
    if (task.status === 'complete') return;

    // Mark as completing
    setCompletingTasks(prev => new Set(prev).add(task.id));
    setCompletionErrors(prev => {
      const next = new Map(prev);
      next.delete(task.id);
      return next;
    });

    try {
      await completeTask(task.id);
      // Success - the polling will update the tasks list
    } catch (err) {
      // Handle error - show error message
      const errorMessage = err instanceof Error ? err.message : 'Failed to complete task';
      setCompletionErrors(prev => new Map(prev).set(task.id, errorMessage));
    } finally {
      setCompletingTasks(prev => {
        const next = new Set(prev);
        next.delete(task.id);
        return next;
      });
    }
  };

  if (error) {
    return (
      <Card>
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-semibold text-gray-900 dark:text-white">Tasks</h3>
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

  // Filter tasks for today and overdue
  const todayTasks = tasks.filter(t => isTaskToday(t) && t.status === 'incomplete');
  const overdueTasks = tasks.filter(t => isTaskOverdue(t));
  const completedToday = tasks.filter(t => isTaskToday(t) && t.status === 'complete');

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold text-gray-900 dark:text-white">Tasks</h3>
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
                  onToggle={handleToggleTask}
                  overdue
                  isCompleting={completingTasks.has(task.id)}
                  error={completionErrors.get(task.id)}
                  userName={userNames?.get(task.userId)}
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
                  onToggle={handleToggleTask}
                  isCompleting={completingTasks.has(task.id)}
                  error={completionErrors.get(task.id)}
                  userName={userNames?.get(task.userId)}
                />
              ))
            ) : completedToday.length > 0 ? (
              <>
                <p className="text-sm text-gray-500 dark:text-gray-400 italic mb-2">
                  All tasks completed!
                </p>
                {completedToday.map(task => (
                  <TaskItem
                    key={task.id}
                    task={task}
                    onToggle={handleToggleTask}
                    isCompleting={false}
                    userName={userNames?.get(task.userId)}
                  />
                ))}
              </>
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
  onToggle: (task: Task) => void;
  overdue?: boolean;
  isCompleting?: boolean;
  error?: string;
  userName?: string;
}

function TaskItem({ task, onToggle, overdue = false, isCompleting = false, error, userName }: TaskItemProps) {
  const isComplete = task.status === 'complete';

  return (
    <div>
      <div
        className={`flex items-start gap-3 p-2 rounded transition-colors ${
          overdue ? 'bg-red-50 dark:bg-red-900/10' : ''
        } ${
          !isComplete ? 'hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer' : ''
        } ${
          isCompleting ? 'opacity-50' : ''
        }`}
        onClick={() => !isCompleting && onToggle(task)}
      >
        {isComplete ? (
          <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
        ) : isCompleting ? (
          <div className="h-5 w-5 flex-shrink-0 mt-0.5">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-gray-400"></div>
          </div>
        ) : (
          <Circle className="h-5 w-5 text-gray-400 dark:text-gray-600 flex-shrink-0 mt-0.5" />
        )}
        <div className="flex-1 min-w-0">
          <p
            className={`text-sm ${
              isComplete
                ? 'line-through text-gray-500 dark:text-gray-400'
                : 'text-gray-900 dark:text-white'
            }`}
          >
            {task.title}
          </p>
          {task.description && (
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              {task.description}
            </p>
          )}
          {userName && (
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              {userName}
            </p>
          )}
        </div>
      </div>
      {error && (
        <p className="text-xs text-red-600 dark:text-red-400 mt-1 ml-8">
          {error}
        </p>
      )}
    </div>
  );
}
