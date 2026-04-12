import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCreatePackage, useDetectCarrier } from "@/lib/hooks/api/use-packages";
import { packageFormSchema, packageFormDefaults, type PackageFormData } from "@/lib/schemas/package.schema";
import { CARRIER_LABELS } from "@/types/models/package";

interface CreatePackageDialogProps {
  open: boolean;
  onClose: () => void;
}

export function CreatePackageDialog({ open, onClose }: CreatePackageDialogProps) {
  const createMutation = useCreatePackage();
  const detectMutation = useDetectCarrier();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors },
  } = useForm<PackageFormData>({
    resolver: zodResolver(packageFormSchema),
    defaultValues: packageFormDefaults,
  });

  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const trackingNumber = watch("trackingNumber");

  useEffect(() => {
    if (!open) {
      reset(packageFormDefaults);
    }
  }, [open, reset]);

  const handleTrackingBlur = () => {
    if (trackingNumber.trim().length >= 8) {
      detectMutation.mutate(trackingNumber.trim(), {
        onSuccess: (resp) => {
          const detected = resp.data.attributes.detectedCarrier;
          if (detected) {
            setValue("carrier", detected as "usps" | "ups" | "fedex");
          }
        },
      });
    }
  };

  const onSubmit = (data: PackageFormData) => {
    const attrs: {
      trackingNumber: string;
      carrier: string;
      label?: string;
      notes?: string;
      private?: boolean;
    } = {
      trackingNumber: data.trackingNumber.trim(),
      carrier: data.carrier,
    };
    if (data.label) attrs.label = data.label;
    if (data.notes) attrs.notes = data.notes;
    if (data.private) attrs.private = data.private;

    createMutation.mutate(
      attrs,
      {
        onSuccess: () => {
          onClose();
        },
      }
    );
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Package</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="trackingNumber">Tracking Number</Label>
            <Input
              id="trackingNumber"
              placeholder="Enter tracking number"
              {...register("trackingNumber")}
              onBlur={handleTrackingBlur}
              onPaste={() => setTimeout(handleTrackingBlur, 100)}
            />
            {errors.trackingNumber && (
              <p className="text-xs text-destructive">{errors.trackingNumber.message}</p>
            )}
            {detectMutation.data?.data.attributes.confidence === "low" && (
              <p className="text-xs text-amber-600">
                Carrier could not be auto-detected. Please select manually.
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="carrier">Carrier</Label>
            <Select
              value={watch("carrier")}
              onValueChange={(v) => setValue("carrier", v as "usps" | "ups" | "fedex")}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select carrier" />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(CARRIER_LABELS).map(([value, label]) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.carrier && (
              <p className="text-xs text-destructive">{errors.carrier.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="label">Label (optional)</Label>
            <Input
              id="label"
              placeholder="e.g., New keyboard"
              {...register("label")}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">Notes (optional)</Label>
            <Textarea
              id="notes"
              placeholder="e.g., Leave at back door"
              rows={2}
              {...register("notes")}
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="private"
              {...register("private")}
              className="rounded border-input"
            />
            <Label htmlFor="private" className="text-sm font-normal">
              Private (hide details from other household members)
            </Label>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "Adding..." : "Add Package"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
