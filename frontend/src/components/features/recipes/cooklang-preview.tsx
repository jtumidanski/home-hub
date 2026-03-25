import { Loader2 } from "lucide-react";
import { RecipeIngredients } from "./recipe-ingredients";
import { RecipeSteps } from "./recipe-steps";
import type { Ingredient, Step, ParseError } from "@/types/models/recipe";

interface CooklangPreviewProps {
  ingredients: Ingredient[];
  steps: Step[];
  errors: ParseError[];
  isLoading: boolean;
}

export function CooklangPreview({ ingredients, steps, errors, isLoading }: CooklangPreviewProps) {
  const isEmpty = ingredients.length === 0 && steps.length === 0 && errors.length === 0;

  return (
    <div className="space-y-4">
      {isLoading && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" /> Parsing...
        </div>
      )}

      {isEmpty && !isLoading && (
        <p className="text-sm text-muted-foreground">
          Start typing Cooklang to see a live preview...
        </p>
      )}

      {errors.length > 0 && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3 space-y-1">
          {errors.map((err, i) => (
            <p key={i} className="text-xs text-destructive">
              Line {err.line}, col {err.column}: {err.message}
            </p>
          ))}
        </div>
      )}

      {ingredients.length > 0 && (
        <div>
          <h3 className="text-base font-semibold mb-2">Ingredients</h3>
          <RecipeIngredients ingredients={ingredients} />
        </div>
      )}

      {steps.length > 0 && (
        <div>
          <h3 className="text-base font-semibold mb-2">Steps</h3>
          <RecipeSteps steps={steps} />
        </div>
      )}
    </div>
  );
}
