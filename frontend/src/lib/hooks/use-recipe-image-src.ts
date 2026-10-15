import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { recipeService } from "@/services/api/recipe";
import { useTenant } from "@/context/tenant-context";

/**
 * Resolves a server image URL (`/api/v1/recipes/{id}/image?size=..&v=..`) to a
 * local object URL. A bare <img src> sends no X-Tenant-ID / X-Household-ID
 * headers, so the image is fetched through the api client instead.
 *
 * The URL carries `?v=`, so it is immutable and a replaced image gets a new
 * query key. The key is deliberately outside `recipeKeys` so recipe
 * invalidation does not refetch (and flicker) immutable image bytes. The object URL is revoked on change and on unmount.
 */
export function useRecipeImageSrc(url: string | null, options: { enabled?: boolean } = {}) {
  const { enabled = true } = options;
  const { tenant, household } = useTenant();
  const query = useQuery({
    queryKey: ["recipe-images", tenant?.id, household?.id, url] as const,
    queryFn: () => recipeService.downloadImage(url!),
    enabled: enabled && !!url && !!tenant?.id && !!household?.id,
    staleTime: Infinity,
    gcTime: 30 * 60 * 1000,
    retry: false,
  });

  const blob = query.data;
  // Paired with its blob so a stale (already revoked) URL is never used for a new blob.
  const [resolved, setResolved] = useState<{ blob: Blob; objectUrl: string } | null>(null);
  useEffect(() => {
    if (!blob) return;
    const objectUrl = URL.createObjectURL(blob);
    // eslint-disable-next-line react-hooks/set-state-in-effect -- object URL is an external resource tied to this effect's cleanup
    setResolved({ blob, objectUrl });
    return () => URL.revokeObjectURL(objectUrl);
  }, [blob]);

  const src = url && blob && resolved?.blob === blob ? resolved.objectUrl : null;
  return { src, isError: query.isError };
}
