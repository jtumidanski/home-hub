import { useState } from "react";
import { Search, Plus } from "lucide-react";
import { toast } from "sonner";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useIngredients, useCreateIngredient } from "@/lib/hooks/api/use-ingredients";
import { useResolveIngredient } from "@/lib/hooks/api/use-ingredient-normalization";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { CanonicalIngredientListItem } from "@/types/models/ingredient";

interface IngredientResolverProps {
  recipeId: string;
  ingredientId: string;
  rawName: string;
}

export function IngredientResolver({ recipeId, ingredientId, rawName }: IngredientResolverProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState(rawName);
  const [saveAsAlias, setSaveAsAlias] = useState(true);

  const { data } = useIngredients({ search: searchQuery, pageSize: 10 });
  const resolveIngredient = useResolveIngredient();
  const createIngredient = useCreateIngredient();

  const suggestions = (data?.data ?? []) as CanonicalIngredientListItem[];

  const handleResolve = async (canonicalId: string) => {
    try {
      await resolveIngredient.mutateAsync({
        recipeId,
        ingredientId,
        canonicalIngredientId: canonicalId,
        saveAsAlias,
      });
      toast.success("Ingredient resolved");
      setIsOpen(false);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to resolve").message);
    }
  };

  const handleCreateAndResolve = async () => {
    try {
      const result = await createIngredient.mutateAsync({ name: searchQuery });
      await handleResolve(result.data.id);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create ingredient").message);
    }
  };

  if (!isOpen) {
    return (
      <Button variant="outline" size="sm" className="text-xs h-6 px-2" onClick={() => setIsOpen(true)}>
        Resolve
      </Button>
    );
  }

  return (
    <div className="mt-1 p-2 border rounded-md bg-background shadow-sm space-y-2 w-64">
      <div className="relative">
        <Search className="absolute left-2 top-1/2 h-3 w-3 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="h-7 text-xs pl-7"
          placeholder="Search ingredients..."
          autoFocus
        />
      </div>

      <div className="max-h-32 overflow-y-auto space-y-0.5">
        {suggestions.map((s) => (
          <button
            key={s.id}
            type="button"
            className="w-full text-left px-2 py-1 text-xs hover:bg-muted rounded transition-colors"
            onClick={() => handleResolve(s.id)}
          >
            {s.attributes.name}
            {s.attributes.unitFamily && (
              <span className="text-muted-foreground ml-1">({s.attributes.unitFamily})</span>
            )}
          </button>
        ))}
        {suggestions.length === 0 && searchQuery && (
          <p className="text-xs text-muted-foreground px-2 py-1">No matches found</p>
        )}
      </div>

      <div className="flex items-center gap-1.5">
        <label className="flex items-center gap-1 text-xs text-muted-foreground cursor-pointer">
          <input
            type="checkbox"
            checked={saveAsAlias}
            onChange={(e) => setSaveAsAlias(e.target.checked)}
            className="h-3 w-3"
          />
          Save as alias
        </label>
      </div>

      <div className="flex gap-1.5">
        {searchQuery && (
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="text-xs h-6 flex-1"
            onClick={handleCreateAndResolve}
            disabled={createIngredient.isPending}
          >
            <Plus className="mr-1 h-3 w-3" />
            Create &quot;{searchQuery}&quot;
          </Button>
        )}
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="text-xs h-6"
          onClick={() => setIsOpen(false)}
        >
          Cancel
        </Button>
      </div>
    </div>
  );
}
