'use client';

import React from 'react';
import { Card, CardSection } from '@/app/components/ui/Card';
import { WeatherResponse, formatTemperature, formatForecastDate } from '@/lib/api/weather';
import { CalendarEvent } from '@/lib/api/calendar';
import { Task } from '@/lib/api/tasks';
import { Calendar, CheckSquare, CloudSun } from 'lucide-react';

interface TomorrowPreviewCardProps {
  weather?: WeatherResponse | null;
  events?: CalendarEvent[] | null;
  tasks?: Task[] | null;
  loading?: boolean;
  error?: string | null;
}

export function TomorrowPreviewCard({
  weather,
  events,
  tasks,
  loading,
  error,
}: TomorrowPreviewCardProps) {
  if (error) {
    return (
      <Card title="Tomorrow">
        <div className="text-center py-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      </Card>
    );
  }

  if (loading) {
    return <Card title="Tomorrow" loading={true}>{null}</Card>;
  }

  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const tomorrowStr = tomorrow.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  });

  // Get tomorrow's forecast
  const tomorrowDate = tomorrow.toISOString().split('T')[0];
  const tomorrowForecast = weather?.daily?.find(d => d.date === tomorrowDate);

  return (
    <Card>
      <div className="mb-4">
        <h3 className="font-semibold text-gray-900 dark:text-white">Tomorrow</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mt-0.5">{tomorrowStr}</p>
      </div>

      <div className="space-y-6">
        {/* Weather */}
        {tomorrowForecast && (
          <CardSection>
            <div className="flex items-center gap-3">
              <CloudSun className="h-8 w-8 text-yellow-500" />
              <div>
                <div className="text-sm text-gray-600 dark:text-gray-400">Weather</div>
                <div className="text-lg font-semibold text-gray-900 dark:text-white">
                  {formatTemperature(tomorrowForecast.tmax_c, weather?.units || 'celsius')} /{' '}
                  {formatTemperature(tomorrowForecast.tmin_c, weather?.units || 'celsius')}
                </div>
              </div>
            </div>
          </CardSection>
        )}

        {/* Events */}
        <CardSection>
          <div className="flex items-center gap-2 mb-3">
            <Calendar className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Events
            </span>
            <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-0.5 rounded">
              Preview
            </span>
          </div>
          {events && events.length > 0 ? (
            <div className="space-y-2">
              {events.slice(0, 3).map(event => (
                <div
                  key={event.id}
                  className="text-sm p-2 bg-gray-50 dark:bg-gray-700/50 rounded"
                >
                  <div className="font-medium text-gray-900 dark:text-white">
                    {event.title}
                  </div>
                  <div className="text-xs text-gray-600 dark:text-gray-400 mt-0.5">
                    {formatEventTime(event.startTime)}
                  </div>
                </div>
              ))}
              {events.length > 3 && (
                <p className="text-xs text-gray-500 dark:text-gray-400 italic">
                  +{events.length - 3} more
                </p>
              )}
            </div>
          ) : (
            <p className="text-sm text-gray-500 dark:text-gray-400 italic">No events</p>
          )}
        </CardSection>

        {/* Tasks */}
        <CardSection>
          <div className="flex items-center gap-2 mb-3">
            <CheckSquare className="h-4 w-4 text-green-600 dark:text-green-400" />
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Tasks
            </span>
          </div>
          {(() => {
            const incompleteTasks = tasks?.filter(t => t.status !== 'complete') || [];
            return incompleteTasks.length > 0 ? (
              <div className="space-y-1">
                {incompleteTasks.slice(0, 3).map(task => (
                  <div
                    key={task.id}
                    className="text-sm text-gray-700 dark:text-gray-300 flex items-center gap-2"
                  >
                    <span className="w-1.5 h-1.5 bg-gray-400 dark:bg-gray-600 rounded-full"></span>
                    {task.title}
                  </div>
                ))}
                {incompleteTasks.length > 3 && (
                  <p className="text-xs text-gray-500 dark:text-gray-400 italic">
                    +{incompleteTasks.length - 3} more
                  </p>
                )}
              </div>
            ) : (
              <p className="text-sm text-gray-500 dark:text-gray-400 italic">No tasks</p>
            );
          })()}
        </CardSection>
      </div>
    </Card>
  );
}

function formatEventTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
}
