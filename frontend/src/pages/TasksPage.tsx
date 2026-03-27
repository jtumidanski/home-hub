import { useState, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { type ColumnDef } from "@tanstack/react-table";
import { toast } from "sonner";
import { useTasks, useUpdateTask, useDeleteTask } from "@/lib/hooks/api/use-tasks";
import { useMemberMap } from "@/lib/hooks/api/use-household-members";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { type Task, isTaskOverdue } from "@/types/models/task";
import { useMobile } from "@/lib/hooks/use-mobile";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { ListFilterBar } from "@/components/common/list-filter-bar";
import { TaskCard } from "@/components/features/tasks/task-card";
import { CreateTaskDialog } from "@/components/features/tasks/create-task-dialog";
import { DataTable } from "@/components/common/data-table";
import { ErrorCard } from "@/components/common/error-card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Check, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";

const TASK_STATUS_OPTIONS = [
  { value: "pending", label: "Pending" },
  { value: "completed", label: "Completed" },
  { value: "overdue", label: "Overdue" },
];

function resolveOwnerName(ownerUserId: string | null | undefined, memberMap: Map<string, string>): string {
  if (!ownerUserId) return "Everyone";
  return memberMap.get(ownerUserId) ?? "Former member";
}

export function TasksPage() {
  const { data, isLoading, isError, refetch } = useTasks();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();
  const [open, setOpen] = useState(false);
  const isMobile = useMobile();
  const memberMap = useMemberMap();
  const [searchParams] = useSearchParams();

  const allTasks = (data?.data ?? []) as Task[];

  const filteredTasks = useMemo(() => {
    let result = allTasks;
    const query = searchParams.get("q")?.toLowerCase();
    const status = searchParams.get("status");
    const owner = searchParams.get("owner");

    if (query) {
      result = result.filter((t) => t.attributes.title.toLowerCase().includes(query));
    }
    if (status && status !== "all") {
      if (status === "overdue") {
        result = result.filter((t) => isTaskOverdue(t));
      } else {
        result = result.filter((t) => t.attributes.status === status);
      }
    }
    if (owner && owner !== "all") {
      if (owner === "everyone") {
        result = result.filter((t) => !t.attributes.ownerUserId);
      } else {
        result = result.filter((t) => t.attributes.ownerUserId === owner);
      }
    }
    return result;
  }, [allTasks, searchParams]);

  const hasActiveFilters = searchParams.has("q") || searchParams.has("status") || searchParams.has("owner");

  const toggleComplete = useCallback(async (id: string, currentStatus: string) => {
    try {
      await updateTask.mutateAsync({
        id,
        attrs: { status: currentStatus === "completed" ? "pending" : "completed" },
      });
      toast.success(currentStatus === "completed" ? "Task reopened" : "Task completed");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to update task").message);
    }
  }, [updateTask]);

  const handleDelete = useCallback(async (id: string) => {
    try {
      await deleteTask.mutateAsync(id);
      toast.success("Task deleted");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete task").message);
    }
  }, [deleteTask]);

  const handleRefresh = useCallback(async () => {
    await refetch();
  }, [refetch]);

  const columns: ColumnDef<Task, unknown>[] = [
    {
      id: "complete",
      header: "",
      size: 40,
      cell: ({ row }) => (
        <Button
          variant="ghost"
          size="sm"
          onClick={(e) => {
            e.stopPropagation();
            toggleComplete(row.original.id, row.original.attributes.status);
          }}
        >
          <Check className={cn("h-4 w-4", row.original.attributes.status === "completed" ? "text-primary" : "text-muted-foreground")} />
        </Button>
      ),
    },
    {
      accessorKey: "attributes.title",
      header: "Title",
      cell: ({ row }) => (
        <div>
          <p className={cn("font-medium", row.original.attributes.status === "completed" && "line-through text-muted-foreground")}>
            {row.original.attributes.title}
          </p>
          {row.original.attributes.dueOn && (
            <p className="text-xs text-muted-foreground">Due: {row.original.attributes.dueOn}</p>
          )}
        </div>
      ),
    },
    {
      id: "status",
      header: "Status",
      cell: ({ row }) => {
        const overdue = isTaskOverdue(row.original);
        return (
          <Badge variant={row.original.attributes.status === "completed" ? "secondary" : overdue ? "destructive" : "default"}>
            {overdue ? "overdue" : row.original.attributes.status}
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
      cell: ({ row }) => (
        <Button
          variant="ghost"
          size="sm"
          onClick={(e) => {
            e.stopPropagation();
            handleDelete(row.original.id);
          }}
        >
          <Trash2 className="h-4 w-4 text-destructive" />
        </Button>
      ),
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
        <ErrorCard message="Failed to load tasks. Try refreshing the page." />
      </div>
    );
  }

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl md:text-2xl font-semibold">Tasks</h1>
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />New Task
          </Button>
        </div>

        <ListFilterBar statusOptions={TASK_STATUS_OPTIONS} />

        <CreateTaskDialog open={open} onOpenChange={setOpen} />

        {isMobile ? (
          filteredTasks.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <p className="text-muted-foreground">
                {hasActiveFilters ? "No items found." : "No tasks yet. Create your first task to get started."}
              </p>
              {hasActiveFilters && (
                <Button variant="link" size="sm" onClick={() => window.history.pushState({}, "", window.location.pathname)}>
                  Clear filters
                </Button>
              )}
            </div>
          ) : (
            <div className="space-y-3">
              {filteredTasks.map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  ownerName={resolveOwnerName(task.attributes.ownerUserId, memberMap)}
                  onToggleComplete={toggleComplete}
                  onDelete={handleDelete}
                />
              ))}
            </div>
          )
        ) : (
          <DataTable
            columns={columns}
            data={filteredTasks}
            emptyMessage={hasActiveFilters ? "No items found." : "No tasks yet. Create your first task to get started."}
          />
        )}
        {allTasks.length === 0 && !hasActiveFilters && (
          <div className="flex justify-center">
            <Button variant="outline" onClick={() => setOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create First Task
            </Button>
          </div>
        )}
      </div>
    </PullToRefresh>
  );
}
