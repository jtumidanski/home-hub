export interface AppError {
  message: string;
  status?: number;
  type: "network" | "auth" | "validation" | "server" | "unknown";
}

export function createErrorFromUnknown(
  error: unknown,
  fallbackMessage = "An unexpected error occurred"
): AppError {
  if (error instanceof Error) {
    return {
      message: error.message || fallbackMessage,
      type: classifyErrorMessage(error.message),
    };
  }

  if (typeof error === "string") {
    return { message: error, type: "unknown" };
  }

  return { message: fallbackMessage, type: "unknown" };
}

function classifyErrorMessage(message: string): AppError["type"] {
  const lower = message.toLowerCase();
  if (lower.includes("network") || lower.includes("fetch")) return "network";
  if (lower.includes("unauthorized") || lower.includes("401")) return "auth";
  if (lower.includes("validation") || lower.includes("422")) return "validation";
  if (lower.includes("500") || lower.includes("server")) return "server";
  return "unknown";
}

export function isRetryableError(error: unknown): boolean {
  const appError = createErrorFromUnknown(error);
  return appError.type === "network" || appError.type === "server";
}

export function requiresAuthentication(error: unknown): boolean {
  const appError = createErrorFromUnknown(error);
  return appError.type === "auth";
}

export function getErrorMessage(error: unknown, fallback?: string): string {
  return createErrorFromUnknown(error, fallback).message;
}
