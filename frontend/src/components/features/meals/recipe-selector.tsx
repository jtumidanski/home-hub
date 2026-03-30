import { useState, useMemo } from "react";
import { Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useRecipes } from "@/lib/hooks/api/use-recipes";
import { CLASSIFICATIONS } from "@/lib/constants/recipe";
import type { Slot } from "@/types/models/meal-plan";
import type { RecipeListItem } from "@/types/models/recipe";

interface RecipeSelectorProps {
  autoClassification?: Slot;
  onSelectRecipe: (recipe: RecipeListItem) => void;
}

export function RecipeSelector({ autoClassification, onSelectRecipe }: RecipeSelectorProps) {
  const [search, setSearch] = useState("");
  const [classification, setClassification] = useState<string>(autoClassification ?? "");

  const { data, isLoading } = useRecipes({
    search: search || undefined,
    classification: classification || undefined,
    plannerReady: true,
    pageSize: 50,
  });

  const recipes = useMemo(() => data?.data ?? [], [data]);

  return (
    <div className="space-y-3">
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search recipes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8"
          />
          {search && (
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-1 top-1 h-6 w-6"
              onClick={() => setSearch("")}
            >
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>
        <Select value={classification} onValueChange={(v) => setClassification(v ?? "")}>
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="All types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All types</SelectItem>
            {CLASSIFICATIONS.map((c) => (
              <SelectItem key={c} value={c}>
                {c.charAt(0).toUpperCase() + c.slice(1)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="max-h-[300px] overflow-y-auto space-y-1">
        {isLoading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))
        ) : recipes.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">No recipes found</p>
        ) : (
          recipes.map((recipe) => (
            <button
              key={recipe.id}
              className="w-full text-left p-2 rounded hover:bg-accent flex items-center justify-between gap-2"
              onClick={() => onSelectRecipe(recipe)}
            >
              <div className="min-w-0">
                <div className="font-medium text-sm truncate">{recipe.attributes.title}</div>
                <div className="text-xs text-muted-foreground flex items-center gap-2">
                  {recipe.attributes.classification && (
                    <Badge variant="secondary" className="text-[10px] px-1 py-0">
                      {recipe.attributes.classification}
                    </Badge>
                  )}
                  {recipe.attributes.servings && (
                    <span>serves {recipe.attributes.servings}</span>
                  )}
                  {recipe.attributes.lastUsedDate && (
                    <span>last used {recipe.attributes.lastUsedDate}</span>
                  )}
                  {(recipe.attributes.usageCount ?? 0) > 0 && (
                    <span>used {recipe.attributes.usageCount}x</span>
                  )}
                </div>
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  );
}
