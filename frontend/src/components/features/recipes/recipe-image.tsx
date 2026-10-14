import { useState } from "react";
import { ChefHat } from "lucide-react";
import { cn } from "@/lib/utils";
import type { RecipeImage as RecipeImageData } from "@/types/models/recipe";

interface RecipeImageProps {
  image: RecipeImageData | null;
  variant: "hero" | "card";
  alt: string;
  className?: string;
}

function RecipeImagePlaceholder({ className }: { className?: string | undefined }) {
  return (
    <div
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
  const src = image ? (isHero ? image.fullUrl : image.thumbUrl) : null;

  if (!src || failedSrc === src) {
    // The detail page has no placeholder (FR-12); cards keep their footprint (FR-16/17).
    return isHero ? null : <RecipeImagePlaceholder className={className} />;
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
