import { describe, it, expect } from "vitest";
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

  it("classifies network errors", () => {
    const result = createErrorFromUnknown(new Error("network error"));
    expect(result.type).toBe("network");
  });

  it("classifies fetch errors as network", () => {
    const result = createErrorFromUnknown(new Error("Failed to fetch"));
    expect(result.type).toBe("network");
  });

  it("classifies unauthorized errors as auth", () => {
    const result = createErrorFromUnknown(new Error("unauthorized"));
    expect(result.type).toBe("auth");
  });

  it("classifies 401 as auth", () => {
    const result = createErrorFromUnknown(new Error("Request failed: 401"));
    expect(result.type).toBe("auth");
  });

  it("classifies validation errors", () => {
    const result = createErrorFromUnknown(new Error("validation failed"));
    expect(result.type).toBe("validation");
  });

  it("classifies server errors", () => {
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
});

describe("isRetryableError", () => {
  it("returns true for network errors", () => {
    expect(isRetryableError(new Error("network timeout"))).toBe(true);
  });

  it("returns true for server errors", () => {
    expect(isRetryableError(new Error("500 server error"))).toBe(true);
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

  it("returns false for non-auth errors", () => {
    expect(requiresAuthentication(new Error("server error"))).toBe(false);
  });
});

describe("getErrorMessage", () => {
  it("returns error message from Error instance", () => {
    expect(getErrorMessage(new Error("test error"))).toBe("test error");
  });

  it("returns fallback for non-Error values", () => {
    expect(getErrorMessage(42, "custom fallback")).toBe("custom fallback");
  });
});
