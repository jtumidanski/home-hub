import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useCreateTask } from "@/lib/hooks/api/use-tasks";
import { createTaskSchema, type CreateTaskFormData, createTaskDefaults } from "@/lib/schemas/task.schema";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { useAuth } from "@/components/providers/auth-provider";
import { OwnerSelect } from "@/components/common/owner-select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Loader2 } from "lucide-react";

interface CreateTaskDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateTaskDialog({ open, onOpenChange }: CreateTaskDialogProps) {
  const createTask = useCreateTask();
  const { user } = useAuth();

  const defaults = { ...createTaskDefaults, ownerUserId: user?.id ?? "" };
  const form = useForm<CreateTaskFormData>({
    resolver: zodResolver(createTaskSchema),
    defaultValues: defaults,
  });

  const handleOpenChange = (next: boolean) => {
    if (form.formState.isSubmitting) return;
    if (!next) form.reset(defaults);
    onOpenChange(next);
  };

  const onSubmit = async (values: CreateTaskFormData) => {
    try {
      await createTask.mutateAsync({
        title: values.title,
        ...(values.notes != null ? { notes: values.notes } : {}),
        ...(values.dueOn ? { dueOn: values.dueOn } : {}),
        ownerUserId: values.ownerUserId || null,
      });
      toast.success("Task created");
      form.reset(defaults);
      onOpenChange(false);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create task").message);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Task</DialogTitle>
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
                    <Input placeholder="Enter task title" {...field} />
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
              name="dueOn"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Due Date</FormLabel>
                  <FormControl>
                    <Input type="date" aria-label="Due Date" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="ownerUserId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Owner</FormLabel>
                  <FormControl>
                    <OwnerSelect value={field.value ?? ""} onChange={field.onChange} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Task
            </Button>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
