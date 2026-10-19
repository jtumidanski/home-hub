import type { Dispatch, ReactNode } from "react";
import { GripVertical, Settings, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
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
      className="group relative flex h-full w-full flex-col rounded-md border border-border bg-card"
      data-testid={`widget-chrome-${widget.id}`}
    >
      <div className="flex items-center justify-between border-b border-border bg-muted px-1 py-1">
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="widget-drag-handle cursor-grab text-muted-foreground"
          aria-label="Drag widget"
          data-testid={`widget-drag-${widget.id}`}
        >
          <GripVertical />
        </Button>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            disabled={!hasDefinition}
            aria-disabled={!hasDefinition}
            title={hasDefinition ? "Configure widget" : "No config for this widget type"}
            aria-label="Configure widget"
            onClick={() => dispatch({ type: "select", id: widget.id })}
            className="text-muted-foreground hover:text-foreground disabled:cursor-not-allowed"
            data-testid={`widget-configure-${widget.id}`}
          >
            <Settings />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label="Remove widget"
            onClick={() => dispatch({ type: "remove", id: widget.id })}
            className="text-muted-foreground hover:text-destructive"
            data-testid={`widget-remove-${widget.id}`}
          >
            <Trash2 />
          </Button>
        </div>
      </div>
      <div className="relative flex-1 overflow-hidden p-2 [pointer-events:none]">
        {children}
        <div aria-hidden="true" className="absolute inset-0 bg-card/40" />
      </div>
    </div>
  );
}
