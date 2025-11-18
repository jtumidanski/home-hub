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

export type CreateTaskInput = Pick<Task, 'title' | 'day'> & {
  description?: string;
  status?: 'incomplete' | 'complete';
};

export type UpdateTaskInput = Partial<Pick<Task, 'title' | 'description' | 'day' | 'status'>>;

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

export async function createTask(input: CreateTaskInput): Promise<Task> {
  const response = await fetch('/api/tasks', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      data: {
        type: 'tasks',
        attributes: {
          title: input.title,
          description: input.description || '',
          day: input.day,
          status: input.status || 'incomplete',
        },
      },
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to create task: ${response.statusText}`);
  }

  const jsonApiData: JsonApiSingleResponse<Omit<Task, 'id'>> = await response.json();
  return flattenResource(jsonApiData.data);
}

export async function updateTask(id: string, input: UpdateTaskInput): Promise<Task> {
  const response = await fetch(`/api/tasks/${id}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      data: {
        type: 'tasks',
        id,
        attributes: input,
      },
    }),
  });

  if (!response.ok) {
    throw new Error(`Failed to update task: ${response.statusText}`);
  }

  const jsonApiData: JsonApiSingleResponse<Omit<Task, 'id'>> = await response.json();
  return flattenResource(jsonApiData.data);
}

export async function completeTask(id: string): Promise<Task> {
  return updateTask(id, {
    status: 'complete',
  });
}

export async function deleteTask(id: string): Promise<void> {
  const response = await fetch(`/api/tasks/${id}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      'Accept': 'application/json',
    },
  });

  if (!response.ok && response.status !== 204) {
    throw new Error(`Failed to delete task: ${response.statusText}`);
  }
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
