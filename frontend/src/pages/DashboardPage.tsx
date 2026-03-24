import { useTaskSummary } from "@/lib/hooks/api/use-tasks";
import { useReminderSummary } from "@/lib/hooks/api/use-reminders";
import { useAuth } from "@/components/providers/auth-provider";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CheckSquare, Bell, AlertTriangle } from "lucide-react";

export function DashboardPage() {
  const { appContext } = useAuth();
  const { data: taskData } = useTaskSummary();
  const { data: reminderData } = useReminderSummary();

  const taskSummary = taskData?.data?.attributes;
  const reminderSummary = reminderData?.data?.attributes;

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-muted-foreground">
          {appContext?.attributes.resolvedRole && `You are ${appContext.attributes.resolvedRole}`}
        </p>
      </div>

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
