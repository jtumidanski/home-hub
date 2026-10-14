export const RECIPE_IMAGE_MAX_BYTES = 10 * 1024 * 1024;
export const RECIPE_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp"] as const;
export const RECIPE_IMAGE_ACCEPT = RECIPE_IMAGE_TYPES.join(",");
export const RECIPE_IMAGE_ERROR = "Use a JPEG, PNG or WebP image under 10 MB.";

/** Returns an error message for an unacceptable file, or null when valid. */
export function validateRecipeImageFile(file: File): string | null {
  const typeOk = (RECIPE_IMAGE_TYPES as readonly string[]).includes(file.type);
  if (!typeOk || file.size > RECIPE_IMAGE_MAX_BYTES) {
    return RECIPE_IMAGE_ERROR;
  }
  return null;
}
