'use client';

import { useAuth } from '@/lib/auth';
import { useMobileDashboard } from '@/lib/hooks/useMobileDashboard';
import { MobileDashboardLayout } from '@/app/components/MobileDashboardLayout';
import { MobileHeader } from '@/components/layout/MobileHeader';
import { WeatherCard } from '@/app/components/WeatherCard';
import { RemindersCountCard } from '@/app/components/RemindersCountCard';
import { TasksCountCard } from '@/app/components/TasksCountCard';
import { SignInPage } from '@/app/components/SignInPage';

export default function Home() {
  const { user, loading: authLoading, error: authError } = useAuth();

  // Loading state
  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center space-y-4">
          <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600"></div>
          <p className="text-lg text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Auth error state
  if (authError) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-gray-50">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <h2 className="text-xl font-semibold text-red-900 mb-2">
            Error
          </h2>
          <p className="text-sm text-red-700">
            {authError.message || 'An unexpected error occurred'}
          </p>
        </div>
      </div>
    );
  }

  // Not authenticated - show sign-in page
  if (!user) {
    return <SignInPage />;
  }

  // Show no household warning
  if (!user.householdId) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-gray-50">
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 max-w-md">
          <h2 className="text-lg font-semibold text-yellow-900 mb-2">
            No Household Associated
          </h2>
          <p className="text-sm text-yellow-700 mb-4">
            You are not currently associated with a household. Visit the admin portal to
            create or join a household.
          </p>
          <a
            href="/admin"
            className="inline-flex px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg transition-colors"
          >
            Go to Admin Portal
          </a>
        </div>
      </div>
    );
  }

  // Render dashboard
  return <MobileDashboard householdId={user.householdId} />;
}

interface MobileDashboardProps {
  householdId: string;
}

function MobileDashboard({ householdId }: MobileDashboardProps) {
  const { data, loading, errors } = useMobileDashboard(householdId);

  return (
    <MobileDashboardLayout header={<MobileHeader />}>
      {/* Weather Card - Top Left */}
      <WeatherCard
        weather={data.weather}
        temperatureUnit="fahrenheit"
        loading={loading.weather}
        error={errors.weather}
      />

      {/* Reminders Count Card - Top Right */}
      <RemindersCountCard
        count={data.remindersCount}
        loading={loading.reminders}
        error={errors.reminders}
      />

      {/* Tasks Count Card - Bottom Left */}
      <TasksCountCard
        count={data.tasksCount}
        loading={loading.tasks}
        error={errors.tasks}
      />
    </MobileDashboardLayout>
  );
}
