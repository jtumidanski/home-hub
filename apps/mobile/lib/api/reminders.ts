export interface Reminder {
  id: string;
  name: string;
  description?: string;
  userId: string;
  householdId: string;
  remindAt: string;
  snoozeCount: number;
  status: 'active' | 'snoozed' | 'dismissed';
  createdAt: string;
  dismissedAt?: string;
  updatedAt: string;
}

interface JsonApiResource<T> {
  type: string;
  id: string;
  attributes: T;
}

interface JsonApiArrayResponse<T> {
  data: JsonApiResource<T>[];
}

function flattenResource<T extends Record<string, any>>(
  resource: JsonApiResource<T>
): T & { id: string } {
  return {
    id: resource.id,
    ...resource.attributes,
  };
}

export async function getReminders(): Promise<Reminder[]> {
  const response = await fetch('/api/reminders', {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch reminders: ${response.statusText}`);
  }

  const jsonApiData: JsonApiArrayResponse<Omit<Reminder, 'id'>> = await response.json();
  return jsonApiData.data.map(flattenResource);
}

/**
 * Filter reminders that are currently active
 * @param reminders - Array of reminders
 * @returns Active reminders only
 */
export function filterActiveReminders(reminders: Reminder[]): Reminder[] {
  return reminders.filter(reminder => reminder.status === 'active');
}
