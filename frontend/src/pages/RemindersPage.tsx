import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useReminders, useCreateReminder, useSnoozeReminder, useDismissReminder, useDeleteReminder } from "@/lib/hooks/api/use-reminders";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Plus, Clock, BellOff, Trash2 } from "lucide-react";

const reminderSchema = z.object({
  title: z.string().min(1, "Title is required"),
  notes: z.string().optional(),
  scheduledFor: z.string().min(1, "Scheduled time is required"),
});

type ReminderForm = z.infer<typeof reminderSchema>;

export function RemindersPage() {
  const { data, isLoading } = useReminders();
  const createReminder = useCreateReminder();
  const snoozeReminder = useSnoozeReminder();
  const dismissReminder = useDismissReminder();
  const deleteReminder = useDeleteReminder();
  const [open, setOpen] = useState(false);

  const form = useForm<ReminderForm>({
    resolver: zodResolver(reminderSchema),
    defaultValues: { title: "", notes: "", scheduledFor: "" },
  });

  const reminders = data?.data ?? [];

  const onSubmit = async (values: ReminderForm) => {
    await createReminder.mutateAsync({
      title: values.title,
      notes: values.notes,
      scheduledFor: new Date(values.scheduledFor).toISOString(),
    });
    form.reset();
    setOpen(false);
  };

  if (isLoading) {
    return <div className="p-6"><div className="h-8 w-48 animate-pulse rounded bg-muted" /></div>;
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Reminders</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger>
            <Button size="sm"><Plus className="mr-2 h-4 w-4" />New Reminder</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Reminder</DialogTitle>
            </DialogHeader>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">Title</Label>
                <Input id="title" {...form.register("title")} />
                {form.formState.errors.title && (
                  <p className="text-sm text-destructive">{form.formState.errors.title.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="notes">Notes</Label>
                <Input id="notes" {...form.register("notes")} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="scheduledFor">Scheduled For</Label>
                <Input id="scheduledFor" type="datetime-local" {...form.register("scheduledFor")} />
              </div>
              <Button type="submit" className="w-full" disabled={createReminder.isPending}>
                {createReminder.isPending ? "Creating..." : "Create Reminder"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {reminders.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          No reminders yet. Create your first reminder to get started.
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
                      <Button variant="ghost" size="sm" onClick={() => snoozeReminder.mutate({ id: rem.id, minutes: 10 })}>
                        <Clock className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => dismissReminder.mutate(rem.id)}>
                        <BellOff className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                  <Button variant="ghost" size="sm" onClick={() => deleteReminder.mutate(rem.id)}>
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
