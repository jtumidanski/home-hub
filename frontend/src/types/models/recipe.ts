export interface Ingredient {
  name: string;
  quantity: string;
  unit: string;
}

export interface Segment {
  type: "text" | "ingredient" | "cookware" | "timer" | "reference";
  value?: string;
  name?: string;
  quantity?: string;
  unit?: string;
  path?: string;
}

export interface Step {
  number: number;
  section?: string;
  segments: Segment[];
}

export interface ParseError {
  line: number;
  column: number;
  message: string;
}

export interface RecipeListAttributes {
  title: string;
  description?: string;
  servings?: number;
  prepTimeMinutes?: number;
  cookTimeMinutes?: number;
  tags: string[];
  createdAt: string;
  updatedAt: string;
}

export interface RecipeDetailAttributes extends RecipeListAttributes {
  sourceUrl?: string;
  source: string;
  ingredients: Ingredient[];
  steps: Step[];
}

export interface RecipeListItem {
  id: string;
  type: "recipes";
  attributes: RecipeListAttributes;
}

export interface RecipeDetail {
  id: string;
  type: "recipes";
  attributes: RecipeDetailAttributes;
}

export interface RecipeCreateAttributes {
  title: string;
  description?: string | undefined;
  source: string;
  servings?: number | undefined;
  prepTimeMinutes?: number | undefined;
  cookTimeMinutes?: number | undefined;
  sourceUrl?: string | undefined;
  tags?: string[] | undefined;
}

export type RecipeUpdateAttributes = {
  title?: string | undefined;
  description?: string | undefined;
  source?: string | undefined;
  servings?: number | undefined;
  prepTimeMinutes?: number | undefined;
  cookTimeMinutes?: number | undefined;
  sourceUrl?: string | undefined;
  tags?: string[] | undefined;
};

export interface RecipeTag {
  id: string;
  type: "recipe-tags";
  attributes: {
    tag: string;
    count: number;
  };
}

export interface RecipeMetadata {
  tags?: string[];
  source?: string;
  title?: string;
  servings?: string;
  prepTime?: string;
  cookTime?: string;
  notes?: string[];
  extra?: Record<string, string>;
}

export interface RecipeParseResult {
  id: string;
  type: "recipe-parse";
  attributes: {
    ingredients: Ingredient[];
    steps: Step[];
    metadata: RecipeMetadata;
    errors?: ParseError[];
  };
}
