import { ApiRequestError } from "./client";

export interface AppError {
  message: string;
  status?: number;
  type: "network" | "auth" | "validation" | "server" | "unknown";
}

export function createErrorFromUnknown(
  error: unknown,
  fallbackMessage = "An unexpected error occurred"
): AppError {
  if (error instanceof ApiRequestError) {
    return {
      message: error.message || fallbackMessage,
      status: error.status,
      type: classifyStatus(error.status),
    };
  }

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

function classifyStatus(status: number): AppError["type"] {
  if (status === 401) return "auth";
  if (status === 400 || status === 422) return "validation";
  if (status >= 500) return "server";
  return "unknown";
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
