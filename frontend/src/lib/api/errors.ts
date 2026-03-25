import { ApiRequestError } from "./client";

export type ErrorType =
  | "network"
  | "auth"
  | "not-found"
  | "validation"
  | "conflict"
  | "rate-limited"
  | "server"
  | "unknown";

export interface AppError {
  message: string;
  status?: number;
  type: ErrorType;
}

export interface NetworkError extends AppError {
  type: "network";
}

export interface AuthError extends AppError {
  type: "auth";
  status: number;
}

export interface NotFoundError extends AppError {
  type: "not-found";
  status: number;
}

export interface ValidationError extends AppError {
  type: "validation";
  status: number;
}

export interface ConflictError extends AppError {
  type: "conflict";
  status: number;
}

export interface RateLimitedError extends AppError {
  type: "rate-limited";
  status: number;
}

export interface ServerError extends AppError {
  type: "server";
}

export function isNetworkError(error: unknown): error is NetworkError {
  return isAppError(error) && error.type === "network";
}

export function isAuthError(error: unknown): error is AuthError {
  return isAppError(error) && error.type === "auth";
}

export function isNotFoundError(error: unknown): error is NotFoundError {
  return isAppError(error) && error.type === "not-found";
}

export function isValidationError(error: unknown): error is ValidationError {
  return isAppError(error) && error.type === "validation";
}

export function isServerError(error: unknown): error is ServerError {
  return isAppError(error) && error.type === "server";
}

function isAppError(error: unknown): error is AppError {
  return (
    typeof error === "object" &&
    error !== null &&
    "type" in error &&
    "message" in error &&
    typeof (error as AppError).type === "string" &&
    typeof (error as AppError).message === "string"
  );
}

export function transformError(error: unknown): AppError {
  return createErrorFromUnknown(error);
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

export function classifyStatus(status: number): ErrorType {
  if (status === 401 || status === 403) return "auth";
  if (status === 404) return "not-found";
  if (status === 400 || status === 422) return "validation";
  if (status === 409) return "conflict";
  if (status === 429) return "rate-limited";
  if (status >= 500) return "server";
  return "unknown";
}

function classifyErrorMessage(message: string): ErrorType {
  const lower = message.toLowerCase();
  if (lower.includes("network") || lower.includes("fetch")) return "network";
  if (lower.includes("unauthorized") || lower.includes("401")) return "auth";
  if (lower.includes("validation") || lower.includes("422")) return "validation";
  if (lower.includes("500") || lower.includes("server")) return "server";
  return "unknown";
}

export function isRetryableError(error: unknown): boolean {
  const appError = createErrorFromUnknown(error);
  return appError.type === "network" || appError.type === "server" || appError.type === "rate-limited";
}

export function requiresAuthentication(error: unknown): boolean {
  const appError = createErrorFromUnknown(error);
  return appError.type === "auth";
}

export function getErrorMessage(error: unknown, fallback?: string): string {
  return createErrorFromUnknown(error, fallback).message;
}
