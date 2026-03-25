import { describe, it, expect } from "vitest";
import { createHouseholdSchema } from "../household.schema";

describe("createHouseholdSchema", () => {
  it("accepts valid data", () => {
    const result = createHouseholdSchema.safeParse({
      name: "Main Home",
      timezone: "America/Chicago",
      units: "imperial",
    });
    expect(result.success).toBe(true);
  });

  it("accepts metric units", () => {
    const result = createHouseholdSchema.safeParse({
      name: "Home",
      timezone: "UTC",
      units: "metric",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty name", () => {
    const result = createHouseholdSchema.safeParse({
      name: "",
      timezone: "UTC",
      units: "imperial",
    });
    expect(result.success).toBe(false);
  });

  it("rejects name exceeding max length", () => {
    const result = createHouseholdSchema.safeParse({
      name: "a".repeat(101),
      timezone: "UTC",
      units: "imperial",
    });
    expect(result.success).toBe(false);
  });

  it("rejects empty timezone", () => {
    const result = createHouseholdSchema.safeParse({
      name: "Home",
      timezone: "",
      units: "imperial",
    });
    expect(result.success).toBe(false);
  });

  it("rejects invalid units", () => {
    const result = createHouseholdSchema.safeParse({
      name: "Home",
      timezone: "UTC",
      units: "kelvin",
    });
    expect(result.success).toBe(false);
  });

  it("rejects missing fields", () => {
    const result = createHouseholdSchema.safeParse({});
    expect(result.success).toBe(false);
  });
});
