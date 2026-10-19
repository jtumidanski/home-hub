import { Link } from "react-router-dom";
import { useWorkoutToday } from "@/lib/hooks/api/use-workouts";
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { WidgetError } from "@/components/features/dashboard-widgets/widget-error";
import { cn } from "@/lib/utils";
import { Dumbbell, BedDouble, Check, Circle, CircleDot, SkipForward, ChevronRight } from "lucide-react";

export function WorkoutWidget() {
  const { household } = useTenant();
  const today = useLocalDate(household?.attributes.timezone);
  const { data, isLoading, isError } = useWorkoutToday(today);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-28" />
        </CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </CardContent>
      </Card>
    );
  }

  if (isError) {
    return <WidgetError message="Failed to load workout" />;
  }

  const attrs = data?.data?.attributes;
  const isRestDay = attrs?.isRestDay ?? false;
  const items = attrs?.items ?? [];

  return (
    <Link to="/app/workouts" className="@container group/workout block h-full rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card className="@container h-full motion-safe:transition-colors group-hover/workout:bg-accent">
        <CardHeader>
          <CardTitle className="font-heading text-lg @xl:text-2xl">Workout</CardTitle>
          <CardAction>
            <ChevronRight className="size-4 text-muted-foreground" aria-hidden="true" />
          </CardAction>
        </CardHeader>
        <CardContent>
          {isRestDay ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <BedDouble className="size-5 @xl:size-6 shrink-0" />
              <p className="text-sm">Rest Day</p>
            </div>
          ) : items.length === 0 ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Dumbbell className="size-5 @xl:size-6 shrink-0" />
              <p className="text-sm">No workout planned for today</p>
            </div>
          ) : (
            <div className="space-y-2">
              {items.map((item) => {
                const status = item.performance?.status;
                const resolved = status === "done" || status === "skipped";
                return (
                  <div key={item.id} className="flex items-center gap-2">
                    {status === "done" ? (
                      <Check className="size-4 @xl:size-6 text-success shrink-0" role="img" aria-label="Completed" />
                    ) : status === "skipped" ? (
                      <SkipForward className="size-4 @xl:size-6 text-muted-foreground shrink-0" role="img" aria-label="Skipped" />
                    ) : status === "partial" ? (
                      <CircleDot className="size-4 @xl:size-6 text-warning-text shrink-0" role="img" aria-label="Partially completed" />
                    ) : (
                      <Circle className="size-4 @xl:size-6 text-muted-foreground shrink-0" role="img" aria-label="Pending" />
                    )}
                    <span className={cn("text-sm @xl:text-xl truncate", resolved && "text-muted-foreground line-through")}>
                      {item.exerciseName}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
