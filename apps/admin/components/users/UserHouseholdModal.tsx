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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { User, associateUserToHousehold } from "@/lib/api/users";
import { listHouseholds, Household } from "@/lib/api/households";
import { UserDisassociateDialog } from "@/components/users/UserDisassociateDialog";
import { toast } from "sonner";
import { AlertCircle } from "lucide-react";

interface UserHouseholdModalProps {
  user: User | null;
  currentHouseholdName?: string;
  open: boolean;
  onClose: () => void;
  onSave: () => void;
}

export function UserHouseholdModal({
  user,
  currentHouseholdName,
  open,
  onClose,
  onSave,
}: UserHouseholdModalProps) {
  const [action, setAction] = useState<"associate" | "disassociate">(
    "associate"
  );
  const [selectedHouseholdId, setSelectedHouseholdId] = useState<string>("");
  const [households, setHouseholds] = useState<Household[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [disassociateDialogOpen, setDisassociateDialogOpen] = useState(false);

  // Fetch households when modal opens
  useEffect(() => {
    if (open) {
      fetchHouseholds();
      // Reset state
      setAction(user?.householdId ? "disassociate" : "associate");
      setSelectedHouseholdId("");
      setSearchQuery("");
    }
  }, [open, user]);

  const fetchHouseholds = async () => {
    try {
      setLoading(true);
      const data = await listHouseholds();
      setHouseholds(data);
    } catch (error) {
      console.error("Failed to fetch households:", error);
      toast.error("Failed to load households");
    } finally {
      setLoading(false);
    }
  };

  const filteredHouseholds = households.filter((h) =>
    h.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleSave = async () => {
    if (!user) return;

    if (action === "associate") {
      if (!selectedHouseholdId) {
        toast.error("Please select a household");
        return;
      }

      try {
        setSaving(true);
        await associateUserToHousehold(user.id, selectedHouseholdId);
        toast.success("User associated with household");
        onSave();
      } catch (error) {
        console.error("Failed to associate user to household:", error);
        toast.error("Failed to associate user to household");
      } finally {
        setSaving(false);
      }
    } else {
      // Disassociate - show confirmation dialog
      setDisassociateDialogOpen(true);
    }
  };

  const handleDisassociateConfirmed = () => {
    setDisassociateDialogOpen(false);
    onSave();
  };

  const canSave = () => {
    if (action === "associate") {
      return selectedHouseholdId.length > 0;
    }
    return true; // Disassociate is always valid if user has a household
  };

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Household Association</DialogTitle>
          <DialogDescription>
            Associate or disassociate user from a household
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Current Household Status */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Current Household</CardTitle>
            </CardHeader>
            <CardContent>
              {user.householdId && currentHouseholdName ? (
                <div className="space-y-1">
                  <p className="text-sm font-medium">{currentHouseholdName}</p>
                  <p className="text-xs text-neutral-500 dark:text-neutral-400 font-mono">
                    {user.householdId}
                  </p>
                </div>
              ) : (
                <div className="flex items-center gap-2 text-neutral-500 dark:text-neutral-400">
                  <AlertCircle className="h-4 w-4" />
                  <p className="text-sm">No household assigned</p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Action Selector */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Action</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Radio Buttons */}
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <input
                    type="radio"
                    id="action-associate"
                    name="action"
                    checked={action === "associate"}
                    onChange={() => setAction("associate")}
                    disabled={saving}
                    className="h-4 w-4"
                  />
                  <Label
                    htmlFor="action-associate"
                    className="text-sm font-normal cursor-pointer"
                  >
                    Associate to Household
                  </Label>
                </div>
                <div className="flex items-center space-x-2">
                  <input
                    type="radio"
                    id="action-disassociate"
                    name="action"
                    checked={action === "disassociate"}
                    onChange={() => setAction("disassociate")}
                    disabled={saving || !user.householdId}
                    className="h-4 w-4"
                  />
                  <Label
                    htmlFor="action-disassociate"
                    className={`text-sm font-normal cursor-pointer ${!user.householdId ? "text-neutral-400 dark:text-neutral-600" : ""}`}
                  >
                    Disassociate from Household
                    {!user.householdId && " (user has no household)"}
                  </Label>
                </div>
              </div>

              {/* Household Selector (shown only for associate) */}
              {action === "associate" && (
                <div className="space-y-3 pt-2">
                  <Label htmlFor="household-search">
                    Select Household <span className="text-red-600">*</span>
                  </Label>

                  {/* Search Input */}
                  <Input
                    id="household-search"
                    placeholder="Search households..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    disabled={loading || saving}
                  />

                  {/* Household List */}
                  <div className="border rounded-md max-h-60 overflow-y-auto">
                    {loading ? (
                      <div className="p-4 space-y-2">
                        {[1, 2, 3].map((i) => (
                          <div
                            key={i}
                            className="h-10 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse"
                          />
                        ))}
                      </div>
                    ) : filteredHouseholds.length === 0 ? (
                      <div className="p-4 text-center text-sm text-neutral-500 dark:text-neutral-400">
                        {searchQuery
                          ? "No households found matching your search"
                          : "No households available"}
                      </div>
                    ) : (
                      <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
                        {filteredHouseholds.map((household) => (
                          <button
                            key={household.id}
                            onClick={() => setSelectedHouseholdId(household.id)}
                            disabled={saving}
                            className={`w-full text-left p-3 hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors ${
                              selectedHouseholdId === household.id
                                ? "bg-blue-50 dark:bg-blue-950 border-l-4 border-blue-600"
                                : ""
                            }`}
                          >
                            <p className="text-sm font-medium">
                              {household.name}
                            </p>
                            <p className="text-xs text-neutral-500 dark:text-neutral-400 font-mono">
                              {household.id}
                            </p>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>

                  {selectedHouseholdId && (
                    <p className="text-xs text-neutral-600 dark:text-neutral-400">
                      Selected:{" "}
                      {
                        households.find((h) => h.id === selectedHouseholdId)
                          ?.name
                      }
                    </p>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={!canSave() || saving}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>

      {/* Disassociate Confirmation Dialog */}
      <UserDisassociateDialog
        user={user}
        householdName={currentHouseholdName}
        open={disassociateDialogOpen}
        onClose={() => setDisassociateDialogOpen(false)}
        onDisassociated={handleDisassociateConfirmed}
      />
    </Dialog>
  );
}
