import { describe, it, expect } from "vitest";
import { packageFormSchema, packageEditSchema } from "../package.schema";

describe("packageFormSchema", () => {
  it("accepts valid data with required fields only", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "1Z999AA10123456784",
      carrier: "ups",
    });
    expect(result.success).toBe(true);
  });

  it("accepts valid data with all fields", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "1Z999AA10123456784",
      carrier: "fedex",
      label: "My Package",
      notes: "Leave at door",
      private: true,
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty tracking number", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "",
      carrier: "usps",
    });
    expect(result.success).toBe(false);
  });

  it("rejects tracking number exceeding max length", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "a".repeat(65),
      carrier: "ups",
    });
    expect(result.success).toBe(false);
  });

  it("rejects invalid carrier", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "ABC123",
      carrier: "dhl",
    });
    expect(result.success).toBe(false);
  });

  it("accepts all valid carriers", () => {
    for (const carrier of ["usps", "ups", "fedex"]) {
      const result = packageFormSchema.safeParse({
        trackingNumber: "ABC123",
        carrier,
      });
      expect(result.success).toBe(true);
    }
  });

  it("rejects label exceeding max length", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "ABC123",
      carrier: "ups",
      label: "a".repeat(256),
    });
    expect(result.success).toBe(false);
  });

  it("allows optional fields to be omitted", () => {
    const result = packageFormSchema.safeParse({
      trackingNumber: "ABC123",
      carrier: "usps",
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.label).toBeUndefined();
      expect(result.data.notes).toBeUndefined();
      expect(result.data.private).toBeUndefined();
    }
  });
});

describe("packageEditSchema", () => {
  it("accepts partial updates", () => {
    const result = packageEditSchema.safeParse({ label: "Updated" });
    expect(result.success).toBe(true);
  });

  it("accepts empty object (no changes)", () => {
    const result = packageEditSchema.safeParse({});
    expect(result.success).toBe(true);
  });

  it("rejects invalid carrier", () => {
    const result = packageEditSchema.safeParse({ carrier: "dhl" });
    expect(result.success).toBe(false);
  });

  it("rejects label exceeding max length", () => {
    const result = packageEditSchema.safeParse({ label: "a".repeat(256) });
    expect(result.success).toBe(false);
  });
});
