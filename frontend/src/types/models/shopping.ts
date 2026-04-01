export interface ShoppingListAttributes {
  name: string;
  status: "active" | "archived";
  item_count: number;
  checked_count: number;
  archived_at: string | null;
  items?: ShoppingItemAttributes[];
  created_at: string;
  updated_at: string;
}

export interface ShoppingList {
  id: string;
  type: "shopping-lists";
  attributes: ShoppingListAttributes;
}

export interface ShoppingItemAttributes {
  name: string;
  quantity: string | null;
  category_id: string | null;
  category_name: string | null;
  category_sort_order: number | null;
  checked: boolean;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface ShoppingItem {
  id: string;
  type: "shopping-items";
  attributes: ShoppingItemAttributes;
}

export interface ShoppingListCreateAttributes {
  name: string;
}

export interface ShoppingListUpdateAttributes {
  name: string;
}

export interface ShoppingItemCreateAttributes {
  name: string;
  quantity?: string;
  category_id?: string;
  position?: number;
}

export interface ShoppingItemUpdateAttributes {
  name?: string;
  quantity?: string;
  category_id?: string;
  position?: number;
}

export interface ShoppingItemCheckAttributes {
  checked: boolean;
}

export interface ShoppingListImportAttributes {
  plan_id: string;
}
