// Read-only widget — no mutations. See PRD §4.1.
import { Link } from "react-router-dom";
import { useTasks } from "@/lib/hooks/api/use-tasks";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { CheckSquare, AlertTriangle, ChevronRight } from "lucide-react";
import type { Task } from "@/types/models/task";

export interface TasksTodayConfig {
  title?: string | undefined;
  includeCompleted: boolean;
}

export function TasksTodayAdapter({ config }: { config: TasksTodayConfig }) {
  const { household } = useTenant();
  const today = useLocalDate(household?.attributes.timezone);
  const { data, isLoading, isError } = useTasks();

  const title = config.title?.trim() || "Today's Tasks";

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader><Skeleton className="h-5 w-28" data-slot="skeleton" /></CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" data-slot="skeleton" />
          <Skeleton className="h-4 w-3/4" data-slot="skeleton" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return (
      <Card className="h-full border-destructive">
        <CardContent className="py-3"><p className="text-sm text-destructive">Failed to load tasks</p></CardContent>
      </Card>
    );
  }

  const all = (data?.data ?? []) as Task[];
  const overdue = all.filter((t) => t.attributes.status === "pending" && t.attributes.dueOn && t.attributes.dueOn < today);
  const todayTasks = all.filter((t) => t.attributes.status === "pending" && t.attributes.dueOn === today);
  const completedToday = all.filter((t) =>
    t.attributes.status === "completed" &&
    t.attributes.completedAt &&
    t.attributes.completedAt.slice(0, 10) === today,
  );

  const showAllCompleted =
    config.includeCompleted &&
    todayTasks.length === 0 &&
    overdue.length === 0 &&
    completedToday.length > 0;

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          <Link to="/app/tasks" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/tasks"><ChevronRight className="h-4 w-4 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent className="space-y-3">
        {overdue.length > 0 && (
          <section>
            <h4 className="text-xs font-medium text-destructive flex items-center gap-1">
              <AlertTriangle className="h-3 w-3" />
              Overdue ({overdue.length})
            </h4>
            <ul className="mt-1 space-y-1">
              {overdue.map((t) => (
                <li key={t.id} className="text-sm">
                  <Link to="/app/tasks" className="hover:underline">{t.attributes.title}</Link>
                </li>
              ))}
            </ul>
          </section>
        )}
        {todayTasks.length > 0 ? (
          <section>
            <h4 className="text-xs font-medium text-muted-foreground">Today</h4>
            <ul className="mt-1 space-y-1">
              {todayTasks.map((t) => (
                <li key={t.id} className="text-sm">
                  <Link to="/app/tasks" className="hover:underline">{t.attributes.title}</Link>
                </li>
              ))}
            </ul>
          </section>
        ) : showAllCompleted ? (
          <section className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground">All tasks completed!</p>
            <ul className="mt-1 space-y-1 opacity-70">
              {completedToday.map((t) => (
                <li key={t.id}>{t.attributes.title}</li>
              ))}
            </ul>
          </section>
        ) : overdue.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CheckSquare className="h-5 w-5" />
            <p className="text-sm">No tasks for today</p>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
