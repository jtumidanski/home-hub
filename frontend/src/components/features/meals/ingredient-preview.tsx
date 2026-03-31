import { useMemo } from "react";
import { AlertTriangle } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { usePlanIngredients } from "@/lib/hooks/api/use-meals";
import type { PlanIngredient } from "@/types/models/meal-plan";

interface IngredientPreviewProps {
  planId: string | null;
}

interface CategoryGroup {
  name: string;
  sortOrder: number;
  ingredients: PlanIngredient[];
}

export function IngredientPreview({ planId }: IngredientPreviewProps) {
  const { data, isLoading } = usePlanIngredients(planId);
  const ingredients = data?.data ?? [];

  const groups = useMemo(() => {
    if (ingredients.length === 0) return [];

    const groupMap = new Map<string, CategoryGroup>();
    const uncategorized: PlanIngredient[] = [];

    for (const ing of ingredients) {
      const catName = ing.attributes.category_name;
      if (catName) {
        let group = groupMap.get(catName);
        if (!group) {
          group = {
            name: catName,
            sortOrder: ing.attributes.category_sort_order ?? 0,
            ingredients: [],
          };
          groupMap.set(catName, group);
        }
        group.ingredients.push(ing);
      } else {
        uncategorized.push(ing);
      }
    }

    // Sort groups by sort order
    const sorted = Array.from(groupMap.values()).sort((a, b) => a.sortOrder - b.sortOrder);

    // Sort ingredients alphabetically within each group
    for (const group of sorted) {
      group.ingredients.sort((a, b) => {
        const na = a.attributes.display_name ?? a.attributes.name;
        const nb = b.attributes.display_name ?? b.attributes.name;
        return na.localeCompare(nb);
      });
    }

    // Sort uncategorized alphabetically too
    uncategorized.sort((a, b) => {
      const na = a.attributes.display_name ?? a.attributes.name;
      const nb = b.attributes.display_name ?? b.attributes.name;
      return na.localeCompare(nb);
    });

    if (uncategorized.length > 0) {
      sorted.push({ name: "Uncategorized", sortOrder: Number.MAX_SAFE_INTEGER, ingredients: uncategorized });
    }

    return sorted;
  }, [ingredients]);

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
    <div className="space-y-3">
      {unresolvedCount > 0 && (
        <div className="flex items-center gap-1 text-xs text-yellow-600">
          <AlertTriangle className="h-3 w-3" />
          {unresolvedCount} ingredient{unresolvedCount > 1 ? "s" : ""} could not be consolidated
        </div>
      )}
      {groups.map((group) => (
        <div key={group.name}>
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide border-t pt-2 mb-1">
            {group.name}
          </h4>
          <ul className="space-y-0.5 text-sm">
            {group.ingredients.map((ing) => {
              const extras = ing.attributes.extra_quantities ?? [];
              const name = ing.attributes.display_name ?? ing.attributes.name;
              return (
                <li key={ing.id} className={cn(!ing.attributes.resolved && "text-muted-foreground italic")}>
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
      ))}
    </div>
  );
}

function formatQuantity(v: number): string {
  if (v === Math.trunc(v)) return String(Math.trunc(v));
  return v.toFixed(1);
}
