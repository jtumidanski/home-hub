// Read-only widget — no mutations. See PRD §4.5.
import { Link } from "react-router-dom";
import { useTasks } from "@/lib/hooks/api/use-tasks";
import { useTenant } from "@/context/tenant-context";
import { useLocalDateOffset } from "@/lib/hooks/use-local-date-offset";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "./widget-error";
import { CheckSquare, ChevronRight } from "lucide-react";
import type { Task } from "@/types/models/task";

export interface TasksTomorrowConfig {
  title?: string | undefined;
  limit: number;
}

export function TasksTomorrowAdapter({ config }: { config: TasksTomorrowConfig }) {
  const { household } = useTenant();
  const tomorrow = useLocalDateOffset(household?.attributes.timezone, 1);
  const { data, isLoading, isError } = useTasks();
  const title = config.title?.trim() || "Tomorrow's Tasks";

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
  const due = all.filter((t) => t.attributes.status === "pending" && t.attributes.dueOn === tomorrow);
  const visible = due.slice(0, config.limit);
  const remainder = due.length - visible.length;

  return (
    <Card className="@container h-full">
      <CardHeader>
        <CardTitle className="font-heading text-lg @xl:text-2xl">
          <Link to="/app/tasks" className="hover:underline">{title}</Link>
        </CardTitle>
        <CardAction><Link to="/app/tasks" aria-label="Open tasks" className="inline-flex size-6 items-center justify-center rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring coarse:size-11"><ChevronRight aria-hidden="true" className="size-4 @xl:size-6 text-muted-foreground" /></Link></CardAction>
      </CardHeader>
      <CardContent>
        {visible.length === 0 ? (
          <div className="flex items-center gap-2 text-muted-foreground">
            <CheckSquare className="size-5 @xl:size-6" />
            <p className="text-sm @xl:text-xl">No tasks for tomorrow</p>
          </div>
        ) : (
          <ul className="space-y-1">
            {visible.map((t) => (
              <li key={t.id} className="text-base @xl:text-xl">
                <Link to="/app/tasks" className="inline-flex min-h-6 items-center hover:underline coarse:min-h-11">{t.attributes.title}</Link>
              </li>
            ))}
            {remainder > 0 && <li className="text-sm @xl:text-base text-muted-foreground">+{remainder} more</li>}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
