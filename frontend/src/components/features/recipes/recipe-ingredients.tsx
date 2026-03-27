import { Check, AlertTriangle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { IngredientResolver } from "./ingredient-resolver";
import type { Ingredient, RecipeIngredient } from "@/types/models/recipe";

interface RecipeIngredientsProps {
  ingredients: Ingredient[] | RecipeIngredient[];
  recipeId?: string;
  onRenormalize?: () => void;
  isRenormalizing?: boolean;
  readOnly?: boolean;
}

function isRecipeIngredient(ing: Ingredient | RecipeIngredient): ing is RecipeIngredient {
  return "rawName" in ing;
}

function hasNormalizationData(ingredients: Ingredient[] | RecipeIngredient[]): ingredients is RecipeIngredient[] {
  const first = ingredients[0];
  return first !== undefined && isRecipeIngredient(first);
}

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case "matched":
    case "alias_matched":
    case "manually_confirmed":
      return <Check className="h-3.5 w-3.5 text-green-500 shrink-0" />;
    case "unresolved":
      return <AlertTriangle className="h-3.5 w-3.5 text-yellow-500 shrink-0" />;
    default:
      return null;
  }
}

export function RecipeIngredients({
  ingredients,
  recipeId,
  onRenormalize,
  isRenormalizing,
  readOnly,
}: RecipeIngredientsProps) {
  if (ingredients.length === 0) {
    return <p className="text-sm text-muted-foreground">No ingredients found.</p>;
  }

  const isNormalized = hasNormalizationData(ingredients);

  const hasUnresolved = isNormalized && ingredients.some((i) => i.normalizationStatus === "unresolved");

  return (
    <div className="space-y-2">
      {isNormalized && onRenormalize && hasUnresolved && (
        <div className="flex justify-end">
          <Button
            variant="outline"
            size="sm"
            className="text-xs h-6"
            onClick={onRenormalize}
            disabled={isRenormalizing}
          >
            <RefreshCw className={cn("mr-1 h-3 w-3", isRenormalizing && "animate-spin")} />
            Re-normalize
          </Button>
        </div>
      )}

      <ul className="space-y-1.5">
        {ingredients.map((ing, i) => {
          if (isRecipeIngredient(ing)) {
            return (
              <li key={ing.id} className="flex items-center gap-1.5 text-sm">
                <StatusIcon status={ing.normalizationStatus} />
                <span className="flex-1">
                  {ing.rawQuantity && (
                    <span className="font-medium text-primary">
                      {ing.rawQuantity}{ing.rawUnit ? ` ${ing.rawUnit}` : ""}
                    </span>
                  )}{" "}
                  {ing.rawName}
                  {ing.canonicalName && ing.normalizationStatus !== "unresolved" && (
                    <span className="text-muted-foreground text-xs ml-1">→ {ing.canonicalName}</span>
                  )}
                </span>
                {ing.normalizationStatus === "unresolved" && !readOnly && recipeId && (
                  <IngredientResolver recipeId={recipeId} ingredientId={ing.id} rawName={ing.rawName} />
                )}
              </li>
            );
          }

          // Plain Ingredient (cooklang preview)
          return (
            <li key={i} className="text-sm list-disc list-inside">
              {ing.quantity && (
                <span className="font-medium text-primary">{ing.quantity}{ing.unit ? ` ${ing.unit}` : ""}</span>
              )}
              {ing.quantity ? " " : ""}
              <span>{ing.name}</span>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
