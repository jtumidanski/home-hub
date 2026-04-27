import { useReducer, useState } from "react";
import { Link, useNavigate, useOutletContext } from "react-router-dom";
import { useMobile } from "@/lib/hooks/use-mobile";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { Dashboard } from "@/types/models/dashboard";
import { useUpdateDashboard } from "@/lib/hooks/api/use-dashboards";
import { draftReducer, fromServer } from "@/pages/dashboard-designer/state";
import { DesignerGrid } from "@/pages/dashboard-designer/designer-grid";
import { PaletteDrawer } from "@/pages/dashboard-designer/palette-drawer";
import { ConfigPanel } from "@/pages/dashboard-designer/config-panel";
import { useUnsavedGuard } from "@/pages/dashboard-designer/use-unsaved-guard";

/**
 * The dashboard designer. Reads the server-fetched dashboard from the
 * parent `<DashboardShell>` outlet context and drives the editor via the
 * `draftReducer`. Save persists to the API; Discard returns to view mode
 * after confirming on a dirty draft. The dirty guard blocks in-app and
 * browser-level navigation while changes are unsaved.
 */
export default function DashboardDesigner() {
  const { dashboard } = useOutletContext<{ dashboard: Dashboard }>();
  const [state, dispatch] = useReducer(draftReducer, dashboard, fromServer);
  const navigate = useNavigate();
  const updateDashboard = useUpdateDashboard();
  const isMobile = useMobile();

  if (isMobile) {
    return (
      <div
        className="flex min-h-[40vh] items-center justify-center p-6 text-center"
        data-testid="designer-mobile-blocker"
      >
        <div className="max-w-md space-y-3">
          <h2 className="text-lg font-semibold">Editing needs a larger screen</h2>
          <p className="text-sm text-muted-foreground">
            The dashboard designer is only available on tablet-or-wider screens.
            Switch to a larger device to make changes.
          </p>
          <Button render={<Link to=".." />} variant="outline" size="sm">
            View only
          </Button>
        </div>
      </div>
    );
  }

  useUnsavedGuard(state.dirty);
  const [discardConfirmOpen, setDiscardConfirmOpen] = useState(false);

  const onSave = () => {
    updateDashboard.mutate(
      { id: dashboard.id, attrs: { name: state.name, layout: state.layout } },
      {
        onSuccess: (res) => {
          dispatch({ type: "saved", server: res.data });
          navigate("..");
        },
      },
    );
  };

  const onDiscardClick = () => {
    if (state.dirty) {
      setDiscardConfirmOpen(true);
      return;
    }
    navigate("..");
  };

  const confirmDiscard = () => {
    setDiscardConfirmOpen(false);
    // Reset draft so the unsaved guard doesn't fire on the subsequent
    // navigation.
    dispatch({ type: "reset", server: dashboard });
    navigate("..");
  };

  const cancelDiscard = () => {
    setDiscardConfirmOpen(false);
  };

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
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onDiscardClick}
            data-testid="designer-discard"
          >
            Discard
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={onSave}
            disabled={updateDashboard.isPending}
            data-testid="designer-save"
          >
            {updateDashboard.isPending ? "Saving…" : "Save"}
          </Button>
        </div>
      </div>
      <div className="p-2 md:p-4">
        <DesignerGrid widgets={state.layout.widgets} dispatch={dispatch} />
      </div>
      <PaletteDrawer
        open={state.paletteOpen}
        onOpenChange={(open) => dispatch({ type: "toggle-palette", open })}
        dispatch={dispatch}
      />
      <ConfigPanel
        widget={
          state.selectedWidgetId
            ? state.layout.widgets.find((w) => w.id === state.selectedWidgetId) ?? null
            : null
        }
        dispatch={dispatch}
      />

      <Dialog open={discardConfirmOpen} onOpenChange={setDiscardConfirmOpen}>
        <DialogContent data-testid="discard-confirm">
          <DialogHeader>
            <DialogTitle>Discard unsaved changes?</DialogTitle>
            <DialogDescription>
              Your edits to this dashboard will be lost. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={cancelDiscard} data-testid="discard-cancel">
              Keep editing
            </Button>
            <Button onClick={confirmDiscard} data-testid="discard-confirm-button">
              Discard
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
