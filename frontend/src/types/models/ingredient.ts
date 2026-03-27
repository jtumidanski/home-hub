export interface CanonicalIngredientAlias {
  id: string;
  name: string;
}

export interface CanonicalIngredientListAttributes {
  name: string;
  displayName?: string;
  unitFamily?: string;
  aliasCount: number;
  usageCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CanonicalIngredientDetailAttributes {
  name: string;
  displayName?: string;
  unitFamily?: string;
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
}

export interface CanonicalIngredientUpdateAttributes {
  name?: string;
  displayName?: string;
  unitFamily?: string;
}

export interface IngredientRecipeRef {
  recipeId: string;
  rawName: string;
}
