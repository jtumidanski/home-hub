import { useCallback, useEffect, useRef, useState } from "react";
import { useRecipeImageSrc } from "@/lib/hooks/use-recipe-image-src";
import type { RecipeImage } from "@/types/models/recipe";

export type ImageDraft =
  | { kind: "unchanged" }
  | { kind: "staged"; file: File; previewUrl: string }
  | { kind: "removed" };

/**
 * Holds the not-yet-saved image change for the recipe form and owns the
 * object-URL lifecycle of the staged preview.
 */
export function useRecipeImageDraft(existing: RecipeImage | null) {
  const [draft, setDraft] = useState<ImageDraft>({ kind: "unchanged" });
  const urlRef = useRef<string | null>(null);

  const revoke = useCallback(() => {
    if (urlRef.current) {
      URL.revokeObjectURL(urlRef.current);
      urlRef.current = null;
    }
  }, []);

  useEffect(() => revoke, [revoke]);

  const stage = useCallback(
    (file: File) => {
      revoke();
      const previewUrl = URL.createObjectURL(file);
      urlRef.current = previewUrl;
      setDraft({ kind: "staged", file, previewUrl });
    },
    [revoke],
  );

  const remove = useCallback(() => {
    revoke();
    setDraft(existing ? { kind: "removed" } : { kind: "unchanged" });
  }, [revoke, existing]);

  const undo = useCallback(() => {
    revoke();
    setDraft({ kind: "unchanged" });
  }, [revoke]);

  // The staged preview is already a local object URL; only the existing server image needs resolving.
  const { src: existingSrc } = useRecipeImageSrc(draft.kind === "unchanged" ? (existing?.fullUrl ?? null) : null);
  const previewSrc = draft.kind === "staged" ? draft.previewUrl : draft.kind === "removed" ? null : existingSrc;

  return { draft, previewSrc, stage, remove, undo };
}
