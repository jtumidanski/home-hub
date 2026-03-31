import { AlertTriangle } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { usePlanIngredients } from "@/lib/hooks/api/use-meals";

interface IngredientPreviewProps {
  planId: string | null;
}

export function IngredientPreview({ planId }: IngredientPreviewProps) {
  const { data, isLoading } = usePlanIngredients(planId);
  const ingredients = data?.data ?? [];

  if (!planId) {
    return (
      <p className="text-sm text-muted-foreground">Save the plan to see ingredients</p>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-5 w-full" />
        ))}
      </div>
    );
  }

  if (ingredients.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No ingredients yet</p>
    );
  }

  const unresolvedCount = ingredients.filter((i) => !i.attributes.resolved).length;

  return (
    <div className="space-y-2">
      {unresolvedCount > 0 && (
        <div className="flex items-center gap-1 text-xs text-yellow-600">
          <AlertTriangle className="h-3 w-3" />
          {unresolvedCount} ingredient{unresolvedCount > 1 ? "s" : ""} could not be consolidated
        </div>
      )}
      <ul className="space-y-1 text-sm">
        {ingredients.map((ing) => {
          const extras = ing.attributes.extra_quantities ?? [];
          const name = ing.attributes.display_name ?? ing.attributes.name;
          return (
            <li key={ing.id} className={`${!ing.attributes.resolved ? "text-muted-foreground italic" : ""}`}>
              {formatQuantity(ing.attributes.quantity)} {ing.attributes.unit}{" "}
              {extras.length > 0 && (
                <>
                  {"+ "}
                  {extras.map((eq, i) => (
                    <span key={i}>
                      {formatQuantity(eq.quantity)} {eq.unit}
                      {i < extras.length - 1 ? " + " : " "}
                    </span>
                  ))}
                </>
              )}
              {name}
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function formatQuantity(v: number): string {
  if (v === Math.trunc(v)) return String(Math.trunc(v));
  return v.toFixed(1);
}
