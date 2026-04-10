import { useEffect } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useCreateTracker } from "@/lib/hooks/api/use-trackers";
import { trackerFormSchema, trackerFormDefaults, type TrackerFormData } from "@/lib/schemas/tracker.schema";
import { COLOR_PALETTE, DAY_LABELS, type ScaleType } from "@/types/models/tracker";
import { cn } from "@/lib/utils";

interface Props {
  open: boolean;
  onClose: () => void;
}

const colorMap: Record<string, string> = {
  red: "bg-red-500", orange: "bg-orange-500", amber: "bg-amber-500", yellow: "bg-yellow-500",
  lime: "bg-lime-500", green: "bg-green-500", emerald: "bg-emerald-500", teal: "bg-teal-500",
  cyan: "bg-cyan-500", blue: "bg-blue-500", indigo: "bg-indigo-500", violet: "bg-violet-500",
  purple: "bg-purple-500", fuchsia: "bg-fuchsia-500", pink: "bg-pink-500", rose: "bg-rose-500",
};

export function CreateTrackerDialog({ open, onClose }: Props) {
  const createMutation = useCreateTracker();
  const { register, control, handleSubmit, watch, reset, formState: { errors } } = useForm<TrackerFormData>({
    resolver: zodResolver(trackerFormSchema),
    defaultValues: trackerFormDefaults,
  });

  const scaleType = watch("scale_type");

  useEffect(() => {
    if (!open) reset(trackerFormDefaults);
  }, [open, reset]);

  const onSubmit = (data: TrackerFormData) => {
    const attrs = {
      name: data.name,
      scale_type: data.scale_type,
      scale_config: data.scale_type === "range" ? { min: data.range_min ?? 0, max: data.range_max ?? 100 } : null,
      schedule: data.schedule,
      color: data.color,
      ...(data.sort_order ? { sort_order: data.sort_order } : {}),
    };
    createMutation.mutate(attrs, { onSuccess: () => onClose() });
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader><DialogTitle>Create Habit</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" placeholder="e.g. Running" {...register("name")} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="space-y-2">
            <Label>Scale Type</Label>
            <Controller control={control} name="scale_type" render={({ field }) => (
              <Select value={field.value} onValueChange={(v) => field.onChange(v as ScaleType)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="sentiment">Sentiment (positive / neutral / negative)</SelectItem>
                  <SelectItem value="numeric">Numeric (counter)</SelectItem>
                  <SelectItem value="range">Range (min-max)</SelectItem>
                </SelectContent>
              </Select>
            )} />
          </div>

          {scaleType === "range" && (
            <div className="grid grid-cols-2 gap-2">
              <div className="space-y-1">
                <Label htmlFor="range_min">Min</Label>
                <Input id="range_min" type="number" {...register("range_min", { valueAsNumber: true })} />
              </div>
              <div className="space-y-1">
                <Label htmlFor="range_max">Max</Label>
                <Input id="range_max" type="number" {...register("range_max", { valueAsNumber: true })} />
                {errors.range_max && <p className="text-xs text-destructive">{errors.range_max.message}</p>}
              </div>
            </div>
          )}

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
                    className={cn("px-2 py-1 text-xs rounded border transition-colors", field.value.includes(i) ? "bg-primary text-primary-foreground" : "bg-muted")}
                    onClick={() => {
                      const next = field.value.includes(i)
                        ? field.value.filter((d: number) => d !== i)
                        : [...field.value, i].sort();
                      field.onChange(next);
                    }}
                  >{label}</button>
                ))}
              </div>
            )} />
          </div>

          <Button type="submit" className="w-full" disabled={createMutation.isPending}>
            {createMutation.isPending ? "Creating..." : "Create"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
