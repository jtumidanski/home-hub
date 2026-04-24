import { useCallback, useEffect } from "react";
import { Link } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useTaskSummary } from "@/lib/hooks/api/use-tasks";
import { useReminderSummary } from "@/lib/hooks/api/use-reminders";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { mealKeys } from "@/lib/hooks/api/use-meals";
import { calendarKeys } from "@/lib/hooks/api/use-calendar";
import { trackerKeys } from "@/lib/hooks/api/use-trackers";
import { workoutKeys } from "@/lib/hooks/api/use-workouts";
import { packageKeys } from "@/lib/hooks/api/use-packages";

import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { ErrorCard } from "@/components/common/error-card";
import { DashboardSkeleton } from "@/components/common/dashboard-skeleton";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CheckSquare, Bell, AlertTriangle } from "lucide-react";
import { WeatherWidget } from "@/components/features/weather/weather-widget";
import { PackageSummaryWidget } from "@/components/features/packages/package-summary-widget";
import { MealPlanWidget } from "@/components/features/meals/meal-plan-widget";
import { CalendarWidget } from "@/components/features/calendar/calendar-widget";
import { HabitsWidget } from "@/components/features/trackers/habits-widget";
import { WorkoutWidget } from "@/components/features/workouts/workout-widget";

export function DashboardPage() {
  const queryClient = useQueryClient();
  const { tenant, household } = useTenant();

  const today = useLocalDate(household?.attributes.timezone);
  const { data: taskData, isLoading: taskLoading, isError: taskError, refetch: refetchTasks } = useTaskSummary(today);
  const { data: reminderData, isLoading: reminderLoading, isError: reminderError, refetch: refetchReminders } = useReminderSummary();

  // Invalidate widget queries on mount so navigating back always shows fresh data
  useEffect(() => {
    queryClient.invalidateQueries({ queryKey: packageKeys.summary(tenant, household) });
    queryClient.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) });
    queryClient.invalidateQueries({ queryKey: calendarKeys.all(tenant, household) });
    queryClient.invalidateQueries({ queryKey: trackerKeys.todayAll(tenant, household) });
    queryClient.invalidateQueries({ queryKey: workoutKeys.todayAll(tenant, household) });
  }, [queryClient, tenant, household]);

  const isLoading = taskLoading || reminderLoading;
  const taskSummary = taskData?.data?.attributes;
  const reminderSummary = reminderData?.data?.attributes;

  const handleRefresh = useCallback(async () => {
    await Promise.all([
      refetchTasks(),
      refetchReminders(),
      queryClient.invalidateQueries({ queryKey: packageKeys.summary(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: calendarKeys.all(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: trackerKeys.todayAll(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: workoutKeys.todayAll(tenant, household) }),
    ]);
  }, [refetchTasks, refetchReminders, queryClient, tenant, household]);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-6">
        <div>
          <h1 className="text-xl md:text-2xl font-semibold">Dashboard</h1>
        </div>

        {(taskError || reminderError) && (
          <ErrorCard message="Failed to load some dashboard data. Try refreshing the page." />
        )}

        <WeatherWidget />

        {/* Tasks / Reminders / Overdue — 3-column row */}
        <div className="grid gap-6 md:gap-4 grid-cols-1 md:grid-cols-3">
          <Link to="/app/tasks?status=pending" className="transition-opacity hover:opacity-80">
            <Card className="h-full">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Pending Tasks</CardTitle>
                <CheckSquare className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{taskSummary?.pendingCount ?? "-"}</div>
                <p className="text-xs text-muted-foreground">
                  {taskSummary?.completedTodayCount ?? 0} completed today
                </p>
              </CardContent>
            </Card>
          </Link>

          <Link to="/app/reminders?status=active" className="transition-opacity hover:opacity-80">
            <Card className="h-full">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Active Reminders</CardTitle>
                <Bell className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{reminderSummary?.dueNowCount ?? "-"}</div>
                <p className="text-xs text-muted-foreground">
                  {reminderSummary?.upcomingCount ?? 0} upcoming
                </p>
              </CardContent>
            </Card>
          </Link>

          <Link to="/app/tasks?status=overdue" className="transition-opacity hover:opacity-80">
            <Card className="h-full">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Overdue</CardTitle>
                <AlertTriangle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{taskSummary?.overdueCount ?? "-"}</div>
                <p className="text-xs text-muted-foreground">
                  {reminderSummary?.snoozedCount ?? 0} snoozed reminders
                </p>
              </CardContent>
            </Card>
          </Link>
        </div>

        {/* Meals / Habits / Packages — 3-column row, equal height */}
        <div className="grid gap-6 md:gap-4 grid-cols-1 md:grid-cols-3 md:auto-rows-fr">
          <MealPlanWidget />
          <HabitsWidget />
          <PackageSummaryWidget />
        </div>

        {/* Calendar / Workout — 2-column row, equal height */}
        <div className="grid gap-6 md:gap-4 grid-cols-1 md:grid-cols-2 md:auto-rows-fr">
          <CalendarWidget />
          <WorkoutWidget />
        </div>
      </div>
    </PullToRefresh>
  );
}
