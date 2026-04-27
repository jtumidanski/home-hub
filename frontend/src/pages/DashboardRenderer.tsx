import { useCallback } from "react";
import { useOutletContext } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { parseConfig } from "@/lib/dashboard/parse-config";
import { UnknownWidgetPlaceholder } from "@/components/features/dashboard-widgets/unknown-widget-placeholder";
import { LossyConfigBadge } from "@/components/features/dashboard-widgets/lossy-config-badge";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import type { Dashboard } from "@/types/models/dashboard";
import { GRID_COLUMNS } from "@/lib/dashboard/widget-types";
import { useTenant } from "@/context/tenant-context";
import { taskKeys } from "@/lib/hooks/api/use-tasks";
import { reminderKeys } from "@/lib/hooks/api/use-reminders";
import { mealKeys } from "@/lib/hooks/api/use-meals";
import { calendarKeys } from "@/lib/hooks/api/use-calendar";
import { trackerKeys } from "@/lib/hooks/api/use-trackers";
import { workoutKeys } from "@/lib/hooks/api/use-workouts";
import { packageKeys } from "@/lib/hooks/api/use-packages";

export function DashboardRenderer() {
  const { dashboard } = useOutletContext<{ dashboard: Dashboard }>();
  const layout = dashboard.attributes.layout;
  const sorted = [...layout.widgets].sort((a, b) => a.y - b.y || a.x - b.x);

  const queryClient = useQueryClient();
  const { tenant, household } = useTenant();

  const handleRefresh = useCallback(async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: taskKeys.all(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: reminderKeys.all(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: packageKeys.summary(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: calendarKeys.all(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: trackerKeys.todayAll(tenant, household) }),
      queryClient.invalidateQueries({ queryKey: workoutKeys.todayAll(tenant, household) }),
    ]);
  }, [queryClient, tenant, household]);

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6">
      <div
        className="hidden md:grid gap-4"
        style={{ gridTemplateColumns: `repeat(${GRID_COLUMNS}, minmax(0, 1fr))` }}
        data-testid="dashboard-renderer-grid"
      >
        {sorted.map((w) => {
          const def = findWidget(w.type);
          if (!def) {
            return (
              <div
                key={w.id}
                style={{ gridColumn: `span ${w.w}`, gridRow: `span ${w.h}` }}
                data-testid={`widget-slot-${w.id}`}
              >
                <UnknownWidgetPlaceholder type={w.type} />
              </div>
            );
          }
          const { config, lossy } = parseConfig(def, w.config);
          const Comp = def.component as React.ComponentType<{ config: unknown }>;
          return (
            <div
              key={w.id}
              style={{ gridColumn: `span ${w.w}`, gridRow: `span ${w.h}` }}
              className="relative"
              data-testid={`widget-slot-${w.id}`}
            >
              <Comp config={config} />
              {lossy && <LossyConfigBadge />}
            </div>
          );
        })}
      </div>
      <div className="grid md:hidden grid-cols-1 gap-4" data-testid="dashboard-renderer-stack">
        {sorted.map((w) => {
          const def = findWidget(w.type);
          if (!def) return <UnknownWidgetPlaceholder key={w.id} type={w.type} />;
          const { config } = parseConfig(def, w.config);
          const Comp = def.component as React.ComponentType<{ config: unknown }>;
          return <Comp key={w.id} config={config} />;
        })}
      </div>
      </div>
    </PullToRefresh>
  );
}
