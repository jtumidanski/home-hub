/**
 * Reminders Service API
 *
 * Client for interacting with the svc-reminders service via the gateway.
 * Handles reminder CRUD operations for users.
 */

import { get, post, patch, del } from "./client";

/**
 * Reminder attributes (what's inside the JSON:API attributes field)
 */
export interface ReminderAttributes {
  name: string;
  description?: string;
  userId: string;
  householdId: string;
  remindAt: string; // ISO 8601 format
  snoozeCount: number;
  status: "active" | "snoozed" | "dismissed";
  createdAt: string; // ISO 8601 format
  dismissedAt?: string; // ISO 8601 format
  updatedAt: string; // ISO 8601 format
}

/**
 * Reminder interface (flattened for easier use in components)
 */
export interface Reminder {
  id: string;
  name: string;
  description?: string;
  userId: string;
  householdId: string;
  remindAt: string;
  snoozeCount: number;
  status: "active" | "snoozed" | "dismissed";
  createdAt: string;
  dismissedAt?: string;
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
 * Input for creating a new reminder
 */
export interface CreateReminderInput {
  userId: string;
  name: string;
  description?: string;
  remindAt: string; // ISO 8601 format
}

/**
 * Input for updating an existing reminder
 */
export interface UpdateReminderInput {
  name?: string;
  description?: string;
  remindAt?: string; // ISO 8601 format
  status?: "active" | "snoozed" | "dismissed";
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
 * List reminders with optional filters
 * @param userId - Filter by user ID
 * @param status - Filter by status (active | snoozed | dismissed)
 */
export async function listReminders(
  userId?: string,
  status?: "active" | "snoozed" | "dismissed"
): Promise<Reminder[]> {
  // Build query params
  const params = new URLSearchParams();
  if (userId) params.append("userId", userId);
  if (status) params.append("status", status);

  const queryString = params.toString();
  const endpoint = queryString ? `/reminders?${queryString}` : "/reminders";

  const response =
    await get<JsonApiArrayResponse<ReminderAttributes>>(endpoint);
  return response.data.map(flattenResource);
}

/**
 * Get a single reminder by ID
 * @param reminderId - Reminder ID
 */
export async function getReminder(reminderId: string): Promise<Reminder> {
  const response = await get<JsonApiResponse<ReminderAttributes>>(
    `/reminders/${reminderId}`
  );
  return flattenResource(response.data);
}

/**
 * Create a new reminder
 * @param input - Reminder creation data
 */
export async function createReminder(
  input: CreateReminderInput
): Promise<Reminder> {
  const response = await post<JsonApiResponse<ReminderAttributes>>(
    "/reminders",
    {
      data: {
        type: "reminders",
        attributes: {
          name: input.name,
          description: input.description,
          remindAt: input.remindAt,
        },
      },
    }
  );
  return flattenResource(response.data);
}

/**
 * Update an existing reminder
 * @param reminderId - Reminder ID
 * @param input - Reminder update data
 */
export async function updateReminder(
  reminderId: string,
  input: UpdateReminderInput
): Promise<Reminder> {
  const response = await patch<JsonApiResponse<ReminderAttributes>>(
    `/reminders/${reminderId}`,
    {
      data: {
        type: "reminders",
        id: reminderId,
        attributes: {
          ...(input.name !== undefined && { name: input.name }),
          ...(input.description !== undefined && {
            description: input.description,
          }),
          ...(input.remindAt !== undefined && { remindAt: input.remindAt }),
          ...(input.status !== undefined && { status: input.status }),
        },
      },
    }
  );
  return flattenResource(response.data);
}

/**
 * Delete a reminder
 * @param reminderId - Reminder ID
 */
export async function deleteReminder(reminderId: string): Promise<void> {
  await del(`/reminders/${reminderId}`);
}

/**
 * Snooze a reminder for a specified duration
 * @param reminderId - Reminder ID
 * @param duration - Duration in minutes to snooze
 */
export async function snoozeReminder(
  reminderId: string,
  duration: number
): Promise<Reminder> {
  const response = await post<JsonApiResponse<ReminderAttributes>>(
    `/reminders/${reminderId}/snooze`,
    {
      data: {
        type: "reminders",
        id: reminderId,
        attributes: {
          remindAt: new Date(Date.now() + duration * 60 * 1000).toISOString(),
        },
      },
    }
  );
  return flattenResource(response.data);
}

/**
 * Dismiss a reminder
 * @param reminderId - Reminder ID
 */
export async function dismissReminder(reminderId: string): Promise<Reminder> {
  const response = await post<JsonApiResponse<ReminderAttributes>>(
    `/reminders/${reminderId}/dismiss`,
    {}
  );
  return flattenResource(response.data);
}
