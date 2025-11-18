'use client';

import { useState, useEffect, useCallback } from 'react';
import {
  getTasks,
  createTask,
  updateTask,
  completeTask,
  deleteTask,
  type Task,
  type CreateTaskInput,
  type UpdateTaskInput,
} from '@/lib/api/tasks';

export function useTasks() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getTasks();
      setTasks(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch tasks';
      setError(message);
      console.error('Error fetching tasks:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleCreateTask = useCallback(async (input: CreateTaskInput): Promise<Task> => {
    try {
      const newTask = await createTask(input);
      setTasks(prev => [...prev, newTask]);
      return newTask;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create task';
      throw new Error(message);
    }
  }, []);

  const handleUpdateTask = useCallback(async (id: string, input: UpdateTaskInput): Promise<Task> => {
    // Optimistic update
    const originalTasks = tasks;
    setTasks(prev =>
      prev.map(task =>
        task.id === id ? { ...task, ...input } : task
      )
    );

    try {
      const updatedTask = await updateTask(id, input);
      setTasks(prev =>
        prev.map(task => (task.id === id ? updatedTask : task))
      );
      return updatedTask;
    } catch (err) {
      // Rollback on error
      setTasks(originalTasks);
      const message = err instanceof Error ? err.message : 'Failed to update task';
      throw new Error(message);
    }
  }, [tasks]);

  const handleCompleteTask = useCallback(async (id: string): Promise<Task> => {
    // Optimistic update
    const originalTasks = tasks;
    const now = new Date().toISOString();
    setTasks(prev =>
      prev.map(task =>
        task.id === id
          ? { ...task, status: 'complete' as const, completedAt: now }
          : task
      )
    );

    try {
      const updatedTask = await completeTask(id);
      setTasks(prev =>
        prev.map(task => (task.id === id ? updatedTask : task))
      );
      return updatedTask;
    } catch (err) {
      // Rollback on error
      setTasks(originalTasks);
      const message = err instanceof Error ? err.message : 'Failed to complete task';
      throw new Error(message);
    }
  }, [tasks]);

  const handleDeleteTask = useCallback(async (id: string): Promise<void> => {
    // Optimistic update
    const originalTasks = tasks;
    setTasks(prev => prev.filter(task => task.id !== id));

    try {
      await deleteTask(id);
    } catch (err) {
      // Rollback on error
      setTasks(originalTasks);
      const message = err instanceof Error ? err.message : 'Failed to delete task';
      throw new Error(message);
    }
  }, [tasks]);

  return {
    tasks,
    loading,
    error,
    refresh: fetchTasks,
    createTask: handleCreateTask,
    updateTask: handleUpdateTask,
    completeTask: handleCompleteTask,
    deleteTask: handleDeleteTask,
  };
}
