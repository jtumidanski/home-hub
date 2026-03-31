import { useState, useMemo } from "react";
import { toast } from "sonner";
import { useIngredients } from "@/lib/hooks/api/use-ingredients";
import { useIngredientCategories, useBulkCategorize } from "@/lib/hooks/api/use-ingredient-categories";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { CanonicalIngredientListItem } from "@/types/models/ingredient";

export function BulkCategorize() {
  const { data: ingredientsData, isLoading: ingredientsLoading } = useIngredients({ pageSize: 200 });
  const { data: categoriesData, isLoading: categoriesLoading } = useIngredientCategories();
  const bulkCategorize = useBulkCategorize();

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [filterCategory, setFilterCategory] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [targetCategoryId, setTargetCategoryId] = useState<string>("");

  const allIngredients: CanonicalIngredientListItem[] = useMemo(() => ingredientsData?.data ?? [], [ingredientsData]);
  const categories = categoriesData?.data ?? [];

  const filteredIngredients = useMemo(() => {
    let result = allIngredients;
    if (filterCategory === "uncategorized") {
      result = result.filter((i) => !i.attributes.categoryId);
    } else if (filterCategory && filterCategory !== "all") {
      result = result.filter((i) => i.attributes.categoryId === filterCategory);
    }
    if (search) {
      const lower = search.toLowerCase();
      result = result.filter((i) => i.attributes.name.toLowerCase().includes(lower));
    }
    return result;
  }, [allIngredients, filterCategory, search]);

  const totalIngredients = ingredientsData?.meta?.total ?? allIngredients.length;
  const categorizedCount = categories.reduce((sum, c) => sum + c.attributes.ingredient_count, 0);
  const uncategorizedCount = Math.max(0, totalIngredients - categorizedCount);

  const toggleSelect = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (selected.size === filteredIngredients.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(filteredIngredients.map((i) => i.id)));
    }
  };

  const handleApply = async () => {
    if (selected.size === 0 || !targetCategoryId) return;
    try {
      await bulkCategorize.mutateAsync({
        ingredientIds: Array.from(selected),
        categoryId: targetCategoryId,
      });
      toast.success(`${selected.size} ingredient${selected.size > 1 ? "s" : ""} categorized`);
      setSelected(new Set());
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to bulk categorize").message);
    }
  };

  if (ingredientsLoading || categoriesLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {uncategorizedCount > 0 && (
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{uncategorizedCount} uncategorized</Badge>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-2">
        <Select value={filterCategory} onValueChange={(v) => setFilterCategory(v ?? "all")}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="uncategorized">Uncategorized</SelectItem>
            {categories.map((cat) => (
              <SelectItem key={cat.id} value={cat.id}>
                {cat.attributes.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Input
          placeholder="Search..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="h-8 w-48 text-sm"
        />
      </div>

      <div className="rounded-md border">
        <div className="flex items-center gap-2 border-b px-3 py-2 text-sm font-medium">
          <input
            type="checkbox"
            checked={selected.size > 0 && selected.size === filteredIngredients.length}
            onChange={toggleAll}
            className="rounded"
          />
          <span>
            {selected.size > 0 ? `${selected.size} selected` : `${filteredIngredients.length} ingredients`}
          </span>
        </div>
        <div className="max-h-64 overflow-y-auto">
          {filteredIngredients.map((ing) => (
            <label
              key={ing.id}
              className="flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted cursor-pointer"
            >
              <input
                type="checkbox"
                checked={selected.has(ing.id)}
                onChange={() => toggleSelect(ing.id)}
                className="rounded"
              />
              <span className="flex-1">{ing.attributes.name}</span>
              {ing.attributes.categoryName ? (
                <span className="text-xs text-muted-foreground">{ing.attributes.categoryName}</span>
              ) : (
                <span className="text-xs text-yellow-600">uncategorized</span>
              )}
            </label>
          ))}
          {filteredIngredients.length === 0 && (
            <p className="px-3 py-4 text-sm text-muted-foreground text-center">No ingredients match filters</p>
          )}
        </div>
      </div>

      {selected.size > 0 && (
        <div className="sticky bottom-0 flex items-center gap-2 rounded-md border bg-background p-3">
          <span className="text-sm font-medium">{selected.size} selected</span>
          <Select value={targetCategoryId} onValueChange={(v) => setTargetCategoryId(v ?? "")}>
            <SelectTrigger>
              <SelectValue placeholder="Choose category..." />
            </SelectTrigger>
            <SelectContent>
              {categories.map((cat) => (
                <SelectItem key={cat.id} value={cat.id}>
                  {cat.attributes.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            size="sm"
            onClick={handleApply}
            disabled={!targetCategoryId || bulkCategorize.isPending}
          >
            Apply
          </Button>
        </div>
      )}
    </div>
  );
}
