import type { Dispatch, ReactNode } from "react";
import { GripVertical, Settings, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { findWidget } from "@/lib/dashboard/widget-registry";
import type { WidgetInstance } from "@/lib/dashboard/schema";
import type { DraftAction } from "@/pages/dashboard-designer/state";

interface WidgetChromeProps {
  widget: WidgetInstance;
  dispatch: Dispatch<DraftAction>;
  children: ReactNode;
}

/**
 * Edit-mode chrome around a widget. Adds drag handle, gear, and trash
 * controls plus a subtle overlay so the widget body reads as "disabled"
 * while the designer is active.
 */
export function WidgetChrome({ widget, dispatch, children }: WidgetChromeProps) {
  const hasDefinition = findWidget(widget.type) !== undefined;

  return (
    <div
      className="group relative flex h-full w-full flex-col rounded-md border border-border bg-background shadow-sm"
      data-testid={`widget-chrome-${widget.id}`}
    >
      <div className="flex items-center justify-between border-b border-border/50 bg-muted/40 px-1 py-1">
        <button
          type="button"
          className={cn(
            "widget-drag-handle inline-flex items-center gap-1 rounded px-1 py-0.5 text-muted-foreground hover:bg-muted",
          )}
          style={{ cursor: "grab" }}
          aria-label="Drag widget"
          data-testid={`widget-drag-${widget.id}`}
        >
          <GripVertical className="h-4 w-4" />
        </button>
        <div className="flex items-center gap-1">
          <button
            type="button"
            disabled={!hasDefinition}
            title={hasDefinition ? "Configure widget" : "No config for this widget type"}
            aria-label="Configure widget"
            onClick={() => dispatch({ type: "select", id: widget.id })}
            className={cn(
              "inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground",
              !hasDefinition && "cursor-not-allowed opacity-40",
            )}
            data-testid={`widget-configure-${widget.id}`}
          >
            <Settings className="h-4 w-4" />
          </button>
          <button
            type="button"
            aria-label="Remove widget"
            onClick={() => dispatch({ type: "remove", id: widget.id })}
            className="inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            data-testid={`widget-remove-${widget.id}`}
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>
      <div className="relative flex-1 overflow-hidden p-2 opacity-75 [pointer-events:none]">
        {children}
      </div>
    </div>
  );
}
