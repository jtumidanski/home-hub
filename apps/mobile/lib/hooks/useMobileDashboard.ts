'use client';

import { useEffect, useState, useCallback } from 'react';
import { getCurrentWeather, type WeatherData } from '@/lib/api/weather';
import { getTasks, filterTasksForToday, type Task } from '@/lib/api/tasks';
import { getReminders, filterActiveReminders, type Reminder } from '@/lib/api/reminders';

export interface DashboardData {
  weather: WeatherData | null;
  tasks: Task[];
  reminders: Reminder[];
  tasksCount: number;
  remindersCount: number;
}

export interface DashboardLoading {
  weather: boolean;
  tasks: boolean;
  reminders: boolean;
}

export interface DashboardErrors {
  weather: string | null;
  tasks: string | null;
  reminders: string | null;
}

export function useMobileDashboard(householdId?: string) {
  const [data, setData] = useState<DashboardData>({
    weather: null,
    tasks: [],
    reminders: [],
    tasksCount: 0,
    remindersCount: 0,
  });

  const [loading, setLoading] = useState<DashboardLoading>({
    weather: true,
    tasks: true,
    reminders: true,
  });

  const [errors, setErrors] = useState<DashboardErrors>({
    weather: null,
    tasks: null,
    reminders: null,
  });

  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchAllData = useCallback(async () => {
    if (!householdId) {
      return;
    }

    // Fetch all data sources in parallel
    const results = await Promise.allSettled([
      getCurrentWeather(householdId),
      getTasks(),
      getReminders(),
    ]);

    // Process weather
    const weatherResult = results[0];
    if (weatherResult.status === 'fulfilled') {
      setData(prev => ({ ...prev, weather: weatherResult.value }));
      setErrors(prev => ({ ...prev, weather: null }));
    } else {
      console.error('Weather fetch failed:', weatherResult.reason);
      setErrors(prev => ({
        ...prev,
        weather: weatherResult.reason instanceof Error ? weatherResult.reason.message : 'Failed to load weather'
      }));
    }
    setLoading(prev => ({ ...prev, weather: false }));

    // Process tasks
    const tasksResult = results[1];
    if (tasksResult.status === 'fulfilled') {
      const allTasks = tasksResult.value;
      const todayTasks = filterTasksForToday(allTasks);
      setData(prev => ({
        ...prev,
        tasks: allTasks,
        tasksCount: todayTasks.length
      }));
      setErrors(prev => ({ ...prev, tasks: null }));
    } else {
      console.error('Tasks fetch failed:', tasksResult.reason);
      setErrors(prev => ({
        ...prev,
        tasks: tasksResult.reason instanceof Error ? tasksResult.reason.message : 'Failed to load tasks'
      }));
    }
    setLoading(prev => ({ ...prev, tasks: false }));

    // Process reminders
    const remindersResult = results[2];
    if (remindersResult.status === 'fulfilled') {
      const allReminders = remindersResult.value;
      const activeReminders = filterActiveReminders(allReminders);
      setData(prev => ({
        ...prev,
        reminders: allReminders,
        remindersCount: activeReminders.length
      }));
      setErrors(prev => ({ ...prev, reminders: null }));
    } else {
      console.error('Reminders fetch failed:', remindersResult.reason);
      setErrors(prev => ({
        ...prev,
        reminders: remindersResult.reason instanceof Error ? remindersResult.reason.message : 'Failed to load reminders'
      }));
    }
    setLoading(prev => ({ ...prev, reminders: false }));

    setIsRefreshing(false);
  }, [householdId]);

  // Initial fetch and polling
  useEffect(() => {
    if (!householdId) {
      return;
    }

    // Initial fetch
    fetchAllData();

    // Set up polling every 30 seconds
    const intervalId = setInterval(fetchAllData, 30000);

    // Cleanup on unmount
    return () => clearInterval(intervalId);
  }, [householdId, fetchAllData]);

  const refresh = useCallback(() => {
    setIsRefreshing(true);
    fetchAllData();
  }, [fetchAllData]);

  return {
    data,
    loading,
    errors,
    isRefreshing,
    refresh,
  };
}
