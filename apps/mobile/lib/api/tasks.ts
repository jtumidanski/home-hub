export interface Task {
  id: string;
  userId: string;
  householdId: string;
  day: string;
  title: string;
  description?: string;
  status: 'incomplete' | 'complete';
  createdAt: string;
  completedAt?: string;
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

export async function getTasks(): Promise<Task[]> {
  const response = await fetch('/api/tasks', {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch tasks: ${response.statusText}`);
  }

  const jsonApiData: JsonApiArrayResponse<Omit<Task, 'id'>> = await response.json();
  return jsonApiData.data.map(flattenResource);
}

/**
 * Filter tasks that are due today
 * @param tasks - Array of tasks
 * @returns Tasks due today
 */
export function filterTasksForToday(tasks: Task[]): Task[] {
  const today = new Date().toISOString().split('T')[0];

  return tasks.filter(task => {
    if (task.status === 'complete') {
      return false;
    }
    return task.day === today;
  });
}
