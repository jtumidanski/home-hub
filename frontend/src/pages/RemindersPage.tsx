import { useState } from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { toast } from "sonner";
import { useReminders, useSnoozeReminder, useDismissReminder, useDeleteReminder } from "@/lib/hooks/api/use-reminders";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { type Reminder, isReminderDismissed, isReminderSnoozed } from "@/types/models/reminder";
import { CreateReminderDialog } from "@/components/features/reminders/create-reminder-dialog";
import { DataTable } from "@/components/common/data-table";
import { ErrorCard } from "@/components/common/error-card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Clock, BellOff, Trash2 } from "lucide-react";

export function RemindersPage() {
  const { data, isLoading, isError } = useReminders();
  const snoozeReminder = useSnoozeReminder();
  const dismissReminder = useDismissReminder();
  const deleteReminder = useDeleteReminder();
  const [open, setOpen] = useState(false);

  const reminders = (data?.data ?? []) as Reminder[];

  const handleSnooze = async (id: string) => {
    try {
      await snoozeReminder.mutateAsync({ id, minutes: 10 });
      toast.success("Reminder snoozed for 10 minutes");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to snooze reminder").message);
    }
  };

  const handleDismiss = async (id: string) => {
    try {
      await dismissReminder.mutateAsync(id);
      toast.success("Reminder dismissed");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to dismiss reminder").message);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteReminder.mutateAsync(id);
      toast.success("Reminder deleted");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete reminder").message);
    }
  };

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
        const rem = row.original;
        return (
          <Badge variant={rem.attributes.active ? "default" : "secondary"}>
            {rem.attributes.active ? "active" : isReminderDismissed(rem) ? "dismissed" : isReminderSnoozed(rem) ? "snoozed" : "inactive"}
          </Badge>
        );
      },
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
      <div className="p-6 space-y-4" role="status" aria-label="Loading">
        <DataTable columns={columns} data={[]} isLoading />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-6">
        <ErrorCard message="Failed to load reminders. Try refreshing the page." />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Reminders</h1>
        <Button size="sm" onClick={() => setOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />New Reminder
        </Button>
      </div>

      <CreateReminderDialog open={open} onOpenChange={setOpen} />

      <DataTable
        columns={columns}
        data={reminders}
        emptyMessage="No reminders yet. Create your first reminder to get started."
      />
      {reminders.length === 0 && (
        <div className="flex justify-center">
          <Button variant="outline" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create First Reminder
          </Button>
        </div>
      )}
    </div>
  );
}
