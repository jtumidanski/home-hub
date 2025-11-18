import { CheckCircle2 } from 'lucide-react';

interface TasksCountCardProps {
  count: number;
  loading: boolean;
  error: string | null;
}

export function TasksCountCard({ count, loading, error }: TasksCountCardProps) {
  if (loading) {
    return (
      <div className="bg-card border border-border rounded-lg p-6 shadow-sm">
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-muted rounded w-1/2"></div>
          <div className="h-24 bg-muted rounded w-1/2 mx-auto"></div>
          <div className="h-4 bg-muted rounded w-2/3 mx-auto"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-card border border-destructive/30 rounded-lg p-6 shadow-sm">
        <p className="text-sm text-destructive">Failed to load tasks</p>
        <p className="text-xs text-muted-foreground mt-1">{error}</p>
      </div>
    );
  }

  return (
    <div className="bg-card text-card-foreground border border-border rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-center gap-2 mb-4">
        <CheckCircle2 className="w-5 h-5 text-blue-600 dark:text-blue-500" />
        <h3 className="text-sm font-medium text-muted-foreground">Tasks Due Today</h3>
      </div>

      <div className="flex flex-col items-center justify-center py-6">
        <div className="text-7xl font-bold text-blue-600 dark:text-blue-500 mb-2">
          {count}
        </div>
        <p className="text-sm text-muted-foreground">
          {count === 1 ? 'task' : 'tasks'} due today
        </p>
      </div>
    </div>
  );
}
