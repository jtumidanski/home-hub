import { useState } from "react";
import { toast } from "sonner";
import { useTasks, useUpdateTask, useDeleteTask } from "@/lib/hooks/api/use-tasks";
import { getErrorMessage } from "@/lib/api/errors";
import { CreateTaskDialog } from "@/components/features/create-task-dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Plus, Check, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";

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
  const { data, isLoading, isError } = useTasks();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();
  const [open, setOpen] = useState(false);

  const tasks = data?.data ?? [];

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

  if (isError) {
    return (
      <div className="p-6">
        <Card className="border-destructive">
          <CardContent className="py-3">
            <p className="text-sm text-destructive">
              Failed to load tasks. Try refreshing the page.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Tasks</h1>
        <Button size="sm" onClick={() => setOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />New Task
        </Button>
      </div>

      <CreateTaskDialog open={open} onOpenChange={setOpen} />

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
                    <Check className={cn("h-4 w-4", task.attributes.status === "completed" ? "text-primary" : "text-muted-foreground")} />
                  </Button>
                  <div>
                    <p className={cn("font-medium", task.attributes.status === "completed" && "line-through text-muted-foreground")}>
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
