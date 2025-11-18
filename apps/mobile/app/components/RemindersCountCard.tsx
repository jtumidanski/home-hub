import { Bell } from 'lucide-react';

interface RemindersCountCardProps {
  count: number;
  loading: boolean;
  error: string | null;
}

export function RemindersCountCard({ count, loading, error }: RemindersCountCardProps) {
  if (loading) {
    return (
      <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
        <div className="animate-pulse space-y-4">
          <div className="h-6 bg-gray-200 rounded w-1/2"></div>
          <div className="h-24 bg-gray-200 rounded w-1/2 mx-auto"></div>
          <div className="h-4 bg-gray-200 rounded w-2/3 mx-auto"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white border border-red-300 rounded-lg p-6 shadow-sm">
        <p className="text-sm text-red-600">Failed to load reminders</p>
        <p className="text-xs text-gray-600 mt-1">{error}</p>
      </div>
    );
  }

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-center gap-2 mb-4">
        <Bell className="w-5 h-5 text-blue-600" />
        <h3 className="text-sm font-medium text-gray-600">Active Reminders</h3>
      </div>

      <div className="flex flex-col items-center justify-center py-6">
        <div className="text-7xl font-bold text-blue-600 mb-2">
          {count}
        </div>
        <p className="text-sm text-gray-600">
          {count === 1 ? 'reminder' : 'reminders'} active
        </p>
      </div>
    </div>
  );
}
