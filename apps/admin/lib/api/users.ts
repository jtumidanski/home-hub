/**
 * Users Service API
 *
 * Client for interacting with the svc-users service via the gateway.
 * Handles users, households, and related operations.
 */

import { get, post, del } from "./client";

/**
 * JSON:API count response shape from the backend
 * Following JSON:API spec with meta field
 */
export interface CountResponse {
  meta: {
    count: number;
  };
}

/**
 * User attributes (what's inside the JSON:API attributes field)
 */
export interface UserAttributes {
  email: string;
  displayName: string;
  provider: string;
  householdId?: string;
  roles?: string[];
  createdAt: string; // ISO 8601 format
  updatedAt: string; // ISO 8601 format
}

/**
 * User interface (flattened for easier use in components)
 */
export interface User {
  id: string;
  email: string;
  displayName: string;
  provider: string;
  householdId?: string;
  roles?: string[];
  createdAt: string;
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
 * Role attributes
 */
export interface RoleAttributes {
  role: string;
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
 * Get total count of households
 */
export async function getHouseholdCount(): Promise<number> {
  const response = await get<CountResponse>("/households/count");
  return response.meta.count;
}

/**
 * Get total count of users
 */
export async function getUserCount(): Promise<number> {
  const response = await get<CountResponse>("/users/count");
  return response.meta.count;
}

/**
 * List all users
 */
export async function listUsers(): Promise<User[]> {
  const response = await get<JsonApiArrayResponse<UserAttributes>>("/users");
  return response.data.map(flattenResource);
}

/**
 * Get a single user by ID
 */
export async function getUser(id: string): Promise<User> {
  const response = await get<JsonApiResponse<UserAttributes>>(`/users/${id}`);
  return flattenResource(response.data);
}

/**
 * Get all roles for a user
 */
export async function getUserRoles(userId: string): Promise<string[]> {
  const response = await get<JsonApiArrayResponse<RoleAttributes>>(
    `/users/${userId}/roles`
  );
  return response.data.map((r) => r.attributes.role);
}

/**
 * Add a role to a user
 */
export async function addUserRole(userId: string, role: string): Promise<void> {
  await post(`/users/${userId}/roles`, {
    data: {
      type: "roles",
      attributes: {
        role,
      },
    },
  });
}

/**
 * Remove a role from a user
 */
export async function removeUserRole(
  userId: string,
  role: string
): Promise<void> {
  await del(`/users/${userId}/roles/${role}`);
}
