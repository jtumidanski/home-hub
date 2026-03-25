import { describe, it, expect } from "vitest";
import { createTenantSchema } from "../tenant.schema";

describe("createTenantSchema", () => {
  it("accepts valid name", () => {
    const result = createTenantSchema.safeParse({ name: "Smith Family" });
    expect(result.success).toBe(true);
  });

  it("rejects empty name", () => {
    const result = createTenantSchema.safeParse({ name: "" });
    expect(result.success).toBe(false);
  });

  it("rejects name exceeding max length", () => {
    const result = createTenantSchema.safeParse({ name: "a".repeat(101) });
    expect(result.success).toBe(false);
  });

  it("rejects missing name", () => {
    const result = createTenantSchema.safeParse({});
    expect(result.success).toBe(false);
  });
});
