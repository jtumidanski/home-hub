import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useCopyDashboardToMine,
  useCreateDashboard,
  useDashboards,
} from "@/lib/hooks/api/use-dashboards";
import type { Layout, WidgetInstance } from "@/lib/dashboard/schema";
import type { Dashboard } from "@/types/models/dashboard";

export const dashboardNameSchema = z.string().trim().min(1, "Name is required").max(80, "Max 80 characters");

export const newDashboardFormSchema = z.object({
  name: dashboardNameSchema,
  scope: z.enum(["household", "user"]),
  copyOf: z.string().optional(),
});

export type NewDashboardFormData = z.infer<typeof newDashboardFormSchema>;

const NONE_VALUE = "__none__";

function deepCloneLayoutWithFreshIds(layout: Layout): Layout {
  const cloned = JSON.parse(JSON.stringify(layout)) as Layout;
  cloned.widgets = cloned.widgets.map((w: WidgetInstance) => ({
    ...w,
    id: (crypto as Crypto).randomUUID(),
  }));
  return cloned;
}

interface NewDashboardModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function NewDashboardModal({ open, onOpenChange }: NewDashboardModalProps) {
  const navigate = useNavigate();
  const { data } = useDashboards();
  const createMutation = useCreateDashboard();
  const copyToMineMutation = useCopyDashboardToMine();

  const dashboards = data?.data ?? [];

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<NewDashboardFormData>({
    resolver: zodResolver(newDashboardFormSchema),
    defaultValues: {
      name: "",
      scope: "household",
      copyOf: NONE_VALUE,
    },
  });

  useEffect(() => {
    if (!open) {
      reset({ name: "", scope: "household", copyOf: NONE_VALUE });
    }
  }, [open, reset]);

  const scope = watch("scope");
  const copyOf = watch("copyOf");

  const onSubmit = async (form: NewDashboardFormData) => {
    const name = form.name.trim();
    const copyId = form.copyOf && form.copyOf !== NONE_VALUE ? form.copyOf : null;
    const source: Dashboard | undefined = copyId
      ? dashboards.find((d) => d.id === copyId)
      : undefined;

    try {
      if (source) {
        if (form.scope === "user" && source.attributes.scope === "household") {
          const resp = await copyToMineMutation.mutateAsync(source.id);
          onOpenChange(false);
          navigate(`/app/dashboards/${resp.data.id}/edit`);
          return;
        }
        const layout = deepCloneLayoutWithFreshIds(source.attributes.layout);
        const resp = await createMutation.mutateAsync({
          name,
          scope: form.scope,
          layout,
        });
        onOpenChange(false);
        navigate(`/app/dashboards/${resp.data.id}/edit`);
        return;
      }

      const resp = await createMutation.mutateAsync({
        name,
        scope: form.scope,
        layout: { version: 1, widgets: [] },
      });
      onOpenChange(false);
      navigate(`/app/dashboards/${resp.data.id}/edit`);
    } catch {
      // Errors handled by mutation hooks (toasts).
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New Dashboard</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="dashboard-name">Name</Label>
            <Input
              id="dashboard-name"
              placeholder="My dashboard"
              {...register("name")}
            />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label>Scope</Label>
            <RadioGroup
              value={scope}
              onValueChange={(v) => setValue("scope", v as "household" | "user")}
            >
              <RadioGroupItem value="household">Household</RadioGroupItem>
              <RadioGroupItem value="user">My Dashboards</RadioGroupItem>
            </RadioGroup>
          </div>

          <div className="space-y-2">
            <Label htmlFor="copyOf">Copy from (optional)</Label>
            <Select
              value={copyOf ?? NONE_VALUE}
              onValueChange={(v) => setValue("copyOf", v ?? NONE_VALUE)}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="None (start blank)" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={NONE_VALUE}>None (start blank)</SelectItem>
                {dashboards.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    {d.attributes.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || createMutation.isPending || copyToMineMutation.isPending}>
              {isSubmitting || createMutation.isPending || copyToMineMutation.isPending
                ? "Creating..."
                : "Create"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
