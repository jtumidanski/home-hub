import { useEffect } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useUpdateTracker } from "@/lib/hooks/api/use-trackers";
import { trackerEditSchema, type TrackerEditData } from "@/lib/schemas/tracker.schema";
import { COLOR_PALETTE, DAY_LABELS, type Tracker } from "@/types/models/tracker";
import { Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { swatchClass } from "@/lib/constants/swatches";

interface Props {
  open: boolean;
  onClose: () => void;
  tracker: Tracker;
}

export function EditTrackerDialog({ open, onClose, tracker }: Props) {
  const updateMutation = useUpdateTracker();
  const { register, control, handleSubmit, reset, formState: { errors } } = useForm<TrackerEditData>({
    resolver: zodResolver(trackerEditSchema),
    defaultValues: {
      name: tracker.attributes.name,
      color: tracker.attributes.color,
      schedule: tracker.attributes.schedule,
      sort_order: tracker.attributes.sort_order,
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        name: tracker.attributes.name,
        color: tracker.attributes.color,
        schedule: tracker.attributes.schedule,
        sort_order: tracker.attributes.sort_order,
      });
    }
  }, [open, tracker, reset]);

  const onSubmit = (data: TrackerEditData) => {
    const attrs: Record<string, unknown> = {};
    if (data.name && data.name !== tracker.attributes.name) attrs.name = data.name;
    if (data.color && data.color !== tracker.attributes.color) attrs.color = data.color;
    if (data.schedule) attrs.schedule = data.schedule;
    if (data.sort_order !== undefined) attrs.sort_order = data.sort_order;

    updateMutation.mutate({ id: tracker.id, attrs }, { onSuccess: () => onClose() });
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader><DialogTitle>Edit {tracker.attributes.name}</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-name">Name</Label>
            <Input id="edit-name" {...register("name")} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="space-y-2">
            <Label>Scale Type</Label>
            <Input value={tracker.attributes.scale_type} disabled className="text-muted-foreground" />
            <p className="text-xs text-muted-foreground">Scale type cannot be changed</p>
          </div>

          <div className="space-y-2">
            <Label>Color</Label>
            <Controller control={control} name="color" render={({ field }) => (
              <div className="flex flex-wrap gap-2">
                {COLOR_PALETTE.map((c) => (
                  <button key={c} type="button" aria-label={c} aria-pressed={field.value === c}
                    className={cn("flex h-8 w-8 items-center justify-center rounded-full border-2 motion-safe:transition-colors", swatchClass(c, "solid"), field.value === c ? "border-foreground motion-safe:scale-110" : "border-transparent")}
                    onClick={() => field.onChange(c)}
                  >
                    {field.value === c && <span className="flex h-4 w-4 items-center justify-center rounded-full bg-background text-foreground"><Check className="h-3 w-3" aria-hidden="true" /></span>}
                  </button>
                ))}
              </div>
            )} />
          </div>

          <div className="space-y-2">
            <Label>Schedule (empty = every day)</Label>
            <Controller control={control} name="schedule" render={({ field }) => (
              <div className="flex gap-1">
                {DAY_LABELS.map((label, i) => (
                  <button key={i} type="button"
                    className={cn("px-2 py-1 text-xs rounded border transition-colors", (field.value ?? []).includes(i) ? "bg-primary text-primary-foreground" : "bg-muted")}
                    onClick={() => {
                      const current = field.value ?? [];
                      const next = current.includes(i)
                        ? current.filter((d: number) => d !== i)
                        : [...current, i].sort();
                      field.onChange(next);
                    }}
                  >{label}</button>
                ))}
              </div>
            )} />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-sort">Sort Order</Label>
            <Input id="edit-sort" type="number" {...register("sort_order", { valueAsNumber: true })} />
          </div>

          <Button type="submit" className="w-full" disabled={updateMutation.isPending}>
            {updateMutation.isPending ? "Saving..." : "Save Changes"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
