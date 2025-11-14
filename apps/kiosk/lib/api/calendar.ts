export interface CalendarEvent {
  id: string;
  title: string;
  startTime: string;
  endTime: string;
  calendar: string;
  allDay?: boolean;
}

/**
 * Mock calendar events - will be replaced with real API call
 */
export async function getCalendarEvents(date?: string): Promise<CalendarEvent[]> {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 300));

  const today = new Date();
  const todayStr = today.toISOString().split('T')[0];

  // If a specific date is requested, filter to that date
  const targetDate = date || todayStr;

  // Create some sample events for today
  if (targetDate === todayStr) {
    return [
      {
        id: '1',
        title: 'Team meeting',
        startTime: `${todayStr}T09:00:00Z`,
        endTime: `${todayStr}T10:00:00Z`,
        calendar: 'Work',
      },
      {
        id: '2',
        title: 'Dentist appointment',
        startTime: `${todayStr}T14:00:00Z`,
        endTime: `${todayStr}T15:00:00Z`,
        calendar: 'Personal',
      },
      {
        id: '3',
        title: 'Soccer practice',
        startTime: `${todayStr}T17:30:00Z`,
        endTime: `${todayStr}T19:00:00Z`,
        calendar: 'Kids',
      },
    ];
  }

  // Tomorrow
  const tomorrow = new Date(today);
  tomorrow.setDate(today.getDate() + 1);
  const tomorrowStr = tomorrow.toISOString().split('T')[0];

  if (targetDate === tomorrowStr) {
    return [
      {
        id: '4',
        title: 'Client presentation',
        startTime: `${tomorrowStr}T10:00:00Z`,
        endTime: `${tomorrowStr}T11:30:00Z`,
        calendar: 'Work',
      },
      {
        id: '5',
        title: 'Grocery shopping',
        startTime: `${tomorrowStr}T15:00:00Z`,
        endTime: `${tomorrowStr}T16:00:00Z`,
        calendar: 'Personal',
      },
    ];
  }

  return [];
}
