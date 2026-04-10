import { useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ChevronLeft, ChevronRight, GripVertical, Search, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useWorkoutWeek,
  useCopyWorkoutWeek,
  useAddPlannedItem,
  useDeletePlannedItem,
  useReorderPlannedItems,
  useWorkoutExercises,
  useWorkoutThemes,
  useWorkoutRegions,
  usePatchWorkoutWeek,
} from "@/lib/hooks/api/use-workouts";
import { addDays, currentWeekStart } from "@/lib/workout-week";
import { DAYS_OF_WEEK_LABELS, type Exercise, type WeekItem } from "@/types/models/workout";
import { toast } from "sonner";

// Weekly planner. Desktop-first layout with one column per day. Drag-and-drop
// reorder uses native HTML5 DnD so we don't pull in a new dependency. The
// empty-week prompt offers Copy Planned / Copy Actual / Start Fresh per
// PRD §4.5; "Start Fresh" lazily creates the week row by toggling rest days
// off, then leaves the planner open for the user to add items.
export function WorkoutWeekPage() {
  const params = useParams<{ weekStart?: string }>();
  const navigate = useNavigate();
  const weekStart = params.weekStart ?? currentWeekStart();

  const week = useWorkoutWeek(weekStart);
  const copy = useCopyWorkoutWeek();
  const patch = usePatchWorkoutWeek();
  const add = useAddPlannedItem();
  const remove = useDeletePlannedItem();
  const reorder = useReorderPlannedItems();
  const exercises = useWorkoutExercises();

  const goWeek = (offset: number) => navigate(`/app/workouts/week/${addDays(weekStart, offset * 7)}`);

  const items = week.data?.data.attributes.items ?? [];
  // Show the empty-week prompt (Copy / Start Fresh) only when the week row
  // does not exist (404). When the row exists but has zero items (e.g. after
  // Start Fresh), show the planner grid so the user can add exercises.
  const weekNotFound = !!week.error;

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

  // startFresh lazily creates the week row by patching restDayFlags to its
  // current (empty) value. The PATCH endpoint creates the row when missing,
  // so subsequent adds have a parent week to attach to.
  const startFresh = () => {
    patch.mutate(
      { weekStart, restDayFlags: [] },
      {
        onSuccess: () => toast.success("Started a fresh week — add exercises below"),
        onError: () => toast.error("Failed to start fresh"),
      },
    );
  };

  // --- DnD reorder ---------------------------------------------------------
  //
  // We track the dragged item id in component state. The drop handler computes
  // the new (day, position) for the dragged item and the shifted positions for
  // every other item that crossed the move, then sends one reorder POST. The
  // server applies it atomically.
  const [draggedItemId, setDraggedItemId] = useState<string | null>(null);

  const onDragStart = (itemId: string) => () => setDraggedItemId(itemId);
  const onDragEnd = () => setDraggedItemId(null);

  const onDropOnDay = (targetDay: number) => (e: React.DragEvent) => {
    e.preventDefault();
    if (!draggedItemId) return;
    const dragged = items.find((it) => it.id === draggedItemId);
    if (!dragged) return;

    // Build the new ordering: remove the dragged item from its current spot,
    // append it to the target day's tail. We rebuild positions for both the
    // source and target days so positions stay contiguous.
    const sourceDay = dragged.dayOfWeek;
    const sourceItems = items
      .filter((it) => it.dayOfWeek === sourceDay && it.id !== dragged.id)
      .sort((a, b) => a.position - b.position);
    const targetItems = items
      .filter((it) => it.dayOfWeek === targetDay && it.id !== dragged.id)
      .sort((a, b) => a.position - b.position);
    targetItems.push(dragged);

    const reorderInputs: Array<{ itemId: string; dayOfWeek: number; position: number }> = [];
    sourceItems.forEach((it, idx) => reorderInputs.push({ itemId: it.id, dayOfWeek: sourceDay, position: idx }));
    targetItems.forEach((it, idx) => reorderInputs.push({ itemId: it.id, dayOfWeek: targetDay, position: idx }));

    reorder.mutate(
      { weekStart, items: reorderInputs },
      { onError: (err) => toast.error((err as Error).message ?? "Reorder failed") },
    );
    setDraggedItemId(null);
  };

  // Allow drop targets to receive the drop event.
  const allowDrop = (e: React.DragEvent) => e.preventDefault();

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

      {weekNotFound ? (
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
          onStartFresh={startFresh}
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-7 gap-3">
          {DAYS_OF_WEEK_LABELS.map((label, day) => {
            const dayItems = items
              .filter((it) => it.dayOfWeek === day)
              .sort((a, b) => a.position - b.position);
            const isRest = restDayFlags.includes(day);
            return (
              <Card
                key={day}
                className="min-h-[8rem]"
                onDragOver={allowDrop}
                onDrop={onDropOnDay(day)}
              >
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
                  {dayItems.map((it) => (
                    <PlannedRow
                      key={it.id}
                      item={it}
                      draggable
                      onDragStart={onDragStart(it.id)}
                      onDragEnd={onDragEnd}
                      onDelete={() =>
                        remove.mutate(
                          { weekStart, itemId: it.id },
                          { onError: () => toast.error("Delete failed") },
                        )
                      }
                    />
                  ))}
                  <ExercisePickerButton
                    exercises={exerciseList}
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
  onStartFresh,
}: {
  weekStart: string;
  onCopy: (mode: "planned" | "actual") => void;
  onStartFresh: () => void;
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
          <Button variant="ghost" onClick={onStartFresh}>
            Start Fresh
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function PlannedRow({
  item,
  draggable,
  onDragStart,
  onDragEnd,
  onDelete,
}: {
  item: WeekItem;
  draggable?: boolean;
  onDragStart?: () => void;
  onDragEnd?: () => void;
  onDelete: () => void;
}) {
  return (
    <div
      className="flex items-start justify-between gap-1 rounded border p-2 text-xs"
      draggable={draggable}
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
    >
      <GripVertical className="h-3 w-3 mt-0.5 text-muted-foreground cursor-grab" />
      <div className="flex-1">
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

// ExercisePickerButton opens the per-day add modal. The modal supports
// theme/region (incl. secondary) filter + free-text search per G8.
function ExercisePickerButton({
  exercises,
  onSelect,
}: {
  exercises: Exercise[];
  onSelect: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger>
        <Button variant="outline" size="sm" className="h-7 w-full text-xs">
          + Add exercise
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add exercise</DialogTitle>
        </DialogHeader>
        <ExercisePickerModal
          exercises={exercises}
          onSelect={(id) => {
            onSelect(id);
            setOpen(false);
          }}
        />
      </DialogContent>
    </Dialog>
  );
}

// ExercisePickerModal is the filterable picker. The region filter matches
// either the primary or any secondary region — same semantics as the backend
// list endpoint.
function ExercisePickerModal({
  exercises,
  onSelect,
}: {
  exercises: Exercise[];
  onSelect: (id: string) => void;
}) {
  const themes = useWorkoutThemes();
  const regions = useWorkoutRegions();
  const [themeId, setThemeId] = useState<string>("all");
  const [regionId, setRegionId] = useState<string>("all");
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return exercises.filter((e) => {
      if (themeId !== "all" && e.attributes.themeId !== themeId) return false;
      if (regionId !== "all") {
        const matchesPrimary = e.attributes.regionId === regionId;
        const matchesSecondary = e.attributes.secondaryRegionIds.includes(regionId);
        if (!matchesPrimary && !matchesSecondary) return false;
      }
      if (q && !e.attributes.name.toLowerCase().includes(q)) return false;
      return true;
    });
  }, [exercises, themeId, regionId, query]);

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-2">
        <Select value={themeId} onValueChange={(v) => setThemeId(v ?? "all")}>
          <SelectTrigger className="h-8 text-xs">
            <SelectValue placeholder="Theme" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All themes</SelectItem>
            {(themes.data?.data ?? []).map((t) => (
              <SelectItem key={t.id} value={t.id}>
                {t.attributes.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={regionId} onValueChange={(v) => setRegionId(v ?? "all")}>
          <SelectTrigger className="h-8 text-xs">
            <SelectValue placeholder="Region" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All regions</SelectItem>
            {(regions.data?.data ?? []).map((r) => (
              <SelectItem key={r.id} value={r.id}>
                {r.attributes.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="relative">
        <Search className="absolute left-2 top-2 h-3 w-3 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search exercises…"
          className="h-8 pl-7 text-xs"
        />
      </div>
      <div className="max-h-72 overflow-y-auto space-y-1 border rounded p-1">
        {filtered.length === 0 ? (
          <p className="text-xs text-muted-foreground p-2">No exercises match the current filters.</p>
        ) : (
          filtered.map((e) => (
            <button
              key={e.id}
              onClick={() => onSelect(e.id)}
              className="w-full text-left px-2 py-1.5 text-xs rounded hover:bg-muted"
            >
              <span className="font-medium">{e.attributes.name}</span>
              <span className="text-muted-foreground ml-1">({e.attributes.kind})</span>
            </button>
          ))
        )}
      </div>
    </div>
  );
}
