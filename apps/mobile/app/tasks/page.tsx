'use client';

import { useState } from 'react';
import { Plus } from 'lucide-react';
import { MobileHeader } from '@/components/layout/MobileHeader';
import { useTasks } from '@/lib/hooks/useTasks';
import { Task, CreateTaskInput, UpdateTaskInput } from '@/lib/api/tasks';
import { TasksList } from './components/TasksList';
import { TaskFormDialog } from './components/TaskFormDialog';
import { Button } from '@/components/ui/button';
import { PullToRefresh } from '@/components/ui/pull-to-refresh';

export default function TasksPage() {
  const { tasks, loading, error, createTask, updateTask, completeTask, deleteTask, refresh } = useTasks();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | undefined>(undefined);

  const handleCreateTask = async (data: CreateTaskInput | UpdateTaskInput) => {
    // For create, we need all required fields
    await createTask(data as CreateTaskInput);
  };

  const handleUpdateTask = async (data: CreateTaskInput | UpdateTaskInput) => {
    if (editingTask) {
      await updateTask(editingTask.id, data);
    }
  };

  const handleEdit = (task: Task) => {
    setEditingTask(task);
  };

  const handleComplete = async (id: string) => {
    await completeTask(id);
  };

  const handleDelete = async (id: string) => {
    await deleteTask(id);
  };

  return (
    <div className="min-h-screen bg-background pb-20">
      <MobileHeader />

      <PullToRefresh onRefresh={refresh}>
        <div className="p-4 md:p-6 max-w-7xl mx-auto">
          <div className="mb-6">
            <h1 className="text-2xl font-bold text-card-foreground mb-2">Tasks</h1>
            <p className="text-sm text-muted-foreground">
              Manage your tasks and to-dos
            </p>
          </div>

          {/* Loading State */}
          {loading && (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="bg-card border border-border rounded-lg p-4 animate-pulse">
                  <div className="h-5 bg-muted rounded w-3/4 mb-2"></div>
                  <div className="h-4 bg-muted rounded w-1/2"></div>
                </div>
              ))}
            </div>
          )}

          {/* Error State */}
          {error && !loading && (
            <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          {/* Tasks List */}
          {!loading && !error && (
            <TasksList
              tasks={tasks}
              onComplete={handleComplete}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          )}
        </div>
      </PullToRefresh>

      {/* Floating Action Button */}
      <div className="fixed bottom-6 right-6 z-40">
        <Button
          size="icon"
          className="h-14 w-14 rounded-full shadow-lg hover:shadow-xl transition-shadow"
          onClick={() => setIsCreateDialogOpen(true)}
        >
          <Plus className="h-6 w-6" />
        </Button>
      </div>

      {/* Create Dialog */}
      <TaskFormDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onSubmit={handleCreateTask}
      />

      {/* Edit Dialog */}
      <TaskFormDialog
        open={!!editingTask}
        onOpenChange={(open) => !open && setEditingTask(undefined)}
        task={editingTask}
        onSubmit={handleUpdateTask}
      />
    </div>
  );
}
