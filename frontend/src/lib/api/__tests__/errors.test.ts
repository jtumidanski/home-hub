import { describe, it, expect } from "vitest";
import { ApiRequestError } from "../client";
import {
  createErrorFromUnknown,
  isRetryableError,
  requiresAuthentication,
  getErrorMessage,
} from "../errors";

describe("createErrorFromUnknown", () => {
  it("extracts message from Error instance", () => {
    const result = createErrorFromUnknown(new Error("something broke"));
    expect(result.message).toBe("something broke");
  });

  it("classifies network errors by message", () => {
    const result = createErrorFromUnknown(new Error("network error"));
    expect(result.type).toBe("network");
  });

  it("classifies fetch errors as network", () => {
    const result = createErrorFromUnknown(new Error("Failed to fetch"));
    expect(result.type).toBe("network");
  });

  it("classifies unauthorized errors as auth by message", () => {
    const result = createErrorFromUnknown(new Error("unauthorized"));
    expect(result.type).toBe("auth");
  });

  it("classifies 401 as auth by message", () => {
    const result = createErrorFromUnknown(new Error("Request failed: 401"));
    expect(result.type).toBe("auth");
  });

  it("classifies validation errors by message", () => {
    const result = createErrorFromUnknown(new Error("validation failed"));
    expect(result.type).toBe("validation");
  });

  it("classifies server errors by message", () => {
    const result = createErrorFromUnknown(new Error("500 internal server error"));
    expect(result.type).toBe("server");
  });

  it("returns unknown for unrecognized errors", () => {
    const result = createErrorFromUnknown(new Error("something weird"));
    expect(result.type).toBe("unknown");
  });

  it("handles string errors", () => {
    const result = createErrorFromUnknown("plain string error");
    expect(result.message).toBe("plain string error");
    expect(result.type).toBe("unknown");
  });

  it("uses fallback message for non-Error/non-string", () => {
    const result = createErrorFromUnknown(42, "custom fallback");
    expect(result.message).toBe("custom fallback");
  });

  it("uses default fallback when none provided", () => {
    const result = createErrorFromUnknown(null);
    expect(result.message).toBe("An unexpected error occurred");
  });

  // ApiRequestError (structured status) tests
  it("classifies ApiRequestError 401 as auth with status", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Unauthorized", 401));
    expect(result.type).toBe("auth");
    expect(result.status).toBe(401);
  });

  it("classifies ApiRequestError 400 as validation with status", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Bad request", 400));
    expect(result.type).toBe("validation");
    expect(result.status).toBe(400);
  });

  it("classifies ApiRequestError 422 as validation with status", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Unprocessable", 422));
    expect(result.type).toBe("validation");
    expect(result.status).toBe(422);
  });

  it("classifies ApiRequestError 500 as server with status", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Internal server error", 500));
    expect(result.type).toBe("server");
    expect(result.status).toBe(500);
  });

  it("classifies ApiRequestError 502 as server with status", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Bad gateway", 502));
    expect(result.type).toBe("server");
    expect(result.status).toBe(502);
  });

  it("classifies ApiRequestError 404 as unknown with status", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Not found", 404));
    expect(result.type).toBe("unknown");
    expect(result.status).toBe(404);
  });

  it("preserves message from ApiRequestError", () => {
    const result = createErrorFromUnknown(new ApiRequestError("Custom detail", 422));
    expect(result.message).toBe("Custom detail");
  });
});

describe("isRetryableError", () => {
  it("returns true for network errors", () => {
    expect(isRetryableError(new Error("network timeout"))).toBe(true);
  });

  it("returns true for server errors", () => {
    expect(isRetryableError(new Error("500 server error"))).toBe(true);
  });

  it("returns true for ApiRequestError 500", () => {
    expect(isRetryableError(new ApiRequestError("Server error", 500))).toBe(true);
  });

  it("returns false for ApiRequestError 401", () => {
    expect(isRetryableError(new ApiRequestError("Unauthorized", 401))).toBe(false);
  });

  it("returns false for auth errors", () => {
    expect(isRetryableError(new Error("unauthorized"))).toBe(false);
  });

  it("returns false for validation errors", () => {
    expect(isRetryableError(new Error("validation error"))).toBe(false);
  });

  it("returns false for unknown errors", () => {
    expect(isRetryableError(new Error("something happened"))).toBe(false);
  });
});

describe("requiresAuthentication", () => {
  it("returns true for auth errors", () => {
    expect(requiresAuthentication(new Error("unauthorized"))).toBe(true);
  });

  it("returns true for ApiRequestError 401", () => {
    expect(requiresAuthentication(new ApiRequestError("Unauthorized", 401))).toBe(true);
  });

  it("returns false for non-auth errors", () => {
    expect(requiresAuthentication(new Error("server error"))).toBe(false);
  });

  it("returns false for ApiRequestError 500", () => {
    expect(requiresAuthentication(new ApiRequestError("Server error", 500))).toBe(false);
  });
});

describe("getErrorMessage", () => {
  it("returns error message from Error instance", () => {
    expect(getErrorMessage(new Error("test error"))).toBe("test error");
  });

  it("returns message from ApiRequestError", () => {
    expect(getErrorMessage(new ApiRequestError("Not found", 404))).toBe("Not found");
  });

  it("returns fallback for non-Error values", () => {
    expect(getErrorMessage(42, "custom fallback")).toBe("custom fallback");
  });
});
