import { useParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useWorkoutWeekSummary } from "@/lib/hooks/api/use-workouts";
import { currentWeekStart } from "@/lib/workout-week";

// Per-week summary view: header totals, per-theme block, per-region block.
// Per-day breakdown is rendered as a simple ordered list — the focus here is
// on totals, which are the headline value of the report.
export function WorkoutSummaryPage() {
  const params = useParams<{ weekStart?: string }>();
  const weekStart = params.weekStart ?? currentWeekStart();
  const { data, isLoading, error } = useWorkoutWeekSummary(weekStart);

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>;
  if (error) return <p className="text-destructive">No summary available — week is empty.</p>;
  if (!data) return null;

  const a = data.data.attributes;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Week of {a.weekStartDate}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4 text-center">
            <Stat label="Planned" value={a.totalPlannedItems} />
            <Stat label="Performed" value={a.totalPerformedItems} />
            <Stat label="Skipped" value={a.totalSkippedItems} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>By Theme</CardTitle>
        </CardHeader>
        <CardContent>
          {a.byTheme.length === 0 ? (
            <p className="text-muted-foreground text-sm">No theme data.</p>
          ) : (
            <ul className="space-y-1 text-sm">
              {a.byTheme.map((g) => (
                <li key={g.themeId} className="flex justify-between">
                  <span>{g.themeName}</span>
                  <span className="text-muted-foreground">
                    {g.itemCount} item(s)
                    {g.strengthVolume && ` · ${g.strengthVolume.value.toFixed(0)} ${g.strengthVolume.unit}`}
                    {g.cardio && ` · ${g.cardio.totalDistance.value.toFixed(1)} ${g.cardio.totalDistance.unit}`}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>By Region</CardTitle>
        </CardHeader>
        <CardContent>
          {a.byRegion.length === 0 ? (
            <p className="text-muted-foreground text-sm">No region data.</p>
          ) : (
            <ul className="space-y-1 text-sm">
              {a.byRegion.map((g) => (
                <li key={g.regionId} className="flex justify-between">
                  <span>{g.regionName}</span>
                  <span className="text-muted-foreground">
                    {g.itemCount} item(s)
                    {g.strengthVolume && ` · ${g.strengthVolume.value.toFixed(0)} ${g.strengthVolume.unit}`}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
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
