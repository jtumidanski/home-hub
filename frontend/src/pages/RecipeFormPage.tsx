import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeft, Loader2, Clock, Users, ExternalLink } from "lucide-react";
import { toast } from "sonner";
import { recipeFormSchema, recipeFormDefaults } from "@/lib/schemas/recipe.schema";
import type { RecipeFormData } from "@/lib/schemas/recipe.schema";
import { useRecipe, useCreateRecipe, useUpdateRecipe } from "@/lib/hooks/api/use-recipes";
import { useCooklangPreview } from "@/lib/hooks/use-cooklang-preview";
import { CooklangPreview } from "@/components/features/recipes/cooklang-preview";
import { CooklangHelp } from "@/components/features/recipes/cooklang-help";
import { useMobile } from "@/lib/hooks/use-mobile";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { createErrorFromUnknown } from "@/lib/api/errors";

export function RecipeFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isMobile = useMobile();
  const isEditing = !!id;

  const { data: existingData } = useRecipe(id ?? "");
  const createRecipe = useCreateRecipe();
  const updateRecipe = useUpdateRecipe();

  const form = useForm<RecipeFormData>({
    resolver: zodResolver(recipeFormSchema),
    defaultValues: recipeFormDefaults,
  });

  const sourceValue = form.watch("source");
  const preview = useCooklangPreview(sourceValue ?? "");

  // Populate form for edit mode
  useEffect(() => {
    if (isEditing && existingData?.data) {
      const attrs = existingData.data.attributes;
      form.reset({
        title: attrs.title,
        description: attrs.description ?? "",
        source: attrs.source,
      });
    }
  }, [isEditing, existingData, form]);

  const onSubmit = async (values: RecipeFormData) => {
    const attrs = {
      title: values.title,
      description: values.description || undefined,
      source: values.source,
    };

    try {
      if (isEditing) {
        await updateRecipe.mutateAsync({ id, attrs });
        toast.success("Recipe updated");
        navigate(`/app/recipes/${id}`);
      } else {
        const result = await createRecipe.mutateAsync(attrs);
        toast.success("Recipe created");
        navigate(`/app/recipes/${result.data.id}`);
      }
    } catch (error) {
      toast.error(createErrorFromUnknown(error, `Failed to ${isEditing ? "update" : "create"} recipe`).message);
    }
  };

  // Derived metadata from the Cooklang source preview
  const meta = preview.metadata;
  const derivedTags = meta?.tags ?? [];
  const derivedSource = meta?.source ?? "";
  const derivedServings = meta?.servings ?? "";
  const derivedPrepTime = meta?.prepTime ?? "";
  const derivedCookTime = meta?.cookTime ?? "";

  return (
    <div className="p-4 md:p-6 max-w-6xl">
      <Button variant="ghost" size="sm" className="mb-4" onClick={() => navigate(isEditing ? `/app/recipes/${id}` : "/app/recipes")}>
        <ArrowLeft className="mr-1 h-4 w-4" /> {isEditing ? "Back to recipe" : "Recipes"}
      </Button>

      <h1 className="text-2xl font-bold mb-6">{isEditing ? "Edit Recipe" : "New Recipe"}</h1>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          {/* Title and description */}
          <div className="space-y-4">
            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input placeholder="Recipe title" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Input placeholder="Brief description (optional)" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          {/* Cooklang editor + preview */}
          <div className={isMobile ? "space-y-4" : "grid grid-cols-2 gap-6"}>
            <div className="space-y-2">
              <FormField
                control={form.control}
                name="source"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Recipe (Cooklang)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={"---\nsource: https://example.com\ntags: italian, pasta\nservings: 4\nprep time: 10 minutes\ncook time: 20 minutes\n---\n\nBring @water{2%l} to a boil in a #large pot{}.\n\nCook @spaghetti{400%g} until al dente."}
                        className="min-h-[400px] font-mono text-sm"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <CooklangHelp />
            </div>

            <div className="rounded-md border p-4 bg-muted/30 min-h-[400px] space-y-4">
              {/* Show derived metadata from Cooklang source */}
              {(derivedTags.length > 0 || derivedSource || derivedServings || derivedPrepTime || derivedCookTime) && (
                <div className="space-y-2 border-b pb-3">
                  <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                    {derivedServings && (
                      <span className="inline-flex items-center gap-1">
                        <Users className="h-3 w-3" /> {derivedServings}
                      </span>
                    )}
                    {derivedPrepTime && (
                      <span className="inline-flex items-center gap-1">
                        <Clock className="h-3 w-3" /> {derivedPrepTime} prep
                      </span>
                    )}
                    {derivedCookTime && (
                      <span className="inline-flex items-center gap-1">
                        <Clock className="h-3 w-3" /> {derivedCookTime} cook
                      </span>
                    )}
                    {derivedSource && (
                      <a href={derivedSource} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1 text-primary hover:underline">
                        <ExternalLink className="h-3 w-3" /> Source
                      </a>
                    )}
                  </div>
                  {derivedTags.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {derivedTags.map((tag) => (
                        <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
                      ))}
                    </div>
                  )}
                </div>
              )}

              <CooklangPreview
                ingredients={preview.ingredients}
                steps={preview.steps}
                errors={preview.errors}
                isLoading={preview.isLoading}
              />
            </div>
          </div>

          <div className="flex gap-3">
            <Button type="submit" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEditing ? "Update Recipe" : "Create Recipe"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => navigate(isEditing ? `/app/recipes/${id}` : "/app/recipes")}
            >
              Cancel
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
