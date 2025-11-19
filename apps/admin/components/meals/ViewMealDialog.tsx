"use client";

import { useEffect, useState } from "react";
import { getMeal, Meal, getIngredients, Ingredient } from "@/lib/api/meals";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, Edit, Sparkles, Loader2 } from "lucide-react";
import { IngredientTable } from "./IngredientTable";

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
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [ingredients, setIngredients] = useState<Ingredient[]>([]);
  const [ingredientsLoading, setIngredientsLoading] = useState(false);
  const [ingredientsError, setIngredientsError] = useState<string | null>(null);

  // Fetch meal and ingredients when mealId changes
  useEffect(() => {
    if (open && mealId) {
      fetchData();
      fetchIngredients();
    }
  }, [open, mealId]);

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setMeal(null);
      setError(null);
      setIngredients([]);
      setIngredientsError(null);
    }
  }, [open]);

  const fetchData = async () => {
    if (!mealId) return;

    try {
      setLoading(true);
      setError(null);

      const mealData = await getMeal(mealId);
      setMeal(mealData);
    } catch (err) {
      console.error("Failed to fetch meal data:", err);
      setError(err instanceof Error ? err.message : "Failed to fetch meal data");
    } finally {
      setLoading(false);
    }
  };

  const fetchIngredients = async () => {
    if (!mealId) return;

    try {
      setIngredientsLoading(true);
      setIngredientsError(null);

      const data = await getIngredients(mealId);
      setIngredients(data);
    } catch (err) {
      console.error("Failed to fetch ingredients:", err);
      setIngredientsError(err instanceof Error ? err.message : "Failed to fetch ingredients");
    } finally {
      setIngredientsLoading(false);
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
      <DialogContent className="w-[90vw] max-w-[1400px] sm:w-[90vw] sm:max-w-[1400px] h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>
            {loading ? "Loading..." : meal?.title || "Meal Details"}
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
        {!loading && !error && meal && (
          <div className="flex-1 overflow-hidden py-4 flex flex-col gap-4">
            {/* Meal Metadata */}
            <div className="flex-shrink-0 grid grid-cols-3 gap-4 text-sm">
              {meal.description && (
                <div className="col-span-3">
                  <span className="font-medium text-neutral-700 dark:text-neutral-300">Description: </span>
                  <span className="text-neutral-600 dark:text-neutral-400">{meal.description}</span>
                </div>
              )}
              <div>
                <span className="font-medium text-neutral-700 dark:text-neutral-300">Created: </span>
                <span className="text-neutral-600 dark:text-neutral-400">{formatDate(meal.createdAt)}</span>
              </div>
              <div>
                <span className="font-medium text-neutral-700 dark:text-neutral-300">Last Updated: </span>
                <span className="text-neutral-600 dark:text-neutral-400">{formatDate(meal.updatedAt)}</span>
              </div>
            </div>

            {/* Ingredients Card */}
            <Card className="flex-1 flex flex-col overflow-hidden">
              <CardHeader className="flex-shrink-0">
                <CardTitle className="flex items-center gap-2">
                  Ingredients
                  {ingredients.length > 0 && (
                    <div className="flex items-center gap-1 text-sm font-normal text-purple-600 dark:text-purple-400">
                      <Sparkles className="h-4 w-4" />
                      <span>{ingredients.length} parsed</span>
                    </div>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent className="flex-1 overflow-hidden">
                {ingredientsLoading && (
                  <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-8 w-8 animate-spin text-neutral-400" />
                    <span className="ml-2 text-neutral-600 dark:text-neutral-400">
                      Loading ingredients...
                    </span>
                  </div>
                )}

                {ingredientsError && !ingredientsLoading && (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>
                      {ingredientsError}
                      <Button
                        variant="outline"
                        size="sm"
                        className="mt-2"
                        onClick={fetchIngredients}
                      >
                        Retry
                      </Button>
                    </AlertDescription>
                  </Alert>
                )}

                {!ingredientsLoading && !ingredientsError && ingredients.length > 0 && (
                  <div className="h-full overflow-auto">
                    <IngredientTable ingredients={ingredients} />
                  </div>
                )}

                {!ingredientsLoading && !ingredientsError && ingredients.length === 0 && (
                  <p className="text-center py-8 text-neutral-500">
                    No ingredients were parsed for this meal
                  </p>
                )}
              </CardContent>
            </Card>
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
