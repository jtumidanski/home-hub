import { useState, type RefCallback } from "react";
import { ChefHat } from "lucide-react";
import { cn } from "@/lib/utils";
import { useInView } from "@/lib/hooks/use-in-view";
import { useRecipeImageSrc } from "@/lib/hooks/use-recipe-image-src";
import type { RecipeImage as RecipeImageData } from "@/types/models/recipe";

interface RecipeImageProps {
  image: RecipeImageData | null;
  variant: "hero" | "card";
  alt: string;
  className?: string;
}

function RecipeImagePlaceholder({
  className,
  innerRef,
}: {
  className?: string | undefined;
  innerRef?: RefCallback<Element>;
}) {
  return (
    <div
      ref={innerRef}
      data-testid="recipe-image-placeholder"
      aria-hidden="true"
      className={cn("flex items-center justify-center bg-muted", className)}
    >
      <ChefHat className="h-6 w-6 text-muted-foreground/50" />
    </div>
  );
}

export function RecipeImage({ image, variant, alt, className }: RecipeImageProps) {
  // Keyed by src so a replaced image (new ?v=) gets a fresh attempt.
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const isHero = variant === "hero";
  const url = image ? (isHero ? image.fullUrl : image.thumbUrl) : null;
  // Cards fetch only once near the viewport (replaces native lazy loading, which a blob URL can't defer).
  const { ref, inView } = useInView("200px");
  const { src, isError } = useRecipeImageSrc(url, { enabled: isHero || inView });

  // While loading (no src yet) or on error: hero renders nothing, card keeps its footprint.
  if (!src || isError || failedSrc === src) {
    // The detail page has no placeholder (FR-12); cards keep their footprint (FR-16/17).
    return isHero ? null : <RecipeImagePlaceholder className={className} innerRef={ref} />;
  }

  return (
    <img
      src={src}
      alt={alt}
      loading={isHero ? "eager" : "lazy"}
      className={cn("object-cover", className)}
      onError={() => setFailedSrc(src)}
    />
  );
}
