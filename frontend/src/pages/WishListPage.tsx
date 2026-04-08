import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Pencil, Trash2, ThumbsUp, Heart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useWishListItems,
  useCreateWishListItem,
  useUpdateWishListItem,
  useDeleteWishListItem,
  useVoteWishListItem,
} from "@/lib/hooks/api/use-wish-list";
import {
  wishListItemSchema,
  type WishListItemFormData,
} from "@/lib/schemas/wish-list.schema";
import type { Urgency, WishListItem } from "@/types/models/wish-list";

const URGENCY_LABELS: Record<Urgency, string> = {
  must_have: "Must have",
  need_to_have: "Need to have",
  want: "Want",
};

const URGENCY_BADGE: Record<Urgency, "default" | "secondary" | "outline"> = {
  must_have: "default",
  need_to_have: "secondary",
  want: "outline",
};

export function WishListPage() {
  const { data, isLoading } = useWishListItems();
  const createItem = useCreateWishListItem();
  const updateItem = useUpdateWishListItem();
  const deleteItem = useDeleteWishListItem();
  const voteItem = useVoteWishListItem();

  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<WishListItem | null>(null);
  const [pendingDelete, setPendingDelete] = useState<WishListItem | null>(null);

  const form = useForm<WishListItemFormData>({
    resolver: zodResolver(wishListItemSchema),
    defaultValues: { name: "", purchase_location: "", urgency: "want" },
  });

  useEffect(() => {
    if (showForm) {
      if (editing) {
        form.reset({
          name: editing.attributes.name,
          purchase_location: editing.attributes.purchase_location ?? "",
          urgency: editing.attributes.urgency,
        });
      } else {
        form.reset({ name: "", purchase_location: "", urgency: "want" });
      }
    }
  }, [showForm, editing, form]);

  const items = data?.data ?? [];

  const onSubmit = (values: WishListItemFormData) => {
    const trimmedLocation = values.purchase_location?.trim() ?? "";
    const attrs: {
      name: string;
      urgency: Urgency;
      purchase_location?: string;
    } = {
      name: values.name,
      urgency: values.urgency,
    };
    if (trimmedLocation) {
      attrs.purchase_location = trimmedLocation;
    }
    if (editing) {
      updateItem.mutate(
        { id: editing.id, attrs },
        {
          onSuccess: () => {
            setShowForm(false);
            setEditing(null);
          },
        },
      );
    } else {
      createItem.mutate(attrs, {
        onSuccess: () => {
          setShowForm(false);
        },
      });
    }
  };

  const openCreate = () => {
    setEditing(null);
    setShowForm(true);
  };

  const openEdit = (item: WishListItem) => {
    setEditing(item);
    setShowForm(true);
  };

  const closeForm = () => {
    setShowForm(false);
    setEditing(null);
  };

  return (
    <div className="p-4 md:p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Wish List</h1>
        <Button onClick={openCreate}>
          <Plus className="h-4 w-4 mr-1" /> Add item
        </Button>
      </div>

      <div className="space-y-3">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24 w-full" />
          ))
        ) : items.length === 0 ? (
          <Card>
            <CardContent className="py-10 text-center text-muted-foreground">
              <Heart className="h-12 w-12 mx-auto mb-3 opacity-50" />
              <p>Nothing on the wish list yet. Add the first thing you've been eyeing.</p>
            </CardContent>
          </Card>
        ) : (
          items.map((item) => {
            const a = item.attributes;
            return (
              <Card key={item.id}>
                <CardContent className="p-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex-1 min-w-0 space-y-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <p className="font-medium truncate">{a.name}</p>
                      <Badge variant={URGENCY_BADGE[a.urgency]}>
                        {URGENCY_LABELS[a.urgency]}
                      </Badge>
                    </div>
                    {a.purchase_location && (
                      <p className="text-sm text-muted-foreground truncate">
                        {a.purchase_location}
                      </p>
                    )}
                  </div>

                  <div className="flex items-center gap-2 sm:justify-end">
                    <Button
                      type="button"
                      variant="default"
                      size="lg"
                      className="h-12 min-w-[88px] gap-2"
                      onClick={() => voteItem.mutate(item.id)}
                      aria-label={`Vote for ${a.name}`}
                    >
                      <ThumbsUp className="h-5 w-5" />
                      <span className="font-semibold tabular-nums">{a.vote_count}</span>
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => openEdit(item)}
                      aria-label={`Edit ${a.name}`}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => setPendingDelete(item)}
                      aria-label={`Delete ${a.name}`}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })
        )}
      </div>

      {/* Create / Edit Dialog */}
      <Dialog open={showForm} onOpenChange={(o) => !o && closeForm()}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit wish list item" : "Add wish list item"}</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="What do you want?" autoFocus {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="purchase_location"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Purchase location (optional)</FormLabel>
                    <FormControl>
                      <Input placeholder="Amazon, Costco, store URL…" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="urgency"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Urgency</FormLabel>
                    <Select
                      value={field.value}
                      onValueChange={(v) => v && field.onChange(v as Urgency)}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="must_have">Must have</SelectItem>
                        <SelectItem value="need_to_have">Need to have</SelectItem>
                        <SelectItem value="want">Want</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="button" variant="outline" onClick={closeForm}>
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={createItem.isPending || updateItem.isPending}
                >
                  {editing
                    ? updateItem.isPending
                      ? "Saving…"
                      : "Save"
                    : createItem.isPending
                      ? "Adding…"
                      : "Add"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <Dialog
        open={!!pendingDelete}
        onOpenChange={(o) => !o && setPendingDelete(null)}
      >
        <DialogContent className="sm:max-w-[380px]">
          <DialogHeader>
            <DialogTitle>Delete wish list item?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            “{pendingDelete?.attributes.name}” will be removed from your wish list.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPendingDelete(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              disabled={deleteItem.isPending}
              onClick={() => {
                if (!pendingDelete) return;
                deleteItem.mutate(pendingDelete.id, {
                  onSuccess: () => setPendingDelete(null),
                });
              }}
            >
              {deleteItem.isPending ? "Deleting…" : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
