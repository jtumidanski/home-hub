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
import { deleteHousehold, Household } from "@/lib/api/households";
import { toast } from "sonner";
import { AlertCircle } from "lucide-react";
import { ApiError } from "@/lib/api/client";

interface HouseholdDeleteDialogProps {
  household: Household | null;
  open: boolean;
  onClose: () => void;
  onDeleted: () => void;
}

export function HouseholdDeleteDialog({
  household,
  open,
  onClose,
  onDeleted,
}: HouseholdDeleteDialogProps) {
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    if (!household) return;

    try {
      setDeleting(true);
      await deleteHousehold(household.id);
      toast.success("Household deleted successfully");
      onDeleted();
    } catch (error) {
      console.error("Failed to delete household:", error);

      // Check if it's a 409 Conflict (household has users)
      if (error instanceof ApiError && error.status === 409) {
        toast.error(
          "Cannot delete household with associated users. Please remove all users first."
        );
      } else {
        toast.error("Failed to delete household");
      }
    } finally {
      setDeleting(false);
    }
  };

  if (!household) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-100 dark:bg-red-950">
              <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
            </div>
            <div>
              <DialogTitle>Delete Household</DialogTitle>
              <DialogDescription>This action cannot be undone</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <p className="text-sm text-neutral-700 dark:text-neutral-300">
            Are you sure you want to delete{" "}
            <span className="font-semibold">{household.name}</span>?
          </p>

          <div className="rounded-lg border border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-950 p-3">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              <strong>Warning:</strong> This household must have no associated users
              before it can be deleted. If users are still associated, the delete
              operation will fail.
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
            {deleting ? "Deleting..." : "Delete Household"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
