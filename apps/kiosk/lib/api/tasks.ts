/**
 * Tasks Service API
 *
 * Client for interacting with the svc-tasks service via the gateway.
 * Handles task retrieval and completion for the kiosk dashboard.
 */

import { get, post } from "./client";

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
 * Get tasks with optional filters
 * @param day - Filter by day (YYYY-MM-DD format)
 * @param status - Filter by status (incomplete | complete)
 */
export async function getTasks(
  day?: string,
  status?: "incomplete" | "complete"
): Promise<Task[]> {
  // Build query params
  const params = new URLSearchParams();
  if (day) params.append("day", day);
  if (status) params.append("status", status);

  const queryString = params.toString();
  const endpoint = queryString ? `/tasks?${queryString}` : "/tasks";

  const response = await get<JsonApiArrayResponse<TaskAttributes>>(endpoint);
  return response.data.map(flattenResource);
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
 * Helper function to determine if a task is overdue
 * A task is overdue if it's incomplete and the day is before today
 */
export function isTaskOverdue(task: Task): boolean {
  if (task.status === "complete") return false;

  const today = new Date().toISOString().split('T')[0];
  return task.day < today;
}

/**
 * Helper function to check if a task is for today
 */
export function isTaskToday(task: Task): boolean {
  const today = new Date().toISOString().split('T')[0];
  return task.day === today;
}

/**
 * Helper function to check if a task is for tomorrow
 */
export function isTaskTomorrow(task: Task): boolean {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const tomorrowStr = tomorrow.toISOString().split('T')[0];
  return task.day === tomorrowStr;
}
