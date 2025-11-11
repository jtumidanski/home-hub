/**
 * Households Service API
 *
 * Client for interacting with the households endpoints via the gateway.
 */

import { get } from "./client";
import { JsonApiResponse, JsonApiArrayResponse, JsonApiResource } from "./users";

/**
 * Household attributes (JSON:API)
 */
export interface HouseholdAttributes {
  name: string;
  timezone: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Household interface (flattened)
 */
export interface Household {
  id: string;
  name: string;
  timezone: string;
  createdAt: string;
  updatedAt: string;
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
 * Get a single household by ID
 */
export async function getHousehold(id: string): Promise<Household> {
  const response = await get<JsonApiResponse<HouseholdAttributes>>(
    `/households/${id}`
  );
  return flattenResource(response.data);
}

/**
 * List all households
 */
export async function listHouseholds(): Promise<Household[]> {
  const response = await get<JsonApiArrayResponse<HouseholdAttributes>>(
    "/households"
  );
  return response.data.map(flattenResource);
}
