/**
 * Reminders Service API
 *
 * Client for interacting with the svc-reminders service via the gateway.
 * Handles reminder retrieval, snoozing, and dismissing for the kiosk dashboard.
 */

import { get, post } from "./client";

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
 * Get all reminders for the current user
 */
export async function getReminders(): Promise<Reminder[]> {
  const response = await get<JsonApiArrayResponse<ReminderAttributes>>("/reminders");
  return response.data.map(flattenResource);
}

/**
 * Snooze a reminder by updating its remind time
 * @param id - Reminder ID
 * @param duration - Duration in milliseconds to snooze
 */
export async function snoozeReminder(id: string, duration: number): Promise<Reminder> {
  const newRemindAt = new Date(Date.now() + duration).toISOString();
  const response = await post<JsonApiResponse<ReminderAttributes>>(
    `/reminders/${id}/snooze`,
    { remindAt: newRemindAt }
  );
  return flattenResource(response.data);
}

/**
 * Dismiss a reminder
 * @param id - Reminder ID
 */
export async function dismissReminder(id: string): Promise<Reminder> {
  const response = await post<JsonApiResponse<ReminderAttributes>>(
    `/reminders/${id}/dismiss`,
    {}
  );
  return flattenResource(response.data);
}
