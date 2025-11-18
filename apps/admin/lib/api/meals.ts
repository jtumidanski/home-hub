/**
 * Meals Service API
 *
 * Client for interacting with the svc-meals service via the gateway.
 * Handles meal CRUD operations with AI-powered ingredient parsing.
 */

import { get, post, patch, del } from "./client";

/**
 * Ingredient parsed from AI
 */
export interface Ingredient {
  id: string;
  mealId: string;
  rawLine: string;
  quantity: number | null;
  quantityRaw: string;
  unit: string | null;
  unitRaw: string | null;
  ingredient: string;
  preparation: string[];
  notes: string[];
  confidence: number;
  createdAt: string;
  updatedAt: string;
}

/**
 * Meal attributes (from JSON:API attributes)
 */
export interface MealAttributes {
  householdId: string;
  userId: string;
  title: string;
  description?: string; // Optional due to omitempty in backend
  rawIngredientText: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * Meal interface (with id from JSON:API)
 */
export interface Meal {
  id: string;
  householdId: string;
  userId: string;
  title: string;
  description: string; // Always present after transformation (defaults to empty string)
  rawIngredientText: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * JSON:API resource object
 */
export interface JsonApiResource<T> {
  type: string;
  id: string;
  attributes: T;
}

/**
 * JSON:API single resource response
 */
export interface JsonApiSingleResponse<T> {
  data: JsonApiResource<T>;
  meta?: Record<string, any>;
}

/**
 * JSON:API array response
 */
export interface JsonApiArrayResponse<T> {
  data: JsonApiResource<T>[];
  meta?: Record<string, any>;
}

/**
 * Transform JSON:API resource to flat object with id
 */
function transformJsonApiResource(resource: JsonApiResource<MealAttributes>): Meal {
  return {
    id: resource.id,
    householdId: resource.attributes.householdId,
    userId: resource.attributes.userId,
    title: resource.attributes.title,
    description: resource.attributes.description || "", // Default to empty string
    rawIngredientText: resource.attributes.rawIngredientText,
    createdAt: resource.attributes.createdAt,
    updatedAt: resource.attributes.updatedAt,
  };
}

/**
 * Transform JSON:API array response to flat array
 */
function transformJsonApiArray(response: JsonApiArrayResponse<MealAttributes>): Meal[] {
  return response.data.map(transformJsonApiResource);
}

/**
 * Transform JSON:API single response to flat object
 */
function transformJsonApiSingle(response: JsonApiSingleResponse<MealAttributes>): Meal {
  return transformJsonApiResource(response.data);
}

/**
 * Input for creating a new meal
 */
export interface CreateMealInput {
  title: string;
  description?: string;
  ingredientText?: string;
}

/**
 * Input for updating an existing meal
 */
export interface UpdateMealInput {
  title?: string;
  description?: string;
}

/**
 * List all meals for the authenticated user's household
 */
export async function listMeals(): Promise<Meal[]> {
  const response = await get<JsonApiArrayResponse<MealAttributes>>("/meals");
  return transformJsonApiArray(response);
}

/**
 * Get a single meal by ID
 * @param mealId - Meal ID
 */
export async function getMeal(mealId: string): Promise<{ meal: Meal; ingredientCount: number }> {
  const response = await get<JsonApiSingleResponse<MealAttributes>>(`/meals/${mealId}`);
  return {
    meal: transformJsonApiSingle(response),
    ingredientCount: response.meta?.ingredient_count ? parseInt(response.meta.ingredient_count[0]) : 0,
  };
}

/**
 * Create a new meal with optional ingredient text for AI parsing
 * @param input - Meal creation data
 */
export async function createMeal(input: CreateMealInput): Promise<Meal> {
  const response = await post<JsonApiSingleResponse<MealAttributes>>("/meals", {
    data: {
      type: "meals",
      attributes: {
        title: input.title,
        description: input.description || "",
        ingredientText: input.ingredientText || "",
      },
    },
  });
  return transformJsonApiSingle(response);
}

/**
 * Update an existing meal
 * @param mealId - Meal ID
 * @param input - Meal update data
 */
export async function updateMeal(
  mealId: string,
  input: UpdateMealInput
): Promise<Meal> {
  const response = await patch<JsonApiSingleResponse<MealAttributes>>(`/meals/${mealId}`, {
    data: {
      type: "meals",
      id: mealId,
      attributes: {
        ...(input.title !== undefined && { title: input.title }),
        ...(input.description !== undefined && { description: input.description }),
      },
    },
  });
  return transformJsonApiSingle(response);
}

/**
 * Delete a meal
 * @param mealId - Meal ID
 */
export async function deleteMeal(mealId: string): Promise<void> {
  await del(`/meals/${mealId}`);
}
