import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, X } from "lucide-react";
import { toast } from "sonner";
import { useIngredients, useCreateIngredient, useDeleteIngredient } from "@/lib/hooks/api/use-ingredients";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardHeader, CardTitle, CardContent, CardAction } from "@/components/ui/card";
import { CardActionMenu } from "@/components/common/card-action-menu";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { CanonicalIngredientListItem } from "@/types/models/ingredient";

export function IngredientsPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState("");

  const { data, isLoading, refetch } = useIngredients(search ? { search } : undefined);
  const createIngredient = useCreateIngredient();
  const deleteIngredient = useDeleteIngredient();

  const ingredients = (data?.data ?? []) as CanonicalIngredientListItem[];

  const handleRefresh = useCallback(async () => {
    await refetch();
  }, [refetch]);

  const handleCreate = async () => {
    if (!newName.trim()) return;
    try {
      const result = await createIngredient.mutateAsync({ name: newName.trim() });
      toast.success("Ingredient created");
      setNewName("");
      setShowCreate(false);
      navigate(`/app/ingredients/${result.data.id}`);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create ingredient").message);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteIngredient.mutateAsync(id);
      toast.success("Ingredient deleted");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete ingredient").message);
    }
  };

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 md:p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Ingredients</h1>
          <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
            <Plus className="mr-1 h-4 w-4" />
            New Ingredient
          </Button>
        </div>

        {showCreate && (
          <div className="flex gap-2">
            <Input
              placeholder="Ingredient name..."
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              autoFocus
            />
            <Button onClick={handleCreate} disabled={!newName.trim() || createIngredient.isPending}>
              Create
            </Button>
            <Button variant="ghost" onClick={() => { setShowCreate(false); setNewName(""); }}>
              Cancel
            </Button>
          </div>
        )}

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search ingredients..."
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

        {/* Loading */}
        {isLoading && (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))}
          </div>
        )}

        {/* Empty state */}
        {!isLoading && ingredients.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground mb-2">No canonical ingredients yet.</p>
            <p className="text-sm text-muted-foreground mb-4">
              Ingredients are added automatically when you create recipes, or you can add them manually.
            </p>
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="mr-1 h-4 w-4" />
              Add your first ingredient
            </Button>
          </div>
        )}

        {/* Ingredient list */}
        {!isLoading && ingredients.length > 0 && (
          <div className="space-y-2">
            {ingredients.map((ingredient) => (
              <Card
                key={ingredient.id}
                size="sm"
                className="cursor-pointer"
                onClick={() => navigate(`/app/ingredients/${ingredient.id}`)}
              >
                <CardHeader>
                  <CardTitle className="text-base">{ingredient.attributes.name}</CardTitle>
                  <CardAction>
                    <CardActionMenu
                      actions={[
                        {
                          icon: <X className="h-4 w-4" />,
                          label: "Delete",
                          onClick: () => handleDelete(ingredient.id),
                          variant: "destructive",
                        },
                      ]}
                    />
                  </CardAction>
                </CardHeader>
                <CardContent>
                  <div className="flex flex-wrap items-center gap-2">
                    {ingredient.attributes.displayName && (
                      <span className="text-sm text-muted-foreground">{ingredient.attributes.displayName}</span>
                    )}
                    {ingredient.attributes.unitFamily && (
                      <Badge variant="secondary" className="text-xs">{ingredient.attributes.unitFamily}</Badge>
                    )}
                    <span className="text-xs text-muted-foreground">
                      {ingredient.attributes.aliasCount} aliases
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {ingredient.attributes.usageCount} uses
                    </span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </PullToRefresh>
  );
}
