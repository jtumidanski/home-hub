import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, Pencil, Trash2, Clock, Users, ExternalLink } from "lucide-react";
import { toast } from "sonner";
import { useRecipe, useDeleteRecipe } from "@/lib/hooks/api/use-recipes";
import { RecipeIngredients } from "@/components/features/recipes/recipe-ingredients";
import { RecipeSteps } from "@/components/features/recipes/recipe-steps";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { createErrorFromUnknown } from "@/lib/api/errors";

export function RecipeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useRecipe(id!);
  const deleteRecipe = useDeleteRecipe();

  const recipe = data?.data;

  const handleDelete = async () => {
    if (!id) return;
    try {
      await deleteRecipe.mutateAsync(id);
      toast.success("Recipe deleted");
      navigate("/app/recipes");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete recipe").message);
    }
  };

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-96" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!recipe) {
    return (
      <div className="p-4 md:p-6">
        <p className="text-muted-foreground">Recipe not found.</p>
        <Button variant="ghost" className="mt-2" onClick={() => navigate("/app/recipes")}>
          <ArrowLeft className="mr-1 h-4 w-4" /> Back to recipes
        </Button>
      </div>
    );
  }

  const attrs = recipe.attributes;
  const totalTime = (attrs.prepTimeMinutes ?? 0) + (attrs.cookTimeMinutes ?? 0);

  return (
    <div className="p-4 md:p-6 space-y-6 max-w-3xl">
      {/* Header */}
      <div className="space-y-3">
        <Button variant="ghost" size="sm" onClick={() => navigate("/app/recipes")}>
          <ArrowLeft className="mr-1 h-4 w-4" /> Recipes
        </Button>

        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{attrs.title}</h1>
            {attrs.description && (
              <p className="mt-1 text-muted-foreground">{attrs.description}</p>
            )}
          </div>
          <div className="flex gap-2 shrink-0">
            <Button variant="outline" size="sm" onClick={() => navigate(`/app/recipes/${id}/edit`)}>
              <Pencil className="mr-1 h-4 w-4" /> Edit
            </Button>
            <Button variant="outline" size="sm" className="text-destructive" onClick={handleDelete}>
              <Trash2 className="mr-1 h-4 w-4" /> Delete
            </Button>
          </div>
        </div>

        {/* Metadata */}
        <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
          {attrs.servings && (
            <span className="inline-flex items-center gap-1">
              <Users className="h-4 w-4" /> {attrs.servings} servings
            </span>
          )}
          {totalTime > 0 && (
            <span className="inline-flex items-center gap-1">
              <Clock className="h-4 w-4" />
              {attrs.prepTimeMinutes ? `${attrs.prepTimeMinutes}m prep` : ""}
              {attrs.prepTimeMinutes && attrs.cookTimeMinutes ? " + " : ""}
              {attrs.cookTimeMinutes ? `${attrs.cookTimeMinutes}m cook` : ""}
            </span>
          )}
          {attrs.sourceUrl && (
            <a
              href={attrs.sourceUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-primary hover:underline"
            >
              <ExternalLink className="h-3 w-3" /> Source
            </a>
          )}
        </div>

        {attrs.tags.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {attrs.tags.map((tag) => (
              <Badge key={tag} variant="secondary">{tag}</Badge>
            ))}
          </div>
        )}
      </div>

      {/* Ingredients */}
      <div>
        <h2 className="text-lg font-semibold mb-3">Ingredients</h2>
        <RecipeIngredients ingredients={attrs.ingredients} />
      </div>

      {/* Steps */}
      <div>
        <h2 className="text-lg font-semibold mb-3">Instructions</h2>
        <RecipeSteps steps={attrs.steps} />
      </div>
    </div>
  );
}
