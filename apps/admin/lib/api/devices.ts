/**
 * Devices Service API
 *
 * Client for interacting with the svc-devices service via the gateway.
 * Handles device management and device preferences.
 */

import { get, post, patch, put, del } from "./client";

/**
 * Device type enum
 */
export type DeviceType = "kiosk";

/**
 * Theme preference
 */
export type Theme = "light" | "dark";

/**
 * Temperature unit preference
 */
export type TemperatureUnit = "household" | "F" | "C";

/**
 * Device attributes
 */
export interface Device {
  id: string;
  name: string;
  type: DeviceType;
  householdId: string;
  created_at: string;
  updated_at: string;
}

/**
 * Device preferences attributes
 */
export interface DevicePreferences {
  id: string;
  deviceId: string;
  theme: Theme;
  temperatureUnit: TemperatureUnit;
  created_at: string;
  updated_at: string;
}

/**
 * JSON:API response wrapper for a single device
 */
export interface JsonApiDeviceResponse {
  data: {
    type: string;
    id: string;
    attributes: Device;
  };
}

/**
 * JSON:API response wrapper for multiple devices
 */
export interface JsonApiDevicesResponse {
  data: Array<{
    type: string;
    id: string;
    attributes: Device;
  }>;
}

/**
 * JSON:API response wrapper for device preferences
 */
export interface JsonApiDevicePreferencesResponse {
  data: {
    type: string;
    id: string;
    attributes: DevicePreferences;
  };
}

/**
 * Request body for creating a device
 * Note: householdId is not included - it comes from auth context
 */
export interface CreateDeviceRequest {
  name: string;
  type: DeviceType;
}

/**
 * Request body for updating a device
 */
export interface UpdateDeviceRequest {
  name?: string;
}

/**
 * Request body for updating device preferences
 */
export interface UpdateDevicePreferencesRequest {
  theme?: Theme;
  temperatureUnit?: TemperatureUnit;
}

/**
 * Get all devices
 *
 * @returns List of all devices for the current household
 * @throws {ApiError} If the request fails
 */
export async function getDevices(): Promise<Device[]> {
  const response = await get<JsonApiDevicesResponse>("/devices");
  return response.data.map((item) => ({
    ...item.attributes,
    id: item.id, // Merge JSON:API id with attributes
  }));
}

/**
 * Get a single device by ID
 *
 * @param deviceId - The UUID of the device
 * @returns Device data
 * @throws {ApiError} If the request fails or device not found
 */
export async function getDevice(deviceId: string): Promise<Device> {
  const response = await get<JsonApiDeviceResponse>(`/devices/${deviceId}`);
  return {
    ...response.data.attributes,
    id: response.data.id, // Merge JSON:API id with attributes
  };
}

/**
 * Create a new device
 *
 * @param data - Device creation data
 * @returns Created device
 * @throws {ApiError} If the request fails or validation errors occur
 */
export async function createDevice(
  data: CreateDeviceRequest
): Promise<Device> {
  const response = await post<JsonApiDeviceResponse>("/devices", {
    data: {
      type: "devices",
      attributes: data,
    },
  });
  return {
    ...response.data.attributes,
    id: response.data.id, // Merge JSON:API id with attributes
  };
}

/**
 * Update a device
 *
 * @param deviceId - The UUID of the device
 * @param data - Device update data
 * @returns Updated device
 * @throws {ApiError} If the request fails or device not found
 */
export async function updateDevice(
  deviceId: string,
  data: UpdateDeviceRequest
): Promise<Device> {
  const response = await patch<JsonApiDeviceResponse>(
    `/devices/${deviceId}`,
    {
      data: {
        type: "devices",
        id: deviceId,
        attributes: data,
      },
    }
  );
  return {
    ...response.data.attributes,
    id: response.data.id, // Merge JSON:API id with attributes
  };
}

/**
 * Delete a device
 *
 * @param deviceId - The UUID of the device
 * @returns void
 * @throws {ApiError} If the request fails or device not found
 */
export async function deleteDevice(deviceId: string): Promise<void> {
  await del<void>(`/devices/${deviceId}`);
}

/**
 * Get device preferences
 *
 * @param deviceId - The UUID of the device
 * @returns Device preferences
 * @throws {ApiError} If the request fails or preferences not found
 */
export async function getDevicePreferences(
  deviceId: string
): Promise<DevicePreferences> {
  const response = await get<JsonApiDevicePreferencesResponse>(
    `/devices/${deviceId}/preferences`
  );
  return response.data.attributes;
}

/**
 * Update device preferences
 *
 * @param deviceId - The UUID of the device
 * @param data - Preferences update data
 * @returns Updated preferences
 * @throws {ApiError} If the request fails
 */
export async function updateDevicePreferences(
  deviceId: string,
  data: UpdateDevicePreferencesRequest
): Promise<DevicePreferences> {
  const response = await put<JsonApiDevicePreferencesResponse>(
    `/devices/${deviceId}/preferences`,
    {
      data: {
        type: "device_preferences",
        attributes: data,
      },
    }
  );
  return response.data.attributes;
}

/**
 * Format device type for display
 *
 * @param type - Device type
 * @returns Formatted device type string
 */
export function formatDeviceType(type: DeviceType): string {
  const typeMap: Record<DeviceType, string> = {
    kiosk: "Kiosk",
  };
  return typeMap[type] || type;
}

/**
 * Format theme for display
 *
 * @param theme - Theme preference
 * @returns Formatted theme string
 */
export function formatTheme(theme: Theme): string {
  const themeMap: Record<Theme, string> = {
    light: "Light",
    dark: "Dark",
  };
  return themeMap[theme] || theme;
}

/**
 * Format temperature unit for display
 *
 * @param unit - Temperature unit preference
 * @returns Formatted temperature unit string
 */
export function formatTemperatureUnit(unit: TemperatureUnit): string {
  const unitMap: Record<TemperatureUnit, string> = {
    household: "Household Default",
    F: "Fahrenheit (°F)",
    C: "Celsius (°C)",
  };
  return unitMap[unit] || unit;
}
