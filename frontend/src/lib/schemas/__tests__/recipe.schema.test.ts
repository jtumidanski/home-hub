import { describe, it, expect } from "vitest";
import { recipeFormSchema } from "../recipe.schema";

describe("recipeFormSchema", () => {
  it("accepts valid data with required fields only", () => {
    const result = recipeFormSchema.safeParse({
      title: "Pasta Carbonara",
      source: "Add @eggs{3}.",
    });
    expect(result.success).toBe(true);
  });

  it("accepts valid data with all fields", () => {
    const result = recipeFormSchema.safeParse({
      title: "Pasta Carbonara",
      description: "Classic Roman dish",
      source: "Add @eggs{3}.\n\nCook @spaghetti{400%g}.",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty title", () => {
    const result = recipeFormSchema.safeParse({
      title: "",
      source: "Add @eggs{3}.",
    });
    expect(result.success).toBe(false);
  });

  it("rejects title exceeding max length", () => {
    const result = recipeFormSchema.safeParse({
      title: "a".repeat(256),
      source: "Add @eggs{3}.",
    });
    expect(result.success).toBe(false);
  });

  it("rejects empty source", () => {
    const result = recipeFormSchema.safeParse({
      title: "Recipe",
      source: "",
    });
    expect(result.success).toBe(false);
  });

  it("rejects missing source", () => {
    const result = recipeFormSchema.safeParse({
      title: "Recipe",
    });
    expect(result.success).toBe(false);
  });

  it("rejects description exceeding max length", () => {
    const result = recipeFormSchema.safeParse({
      title: "Recipe",
      source: "Add @salt.",
      description: "a".repeat(2001),
    });
    expect(result.success).toBe(false);
  });

  it("allows optional description", () => {
    const result = recipeFormSchema.safeParse({
      title: "Recipe",
      source: "Add @salt.",
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.description).toBeUndefined();
    }
  });
});
