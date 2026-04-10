import { Link } from "react-router-dom";
import { useWorkoutToday } from "@/lib/hooks/api/use-workouts";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Dumbbell, BedDouble, Check, Circle, ChevronRight } from "lucide-react";

export function WorkoutWidget() {
  const { data, isLoading, isError } = useWorkoutToday();

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
    return (
      <Card className="border-destructive">
        <CardContent className="py-3">
          <p className="text-sm text-destructive">Failed to load workout</p>
        </CardContent>
      </Card>
    );
  }

  const attrs = data?.data?.attributes;
  const isRestDay = attrs?.isRestDay ?? false;
  const items = attrs?.items ?? [];

  return (
    <Link to="/app/workouts" className="transition-opacity hover:opacity-80">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Workout</CardTitle>
          <CardAction>
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          </CardAction>
        </CardHeader>
        <CardContent>
          {isRestDay ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <BedDouble className="h-5 w-5" />
              <p className="text-sm">Rest Day</p>
            </div>
          ) : items.length === 0 ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Dumbbell className="h-5 w-5" />
              <p className="text-sm">No workout planned for today</p>
            </div>
          ) : (
            <div className="space-y-2">
              {items.map((item) => {
                const done = item.performance?.status === "done";
                return (
                  <div key={item.id} className="flex items-center gap-2">
                    {done ? (
                      <Check className="h-4 w-4 text-green-600 shrink-0" />
                    ) : (
                      <Circle className="h-4 w-4 text-muted-foreground shrink-0" />
                    )}
                    <span className={`text-sm truncate ${done ? "text-muted-foreground line-through" : ""}`}>
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
