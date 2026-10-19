import { useState, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { useTasks, useUpdateTask, useDeleteTask } from "@/lib/hooks/api/use-tasks";
import { useMemberMap, useHouseholdMembers } from "@/lib/hooks/api/use-household-members";
import { useAuth } from "@/components/providers/auth-provider";
import type { Member } from "@/types/models/member";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { type Task, isTaskOverdue } from "@/types/models/task";
import { useMobile } from "@/lib/hooks/use-mobile";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { ListFilterBar } from "@/components/common/list-filter-bar";
import { TaskCard } from "@/components/features/tasks/task-card";
import { CreateTaskDialog } from "@/components/features/tasks/create-task-dialog";
import { DataTable, type DataTableColumnDef } from "@/components/common/data-table";
import { ErrorCard } from "@/components/common/error-card";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Circle, CheckCircle2, Trash2, ListChecks } from "lucide-react";
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
  const { user } = useAuth();
  const { data: membersData } = useHouseholdMembers();
  const members = (membersData?.data ?? []) as Member[];
  const memberMap = useMemberMap();
  const [searchParams] = useSearchParams();

  const allTasks = useMemo(() => (data?.data ?? []) as Task[], [data]);

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

  const columns: DataTableColumnDef<Task>[] = [
    {
      id: "complete",
      header: "",
      size: 40,
      cell: ({ row }) => {
        const isCompleted = row.original.attributes.status === "completed";
        const Icon = isCompleted ? CheckCircle2 : Circle;
        return (
          <Button
            variant="ghost"
            size="sm"
            className="coarse:min-h-11 coarse:min-w-11"
            aria-label={isCompleted ? "Mark incomplete" : "Mark complete"}
            onClick={(e) => {
              e.stopPropagation();
              toggleComplete(row.original.id, row.original.attributes.status);
            }}
          >
            <Icon className={cn("h-4 w-4", isCompleted ? "text-primary" : "text-muted-foreground")} />
          </Button>
        );
      },
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
            <p className="text-sm text-muted-foreground tabular-nums">Due {row.original.attributes.dueOn}</p>
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
          className="coarse:min-h-11 coarse:min-w-11"
          aria-label="Delete task"
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
        <PageHeader
          title="Tasks"
          actions={
            <Button size="sm" className="coarse:min-h-11" onClick={() => setOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />New Task
            </Button>
          }
        />

        <ListFilterBar statusOptions={TASK_STATUS_OPTIONS} />

        <CreateTaskDialog open={open} onOpenChange={setOpen} currentUserId={user?.id ?? ""} members={members} />

        {isMobile ? (
          filteredTasks.length === 0 ? (
            <EmptyState
              icon={ListChecks}
              title={hasActiveFilters ? "No items found." : "No tasks yet. Create your first task to get started."}
              action={
                hasActiveFilters ? (
                  <Button variant="link" size="sm" onClick={() => window.history.pushState({}, "", window.location.pathname)}>
                    Clear filters
                  </Button>
                ) : undefined
              }
            />
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
