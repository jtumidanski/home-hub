import type { Ingredient, RecipeIngredient } from "@/types/models/recipe";

interface RecipeIngredientsProps {
  ingredients: Ingredient[] | RecipeIngredient[];
}

function isRecipeIngredient(ing: Ingredient | RecipeIngredient): ing is RecipeIngredient {
  return "rawName" in ing;
}

export function RecipeIngredients({ ingredients }: RecipeIngredientsProps) {
  if (ingredients.length === 0) {
    return <p className="text-sm text-muted-foreground">No ingredients found.</p>;
  }

  return (
    <ul className="space-y-1.5 list-disc list-inside">
      {ingredients.map((ing, i) => {
        const name = isRecipeIngredient(ing) ? ing.rawName : ing.name;
        const quantity = isRecipeIngredient(ing) ? ing.rawQuantity : ing.quantity;
        const unit = isRecipeIngredient(ing) ? ing.rawUnit : ing.unit;

        return (
          <li key={i} className="text-sm">
            {quantity && (
              <span className="font-medium text-primary">{quantity}{unit ? ` ${unit}` : ""}</span>
            )}
            {quantity ? " " : ""}
            <span>{name}</span>
          </li>
        );
      })}
    </ul>
  );
}
