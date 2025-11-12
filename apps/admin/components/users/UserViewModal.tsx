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
import { User, getUserRoles } from "@/lib/api/users";
import { Badge } from "@/components/ui/badge";

interface UserViewModalProps {
  user: User | null;
  householdName?: string;
  open: boolean;
  onClose: () => void;
}

// Role display names
const ROLE_LABELS: Record<string, string> = {
  admin: "Administrator",
  user: "User",
  household_admin: "Household Admin",
  device_manager: "Device Manager",
};

export function UserViewModal({
  user,
  householdName,
  open,
  onClose,
}: UserViewModalProps) {
  const [roles, setRoles] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  // Fetch user roles when modal opens
  useEffect(() => {
    if (user && open) {
      fetchRoles();
    }
  }, [user, open]);

  const fetchRoles = async () => {
    if (!user) return;

    try {
      setLoading(true);
      const userRoles = await getUserRoles(user.id);
      setRoles(userRoles);
    } catch (error) {
      console.error("Failed to fetch roles:", error);
      setRoles([]);
    } finally {
      setLoading(false);
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
          <DialogTitle>User Details</DialogTitle>
          <DialogDescription>
            View user information and roles
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* User Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">User Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="grid grid-cols-3 gap-2">
                <div className="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                  Display Name
                </div>
                <div className="col-span-2 text-sm">{user.displayName}</div>
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

          {/* Roles Display */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Roles</CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex gap-2">
                  {[1, 2, 3].map((i) => (
                    <div
                      key={i}
                      className="h-6 w-24 bg-neutral-200 dark:bg-neutral-800 rounded animate-pulse"
                    />
                  ))}
                </div>
              ) : roles.length === 0 ? (
                <p className="text-sm text-neutral-500 dark:text-neutral-400">
                  No roles assigned
                </p>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {roles.map((role) => (
                    <Badge key={role} variant="secondary">
                      {ROLE_LABELS[role] || role}
                    </Badge>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <DialogFooter>
          <Button onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
