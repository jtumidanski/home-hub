import type { Dispatch } from "react";
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { XIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { widgetRegistry } from "@/lib/dashboard/widget-registry";
import type { DraftAction } from "@/pages/dashboard-designer/state";
import type { WidgetInstance } from "@/lib/dashboard/schema";

interface PaletteDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dispatch: Dispatch<DraftAction>;
}

function uuid(): string {
  return (crypto as Crypto).randomUUID();
}

/**
 * Side drawer that lists every widget in the registry. Clicking a palette
 * entry instantiates a new widget at `(0, 0)` with its default size + config
 * and dispatches `add`. Dragging from palette directly onto the grid is a
 * later enhancement — the click-to-add flow covers v1.
 */
export function PaletteDrawer({ open, onOpenChange, dispatch }: PaletteDrawerProps) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Backdrop
          className={cn(
            "fixed inset-0 z-50 bg-black/30 duration-100 data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
          )}
        />
        <DialogPrimitive.Popup
          data-testid="palette-drawer"
          className={cn(
            "fixed top-0 right-0 z-50 flex h-full w-full max-w-sm flex-col gap-2 border-l border-border bg-popover p-4 text-popover-foreground shadow-lg outline-none",
          )}
        >
          <div className="flex items-center justify-between">
            <DialogPrimitive.Title className="font-heading text-base font-medium">
              Add widget
            </DialogPrimitive.Title>
            <DialogPrimitive.Close
              render={<Button variant="ghost" size="icon-sm" />}
              aria-label="Close palette"
            >
              <XIcon />
            </DialogPrimitive.Close>
          </div>
          <DialogPrimitive.Description className="text-sm text-muted-foreground">
            Choose a widget to add to your dashboard. New widgets land at the top.
          </DialogPrimitive.Description>
          <ul className="mt-2 flex flex-col gap-2 overflow-y-auto">
            {widgetRegistry.map((def) => (
              <li key={def.type}>
                <button
                  type="button"
                  data-testid={`palette-add-${def.type}`}
                  onClick={() => {
                    const widget: WidgetInstance = {
                      id: uuid(),
                      type: def.type,
                      x: 0,
                      y: 0,
                      w: def.defaultSize.w,
                      h: def.defaultSize.h,
                      config: def.defaultConfig as Record<string, unknown>,
                    };
                    dispatch({ type: "add", widget });
                    onOpenChange(false);
                  }}
                  className="flex w-full flex-col items-start rounded-md border border-border bg-background p-3 text-left hover:bg-accent"
                >
                  <span className="font-medium">{def.displayName}</span>
                  <span className="text-xs text-muted-foreground">{def.description}</span>
                  <span className="mt-1 text-[10px] text-muted-foreground">
                    Default size {def.defaultSize.w} × {def.defaultSize.h}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
