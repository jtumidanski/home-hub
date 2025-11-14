'use client';

import { useEffect, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  AlertCircle,
  CheckCircle2,
  Circle,
  Plus,
  Trash2,
  Pencil,
  X,
  Filter,
} from 'lucide-react';
import {
  Task,
  listTasks,
  createTask,
  updateTask,
  deleteTask,
  completeTask,
  uncompleteTask,
  CreateTaskInput,
  UpdateTaskInput,
} from '@/lib/api/tasks';
import type { User } from '@/lib/api/users';
import { toast } from 'sonner';
import { Separator } from '@/components/ui/separator';
import { TaskDeleteDialog } from './TaskDeleteDialog';

interface UserTasksModalProps {
  user: User | null;
  open: boolean;
  onClose: () => void;
  onSave?: () => void;
}

type FilterStatus = 'all' | 'incomplete' | 'complete';

export function UserTasksModal({ user, open, onClose, onSave }: UserTasksModalProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [filterDay, setFilterDay] = useState<string>('');
  const [filterStatus, setFilterStatus] = useState<FilterStatus>('all');
  const [showFilters, setShowFilters] = useState(false);

  // Create form
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createForm, setCreateForm] = useState<CreateTaskInput>({
    userId: '',
    day: new Date().toISOString().split('T')[0],
    title: '',
    description: '',
  });
  const [creating, setCreating] = useState(false);

  // Edit form
  const [editingTaskId, setEditingTaskId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<UpdateTaskInput>({});
  const [updating, setUpdating] = useState(false);

  // Delete dialog
  const [taskToDelete, setTaskToDelete] = useState<Task | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  // Fetch tasks when modal opens
  useEffect(() => {
    if (user && open) {
      fetchTasks();
      setCreateForm(prev => ({ ...prev, userId: user.id }));
    }
  }, [user, open]);

  const fetchTasks = async () => {
    if (!user) return;

    try {
      setLoading(true);
      setError(null);

      const fetchedTasks = await listTasks(user.id);
      setTasks(fetchedTasks);
    } catch (err) {
      console.error('Failed to fetch tasks:', err);
      setError(err instanceof Error ? err.message : 'Failed to load tasks');
      toast.error('Failed to load tasks');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!createForm.title.trim()) {
      toast.error('Task title is required');
      return;
    }

    try {
      setCreating(true);
      const newTask = await createTask(createForm);
      setTasks(prev => [...prev, newTask]);

      // Reset form
      setCreateForm({
        userId: user!.id,
        day: new Date().toISOString().split('T')[0],
        title: '',
        description: '',
      });
      setShowCreateForm(false);
      toast.success('Task created successfully');
    } catch (err) {
      console.error('Failed to create task:', err);
      toast.error('Failed to create task');
    } finally {
      setCreating(false);
    }
  };

  const handleStartEdit = (task: Task) => {
    setEditingTaskId(task.id);
    setEditForm({
      title: task.title,
      description: task.description,
      day: task.day,
      status: task.status,
    });
  };

  const handleCancelEdit = () => {
    setEditingTaskId(null);
    setEditForm({});
  };

  const handleUpdate = async (taskId: string) => {
    try {
      setUpdating(true);
      const updatedTask = await updateTask(taskId, editForm);
      setTasks(prev => prev.map(t => t.id === taskId ? updatedTask : t));
      setEditingTaskId(null);
      setEditForm({});
      toast.success('Task updated successfully');
    } catch (err) {
      console.error('Failed to update task:', err);
      toast.error('Failed to update task');
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = (task: Task) => {
    setTaskToDelete(task);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirmed = () => {
    setDeleteDialogOpen(false);
    if (taskToDelete) {
      setTasks(prev => prev.filter(t => t.id !== taskToDelete.id));
    }
    setTaskToDelete(null);
  };

  const handleToggleComplete = async (task: Task) => {
    try {
      const updatedTask = task.status === 'complete'
        ? await uncompleteTask(task.id)
        : await completeTask(task.id);

      setTasks(prev => prev.map(t => t.id === task.id ? updatedTask : t));
      toast.success(task.status === 'complete' ? 'Task marked incomplete' : 'Task completed');
    } catch (err) {
      console.error('Failed to toggle task status:', err);
      toast.error('Failed to update task status');
    }
  };

  // Apply filters
  const filteredTasks = tasks.filter(task => {
    if (filterDay && task.day !== filterDay) return false;
    if (filterStatus !== 'all' && task.status !== filterStatus) return false;
    return true;
  });

  // Sort tasks: incomplete first, then by day (oldest first), then by created date
  const sortedTasks = [...filteredTasks].sort((a, b) => {
    if (a.status !== b.status) {
      return a.status === 'incomplete' ? -1 : 1;
    }
    if (a.day !== b.day) {
      return a.day.localeCompare(b.day);
    }
    return a.createdAt.localeCompare(b.createdAt);
  });

  if (!user) return null;

  return (
    <>
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Tasks - {user.displayName}</DialogTitle>
          <DialogDescription>
            View and manage tasks for this user
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Action Bar */}
          <div className="flex items-center justify-between gap-2">
            <Button
              onClick={() => setShowCreateForm(!showCreateForm)}
              size="sm"
              variant={showCreateForm ? 'secondary' : 'default'}
            >
              {showCreateForm ? (
                <>
                  <X className="h-4 w-4 mr-2" />
                  Cancel
                </>
              ) : (
                <>
                  <Plus className="h-4 w-4 mr-2" />
                  New Task
                </>
              )}
            </Button>

            <Button
              onClick={() => setShowFilters(!showFilters)}
              size="sm"
              variant="outline"
            >
              <Filter className="h-4 w-4 mr-2" />
              Filters
            </Button>
          </div>

          {/* Create Form */}
          {showCreateForm && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Create New Task</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="create-title">Title *</Label>
                  <Input
                    id="create-title"
                    value={createForm.title}
                    onChange={(e) => setCreateForm(prev => ({ ...prev, title: e.target.value }))}
                    placeholder="Task title"
                    maxLength={200}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="create-description">Description</Label>
                  <Input
                    id="create-description"
                    value={createForm.description}
                    onChange={(e) => setCreateForm(prev => ({ ...prev, description: e.target.value }))}
                    placeholder="Optional description"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="create-day">Day *</Label>
                  <Input
                    id="create-day"
                    type="date"
                    value={createForm.day}
                    onChange={(e) => setCreateForm(prev => ({ ...prev, day: e.target.value }))}
                  />
                </div>

                <div className="flex justify-end gap-2">
                  <Button
                    onClick={() => setShowCreateForm(false)}
                    variant="outline"
                    size="sm"
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={handleCreate}
                    disabled={creating || !createForm.title.trim()}
                    size="sm"
                  >
                    {creating ? 'Creating...' : 'Create Task'}
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Filters */}
          {showFilters && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Filters</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="filter-day">Filter by Day</Label>
                    <Input
                      id="filter-day"
                      type="date"
                      value={filterDay}
                      onChange={(e) => setFilterDay(e.target.value)}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="filter-status">Filter by Status</Label>
                    <select
                      id="filter-status"
                      value={filterStatus}
                      onChange={(e) => setFilterStatus(e.target.value as FilterStatus)}
                      className="w-full h-10 px-3 py-2 text-sm border border-input bg-background rounded-md"
                    >
                      <option value="all">All Tasks</option>
                      <option value="incomplete">Incomplete</option>
                      <option value="complete">Complete</option>
                    </select>
                  </div>
                </div>

                <Button
                  onClick={() => {
                    setFilterDay('');
                    setFilterStatus('all');
                  }}
                  variant="outline"
                  size="sm"
                >
                  Clear Filters
                </Button>
              </CardContent>
            </Card>
          )}

          {/* Task List */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">
                Tasks ({filteredTasks.length})
              </CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="space-y-2">
                  {[1, 2, 3].map(i => (
                    <div key={i} className="h-16 bg-neutral-100 dark:bg-neutral-800 rounded animate-pulse" />
                  ))}
                </div>
              ) : error ? (
                <div className="text-center py-8">
                  <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-2" />
                  <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
                  <Button onClick={fetchTasks} variant="outline" size="sm" className="mt-4">
                    Retry
                  </Button>
                </div>
              ) : sortedTasks.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-sm text-neutral-500 dark:text-neutral-400">
                    {tasks.length === 0 ? 'No tasks yet' : 'No tasks match filters'}
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {sortedTasks.map(task => (
                    <div key={task.id}>
                      {editingTaskId === task.id ? (
                        // Edit mode
                        <div className="border border-neutral-200 dark:border-neutral-700 rounded-lg p-4 space-y-3">
                          <div className="space-y-2">
                            <Label htmlFor={`edit-title-${task.id}`}>Title</Label>
                            <Input
                              id={`edit-title-${task.id}`}
                              value={editForm.title || ''}
                              onChange={(e) => setEditForm(prev => ({ ...prev, title: e.target.value }))}
                              maxLength={200}
                            />
                          </div>

                          <div className="space-y-2">
                            <Label htmlFor={`edit-description-${task.id}`}>Description</Label>
                            <Input
                              id={`edit-description-${task.id}`}
                              value={editForm.description || ''}
                              onChange={(e) => setEditForm(prev => ({ ...prev, description: e.target.value }))}
                            />
                          </div>

                          <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                              <Label htmlFor={`edit-day-${task.id}`}>Day</Label>
                              <Input
                                id={`edit-day-${task.id}`}
                                type="date"
                                value={editForm.day || ''}
                                onChange={(e) => setEditForm(prev => ({ ...prev, day: e.target.value }))}
                              />
                            </div>

                            <div className="space-y-2">
                              <Label htmlFor={`edit-status-${task.id}`}>Status</Label>
                              <select
                                id={`edit-status-${task.id}`}
                                value={editForm.status || ''}
                                onChange={(e) => setEditForm(prev => ({ ...prev, status: e.target.value as 'incomplete' | 'complete' }))}
                                className="w-full h-10 px-3 py-2 text-sm border border-input bg-background rounded-md"
                              >
                                <option value="incomplete">Incomplete</option>
                                <option value="complete">Complete</option>
                              </select>
                            </div>
                          </div>

                          <div className="flex justify-end gap-2">
                            <Button onClick={handleCancelEdit} variant="outline" size="sm">
                              Cancel
                            </Button>
                            <Button
                              onClick={() => handleUpdate(task.id)}
                              disabled={updating}
                              size="sm"
                            >
                              {updating ? 'Saving...' : 'Save'}
                            </Button>
                          </div>
                        </div>
                      ) : (
                        // View mode
                        <div className="flex items-start gap-3 p-3 border border-neutral-200 dark:border-neutral-700 rounded-lg hover:bg-neutral-50 dark:hover:bg-neutral-800/50 transition-colors">
                          <button
                            onClick={() => handleToggleComplete(task)}
                            className="mt-0.5"
                          >
                            {task.status === 'complete' ? (
                              <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400" />
                            ) : (
                              <Circle className="h-5 w-5 text-neutral-400 dark:text-neutral-600" />
                            )}
                          </button>

                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-2">
                              <div className="flex-1">
                                <p className={`text-sm font-medium ${
                                  task.status === 'complete'
                                    ? 'line-through text-neutral-500 dark:text-neutral-400'
                                    : 'text-neutral-900 dark:text-white'
                                }`}>
                                  {task.title}
                                </p>
                                {task.description && (
                                  <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">
                                    {task.description}
                                  </p>
                                )}
                                <p className="text-xs text-neutral-400 dark:text-neutral-500 mt-1">
                                  {task.day}
                                  {task.status === 'incomplete' && task.day < new Date().toISOString().split('T')[0] && (
                                    <span className="ml-2 text-red-600 dark:text-red-400">• Overdue</span>
                                  )}
                                </p>
                              </div>

                              <div className="flex items-center gap-1">
                                <Button
                                  onClick={() => handleStartEdit(task)}
                                  variant="ghost"
                                  size="sm"
                                >
                                  <Pencil className="h-4 w-4" />
                                </Button>
                                <Button
                                  onClick={() => handleDelete(task)}
                                  variant="ghost"
                                  size="sm"
                                >
                                  <Trash2 className="h-4 w-4 text-red-600 dark:text-red-400" />
                                </Button>
                              </div>
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    {/* Delete Confirmation Dialog */}
    <TaskDeleteDialog
      task={taskToDelete}
      open={deleteDialogOpen}
      onClose={() => {
        setDeleteDialogOpen(false);
        setTaskToDelete(null);
      }}
      onDeleted={handleDeleteConfirmed}
    />
  </>
  );
}
