import { useState, useCallback, useRef } from 'react';
import { usePolling } from './usePolling';
import { getWeather, WeatherResponse } from '../api/weather';
import { getTasks, Task } from '../api/tasks';
import { getMealPlan, MealPlan } from '../api/meals';
import { getCalendarEvents, CalendarEvent } from '../api/calendar';
import { getReminders, Reminder } from '../api/reminders';

export interface DashboardData {
  weather: WeatherResponse | null;
  tasks: Task[] | null;
  mealPlan: MealPlan | null;
  todayEvents: CalendarEvent[] | null;
  tomorrowEvents: CalendarEvent[] | null;
  reminders: Reminder[] | null;
}

export interface DashboardState {
  data: DashboardData;
  loading: {
    weather: boolean;
    tasks: boolean;
    mealPlan: boolean;
    todayEvents: boolean;
    tomorrowEvents: boolean;
    reminders: boolean;
  };
  errors: {
    weather: string | null;
    tasks: string | null;
    mealPlan: string | null;
    todayEvents: string | null;
    tomorrowEvents: string | null;
    reminders: string | null;
  };
  isRefreshing: boolean;
  lastUpdate: Date | null;
}

/**
 * Hook to fetch and manage all dashboard data with 20-second polling
 */
export function useDashboardData(householdId?: string) {
  const [state, setState] = useState<DashboardState>({
    data: {
      weather: null,
      tasks: null,
      mealPlan: null,
      todayEvents: null,
      tomorrowEvents: null,
      reminders: null,
    },
    loading: {
      weather: true,
      tasks: true,
      mealPlan: true,
      todayEvents: true,
      tomorrowEvents: true,
      reminders: true,
    },
    errors: {
      weather: null,
      tasks: null,
      mealPlan: null,
      todayEvents: null,
      tomorrowEvents: null,
      reminders: null,
    },
    isRefreshing: false,
    lastUpdate: null,
  });

  const abortControllerRef = useRef<AbortController | null>(null);
  const cacheRef = useRef<{
    data: Partial<DashboardData>;
    timestamp: number;
  }>({
    data: {},
    timestamp: 0,
  });

  const fetchAllData = useCallback(async () => {
    // Abort any pending requests
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    abortControllerRef.current = new AbortController();
    const signal = abortControllerRef.current.signal;

    setState(prev => ({ ...prev, isRefreshing: true }));

    // Check cache (20s TTL)
    const now = Date.now();
    const cacheAge = now - cacheRef.current.timestamp;
    const useCached = cacheAge < 20000 && Object.keys(cacheRef.current.data).length > 0;

    if (useCached) {
      // Return cached data immediately
      setState(prev => ({
        ...prev,
        data: { ...prev.data, ...cacheRef.current.data } as DashboardData,
        loading: {
          weather: false,
          tasks: false,
          mealPlan: false,
          todayEvents: false,
          tomorrowEvents: false,
          reminders: false,
        },
        isRefreshing: false,
      }));
    }

    // Fetch all data in parallel
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const tomorrowStr = tomorrow.toISOString().split('T')[0];

    const promises = [
      householdId
        ? getWeather(householdId, signal).catch(err => ({ error: 'weather', message: err.message }))
        : Promise.resolve({ error: 'weather', message: 'No household ID' }),
      getTasks().catch(err => ({ error: 'tasks', message: err.message })),
      getMealPlan().catch(err => ({ error: 'mealPlan', message: err.message })),
      getCalendarEvents().catch(err => ({ error: 'todayEvents', message: err.message })),
      getCalendarEvents(tomorrowStr).catch(err => ({ error: 'tomorrowEvents', message: err.message })),
      getReminders().catch(err => ({ error: 'reminders', message: err.message })),
    ];

    const results = await Promise.allSettled(promises);

    const newData: Partial<DashboardData> = {};
    const newErrors: Partial<typeof state.errors> = {};

    // Process weather
    if (results[0].status === 'fulfilled') {
      const result = results[0].value;
      if ('error' in result) {
        newErrors.weather = result.message;
      } else {
        newData.weather = result as WeatherResponse;
        newErrors.weather = null;
      }
    }

    // Process tasks
    if (results[1].status === 'fulfilled') {
      const result = results[1].value;
      if ('error' in result) {
        newErrors.tasks = result.message;
      } else {
        newData.tasks = result as Task[];
        newErrors.tasks = null;
      }
    }

    // Process meal plan
    if (results[2].status === 'fulfilled') {
      const result = results[2].value;
      if ('error' in result) {
        newErrors.mealPlan = result.message;
      } else {
        newData.mealPlan = result as MealPlan;
        newErrors.mealPlan = null;
      }
    }

    // Process today's events
    if (results[3].status === 'fulfilled') {
      const result = results[3].value;
      if ('error' in result) {
        newErrors.todayEvents = result.message;
      } else {
        newData.todayEvents = result as CalendarEvent[];
        newErrors.todayEvents = null;
      }
    }

    // Process tomorrow's events
    if (results[4].status === 'fulfilled') {
      const result = results[4].value;
      if ('error' in result) {
        newErrors.tomorrowEvents = result.message;
      } else {
        newData.tomorrowEvents = result as CalendarEvent[];
        newErrors.tomorrowEvents = null;
      }
    }

    // Process reminders
    if (results[5].status === 'fulfilled') {
      const result = results[5].value;
      if ('error' in result) {
        newErrors.reminders = result.message;
      } else {
        newData.reminders = result as Reminder[];
        newErrors.reminders = null;
      }
    }

    // Update cache
    cacheRef.current = {
      data: newData,
      timestamp: now,
    };

    setState(prev => ({
      ...prev,
      data: { ...prev.data, ...newData } as DashboardData,
      errors: { ...prev.errors, ...newErrors } as typeof state.errors,
      loading: {
        weather: false,
        tasks: false,
        mealPlan: false,
        todayEvents: false,
        tomorrowEvents: false,
        reminders: false,
      },
      isRefreshing: false,
      lastUpdate: new Date(),
    }));
  }, [householdId]);

  // Set up polling with 20-second interval
  usePolling(fetchAllData, {
    interval: 20000,
    enabled: true,
    pauseOnHidden: true,
  });

  return {
    ...state,
    refresh: fetchAllData,
  };
}
