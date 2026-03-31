import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, X } from "lucide-react";
import { toast } from "sonner";
import { useRecipes, useRecipeTags, useDeleteRecipe } from "@/lib/hooks/api/use-recipes";
import { useMobile } from "@/lib/hooks/use-mobile";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { RecipeCard } from "@/components/features/recipes/recipe-card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { CLASSIFICATIONS } from "@/lib/constants/recipe";
import { toTitleCase } from "@/lib/utils";
import type { RecipeListItem } from "@/types/models/recipe";

export function RecipesPage() {
  const navigate = useNavigate();
  const isMobile = useMobile();
  const [search, setSearch] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [plannerFilter, setPlannerFilter] = useState<"" | "ready" | "not-ready">("");

  // Extract classification from selected tags
  const selectedClassification = selectedTags.find((t) => (CLASSIFICATIONS as readonly string[]).includes(t));
  const selectedNonClassTags = selectedTags.filter((t) => !(CLASSIFICATIONS as readonly string[]).includes(t));

  const { data, isLoading, refetch } = useRecipes({
    search: search || undefined,
    tags: selectedNonClassTags.length > 0 ? selectedNonClassTags : undefined,
    classification: selectedClassification,
    plannerReady: plannerFilter === "ready" ? true : plannerFilter === "not-ready" ? false : undefined,
  });
  const { data: tagsData } = useRecipeTags();
  const deleteRecipe = useDeleteRecipe();

  const recipes: RecipeListItem[] = data?.data ?? [];
  const availableTags = (tagsData?.data ?? []) as Array<{ attributes: { tag: string; count: number } }>;

  const handleRefresh = useCallback(async () => {
    await refetch();
  }, [refetch]);

  const handleDelete = useCallback(
    async (id: string) => {
      try {
        await deleteRecipe.mutateAsync(id);
        toast.success("Recipe deleted");
      } catch (error) {
        toast.error(createErrorFromUnknown(error, "Failed to delete recipe").message);
      }
    },
    [deleteRecipe],
  );

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag],
    );
  };

  const hasActiveFilters = selectedTags.length > 0 || plannerFilter !== "";

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Recipes</h1>
          <Button size="sm" onClick={() => navigate("/app/recipes/new")}>
            <Plus className="mr-1 h-4 w-4" />
            New Recipe
          </Button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search recipes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
          {search && (
            <button
              type="button"
              onClick={() => setSearch("")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        {/* Unified filter row */}
        <div className="flex flex-wrap gap-1.5">
            {/* Planner status filters */}
            <Badge
              variant={plannerFilter === "ready" ? "default" : "outline"}
              className="cursor-pointer text-xs"
              onClick={() => setPlannerFilter(plannerFilter === "ready" ? "" : "ready")}
            >
              Planner Ready
            </Badge>
            <Badge
              variant={plannerFilter === "not-ready" ? "default" : "outline"}
              className="cursor-pointer text-xs"
              onClick={() => setPlannerFilter(plannerFilter === "not-ready" ? "" : "not-ready")}
            >
              Not Ready
            </Badge>

            {/* Separator */}
            {availableTags.length > 0 && (
              <span className="border-l border-border mx-1" />
            )}

            {/* Tag filters (includes classification tags) */}
            {availableTags.map((t) => (
              <Badge
                key={t.attributes.tag}
                variant={selectedTags.includes(t.attributes.tag) ? "default" : "outline"}
                className="cursor-pointer text-xs"
                onClick={() => toggleTag(t.attributes.tag)}
              >
                {toTitleCase(t.attributes.tag)} ({t.attributes.count})
              </Badge>
            ))}

            {hasActiveFilters && (
              <button
                type="button"
                className="inline-flex items-center rounded-md px-2 py-0.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
                onClick={() => { setSelectedTags([]); setPlannerFilter(""); }}
              >
                Clear
              </button>
            )}
          </div>

        {/* Loading */}
        {isLoading && (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-24 w-full" />
            ))}
          </div>
        )}

        {/* Empty state */}
        {!isLoading && recipes.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground mb-4">No recipes yet.</p>
            <Button onClick={() => navigate("/app/recipes/new")}>
              <Plus className="mr-1 h-4 w-4" />
              Create your first recipe
            </Button>
          </div>
        )}

        {/* Recipe list */}
        {!isLoading && recipes.length > 0 && (
          <div className={isMobile ? "space-y-3" : "grid grid-cols-2 lg:grid-cols-3 gap-4"}>
            {recipes.map((recipe) => (
              <RecipeCard key={recipe.id} recipe={recipe} onDelete={handleDelete} />
            ))}
          </div>
        )}
      </div>
    </PullToRefresh>
  );
}
