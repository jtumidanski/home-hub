import { useState, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  ArrowLeft,
  Plus,
  ShoppingCart,
  Check,
  RotateCcw,
  Archive,
  Trash2,
  FileDown,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import {
  useShoppingList,
  useAddShoppingItem,
  useRemoveShoppingItem,
  useCheckShoppingItem,
  useUncheckAllItems,
  useArchiveShoppingList,
  useUnarchiveShoppingList,
  useDeleteShoppingList,
  useImportMealPlan,
} from "@/lib/hooks/api/use-shopping";
import { usePlans } from "@/lib/hooks/api/use-meals";
import type { NestedShoppingItem } from "@/types/models/shopping";

interface GroupedItems {
  categoryName: string;
  sortOrder: number;
  items: NestedShoppingItem[];
}

function groupItemsByCategory(items: NestedShoppingItem[]): GroupedItems[] {
  const groups = new Map<string, GroupedItems>();

  for (const item of items) {
    const key = item.category_name ?? "__uncategorized__";
    if (!groups.has(key)) {
      groups.set(key, {
        categoryName: item.category_name ?? "Uncategorized",
        sortOrder: item.category_sort_order ?? 999999,
        items: [],
      });
    }
    groups.get(key)!.items.push(item);
  }

  return Array.from(groups.values()).sort((a, b) => a.sortOrder - b.sortOrder);
}

export function ShoppingListDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [shoppingMode, setShoppingMode] = useState(false);
  const [quickAdd, setQuickAdd] = useState("");
  const [showImport, setShowImport] = useState(false);
  const [showFinishConfirm, setShowFinishConfirm] = useState(false);

  const { data: listData, isLoading } = useShoppingList(id ?? null);
  const { data: plansData } = usePlans();

  const list = listData?.data ?? null;
  const isArchived = list?.attributes.status === "archived";

  const addItem = useAddShoppingItem(id ?? "");
  const removeItem = useRemoveShoppingItem(id ?? "");
  const checkItem = useCheckShoppingItem(id ?? "");
  const uncheckAll = useUncheckAllItems(id ?? "");
  const archiveList = useArchiveShoppingList();
  const unarchiveList = useUnarchiveShoppingList();
  const deleteList = useDeleteShoppingList();
  const importMealPlan = useImportMealPlan(id ?? "");

  const items: NestedShoppingItem[] = useMemo(
    () => listData?.data?.attributes?.items ?? [],
    [listData],
  );

  const grouped = useMemo(() => groupItemsByCategory(items), [items]);

  const checkedCount = items.filter((i) => i.checked).length;
  const totalCount = items.length;

  const handleQuickAdd = () => {
    if (!quickAdd.trim()) return;
    addItem.mutate({ name: quickAdd.trim() });
    setQuickAdd("");
  };

  const handleFinishShopping = () => {
    if (!id) return;
    archiveList.mutate(id, {
      onSuccess: () => {
        setShowFinishConfirm(false);
        navigate("/app/shopping/grocery");
      },
    });
  };

  const handleImport = (planId: string) => {
    importMealPlan.mutate(planId, {
      onSuccess: () => setShowImport(false),
    });
  };

  if (isLoading) {
    return (
      <div className="p-4 md:p-6 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!list) {
    return (
      <div className="p-4 md:p-6">
        <p className="text-muted-foreground">Shopping list not found.</p>
        <Button variant="link" onClick={() => navigate("/app/shopping/grocery")}>
          Back to lists
        </Button>
      </div>
    );
  }

  return (
    <div className="p-4 md:p-6 space-y-4">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => navigate("/app/shopping/grocery")}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">{list.attributes.name}</h1>
          {isArchived && list.attributes.archived_at && (
            <p className="text-sm text-muted-foreground">
              Archived on{" "}
              {new Date(list.attributes.archived_at).toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                year: "numeric",
              })}
            </p>
          )}
        </div>
      </div>

      {/* Action Bar */}
      {!isArchived && (
        <div className="flex gap-2 flex-wrap">
          {shoppingMode ? (
            <>
              <Button variant="outline" size="sm" onClick={() => setShoppingMode(false)}>
                <ArrowLeft className="h-4 w-4 mr-1" /> Back to Edit
              </Button>
              <Button variant="outline" size="sm" onClick={() => uncheckAll.mutate()}>
                <RotateCcw className="h-4 w-4 mr-1" /> Uncheck All
              </Button>
              <Button size="sm" onClick={() => setShowFinishConfirm(true)}>
                <Archive className="h-4 w-4 mr-1" /> Finish Shopping
              </Button>
              <span className="flex items-center text-sm text-muted-foreground ml-auto">
                {checkedCount} of {totalCount} items
              </span>
            </>
          ) : (
            <>
              <Button size="sm" onClick={() => setShoppingMode(true)}>
                <ShoppingCart className="h-4 w-4 mr-1" /> Start Shopping
              </Button>
              <Button variant="outline" size="sm" onClick={() => setShowImport(true)}>
                <FileDown className="h-4 w-4 mr-1" /> Import from Meal Plan
              </Button>
            </>
          )}
        </div>
      )}

      {isArchived && (
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => id && unarchiveList.mutate(id, { onSuccess: () => setShoppingMode(false) })}
          >
            Reopen List
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              id && deleteList.mutate(id, { onSuccess: () => navigate("/app/shopping/grocery") })
            }
          >
            <Trash2 className="h-4 w-4 mr-1" /> Delete
          </Button>
        </div>
      )}

      {/* Quick Add (edit mode only) */}
      {!isArchived && !shoppingMode && (
        <div className="flex gap-2">
          <Input
            placeholder="Quick add item..."
            value={quickAdd}
            onChange={(e) => setQuickAdd(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleQuickAdd()}
          />
          <Button onClick={handleQuickAdd} disabled={!quickAdd.trim()}>
            <Plus className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Items grouped by category */}
      {grouped.length === 0 ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            <ShoppingCart className="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>No items yet</p>
            <p className="text-sm">Add items to get started</p>
          </CardContent>
        </Card>
      ) : (
        grouped.map((group) => {
          const unchecked = group.items.filter((i) => !i.checked);
          const checked = group.items.filter((i) => i.checked);
          const displayItems = shoppingMode ? [...unchecked, ...checked] : group.items;

          return (
            <Card key={group.categoryName}>
              <CardHeader className="py-2 px-4">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {group.categoryName}
                </CardTitle>
              </CardHeader>
              <CardContent className="py-0 px-2">
                {displayItems.map((item) => (
                  <div
                    key={item.id}
                    className={cn(
                      "flex items-center gap-3 px-2 py-2 rounded hover:bg-accent/30 transition-colors",
                      shoppingMode && "cursor-pointer",
                      item.checked && "opacity-60",
                    )}
                    onClick={() => {
                      if (shoppingMode && !isArchived) {
                        checkItem.mutate({ itemId: item.id, checked: !item.checked });
                      }
                    }}
                  >
                    {(shoppingMode || isArchived) && (
                      <div
                        className={cn(
                          "h-5 w-5 rounded border flex items-center justify-center flex-shrink-0",
                          item.checked
                            ? "bg-primary border-primary text-primary-foreground"
                            : "border-muted-foreground",
                        )}
                      >
                        {item.checked && <Check className="h-3 w-3" />}
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <span className={cn("text-sm", item.checked && shoppingMode && "line-through")}>
                        {item.name}
                      </span>
                      {item.quantity && (
                        <Badge variant="secondary" className="ml-2 text-xs">
                          {item.quantity}
                        </Badge>
                      )}
                    </div>
                    {!shoppingMode && !isArchived && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 flex-shrink-0"
                        onClick={(e) => {
                          e.stopPropagation();
                          removeItem.mutate(item.id);
                        }}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                ))}
              </CardContent>
            </Card>
          );
        })
      )}

      {/* Finish Shopping Confirmation */}
      <Dialog open={showFinishConfirm} onOpenChange={(o) => !o && setShowFinishConfirm(false)}>
        <DialogContent className="sm:max-w-[350px]">
          <DialogHeader>
            <DialogTitle>Finish Shopping?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            This will archive the list. You can view it later in the History tab.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowFinishConfirm(false)}>
              Cancel
            </Button>
            <Button onClick={handleFinishShopping} disabled={archiveList.isPending}>
              {archiveList.isPending ? "Archiving..." : "Finish Shopping"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Import from Meal Plan */}
      <Dialog open={showImport} onOpenChange={(o) => !o && setShowImport(false)}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>Import from Meal Plan</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {(plansData?.data ?? []).length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                No meal plans available
              </p>
            ) : (
              (plansData?.data ?? []).map((plan) => (
                <Button
                  key={plan.id}
                  variant="outline"
                  className="w-full justify-start"
                  onClick={() => handleImport(plan.id)}
                  disabled={importMealPlan.isPending}
                >
                  <span className="truncate">{plan.attributes.name}</span>
                  <span className="ml-auto text-xs text-muted-foreground">
                    {plan.attributes.item_count} recipes
                  </span>
                </Button>
              ))
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
