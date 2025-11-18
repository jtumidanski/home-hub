"use client";

import { useEffect, useState } from "react";
import { getMeal, Meal } from "@/lib/api/meals";
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
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, Edit, Sparkles } from "lucide-react";

interface ViewMealDialogProps {
  mealId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onEdit: (mealId: string) => void;
}

export function ViewMealDialog({
  mealId,
  open,
  onOpenChange,
  onEdit,
}: ViewMealDialogProps) {
  const [meal, setMeal] = useState<Meal | null>(null);
  const [ingredientCount, setIngredientCount] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
      setIngredientCount(0);
      setError(null);
    }
  }, [open]);

  const fetchData = async () => {
    if (!mealId) return;

    try {
      setLoading(true);
      setError(null);

      const { meal: mealData, ingredientCount: count } = await getMeal(mealId);
      setMeal(mealData);
      setIngredientCount(count);
    } catch (err) {
      console.error("Failed to fetch meal data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch meal data");
    } finally {
      setLoading(false);
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

  const handleEdit = () => {
    if (mealId) {
      onEdit(mealId);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {loading ? "Loading..." : meal?.title || "Meal Details"}
          </DialogTitle>
          <DialogDescription>
            {meal && (
              <div className="flex items-center gap-2 mt-1">
                <span>ID: {meal.id}</span>
                {ingredientCount > 0 && (
                  <>
                    <span>•</span>
                    <Sparkles className="h-3 w-3 text-purple-500" />
                    <span className="text-purple-600 dark:text-purple-400">
                      {ingredientCount} parsed ingredient{ingredientCount !== 1 ? 's' : ''}
                    </span>
                  </>
                )}
              </div>
            )}
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
        {!loading && !error && meal && (
          <div className="space-y-6 py-4">
            {/* Meal Information Card */}
            <Card>
              <CardHeader>
                <CardTitle>Meal Information</CardTitle>
                <CardDescription>Details about this meal</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Title</Label>
                  <div className="text-lg font-semibold">{meal.title}</div>
                </div>

                {meal.description && (
                  <div className="space-y-2">
                    <Label>Description</Label>
                    <div className="text-neutral-700 dark:text-neutral-300">
                      {meal.description}
                    </div>
                  </div>
                )}

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
              </CardContent>
            </Card>

            {/* Ingredients Card */}
            {meal.rawIngredientText && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    Ingredients
                    {ingredientCount > 0 && (
                      <div className="flex items-center gap-1 text-sm font-normal text-purple-600 dark:text-purple-400">
                        <Sparkles className="h-4 w-4" />
                        <span>{ingredientCount} parsed</span>
                      </div>
                    )}
                  </CardTitle>
                  <CardDescription>
                    Raw ingredient list{ingredientCount > 0 ? ' (processed by AI)' : ''}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Textarea
                    value={meal.rawIngredientText}
                    disabled
                    rows={10}
                    className="font-mono text-sm"
                  />
                </CardContent>
              </Card>
            )}
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {meal && (
            <Button onClick={handleEdit}>
              <Edit className="h-4 w-4 mr-2" />
              Edit
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
