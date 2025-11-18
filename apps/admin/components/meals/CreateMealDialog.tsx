"use client";

import { useState, useEffect } from "react";
import { createMeal } from "@/lib/api/meals";
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
import { AlertCircle, Sparkles } from "lucide-react";

interface CreateMealDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function CreateMealDialog({
  open,
  onOpenChange,
  onSuccess,
}: CreateMealDialogProps) {
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [ingredientText, setIngredientText] = useState("");

  // Reset form when dialog closes
  useEffect(() => {
    if (!open) {
      setTitle("");
      setDescription("");
      setIngredientText("");
      setError(null);
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim()) {
      setError("Title is required");
      return;
    }

    try {
      setCreating(true);
      setError(null);

      await createMeal({
        title: title.trim(),
        description: description.trim() || undefined,
        ingredientText: ingredientText.trim() || undefined,
      });

      // Success - notify parent and close
      onSuccess();
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to create meal:", err);
      setError(err instanceof Error ? err.message : "Failed to create meal");
    } finally {
      setCreating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add New Meal</DialogTitle>
          <DialogDescription>
            Create a new meal or recipe with AI-powered ingredient parsing
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-4 py-4">
            {/* Title */}
            <div className="space-y-2">
              <Label htmlFor="title">Title *</Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g., Spaghetti Carbonara, Chicken Tacos"
                required
                autoFocus
              />
              <p className="text-sm text-neutral-500 dark:text-neutral-400">
                Give your meal a descriptive name
              </p>
            </div>

            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="A quick and easy weeknight dinner..."
                rows={3}
              />
              <p className="text-sm text-neutral-500 dark:text-neutral-400">
                Optional notes or instructions for this meal
              </p>
            </div>

            {/* Ingredient Text */}
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Label htmlFor="ingredientText">Ingredients</Label>
                <Sparkles className="h-4 w-4 text-purple-500" />
                <span className="text-sm text-purple-600 dark:text-purple-400">
                  AI-powered parsing
                </span>
              </div>
              <Textarea
                id="ingredientText"
                value={ingredientText}
                onChange={(e) => setIngredientText(e.target.value)}
                placeholder={"1 lb ground beef\n2 cups diced tomatoes\n1 tbsp olive oil\n1/2 tsp salt\n1/4 tsp black pepper"}
                rows={8}
                className="font-mono text-sm"
              />
              <p className="text-sm text-neutral-500 dark:text-neutral-400">
                Enter ingredients one per line. Our AI will parse quantities, units, and preparation steps.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={creating}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={creating || !title.trim()}
            >
              {creating ? "Creating..." : "Create Meal"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
