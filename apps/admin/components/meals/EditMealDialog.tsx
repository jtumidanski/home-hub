"use client";

import { useEffect, useState } from "react";
import { getMeal, updateMeal, Meal } from "@/lib/api/meals";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, Save } from "lucide-react";

interface EditMealDialogProps {
  mealId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function EditMealDialog({
  mealId,
  open,
  onOpenChange,
  onSuccess,
}: EditMealDialogProps) {
  const [meal, setMeal] = useState<Meal | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");

  // Fetch meal when mealId changes
  useEffect(() => {
    if (open && mealId) {
      fetchData();
    }
  }, [open, mealId]);

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setMeal(null);
      setTitle("");
      setDescription("");
      setError(null);
    }
  }, [open]);

  const fetchData = async () => {
    if (!mealId) return;

    try {
      setLoading(true);
      setError(null);

      const mealData = await getMeal(mealId);
      setMeal(mealData);

      // Set form state
      setTitle(mealData.title);
      setDescription(mealData.description || "");
    } catch (err) {
      console.error("Failed to fetch meal data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch meal data");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!meal || !mealId) return;

    if (!title.trim()) {
      setError("Title is required");
      return;
    }

    try {
      setSaving(true);
      setError(null);

      // Determine what changed
      const titleChanged = title !== meal.title;
      const descriptionChanged = description !== (meal.description || "");

      // Only send changed fields
      const updates: { title?: string; description?: string } = {};
      if (titleChanged) {
        updates.title = title.trim();
      }
      if (descriptionChanged) {
        updates.description = description.trim();
      }

      // Only make API call if something changed
      if (Object.keys(updates).length > 0) {
        await updateMeal(mealId, updates);
      }

      onSuccess();
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to save meal:", err);
      const errorMessage = err instanceof Error ? err.message : "Failed to save meal";
      setError(errorMessage);
    } finally {
      setSaving(false);
    }
  };

  const formatDate = (dateString: string): string => {
    if (!dateString) return "—";
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return "Invalid Date";
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {loading ? "Loading..." : `Edit ${meal?.title || "Meal"}`}
          </DialogTitle>
          <DialogDescription>
            {meal && `ID: ${meal.id}`}
          </DialogDescription>
        </DialogHeader>

        {/* Loading State */}
        {loading && (
          <div className="text-center py-12">
            <p className="text-neutral-600 dark:text-neutral-400">Loading meal...</p>
          </div>
        )}

        {/* Error State */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Content */}
        {!loading && meal && (
          <form onSubmit={handleSave}>
            <div className="py-4">
              <Card>
                <CardHeader>
                  <CardTitle>Meal Information</CardTitle>
                  <CardDescription>
                    Update the title and description
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Title */}
                  <div className="space-y-2">
                    <Label htmlFor="title">Title *</Label>
                    <Input
                      id="title"
                      value={title}
                      onChange={(e) => setTitle(e.target.value)}
                      placeholder="e.g., Spaghetti Carbonara"
                      required
                      autoFocus
                    />
                  </div>

                  {/* Description */}
                  <div className="space-y-2">
                    <Label htmlFor="description">Description</Label>
                    <Textarea
                      id="description"
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      placeholder="A quick and easy weeknight dinner..."
                      rows={4}
                    />
                  </div>

                  {/* Timestamps (read-only) */}
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Created</Label>
                      <div className="text-sm text-neutral-600 dark:text-neutral-400">
                        {formatDate(meal.createdAt)}
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label>Last Updated</Label>
                      <div className="text-sm text-neutral-600 dark:text-neutral-400">
                        {formatDate(meal.updatedAt)}
                      </div>
                    </div>
                  </div>

                  {/* Note about ingredients */}
                  <Alert>
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Note</AlertTitle>
                    <AlertDescription>
                      To modify ingredients, you'll need to delete this meal and create a new one.
                      Ingredient parsing happens only during meal creation.
                    </AlertDescription>
                  </Alert>
                </CardContent>
              </Card>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={saving}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={saving || !title.trim()}>
                <Save className="h-4 w-4 mr-2" />
                {saving ? "Saving..." : "Save Changes"}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
