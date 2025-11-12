/**
 * Households Service API
 *
 * Client for interacting with the households endpoints via the gateway.
 */

import { get, post, patch, del } from "./client";
import { JsonApiResponse, JsonApiArrayResponse, JsonApiResource, User, UserAttributes } from "./users";

/**
 * Household attributes (JSON:API)
 */
export interface HouseholdAttributes {
  name: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
  created_at: string;
  updated_at: string;
}

/**
 * Household interface (flattened)
 */
export interface Household {
  id: string;
  name: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Request payload for creating a household
 */
export interface CreateHouseholdRequest {
  name: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
}

/**
 * Request payload for updating a household
 */
export interface UpdateHouseholdRequest {
  name?: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
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
 * Helper to flatten household attributes to Household
 */
function flattenHousehold(resource: JsonApiResource<HouseholdAttributes>): Household {
  return {
    id: resource.id,
    name: resource.attributes.name,
    latitude: resource.attributes.latitude,
    longitude: resource.attributes.longitude,
    timezone: resource.attributes.timezone,
    createdAt: resource.attributes.created_at,
    updatedAt: resource.attributes.updated_at,
  };
}

/**
 * Get a single household by ID
 */
export async function getHousehold(id: string): Promise<Household> {
  const response = await get<JsonApiResponse<HouseholdAttributes>>(
    `/households/${id}`
  );
  return flattenHousehold(response.data);
}

/**
 * List all households
 */
export async function listHouseholds(): Promise<Household[]> {
  const response = await get<JsonApiArrayResponse<HouseholdAttributes>>(
    "/households"
  );
  return response.data.map(flattenHousehold);
}

/**
 * Create a new household
 */
export async function createHousehold(
  request: CreateHouseholdRequest
): Promise<Household> {
  const response = await post<JsonApiResponse<HouseholdAttributes>>(
    "/households",
    {
      data: {
        type: "households",
        attributes: request,
      },
    }
  );
  return flattenHousehold(response.data);
}

/**
 * Update an existing household
 */
export async function updateHousehold(
  id: string,
  request: UpdateHouseholdRequest
): Promise<Household> {
  const response = await patch<JsonApiResponse<HouseholdAttributes>>(
    `/households/${id}`,
    {
      data: {
        type: "households",
        id,
        attributes: request,
      },
    }
  );
  return flattenHousehold(response.data);
}

/**
 * Delete a household
 * Note: Household must have no associated users
 */
export async function deleteHousehold(id: string): Promise<void> {
  await del(`/households/${id}`);
}

/**
 * Get all users in a household
 */
export async function getHouseholdUsers(householdId: string): Promise<User[]> {
  const response = await get<JsonApiArrayResponse<UserAttributes>>(
    `/households/${householdId}/users`
  );
  return response.data.map((resource) => ({
    id: resource.id,
    email: resource.attributes.email,
    displayName: resource.attributes.displayName,
    provider: resource.attributes.provider,
    householdId: resource.attributes.householdId,
    roles: resource.attributes.roles,
    createdAt: resource.attributes.createdAt,
    updatedAt: resource.attributes.updatedAt,
  }));
}
