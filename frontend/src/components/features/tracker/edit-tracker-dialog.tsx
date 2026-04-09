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
import { cn } from "@/lib/utils";

interface Props {
  open: boolean;
  onClose: () => void;
  tracker: Tracker;
}

const colorMap: Record<string, string> = {
  red: "bg-red-500", orange: "bg-orange-500", amber: "bg-amber-500", yellow: "bg-yellow-500",
  lime: "bg-lime-500", green: "bg-green-500", emerald: "bg-emerald-500", teal: "bg-teal-500",
  cyan: "bg-cyan-500", blue: "bg-blue-500", indigo: "bg-indigo-500", violet: "bg-violet-500",
  purple: "bg-purple-500", fuchsia: "bg-fuchsia-500", pink: "bg-pink-500", rose: "bg-rose-500",
};

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
                  <button key={c} type="button"
                    className={cn("w-7 h-7 rounded-full border-2 transition-all", colorMap[c], field.value === c ? "border-foreground scale-110" : "border-transparent")}
                    onClick={() => field.onChange(c)}
                  />
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
