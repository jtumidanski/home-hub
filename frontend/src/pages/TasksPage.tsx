import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask } from "@/lib/hooks/api/use-tasks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Plus, Check, Trash2 } from "lucide-react";

const taskSchema = z.object({
  title: z.string().min(1, "Title is required"),
  notes: z.string().optional(),
  dueOn: z.string().optional(),
});

type TaskForm = z.infer<typeof taskSchema>;

export function TasksPage() {
  const { data, isLoading } = useTasks();
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();
  const [open, setOpen] = useState(false);

  const form = useForm<TaskForm>({
    resolver: zodResolver(taskSchema),
    defaultValues: { title: "", notes: "", dueOn: "" },
  });

  const tasks = data?.data ?? [];

  const onSubmit = async (values: TaskForm) => {
    await createTask.mutateAsync({
      title: values.title,
      notes: values.notes,
      dueOn: values.dueOn || undefined,
    });
    form.reset();
    setOpen(false);
  };

  const toggleComplete = (id: string, currentStatus: string) => {
    updateTask.mutate({
      id,
      attrs: { status: currentStatus === "completed" ? "pending" : "completed" },
    });
  };

  if (isLoading) {
    return <div className="p-6"><div className="h-8 w-48 animate-pulse rounded bg-muted" /></div>;
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
                <Label htmlFor="dueOn">Due Date</Label>
                <Input id="dueOn" type="date" {...form.register("dueOn")} />
              </div>
              <Button type="submit" className="w-full" disabled={createTask.isPending}>
                {createTask.isPending ? "Creating..." : "Create Task"}
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {tasks.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          No tasks yet. Create your first task to get started.
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
                    <Check className={`h-4 w-4 ${task.attributes.status === "completed" ? "text-green-500" : "text-muted-foreground"}`} />
                  </Button>
                  <div>
                    <p className={`font-medium ${task.attributes.status === "completed" ? "line-through text-muted-foreground" : ""}`}>
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
                  <Button variant="ghost" size="sm" onClick={() => deleteTask.mutate(task.id)}>
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
