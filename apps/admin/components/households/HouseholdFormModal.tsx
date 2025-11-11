"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createHousehold, updateHousehold, Household } from "@/lib/api/households";
import { toast } from "sonner";

interface HouseholdFormModalProps {
  open: boolean;
  mode: "create" | "edit";
  household?: Household;
  onClose: () => void;
  onSave: () => void;
}

export function HouseholdFormModal({
  open,
  mode,
  household,
  onClose,
  onSave,
}: HouseholdFormModalProps) {
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);

  // Reset or populate form when modal opens
  useEffect(() => {
    if (open) {
      if (mode === "edit" && household) {
        setName(household.name);
      } else if (mode === "create") {
        setName("");
      }
    }
  }, [open, mode, household]);

  // Validation
  const isValid = name.trim().length > 0;
  const hasChanges =
    mode === "create" || (household && name !== household.name);

  const handleSave = async () => {
    if (!isValid) return;

    try {
      setSaving(true);

      if (mode === "create") {
        await createHousehold({ name });
        toast.success("Household created successfully");
      } else if (household) {
        await updateHousehold(household.id, { name });
        toast.success("Household updated successfully");
      }

      onSave();
    } catch (error) {
      console.error("Failed to save household:", error);
      toast.error(`Failed to ${mode} household`);
    } finally {
      setSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && isValid && hasChanges && !saving) {
      handleSave();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {mode === "create" ? "Create Household" : "Edit Household"}
          </DialogTitle>
          <DialogDescription>
            {mode === "create"
              ? "Create a new household"
              : "Update household information"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Name field */}
          <div className="space-y-2">
            <Label htmlFor="name">
              Name <span className="text-red-600">*</span>
            </Label>
            <Input
              id="name"
              placeholder="Enter household name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={saving}
              autoFocus
            />
            {name.trim().length === 0 && name.length > 0 && (
              <p className="text-sm text-red-600">Name cannot be empty</p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={!isValid || !hasChanges || saving}
          >
            {saving
              ? mode === "create"
                ? "Creating..."
                : "Saving..."
              : mode === "create"
                ? "Create"
                : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
