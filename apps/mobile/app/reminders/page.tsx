'use client';

import { useState } from 'react';
import { Plus } from 'lucide-react';
import { MobileHeader } from '@/components/layout/MobileHeader';
import { useReminders } from '@/lib/hooks/useReminders';
import { Reminder, CreateReminderInput, UpdateReminderInput } from '@/lib/api/reminders';
import { RemindersList } from './components/RemindersList';
import { ReminderFormDialog } from './components/ReminderFormDialog';
import { Button } from '@/components/ui/button';
import { PullToRefresh } from '@/components/ui/pull-to-refresh';

export default function RemindersPage() {
  const { reminders, loading, error, createReminder, updateReminder, snoozeReminder, dismissReminder, deleteReminder, refresh } = useReminders();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [editingReminder, setEditingReminder] = useState<Reminder | undefined>(undefined);

  const handleCreateReminder = async (data: CreateReminderInput | UpdateReminderInput) => {
    // For create, we need all required fields
    await createReminder(data as CreateReminderInput);
  };

  const handleUpdateReminder = async (data: CreateReminderInput | UpdateReminderInput) => {
    if (editingReminder) {
      await updateReminder(editingReminder.id, data);
    }
  };

  const handleSnooze = async (id: string, minutes: number) => {
    await snoozeReminder(id, minutes);
  };

  const handleDismiss = async (id: string) => {
    await dismissReminder(id);
  };

  const handleEdit = (reminder: Reminder) => {
    setEditingReminder(reminder);
  };

  const handleDelete = async (id: string) => {
    await deleteReminder(id);
  };

  return (
    <div className="min-h-screen bg-background pb-20">
      <MobileHeader />

      <PullToRefresh onRefresh={refresh}>
        <div className="p-4 md:p-6 max-w-7xl mx-auto">
          <div className="mb-6">
            <h1 className="text-2xl font-bold text-card-foreground mb-2">Reminders</h1>
            <p className="text-sm text-muted-foreground">
              Manage your reminders and notifications
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

          {/* Reminders List */}
          {!loading && !error && (
            <RemindersList
              reminders={reminders}
              onSnooze={handleSnooze}
              onDismiss={handleDismiss}
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
      <ReminderFormDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onSubmit={handleCreateReminder}
      />

      {/* Edit Dialog */}
      <ReminderFormDialog
        open={!!editingReminder}
        onOpenChange={(open) => !open && setEditingReminder(undefined)}
        reminder={editingReminder}
        onSubmit={handleUpdateReminder}
      />
    </div>
  );
}
