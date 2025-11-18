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

export type CreateReminderInput = Pick<Reminder, 'name' | 'remindAt'> & {
  description?: string;
};

export type UpdateReminderInput = Partial<Pick<Reminder, 'name' | 'description' | 'remindAt'>>;

interface JsonApiResource<T> {
  type: string;
  id: string;
  attributes: T;
}

interface JsonApiArrayResponse<T> {
  data: JsonApiResource<T>[];
}

interface JsonApiSingleResponse<T> {
  data: JsonApiResource<T>;
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

export async function createReminder(input: CreateReminderInput): Promise<Reminder> {
  const response = await fetch('/api/reminders', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      data: {
        type: 'reminders',
        attributes: {
          name: input.name,
          description: input.description || '',
          remindAt: input.remindAt,
        },
      },
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to create reminder: ${response.statusText}`);
  }

  const jsonApiData: JsonApiSingleResponse<Omit<Reminder, 'id'>> = await response.json();
  return flattenResource(jsonApiData.data);
}

export async function updateReminder(id: string, input: UpdateReminderInput): Promise<Reminder> {
  const response = await fetch(`/api/reminders/${id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      data: {
        type: 'reminders',
        id,
        attributes: input,
      },
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to update reminder: ${response.statusText}`);
  }

  const jsonApiData: JsonApiSingleResponse<Omit<Reminder, 'id'>> = await response.json();
  return flattenResource(jsonApiData.data);
}

export async function snoozeReminder(id: string, minutes: number): Promise<Reminder> {
  // Calculate new remind time
  const newRemindAt = new Date(Date.now() + minutes * 60 * 1000).toISOString();

  const response = await fetch(`/api/reminders/${id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      data: {
        type: 'reminders',
        id,
        attributes: {
          remindAt: newRemindAt,
          status: 'snoozed',
        },
      },
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to snooze reminder: ${response.statusText}`);
  }

  const jsonApiData: JsonApiSingleResponse<Omit<Reminder, 'id'>> = await response.json();
  return flattenResource(jsonApiData.data);
}

export async function dismissReminder(id: string): Promise<Reminder> {
  const response = await fetch(`/api/reminders/${id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      data: {
        type: 'reminders',
        id,
        attributes: {
          status: 'dismissed',
        },
      },
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to dismiss reminder: ${response.statusText}`);
  }

  const jsonApiData: JsonApiSingleResponse<Omit<Reminder, 'id'>> = await response.json();
  return flattenResource(jsonApiData.data);
}

export async function deleteReminder(id: string): Promise<void> {
  const response = await fetch(`/api/reminders/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      'Accept': 'application/json',
    },
  });

  if (!response.ok && response.status !== 204) {
    throw new Error(`Failed to delete reminder: ${response.statusText}`);
  }
}

/**
 * Filter reminders that are currently active
 * @param reminders - Array of reminders
 * @returns Active reminders only
 */
export function filterActiveReminders(reminders: Reminder[]): Reminder[] {
  return reminders.filter(reminder => reminder.status === 'active');
}
