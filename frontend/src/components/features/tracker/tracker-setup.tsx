import { useState } from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useTrackers, useDeleteTracker } from "@/lib/hooks/api/use-trackers";
import { CreateTrackerDialog } from "./create-tracker-dialog";
import { EditTrackerDialog } from "./edit-tracker-dialog";
import { DAY_LABELS, SCALE_TYPE_LABELS, type Tracker } from "@/types/models/tracker";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

const colorDot: Record<string, string> = {
  red: "bg-red-500", orange: "bg-orange-500", amber: "bg-amber-500", yellow: "bg-yellow-500",
  lime: "bg-lime-500", green: "bg-green-500", emerald: "bg-emerald-500", teal: "bg-teal-500",
  cyan: "bg-cyan-500", blue: "bg-blue-500", indigo: "bg-indigo-500", violet: "bg-violet-500",
  purple: "bg-purple-500", fuchsia: "bg-fuchsia-500", pink: "bg-pink-500", rose: "bg-rose-500",
};

function scheduleLabel(schedule: number[]): string {
  if (schedule.length === 0) return "Every day";
  return schedule.map((d) => DAY_LABELS[d]).join(", ");
}

export function TrackerSetup() {
  const { data, isLoading } = useTrackers();
  const deleteMutation = useDeleteTracker();
  const [createOpen, setCreateOpen] = useState(false);
  const [editTracker, setEditTracker] = useState<Tracker | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Tracker | null>(null);

  const trackers = data?.data ?? [];

  if (isLoading) {
    return <div className="space-y-2">{[1, 2, 3].map((i) => <Skeleton key={i} className="h-16 w-full" />)}</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">My Tracking Items</h2>
        <Button size="sm" onClick={() => setCreateOpen(true)}><Plus className="h-4 w-4 mr-1" /> Add</Button>
      </div>

      {trackers.length === 0 && (
        <p className="text-muted-foreground text-sm">No tracking items yet. Click "Add" to create your first one.</p>
      )}

      <div className="space-y-2">
        {trackers.map((t, i) => (
          <Card key={t.id}>
            <CardContent className="flex items-center justify-between py-3 px-4">
              <div className="flex items-center gap-3">
                <span className="text-sm text-muted-foreground w-5">{i + 1}.</span>
                <span className={cn("w-3 h-3 rounded-full", colorDot[t.attributes.color])} />
                <div>
                  <span className="font-medium">{t.attributes.name}</span>
                  <span className="text-xs text-muted-foreground ml-2">
                    {SCALE_TYPE_LABELS[t.attributes.scale_type]}
                    {t.attributes.scale_type === "range" && t.attributes.scale_config && ` ${t.attributes.scale_config.min}-${t.attributes.scale_config.max}`}
                  </span>
                  <span className="text-xs text-muted-foreground ml-2">{scheduleLabel(t.attributes.schedule)}</span>
                </div>
              </div>
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" onClick={() => setEditTracker(t)}><Pencil className="h-4 w-4" /></Button>
                <Button variant="ghost" size="icon" onClick={() => setDeleteTarget(t)}><Trash2 className="h-4 w-4" /></Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <CreateTrackerDialog open={createOpen} onClose={() => setCreateOpen(false)} />
      {editTracker && (
        <EditTrackerDialog open={!!editTracker} onClose={() => setEditTracker(null)} tracker={editTracker} />
      )}

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete tracking item</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Delete &ldquo;{deleteTarget?.attributes.name}&rdquo;? Historical entries for completed
            months will still be visible in past reports.
          </p>
          <div className="flex gap-2 justify-end">
            <Button variant="outline" size="sm" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              disabled={deleteMutation.isPending}
              onClick={() => {
                if (!deleteTarget) return;
                deleteMutation.mutate(deleteTarget.id, {
                  onSuccess: () => setDeleteTarget(null),
                });
              }}
            >
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
