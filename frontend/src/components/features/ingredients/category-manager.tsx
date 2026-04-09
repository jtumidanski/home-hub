import { useState } from "react";
import { Plus, X, Pencil, Check } from "lucide-react";
import { toast } from "sonner";
import {
  useIngredientCategories,
  useCreateCategory,
  useUpdateCategory,
  useDeleteCategory,
} from "@/lib/hooks/api/use-ingredient-categories";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { categoryNameSchema } from "@/lib/schemas/ingredient-category.schema";

export function CategoryManager() {
  const { data, isLoading } = useIngredientCategories();
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory();
  const deleteCategory = useDeleteCategory();

  const [newName, setNewName] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string; ingredientCount: number } | null>(null);

  const categories = data?.data ?? [];

  const handleCreate = async () => {
    const result = categoryNameSchema.safeParse({ name: newName });
    if (!result.success) {
      toast.error(result.error?.issues[0]?.message ?? "Validation failed");
      return;
    }
    try {
      await createCategory.mutateAsync({ name: result.data.name });
      toast.success("Category created");
      setNewName("");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to create category").message);
    }
  };

  const handleRename = async (id: string) => {
    const result = categoryNameSchema.safeParse({ name: editName });
    if (!result.success) {
      toast.error(result.error?.issues[0]?.message ?? "Validation failed");
      return;
    }
    try {
      await updateCategory.mutateAsync({ id, attrs: { name: result.data.name } });
      toast.success("Category renamed");
      setEditingId(null);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to rename category").message);
    }
  };

  const requestDelete = (id: string, name: string, ingredientCount: number) => {
    if (ingredientCount > 0) {
      setDeleteTarget({ id, name, ingredientCount });
      return;
    }
    void performDelete(id);
  };

  const performDelete = async (id: string) => {
    try {
      await deleteCategory.mutateAsync(id);
      toast.success("Category deleted");
      setDeleteTarget(null);
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to delete category").message);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="space-y-1">
        {categories.map((cat) => (
          <div key={cat.id} className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
            {editingId === cat.id ? (
              <>
                <Input
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleRename(cat.id)}
                  className="h-7 text-sm flex-1"
                  autoFocus
                />
                <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={() => handleRename(cat.id)}>
                  <Check className="h-3.5 w-3.5" />
                </Button>
                <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={() => setEditingId(null)}>
                  <X className="h-3.5 w-3.5" />
                </Button>
              </>
            ) : (
              <>
                <span className="flex-1">{cat.attributes.name}</span>
                <span className="text-xs text-muted-foreground">
                  {cat.attributes.ingredient_count ?? 0} ingredient{(cat.attributes.ingredient_count ?? 0) !== 1 ? "s" : ""}
                </span>
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-7 w-7 p-0"
                  onClick={() => { setEditingId(cat.id); setEditName(cat.attributes.name); }}
                >
                  <Pencil className="h-3.5 w-3.5" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                  onClick={() => requestDelete(cat.id, cat.attributes.name, cat.attributes.ingredient_count ?? 0)}
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              </>
            )}
          </div>
        ))}
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="New category name..."
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleCreate()}
          className="h-8 text-sm"
        />
        <Button size="sm" onClick={handleCreate} disabled={!newName.trim() || createCategory.isPending}>
          <Plus className="mr-1 h-4 w-4" />
          Add
        </Button>
      </div>

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete category</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Delete &ldquo;{deleteTarget?.name}&rdquo;? {deleteTarget?.ingredientCount} ingredient
            {(deleteTarget?.ingredientCount ?? 0) > 1 ? "s" : ""} will become uncategorized.
          </p>
          <div className="flex gap-2 justify-end">
            <Button variant="outline" size="sm" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              disabled={deleteCategory.isPending}
              onClick={() => deleteTarget && void performDelete(deleteTarget.id)}
            >
              Delete
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
