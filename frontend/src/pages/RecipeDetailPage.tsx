import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, Pencil, Trash2, Clock, Users, ExternalLink, CheckCircle2, AlertCircle, Maximize2 } from "lucide-react";
import { toast } from "sonner";
import { useRecipe, useDeleteRecipe } from "@/lib/hooks/api/use-recipes";
import { useRecipeImageSrc } from "@/lib/hooks/use-recipe-image-src";
import { useRenormalize } from "@/lib/hooks/api/use-ingredient-normalization";
import { RecipeIngredients } from "@/components/features/recipes/recipe-ingredients";
import { RecipeSteps } from "@/components/features/recipes/recipe-steps";
import { CookMode } from "@/components/features/recipes/cook-mode";
import { RecipeImage } from "@/components/features/recipes/recipe-image";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { filterClassificationTags } from "@/lib/constants/recipe";
import { toTitleCase } from "@/lib/utils";

export function RecipeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useRecipe(id!);
  const deleteRecipe = useDeleteRecipe();
  const renormalize = useRenormalize();
  const [cookModeOpen, setCookModeOpen] = useState(false);
  const [lightboxOpen, setLightboxOpen] = useState(false);

  const recipe = data?.data;
  // Same query key as the hero, so this shares its cached bytes.
  const { src: lightboxSrc } = useRecipeImageSrc(recipe?.attributes.image?.fullUrl ?? null);

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
  const pc = attrs.plannerConfig;

  // Build tags: classification first, then non-classification tags
  const nonClassTags = filterClassificationTags(attrs.tags);
  const allTags = [
    ...(pc?.classification ? [pc.classification] : []),
    ...nonClassTags,
  ];

  // Planner tooltip
  const plannerTooltip = attrs.plannerReady
    ? "Planner Ready"
    : `Not Planner Ready${attrs.plannerIssues?.length ? ": " + attrs.plannerIssues.join(", ") : ""}`;

  return (
    <div className="p-4 md:p-6 space-y-6 max-w-5xl">
      {/* Header */}
      <div className="space-y-3">
        <Button variant="ghost" size="sm" onClick={() => navigate("/app/recipes")}>
          <ArrowLeft className="mr-1 h-4 w-4" /> Recipes
        </Button>

        {attrs.image && lightboxSrc && (
          <button
            type="button"
            aria-label="View photo full size"
            className="-mx-4 block w-[calc(100%+2rem)] md:mx-0 md:w-full"
            onClick={() => setLightboxOpen(true)}
          >
            <RecipeImage
              image={attrs.image}
              variant="hero"
              alt={attrs.title}
              className="aspect-[4/3] w-full md:aspect-video md:rounded-lg"
            />
          </button>
        )}

        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span title={plannerTooltip}>
                {attrs.plannerReady
                  ? <CheckCircle2 className="h-5 w-5 shrink-0 text-green-500" />
                  : <AlertCircle className="h-5 w-5 shrink-0 text-yellow-500" />
                }
              </span>
              <h1 className="text-xl sm:text-2xl font-bold break-words">{attrs.title}</h1>
            </div>
            {attrs.description && (
              <p className="mt-1 text-sm sm:text-base text-muted-foreground">{attrs.description}</p>
            )}
          </div>
          <div className="flex flex-wrap gap-2 sm:shrink-0">
            <Button variant="outline" size="sm" onClick={() => setCookModeOpen(true)}>
              <Maximize2 className="h-4 w-4 sm:mr-1" />
              <span className="hidden sm:inline">Cook Mode</span>
            </Button>
            <Button variant="outline" size="sm" onClick={() => navigate(`/app/recipes/${id}/edit`)}>
              <Pencil className="h-4 w-4 sm:mr-1" />
              <span className="hidden sm:inline">Edit</span>
            </Button>
            <Button variant="outline" size="sm" className="text-destructive" onClick={handleDelete}>
              <Trash2 className="h-4 w-4 sm:mr-1" />
              <span className="hidden sm:inline">Delete</span>
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
          {pc?.eatWithinDays && (
            <span>Eat within {pc.eatWithinDays}d</span>
          )}
          {pc?.minGapDays != null && pc.minGapDays > 0 && (
            <span>Gap {pc.minGapDays}d+</span>
          )}
          {pc?.maxConsecutiveDays && (
            <span>Max {pc.maxConsecutiveDays}d</span>
          )}
        </div>

        {allTags.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {allTags.map((tag) => (
              <Badge key={tag} variant="secondary">{toTitleCase(tag)}</Badge>
            ))}
          </div>
        )}
      </div>

      {/* Ingredients and Instructions — side-by-side on desktop */}
      <div className="grid grid-cols-1 md:grid-cols-[280px_1fr] gap-6 md:gap-8">
        <div>
          <h2 className="text-lg font-semibold mb-3">Ingredients</h2>
          <RecipeIngredients
            ingredients={attrs.ingredients}
            recipeId={id!}
            onRenormalize={() => renormalize.mutate(id!)}
            isRenormalizing={renormalize.isPending}
          />
        </div>

        <div>
          <h2 className="text-lg font-semibold mb-3">Instructions</h2>
          <RecipeSteps steps={attrs.steps} notes={attrs.notes} />
        </div>
      </div>

      {attrs.image && lightboxSrc && (
        <Dialog open={lightboxOpen} onOpenChange={setLightboxOpen}>
          <DialogContent className="max-w-[95vw] p-2 sm:max-w-[95vw]">
            <DialogTitle className="sr-only">{attrs.title}</DialogTitle>
            <img
              src={lightboxSrc}
              alt={attrs.title}
              className="max-h-[85vh] w-full object-contain"
            />
          </DialogContent>
        </Dialog>
      )}

      <CookMode
        steps={attrs.steps}
        title={attrs.title}
        open={cookModeOpen}
        onClose={() => setCookModeOpen(false)}
      />
    </div>
  );
}
