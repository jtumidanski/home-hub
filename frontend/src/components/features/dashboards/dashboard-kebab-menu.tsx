import { useState } from "react";
import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { MoreVertical, Pencil, Star, Trash2, ArrowUpCircle, Copy } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  useCopyDashboardToMine,
  useDeleteDashboard,
  usePromoteDashboard,
  useUpdateDashboard,
} from "@/lib/hooks/api/use-dashboards";
import {
  useHouseholdPreferences,
  useUpdateHouseholdPreferences,
} from "@/lib/hooks/api/use-household-preferences";
import { dashboardNameSchema } from "./new-dashboard-modal";
import type { Dashboard } from "@/types/models/dashboard";

const renameSchema = z.object({ name: dashboardNameSchema });
type RenameFormData = z.infer<typeof renameSchema>;

interface DashboardKebabMenuProps {
  dashboard: Dashboard;
  isDefault: boolean;
}

export function DashboardKebabMenu({ dashboard, isDefault }: DashboardKebabMenuProps) {
  const navigate = useNavigate();
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const updateMutation = useUpdateDashboard();
  const deleteMutation = useDeleteDashboard();
  const promoteMutation = usePromoteDashboard();
  const copyToMineMutation = useCopyDashboardToMine();
  const { data: prefsData } = useHouseholdPreferences();
  const updatePrefsMutation = useUpdateHouseholdPreferences();

  const isUserScope = dashboard.attributes.scope === "user";
  const isHouseholdScope = dashboard.attributes.scope === "household";

  const handleSetDefault = () => {
    const prefs = prefsData?.data?.[0];
    if (!prefs) return;
    updatePrefsMutation.mutate({
      id: prefs.id,
      attrs: { defaultDashboardId: dashboard.id },
    });
  };

  const handlePromote = () => {
    promoteMutation.mutate(dashboard.id, {
      onSuccess: (resp) => {
        navigate(`/app/dashboards/${resp.data.id}`);
      },
    });
  };

  const handleCopyToMine = () => {
    copyToMineMutation.mutate(dashboard.id, {
      onSuccess: (resp) => {
        navigate(`/app/dashboards/${resp.data.id}`);
      },
    });
  };

  const handleDelete = () => {
    deleteMutation.mutate(dashboard.id, {
      onSuccess: () => {
        setDeleteOpen(false);
      },
    });
  };

  return (
    <>
      <MenuPrimitive.Root>
        <MenuPrimitive.Trigger
          aria-label={`Dashboard actions for ${dashboard.attributes.name}`}
          className={cn(
            "flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground opacity-0 transition-opacity hover:bg-sidebar-accent/50 hover:text-sidebar-foreground focus-visible:opacity-100 group-hover/row:opacity-100 data-[popup-open]:opacity-100 outline-none",
          )}
        >
          <MoreVertical className="h-4 w-4" />
        </MenuPrimitive.Trigger>
        <MenuPrimitive.Portal>
          <MenuPrimitive.Positioner side="right" align="start" sideOffset={4} className="z-50">
            <MenuPrimitive.Popup
              className={cn(
                "min-w-48 rounded-lg bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/10",
                "origin-(--transform-origin) transition-[transform,scale,opacity] duration-100",
                "data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95",
                "data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95",
              )}
            >
              <MenuPrimitive.Item
                className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
                onClick={() => setRenameOpen(true)}
              >
                <Pencil className="h-4 w-4" />
                Rename
              </MenuPrimitive.Item>
              <MenuPrimitive.Item
                className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50"
                disabled={isDefault}
                onClick={handleSetDefault}
              >
                <Star className="h-4 w-4" />
                {isDefault ? "Already default" : "Set as my default"}
              </MenuPrimitive.Item>
              {isUserScope && (
                <MenuPrimitive.Item
                  className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
                  onClick={handlePromote}
                >
                  <ArrowUpCircle className="h-4 w-4" />
                  Promote to household
                </MenuPrimitive.Item>
              )}
              {isHouseholdScope && (
                <MenuPrimitive.Item
                  className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm outline-none select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
                  onClick={handleCopyToMine}
                >
                  <Copy className="h-4 w-4" />
                  Copy to mine
                </MenuPrimitive.Item>
              )}
              <MenuPrimitive.Item
                className="flex w-full cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm text-destructive outline-none select-none hover:bg-accent focus:bg-accent"
                onClick={() => setDeleteOpen(true)}
              >
                <Trash2 className="h-4 w-4" />
                Delete
              </MenuPrimitive.Item>
            </MenuPrimitive.Popup>
          </MenuPrimitive.Positioner>
        </MenuPrimitive.Portal>
      </MenuPrimitive.Root>

      <RenameDialog
        open={renameOpen}
        onOpenChange={setRenameOpen}
        dashboard={dashboard}
        onSubmit={(name) => {
          updateMutation.mutate(
            { id: dashboard.id, attrs: { name } },
            { onSuccess: () => setRenameOpen(false) },
          );
        }}
        isPending={updateMutation.isPending}
      />

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Delete dashboard?</DialogTitle>
            <DialogDescription>
              Delete &ldquo;{dashboard.attributes.name}&rdquo;? This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => setDeleteOpen(false)}>
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={handleDelete}
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

interface RenameDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dashboard: Dashboard;
  onSubmit: (name: string) => void;
  isPending: boolean;
}

function RenameDialog({ open, onOpenChange, dashboard, onSubmit, isPending }: RenameDialogProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<RenameFormData>({
    resolver: zodResolver(renameSchema),
    defaultValues: { name: dashboard.attributes.name },
  });

  const submit = (data: RenameFormData) => {
    onSubmit(data.name.trim());
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (v) {
          reset({ name: dashboard.attributes.name });
        }
        onOpenChange(v);
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Rename dashboard</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(submit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="rename-name">Name</Label>
            <Input id="rename-name" {...register("name")} />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name.message}</p>
            )}
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving..." : "Save"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
