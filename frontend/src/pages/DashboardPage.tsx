import { useTaskSummary } from "@/lib/hooks/api/use-tasks";
import { useReminderSummary } from "@/lib/hooks/api/use-reminders";
import { useAuth } from "@/components/providers/auth-provider";
import { ErrorCard } from "@/components/common/error-card";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckSquare, Bell, AlertTriangle } from "lucide-react";

function DashboardSkeleton() {
  return (
    <div className="p-6 space-y-6" role="status" aria-label="Loading">
      <div>
        <Skeleton className="h-8 w-40" />
        <Skeleton className="mt-1 h-4 w-32" />
      </div>
      <div className="grid gap-4 md:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
    </div>
  );
}

export function DashboardPage() {
  const { appContext } = useAuth();
  const { data: taskData, isLoading: taskLoading, isError: taskError } = useTaskSummary();
  const { data: reminderData, isLoading: reminderLoading, isError: reminderError } = useReminderSummary();

  const isLoading = taskLoading || reminderLoading;
  const taskSummary = taskData?.data?.attributes;
  const reminderSummary = reminderData?.data?.attributes;

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-muted-foreground">
          {appContext?.attributes.resolvedRole && `You are ${appContext.attributes.resolvedRole}`}
        </p>
      </div>

      {(taskError || reminderError) && (
        <ErrorCard message="Failed to load some dashboard data. Try refreshing the page." />
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
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

        <Card>
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

        <Card>
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
      </div>
    </div>
  );
}
