import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from "@/lib/hooks/api/use-tasks";
import { createTaskSchema, type CreateTaskFormData, createTaskDefaults } from "@/lib/schemas/task.schema";
import { getErrorMessage } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus, Check, Trash2, Loader2 } from "lucide-react";

function TasksPageSkeleton() {
  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-9 w-28" />
      </div>
      <div className="space-y-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    </div>
  );
}

export function TasksPage() {
  const { data, isLoading } = useTasks();
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();
  const [open, setOpen] = useState(false);

  const form = useForm<CreateTaskFormData>({
    resolver: zodResolver(createTaskSchema),
    defaultValues: createTaskDefaults,
  });

  const tasks = data?.data ?? [];

  const onSubmit = async (values: CreateTaskFormData) => {
    try {
      await createTask.mutateAsync({
        title: values.title,
        notes: values.notes,
        dueOn: values.dueOn || undefined,
      });
      toast.success("Task created");
      form.reset(createTaskDefaults);
      setOpen(false);
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to create task"));
    }
  };

  const toggleComplete = async (id: string, currentStatus: string) => {
    try {
      await updateTask.mutateAsync({
        id,
        attrs: { status: currentStatus === "completed" ? "pending" : "completed" },
      });
      toast.success(currentStatus === "completed" ? "Task reopened" : "Task completed");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to update task"));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteTask.mutateAsync(id);
      toast.success("Task deleted");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to delete task"));
    }
  };

  if (isLoading) {
    return <TasksPageSkeleton />;
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Tasks</h1>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger>
            <Button size="sm"><Plus className="mr-2 h-4 w-4" />New Task</Button>
          </DialogTrigger>
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
                        <Input type="date" {...field} />
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
      </div>

      {tasks.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-muted-foreground">No tasks yet. Create your first task to get started.</p>
          <Button variant="outline" className="mt-4" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create First Task
          </Button>
        </div>
      ) : (
        <div className="space-y-2">
          {tasks.map((task) => (
            <Card key={task.id}>
              <CardContent className="flex items-center justify-between py-3">
                <div className="flex items-center gap-3">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => toggleComplete(task.id, task.attributes.status)}
                  >
                    <Check className={task.attributes.status === "completed" ? "h-4 w-4 text-green-500" : "h-4 w-4 text-muted-foreground"} />
                  </Button>
                  <div>
                    <p className={task.attributes.status === "completed" ? "font-medium line-through text-muted-foreground" : "font-medium"}>
                      {task.attributes.title}
                    </p>
                    {task.attributes.dueOn && (
                      <p className="text-xs text-muted-foreground">Due: {task.attributes.dueOn}</p>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={task.attributes.status === "completed" ? "secondary" : "default"}>
                    {task.attributes.status}
                  </Badge>
                  <Button variant="ghost" size="sm" onClick={() => handleDelete(task.id)}>
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
