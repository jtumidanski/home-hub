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
  useUpdatePlannedItem,
} from "@/lib/hooks/api/use-workouts";
import { addDays, currentWeekStart } from "@/lib/workout-week";
import { DAYS_OF_WEEK_LABELS, type Exercise, type WeekItem } from "@/types/models/workout";
import { toast } from "sonner";

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
  const updateItem = useUpdatePlannedItem();

  const goWeek = (offset: number) => navigate(`/app/workouts/week/${addDays(weekStart, offset * 7)}`);

  const items = week.data?.data.attributes.items ?? [];
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

  const startFresh = () => {
    patch.mutate(
      { weekStart, restDayFlags: [] },
      {
        onSuccess: () => toast.success("Started a fresh week — add exercises below"),
        onError: () => toast.error("Failed to start fresh"),
      },
    );
  };

  // --- DnD reorder with positional drops -----------------------------------

  const [draggedItemId, setDraggedItemId] = useState<string | null>(null);
  const [dropIndicator, setDropIndicator] = useState<{ day: number; position: number } | null>(null);

  const onDragStart = (itemId: string) => (e: React.DragEvent) => {
    setDraggedItemId(itemId);
    e.dataTransfer.effectAllowed = "move";
  };

  const onDragEnd = () => {
    setDraggedItemId(null);
    setDropIndicator(null);
  };

  // Per-item drag over: detect top/bottom half to set insertion position.
  const onDragOverItem = (day: number, idx: number) => (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    const pos = e.clientY < midY ? idx : idx + 1;
    setDropIndicator((prev) =>
      prev?.day === day && prev?.position === pos ? prev : { day, position: pos },
    );
  };

  // Card-level drag over: fires when hovering empty area below items.
  const onDragOverDay = (day: number, itemCount: number) => (e: React.DragEvent) => {
    e.preventDefault();
    setDropIndicator((prev) =>
      prev?.day === day && prev?.position === itemCount ? prev : { day, position: itemCount },
    );
  };

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (!draggedItemId || !dropIndicator) return;
    const dragged = items.find((it) => it.id === draggedItemId);
    if (!dragged) return;

    const sourceDay = dragged.dayOfWeek;
    const targetDay = dropIndicator.day;

    const sourceItems = items
      .filter((it) => it.dayOfWeek === sourceDay && it.id !== draggedItemId)
      .sort((a, b) => a.position - b.position);
    const targetDayItems = items
      .filter((it) => it.dayOfWeek === targetDay && it.id !== draggedItemId)
      .sort((a, b) => a.position - b.position);

    // Adjust insertion index when dragging within the same day.
    let insertIdx = dropIndicator.position;
    if (sourceDay === targetDay) {
      const origIdx = items
        .filter((it) => it.dayOfWeek === sourceDay)
        .sort((a, b) => a.position - b.position)
        .findIndex((it) => it.id === draggedItemId);
      if (origIdx >= 0 && origIdx < dropIndicator.position) insertIdx--;
    }
    targetDayItems.splice(insertIdx, 0, dragged);

    const reorderInputs: Array<{ itemId: string; dayOfWeek: number; position: number }> = [];
    if (sourceDay !== targetDay) {
      sourceItems.forEach((it, idx) =>
        reorderInputs.push({ itemId: it.id, dayOfWeek: sourceDay, position: idx }),
      );
    }
    targetDayItems.forEach((it, idx) =>
      reorderInputs.push({ itemId: it.id, dayOfWeek: targetDay, position: idx }),
    );

    reorder.mutate(
      { weekStart, items: reorderInputs },
      { onError: (err) => toast.error((err as Error).message ?? "Reorder failed") },
    );
    setDraggedItemId(null);
    setDropIndicator(null);
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
                onDragOver={onDragOverDay(day, dayItems.length)}
                onDrop={onDrop}
              >
                <CardHeader className="p-3">
                  <CardTitle className="flex items-center justify-between text-sm">
                    <span>{label}</span>
                    <button
                      onClick={() => toggleRestDay(day)}
                      className={`text-xs rounded-full px-2 py-0.5 border transition-colors ${
                        isRest
                          ? "bg-secondary text-secondary-foreground border-secondary"
                          : "text-muted-foreground border-transparent hover:border-muted-foreground/30"
                      }`}
                    >
                      Rest
                    </button>
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-3 pt-0 space-y-1">
                  {dayItems.map((it, idx) => (
                    <div key={it.id}>
                      {dropIndicator?.day === day && dropIndicator.position === idx && (
                        <div className="h-0.5 bg-primary rounded my-1" />
                      )}
                      <PlannedRow
                        item={it}
                        draggable
                        onDragStart={onDragStart(it.id)}
                        onDragEnd={onDragEnd}
                        onDragOver={onDragOverItem(day, idx)}
                        onDelete={() =>
                          remove.mutate(
                            { weekStart, itemId: it.id },
                            { onError: () => toast.error("Delete failed") },
                          )
                        }
                        onUpdatePlanned={(planned) =>
                          updateItem.mutate(
                            { weekStart, itemId: it.id, attrs: { planned } },
                            { onError: () => toast.error("Update failed") },
                          )
                        }
                      />
                    </div>
                  ))}
                  {dropIndicator?.day === day && dropIndicator.position === dayItems.length && (
                    <div className="h-0.5 bg-primary rounded my-1" />
                  )}
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

// --- Planned item row with inline editing ----------------------------------

function PlannedRow({
  item,
  draggable,
  onDragStart,
  onDragEnd,
  onDragOver,
  onDelete,
  onUpdatePlanned,
}: {
  item: WeekItem;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragEnd?: () => void;
  onDragOver?: (e: React.DragEvent) => void;
  onDelete: () => void;
  onUpdatePlanned: (planned: Record<string, unknown>) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [sets, setSets] = useState(item.planned.sets?.toString() ?? "");
  const [reps, setReps] = useState(item.planned.reps?.toString() ?? "");
  const [weight, setWeight] = useState(item.planned.weight?.toString() ?? "");
  const [weightUnit, setWeightUnit] = useState(item.planned.weightUnit ?? "lb");
  const [durationSeconds, setDurationSeconds] = useState(item.planned.durationSeconds?.toString() ?? "");
  const [distance, setDistance] = useState(item.planned.distance?.toString() ?? "");
  const [distanceUnit, setDistanceUnit] = useState(item.planned.distanceUnit ?? "mi");

  const save = () => {
    const planned: Record<string, unknown> = {};
    switch (item.kind) {
      case "strength":
        if (sets) planned.sets = parseInt(sets);
        if (reps) planned.reps = parseInt(reps);
        if (weight) planned.weight = parseFloat(weight);
        planned.weightUnit = weightUnit;
        break;
      case "isometric":
        if (sets) planned.sets = parseInt(sets);
        if (durationSeconds) planned.durationSeconds = parseInt(durationSeconds);
        if (weight) planned.weight = parseFloat(weight);
        if (weight) planned.weightUnit = weightUnit;
        break;
      case "cardio":
        if (durationSeconds) planned.durationSeconds = parseInt(durationSeconds);
        if (distance) planned.distance = parseFloat(distance);
        planned.distanceUnit = distanceUnit;
        break;
    }
    onUpdatePlanned(planned);
    setEditing(false);
  };

  return (
    <div
      className="rounded border p-2 text-xs"
      draggable={draggable}
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
      onDragOver={onDragOver}
    >
      <div className="flex items-start gap-1">
        <GripVertical className="h-3 w-3 mt-0.5 shrink-0 text-muted-foreground cursor-grab" />
        <div className="flex-1 min-w-0 cursor-pointer" onClick={() => setEditing(!editing)}>
          <p className="font-medium truncate">{item.exerciseName}</p>
          <p className="text-muted-foreground truncate">{summarize(item)}</p>
        </div>
        <Button
          size="icon"
          variant="ghost"
          onClick={onDelete}
          aria-label="Delete"
          className="shrink-0 h-6 w-6"
        >
          <Trash2 className="h-3 w-3" />
        </Button>
      </div>
      {editing && (
        <div className="mt-2 space-y-1.5 border-t pt-2">
          {item.kind === "strength" && (
            <div className="grid grid-cols-2 gap-1">
              <Input className="h-6 text-xs" placeholder="Sets" value={sets} onChange={(e) => setSets(e.target.value)} type="number" />
              <Input className="h-6 text-xs" placeholder="Reps" value={reps} onChange={(e) => setReps(e.target.value)} type="number" />
              <Input className="h-6 text-xs" placeholder="Weight" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" />
              <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
                <SelectTrigger className="h-6 text-xs"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="lb">lb</SelectItem>
                  <SelectItem value="kg">kg</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}
          {item.kind === "isometric" && (
            <div className="grid grid-cols-2 gap-1">
              <Input className="h-6 text-xs" placeholder="Sets" value={sets} onChange={(e) => setSets(e.target.value)} type="number" />
              <Input className="h-6 text-xs" placeholder="Duration (s)" value={durationSeconds} onChange={(e) => setDurationSeconds(e.target.value)} type="number" />
              <Input className="h-6 text-xs" placeholder="Weight" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" />
              <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
                <SelectTrigger className="h-6 text-xs"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="lb">lb</SelectItem>
                  <SelectItem value="kg">kg</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}
          {item.kind === "cardio" && (
            <div className="grid grid-cols-2 gap-1">
              <Input className="h-6 text-xs" placeholder="Duration (s)" value={durationSeconds} onChange={(e) => setDurationSeconds(e.target.value)} type="number" />
              <Input className="h-6 text-xs" placeholder="Distance" value={distance} onChange={(e) => setDistance(e.target.value)} type="number" />
              <Select value={distanceUnit} onValueChange={(v) => v && setDistanceUnit(v)}>
                <SelectTrigger className="h-6 text-xs"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="mi">mi</SelectItem>
                  <SelectItem value="km">km</SelectItem>
                  <SelectItem value="m">m</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}
          <div className="flex gap-1">
            <Button size="sm" className="h-6 text-xs" onClick={save}>Save</Button>
            <Button size="sm" variant="ghost" className="h-6 text-xs" onClick={() => setEditing(false)}>Cancel</Button>
          </div>
        </div>
      )}
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

// --- Exercise picker (unchanged) -------------------------------------------

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
