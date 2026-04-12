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
import { useUpdatePackage } from "@/lib/hooks/api/use-packages";
import { packageEditSchema, type PackageEditData } from "@/lib/schemas/package.schema";
import { CARRIER_LABELS } from "@/types/models/package";
import type { Package } from "@/types/models/package";

interface EditPackageDialogProps {
  pkg: Package;
  open: boolean;
  onClose: () => void;
}

export function EditPackageDialog({ pkg, open, onClose }: EditPackageDialogProps) {
  const updateMutation = useUpdatePackage();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors },
  } = useForm<PackageEditData>({
    resolver: zodResolver(packageEditSchema),
    defaultValues: {
      label: pkg.attributes.label ?? "",
      notes: pkg.attributes.notes ?? "",
      carrier: pkg.attributes.carrier as "usps" | "ups" | "fedex",
      private: pkg.attributes.private,
    },
  });

  useEffect(() => {
    if (open) {
      reset({
        label: pkg.attributes.label ?? "",
        notes: pkg.attributes.notes ?? "",
        carrier: pkg.attributes.carrier as "usps" | "ups" | "fedex",
        private: pkg.attributes.private,
      });
    }
  }, [open, pkg, reset]);

  const onSubmit = (data: PackageEditData) => {
    const attrs: { label?: string; notes?: string; carrier?: string; private?: boolean } = {};
    if (data.label !== undefined) attrs.label = data.label;
    if (data.notes !== undefined) attrs.notes = data.notes;
    if (data.carrier !== undefined) attrs.carrier = data.carrier;
    if (data.private !== undefined) attrs.private = data.private;

    updateMutation.mutate(
      { id: pkg.id, attrs },
      { onSuccess: () => onClose() }
    );
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Package</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-carrier">Carrier</Label>
            <Select
              // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
              value={watch("carrier")}
              onValueChange={(v) => setValue("carrier", v as "usps" | "ups" | "fedex")}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(CARRIER_LABELS).map(([value, label]) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-label">Label</Label>
            <Input id="edit-label" {...register("label")} />
            {errors.label && (
              <p className="text-xs text-destructive">{errors.label.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-notes">Notes</Label>
            <Textarea id="edit-notes" rows={2} {...register("notes")} />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="edit-private"
              {...register("private")}
              className="rounded border-input"
            />
            <Label htmlFor="edit-private" className="text-sm font-normal">
              Private
            </Label>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
