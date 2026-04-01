export interface CanonicalIngredientAlias {
  id: string;
  name: string;
}

export interface CanonicalIngredientListAttributes {
  name: string;
  displayName?: string;
  unitFamily?: string;
  categoryId?: string;
  categoryName?: string;
  aliasCount: number;
  usageCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CanonicalIngredientDetailAttributes {
  name: string;
  displayName?: string;
  unitFamily?: string;
  categoryId?: string;
  categoryName?: string;
  aliases: CanonicalIngredientAlias[];
  createdAt: string;
  updatedAt: string;
}

export interface CanonicalIngredientListItem {
  id: string;
  type: "ingredients";
  attributes: CanonicalIngredientListAttributes;
}

export interface CanonicalIngredientDetail {
  id: string;
  type: "ingredients";
  attributes: CanonicalIngredientDetailAttributes;
}

export interface CanonicalIngredientCreateAttributes {
  name: string;
  displayName?: string;
  unitFamily?: string;
  categoryId?: string;
}

export interface CanonicalIngredientUpdateAttributes {
  name?: string;
  displayName?: string;
  unitFamily?: string;
  categoryId?: string | null;
}

export interface IngredientRecipeRef {
  recipeId: string;
  rawName: string;
}

export interface IngredientCategoryAttributes {
  name: string;
  sort_order: number;
  ingredient_count?: number;
  created_at: string;
  updated_at: string;
}

export interface IngredientCategory {
  id: string;
  type: "categories";
  attributes: IngredientCategoryAttributes;
}

export interface IngredientCategoryCreateAttributes {
  name: string;
}

export interface IngredientCategoryUpdateAttributes {
  name?: string;
  sort_order?: number;
}
