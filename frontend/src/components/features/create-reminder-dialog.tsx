import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useCreateReminder } from "@/lib/hooks/api/use-reminders";
import { createReminderSchema, type CreateReminderFormData, createReminderDefaults } from "@/lib/schemas/reminder.schema";
import { getErrorMessage } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Loader2 } from "lucide-react";

interface CreateReminderDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateReminderDialog({ open, onOpenChange }: CreateReminderDialogProps) {
  const createReminder = useCreateReminder();

  const form = useForm<CreateReminderFormData>({
    resolver: zodResolver(createReminderSchema),
    defaultValues: createReminderDefaults,
  });

  const handleOpenChange = (next: boolean) => {
    if (form.formState.isSubmitting) return;
    if (!next) form.reset(createReminderDefaults);
    onOpenChange(next);
  };

  const onSubmit = async (values: CreateReminderFormData) => {
    try {
      await createReminder.mutateAsync({
        title: values.title,
        notes: values.notes,
        scheduledFor: new Date(values.scheduledFor).toISOString(),
      });
      toast.success("Reminder created");
      form.reset(createReminderDefaults);
      onOpenChange(false);
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to create reminder"));
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Reminder</DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter reminder title" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="notes"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Notes</FormLabel>
                  <FormControl>
                    <Input placeholder="Optional notes" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="scheduledFor"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Scheduled For</FormLabel>
                  <FormControl>
                    <Input type="datetime-local" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Reminder
            </Button>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
