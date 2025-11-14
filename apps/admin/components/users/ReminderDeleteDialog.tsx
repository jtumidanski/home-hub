"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Reminder, deleteReminder } from "@/lib/api/reminders";
import { toast } from "sonner";
import { AlertCircle } from "lucide-react";

interface ReminderDeleteDialogProps {
  reminder: Reminder | null;
  open: boolean;
  onClose: () => void;
  onDeleted: () => void;
}

export function ReminderDeleteDialog({
  reminder,
  open,
  onClose,
  onDeleted,
}: ReminderDeleteDialogProps) {
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    if (!reminder) return;

    try {
      setDeleting(true);
      await deleteReminder(reminder.id);
      toast.success("Reminder deleted successfully");
      onDeleted();
    } catch (error) {
      console.error("Failed to delete reminder:", error);
      toast.error("Failed to delete reminder");
    } finally {
      setDeleting(false);
    }
  };

  if (!reminder) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-100 dark:bg-red-950">
              <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
            </div>
            <div>
              <DialogTitle>Delete Reminder</DialogTitle>
              <DialogDescription>
                Confirm reminder deletion
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <p className="text-sm text-neutral-700 dark:text-neutral-300">
            Are you sure you want to delete the reminder{" "}
            <span className="font-semibold">&quot;{reminder.name}&quot;</span>?
          </p>

          <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950 p-3">
            <p className="text-sm text-red-800 dark:text-red-200">
              This action cannot be undone. The reminder will be permanently
              deleted.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={deleting}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleting}
          >
            {deleting ? "Deleting..." : "Delete Reminder"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
