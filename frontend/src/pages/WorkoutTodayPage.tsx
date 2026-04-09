import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useWorkoutToday, usePatchPerformance } from "@/lib/hooks/api/use-workouts";
import type { WeekItem } from "@/types/models/workout";
import { toast } from "sonner";

// Mobile-first today view. The page is intentionally minimal: large status
// buttons (done / skipped) and a planned-vs-actual readout. Per-set logging is
// reachable from the weekly planner; the today view is meant for one-handed
// quick logging during a workout, not for complex set-by-set entry.
export function WorkoutTodayPage() {
  const { data, isLoading, error } = useWorkoutToday();
  const patch = usePatchPerformance();

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>;
  if (error) return <p className="text-destructive">Failed to load today's workout.</p>;
  if (!data) return null;

  const attrs = data.data.attributes;
  const items = attrs.items;
  const weekStart = attrs.weekStartDate;

  const onMark = (item: WeekItem, status: "done" | "skipped" | "pending") => {
    patch.mutate(
      { weekStart, itemId: item.id, attrs: { status } },
      {
        onError: () => toast.error("Failed to update status"),
      },
    );
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>{attrs.date}</span>
            {attrs.isRestDay && <Badge variant="secondary">Rest day</Badge>}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {items.length === 0 ? (
            <p className="text-muted-foreground">Nothing planned for today.</p>
          ) : (
            <ul className="space-y-2">
              {items.map((it) => (
                <li key={it.id} className="rounded-lg border p-3 space-y-2">
                  <div className="flex items-center justify-between gap-2">
                    <div>
                      <p className="font-medium">
                        {it.exerciseName}
                        {it.exerciseDeleted && <span className="text-muted-foreground"> (deleted)</span>}
                      </p>
                      <p className="text-xs text-muted-foreground">{summarizePlanned(it)}</p>
                    </div>
                    <StatusBadge status={it.performance?.status ?? "pending"} />
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button size="sm" onClick={() => onMark(it, "done")}>
                      Done
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => onMark(it, "skipped")}>
                      Skip
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => onMark(it, "pending")}>
                      Reset
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const tone =
    status === "done"
      ? "bg-emerald-500/10 text-emerald-600"
      : status === "skipped"
      ? "bg-rose-500/10 text-rose-600"
      : status === "partial"
      ? "bg-amber-500/10 text-amber-600"
      : "bg-muted text-muted-foreground";
  return <span className={`rounded px-2 py-0.5 text-xs font-medium ${tone}`}>{status}</span>;
}

function summarizePlanned(it: WeekItem): string {
  const p = it.planned;
  switch (it.kind) {
    case "strength":
      if (p.sets && p.reps) return `${p.sets} × ${p.reps} @ ${p.weight ?? "?"} ${p.weightUnit ?? ""}`;
      return "—";
    case "isometric":
      if (p.sets && p.durationSeconds) return `${p.sets} × ${p.durationSeconds}s`;
      return "—";
    case "cardio":
      if (p.distance) return `${p.distance} ${p.distanceUnit ?? ""}`;
      if (p.durationSeconds) return `${Math.round(p.durationSeconds / 60)} min`;
      return "—";
  }
}
