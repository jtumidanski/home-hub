import { useEffect, useState } from "react";
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
import { PlannerConfigForm } from "@/components/features/recipes/planner-config-form";
import { RecipeIngredients } from "@/components/features/recipes/recipe-ingredients";
import { useMobile } from "@/lib/hooks/use-mobile";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { extractClassification } from "@/lib/constants/recipe";
import { ensureFrontmatter } from "@/lib/cooklang/frontmatter";
import type { PlannerConfig } from "@/types/models/recipe";

export function RecipeFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isMobile = useMobile();
  const isEditing = !!id;

  const { data: existingData } = useRecipe(id ?? "");
  const createRecipe = useCreateRecipe();
  const updateRecipe = useUpdateRecipe();

  const [plannerConfig, setPlannerConfig] = useState<PlannerConfig>({});

  const form = useForm<RecipeFormData>({
    resolver: zodResolver(recipeFormSchema),
    defaultValues: recipeFormDefaults,
  });

  // eslint-disable-next-line react-hooks/incompatible-library -- form.watch() returns unmemoizable values; library-level React Compiler limitation
  const sourceValue = form.watch("source");
  const preview = useCooklangPreview(sourceValue ?? "");

  // Populate form for edit mode. For legacy recipes whose source lacks
  // title/description in frontmatter, inject them from the DB columns so the
  // user sees and edits the values in the cooklang editor.
  useEffect(() => {
    if (isEditing && existingData?.data) {
      const attrs = existingData.data.attributes;
      form.reset({
        source: ensureFrontmatter(attrs.source, attrs.title, attrs.description ?? ""),
      });
      if (attrs.plannerConfig) {
        setPlannerConfig(attrs.plannerConfig);
      }
    }
  }, [isEditing, existingData, form]);

  // Derived metadata from the Cooklang source preview
  const meta = preview.metadata;
  const derivedTitle = meta?.title?.trim() ?? "";
  const derivedDescription = meta?.description?.trim() ?? "";
  const derivedTags = meta?.tags ?? [];
  const derivedSource = meta?.source ?? "";
  const derivedServings = meta?.servings ?? "";
  const derivedPrepTime = meta?.prepTime ?? "";
  const derivedCookTime = meta?.cookTime ?? "";

  const onSubmit = async (values: RecipeFormData) => {
    if (!derivedTitle) {
      toast.error("Add a `title:` line to the recipe frontmatter");
      return;
    }

    const classification = extractClassification(derivedTags);
    const parsedServings = derivedServings ? parseInt(derivedServings, 10) : undefined;
    const servingsYield = parsedServings && !isNaN(parsedServings) ? parsedServings : undefined;

    const mergedConfig: PlannerConfig = {
      ...(plannerConfig.eatWithinDays ? { eatWithinDays: plannerConfig.eatWithinDays } : {}),
      ...(plannerConfig.minGapDays ? { minGapDays: plannerConfig.minGapDays } : {}),
      ...(plannerConfig.maxConsecutiveDays ? { maxConsecutiveDays: plannerConfig.maxConsecutiveDays } : {}),
      ...(classification ? { classification } : {}),
      ...(servingsYield ? { servingsYield } : {}),
    };
    const hasPlanner = mergedConfig.classification || mergedConfig.servingsYield || mergedConfig.eatWithinDays || mergedConfig.minGapDays || mergedConfig.maxConsecutiveDays;
    const attrs = {
      title: derivedTitle,
      description: derivedDescription || undefined,
      source: values.source,
      plannerConfig: hasPlanner ? mergedConfig : undefined,
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

  const existingIngredients = existingData?.data?.attributes?.ingredients ?? [];

  return (
    <div className="p-4 md:p-6 max-w-6xl">
      <Button variant="ghost" size="sm" className="mb-4" onClick={() => navigate(isEditing ? `/app/recipes/${id}` : "/app/recipes")}>
        <ArrowLeft className="mr-1 h-4 w-4" /> {isEditing ? "Back to recipe" : "Recipes"}
      </Button>

      <h1 className="text-2xl font-bold mb-6">{isEditing ? "Edit Recipe" : "New Recipe"}</h1>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <PlannerConfigForm value={plannerConfig} onChange={setPlannerConfig} />

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
                        placeholder={"---\ntitle: My Recipe\ndescription: A short summary\nsource: https://example.com\ntags: italian, pasta\nservings: 4\nprep time: 10 minutes\ncook time: 20 minutes\n---\n\nBring @water{2%l} to a boil in a #large pot{}.\n\nCook @spaghetti{400%g} until al dente."}
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

            <div className="space-y-2">
              <label className="text-sm font-medium leading-none">Preview</label>
              <div className="rounded-md border p-4 bg-muted/30 min-h-[400px] space-y-4">
                {(derivedTitle || derivedDescription) && (
                  <div className="space-y-1 border-b pb-3">
                    {derivedTitle ? (
                      <h2 className="text-lg font-semibold leading-tight">{derivedTitle}</h2>
                    ) : (
                      <p className="text-sm italic text-muted-foreground">Missing title — add `title:` to frontmatter</p>
                    )}
                    {derivedDescription && (
                      <p className="text-sm text-muted-foreground">{derivedDescription}</p>
                    )}
                  </div>
                )}

                {(derivedTags.length > 0 || derivedSource || derivedServings || derivedPrepTime || derivedCookTime) && (
                  <div className="space-y-2 border-b pb-3">
                    <div className="space-y-1 text-xs text-muted-foreground">
                      {derivedServings && (
                        <div className="flex items-center gap-1.5">
                          <Users className="h-3 w-3 shrink-0" />
                          <span className="text-muted-foreground/70">Servings:</span>
                          <span>{derivedServings}</span>
                        </div>
                      )}
                      {derivedPrepTime && (
                        <div className="flex items-center gap-1.5">
                          <Clock className="h-3 w-3 shrink-0" />
                          <span className="text-muted-foreground/70">Prep Time:</span>
                          <span>{derivedPrepTime}</span>
                        </div>
                      )}
                      {derivedCookTime && (
                        <div className="flex items-center gap-1.5">
                          <Clock className="h-3 w-3 shrink-0" />
                          <span className="text-muted-foreground/70">Cook Time:</span>
                          <span>{derivedCookTime}</span>
                        </div>
                      )}
                      {derivedSource && (
                        <div className="flex items-center gap-1.5">
                          <ExternalLink className="h-3 w-3 shrink-0" />
                          <span className="text-muted-foreground/70">Source:</span>
                          <a href={derivedSource} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline truncate">
                            {derivedSource}
                          </a>
                        </div>
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
                  notes={preview.notes}
                  isLoading={preview.isLoading}
                />

                {preview.normalization && preview.normalization.length > 0 && (
                  <div className="border-t pt-3">
                    <h4 className="text-xs font-medium text-muted-foreground mb-2">Normalization Preview</h4>
                    <ul className="space-y-1">
                      {preview.normalization.map((n, i) => (
                        <li key={i} className="text-xs flex items-center gap-1.5">
                          <span className={n.status === "unresolved" ? "text-yellow-500" : "text-green-500"}>
                            {n.status === "unresolved" ? "?" : "✓"}
                          </span>
                          <span>{n.rawName}</span>
                          {n.canonicalName && (
                            <span className="text-muted-foreground">→ {n.canonicalName}</span>
                          )}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </div>
          </div>

          {isEditing && existingIngredients.length > 0 && (
            <div>
              <h3 className="text-sm font-medium mb-2">Ingredient Normalization</h3>
              <RecipeIngredients ingredients={existingIngredients} recipeId={id!} />
            </div>
          )}

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
