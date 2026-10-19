import { useEffect, useMemo, useRef, useState } from "react";
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

  // Shared by the drop handler and the single-pointer "Move to day" control.
  const moveItem = (itemId: string, targetDay: number, position: number) => {
    const dragged = items.find((it) => it.id === itemId);
    if (!dragged) return;

    const sourceDay = dragged.dayOfWeek;

    const sourceItems = items
      .filter((it) => it.dayOfWeek === sourceDay && it.id !== itemId)
      .sort((a, b) => a.position - b.position);
    const targetDayItems = items
      .filter((it) => it.dayOfWeek === targetDay && it.id !== itemId)
      .sort((a, b) => a.position - b.position);

    // Adjust insertion index when dragging within the same day.
    let insertIdx = position;
    if (sourceDay === targetDay) {
      const origIdx = items
        .filter((it) => it.dayOfWeek === sourceDay)
        .sort((a, b) => a.position - b.position)
        .findIndex((it) => it.id === itemId);
      if (origIdx >= 0 && origIdx < position) insertIdx--;
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
  };

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (!draggedItemId || !dropIndicator) return;
    moveItem(draggedItemId, dropIndicator.day, dropIndicator.position);
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
                    <Button
                      size="xs"
                      variant={isRest ? "secondary" : "ghost"}
                      aria-pressed={isRest}
                      aria-label={`Rest day, ${label}`}
                      onClick={() => toggleRestDay(day)}
                      className={isRest ? "" : "text-muted-foreground"}
                    >
                      Rest
                    </Button>
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
                        currentDay={day}
                        onMoveToDay={(targetDay) =>
                          moveItem(it.id, targetDay, items.filter((x) => x.dayOfWeek === targetDay).length)
                        }
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
                    onSelect={(exerciseId, planned) =>
                      add.mutate(
                        {
                          weekStart,
                          attrs: {
                            exerciseId,
                            dayOfWeek: day,
                            ...(planned && Object.keys(planned).length > 0 ? { planned } : {}),
                          },
                        },
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
  currentDay,
  onMoveToDay,
  onDelete,
  onUpdatePlanned,
}: {
  item: WeekItem;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragEnd?: () => void;
  onDragOver?: (e: React.DragEvent) => void;
  currentDay: number;
  onMoveToDay: (day: number) => void;
  onDelete: () => void;
  onUpdatePlanned: (planned: Record<string, unknown>) => void;
}) {
  const [editing, setEditing] = useState(false);
  const firstInputRef = useRef<HTMLInputElement>(null);
  const [sets, setSets] = useState(item.planned.sets?.toString() ?? "");
  const [reps, setReps] = useState(item.planned.reps?.toString() ?? "");
  const [weight, setWeight] = useState(item.planned.weight?.toString() ?? "");
  const [weightUnit, setWeightUnit] = useState(item.planned.weightUnit ?? "lb");
  const [durMin, setDurMin] = useState(item.planned.durationSeconds ? Math.floor(item.planned.durationSeconds / 60).toString() : "");
  const [durSec, setDurSec] = useState(item.planned.durationSeconds ? (item.planned.durationSeconds % 60).toString() : "");
  const [distance, setDistance] = useState(item.planned.distance?.toString() ?? "");
  const [distanceUnit, setDistanceUnit] = useState(item.planned.distanceUnit ?? "mi");

  useEffect(() => {
    if (editing) firstInputRef.current?.focus();
  }, [editing]);

  const isBw = item.weightType === "bodyweight";

  const save = () => {
    const planned: Record<string, unknown> = {};
    const totalSec = (parseInt(durMin) || 0) * 60 + (parseInt(durSec) || 0);
    switch (item.kind) {
      case "strength":
        if (sets) planned.sets = parseInt(sets);
        if (reps) planned.reps = parseInt(reps);
        if (!isBw && weight) planned.weight = parseFloat(weight);
        if (!isBw) planned.weightUnit = weightUnit;
        break;
      case "isometric":
        if (sets) planned.sets = parseInt(sets);
        if (totalSec) planned.durationSeconds = totalSec;
        if (weight) planned.weight = parseFloat(weight);
        if (weight) planned.weightUnit = weightUnit;
        break;
      case "cardio":
        if (totalSec) planned.durationSeconds = totalSec;
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
        <button
          type="button"
          className="flex-1 min-w-0 cursor-pointer rounded text-left focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          aria-expanded={editing}
          onClick={() => setEditing(!editing)}
        >
          <span className="block font-medium truncate">{item.exerciseName}</span>
          <span className="block text-muted-foreground truncate tabular-nums">{summarize(item)}</span>
        </button>
        <Button
          size="icon-xs"
          variant="ghost"
          onClick={onDelete}
          aria-label={`Delete ${item.exerciseName}`}
          className="shrink-0"
        >
          <Trash2 className="h-3 w-3" />
        </Button>
      </div>
      <Select value="" onValueChange={(v) => v && onMoveToDay(Number(v))}>
        <SelectTrigger
          size="sm"
          aria-label={`Move ${item.exerciseName} to day`}
          className="mt-1 w-full text-xs coarse:min-h-11"
        >
          <SelectValue placeholder="Move to day…" />
        </SelectTrigger>
        <SelectContent>
          {DAYS_OF_WEEK_LABELS.map((label, d) =>
            d === currentDay ? null : (
              <SelectItem key={label} value={String(d)}>{label}</SelectItem>
            ),
          )}
        </SelectContent>
      </Select>
      {editing && (
        <div className="mt-2 space-y-1.5 border-t pt-2">
          {item.kind === "strength" && (
            <div className={`grid gap-1 ${isBw ? "grid-cols-2" : "grid-cols-2"}`}>
              <Input ref={firstInputRef} aria-label="Sets" className="tabular-nums" placeholder="Sets" value={sets} onChange={(e) => setSets(e.target.value)} type="number" />
              <Input aria-label="Reps" className="tabular-nums" placeholder="Reps" value={reps} onChange={(e) => setReps(e.target.value)} type="number" />
              {!isBw && (
                <>
                  <Input aria-label="Weight" className="tabular-nums" placeholder="Weight" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" />
                  <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
                    <SelectTrigger aria-label="Weight unit" className="text-xs"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="lb">lb</SelectItem>
                      <SelectItem value="kg">kg</SelectItem>
                    </SelectContent>
                  </Select>
                </>
              )}
            </div>
          )}
          {item.kind === "isometric" && (
            <div className="grid grid-cols-2 gap-1">
              <Input ref={firstInputRef} aria-label="Sets" className="tabular-nums" placeholder="Sets" value={sets} onChange={(e) => setSets(e.target.value)} type="number" />
              <div className="flex gap-0.5">
                <Input aria-label="Duration minutes" className="tabular-nums" placeholder="Min" value={durMin} onChange={(e) => setDurMin(e.target.value)} type="number" />
                <Input aria-label="Duration seconds" className="tabular-nums" placeholder="Sec" value={durSec} onChange={(e) => setDurSec(e.target.value)} type="number" />
              </div>
              <Input aria-label="Weight" className="tabular-nums" placeholder="Weight" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" />
              <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
                <SelectTrigger aria-label="Weight unit" className="text-xs"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="lb">lb</SelectItem>
                  <SelectItem value="kg">kg</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}
          {item.kind === "cardio" && (
            <div className="space-y-1">
              <div className="flex gap-0.5">
                <Input ref={firstInputRef} aria-label="Duration minutes" className="tabular-nums" placeholder="Min" value={durMin} onChange={(e) => setDurMin(e.target.value)} type="number" />
                <Input aria-label="Duration seconds" className="tabular-nums" placeholder="Sec" value={durSec} onChange={(e) => setDurSec(e.target.value)} type="number" />
              </div>
              <div className="flex gap-0.5">
                <Input aria-label="Distance" className="tabular-nums flex-1" placeholder="Distance" value={distance} onChange={(e) => setDistance(e.target.value)} type="number" />
                <Select value={distanceUnit} onValueChange={(v) => v && setDistanceUnit(v)}>
                  <SelectTrigger aria-label="Distance unit" className="w-16 text-xs"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="mi">mi</SelectItem>
                    <SelectItem value="km">km</SelectItem>
                    <SelectItem value="m">m</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}
          <div className="flex gap-1">
            <Button size="sm" onClick={save}>Save</Button>
            <Button size="sm" variant="ghost" onClick={() => setEditing(false)}>Cancel</Button>
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
  onSelect: (id: string, planned?: Record<string, unknown>) => void;
}) {
  const [open, setOpen] = useState(false);
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger>
        <Button variant="outline" size="sm" className="w-full">
          + Add Exercise
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Exercise</DialogTitle>
        </DialogHeader>
        <ExercisePickerModal
          exercises={exercises}
          onSelect={(id, planned) => {
            onSelect(id, planned);
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
  onSelect: (id: string, planned?: Record<string, unknown>) => void;
}) {
  const themes = useWorkoutThemes();
  const regions = useWorkoutRegions();
  const [themeId, setThemeId] = useState<string>("all");
  const [regionId, setRegionId] = useState<string>("all");
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<Exercise | null>(null);

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

  const themeLabel =
    themeId === "all"
      ? "All themes"
      : themes.data?.data?.find((t) => t.id === themeId)?.attributes.name ?? "All themes";
  const regionLabel =
    regionId === "all"
      ? "All regions"
      : regions.data?.data?.find((r) => r.id === regionId)?.attributes.name ?? "All regions";

  if (selected) {
    return (
      <ExerciseInitialValuesForm
        exercise={selected}
        onCancel={() => setSelected(null)}
        onConfirm={(planned) => onSelect(selected.id, planned)}
      />
    );
  }

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-2">
        <Select value={themeId} onValueChange={(v) => setThemeId(v ?? "all")}>
          <SelectTrigger aria-label="Theme" className="text-xs">
            <SelectValue placeholder="Theme">{themeLabel}</SelectValue>
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
          <SelectTrigger aria-label="Region" className="text-xs">
            <SelectValue placeholder="Region">{regionLabel}</SelectValue>
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
          aria-label="Search exercises"
          className="pl-7"
        />
      </div>
      <div className="max-h-72 overflow-y-auto space-y-1 border rounded p-1">
        {filtered.length === 0 ? (
          <p className="text-xs text-muted-foreground p-2">No exercises match the current filters.</p>
        ) : (
          filtered.map((e) => (
            <button
              key={e.id}
              onClick={() => setSelected(e)}
              className="w-full text-left px-2 py-2 text-sm rounded hover:bg-muted coarse:min-h-11 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              <span className="font-medium">{e.attributes.name}</span>
              <span className="text-muted-foreground ml-1">({e.attributes.kind.charAt(0).toUpperCase() + e.attributes.kind.slice(1)})</span>
            </button>
          ))
        )}
      </div>
    </div>
  );
}

function ExerciseInitialValuesForm({
  exercise,
  onCancel,
  onConfirm,
}: {
  exercise: Exercise;
  onCancel: () => void;
  onConfirm: (planned: Record<string, unknown>) => void;
}) {
  const kind = exercise.attributes.kind;
  const isBw = exercise.attributes.weightType === "bodyweight";
  const defaults = exercise.attributes.defaults ?? {};

  const [sets, setSets] = useState(defaults.sets?.toString() ?? "");
  const [reps, setReps] = useState(defaults.reps?.toString() ?? "");
  const [weight, setWeight] = useState(defaults.weight?.toString() ?? "");
  const [weightUnit, setWeightUnit] = useState<string>(defaults.weightUnit ?? "lb");
  const [durMin, setDurMin] = useState(
    defaults.durationSeconds ? Math.floor(defaults.durationSeconds / 60).toString() : "",
  );
  const [durSec, setDurSec] = useState(
    defaults.durationSeconds ? (defaults.durationSeconds % 60).toString() : "",
  );
  const [distance, setDistance] = useState(defaults.distance?.toString() ?? "");
  const [distanceUnit, setDistanceUnit] = useState<string>(defaults.distanceUnit ?? "mi");

  const submit = () => {
    const planned: Record<string, unknown> = {};
    const totalSec = (parseInt(durMin) || 0) * 60 + (parseInt(durSec) || 0);
    switch (kind) {
      case "strength":
        if (sets) planned.sets = parseInt(sets);
        if (reps) planned.reps = parseInt(reps);
        if (!isBw && weight) planned.weight = parseFloat(weight);
        if (!isBw) planned.weightUnit = weightUnit;
        break;
      case "isometric":
        if (sets) planned.sets = parseInt(sets);
        if (totalSec) planned.durationSeconds = totalSec;
        if (weight) planned.weight = parseFloat(weight);
        if (weight) planned.weightUnit = weightUnit;
        break;
      case "cardio":
        if (totalSec) planned.durationSeconds = totalSec;
        if (distance) planned.distance = parseFloat(distance);
        planned.distanceUnit = distanceUnit;
        break;
    }
    onConfirm(planned);
  };

  return (
    <div className="space-y-3">
      <div>
        <p className="font-medium text-sm">{exercise.attributes.name}</p>
        <p className="text-xs text-muted-foreground">
          {kind.charAt(0).toUpperCase() + kind.slice(1)}
          {kind === "strength" && ` · ${isBw ? "Bodyweight" : "Free"}`}
        </p>
      </div>
      {kind === "strength" && (
        <div className="grid grid-cols-2 gap-2">
          <Input aria-label="Sets" className="tabular-nums" placeholder="Sets" value={sets} onChange={(e) => setSets(e.target.value)} type="number" />
          <Input aria-label="Reps" className="tabular-nums" placeholder="Reps" value={reps} onChange={(e) => setReps(e.target.value)} type="number" />
          {!isBw && (
            <>
              <Input aria-label="Weight" className="tabular-nums" placeholder="Weight" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" />
              <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
                <SelectTrigger aria-label="Weight unit"><SelectValue>{weightUnit}</SelectValue></SelectTrigger>
                <SelectContent>
                  <SelectItem value="lb">lb</SelectItem>
                  <SelectItem value="kg">kg</SelectItem>
                </SelectContent>
              </Select>
            </>
          )}
        </div>
      )}
      {kind === "isometric" && (
        <div className="grid grid-cols-2 gap-2">
          <Input aria-label="Sets" className="tabular-nums" placeholder="Sets" value={sets} onChange={(e) => setSets(e.target.value)} type="number" />
          <div className="flex gap-1">
            <Input aria-label="Duration minutes" className="tabular-nums" placeholder="Min" value={durMin} onChange={(e) => setDurMin(e.target.value)} type="number" />
            <Input aria-label="Duration seconds" className="tabular-nums" placeholder="Sec" value={durSec} onChange={(e) => setDurSec(e.target.value)} type="number" />
          </div>
          <Input aria-label="Weight" className="tabular-nums" placeholder="Weight" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" />
          <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
            <SelectTrigger aria-label="Weight unit"><SelectValue>{weightUnit}</SelectValue></SelectTrigger>
            <SelectContent>
              <SelectItem value="lb">lb</SelectItem>
              <SelectItem value="kg">kg</SelectItem>
            </SelectContent>
          </Select>
        </div>
      )}
      {kind === "cardio" && (
        <div className="space-y-2">
          <div className="flex gap-1">
            <Input aria-label="Duration minutes" className="tabular-nums" placeholder="Min" value={durMin} onChange={(e) => setDurMin(e.target.value)} type="number" />
            <Input aria-label="Duration seconds" className="tabular-nums" placeholder="Sec" value={durSec} onChange={(e) => setDurSec(e.target.value)} type="number" />
          </div>
          <div className="flex gap-1">
            <Input aria-label="Distance" className="flex-1 tabular-nums" placeholder="Distance" value={distance} onChange={(e) => setDistance(e.target.value)} type="number" />
            <Select value={distanceUnit} onValueChange={(v) => v && setDistanceUnit(v)}>
              <SelectTrigger aria-label="Distance unit" className="w-20"><SelectValue>{distanceUnit}</SelectValue></SelectTrigger>
              <SelectContent>
                <SelectItem value="mi">mi</SelectItem>
                <SelectItem value="km">km</SelectItem>
                <SelectItem value="m">m</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      )}
      <div className="flex gap-2 justify-end">
        <Button variant="ghost" size="sm" onClick={onCancel}>Back</Button>
        <Button size="sm" onClick={submit}>Add</Button>
      </div>
    </div>
  );
}
