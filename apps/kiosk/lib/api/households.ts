/**
 * Households Service API
 *
 * Client for interacting with the households endpoints via the gateway.
 */

import { get } from "./client";

/**
 * JSON:API resource structure
 */
export interface JsonApiResource<T> {
  type: string;
  id: string;
  attributes: T;
}

export interface JsonApiResponse<T> {
  data: JsonApiResource<T>;
}

export interface JsonApiArrayResponse<T> {
  data: JsonApiResource<T>[];
}

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
 * User attributes (JSON:API)
 */
export interface UserAttributes {
  email: string;
  displayName: string;
  provider: string;
  householdId?: string;
  roles: string[];
  createdAt: string;
  updatedAt: string;
}

/**
 * User interface (flattened)
 */
export interface User {
  id: string;
  email: string;
  displayName: string;
  provider: 'google' | 'github';
  householdId?: string;
  roles: string[];
  createdAt: string;
  updatedAt: string;
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
 * Helper to flatten user attributes to User
 */
function flattenUser(resource: JsonApiResource<UserAttributes>): User {
  return {
    id: resource.id,
    email: resource.attributes.email,
    displayName: resource.attributes.displayName,
    provider: resource.attributes.provider as 'google' | 'github',
    householdId: resource.attributes.householdId,
    roles: resource.attributes.roles,
    createdAt: resource.attributes.createdAt,
    updatedAt: resource.attributes.updatedAt,
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
 * Get all users in a household
 */
export async function getHouseholdUsers(householdId: string): Promise<User[]> {
  const response = await get<JsonApiArrayResponse<UserAttributes>>(
    `/households/${householdId}/users`
  );
  return response.data.map(flattenUser);
}
