"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
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
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  User,
  getUserRoles,
  addUserRole,
  removeUserRole,
  updateUser,
} from "@/lib/api/users";
import { toast } from "sonner";

interface UserEditModalProps {
  user: User | null;
  householdName?: string;
  open: boolean;
  onClose: () => void;
  onSave: () => void;
}

// Valid roles from backend
const VALID_ROLES = ["admin", "user", "household_admin", "device_manager"];

// Role display names
const ROLE_LABELS: Record<string, string> = {
  admin: "Administrator",
  user: "User",
  household_admin: "Household Admin",
  device_manager: "Device Manager",
};

export function UserEditModal({
  user,
  householdName,
  open,
  onClose,
  onSave,
}: UserEditModalProps) {
  const [displayName, setDisplayName] = useState("");
  const [originalDisplayName, setOriginalDisplayName] = useState("");
  const [roles, setRoles] = useState<Set<string>>(new Set());
  const [originalRoles, setOriginalRoles] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // Fetch user roles and set display name when modal opens
  useEffect(() => {
    if (user && open) {
      setDisplayName(user.displayName);
      setOriginalDisplayName(user.displayName);
      fetchRoles();
    }
  }, [user, open]);

  const fetchRoles = async () => {
    if (!user) return;

    try {
      setLoading(true);
      const userRoles = await getUserRoles(user.id);
      const rolesSet = new Set(userRoles);
      setRoles(rolesSet);
      setOriginalRoles(new Set(rolesSet));
    } catch (error) {
      console.error("Failed to fetch roles:", error);
      toast.error("Failed to load user roles");
    } finally {
      setLoading(false);
    }
  };

  const handleRoleToggle = (role: string, checked: boolean) => {
    const newRoles = new Set(roles);
    if (checked) {
      newRoles.add(role);
    } else {
      newRoles.delete(role);
    }
    setRoles(newRoles);
  };

  const hasChanges = () => {
    // Check display name change
    if (displayName.trim() !== originalDisplayName) return true;

    // Check role changes
    if (roles.size !== originalRoles.size) return true;
    for (const role of roles) {
      if (!originalRoles.has(role)) return true;
    }
    return false;
  };

  const isValid = () => {
    return displayName.trim().length > 0;
  };

  const handleSave = async () => {
    if (!user || !isValid()) return;

    try {
      setSaving(true);

      const operations: Promise<unknown>[] = [];

      // Update display name if changed
      if (displayName.trim() !== originalDisplayName) {
        operations.push(updateUser(user.id, displayName.trim()));
      }

      // Determine which roles were added and removed
      const added = Array.from(roles).filter((r) => !originalRoles.has(r));
      const removed = Array.from(originalRoles).filter((r) => !roles.has(r));

      // Execute role operations
      operations.push(
        ...added.map((role) => addUserRole(user.id, role)),
        ...removed.map((role) => removeUserRole(user.id, role))
      );

      await Promise.all(operations);

      toast.success("User updated successfully");
      onSave();
    } catch (error) {
      console.error("Failed to save user:", error);
      toast.error("Failed to update user");
    } finally {
      setSaving(false);
    }
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

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit User</DialogTitle>
          <DialogDescription>
            Update user information and manage roles
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* User Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">User Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {/* Editable Display Name */}
              <div className="space-y-2">
                <Label htmlFor="displayName">
                  Display Name <span className="text-red-600">*</span>
                </Label>
                <Input
                  id="displayName"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  disabled={saving}
                  placeholder="Enter display name"
                />
                {displayName.trim().length === 0 && displayName.length > 0 && (
                  <p className="text-sm text-red-600">
                    Display name cannot be empty
                  </p>
                )}
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Email
                </div>
                <div className="col-span-2 text-sm">{user.email}</div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Household
                </div>
                <div className="col-span-2 text-sm">
                  {householdName && user.householdId ? (
                    <Link
                      href={`/households?householdId=${user.householdId}`}
                      className="text-blue-600 hover:underline dark:text-blue-400"
                    >
                      {householdName}
                    </Link>
                  ) : (
                    "—"
                  )}
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Provider
                </div>
                <div className="col-span-2 text-sm capitalize">
                  {user.provider}
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Created
                </div>
                <div className="col-span-2 text-sm">
                  {formatDate(user.createdAt)}
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Updated
                </div>
                <div className="col-span-2 text-sm">
                  {formatDate(user.updatedAt)}
                </div>
              </div>
              <Separator />
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  ID
                </div>
                <div className="col-span-2 text-xs text-neutral-500 dark:text-neutral-400 font-mono">
                  {user.id}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Role Management */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Roles</CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="space-y-2">
                  {VALID_ROLES.map((role) => (
                    <div key={role} className="flex items-center space-x-2">
                      <div className="h-4 w-4 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
                      <div className="h-4 w-32 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse" />
                    </div>
                  ))}
                </div>
              ) : (
                <div className="space-y-3">
                  {VALID_ROLES.map((role) => (
                    <div key={role} className="flex items-center space-x-2">
                      <Checkbox
                        id={`role-${role}`}
                        checked={roles.has(role)}
                        onCheckedChange={(checked) =>
                          handleRoleToggle(role, checked as boolean)
                        }
                        disabled={saving}
                      />
                      <Label
                        htmlFor={`role-${role}`}
                        className="text-sm font-normal cursor-pointer"
                      >
                        {ROLE_LABELS[role] || role}
                      </Label>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={!hasChanges() || !isValid() || saving || loading}
          >
            {saving ? "Saving..." : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
