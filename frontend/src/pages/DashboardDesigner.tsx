import { useReducer } from "react";
import { useOutletContext } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Dashboard } from "@/types/models/dashboard";
import { draftReducer, fromServer } from "@/pages/dashboard-designer/state";
import { DesignerGrid } from "@/pages/dashboard-designer/designer-grid";

/**
 * The dashboard designer. Reads the server-fetched dashboard from the
 * parent `<DashboardShell>` outlet context and drives the editor via the
 * `draftReducer`. Palette, config panel, save/discard, and below-tablet
 * blocker all land in subsequent tasks.
 */
export default function DashboardDesigner() {
  const { dashboard } = useOutletContext<{ dashboard: Dashboard }>();
  const [state, dispatch] = useReducer(draftReducer, dashboard, fromServer);

  return (
    <div className="flex flex-col" data-testid="dashboard-designer">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border bg-muted/40 p-3 md:px-6">
        <div className="flex items-center gap-2">
          <Input
            aria-label="Dashboard name"
            value={state.name}
            onChange={(e) => dispatch({ type: "rename", name: e.target.value })}
            className="max-w-xs"
          />
          {state.dirty ? (
            <span className="text-xs text-muted-foreground">Unsaved changes</span>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => dispatch({ type: "toggle-palette", open: !state.paletteOpen })}
            data-testid="designer-toggle-palette"
          >
            Add widget
          </Button>
          <Button type="button" variant="outline" size="sm" data-testid="designer-discard">
            Discard
          </Button>
          <Button type="button" size="sm" data-testid="designer-save">
            Save
          </Button>
        </div>
      </div>
      <div className="p-2 md:p-4">
        <DesignerGrid widgets={state.layout.widgets} dispatch={dispatch} />
      </div>
    </div>
  );
}
