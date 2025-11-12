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
import { User, removeUserFromHousehold } from "@/lib/api/users";
import { toast } from "sonner";
import { AlertCircle } from "lucide-react";

interface UserDisassociateDialogProps {
  user: User | null;
  householdName?: string;
  open: boolean;
  onClose: () => void;
  onDisassociated: () => void;
}

export function UserDisassociateDialog({
  user,
  householdName,
  open,
  onClose,
  onDisassociated,
}: UserDisassociateDialogProps) {
  const [removing, setRemoving] = useState(false);

  const handleDisassociate = async () => {
    if (!user) return;

    try {
      setRemoving(true);
      await removeUserFromHousehold(user.id);
      toast.success("User removed from household");
      onDisassociated();
    } catch (error) {
      console.error("Failed to remove user from household:", error);
      toast.error("Failed to remove user from household");
    } finally {
      setRemoving(false);
    }
  };

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-yellow-100 dark:bg-yellow-950">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
            </div>
            <div>
              <DialogTitle>Remove User from Household</DialogTitle>
              <DialogDescription>
                Confirm household disassociation
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <p className="text-sm text-neutral-700 dark:text-neutral-300">
            Are you sure you want to remove{" "}
            <span className="font-semibold">{user.displayName}</span> from{" "}
            <span className="font-semibold">
              {householdName || "this household"}
            </span>
            ?
          </p>

          <div className="rounded-lg border border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-950 p-3">
            <p className="text-sm text-blue-800 dark:text-blue-200">
              The user will no longer have access to this household&apos;s data
              and settings. They can be re-associated later if needed.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={removing}>
            Cancel
          </Button>
          <Button onClick={handleDisassociate} disabled={removing}>
            {removing ? "Removing..." : "Remove from Household"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
