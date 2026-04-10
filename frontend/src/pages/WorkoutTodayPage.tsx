import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useWorkoutToday, usePatchPerformance } from "@/lib/hooks/api/use-workouts";
import type { WeekItem, WeightUnit } from "@/types/models/workout";
import { toast } from "sonner";

export function WorkoutTodayPage() {
  const { data, isLoading, error } = useWorkoutToday();

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>;
  if (error) return <p className="text-destructive">Failed to load today's workout.</p>;
  if (!data) return null;

  const attrs = data.data.attributes;
  const items = attrs.items;
  const weekStart = attrs.weekStartDate;

  return (
    <div className="space-y-4 max-w-lg mx-auto">
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
                <TodayItem key={it.id} item={it} weekStart={weekStart} />
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function TodayItem({ item, weekStart }: { item: WeekItem; weekStart: string }) {
  const [expanded, setExpanded] = useState(false);
  const patch = usePatchPerformance();

  const perf = item.performance;
  const [sets, setSets] = useState(perf?.actuals?.sets?.toString() ?? item.planned.sets?.toString() ?? "");
  const [reps, setReps] = useState(perf?.actuals?.reps?.toString() ?? item.planned.reps?.toString() ?? "");
  const [weight, setWeight] = useState(perf?.actuals?.weight?.toString() ?? item.planned.weight?.toString() ?? "");
  const [weightUnit, setWeightUnit] = useState<string>(perf?.weightUnit ?? item.planned.weightUnit ?? "lb");
  const [durationSeconds, setDurationSeconds] = useState(
    perf?.actuals?.durationSeconds?.toString() ?? item.planned.durationSeconds?.toString() ?? "",
  );
  const [distance, setDistance] = useState(
    perf?.actuals?.distance?.toString() ?? item.planned.distance?.toString() ?? "",
  );
  const [distanceUnit, setDistanceUnit] = useState<string>(
    perf?.actuals?.distanceUnit ?? item.planned.distanceUnit ?? "mi",
  );

  const onMark = (status: "done" | "skipped" | "pending") => {
    patch.mutate(
      { weekStart, itemId: item.id, attrs: { status } },
      { onError: () => toast.error("Failed to update status") },
    );
  };

  const logActuals = () => {
    const actuals: Record<string, unknown> = {};
    switch (item.kind) {
      case "strength":
        if (sets) actuals.sets = parseInt(sets);
        if (reps) actuals.reps = parseInt(reps);
        if (weight) actuals.weight = parseFloat(weight);
        break;
      case "isometric":
        if (sets) actuals.sets = parseInt(sets);
        if (durationSeconds) actuals.durationSeconds = parseInt(durationSeconds);
        if (weight) actuals.weight = parseFloat(weight);
        break;
      case "cardio":
        if (durationSeconds) actuals.durationSeconds = parseInt(durationSeconds);
        if (distance) actuals.distance = parseFloat(distance);
        actuals.distanceUnit = distanceUnit;
        break;
    }
    patch.mutate(
      {
        weekStart,
        itemId: item.id,
        attrs: {
          status: "done",
          ...(item.kind !== "cardio" ? { weightUnit: weightUnit as WeightUnit } : {}),
          actuals,
        },
      },
      {
        onSuccess: () => {
          toast.success("Logged");
          setExpanded(false);
        },
        onError: () => toast.error("Failed to log"),
      },
    );
  };

  const status = perf?.status ?? "pending";

  return (
    <li className="rounded-lg border p-3 space-y-2">
      <div
        className="flex items-center justify-between gap-2 cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="min-w-0 flex-1">
          <p className="font-medium truncate">
            {item.exerciseName}
            {item.exerciseDeleted && <span className="text-muted-foreground"> (deleted)</span>}
          </p>
          <p className="text-xs text-muted-foreground">{summarizePlanned(item)}</p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <StatusBadge status={status} />
          {expanded ? <ChevronUp className="h-4 w-4 text-muted-foreground" /> : <ChevronDown className="h-4 w-4 text-muted-foreground" />}
        </div>
      </div>

      {expanded ? (
        <div className="space-y-3 border-t pt-3">
          {item.kind === "strength" && (
            <div className="grid grid-cols-4 gap-1.5">
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Sets</label>
                <Input className="h-8" value={sets} onChange={(e) => setSets(e.target.value)} type="number" inputMode="numeric" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Reps</label>
                <Input className="h-8" value={reps} onChange={(e) => setReps(e.target.value)} type="number" inputMode="numeric" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Weight</label>
                <Input className="h-8" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" inputMode="decimal" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Unit</label>
                <Select value={weightUnit} onValueChange={(v) => v && setWeightUnit(v)}>
                  <SelectTrigger className="h-8"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="lb">lb</SelectItem>
                    <SelectItem value="kg">kg</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}
          {item.kind === "isometric" && (
            <div className="grid grid-cols-3 gap-1.5">
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Sets</label>
                <Input className="h-8" value={sets} onChange={(e) => setSets(e.target.value)} type="number" inputMode="numeric" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Duration (s)</label>
                <Input className="h-8" value={durationSeconds} onChange={(e) => setDurationSeconds(e.target.value)} type="number" inputMode="numeric" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Weight</label>
                <Input className="h-8" value={weight} onChange={(e) => setWeight(e.target.value)} type="number" inputMode="decimal" />
              </div>
            </div>
          )}
          {item.kind === "cardio" && (
            <div className="grid grid-cols-3 gap-1.5">
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Duration (s)</label>
                <Input className="h-8" value={durationSeconds} onChange={(e) => setDurationSeconds(e.target.value)} type="number" inputMode="numeric" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Distance</label>
                <Input className="h-8" value={distance} onChange={(e) => setDistance(e.target.value)} type="number" inputMode="decimal" />
              </div>
              <div>
                <label className="text-[10px] text-muted-foreground block mb-0.5">Unit</label>
                <Select value={distanceUnit} onValueChange={(v) => v && setDistanceUnit(v)}>
                  <SelectTrigger className="h-8"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="mi">mi</SelectItem>
                    <SelectItem value="km">km</SelectItem>
                    <SelectItem value="m">m</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}
          <div className="flex flex-wrap gap-2">
            <Button size="sm" onClick={logActuals}>Log & Done</Button>
            <Button size="sm" variant="outline" onClick={() => onMark("skipped")}>Skip</Button>
            <Button size="sm" variant="ghost" onClick={() => onMark("pending")}>Reset</Button>
          </div>
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          <Button size="sm" onClick={() => onMark("done")}>Done</Button>
          <Button size="sm" variant="outline" onClick={() => onMark("skipped")}>Skip</Button>
          <Button size="sm" variant="ghost" onClick={() => onMark("pending")}>Reset</Button>
        </div>
      )}
    </li>
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
