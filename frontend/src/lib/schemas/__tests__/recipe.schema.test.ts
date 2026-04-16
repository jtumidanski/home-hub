import { describe, it, expect } from "vitest";
import { recipeFormSchema } from "../recipe.schema";

describe("recipeFormSchema", () => {
  it("accepts valid source", () => {
    const result = recipeFormSchema.safeParse({
      source: "Add @eggs{3}.",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty source", () => {
    const result = recipeFormSchema.safeParse({ source: "" });
    expect(result.success).toBe(false);
  });

  it("rejects missing source", () => {
    const result = recipeFormSchema.safeParse({});
    expect(result.success).toBe(false);
  });
});
