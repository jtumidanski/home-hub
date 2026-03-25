export interface ApiResponse<T = unknown> {
  data: T;
  included?: Array<Record<string, unknown>>;
}

export interface ApiListResponse<T = unknown> extends ApiResponse<T[]> {}

export interface ApiSingleResponse<T = unknown> extends ApiResponse<T> {}

export interface JsonApiError {
  status: string;
  code?: string;
  title: string;
  detail?: string;
}

export interface JsonApiErrorResponse {
  errors: JsonApiError[];
}

// --- Type guards (F11) ---

export function isApiResponse<T>(
  value: unknown,
): value is ApiResponse<T> {
  return (
    typeof value === "object" &&
    value !== null &&
    "data" in value &&
    !("errors" in value)
  );
}

export function isApiErrorResponse(
  value: unknown,
): value is JsonApiErrorResponse {
  return (
    typeof value === "object" &&
    value !== null &&
    "errors" in value &&
    Array.isArray((value as JsonApiErrorResponse).errors)
  );
}

// --- Error hierarchy (F12) ---

export type ApiErrorCategory =
  | "validation"
  | "not-found"
  | "auth"
  | "server";

export interface CategorizedApiError {
  category: ApiErrorCategory;
  source: JsonApiError;
}

export function categorizeApiError(error: JsonApiError): CategorizedApiError {
  const status = parseInt(error.status, 10);
  let category: ApiErrorCategory;
  if (status === 401 || status === 403) {
    category = "auth";
  } else if (status === 404) {
    category = "not-found";
  } else if (status === 422 || status === 400) {
    category = "validation";
  } else {
    category = "server";
  }
  return { category, source: error };
}

// --- Result pattern (F13) ---

export type Result<T, E = JsonApiErrorResponse> =
  | { success: true; data: T }
  | { success: false; error: E };
