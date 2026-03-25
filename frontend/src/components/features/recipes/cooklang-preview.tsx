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
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">Preview</h3>
        {isLoading && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
      </div>

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
          <h4 className="text-xs font-semibold uppercase text-muted-foreground mb-2">Ingredients</h4>
          <RecipeIngredients ingredients={ingredients} />
        </div>
      )}

      {steps.length > 0 && (
        <div>
          <h4 className="text-xs font-semibold uppercase text-muted-foreground mb-2">Steps</h4>
          <RecipeSteps steps={steps} />
        </div>
      )}
    </div>
  );
}
