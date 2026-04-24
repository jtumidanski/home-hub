import { useMemo, type Dispatch } from "react";
import RGL, { WidthProvider } from "react-grid-layout/legacy";

// react-grid-layout v2's ESM `/legacy` entrypoint preserves the classic
// HOC API used here; the `./dist/legacy.d.mts` file types it as a
// namespace, so `RGL.Layout` is the single-cell shape RGL calls back
// with in `onLayoutChange`.
type RglLayout = {
  i: string;
  x: number;
  y: number;
  w: number;
  h: number;
};
import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { parseConfig } from "@/lib/dashboard/parse-config";
import { UnknownWidgetPlaceholder } from "@/components/features/dashboard-widgets/unknown-widget-placeholder";
import { GRID_COLUMNS } from "@/lib/dashboard/widget-types";
import type { WidgetInstance } from "@/lib/dashboard/schema";
import type { DraftAction } from "@/pages/dashboard-designer/state";
import { WidgetChrome } from "@/pages/dashboard-designer/widget-chrome";

// WidthProvider measures the container width so <GridLayout> can auto-size.
const ResponsiveGrid = WidthProvider(RGL);

interface DesignerGridProps {
  widgets: WidgetInstance[];
  dispatch: Dispatch<DraftAction>;
}

/**
 * Merges an RGL Layout[] feedback array onto the current WidgetInstance[]
 * preserving type/config/etc. that RGL doesn't track.
 */
function mergeLayout(widgets: WidgetInstance[], next: RglLayout[]): WidgetInstance[] {
  const byId = new Map(next.map((l) => [l.i, l] as const));
  return widgets.map((w) => {
    const l = byId.get(w.id);
    if (!l) return w;
    return { ...w, x: l.x, y: l.y, w: l.w, h: l.h };
  });
}

export function DesignerGrid({ widgets, dispatch }: DesignerGridProps) {
  // Build data-grid per item from the registry (min/max sizing) + current position.
  const items = useMemo(() => {
    return widgets.map((w) => {
      const def = findWidget(w.type);
      const dataGrid: RglLayout = {
        i: w.id,
        x: w.x,
        y: w.y,
        w: w.w,
        h: w.h,
        minW: def?.minSize.w,
        minH: def?.minSize.h,
        maxW: def?.maxSize.w,
        maxH: def?.maxSize.h,
      };
      return { widget: w, dataGrid };
    });
  }, [widgets]);

  return (
    <ResponsiveGrid
      className="designer-grid"
      cols={GRID_COLUMNS}
      rowHeight={60}
      margin={[8, 8]}
      compactType="vertical"
      isBounded
      draggableHandle=".widget-drag-handle"
      onLayoutChange={(next: RglLayout[]) =>
        dispatch({ type: "move-or-resize", widgets: mergeLayout(widgets, next) })
      }
    >
      {items.map(({ widget, dataGrid }) => {
        const def = findWidget(widget.type);
        return (
          <div key={widget.id} data-grid={dataGrid} data-testid={`grid-item-${widget.id}`}>
            <WidgetChrome widget={widget} dispatch={dispatch}>
              {def ? (
                (() => {
                  const { config } = parseConfig(def, widget.config);
                  const Comp = def.component as React.ComponentType<{ config: unknown }>;
                  return <Comp config={config} />;
                })()
              ) : (
                <UnknownWidgetPlaceholder type={widget.type} />
              )}
            </WidgetChrome>
          </div>
        );
      })}
    </ResponsiveGrid>
  );
}
