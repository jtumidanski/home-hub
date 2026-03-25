import type { Ingredient } from "@/types/models/recipe";

interface RecipeIngredientsProps {
  ingredients: Ingredient[];
}

export function RecipeIngredients({ ingredients }: RecipeIngredientsProps) {
  if (ingredients.length === 0) {
    return <p className="text-sm text-muted-foreground">No ingredients found.</p>;
  }

  return (
    <ul className="space-y-1.5">
      {ingredients.map((ing, i) => (
        <li key={i} className="flex items-baseline gap-2 text-sm">
          <span className="font-medium text-primary min-w-fit">
            {ing.quantity && `${ing.quantity}`}
            {ing.unit && ` ${ing.unit}`}
          </span>
          <span>{ing.name}</span>
        </li>
      ))}
    </ul>
  );
}
