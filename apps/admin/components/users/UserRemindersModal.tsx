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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  AlertCircle,
  Bell,
  Plus,
  Trash2,
  Pencil,
  X,
  Filter,
  Clock,
} from "lucide-react";
import {
  Reminder,
  listReminders,
  createReminder,
  updateReminder,
  CreateReminderInput,
  UpdateReminderInput,
} from "@/lib/api/reminders";
import type { User } from "@/lib/api/users";
import { toast } from "sonner";
import { ReminderDeleteDialog } from "./ReminderDeleteDialog";

interface UserRemindersModalProps {
  user: User | null;
  open: boolean;
  onClose: () => void;
  onSave?: () => void;
}

type FilterStatus = "all" | "active" | "snoozed" | "dismissed";

export function UserRemindersModal({
  user,
  open,
  onClose,
  onSave,
}: UserRemindersModalProps) {
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [filterStatus, setFilterStatus] = useState<FilterStatus>("all");
  const [filterFromDate, setFilterFromDate] = useState<string>("");
  const [filterToDate, setFilterToDate] = useState<string>("");
  const [showFilters, setShowFilters] = useState(false);

  // Create form
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createForm, setCreateForm] = useState<CreateReminderInput>({
    userId: "",
    name: "",
    description: "",
    remindAt: new Date(Date.now() + 3600000).toISOString(), // Default: 1 hour from now
  });
  const [creating, setCreating] = useState(false);

  // Edit form
  const [editingReminderId, setEditingReminderId] = useState<string | null>(
    null
  );
  const [editForm, setEditForm] = useState<UpdateReminderInput>({});
  const [updating, setUpdating] = useState(false);

  // Delete dialog
  const [reminderToDelete, setReminderToDelete] = useState<Reminder | null>(
    null
  );
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  // Fetch reminders when modal opens
  useEffect(() => {
    if (user && open) {
      fetchReminders();
      setCreateForm((prev) => ({
        ...prev,
        userId: user.id,
        remindAt: new Date(Date.now() + 3600000).toISOString(),
      }));
    }
  }, [user, open]);

  const fetchReminders = async () => {
    if (!user) return;

    try {
      setLoading(true);
      setError(null);

      const fetchedReminders = await listReminders(user.id);
      setReminders(fetchedReminders);
    } catch (err) {
      console.error("Failed to fetch reminders:", err);
      setError(
        err instanceof Error ? err.message : "Failed to load reminders"
      );
      toast.error("Failed to load reminders");
    } finally {
      setLoading(false);
    }
  };

  // Convert ISO string to datetime-local format for input
  const toDatetimeLocal = (isoString: string): string => {
    return new Date(isoString).toISOString().slice(0, 16);
  };

  // Convert datetime-local format back to ISO string
  const fromDatetimeLocal = (localString: string): string => {
    return new Date(localString).toISOString();
  };

  const handleCreate = async () => {
    if (!createForm.name.trim()) {
      toast.error("Reminder name is required");
      return;
    }

    try {
      setCreating(true);
      const newReminder = await createReminder(createForm);
      setReminders((prev) => [...prev, newReminder]);

      // Reset form
      setCreateForm({
        userId: user!.id,
        name: "",
        description: "",
        remindAt: new Date(Date.now() + 3600000).toISOString(),
      });
      setShowCreateForm(false);
      toast.success("Reminder created successfully");
    } catch (err) {
      console.error("Failed to create reminder:", err);
      toast.error("Failed to create reminder");
    } finally {
      setCreating(false);
    }
  };

  const handleStartEdit = (reminder: Reminder) => {
    setEditingReminderId(reminder.id);
    setEditForm({
      name: reminder.name,
      description: reminder.description,
      remindAt: reminder.remindAt,
      status: reminder.status,
    });
  };

  const handleCancelEdit = () => {
    setEditingReminderId(null);
    setEditForm({});
  };

  const handleUpdate = async (reminderId: string) => {
    try {
      setUpdating(true);
      const updatedReminder = await updateReminder(reminderId, editForm);
      setReminders((prev) =>
        prev.map((r) => (r.id === reminderId ? updatedReminder : r))
      );
      setEditingReminderId(null);
      setEditForm({});
      toast.success("Reminder updated successfully");
    } catch (err) {
      console.error("Failed to update reminder:", err);
      toast.error("Failed to update reminder");
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = (reminder: Reminder) => {
    setReminderToDelete(reminder);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirmed = () => {
    setDeleteDialogOpen(false);
    if (reminderToDelete) {
      setReminders((prev) => prev.filter((r) => r.id !== reminderToDelete.id));
    }
    setReminderToDelete(null);
  };

  // Apply filters
  const filteredReminders = reminders.filter((reminder) => {
    if (filterStatus !== "all" && reminder.status !== filterStatus)
      return false;

    if (filterFromDate) {
      const fromDate = new Date(filterFromDate).getTime();
      const remindDate = new Date(reminder.remindAt).getTime();
      if (remindDate < fromDate) return false;
    }

    if (filterToDate) {
      const toDate = new Date(filterToDate).getTime();
      const remindDate = new Date(reminder.remindAt).getTime();
      if (remindDate > toDate) return false;
    }

    return true;
  });

  // Sort reminders: active first, then snoozed, then dismissed, then by remindAt (soonest first), then by createdAt
  const sortedReminders = [...filteredReminders].sort((a, b) => {
    // Sort by status priority
    const statusOrder = { active: 0, snoozed: 1, dismissed: 2 };
    if (a.status !== b.status) {
      return statusOrder[a.status] - statusOrder[b.status];
    }

    // Then by remindAt (soonest first)
    if (a.remindAt !== b.remindAt) {
      return a.remindAt.localeCompare(b.remindAt);
    }

    // Then by createdAt (oldest first)
    return a.createdAt.localeCompare(b.createdAt);
  });

  // Format datetime for display
  const formatRemindAt = (dateString: string): string => {
    if (!dateString) return "—";

    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return "Invalid Date";
    }

    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffHours = diffMs / (1000 * 60 * 60);

    // Show relative time if within 48 hours
    if (Math.abs(diffHours) < 48) {
      if (diffMs < 0) {
        const absDiffHours = Math.abs(diffHours);
        if (absDiffHours < 1) {
          return "Just now";
        } else if (absDiffHours < 24) {
          return `${Math.floor(absDiffHours)}h ago`;
        } else {
          return `${Math.floor(absDiffHours / 24)}d ago`;
        }
      } else {
        if (diffHours < 1) {
          return `In ${Math.ceil(diffHours * 60)}m`;
        } else if (diffHours < 24) {
          return `In ${Math.ceil(diffHours)}h`;
        } else {
          return `In ${Math.ceil(diffHours / 24)}d`;
        }
      }
    }

    // Otherwise absolute time
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  };

  // Get status badge variant
  const getStatusBadge = (status: Reminder["status"]) => {
    const variants = {
      active: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
      snoozed:
        "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
      dismissed: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
    };

    return (
      <Badge className={variants[status]} variant="secondary">
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    );
  };

  if (!user) return null;

  return (
    <>
      <Dialog open={open} onOpenChange={onClose}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Manage Reminders - {user.displayName}</DialogTitle>
            <DialogDescription>
              View and manage reminders for this user
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {/* Action Bar */}
            <div className="flex items-center justify-between gap-2">
              <Button
                onClick={() => setShowCreateForm(!showCreateForm)}
                size="sm"
                variant={showCreateForm ? "secondary" : "default"}
              >
                {showCreateForm ? (
                  <>
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </>
                ) : (
                  <>
                    <Plus className="h-4 w-4 mr-2" />
                    New Reminder
                  </>
                )}
              </Button>

              <Button
                onClick={() => setShowFilters(!showFilters)}
                size="sm"
                variant="outline"
              >
                <Filter className="h-4 w-4 mr-2" />
                Filters
              </Button>
            </div>

            {/* Create Form */}
            {showCreateForm && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Create New Reminder</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="create-name">Name *</Label>
                    <Input
                      id="create-name"
                      value={createForm.name}
                      onChange={(e) =>
                        setCreateForm((prev) => ({
                          ...prev,
                          name: e.target.value,
                        }))
                      }
                      placeholder="Reminder name"
                      maxLength={200}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="create-description">Description</Label>
                    <Input
                      id="create-description"
                      value={createForm.description}
                      onChange={(e) =>
                        setCreateForm((prev) => ({
                          ...prev,
                          description: e.target.value,
                        }))
                      }
                      placeholder="Optional description"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="create-remindAt">Remind At *</Label>
                    <Input
                      id="create-remindAt"
                      type="datetime-local"
                      value={toDatetimeLocal(createForm.remindAt)}
                      onChange={(e) =>
                        setCreateForm((prev) => ({
                          ...prev,
                          remindAt: fromDatetimeLocal(e.target.value),
                        }))
                      }
                    />
                  </div>

                  <div className="flex justify-end gap-2">
                    <Button
                      onClick={() => setShowCreateForm(false)}
                      variant="outline"
                      size="sm"
                    >
                      Cancel
                    </Button>
                    <Button
                      onClick={handleCreate}
                      disabled={creating || !createForm.name.trim()}
                      size="sm"
                    >
                      {creating ? "Creating..." : "Create Reminder"}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Filters */}
            {showFilters && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Filters</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-3 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="filter-status">Filter by Status</Label>
                      <select
                        id="filter-status"
                        value={filterStatus}
                        onChange={(e) =>
                          setFilterStatus(e.target.value as FilterStatus)
                        }
                        className="w-full h-10 px-3 py-2 text-sm border border-input bg-background rounded-md"
                      >
                        <option value="all">All Reminders</option>
                        <option value="active">Active</option>
                        <option value="snoozed">Snoozed</option>
                        <option value="dismissed">Dismissed</option>
                      </select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="filter-from">From Date</Label>
                      <Input
                        id="filter-from"
                        type="datetime-local"
                        value={filterFromDate}
                        onChange={(e) => setFilterFromDate(e.target.value)}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="filter-to">To Date</Label>
                      <Input
                        id="filter-to"
                        type="datetime-local"
                        value={filterToDate}
                        onChange={(e) => setFilterToDate(e.target.value)}
                      />
                    </div>
                  </div>

                  <Button
                    onClick={() => {
                      setFilterStatus("all");
                      setFilterFromDate("");
                      setFilterToDate("");
                    }}
                    variant="outline"
                    size="sm"
                  >
                    Clear Filters
                  </Button>
                </CardContent>
              </Card>
            )}

            {/* Reminder List */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">
                  Reminders ({filteredReminders.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <div className="space-y-2">
                    {[1, 2, 3].map((i) => (
                      <div
                        key={i}
                        className="h-20 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse"
                      />
                    ))}
                  </div>
                ) : error ? (
                  <div className="text-center py-8">
                    <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-2" />
                    <p className="text-sm text-red-600 dark:text-red-400">
                      {error}
                    </p>
                    <Button
                      onClick={fetchReminders}
                      variant="outline"
                      size="sm"
                      className="mt-4"
                    >
                      Retry
                    </Button>
                  </div>
                ) : sortedReminders.length === 0 ? (
                  <div className="text-center py-8">
                    <Bell className="h-12 w-12 text-neutral-400 mx-auto mb-2" />
                    <p className="text-sm text-neutral-500 dark:text-neutral-400">
                      {reminders.length === 0
                        ? "No reminders yet"
                        : "No reminders match filters"}
                    </p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {sortedReminders.map((reminder) => (
                      <div key={reminder.id}>
                        {editingReminderId === reminder.id ? (
                          // Edit mode
                          <div className="border border-neutral-200 dark:border-neutral-700 rounded-lg p-4 space-y-3">
                            <div className="space-y-2">
                              <Label htmlFor={`edit-name-${reminder.id}`}>
                                Name
                              </Label>
                              <Input
                                id={`edit-name-${reminder.id}`}
                                value={editForm.name || ""}
                                onChange={(e) =>
                                  setEditForm((prev) => ({
                                    ...prev,
                                    name: e.target.value,
                                  }))
                                }
                                maxLength={200}
                              />
                            </div>

                            <div className="space-y-2">
                              <Label
                                htmlFor={`edit-description-${reminder.id}`}
                              >
                                Description
                              </Label>
                              <Input
                                id={`edit-description-${reminder.id}`}
                                value={editForm.description || ""}
                                onChange={(e) =>
                                  setEditForm((prev) => ({
                                    ...prev,
                                    description: e.target.value,
                                  }))
                                }
                              />
                            </div>

                            <div className="grid grid-cols-2 gap-4">
                              <div className="space-y-2">
                                <Label htmlFor={`edit-remindAt-${reminder.id}`}>
                                  Remind At
                                </Label>
                                <Input
                                  id={`edit-remindAt-${reminder.id}`}
                                  type="datetime-local"
                                  value={
                                    editForm.remindAt
                                      ? toDatetimeLocal(editForm.remindAt)
                                      : ""
                                  }
                                  onChange={(e) =>
                                    setEditForm((prev) => ({
                                      ...prev,
                                      remindAt: fromDatetimeLocal(
                                        e.target.value
                                      ),
                                    }))
                                  }
                                />
                              </div>

                              <div className="space-y-2">
                                <Label htmlFor={`edit-status-${reminder.id}`}>
                                  Status
                                </Label>
                                <select
                                  id={`edit-status-${reminder.id}`}
                                  value={editForm.status || ""}
                                  onChange={(e) =>
                                    setEditForm((prev) => ({
                                      ...prev,
                                      status: e.target.value as
                                        | "active"
                                        | "snoozed"
                                        | "dismissed",
                                    }))
                                  }
                                  className="w-full h-10 px-3 py-2 text-sm border border-input bg-background rounded-md"
                                >
                                  <option value="active">Active</option>
                                  <option value="snoozed">Snoozed</option>
                                  <option value="dismissed">Dismissed</option>
                                </select>
                              </div>
                            </div>

                            <div className="flex justify-end gap-2">
                              <Button
                                onClick={handleCancelEdit}
                                variant="outline"
                                size="sm"
                              >
                                Cancel
                              </Button>
                              <Button
                                onClick={() => handleUpdate(reminder.id)}
                                disabled={updating}
                                size="sm"
                              >
                                {updating ? "Saving..." : "Save"}
                              </Button>
                            </div>
                          </div>
                        ) : (
                          // View mode
                          <div className="p-3 border border-neutral-200 dark:border-neutral-700 rounded-lg hover:bg-neutral-50 dark:hover:bg-neutral-800/50 transition-colors">
                            <div className="flex items-start justify-between gap-3">
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 mb-1">
                                  {getStatusBadge(reminder.status)}
                                  {reminder.snoozeCount > 0 && (
                                    <Badge variant="outline" className="text-xs">
                                      <Clock className="h-3 w-3 mr-1" />
                                      Snoozed {reminder.snoozeCount}x
                                    </Badge>
                                  )}
                                </div>

                                <p
                                  className={`text-sm font-medium ${
                                    reminder.status === "dismissed"
                                      ? "line-through text-neutral-500 dark:text-neutral-400"
                                      : "text-neutral-900 dark:text-white"
                                  }`}
                                >
                                  {reminder.name}
                                </p>

                                {reminder.description && (
                                  <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">
                                    {reminder.description}
                                  </p>
                                )}

                                <p className="text-xs text-neutral-400 dark:text-neutral-500 mt-1">
                                  {formatRemindAt(reminder.remindAt)}
                                  {reminder.status === "active" &&
                                    new Date(reminder.remindAt) < new Date() && (
                                      <span className="ml-2 text-red-600 dark:text-red-400">
                                        • Overdue
                                      </span>
                                    )}
                                </p>
                              </div>

                              <div className="flex items-center gap-1">
                                <Button
                                  onClick={() => handleStartEdit(reminder)}
                                  variant="ghost"
                                  size="sm"
                                >
                                  <Pencil className="h-4 w-4" />
                                </Button>
                                <Button
                                  onClick={() => handleDelete(reminder)}
                                  variant="ghost"
                                  size="sm"
                                >
                                  <Trash2 className="h-4 w-4 text-red-600 dark:text-red-400" />
                                </Button>
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={onClose}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <ReminderDeleteDialog
        reminder={reminderToDelete}
        open={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false);
          setReminderToDelete(null);
        }}
        onDeleted={handleDeleteConfirmed}
      />
    </>
  );
}
