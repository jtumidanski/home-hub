import { Check, AlertTriangle, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { IngredientResolver } from "./ingredient-resolver";
import type { RecipeIngredient } from "@/types/models/recipe";

interface IngredientNormalizationPanelProps {
  ingredients: RecipeIngredient[];
  recipeId: string;
  onRenormalize?: () => void;
  isRenormalizing?: boolean;
  readOnly?: boolean;
}

export function IngredientNormalizationPanel({
  ingredients,
  recipeId,
  onRenormalize,
  isRenormalizing,
  readOnly,
}: IngredientNormalizationPanelProps) {
  if (ingredients.length === 0) return null;

  const resolved = ingredients.filter((i) => i.normalizationStatus !== "unresolved").length;
  const total = ingredients.length;
  const hasUnresolved = resolved < total;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium">Ingredient Normalization</h3>
          <Badge variant={hasUnresolved ? "outline" : "secondary"} className="text-xs">
            {resolved}/{total} resolved
          </Badge>
        </div>
        {onRenormalize && hasUnresolved && (
          <Button
            variant="outline"
            size="sm"
            className="text-xs h-7"
            onClick={onRenormalize}
            disabled={isRenormalizing}
          >
            <RefreshCw className={cn("mr-1 h-3 w-3", isRenormalizing && "animate-spin")} />
            Re-normalize
          </Button>
        )}
      </div>

      <ul className="space-y-1.5">
        {ingredients.map((ing) => (
          <li key={ing.id} className="flex items-center gap-2 text-sm">
            <StatusIcon status={ing.normalizationStatus} />
            <span className="flex-1">
              {ing.rawQuantity && (
                <span className="font-medium text-primary">
                  {ing.rawQuantity}{ing.rawUnit ? ` ${ing.rawUnit}` : ""}
                </span>
              )}{" "}
              <span>{ing.rawName}</span>
              {ing.canonicalName && ing.normalizationStatus !== "unresolved" && (
                <span className="text-muted-foreground text-xs ml-1">
                  → {ing.canonicalName}
                </span>
              )}
            </span>
            {ing.normalizationStatus === "unresolved" && !readOnly && (
              <IngredientResolver recipeId={recipeId} ingredientId={ing.id} rawName={ing.rawName} />
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case "matched":
    case "alias_matched":
    case "manually_confirmed":
      return <Check className="h-4 w-4 text-green-500 shrink-0" />;
    case "unresolved":
      return <AlertTriangle className="h-4 w-4 text-yellow-500 shrink-0" />;
    default:
      return null;
  }
}
