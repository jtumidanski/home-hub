import { useState, useEffect, useRef } from "react";
import { recipeService } from "@/services/api/recipe";
import { useTenant } from "@/context/tenant-context";
import type { Ingredient, Step, ParseError, RecipeMetadata } from "@/types/models/recipe";

interface CooklangPreview {
  ingredients: Ingredient[];
  steps: Step[];
  errors: ParseError[];
  metadata: RecipeMetadata | null;
  isLoading: boolean;
}

export function useCooklangPreview(source: string, debounceMs = 300): CooklangPreview {
  const { tenant } = useTenant();
  const [ingredients, setIngredients] = useState<Ingredient[]>([]);
  const [steps, setSteps] = useState<Step[]>([]);
  const [errors, setErrors] = useState<ParseError[]>([]);
  const [metadata, setMetadata] = useState<RecipeMetadata | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const cancelledRef = useRef(false);
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    cancelledRef.current = true;

    if (!source.trim() || !tenant) {
      setIngredients([]);
      setSteps([]);
      setErrors([]);
      setMetadata(null);
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    cancelledRef.current = false;

    timeoutRef.current = setTimeout(async () => {
      try {
        const result = await recipeService.parseSource(tenant, source);
        if (cancelledRef.current) return;
        const attrs = result.data.attributes;
        setIngredients(attrs.ingredients ?? []);
        setSteps(attrs.steps ?? []);
        setErrors(attrs.errors ?? []);
        setMetadata(attrs.metadata ?? null);
      } catch {
        // Silently ignore errors from cancelled/failed requests
      } finally {
        if (!cancelledRef.current) {
          setIsLoading(false);
        }
      }
    }, debounceMs);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      cancelledRef.current = true;
    };
  }, [source, tenant, debounceMs]);

  return { ingredients, steps, errors, metadata, isLoading };
}
