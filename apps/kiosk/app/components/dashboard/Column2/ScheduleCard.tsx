'use client';

import React from 'react';
import { Card, CardSection } from '@/app/components/ui/Card';
import { CalendarEvent } from '@/lib/api/calendar';
import { Calendar, Clock } from 'lucide-react';

interface ScheduleCardProps {
  events?: CalendarEvent[] | null;
  loading?: boolean;
  error?: string | null;
}

export function ScheduleCard({ events, loading, error }: ScheduleCardProps) {
  if (error) {
    return (
      <Card title="Schedule">
        <div className="flex justify-between items-center mb-2">
          <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
            Preview
          </span>
        </div>
        <div className="text-center py-4">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      </Card>
    );
  }

  if (loading || !events) {
    return <Card title="Schedule" loading={true}>{null}</Card>;
  }

  const today = new Date();
  const dateStr = today.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  });

  // Sort events by start time
  const sortedEvents = [...events].sort(
    (a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
  );

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="font-semibold text-gray-900 dark:text-white">Schedule</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-0.5">{dateStr}</p>
        </div>
        <span className="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
          Preview
        </span>
      </div>

      {sortedEvents.length === 0 ? (
        <div className="text-center py-8">
          <Calendar className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600 mb-2" />
          <p className="text-sm text-gray-500 dark:text-gray-400">No events today</p>
        </div>
      ) : (
        <div className="space-y-3">
          {sortedEvents.map(event => (
            <EventItem key={event.id} event={event} />
          ))}
        </div>
      )}
    </Card>
  );
}

function EventItem({ event }: { event: CalendarEvent }) {
  const startTime = new Date(event.startTime);
  const endTime = new Date(event.endTime);
  const timeStr = `${formatTime(startTime)} - ${formatTime(endTime)}`;

  return (
    <div className="flex gap-3 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
      <div className="flex-shrink-0">
        <Clock className="h-5 w-5 text-blue-600 dark:text-blue-400" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-gray-900 dark:text-white">
          {event.title}
        </div>
        <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
          {timeStr}
        </div>
        <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          {event.calendar}
        </div>
      </div>
    </div>
  );
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
}
