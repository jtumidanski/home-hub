import { describe, it, expect } from "vitest";
import { validateRecipeImageFile, RECIPE_IMAGE_ERROR, RECIPE_IMAGE_MAX_BYTES } from "../recipe-image";

function fileOf(type: string, size: number): File {
  const f = new File(["x"], "photo", { type });
  Object.defineProperty(f, "size", { value: size });
  return f;
}

describe("validateRecipeImageFile", () => {
  it.each(["image/jpeg", "image/png", "image/webp"])("accepts %s", (type) => {
    expect(validateRecipeImageFile(fileOf(type, 1000))).toBeNull();
  });

  it("accepts exactly 10 MiB", () => {
    expect(validateRecipeImageFile(fileOf("image/jpeg", RECIPE_IMAGE_MAX_BYTES))).toBeNull();
  });

  it("rejects one byte over 10 MiB", () => {
    expect(validateRecipeImageFile(fileOf("image/jpeg", RECIPE_IMAGE_MAX_BYTES + 1))).toBe(RECIPE_IMAGE_ERROR);
  });

  it.each(["image/gif", "image/heic", "application/pdf"])("rejects %s", (type) => {
    expect(validateRecipeImageFile(fileOf(type, 1000))).toBe(RECIPE_IMAGE_ERROR);
  });

  it("rejects empty mime type", () => {
    expect(validateRecipeImageFile(fileOf("", 1000))).toBe(RECIPE_IMAGE_ERROR);
  });

  it("uses the PRD message", () => {
    expect(RECIPE_IMAGE_ERROR).toBe("Use a JPEG, PNG or WebP image under 10 MB.");
  });
});
