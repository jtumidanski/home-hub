// Read-only widget — no mutations. See PRD §4.1.
import { Link } from "react-router-dom";
import { useTasks } from "@/lib/hooks/api/use-tasks";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "./widget-error";
import { Check, CheckSquare, AlertTriangle, ChevronRight } from "lucide-react";
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
      <Card className="@container h-full">
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
      <WidgetError message="Failed to load tasks" />
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
    <Card className="@container h-full">
      <CardHeader>
        <CardTitle className="font-heading text-lg @xl:text-2xl">
          <Link to="/app/tasks" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/tasks" aria-label="Open tasks" className="inline-flex size-6 items-center justify-center rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring coarse:size-11"><ChevronRight aria-hidden="true" className="size-4 @xl:size-6 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent className="space-y-3">
        {overdue.length > 0 && (
          <section>
            <h4 className="text-sm @xl:text-base font-medium text-destructive flex items-center gap-1">
              <AlertTriangle className="size-3 @xl:size-5" />
              Overdue ({overdue.length})
            </h4>
            <ul className="mt-1 space-y-1">
              {overdue.map((t) => (
                <li key={t.id} className="text-base @xl:text-xl">
                  <Link to="/app/tasks" className="inline-flex min-h-6 items-center hover:underline coarse:min-h-11">{t.attributes.title}</Link>
                </li>
              ))}
            </ul>
          </section>
        )}
        {todayTasks.length > 0 ? (
          <section>
            <h4 className="text-sm @xl:text-base font-medium text-muted-foreground">Today</h4>
            <ul className="mt-1 space-y-1">
              {todayTasks.map((t) => (
                <li key={t.id} className="text-base @xl:text-xl">
                  <Link to="/app/tasks" className="inline-flex min-h-6 items-center hover:underline coarse:min-h-11">{t.attributes.title}</Link>
                </li>
              ))}
            </ul>
          </section>
        ) : showAllCompleted ? (
          <section className="text-base @xl:text-xl text-muted-foreground">
            <p className="font-medium text-foreground">All tasks completed!</p>
            <ul className="mt-1 space-y-1">
              {completedToday.map((t) => (
                <li key={t.id} className="flex items-center gap-1.5 line-through">
                  <Check aria-hidden="true" className="size-4 shrink-0 @xl:size-5" />
                  {t.attributes.title}
                </li>
              ))}
            </ul>
          </section>
        ) : overdue.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CheckSquare className="size-5 @xl:size-6" />
            <p className="text-sm @xl:text-xl">No tasks for today</p>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
