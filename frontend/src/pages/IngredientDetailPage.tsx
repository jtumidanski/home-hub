import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, Plus, X, Trash2 } from "lucide-react";
import { toast } from "sonner";
import {
  useIngredient,
  useUpdateIngredient,
  useDeleteIngredient,
  useReassignIngredient,
  useAddAlias,
  useRemoveAlias,
  useIngredientRecipes,
} from "@/lib/hooks/api/use-ingredients";
import { useIngredients } from "@/lib/hooks/api/use-ingredients";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { UNIT_FAMILIES } from "@/lib/constants/recipe";
import { useIngredientCategories } from "@/lib/hooks/api/use-ingredient-categories";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { CanonicalIngredientListItem } from "@/types/models/ingredient";

export function IngredientDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useIngredient(id!);
  const { data: recipesData } = useIngredientRecipes(id!);
  const updateIngredient = useUpdateIngredient();
  const deleteIngredient = useDeleteIngredient();
  const reassignIngredient = useReassignIngredient();
  const addAlias = useAddAlias();
  const removeAlias = useRemoveAlias();

  const { data: categoriesData } = useIngredientCategories();
  const [newAlias, setNewAlias] = useState("");
  const [showReassign, setShowReassign] = useState(false);
  const [reassignSearch, setReassignSearch] = useState("");
  const { data: reassignCandidates } = useIngredients({ search: reassignSearch, pageSize: 10 });

  const ingredient = data?.data;
  const recipes = recipesData?.data ?? [];
  const categories = categoriesData?.data ?? [];
  const allCandidates: CanonicalIngredientListItem[] = reassignCandidates?.data ?? [];
  const candidates = allCandidates.filter((c) => c.id !== id);

  const handleAddAlias = async () => {
    if (!newAlias.trim() || !id) return;
    try {
      await addAlias.mutateAsync({ ingredientId: id, aliasName: newAlias.trim() });
      toast.success("Alias added");
      setNewAlias("");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to add alias").message);
    }
  };

  const handleRemoveAlias = async (aliasId: string) => {
    if (!id) return;
    try {
      await removeAlias.mutateAsync({ ingredientId: id, aliasId });
      toast.success("Alias removed");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to remove alias").message);
    }
  };

  const handleDelete = async () => {
    if (!id) return;
    try {
      await deleteIngredient.mutateAsync(id);
      toast.success("Ingredient deleted");
      navigate(-1);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete ingredient").message);
      // If has references, show reassign option
      setShowReassign(true);
    }
  };

  const handleReassign = async (targetId: string) => {
    if (!id) return;
    try {
      await reassignIngredient.mutateAsync({ id, targetId });
      toast.success("Ingredient reassigned and deleted");
      navigate(-1);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to reassign").message);
    }
  };

  const handleUpdateCategory = async (categoryId: string) => {
    if (!id) return;
    try {
      const value = categoryId === "none" ? null : categoryId;
      await updateIngredient.mutateAsync({ id, attrs: { categoryId: value } });
      toast.success("Category updated");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to update").message);
    }
  };

  const handleUpdateUnitFamily = async (unitFamily: string) => {
    if (!id) return;
    try {
      await updateIngredient.mutateAsync({ id, attrs: unitFamily ? { unitFamily } : { unitFamily: "" } });
      toast.success("Unit family updated");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to update").message);
    }
  };

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-96" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (!ingredient) {
    return (
      <div className="p-4 md:p-6">
        <p className="text-muted-foreground">Ingredient not found.</p>
        <Button variant="ghost" className="mt-2" onClick={() => navigate(-1)}>
          <ArrowLeft className="mr-1 h-4 w-4" /> Back to ingredients
        </Button>
      </div>
    );
  }

  const attrs = ingredient.attributes;

  return (
    <div className="p-4 md:p-6 space-y-6 max-w-3xl">
      {/* Header */}
      <div className="space-y-3">
        <Button variant="ghost" size="sm" onClick={() => navigate(-1)}>
          <ArrowLeft className="mr-1 h-4 w-4" /> Ingredients
        </Button>

        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{attrs.name}</h1>
            {attrs.displayName && (
              <p className="mt-1 text-muted-foreground">{attrs.displayName}</p>
            )}
          </div>
          <Button variant="outline" size="sm" className="text-destructive shrink-0" onClick={handleDelete}>
            <Trash2 className="mr-1 h-4 w-4" /> Delete
          </Button>
        </div>
      </div>

      {/* Unit family */}
      <div>
        <h3 className="text-sm font-medium mb-2">Unit Family</h3>
        <div className="flex gap-1.5">
          {UNIT_FAMILIES.map((uf) => (
            <Button
              key={uf || "none"}
              variant={(attrs.unitFamily ?? "") === uf ? "default" : "outline"}
              size="sm"
              className="text-xs"
              onClick={() => handleUpdateUnitFamily(uf)}
            >
              {uf || "Unset"}
            </Button>
          ))}
        </div>
      </div>

      {/* Category */}
      <div>
        <h3 className="text-sm font-medium mb-2">Category</h3>
        <Select value={attrs.categoryId ?? "none"} onValueChange={(v) => v && handleUpdateCategory(v)}>
          <SelectTrigger>
            <SelectValue placeholder="Uncategorized">
              {attrs.categoryId
                ? categories.find((c) => c.id === attrs.categoryId)?.attributes.name ?? "Uncategorized"
                : "Uncategorized"}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">Uncategorized</SelectItem>
            {categories.map((cat) => (
              <SelectItem key={cat.id} value={cat.id}>
                {cat.attributes.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Aliases */}
      <div>
        <h3 className="text-sm font-medium mb-2">Aliases</h3>
        <div className="space-y-2">
          {attrs.aliases.length > 0 ? (
            <div className="flex flex-wrap gap-1.5">
              {attrs.aliases.map((alias) => (
                <Badge key={alias.id} variant="secondary" className="pr-1">
                  {alias.name}
                  <button
                    type="button"
                    className="ml-1 hover:text-destructive"
                    onClick={() => handleRemoveAlias(alias.id)}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </Badge>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No aliases.</p>
          )}
          <div className="flex gap-2">
            <Input
              placeholder="Add alias..."
              value={newAlias}
              onChange={(e) => setNewAlias(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleAddAlias()}
              className="h-8 text-sm"
            />
            <Button size="sm" onClick={handleAddAlias} disabled={!newAlias.trim() || addAlias.isPending}>
              <Plus className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Linked recipes */}
      <div>
        <h3 className="text-sm font-medium mb-2">Used in Recipes</h3>
        {recipes.length > 0 ? (
          <ul className="space-y-1">
            {recipes.map((ref, i) => (
              <li key={i}>
                <button
                  type="button"
                  className="text-sm text-primary hover:underline"
                  onClick={() => navigate(`/app/recipes/${ref.recipeId}`)}
                >
                  {ref.recipeName || "(unknown recipe)"}
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">Not used in any recipes.</p>
        )}
      </div>

      {/* Reassign dialog */}
      {showReassign && (
        <div className="border rounded-md p-4 space-y-3">
          <h3 className="text-sm font-medium">Reassign references before deleting</h3>
          <p className="text-xs text-muted-foreground">
            This ingredient is referenced by recipe ingredients. Choose another ingredient to reassign all references to.
          </p>
          <Input
            placeholder="Search for target ingredient..."
            value={reassignSearch}
            onChange={(e) => setReassignSearch(e.target.value)}
            className="h-8 text-sm"
          />
          <div className="max-h-32 overflow-y-auto space-y-1">
            {candidates.map((c) => (
              <button
                key={c.id}
                type="button"
                className="w-full text-left px-2 py-1 text-sm hover:bg-muted rounded"
                onClick={() => handleReassign(c.id)}
              >
                {c.attributes.name}
              </button>
            ))}
          </div>
          <Button variant="ghost" size="sm" onClick={() => setShowReassign(false)}>
            Cancel
          </Button>
        </div>
      )}
    </div>
  );
}
