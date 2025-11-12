"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { getHouseholdUsers, Household } from "@/lib/api/households";
import { User } from "@/lib/api/users";
import { UserDisassociateDialog } from "@/components/users/UserDisassociateDialog";
import { toast } from "sonner";
import { AlertCircle } from "lucide-react";

interface HouseholdDetailModalProps {
  household: Household | null;
  open: boolean;
  onClose: () => void;
  onUpdate: () => void;
}

export function HouseholdDetailModal({
  household,
  open,
  onClose,
  onUpdate,
}: HouseholdDetailModalProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [userToRemove, setUserToRemove] = useState<User | null>(null);
  const [disassociateDialogOpen, setDisassociateDialogOpen] = useState(false);

  // Fetch users when modal opens
  useEffect(() => {
    if (household && open) {
      fetchUsers();
    }
  }, [household, open]);

  const fetchUsers = async () => {
    if (!household) return;

    try {
      setLoading(true);
      const data = await getHouseholdUsers(household.id);
      setUsers(data);
    } catch (error) {
      console.error("Failed to fetch users:", error);
      toast.error("Failed to load users");
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveUser = (user: User) => {
    setUserToRemove(user);
    setDisassociateDialogOpen(true);
  };

  const handleDisassociateConfirmed = async () => {
    setDisassociateDialogOpen(false);
    setUserToRemove(null);
    await fetchUsers(); // Refresh users list
    onUpdate(); // Notify parent to refresh
  };

  const formatDate = (dateString: string): string => {
    if (!dateString) return "—";

    const date = new Date(dateString);

    if (isNaN(date.getTime())) {
      console.error("Invalid date string:", dateString);
      return "Invalid Date";
    }

    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (!household) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Household Details</DialogTitle>
          <DialogDescription>
            View household information and manage associated users
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Household Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Household Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Name
                </div>
                <div className="col-span-2 text-sm">{household.name}</div>
              </div>
              <Separator />
              {/* Location Information */}
              {(household.latitude !== undefined || household.longitude !== undefined || household.timezone) && (
                <>
                  {household.latitude !== undefined && (
                    <>
                      <div className="grid grid-cols-3 gap-2">
                        <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                          Latitude
                        </div>
                        <div className="col-span-2 text-sm font-mono">
                          {household.latitude.toFixed(6)}
                        </div>
                      </div>
                      <Separator />
                    </>
                  )}
                  {household.longitude !== undefined && (
                    <>
                      <div className="grid grid-cols-3 gap-2">
                        <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                          Longitude
                        </div>
                        <div className="col-span-2 text-sm font-mono">
                          {household.longitude.toFixed(6)}
                        </div>
                      </div>
                      <Separator />
                    </>
                  )}
                  {household.timezone && (
                    <>
                      <div className="grid grid-cols-3 gap-2">
                        <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                          Timezone
                        </div>
                        <div className="col-span-2 text-sm font-mono">
                          {household.timezone}
                        </div>
                      </div>
                      <Separator />
                    </>
                  )}
                </>
              )}
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Created
                </div>
                <div className="col-span-2 text-sm">
                  {formatDate(household.createdAt)}
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Updated
                </div>
                <div className="col-span-2 text-sm">
                  {formatDate(household.updatedAt)}
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  ID
                </div>
                <div className="col-span-2 text-xs text-neutral-500 dark:text-neutral-400 font-mono">
                  {household.id}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Users List */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Associated Users</CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="space-y-2">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="flex items-center justify-between">
                      <div className="space-y-1 flex-1">
                        <div className="h-4 w-32 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
                        <div className="h-3 w-48 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
                      </div>
                      <div className="h-8 w-20 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
                    </div>
                  ))}
                </div>
              ) : users.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 text-center">
                  <AlertCircle className="h-8 w-8 text-neutral-400 mb-2" />
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    No users in this household
                  </p>
                </div>
              ) : (
                <div className="space-y-3">
                  {users.map((user) => (
                    <div
                      key={user.id}
                      className="flex items-center justify-between p-3 rounded-lg border border-neutral-200 dark:border-neutral-800"
                    >
                      <div className="flex-1">
                        <p className="text-sm font-medium">{user.displayName}</p>
                        <p className="text-xs text-neutral-500 dark:text-neutral-400">
                          {user.email}
                        </p>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRemoveUser(user)}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                      >
                        Remove
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </DialogContent>

      {/* Disassociate Confirmation Dialog */}
      <UserDisassociateDialog
        user={userToRemove}
        householdName={household?.name}
        open={disassociateDialogOpen}
        onClose={() => {
          setDisassociateDialogOpen(false);
          setUserToRemove(null);
        }}
        onDisassociated={handleDisassociateConfirmed}
      />
    </Dialog>
  );
}
