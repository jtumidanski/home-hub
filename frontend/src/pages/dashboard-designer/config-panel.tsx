import { useEffect, type Dispatch } from "react";
import { useForm } from "react-hook-form";
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { XIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { findWidget } from "@/lib/dashboard/widget-registry";
import { parseConfig } from "@/lib/dashboard/parse-config";
import type { WidgetInstance } from "@/lib/dashboard/schema";
import type { DraftAction } from "@/pages/dashboard-designer/state";
import { ZodFormFields } from "@/pages/dashboard-designer/zod-form";

interface ConfigPanelProps {
  widget: WidgetInstance | null;
  dispatch: Dispatch<DraftAction>;
}

/**
 * Right-hand panel that edits a widget's `config` against its registered
 * `configSchema`. Opens when a widget is selected; closes via the
 * `Cancel` button, the backdrop, or a successful `Apply`.
 */
export function ConfigPanel({ widget, dispatch }: ConfigPanelProps) {
  const def = widget ? findWidget(widget.type) : undefined;

  const defaultValues = widget && def ? parseConfig(def, widget.config).config : {};

  const form = useForm<Record<string, unknown>>({
    defaultValues: defaultValues as Record<string, unknown>,
  });

  // Whenever a different widget is selected, reset the form to its values.
  useEffect(() => {
    if (widget && def) {
      form.reset(parseConfig(def, widget.config).config as Record<string, unknown>);
    }
  }, [widget, def, form]);

  const open = widget !== null;

  const close = () => dispatch({ type: "select", id: null });

  if (!open || !widget || !def) {
    // Still render the Root so base-ui can manage focus/portal state, but
    // keep it closed. The form inside never instantiates without a widget.
    return (
      <DialogPrimitive.Root open={false} onOpenChange={close}>
        <DialogPrimitive.Portal>
          <DialogPrimitive.Popup className="hidden" />
        </DialogPrimitive.Portal>
      </DialogPrimitive.Root>
    );
  }

  const onApply = form.handleSubmit((values) => {
    const parsed = def.configSchema.safeParse(values);
    if (!parsed.success) {
      form.setError("root", { message: parsed.error.message });
      return;
    }
    dispatch({
      type: "update-config",
      id: widget.id,
      config: parsed.data as Record<string, unknown>,
    });
    close();
  });

  const onResetDefaults = () => {
    // Build a full record from the schema shape so keys not present in
    // `defaultConfig` are explicitly cleared rather than left at their
    // previous value.
    const shape =
      ((def.configSchema as unknown as { shape?: Record<string, unknown> }).shape) ?? {};
    const cleared: Record<string, unknown> = {};
    // RHF's `reset()` skips keys whose new value is `undefined`, so we
    // seed every schema field with `""` to guarantee a state change. Any
    // key that has a real default overrides it below.
    for (const key of Object.keys(shape)) cleared[key] = "";
    const merged: Record<string, unknown> = {
      ...cleared,
      ...(def.defaultConfig as Record<string, unknown>),
    };
    form.reset(merged);
  };

  return (
    <DialogPrimitive.Root open={open} onOpenChange={(next) => (next ? null : close())}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Backdrop
          className={cn(
            "fixed inset-0 z-50 bg-black/30 duration-100 data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
          )}
        />
        <DialogPrimitive.Popup
          data-testid="config-panel"
          className={cn(
            "fixed top-0 right-0 z-50 flex h-full w-full max-w-sm flex-col gap-3 border-l border-border bg-popover p-4 text-popover-foreground shadow-lg outline-none",
          )}
        >
          <div className="flex items-center justify-between">
            <DialogPrimitive.Title className="font-heading text-base font-medium">
              Configure {def.displayName}
            </DialogPrimitive.Title>
            <DialogPrimitive.Close
              render={<Button variant="ghost" size="icon-sm" />}
              aria-label="Close config panel"
            >
              <XIcon />
            </DialogPrimitive.Close>
          </div>
          <form
            onSubmit={onApply}
            className="flex flex-1 flex-col gap-3 overflow-y-auto"
            data-testid="config-form"
          >
            <ZodFormFields schema={def.configSchema} form={form} />
            <div className="mt-auto flex flex-col gap-2 border-t border-border pt-2 sm:flex-row sm:justify-end">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={onResetDefaults}
                data-testid="config-reset"
              >
                Reset to defaults
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={close}
                data-testid="config-cancel"
              >
                Cancel
              </Button>
              <Button type="submit" size="sm" data-testid="config-apply">
                Apply
              </Button>
            </div>
          </form>
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
