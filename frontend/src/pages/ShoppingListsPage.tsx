import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Archive, Trash2, Pencil, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import {
  useShoppingLists,
  useCreateShoppingList,
  useDeleteShoppingList,
  useUpdateShoppingList,
  useUnarchiveShoppingList,
} from "@/lib/hooks/api/use-shopping";
import type { ShoppingList } from "@/types/models/shopping";
import { shoppingListNameSchema } from "@/lib/schemas/shopping.schema";

export function ShoppingListsPage() {
  const navigate = useNavigate();
  const [tab, setTab] = useState<"active" | "archived">("active");
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [editList, setEditList] = useState<ShoppingList | null>(null);
  const [editName, setEditName] = useState("");
  const [editError, setEditError] = useState<string | null>(null);

  const { data: activeData, isLoading: activeLoading } = useShoppingLists("active");
  const { data: archivedData, isLoading: archivedLoading } = useShoppingLists("archived");

  const createList = useCreateShoppingList();
  const deleteList = useDeleteShoppingList();
  const updateList = useUpdateShoppingList();
  const unarchiveList = useUnarchiveShoppingList();

  const activeLists = activeData?.data ?? [];
  const archivedLists = archivedData?.data ?? [];
  const isLoading = tab === "active" ? activeLoading : archivedLoading;
  const lists = tab === "active" ? activeLists : archivedLists;

  const handleCreate = () => {
    const result = shoppingListNameSchema.safeParse({ name: newName });
    if (!result.success) {
      setCreateError(result.error.issues[0]?.message ?? "Invalid name");
      return;
    }
    createList.mutate(result.data.name, {
      onSuccess: (data) => {
        setShowCreate(false);
        setNewName("");
        setCreateError(null);
        navigate(`/app/shopping/${data.data.id}`);
      },
    });
  };

  const handleRename = () => {
    if (!editList) return;
    const result = shoppingListNameSchema.safeParse({ name: editName });
    if (!result.success) {
      setEditError(result.error.issues[0]?.message ?? "Invalid name");
      return;
    }
    updateList.mutate(
      { id: editList.id, name: result.data.name },
      { onSuccess: () => { setEditList(null); setEditError(null); } },
    );
  };

  const formatDate = (dateStr: string) =>
    new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });

  return (
    <div className="p-4 md:p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Shopping Lists</h1>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4 mr-1" /> New List
        </Button>
      </div>

      {/* Tab buttons */}
      <div className="flex gap-1 border-b">
        <button
          className={cn(
            "px-4 py-2 text-sm font-medium border-b-2 transition-colors",
            tab === "active"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground",
          )}
          onClick={() => setTab("active")}
        >
          Active ({activeLists.length})
        </button>
        <button
          className={cn(
            "px-4 py-2 text-sm font-medium border-b-2 transition-colors",
            tab === "archived"
              ? "border-primary text-primary"
              : "border-transparent text-muted-foreground hover:text-foreground",
          )}
          onClick={() => setTab("archived")}
        >
          History ({archivedLists.length})
        </button>
      </div>

      {/* List content */}
      <div className="space-y-3">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-20 w-full" />)
        ) : lists.length === 0 ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              {tab === "active" ? (
                <>
                  <ShoppingCart className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p>No active shopping lists</p>
                  <p className="text-sm">Create one to get started</p>
                </>
              ) : (
                <>
                  <Archive className="h-12 w-12 mx-auto mb-3 opacity-50" />
                  <p>No archived lists</p>
                </>
              )}
            </CardContent>
          </Card>
        ) : (
          lists.map((list) => (
            <Card
              key={list.id}
              className="cursor-pointer hover:bg-accent/50 transition-colors"
              onClick={() => navigate(`/app/shopping/${list.id}`)}
            >
              <CardContent className="py-3 px-4 flex items-center justify-between">
                <div>
                  <p className="font-medium">{list.attributes.name}</p>
                  <p className="text-sm text-muted-foreground">
                    {tab === "active" ? (
                      <>
                        {list.attributes.item_count} items
                        {list.attributes.checked_count > 0 &&
                          `, ${list.attributes.checked_count} checked`}
                      </>
                    ) : (
                      <>
                        {list.attributes.item_count} items
                        {list.attributes.archived_at &&
                          ` · Archived ${formatDate(list.attributes.archived_at)}`}
                      </>
                    )}
                  </p>
                </div>
                <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
                  {tab === "active" ? (
                    <>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          setEditList(list);
                          setEditName(list.attributes.name);
                        }}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => deleteList.mutate(list.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => unarchiveList.mutate(list.id)}
                      >
                        Reopen
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => deleteList.mutate(list.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={showCreate} onOpenChange={(o) => { if (!o) { setShowCreate(false); setCreateError(null); } }}>
        <DialogContent className="sm:max-w-[350px]">
          <DialogHeader>
            <DialogTitle>New Shopping List</DialogTitle>
          </DialogHeader>
          <div className="space-y-1">
            <Input
              placeholder="List name"
              value={newName}
              onChange={(e) => { setNewName(e.target.value); setCreateError(null); }}
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              autoFocus
            />
            {createError && <p className="text-sm text-destructive">{createError}</p>}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={createList.isPending}>
              {createList.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rename Dialog */}
      <Dialog open={!!editList} onOpenChange={(o) => { if (!o) { setEditList(null); setEditError(null); } }}>
        <DialogContent className="sm:max-w-[350px]">
          <DialogHeader>
            <DialogTitle>Rename List</DialogTitle>
          </DialogHeader>
          <div className="space-y-1">
            <Input
              value={editName}
              onChange={(e) => { setEditName(e.target.value); setEditError(null); }}
              onKeyDown={(e) => e.key === "Enter" && handleRename()}
              autoFocus
            />
            {editError && <p className="text-sm text-destructive">{editError}</p>}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditList(null)}>
              Cancel
            </Button>
            <Button onClick={handleRename} disabled={updateList.isPending}>
              {updateList.isPending ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
