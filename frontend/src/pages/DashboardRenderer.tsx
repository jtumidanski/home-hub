import { useOutletContext } from "react-router-dom";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { parseConfig } from "@/lib/dashboard/parse-config";
import { UnknownWidgetPlaceholder } from "@/components/features/dashboard-widgets/unknown-widget-placeholder";
import { LossyConfigBadge } from "@/components/features/dashboard-widgets/lossy-config-badge";
import type { Dashboard } from "@/types/models/dashboard";
import { GRID_COLUMNS } from "@/lib/dashboard/widget-types";

export function DashboardRenderer() {
  const { dashboard } = useOutletContext<{ dashboard: Dashboard }>();
  const layout = dashboard.attributes.layout;
  const sorted = [...layout.widgets].sort((a, b) => a.y - b.y || a.x - b.x);

  return (
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
  );
}
