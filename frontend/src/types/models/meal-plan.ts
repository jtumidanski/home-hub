// Plan list item (from GET /meals/plans)
export interface PlanListAttributes {
  starts_on: string;
  name: string;
  locked: boolean;
  item_count: number;
  created_at: string;
  updated_at: string;
}

export interface PlanListItem {
  id: string;
  type: "plans";
  attributes: PlanListAttributes;
}

// Plan item within detail response
export interface PlanItemAttributes {
  id: string;
  day: string;
  slot: Slot;
  recipe_id: string;
  recipe_title: string;
  recipe_servings: number | null;
  recipe_classification?: string;
  recipe_deleted: boolean;
  serving_multiplier: number | null;
  planned_servings: number | null;
  notes: string | null;
  position: number;
}

// Plan detail (from GET /meals/plans/{planId})
export interface PlanDetailAttributes {
  starts_on: string;
  name: string;
  locked: boolean;
  created_by: string;
  items: PlanItemAttributes[];
  created_at: string;
  updated_at: string;
}

export interface PlanDetail {
  id: string;
  type: "plans";
  attributes: PlanDetailAttributes;
}

// Plan item (from POST/PATCH /meals/plans/{planId}/items)
export interface PlanItemResponseAttributes {
  day: string;
  slot: Slot;
  recipe_id: string;
  serving_multiplier: number | null;
  planned_servings: number | null;
  notes: string | null;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface PlanItemResponse {
  id: string;
  type: "plan-items";
  attributes: PlanItemResponseAttributes;
}

// Consolidated ingredient (from GET /meals/plans/{planId}/ingredients)
export interface PlanIngredientAttributes {
  name: string;
  display_name: string | null;
  quantity: number;
  unit: string;
  unit_family: string;
  resolved: boolean;
}

export interface PlanIngredient {
  id: string;
  type: "plan-ingredients";
  attributes: PlanIngredientAttributes;
}

// Slot enum
export type Slot = "breakfast" | "lunch" | "dinner" | "snack" | "side";
export const SLOTS: Slot[] = ["breakfast", "lunch", "dinner", "snack", "side"];

// Create/update types
export interface PlanCreateAttributes {
  starts_on: string;
  name?: string;
}

export interface PlanUpdateAttributes {
  name: string;
}

export interface PlanDuplicateAttributes {
  starts_on: string;
}

export interface PlanItemCreateAttributes {
  day: string;
  slot: Slot;
  recipe_id: string;
  serving_multiplier?: number | null;
  planned_servings?: number | null;
  notes?: string | null;
  position?: number;
}

export interface PlanItemUpdateAttributes {
  day?: string;
  slot?: Slot;
  serving_multiplier?: number | null;
  planned_servings?: number | null;
  notes?: string | null;
  position?: number;
}
