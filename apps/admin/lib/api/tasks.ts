/**
 * Tasks Service API
 *
 * Client for interacting with the svc-tasks service via the gateway.
 * Handles task CRUD operations for users.
 */

import { get, post, patch, del } from "./client";

/**
 * Task attributes (what's inside the JSON:API attributes field)
 */
export interface TaskAttributes {
  userId: string;
  householdId: string;
  day: string; // YYYY-MM-DD format
  title: string;
  description?: string;
  status: "incomplete" | "complete";
  createdAt: string; // ISO 8601 format
  completedAt?: string; // ISO 8601 format
  updatedAt: string; // ISO 8601 format
}

/**
 * Task interface (flattened for easier use in components)
 */
export interface Task {
  id: string;
  userId: string;
  householdId: string;
  day: string;
  title: string;
  description?: string;
  status: "incomplete" | "complete";
  createdAt: string;
  completedAt?: string;
  updatedAt: string;
}

/**
 * JSON:API resource structure
 */
export interface JsonApiResource<T> {
  type: string;
  id: string;
  attributes: T;
}

/**
 * JSON:API response wrapper for single resource
 */
export interface JsonApiResponse<T> {
  data: JsonApiResource<T>;
}

/**
 * JSON:API response wrapper for array of resources
 */
export interface JsonApiArrayResponse<T> {
  data: JsonApiResource<T>[];
}

/**
 * Input for creating a new task
 */
export interface CreateTaskInput {
  userId: string;
  day: string; // YYYY-MM-DD format
  title: string;
  description?: string;
}

/**
 * Input for updating an existing task
 */
export interface UpdateTaskInput {
  title?: string;
  description?: string;
  day?: string; // YYYY-MM-DD format
  status?: "incomplete" | "complete";
}

/**
 * Helper to flatten JSON:API resource to plain object
 */
function flattenResource<T extends Record<string, any>>(
  resource: JsonApiResource<T>
): T & { id: string } {
  return {
    id: resource.id,
    ...resource.attributes,
  };
}

/**
 * List tasks with optional filters
 * @param userId - Filter by user ID
 * @param day - Filter by day (YYYY-MM-DD)
 * @param status - Filter by status (incomplete | complete)
 */
export async function listTasks(
  userId?: string,
  day?: string,
  status?: "incomplete" | "complete"
): Promise<Task[]> {
  // Build query params
  const params = new URLSearchParams();
  if (userId) params.append("userId", userId);
  if (day) params.append("day", day);
  if (status) params.append("status", status);

  const queryString = params.toString();
  const endpoint = queryString ? `/tasks?${queryString}` : "/tasks";

  const response = await get<JsonApiArrayResponse<TaskAttributes>>(endpoint);
  return response.data.map(flattenResource);
}

/**
 * Get a single task by ID
 * @param taskId - Task ID
 */
export async function getTask(taskId: string): Promise<Task> {
  const response = await get<JsonApiResponse<TaskAttributes>>(
    `/tasks/${taskId}`
  );
  return flattenResource(response.data);
}

/**
 * Create a new task
 * @param input - Task creation data
 */
export async function createTask(input: CreateTaskInput): Promise<Task> {
  const response = await post<JsonApiResponse<TaskAttributes>>("/tasks", {
    data: {
      type: "tasks",
      attributes: {
        day: input.day,
        title: input.title,
        description: input.description,
      },
    },
  });
  return flattenResource(response.data);
}

/**
 * Update an existing task
 * @param taskId - Task ID
 * @param input - Task update data
 */
export async function updateTask(
  taskId: string,
  input: UpdateTaskInput
): Promise<Task> {
  const response = await patch<JsonApiResponse<TaskAttributes>>(
    `/tasks/${taskId}`,
    {
      data: {
        type: "tasks",
        id: taskId,
        attributes: {
          ...(input.title !== undefined && { title: input.title }),
          ...(input.description !== undefined && {
            description: input.description,
          }),
          ...(input.day !== undefined && { day: input.day }),
          ...(input.status !== undefined && { status: input.status }),
        },
      },
    }
  );
  return flattenResource(response.data);
}

/**
 * Delete a task
 * @param taskId - Task ID
 */
export async function deleteTask(taskId: string): Promise<void> {
  await del(`/tasks/${taskId}`);
}

/**
 * Mark a task as complete
 * @param taskId - Task ID
 */
export async function completeTask(taskId: string): Promise<Task> {
  const response = await post<JsonApiResponse<TaskAttributes>>(
    `/tasks/${taskId}/complete`,
    {}
  );
  return flattenResource(response.data);
}

/**
 * Mark a task as incomplete (uncomplete)
 * @param taskId - Task ID
 */
export async function uncompleteTask(taskId: string): Promise<Task> {
  return updateTask(taskId, { status: "incomplete" });
}
