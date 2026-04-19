import { ChevronLeft, ChevronRight, Check } from "lucide-react";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  useWorkoutNearestPopulatedWeek,
  useWorkoutWeekSummary,
} from "@/lib/hooks/api/use-workouts";
import { addDays, currentWeekStart } from "@/lib/workout-week";
import {
  DAYS_OF_WEEK_LABELS,
  type PerformanceStatus,
  type SummaryActual,
  type SummaryItem,
  type SummaryPlanned,
  type WorkoutKind,
} from "@/types/models/workout";

// WorkoutReviewPage renders the post-workout review for a single ISO week:
// per-day planned-vs-actual breakdown, status totals, and navigation to
// adjacent weeks (including jump-to-populated when the current week is empty).
export function WorkoutReviewPage() {
  const params = useParams<{ weekStart?: string }>();
  const navigate = useNavigate();
  const weekStart = params.weekStart ?? currentWeekStart();
  const summary = useWorkoutWeekSummary(weekStart);

  const attrs = summary.data?.data.attributes;
  const isEmpty = !!summary.error && !summary.isLoading;

  // Only query /weeks/nearest for the empty-week flow. When summary loaded
  // successfully, we already have previousPopulatedWeek/nextPopulatedWeek
  // threaded into the response.
  const prevJump = useWorkoutNearestPopulatedWeek(weekStart, "prev", isEmpty);
  const nextJump = useWorkoutNearestPopulatedWeek(weekStart, "next", isEmpty);

  const goTo = (target: string | null | undefined) => {
    if (!target) return;
    navigate(`/app/workouts/review/${target}`);
  };

  const previousPopulated = attrs?.previousPopulatedWeek ?? prevJump.data?.data.attributes.weekStartDate ?? null;
  const nextPopulated = attrs?.nextPopulatedWeek ?? nextJump.data?.data.attributes.weekStartDate ?? null;

  return (
    <div className="space-y-4">
      <ReviewNavHeader
        weekStart={weekStart}
        onGoToWeek={(offset) => goTo(addDays(weekStart, offset * 7))}
        previousPopulated={previousPopulated}
        nextPopulated={nextPopulated}
        onJump={goTo}
      />

      {summary.isLoading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : isEmpty ? (
        <EmptyWeekCard weekStart={weekStart} />
      ) : attrs ? (
        <>
          <TotalsCard
            weekStart={attrs.weekStartDate}
            planned={attrs.totalPlannedItems}
            performed={attrs.totalPerformedItems}
            skipped={attrs.totalSkippedItems}
          />
          <PerDayGrid byDay={attrs.byDay} restDayFlags={attrs.restDayFlags} />
        </>
      ) : null}
    </div>
  );
}

// --- Navigation header -----------------------------------------------------

function ReviewNavHeader({
  weekStart,
  onGoToWeek,
  previousPopulated,
  nextPopulated,
  onJump,
}: {
  weekStart: string;
  onGoToWeek: (offset: number) => void;
  previousPopulated: string | null;
  nextPopulated: string | null;
  onJump: (target: string) => void;
}) {
  const jumpDisabled = (target: string | null) => !target || target === weekStart;
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <Button variant="outline" size="sm" onClick={() => onGoToWeek(-1)} aria-label="Previous week">
          <ChevronLeft className="h-4 w-4" />
          Prev
        </Button>
        <span className="text-sm font-medium">Week of {weekStart}</span>
        <Button variant="outline" size="sm" onClick={() => onGoToWeek(1)} aria-label="Next week">
          Next
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
      <div className="flex items-center justify-between text-xs">
        <Button
          variant="ghost"
          size="sm"
          disabled={jumpDisabled(previousPopulated)}
          onClick={() => previousPopulated && onJump(previousPopulated)}
          aria-label={
            previousPopulated
              ? `Jump to previous populated week ${previousPopulated}`
              : "No earlier populated week"
          }
        >
          ↞ Previous populated{previousPopulated ? ` (${previousPopulated})` : ""}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          disabled={jumpDisabled(nextPopulated)}
          onClick={() => nextPopulated && onJump(nextPopulated)}
          aria-label={
            nextPopulated
              ? `Jump to next populated week ${nextPopulated}`
              : "No later populated week"
          }
        >
          Next populated{nextPopulated ? ` (${nextPopulated})` : ""} ↠
        </Button>
      </div>
    </div>
  );
}

// --- Totals card -----------------------------------------------------------

function TotalsCard({
  weekStart,
  planned,
  performed,
  skipped,
}: {
  weekStart: string;
  planned: number;
  performed: number;
  skipped: number;
}) {
  // Pending is derived: anything that was planned but neither done nor
  // explicitly skipped is pending. We floor at zero to guard against weeks
  // where transient data has `performed + skipped > planned`.
  const pending = Math.max(planned - performed - skipped, 0);
  return (
    <Card>
      <CardHeader>
        <CardTitle>Week of {weekStart}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-4 gap-4 text-center">
          <Stat label="Planned" value={planned} />
          <Stat label="Performed" value={performed} />
          <Stat label="Pending" value={pending} />
          <Stat label="Skipped" value={skipped} />
        </div>
      </CardContent>
    </Card>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-xs text-muted-foreground">{label}</p>
    </div>
  );
}

// --- Per-day grid ----------------------------------------------------------

function PerDayGrid({
  byDay,
  restDayFlags,
}: {
  byDay: Array<{ dayOfWeek: number; isRestDay: boolean; items: SummaryItem[] }>;
  restDayFlags: number[];
}) {
  // byDay arrives Mon→Sun from the backend; labels match that ordering.
  return (
    <div className="grid grid-cols-1 md:grid-cols-7 gap-3">
      {DAYS_OF_WEEK_LABELS.map((label, dayIdx) => {
        const block = byDay.find((b) => b.dayOfWeek === dayIdx) ?? {
          dayOfWeek: dayIdx,
          isRestDay: restDayFlags.includes(dayIdx),
          items: [],
        };
        return (
          <section key={dayIdx} aria-label={label}>
            <Card className="min-h-[10rem]">
              <CardHeader className="p-3 pb-2">
                <CardTitle className="flex items-center justify-between text-sm">
                  <h2 className="font-semibold">{label}</h2>
                  {block.isRestDay && (
                    <Badge variant="secondary" className="text-[10px]">Rest</Badge>
                  )}
                </CardTitle>
                <p className="text-xs text-muted-foreground">
                  {block.items.length === 0
                    ? block.isRestDay
                      ? "—"
                      : "Nothing scheduled"
                    : `${block.items.length} ${block.items.length === 1 ? "exercise" : "exercises"}`}
                </p>
              </CardHeader>
              <CardContent className="p-3 pt-0 space-y-2">
                {block.items.map((it) => (
                  <ReviewItemCard key={it.itemId} item={it} />
                ))}
              </CardContent>
            </Card>
          </section>
        );
      })}
    </div>
  );
}

// --- Per-item card ---------------------------------------------------------

function ReviewItemCard({ item }: { item: SummaryItem }) {
  const status = item.status;
  const actual = item.actualSummary;
  const nameClass =
    status === "skipped"
      ? "line-through text-muted-foreground"
      : status === "pending"
      ? "italic text-muted-foreground"
      : "";

  return (
    <div className="rounded border p-2 text-xs space-y-1">
      <div className="flex items-start justify-between gap-2">
        <span className={`font-medium truncate ${nameClass}`}>{item.exerciseName}</span>
        <StatusBadge status={status} />
      </div>
      <p className="text-muted-foreground">Planned: {renderPlanned(item.kind, item.planned)}</p>
      <p className="flex items-center gap-1">
        {renderActual(item.kind, item.status, actual)}
        {targetMet(item.kind, item.planned, actual) && (
          <span aria-label="Target met" title="Target met" className="text-emerald-600">
            <Check className="inline h-3 w-3" />
          </span>
        )}
      </p>
    </div>
  );
}

function StatusBadge({ status }: { status: PerformanceStatus }) {
  const map: Record<PerformanceStatus, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
    done: { label: "Done", variant: "default" },
    partial: { label: "Partial", variant: "outline" },
    skipped: { label: "Skipped", variant: "destructive" },
    pending: { label: "Pending", variant: "secondary" },
  };
  const m = map[status];
  return (
    <Badge variant={m.variant} className="text-[10px] shrink-0">
      {m.label}
    </Badge>
  );
}

// --- planned / actual formatters ------------------------------------------

function renderPlanned(kind: WorkoutKind, p: SummaryPlanned): string {
  switch (kind) {
    case "strength":
      return p.sets && p.reps
        ? `${p.sets}×${p.reps}${p.weight != null ? ` @ ${p.weight} ${p.weightUnit ?? ""}` : ""}`
        : "—";
    case "isometric":
      return p.sets && p.durationSeconds
        ? `${p.sets}×${formatDuration(p.durationSeconds)}`
        : "—";
    case "cardio":
      return formatCardio(p.durationSeconds ?? null, p.distance ?? null, p.distanceUnit ?? null);
  }
}

function renderActual(kind: WorkoutKind, status: PerformanceStatus, a: SummaryActual | null) {
  if (status === "skipped") return <span>Actual: Skipped</span>;
  if (status === "pending" || a == null) return <span>Actual: —</span>;

  if (kind === "strength" && a.setRows && a.setRows.length > 0) {
    // Per-set: list every set. Wraps naturally thanks to flex-wrap at ≥6 sets.
    return (
      <span className="flex flex-wrap gap-x-1">
        <span>Actual:</span>
        {a.setRows.map((row, i) => (
          <span key={row.setNumber} className="text-muted-foreground">
            set {row.setNumber}: {row.reps} @ {row.weight}
            {i < a.setRows!.length - 1 ? " · " : ""}
          </span>
        ))}
      </span>
    );
  }

  switch (kind) {
    case "strength":
      return (
        <span>
          Actual: {a.sets ?? "?"}×{a.reps ?? "?"}
          {a.weight != null ? ` @ ${a.weight} ${a.weightUnit ?? ""}` : ""}
        </span>
      );
    case "isometric":
      return (
        <span>
          Actual: {a.sets ?? "?"}×{a.durationSeconds != null ? formatDuration(a.durationSeconds) : "?"}
        </span>
      );
    case "cardio":
      return (
        <span>Actual: {formatCardio(a.durationSeconds ?? null, a.distance ?? null, a.distanceUnit ?? null)}</span>
      );
  }
}

function formatDuration(totalSeconds: number): string {
  const m = Math.floor(totalSeconds / 60);
  const s = totalSeconds % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

function formatCardio(duration: number | null, distance: number | null, distanceUnit: string | null): string {
  const parts: string[] = [];
  if (duration != null) parts.push(formatDuration(duration));
  if (distance != null) parts.push(`${distance}${distanceUnit ? ` ${distanceUnit}` : ""}`);
  return parts.length === 0 ? "—" : parts.join(" · ");
}

// --- target-met predicate -------------------------------------------------

function targetMet(kind: WorkoutKind, p: SummaryPlanned, a: SummaryActual | null): boolean {
  if (!a) return false;
  switch (kind) {
    case "strength": {
      // Per-set mode: sum actual volume (reps × weight) across all set rows.
      const plannedVolume =
        (p.sets ?? 0) * (p.reps ?? 0) * (p.weight ?? 0);
      if (plannedVolume <= 0) return false;
      let actualVolume: number;
      if (a.setRows && a.setRows.length > 0) {
        actualVolume = a.setRows.reduce((sum, r) => sum + r.reps * r.weight, 0);
      } else {
        if (a.sets == null || a.reps == null || a.weight == null) return false;
        actualVolume = a.sets * a.reps * a.weight;
      }
      return actualVolume >= plannedVolume;
    }
    case "isometric": {
      if (p.sets == null || p.durationSeconds == null) return false;
      if (a.sets == null || a.durationSeconds == null) return false;
      return a.sets * a.durationSeconds >= p.sets * p.durationSeconds;
    }
    case "cardio": {
      if (p.distance != null && a.distance != null) {
        return a.distance >= p.distance;
      }
      if (p.durationSeconds != null && a.durationSeconds != null) {
        return a.durationSeconds >= p.durationSeconds;
      }
      return false;
    }
  }
}

// --- empty-week card ------------------------------------------------------

function EmptyWeekCard({ weekStart }: { weekStart: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Week of {weekStart}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">No workouts logged for this week.</p>
      </CardContent>
    </Card>
  );
}
