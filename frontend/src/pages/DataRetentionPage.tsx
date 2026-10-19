import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  useRetentionPolicies,
  useRetentionRuns,
  usePatchHouseholdRetention,
  usePatchUserRetention,
  usePurgeRetention,
} from "@/lib/hooks/api/use-retention";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorCard } from "@/components/common/error-card";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { retentionService, type RetentionCategoryView } from "@/services/api/retention";
import { useTenant } from "@/context/tenant-context";
import { History } from "lucide-react";
import { toast } from "sonner";

function retentionDaysSchema(max: number) {
  return z.object({
    days: z.number({ error: "Must be a number" }).int().min(1, "Minimum 1 day").max(max, `Maximum ${max} days`),
  });
}
type RetentionDaysData = z.infer<ReturnType<typeof retentionDaysSchema>>;

const CATEGORY_LABELS: Record<string, string> = {
  "productivity.completed_tasks": "Completed tasks",
  "productivity.deleted_tasks_restore_window": "Deleted tasks (restore window)",
  "productivity.reminders": "Reminders",
  "productivity.deleted_reminders_restore_window": "Deleted reminders (restore window)",
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
  onPurge: (category: string, scope: "household" | "user") => void;
  saving: boolean;
}

function CategoryRow({ category, current, scope, onSave, onPurge, saving }: RowProps) {
  const max = categoryMax(category);
  const schema = useMemo(() => retentionDaysSchema(max), [max]);
  const { register, handleSubmit, formState: { errors, isDirty } } = useForm<RetentionDaysData>({
    resolver: zodResolver(schema),
    defaultValues: { days: current.days },
  });
  const label = CATEGORY_LABELS[category] ?? category;
  const errorId = `${category}-days-error`;

  return (
    <div className="flex flex-col gap-3 border-b py-3 last:border-b-0 sm:flex-row sm:items-center">
      <div className="flex min-w-0 flex-1 items-center gap-2" title={category}>
        <span className="text-sm font-medium">{label}</span>
        <Badge variant={current.source === "default" ? "secondary" : "default"}>
          {current.source === "default" ? "Default" : "Custom"}
        </Badge>
      </div>
      <div className="flex items-start gap-2">
        <div>
          <Input
            type="number"
            {...register("days", { valueAsNumber: true })}
            aria-label={`${label} retention days`}
            aria-invalid={errors.days ? true : undefined}
            aria-describedby={errors.days ? errorId : undefined}
            className="w-24"
          />
          {errors.days && (
            <p id={errorId} className="mt-1 text-xs text-destructive">
              {errors.days.message}
            </p>
          )}
        </div>
        <span className="pt-2 text-xs text-muted-foreground">days</span>
        <Button
          size="sm"
          disabled={!isDirty || saving || !!errors.days}
          onClick={handleSubmit((data) => onSave(category, data.days, current.days))}
        >
          Save
        </Button>
      </div>
      <Button
        size="sm"
        variant="outline"
        className="text-destructive sm:ml-2"
        aria-label={`Purge ${label} now`}
        onClick={() => onPurge(category, scope)}
      >
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

  const [purgeTarget, setPurgeTarget] = useState<{
    category: string;
    scope: "household" | "user";
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
      } catch {
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

  function handlePurge(category: string, scope: "household" | "user") {
    setPurgeTarget({ category, scope });
  }

  async function confirmPurge() {
    if (!purgeTarget) return;
    const { category, scope } = purgeTarget;
    try {
      const result = await purge.mutateAsync({ category, scope });
      toast.success(`Purged ${result.data?.attributes?.deleted ?? 0} rows`);
    } catch {
      // toast handled in hook
    } finally {
      setPurgeTarget(null);
    }
  }

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-6" role="status" aria-label="Loading">
        <PageHeader title="Data Retention" />
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
      <PageHeader
        title="Data Retention"
        description="Configure how long Home Hub keeps each kind of data. Lower windows shrink your footprint at the cost of less history."
      />

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
            <EmptyState
              icon={History}
              title="No reaper runs yet"
              description="Purges will be listed here once the retention reaper has run."
              headingLevel="h3"
            />
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

      <ConfirmDialog
        open={!!shrinkWarning}
        onOpenChange={(o) => {
          if (!o) setShrinkWarning(null);
        }}
        title="Shrink retention window?"
        description={
          shrinkWarning && (
            <>
              Lowering <strong>{CATEGORY_LABELS[shrinkWarning.category] ?? shrinkWarning.category}</strong> to{" "}
              <strong>{shrinkWarning.days} days</strong> will permanently delete approximately{" "}
              <strong>{shrinkWarning.estimate ?? "?"}</strong> rows on the next reaper run.
            </>
          )
        }
        confirmLabel="Confirm"
        pending={patchHousehold.isPending || patchUser.isPending}
        onConfirm={async () => {
          if (!shrinkWarning) return;
          await applyPatch(shrinkWarning.scope, shrinkWarning.category, shrinkWarning.days);
          setShrinkWarning(null);
        }}
      />

      <ConfirmDialog
        open={!!purgeTarget}
        onOpenChange={(o) => {
          if (!o) setPurgeTarget(null);
        }}
        title="Purge now?"
        description={
          purgeTarget && (
            <>
              This will permanently delete expired{" "}
              <strong>{CATEGORY_LABELS[purgeTarget.category] ?? purgeTarget.category}</strong> data immediately. This cannot be undone.
            </>
          )
        }
        confirmLabel="Purge now"
        pending={purge.isPending}
        onConfirm={confirmPurge}
      />
    </div>
  );
}
