import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ChevronLeft, ChevronRight, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  useWorkoutWeek,
  useCopyWorkoutWeek,
  useAddPlannedItem,
  useDeletePlannedItem,
  useWorkoutExercises,
  usePatchWorkoutWeek,
} from "@/lib/hooks/api/use-workouts";
import { addDays, currentWeekStart } from "@/lib/workout-week";
import { DAYS_OF_WEEK_LABELS, type WeekItem } from "@/types/models/workout";
import { toast } from "sonner";

// Weekly planner. Desktop-first layout with one column per day. Drag-and-drop
// reorder is intentionally deferred — items can be added/removed and the
// rest-day toggle is wired through. The empty-week prompt offers Copy Planned
// / Copy Actual / Start Fresh per PRD §4.5.
export function WorkoutWeekPage() {
  const params = useParams<{ weekStart?: string }>();
  const navigate = useNavigate();
  const weekStart = params.weekStart ?? currentWeekStart();

  const week = useWorkoutWeek(weekStart);
  const copy = useCopyWorkoutWeek();
  const patch = usePatchWorkoutWeek();
  const add = useAddPlannedItem();
  const remove = useDeletePlannedItem();
  const exercises = useWorkoutExercises();

  const goWeek = (offset: number) => navigate(`/app/workouts/week/${addDays(weekStart, offset * 7)}`);

  const isEmpty = week.error || (week.data && week.data.data.attributes.items.length === 0);
  const exerciseList = exercises.data?.data ?? [];

  const restDayFlags = week.data?.data.attributes.restDayFlags ?? [];

  const toggleRestDay = (day: number) => {
    const next = restDayFlags.includes(day)
      ? restDayFlags.filter((d) => d !== day)
      : [...restDayFlags, day];
    patch.mutate(
      { weekStart, restDayFlags: next },
      { onError: () => toast.error("Failed to update rest day") },
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Button variant="outline" size="sm" onClick={() => goWeek(-1)}>
          <ChevronLeft className="h-4 w-4" />
          Prev
        </Button>
        <span className="text-sm font-medium">Week of {weekStart}</span>
        <Button variant="outline" size="sm" onClick={() => goWeek(1)}>
          Next
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>

      {isEmpty ? (
        <EmptyWeek
          weekStart={weekStart}
          onCopy={(mode) =>
            copy.mutate(
              { weekStart, mode },
              {
                onSuccess: () => toast.success(`Copied ${mode} from previous week`),
                onError: (e) => toast.error((e as Error).message ?? "Copy failed"),
              },
            )
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-7 gap-3">
          {DAYS_OF_WEEK_LABELS.map((label, day) => {
            const items =
              week.data?.data.attributes.items.filter((it) => it.dayOfWeek === day) ?? [];
            const isRest = restDayFlags.includes(day);
            return (
              <Card key={day} className="min-h-[8rem]">
                <CardHeader className="p-3">
                  <CardTitle className="flex items-center justify-between text-sm">
                    <span>{label}</span>
                    <button
                      onClick={() => toggleRestDay(day)}
                      className={`text-xs rounded px-1.5 py-0.5 ${
                        isRest ? "bg-secondary" : "text-muted-foreground hover:bg-muted"
                      }`}
                    >
                      {isRest ? "Rest" : "·"}
                    </button>
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-3 pt-0 space-y-2">
                  {items.map((it) => (
                    <PlannedRow
                      key={it.id}
                      item={it}
                      onDelete={() =>
                        remove.mutate(
                          { weekStart, itemId: it.id },
                          { onError: () => toast.error("Delete failed") },
                        )
                      }
                    />
                  ))}
                  <ExercisePicker
                    exerciseOptions={exerciseList}
                    onSelect={(exerciseId) =>
                      add.mutate(
                        { weekStart, attrs: { exerciseId, dayOfWeek: day } },
                        { onError: (e) => toast.error((e as Error).message ?? "Add failed") },
                      )
                    }
                  />
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}

function EmptyWeek({
  weekStart,
  onCopy,
}: {
  weekStart: string;
  onCopy: (mode: "planned" | "actual") => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Nothing planned for {weekStart}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <p className="text-sm text-muted-foreground">
          Copy a previous week or start from scratch by adding exercises below.
        </p>
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => onCopy("planned")}>Copy Planned</Button>
          <Button onClick={() => onCopy("actual")} variant="outline">
            Copy Actual
          </Button>
          <Button variant="ghost">Start Fresh</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function PlannedRow({ item, onDelete }: { item: WeekItem; onDelete: () => void }) {
  return (
    <div className="flex items-start justify-between gap-1 rounded border p-2 text-xs">
      <div>
        <p className="font-medium">{item.exerciseName}</p>
        <p className="text-muted-foreground">{summarize(item)}</p>
      </div>
      <Button size="icon" variant="ghost" onClick={onDelete} aria-label="Delete">
        <Trash2 className="h-3 w-3" />
      </Button>
    </div>
  );
}

function summarize(it: WeekItem): string {
  const p = it.planned;
  switch (it.kind) {
    case "strength":
      return p.sets && p.reps ? `${p.sets}×${p.reps} @ ${p.weight ?? "?"} ${p.weightUnit ?? ""}` : "—";
    case "isometric":
      return p.sets && p.durationSeconds ? `${p.sets}×${p.durationSeconds}s` : "—";
    case "cardio":
      return p.distance ? `${p.distance} ${p.distanceUnit ?? ""}` : "—";
  }
}

// ExercisePicker is the per-day "add exercise" affordance. The actual catalog
// management lives on the Exercises tab; here we just need a select to attach
// an existing exercise to the current day.
function ExercisePicker({
  exerciseOptions,
  onSelect,
}: {
  exerciseOptions: Array<{ id: string; attributes: { name: string } }>;
  onSelect: (id: string) => void;
}) {
  const [value, setValue] = useState<string>("");
  return (
    <Select
      value={value}
      onValueChange={(v) => {
        if (!v) return;
        setValue("");
        onSelect(v);
      }}
    >
      <SelectTrigger className="h-7 text-xs">
        <SelectValue placeholder="+ Add exercise" />
      </SelectTrigger>
      <SelectContent>
        {exerciseOptions.map((e) => (
          <SelectItem key={e.id} value={e.id}>
            {e.attributes.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
