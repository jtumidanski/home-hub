import { useState } from "react";
import { toast } from "sonner";
import { useReminders, useSnoozeReminder, useDismissReminder, useDeleteReminder } from "@/lib/hooks/api/use-reminders";
import { getErrorMessage } from "@/lib/api/errors";
import { CreateReminderDialog } from "@/components/features/create-reminder-dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus, Clock, BellOff, Trash2 } from "lucide-react";

function RemindersPageSkeleton() {
  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-9 w-36" />
      </div>
      <div className="space-y-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    </div>
  );
}

export function RemindersPage() {
  const { data, isLoading, isError } = useReminders();
  const snoozeReminder = useSnoozeReminder();
  const dismissReminder = useDismissReminder();
  const deleteReminder = useDeleteReminder();
  const [open, setOpen] = useState(false);

  const reminders = data?.data ?? [];

  const handleSnooze = async (id: string) => {
    try {
      await snoozeReminder.mutateAsync({ id, minutes: 10 });
      toast.success("Reminder snoozed for 10 minutes");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to snooze reminder"));
    }
  };

  const handleDismiss = async (id: string) => {
    try {
      await dismissReminder.mutateAsync(id);
      toast.success("Reminder dismissed");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to dismiss reminder"));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteReminder.mutateAsync(id);
      toast.success("Reminder deleted");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to delete reminder"));
    }
  };

  if (isLoading) {
    return <RemindersPageSkeleton />;
  }

  if (isError) {
    return (
      <div className="p-6">
        <Card className="border-destructive">
          <CardContent className="py-3">
            <p className="text-sm text-destructive">
              Failed to load reminders. Try refreshing the page.
            </p>
          </CardContent>
        </Card>
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

      {reminders.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-muted-foreground">No reminders yet. Create your first reminder to get started.</p>
          <Button variant="outline" className="mt-4" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create First Reminder
          </Button>
        </div>
      ) : (
        <div className="space-y-2">
          {reminders.map((rem) => (
            <Card key={rem.id}>
              <CardContent className="flex items-center justify-between py-3">
                <div>
                  <p className="font-medium">{rem.attributes.title}</p>
                  <p className="text-xs text-muted-foreground">
                    {new Date(rem.attributes.scheduledFor).toLocaleString()}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={rem.attributes.active ? "default" : "secondary"}>
                    {rem.attributes.active ? "active" : rem.attributes.lastDismissedAt ? "dismissed" : "snoozed"}
                  </Badge>
                  {rem.attributes.active && (
                    <>
                      <Button variant="ghost" size="sm" onClick={() => handleSnooze(rem.id)}>
                        <Clock className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => handleDismiss(rem.id)}>
                        <BellOff className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                  <Button variant="ghost" size="sm" onClick={() => handleDelete(rem.id)}>
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
