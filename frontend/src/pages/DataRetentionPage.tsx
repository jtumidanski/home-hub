import { useMemo, useState } from "react";
import {
  useRetentionPolicies,
  useRetentionRuns,
  usePatchHouseholdRetention,
  usePatchUserRetention,
  usePurgeRetention,
} from "@/lib/hooks/api/use-retention";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorCard } from "@/components/common/error-card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { retentionService, type RetentionCategoryView } from "@/services/api/retention";
import { useTenant } from "@/context/tenant-context";
import { toast } from "sonner";

const CATEGORY_LABELS: Record<string, string> = {
  "productivity.completed_tasks": "Completed tasks",
  "productivity.deleted_tasks_restore_window": "Deleted tasks (restore window)",
  "recipe.deleted_recipes_restore_window": "Deleted recipes (restore window)",
  "recipe.restoration_audit": "Recipe restoration audit",
  "tracker.entries": "Tracker entries",
  "tracker.deleted_items_restore_window": "Deleted trackers (restore window)",
  "workout.performances": "Workout performances",
  "workout.deleted_catalog_restore_window": "Deleted workouts (restore window)",
  "calendar.past_events": "Past calendar events",
  "package.archive_window": "Package archive window",
  "package.archived_delete_window": "Archived packages (delete after)",
  "system.retention_audit": "Retention audit log",
};

function categoryMax(category: string): number {
  return category.endsWith("_restore_window") ? 365 : 3650;
}

interface RowProps {
  category: string;
  current: RetentionCategoryView;
  scope: "household" | "user";
  onSave: (category: string, days: number, currentValue: number) => Promise<void>;
  onPurge: (category: string, scope: "household" | "user") => Promise<void>;
  saving: boolean;
}

function CategoryRow({ category, current, scope, onSave, onPurge, saving }: RowProps) {
  const [value, setValue] = useState<number>(current.days);
  const dirty = value !== current.days;
  const max = categoryMax(category);

  return (
    <div className="flex flex-wrap items-center gap-3 py-3 border-b last:border-b-0">
      <div className="flex-1 min-w-[180px]">
        <div className="font-medium text-sm">{CATEGORY_LABELS[category] ?? category}</div>
        <div className="text-xs text-muted-foreground">{category}</div>
      </div>
      <Badge variant={current.source === "default" ? "secondary" : "default"}>
        {current.source}
      </Badge>
      <div className="flex items-center gap-2">
        <Input
          type="number"
          min={1}
          max={max}
          value={value}
          onChange={(e) => setValue(parseInt(e.target.value || "0", 10))}
          className="w-24"
        />
        <span className="text-xs text-muted-foreground">days</span>
      </div>
      <Button
        size="sm"
        disabled={!dirty || saving || value < 1 || value > max}
        onClick={() => onSave(category, value, current.days)}
      >
        Save
      </Button>
      <Button size="sm" variant="outline" onClick={() => onPurge(category, scope)}>
        Purge now
      </Button>
    </div>
  );
}

export function DataRetentionPage() {
  const { tenant } = useTenant();
  const { data, isLoading, error } = useRetentionPolicies();
  const runsQuery = useRetentionRuns({ limit: 20 });
  const patchHousehold = usePatchHouseholdRetention();
  const patchUser = usePatchUserRetention();
  const purge = usePurgeRetention();

  const [shrinkWarning, setShrinkWarning] = useState<{
    category: string;
    scope: "household" | "user";
    days: number;
    estimate: number | null;
  } | null>(null);

  const policies = data?.data?.attributes;
  const householdId = policies?.household?.id;

  const householdRows = useMemo(() => {
    if (!policies?.household) return [];
    return Object.entries(policies.household.categories).sort(([a], [b]) => a.localeCompare(b));
  }, [policies?.household]);

  const userRows = useMemo(() => {
    if (!policies?.user) return [];
    return Object.entries(policies.user.categories).sort(([a], [b]) => a.localeCompare(b));
  }, [policies?.user]);

  async function handleSave(category: string, days: number, currentValue: number) {
    const scope: "household" | "user" = category in (policies?.household?.categories ?? {}) ? "household" : "user";

    // Shrink-warning: when reducing the window below the current effective value,
    // call the dry-run path first to preview row counts and confirm.
    if (days < currentValue && tenant) {
      try {
        const preview = await retentionService.purge(tenant, category, scope, true);
        setShrinkWarning({
          category,
          scope,
          days,
          estimate: preview.data?.attributes?.deleted ?? 0,
        });
        return;
      } catch (e) {
        // If preview fails, still let the user opt in via the modal with no estimate.
        setShrinkWarning({ category, scope, days, estimate: null });
        return;
      }
    }

    await applyPatch(scope, category, days);
  }

  async function applyPatch(scope: "household" | "user", category: string, days: number) {
    if (scope === "household") {
      if (!householdId) {
        toast.error("No active household");
        return;
      }
      await patchHousehold.mutateAsync({ householdId, categories: { [category]: days } });
    } else {
      await patchUser.mutateAsync({ [category]: days });
    }
  }

  async function handlePurge(category: string, scope: "household" | "user") {
    try {
      const result = await purge.mutateAsync({ category, scope });
      toast.success(`Purged ${result.data?.attributes?.deleted ?? 0} rows`);
    } catch {
      // toast handled in hook
    }
  }

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (error || !policies) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load retention policies." />
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 space-y-6">
      <div>
        <h1 className="text-xl md:text-2xl font-semibold">Data Retention</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Configure how long Home Hub keeps each kind of data. Lower windows shrink your footprint at the cost of less history.
        </p>
      </div>

      {policies.household && (
        <Card>
          <CardHeader>
            <CardTitle>Household data</CardTitle>
            <CardDescription>Applies to everyone in this household.</CardDescription>
          </CardHeader>
          <CardContent>
            {householdRows.map(([cat, view]) => (
              <CategoryRow
                key={cat}
                category={cat}
                current={view}
                scope="household"
                onSave={handleSave}
                onPurge={handlePurge}
                saving={patchHousehold.isPending}
              />
            ))}
          </CardContent>
        </Card>
      )}

      {policies.user && (
        <Card>
          <CardHeader>
            <CardTitle>My personal data</CardTitle>
            <CardDescription>Only applies to your own data.</CardDescription>
          </CardHeader>
          <CardContent>
            {userRows.map(([cat, view]) => (
              <CategoryRow
                key={cat}
                category={cat}
                current={view}
                scope="user"
                onSave={handleSave}
                onPurge={handlePurge}
                saving={patchUser.isPending}
              />
            ))}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Recent purges</CardTitle>
          <CardDescription>Last 20 retention reaper runs across services.</CardDescription>
        </CardHeader>
        <CardContent>
          {runsQuery.isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : !runsQuery.data?.data?.length ? (
            <p className="text-sm text-muted-foreground">No reaper runs yet.</p>
          ) : (
            <div className="space-y-1 text-sm">
              {runsQuery.data.data.map((run) => (
                <div key={run.id} className="flex flex-wrap items-center gap-2 py-2 border-b last:border-b-0">
                  <Badge variant="outline">{run.attributes.service}</Badge>
                  <span className="font-medium">{CATEGORY_LABELS[run.attributes.category] ?? run.attributes.category}</span>
                  <span className="text-muted-foreground">{run.attributes.trigger}</span>
                  <span className="ml-auto text-muted-foreground">
                    {run.attributes.deleted} deleted of {run.attributes.scanned} scanned
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {new Date(run.attributes.started_at).toLocaleString()}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!shrinkWarning} onOpenChange={(o) => !o && setShrinkWarning(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Shrink retention window?</DialogTitle>
            <DialogDescription>
              {shrinkWarning && (
                <>
                  Lowering <strong>{CATEGORY_LABELS[shrinkWarning.category] ?? shrinkWarning.category}</strong> to{" "}
                  <strong>{shrinkWarning.days} days</strong> will permanently delete approximately{" "}
                  <strong>{shrinkWarning.estimate ?? "?"}</strong> rows on the next reaper run.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShrinkWarning(null)}>
              Cancel
            </Button>
            <Button
              onClick={async () => {
                if (!shrinkWarning) return;
                await applyPatch(shrinkWarning.scope, shrinkWarning.category, shrinkWarning.days);
                setShrinkWarning(null);
              }}
            >
              Confirm
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
