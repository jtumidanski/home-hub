import { useState, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { type ColumnDef } from "@tanstack/react-table";
import { toast } from "sonner";
import { useReminders, useSnoozeReminder, useDismissReminder, useDeleteReminder } from "@/lib/hooks/api/use-reminders";
import { useMemberMap, useHouseholdMembers } from "@/lib/hooks/api/use-household-members";
import { useAuth } from "@/components/providers/auth-provider";
import type { Member } from "@/types/models/member";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { type Reminder, isReminderDismissed, isReminderSnoozed } from "@/types/models/reminder";
import { useMobile } from "@/lib/hooks/use-mobile";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { ListFilterBar } from "@/components/common/list-filter-bar";
import { ReminderCard } from "@/components/features/reminders/reminder-card";
import { CreateReminderDialog } from "@/components/features/reminders/create-reminder-dialog";
import { DataTable } from "@/components/common/data-table";
import { ErrorCard } from "@/components/common/error-card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Clock, BellOff, Trash2 } from "lucide-react";

const REMINDER_STATUS_OPTIONS = [
  { value: "active", label: "Active" },
  { value: "dismissed", label: "Dismissed" },
  { value: "snoozed", label: "Snoozed" },
  { value: "upcoming", label: "Upcoming" },
];

function getReminderStatus(rem: Reminder): string {
  if (rem.attributes.active) return "active";
  if (isReminderDismissed(rem)) return "dismissed";
  if (isReminderSnoozed(rem)) return "snoozed";
  if (new Date(rem.attributes.scheduledFor) > new Date()) return "upcoming";
  return "inactive";
}

function resolveOwnerName(ownerUserId: string | null | undefined, memberMap: Map<string, string>): string {
  if (!ownerUserId) return "Everyone";
  return memberMap.get(ownerUserId) ?? "Former member";
}

export function RemindersPage() {
  const { data, isLoading, isError, refetch } = useReminders();
  const snoozeReminder = useSnoozeReminder();
  const dismissReminder = useDismissReminder();
  const deleteReminder = useDeleteReminder();
  const [open, setOpen] = useState(false);
  const isMobile = useMobile();
  const { user } = useAuth();
  const { data: membersData } = useHouseholdMembers();
  const members = (membersData?.data ?? []) as Member[];
  const memberMap = useMemberMap();
  const [searchParams] = useSearchParams();

  const allReminders = (data?.data ?? []) as Reminder[];

  const filteredReminders = useMemo(() => {
    let result = allReminders;
    const query = searchParams.get("q")?.toLowerCase();
    const status = searchParams.get("status");
    const owner = searchParams.get("owner");

    if (query) {
      result = result.filter((r) => r.attributes.title.toLowerCase().includes(query));
    }
    if (status && status !== "all") {
      result = result.filter((r) => getReminderStatus(r) === status);
    }
    if (owner && owner !== "all") {
      if (owner === "everyone") {
        result = result.filter((r) => !r.attributes.ownerUserId);
      } else {
        result = result.filter((r) => r.attributes.ownerUserId === owner);
      }
    }
    return result;
  }, [allReminders, searchParams]);

  const hasActiveFilters = searchParams.has("q") || searchParams.has("status") || searchParams.has("owner");

  const handleSnooze = useCallback(async (id: string) => {
    try {
      await snoozeReminder.mutateAsync({ id, minutes: 10 });
      toast.success("Reminder snoozed for 10 minutes");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to snooze reminder").message);
    }
  }, [snoozeReminder]);

  const handleDismiss = useCallback(async (id: string) => {
    try {
      await dismissReminder.mutateAsync(id);
      toast.success("Reminder dismissed");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to dismiss reminder").message);
    }
  }, [dismissReminder]);

  const handleDelete = useCallback(async (id: string) => {
    try {
      await deleteReminder.mutateAsync(id);
      toast.success("Reminder deleted");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete reminder").message);
    }
  }, [deleteReminder]);

  const handleRefresh = useCallback(async () => {
    await refetch();
  }, [refetch]);

  const columns: ColumnDef<Reminder, unknown>[] = [
    {
      accessorKey: "attributes.title",
      header: "Title",
      cell: ({ row }) => (
        <div>
          <p className="font-medium">{row.original.attributes.title}</p>
          <p className="text-xs text-muted-foreground">
            {new Date(row.original.attributes.scheduledFor).toLocaleString()}
          </p>
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: ({ row }) => {
        const statusLabel = getReminderStatus(row.original);
        return (
          <Badge variant={row.original.attributes.active ? "default" : "secondary"}>
            {statusLabel}
          </Badge>
        );
      },
    },
    {
      id: "owner",
      header: "Owner",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {resolveOwnerName(row.original.attributes.ownerUserId, memberMap)}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const rem = row.original;
        return (
          <div className="flex items-center gap-1">
            {rem.attributes.active && (
              <>
                <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleSnooze(rem.id); }}>
                  <Clock className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleDismiss(rem.id); }}>
                  <BellOff className="h-4 w-4" />
                </Button>
              </>
            )}
            <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleDelete(rem.id); }}>
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        );
      },
    },
  ];

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4" role="status" aria-label="Loading">
        <DataTable columns={columns} data={[]} isLoading />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load reminders. Try refreshing the page." />
      </div>
    );
  }

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl md:text-2xl font-semibold">Reminders</h1>
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />New Reminder
          </Button>
        </div>

        <ListFilterBar statusOptions={REMINDER_STATUS_OPTIONS} />

        <CreateReminderDialog open={open} onOpenChange={setOpen} currentUserId={user?.id ?? ""} members={members} />

        {isMobile ? (
          filteredReminders.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <p className="text-muted-foreground">
                {hasActiveFilters ? "No items found." : "No reminders yet. Create your first reminder to get started."}
              </p>
              {hasActiveFilters && (
                <Button variant="link" size="sm" onClick={() => window.history.pushState({}, "", window.location.pathname)}>
                  Clear filters
                </Button>
              )}
            </div>
          ) : (
            <div className="space-y-3">
              {filteredReminders.map((reminder) => (
                <ReminderCard
                  key={reminder.id}
                  reminder={reminder}
                  ownerName={resolveOwnerName(reminder.attributes.ownerUserId, memberMap)}
                  onSnooze={handleSnooze}
                  onDismiss={handleDismiss}
                  onDelete={handleDelete}
                />
              ))}
            </div>
          )
        ) : (
          <DataTable
            columns={columns}
            data={filteredReminders}
            emptyMessage={hasActiveFilters ? "No items found." : "No reminders yet. Create your first reminder to get started."}
          />
        )}
        {allReminders.length === 0 && !hasActiveFilters && (
          <div className="flex justify-center">
            <Button variant="outline" onClick={() => setOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create First Reminder
            </Button>
          </div>
        )}
      </div>
    </PullToRefresh>
  );
}
