export interface Reminder {
  id: string;
  text: string;
  triggerAt: string;
  status: 'active' | 'snoozed' | 'dismissed';
  snoozedUntil?: string;
}

/**
 * Mock reminders data - will be replaced with real API call
 */
export async function getReminders(): Promise<Reminder[]> {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 300));

  const now = new Date();
  const later = new Date(now.getTime() + 3600000); // 1 hour from now

  return [
    {
      id: '1',
      text: 'Doctor appointment tomorrow at 2pm',
      triggerAt: later.toISOString(),
      status: 'active',
    },
    {
      id: '2',
      text: 'Pick up prescription',
      triggerAt: now.toISOString(),
      status: 'active',
    },
    {
      id: '3',
      text: 'Return library books',
      triggerAt: new Date(now.getTime() - 1800000).toISOString(), // 30 min ago
      status: 'active',
    },
  ];
}

/**
 * Mock snooze reminder - will be replaced with real API call
 */
export async function snoozeReminder(id: string, duration: number): Promise<void> {
  console.log(`[Mock] Snoozing reminder ${id} for ${duration}ms`);
  await new Promise(resolve => setTimeout(resolve, 100));
}

/**
 * Mock dismiss reminder - will be replaced with real API call
 */
export async function dismissReminder(id: string): Promise<void> {
  console.log(`[Mock] Dismissing reminder ${id}`);
  await new Promise(resolve => setTimeout(resolve, 100));
}
