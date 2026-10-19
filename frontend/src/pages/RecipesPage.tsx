import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { BookOpen, Plus, Search, X } from "lucide-react";
import { toast } from "sonner";
import { useRecipes, useRecipeTags, useDeleteRecipe } from "@/lib/hooks/api/use-recipes";
import { useMobile } from "@/lib/hooks/use-mobile";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { RecipeCard } from "@/components/features/recipes/recipe-card";
import { RecipeSortSelect } from "@/components/features/recipes/recipe-sort-select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { CLASSIFICATIONS } from "@/lib/constants/recipe";
import { toTitleCase } from "@/lib/utils";
import type { RecipeListItem, RecipeSort } from "@/types/models/recipe";

export function RecipesPage() {
  const navigate = useNavigate();
  const isMobile = useMobile();
  const [search, setSearch] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [plannerFilter, setPlannerFilter] = useState<"" | "ready" | "not-ready">("");
  const [sort, setSort] = useState<RecipeSort | undefined>(undefined);

  // Extract classification from selected tags
  const selectedClassification = selectedTags.find((t) => (CLASSIFICATIONS as readonly string[]).includes(t));
  const selectedNonClassTags = selectedTags.filter((t) => !(CLASSIFICATIONS as readonly string[]).includes(t));

  const { data, isLoading, refetch } = useRecipes({
    search: search || undefined,
    tags: selectedNonClassTags.length > 0 ? selectedNonClassTags : undefined,
    classification: selectedClassification,
    plannerReady: plannerFilter === "ready" ? true : plannerFilter === "not-ready" ? false : undefined,
    sort,
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
        <PageHeader
          title="Recipes"
          actions={
            <Button size="sm" onClick={() => navigate("/app/recipes/new")}>
              <Plus className="mr-1 h-4 w-4" />
              New Recipe
            </Button>
          }
        />

        {/* Search + sort */}
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search aria-hidden="true" className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              aria-label="Search recipes"
              placeholder="Search recipes..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
            />
            {search && (
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                aria-label="Clear search"
                onClick={() => setSearch("")}
                className="absolute right-1.5 top-1/2 -translate-y-1/2 text-muted-foreground"
              >
                <X aria-hidden="true" className="h-4 w-4" />
              </Button>
            )}
          </div>
          <RecipeSortSelect value={sort} onChange={setSort} />
        </div>

        {/* Unified filter row */}
        <div className="flex flex-wrap gap-1.5">
            {/* Planner status filters */}
            <Button
              type="button"
              size="sm"
              variant={plannerFilter === "ready" ? "default" : "outline"}
              aria-pressed={plannerFilter === "ready"}
              onClick={() => setPlannerFilter(plannerFilter === "ready" ? "" : "ready")}
            >
              Planner Ready
            </Button>
            <Button
              type="button"
              size="sm"
              variant={plannerFilter === "not-ready" ? "default" : "outline"}
              aria-pressed={plannerFilter === "not-ready"}
              onClick={() => setPlannerFilter(plannerFilter === "not-ready" ? "" : "not-ready")}
            >
              Not Ready
            </Button>

            {/* Separator */}
            {availableTags.length > 0 && (
              <span className="border-l border-border mx-1" />
            )}

            {/* Tag filters (includes classification tags) */}
            {availableTags.map((t) => (
              <Button
                key={t.attributes.tag}
                type="button"
                size="sm"
                variant={selectedTags.includes(t.attributes.tag) ? "default" : "outline"}
                aria-pressed={selectedTags.includes(t.attributes.tag)}
                onClick={() => toggleTag(t.attributes.tag)}
              >
                {toTitleCase(t.attributes.tag)} ({t.attributes.count})
              </Button>
            ))}

            {hasActiveFilters && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => { setSelectedTags([]); setPlannerFilter(""); }}
              >
                Clear
              </Button>
            )}
          </div>

        {/* Loading */}
        {isLoading && (
          <div className={isMobile ? "space-y-3" : "grid grid-cols-2 lg:grid-cols-3 gap-4"}>
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-28 w-full" />
            ))}
          </div>
        )}

        {/* Empty state */}
        {!isLoading && recipes.length === 0 && (
          <EmptyState
            icon={BookOpen}
            title="No recipes yet"
            description="Add a recipe to start building meal plans."
            action={
              <Button onClick={() => navigate("/app/recipes/new")}>
                <Plus className="mr-1 h-4 w-4" />
                Create your first recipe
              </Button>
            }
          />
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
