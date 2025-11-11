/**
 * Users Service API
 *
 * Client for interacting with the svc-users service via the gateway.
 * Handles users, households, and related operations.
 */

import { get } from "./client";

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
