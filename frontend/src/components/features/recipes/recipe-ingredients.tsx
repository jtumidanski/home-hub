import type { Ingredient } from "@/types/models/recipe";

interface RecipeIngredientsProps {
  ingredients: Ingredient[];
}

export function RecipeIngredients({ ingredients }: RecipeIngredientsProps) {
  if (ingredients.length === 0) {
    return <p className="text-sm text-muted-foreground">No ingredients found.</p>;
  }

  return (
    <ul className="space-y-1.5 list-disc list-inside">
      {ingredients.map((ing, i) => (
        <li key={i} className="text-sm">
          {ing.quantity && (
            <span className="font-medium text-primary">{ing.quantity}{ing.unit ? ` ${ing.unit}` : ""}</span>
          )}
          {ing.quantity ? " " : ""}
          <span>{ing.name}</span>
        </li>
      ))}
    </ul>
  );
}
