import { useState, useCallback } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { AlertTriangle, Plus, Search, X, Settings, ListChecks, ChevronLeft, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { useIngredients, useCreateIngredient, useDeleteIngredient } from "@/lib/hooks/api/use-ingredients";
import { useIngredientCategories } from "@/lib/hooks/api/use-ingredient-categories";
import { PageHeader } from "@/components/common/page-header";
import { EmptyState } from "@/components/common/empty-state";
import { PullToRefresh } from "@/components/common/pull-to-refresh";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardHeader, CardTitle, CardContent, CardAction } from "@/components/ui/card";
import { CardActionMenu } from "@/components/common/card-action-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CategoryManager } from "@/components/features/ingredients/category-manager";
import { BulkCategorize } from "@/components/features/ingredients/bulk-categorize";
import { createErrorFromUnknown } from "@/lib/api/errors";
import type { CanonicalIngredientListItem } from "@/types/models/ingredient";

export function IngredientsPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [search, setSearch] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState("");
  const [view, setView] = useState<"list" | "categories" | "bulk">("list");
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const filterCategory = searchParams.get("category") ?? "all";
  const setFilterCategory = useCallback((value: string) => {
    setPage(1);
    setSearchParams((prev) => {
      if (value === "all") {
        prev.delete("category");
      } else {
        prev.set("category", value);
      }
      return prev;
    }, { replace: true });
  }, [setSearchParams]);

  const ingredientParams = {
    ...(search ? { search } : {}),
    ...(filterCategory !== "all" ? { categoryId: filterCategory === "uncategorized" ? "null" : filterCategory } : {}),
    page,
    pageSize,
  };
  const { data, isLoading, refetch } = useIngredients(ingredientParams);
  const { data: categoriesData } = useIngredientCategories();
  const createIngredient = useCreateIngredient();
  const deleteIngredient = useDeleteIngredient();

  const ingredients: CanonicalIngredientListItem[] = data?.data ?? [];
  const categories = categoriesData?.data ?? [];
  const total = data?.meta?.total ?? 0;
  const totalPages = Math.ceil(total / pageSize);

  // When viewing "all", we can derive uncategorized count from total - sum(category counts)
  // When filtered, the uncategorized count from "uncategorized" filter's total is exact
  const categorizedCount = categories.reduce((sum, c) => sum + (c.attributes.ingredient_count ?? 0), 0);
  const uncategorizedCount = filterCategory === "uncategorized"
    ? total
    : filterCategory === "all"
      ? Math.max(0, total - categorizedCount)
      : 0;

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
        <PageHeader
          title={
            <>
              Ingredients
              {uncategorizedCount > 0 && (
                <Badge variant="secondary" className="ml-3 align-middle text-xs">
                  {uncategorizedCount} uncategorized
                </Badge>
              )}
            </>
          }
          actions={
            <>
              <Button
                size="sm"
                variant={view === "categories" ? "default" : "outline"}
                aria-pressed={view === "categories"}
                onClick={() => setView(view === "categories" ? "list" : "categories")}
              >
                <Settings className="mr-1 h-4 w-4" />
                Categories
              </Button>
              <Button
                size="sm"
                variant={view === "bulk" ? "default" : "outline"}
                aria-pressed={view === "bulk"}
                onClick={() => setView(view === "bulk" ? "list" : "bulk")}
              >
                <ListChecks className="mr-1 h-4 w-4" />
                Bulk Edit
              </Button>
              <Button size="sm" onClick={() => { setView("list"); setShowCreate(!showCreate); }}>
                <Plus className="mr-1 h-4 w-4" />
                New
              </Button>
            </>
          }
        />

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

        {view === "categories" && <CategoryManager />}
        {view === "bulk" && <BulkCategorize />}

        {view === "list" && (
        <>
        {/* Category filter */}
        <div className="flex items-center gap-2">
          <Select value={filterCategory} onValueChange={(v) => setFilterCategory(v ?? "all")}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All categories</SelectItem>
              <SelectItem value="uncategorized">Uncategorized</SelectItem>
              {categories.map((cat) => (
                <SelectItem key={cat.id} value={cat.id}>
                  {cat.attributes.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search ingredients..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="pl-9"
          />
          {search && (
            <button
              type="button"
              onClick={() => setSearch("")}
              aria-label="Clear search"
              className="absolute right-1 top-1/2 flex size-8 -translate-y-1/2 cursor-pointer items-center justify-center text-muted-foreground hover:text-foreground coarse:size-11"
            >
              <X className="h-4 w-4" />
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
        {!isLoading && ingredients.length === 0 && (
          <EmptyState
            title="No canonical ingredients yet"
            description="Ingredients are added automatically when you create recipes, or you can add them manually."
            action={
              <Button onClick={() => setShowCreate(true)}>
                <Plus className="mr-1 h-4 w-4" />
                Add your first ingredient
              </Button>
            }
          />
        )}

        {/* Ingredient list */}
        {!isLoading && ingredients.length > 0 && (
          <div className="space-y-2">
            {ingredients.map((ingredient) => (
              <Card
                key={ingredient.id}
                size="sm"
                className="relative focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-ring"
              >
                <CardHeader>
                  <CardTitle className="text-base">
                    <Link
                      to={`/app/ingredients/${ingredient.id}`}
                      className="outline-none after:absolute after:inset-0 after:content-['']"
                    >
                      {ingredient.attributes.name}
                    </Link>
                  </CardTitle>
                  <CardAction className="relative z-10">
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
                    {(() => {
                      const categoryName = ingredient.attributes.categoryId
                        ? categories.find((c) => c.id === ingredient.attributes.categoryId)?.attributes.name
                        : undefined;
                      return categoryName ? (
                        <Badge variant="outline" className="text-xs">{categoryName}</Badge>
                      ) : (
                        <Badge variant="warning" className="text-xs"><AlertTriangle aria-hidden="true" />uncategorized</Badge>
                      );
                    })()}
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

        {/* Pagination */}
        {!isLoading && ingredients.length > 0 && (
          <div className="flex items-center justify-between pt-2">
            <span className="text-sm text-muted-foreground">
              Showing {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)} of {total} ingredients
            </span>
            {totalPages > 1 && (
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm px-2">
                  {page} / {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage(page + 1)}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            )}
          </div>
        )}
        </>
        )}
      </div>
    </PullToRefresh>
  );
}
