import { useState, useCallback, useMemo } from "react";
import { Lock, Unlock, Copy, FileDown, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { WeekSelector } from "@/components/features/meals/week-selector";
import { WeekGrid } from "@/components/features/meals/week-grid";
import { RecipeSelector } from "@/components/features/meals/recipe-selector";
import { PlanItemPopover } from "@/components/features/meals/plan-item-popover";
import { IngredientPreview } from "@/components/features/meals/ingredient-preview";
import { ExportModal } from "@/components/features/meals/export-modal";
import {
  usePlans,
  usePlan,
  useCreatePlan,
  useLockPlan,
  useUnlockPlan,
  useDuplicatePlan,
  useAddPlanItem,
  useUpdatePlanItem,
  useRemovePlanItem,
} from "@/lib/hooks/api/use-meals";
import type { Slot, PlanItemAttributes } from "@/types/models/meal-plan";
import type { RecipeListItem } from "@/types/models/recipe";

function getMonday(): Date {
  const today = new Date();
  const day = today.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  const monday = new Date(today);
  monday.setDate(today.getDate() + diff);
  monday.setHours(0, 0, 0, 0);
  return monday;
}

function formatDateStr(d: Date): string {
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function getWeekDays(startsOn: Date) {
  const days = [];
  for (let i = 0; i < 7; i++) {
    const d = new Date(startsOn);
    d.setDate(d.getDate() + i);
    days.push({
      dateStr: formatDateStr(d),
      label: d.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" }),
    });
  }
  return days;
}

export function MealsPage() {
  const [startsOn, setStartsOn] = useState<Date>(getMonday);
  const startsOnStr = formatDateStr(startsOn);

  // Find existing plan for this week
  const { data: plansData, isLoading: plansLoading } = usePlans({ starts_on: startsOnStr });
  const existingPlanId = plansData?.data?.[0]?.id ?? null;

  // Load plan detail if exists
  const { data: planData, isLoading: planLoading } = usePlan(existingPlanId);
  const plan = planData?.data ?? null;
  const items: PlanItemAttributes[] = plan?.attributes?.items ?? [];
  const locked = plan?.attributes?.locked ?? false;

  // Mutations
  const createPlan = useCreatePlan();
  const lockPlan = useLockPlan();
  const unlockPlan = useUnlockPlan();
  const duplicatePlan = useDuplicatePlan();
  const addItem = useAddPlanItem();
  const updateItem = useUpdatePlanItem();
  const removeItem = useRemovePlanItem();

  // UI state
  const [selectorSlot, setSelectorSlot] = useState<Slot | null>(null);
  const [selectorDay, setSelectorDay] = useState<string | null>(null);
  const [selectedRecipe, setSelectedRecipe] = useState<RecipeListItem | null>(null);
  const [editItem, setEditItem] = useState<PlanItemAttributes | null>(null);
  const [showExport, setShowExport] = useState(false);
  const [showDuplicate, setShowDuplicate] = useState(false);
  const [duplicateTarget, setDuplicateTarget] = useState("");

  const weekDays = useMemo(() => getWeekDays(startsOn), [startsOn]);

  const isLoading = plansLoading || planLoading;

  // Ensure plan exists before adding items
  const ensurePlan = useCallback(async (): Promise<string | null> => {
    if (existingPlanId) return existingPlanId;
    try {
      const result = await createPlan.mutateAsync({ starts_on: startsOnStr });
      return result.data.id;
    } catch {
      return null;
    }
  }, [existingPlanId, createPlan, startsOnStr]);

  const handleCellClick = (day: string, slot: Slot) => {
    setSelectorDay(day);
    setSelectorSlot(slot);
    setSelectedRecipe(null);
    setEditItem(null);
  };

  const handleSelectRecipe = (recipe: RecipeListItem) => {
    setSelectedRecipe(recipe);
  };

  const handleAddItem = async (data: {
    day: string;
    slot: Slot;
    serving_multiplier?: number | null;
    planned_servings?: number | null;
    notes?: string | null;
  }) => {
    if (!selectedRecipe) return;
    const planId = await ensurePlan();
    if (!planId) return;
    addItem.mutate({
      planId,
      attrs: {
        ...data,
        recipe_id: selectedRecipe.id,
      },
    });
    setSelectedRecipe(null);
    setSelectorDay(null);
    setSelectorSlot(null);
  };

  const handleEditItem = async (data: {
    day: string;
    slot: Slot;
    serving_multiplier?: number | null;
    planned_servings?: number | null;
    notes?: string | null;
  }) => {
    if (!editItem || !existingPlanId) return;
    updateItem.mutate({
      planId: existingPlanId,
      itemId: editItem.id,
      attrs: data,
    });
    setEditItem(null);
  };

  const handleRemoveItem = (itemId: string) => {
    if (!existingPlanId) return;
    removeItem.mutate({ planId: existingPlanId, itemId });
  };

  const handleLockToggle = () => {
    if (!existingPlanId) return;
    if (locked) {
      unlockPlan.mutate(existingPlanId);
    } else {
      lockPlan.mutate(existingPlanId);
    }
  };

  const handleDuplicate = () => {
    if (!existingPlanId || !duplicateTarget) return;
    duplicatePlan.mutate(
      { id: existingPlanId, attrs: { starts_on: duplicateTarget } },
      { onSuccess: () => setShowDuplicate(false) },
    );
  };

  return (
    <div className="p-4 md:p-6 space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Meal Planner</h1>
          <p className="text-sm text-muted-foreground">
            {plan ? plan.attributes.name : "No plan for this week"}
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <WeekSelector startsOn={startsOn} onWeekChange={setStartsOn} />
        </div>
      </div>

      {/* Actions */}
      {existingPlanId && (
        <div className="flex gap-2 flex-wrap">
          <Button variant="outline" size="sm" onClick={handleLockToggle}>
            {locked ? <Unlock className="h-4 w-4 mr-1" /> : <Lock className="h-4 w-4 mr-1" />}
            {locked ? "Unlock" : "Lock"}
          </Button>
          <Button variant="outline" size="sm" onClick={() => setShowDuplicate(true)}>
            <Copy className="h-4 w-4 mr-1" /> Duplicate
          </Button>
          <Button variant="outline" size="sm" onClick={() => setShowExport(true)}>
            <FileDown className="h-4 w-4 mr-1" /> Export
          </Button>
        </div>
      )}

      {/* Grid */}
      <Card>
        <CardContent className="p-2 sm:p-4">
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : (
            <WeekGrid
              startsOn={startsOn}
              items={items}
              locked={locked}
              onCellClick={handleCellClick}
              onItemClick={(item) => {
                if (!locked) setEditItem(item);
              }}
              onRemoveItem={handleRemoveItem}
            />
          )}
        </CardContent>
      </Card>

      {/* Recipe Selector */}
      {selectorDay && selectorSlot && !selectedRecipe && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <Plus className="h-4 w-4" />
              Add Recipe to {selectorSlot} on{" "}
              {weekDays.find((d) => d.dateStr === selectorDay)?.label}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <RecipeSelector
              autoClassification={selectorSlot}
              onSelectRecipe={handleSelectRecipe}
            />
          </CardContent>
        </Card>
      )}

      {/* Add item popover (after recipe selected) */}
      {selectedRecipe && selectorDay && selectorSlot && (
        <PlanItemPopover
          open={true}
          onClose={() => {
            setSelectedRecipe(null);
            setSelectorDay(null);
            setSelectorSlot(null);
          }}
          onSave={handleAddItem}
          weekDays={weekDays}
          initialDay={selectorDay}
          initialSlot={selectorSlot}
          recipeServings={selectedRecipe.attributes.servings}
        />
      )}

      {/* Edit item popover */}
      {editItem && (
        <PlanItemPopover
          open={true}
          onClose={() => setEditItem(null)}
          onSave={handleEditItem}
          weekDays={weekDays}
          editItem={editItem}
          recipeServings={editItem.recipe_servings}
        />
      )}

      {/* Ingredient Preview */}
      {existingPlanId && items.length > 0 && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Consolidated Ingredients</CardTitle>
          </CardHeader>
          <CardContent>
            <IngredientPreview planId={existingPlanId} />
          </CardContent>
        </Card>
      )}

      {/* Export Modal */}
      {showExport && existingPlanId && plan && (
        <ExportModal
          open={showExport}
          onClose={() => setShowExport(false)}
          planId={existingPlanId}
          planName={plan.attributes.name}
        />
      )}

      {/* Duplicate Dialog */}
      <Dialog open={showDuplicate} onOpenChange={(o) => !o && setShowDuplicate(false)}>
        <DialogContent className="sm:max-w-[350px]">
          <DialogHeader>
            <DialogTitle>Duplicate Plan</DialogTitle>
          </DialogHeader>
          <div className="py-2">
            <label className="text-sm font-medium">Target week start date</label>
            <Input
              type="date"
              value={duplicateTarget}
              onChange={(e) => setDuplicateTarget(e.target.value)}
              className="mt-1"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDuplicate(false)}>Cancel</Button>
            <Button onClick={handleDuplicate} disabled={!duplicateTarget || duplicatePlan.isPending}>
              {duplicatePlan.isPending ? "Duplicating..." : "Duplicate"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
