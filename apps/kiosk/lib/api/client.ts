/**
 * API Client for Home Hub Kiosk
 *
 * Communicates with the backend gateway at /api/*
 * Gateway injects tenant/household context via headers
 *
 * Architecture Notes:
 * - Never include tenant_id or household_id in request bodies
 * - All authentication handled by gateway (Google/GitHub OAuth)
 * - Gateway returns tenant/household context in responses
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "/api";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
    public details?: unknown
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export interface ApiResponse<T> {
  data: T;
  meta?: {
    page?: number;
    perPage?: number;
    total?: number;
  };
}

/**
 * Base API client using fetch
 * Automatically handles:
 * - Authentication headers (session cookies)
 * - JSON content-type
 * - Error responses
 * - Type-safe responses
 */
export async function apiClient<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  try {
    const response = await fetch(url, {
      ...options,
      credentials: 'include', // Include cookies for session authentication
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
    });

    // Handle HTTP errors
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new ApiError(
        errorData.message || `HTTP ${response.status}: ${response.statusText}`,
        response.status,
        errorData.code,
        errorData.details
      );
    }

    // Handle 204 No Content (no response body)
    if (response.status === 204) {
      return undefined as T;
    }

    // Parse successful response
    const data = await response.json();
    return data as T;
  } catch (error) {
    // Re-throw ApiError as-is
    if (error instanceof ApiError) {
      throw error;
    }

    // Handle network errors
    if (error instanceof TypeError) {
      throw new ApiError("Network error: Unable to reach API", 0);
    }

    // Handle unknown errors
    throw new ApiError(
      error instanceof Error ? error.message : "Unknown error occurred",
      0
    );
  }
}

/**
 * GET request helper
 */
export async function get<T>(endpoint: string): Promise<T> {
  return apiClient<T>(endpoint, {
    method: "GET",
  });
}

/**
 * POST request helper
 */
export async function post<T>(endpoint: string, body?: unknown): Promise<T> {
  return apiClient<T>(endpoint, {
    method: "POST",
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * PUT request helper
 */
export async function put<T>(endpoint: string, body?: unknown): Promise<T> {
  return apiClient<T>(endpoint, {
    method: "PUT",
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * PATCH request helper
 */
export async function patch<T>(endpoint: string, body?: unknown): Promise<T> {
  return apiClient<T>(endpoint, {
    method: "PATCH",
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * DELETE request helper
 */
export async function del<T>(endpoint: string): Promise<T> {
  return apiClient<T>(endpoint, {
    method: "DELETE",
  });
}
