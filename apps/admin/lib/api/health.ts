/**
 * Health Check API Client
 *
 * Provides functions for querying service health endpoints and
 * aggregating health status across all microservices.
 */

import { get } from "./client";

// ==========================================
// Type Definitions
// ==========================================

/**
 * Health status values for services and individual checks
 */
export type HealthStatus = "healthy" | "degraded" | "unhealthy" | "unknown";

/**
 * Result of a single health check (e.g., database, cache, etc.)
 */
export interface HealthCheckDetail {
  status: HealthStatus;
  response_time_ms?: number;
  error?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Complete health response from a service
 */
export interface HealthCheckResponse {
  status: HealthStatus;
  service: string;
  version: string;
  uptime_seconds: number;
  timestamp: string;
  checks: Record<string, HealthCheckDetail>;
}

/**
 * Health information for a single service (includes error handling)
 */
export interface ServiceHealth {
  service: string;
  health: HealthCheckResponse | null;
  error?: string;
  available: boolean;
  responseTime?: number;
}

// ==========================================
// Service List
// ==========================================

/**
 * List of all microservices that expose health endpoints
 */
export const SERVICES = ["users", "tasks", "reminders", "weather", "devices"] as const;

export type ServiceName = (typeof SERVICES)[number];

// ==========================================
// API Functions
// ==========================================

/**
 * Fetch health status for a single service
 * @param service - Service name (users, tasks, reminders, weather, devices)
 * @returns Health check response
 * @throws ApiError if the request fails
 */
export async function getServiceHealth(
  service: ServiceName
): Promise<HealthCheckResponse> {
  return get<HealthCheckResponse>(`/health/${service}`);
}

/**
 * Fetch health status for a single service (with error handling)
 * Returns 'unknown' status instead of throwing on error
 * @param service - Service name
 * @returns Service health with availability information
 */
export async function getServiceHealthSafe(
  service: ServiceName
): Promise<ServiceHealth> {
  const startTime = performance.now();

  try {
    const health = await getServiceHealth(service);
    const responseTime = performance.now() - startTime;

    return {
      service,
      health,
      available: true,
      responseTime,
    };
  } catch (error) {
    const responseTime = performance.now() - startTime;

    return {
      service,
      health: null,
      error: error instanceof Error ? error.message : "Unknown error",
      available: false,
      responseTime,
    };
  }
}

/**
 * Fetch health status for all services in parallel
 * @returns Array of service health information
 */
export async function getAllServiceHealth(): Promise<ServiceHealth[]> {
  const healthPromises = SERVICES.map((service) =>
    getServiceHealthSafe(service)
  );

  return Promise.all(healthPromises);
}

// ==========================================
// Utility Functions
// ==========================================

/**
 * Determine aggregate health status from multiple services
 * @param services - Array of service health data
 * @returns Overall system health status
 */
export function aggregateHealthStatus(services: ServiceHealth[]): HealthStatus {
  // If any service is unavailable, system is unknown
  if (services.some((s) => !s.available)) {
    return "unknown";
  }

  // Get all service statuses
  const statuses = services
    .filter((s) => s.health !== null)
    .map((s) => s.health!.status);

  // If no statuses available, system is unknown
  if (statuses.length === 0) {
    return "unknown";
  }

  // If any service is unhealthy, system is unhealthy
  if (statuses.includes("unhealthy")) {
    return "unhealthy";
  }

  // If any service is degraded, system is degraded
  if (statuses.includes("degraded")) {
    return "degraded";
  }

  // Otherwise, system is healthy
  return "healthy";
}

/**
 * Get Tailwind color class for a health status
 * @param status - Health status
 * @returns Tailwind color class (e.g., "text-green-600")
 */
export function getStatusColor(status: HealthStatus): string {
  switch (status) {
    case "healthy":
      return "text-green-600 dark:text-green-400";
    case "degraded":
      return "text-yellow-600 dark:text-yellow-400";
    case "unhealthy":
      return "text-red-600 dark:text-red-400";
    case "unknown":
      return "text-neutral-600 dark:text-neutral-400";
  }
}

/**
 * Get background color class for status badges
 * @param status - Health status
 * @returns Tailwind background color class
 */
export function getStatusBgColor(status: HealthStatus): string {
  switch (status) {
    case "healthy":
      return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300";
    case "degraded":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300";
    case "unhealthy":
      return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300";
    case "unknown":
      return "bg-neutral-100 text-neutral-800 dark:bg-neutral-800 dark:text-neutral-300";
  }
}

/**
 * Get human-readable message for a health status
 * @param status - Health status
 * @param serviceCount - Number of services (optional, for aggregate messages)
 * @returns Status message
 */
export function getStatusMessage(
  status: HealthStatus,
  serviceCount?: number
): string {
  const count = serviceCount ? ` (${serviceCount} services)` : "";

  switch (status) {
    case "healthy":
      return `All systems operational${count}`;
    case "degraded":
      return `Some services degraded${count}`;
    case "unhealthy":
      return `Service issues detected${count}`;
    case "unknown":
      return `Unable to fetch status${count}`;
  }
}

/**
 * Get icon name for a health status (Lucide React icon names)
 * @param status - Health status
 * @returns Icon name
 */
export function getStatusIconName(status: HealthStatus): string {
  switch (status) {
    case "healthy":
      return "CheckCircle2";
    case "degraded":
      return "AlertTriangle";
    case "unhealthy":
      return "XCircle";
    case "unknown":
      return "HelpCircle";
  }
}

/**
 * Format uptime seconds to human-readable string
 * @param seconds - Uptime in seconds
 * @returns Formatted uptime string (e.g., "2d 3h 15m")
 */
export function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  const parts: string[] = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);

  return parts.length > 0 ? parts.join(" ") : "< 1m";
}

/**
 * Format relative time (e.g., "2 seconds ago")
 * @param timestamp - ISO 8601 timestamp or Date object
 * @returns Relative time string
 */
export function formatRelativeTime(timestamp: string | Date): string {
  const now = new Date();
  const then = typeof timestamp === "string" ? new Date(timestamp) : timestamp;
  const secondsAgo = Math.floor((now.getTime() - then.getTime()) / 1000);

  if (secondsAgo < 10) return "just now";
  if (secondsAgo < 60) return `${secondsAgo} seconds ago`;

  const minutesAgo = Math.floor(secondsAgo / 60);
  if (minutesAgo < 60) return `${minutesAgo} minute${minutesAgo === 1 ? "" : "s"} ago`;

  const hoursAgo = Math.floor(minutesAgo / 60);
  if (hoursAgo < 24) return `${hoursAgo} hour${hoursAgo === 1 ? "" : "s"} ago`;

  const daysAgo = Math.floor(hoursAgo / 24);
  return `${daysAgo} day${daysAgo === 1 ? "" : "s"} ago`;
}
