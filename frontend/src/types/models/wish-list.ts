export type Urgency = "must_have" | "need_to_have" | "want";

export interface WishListItemAttributes {
  name: string;
  purchase_location: string | null;
  urgency: Urgency;
  vote_count: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface WishListItem {
  id: string;
  type: "wish-items";
  attributes: WishListItemAttributes;
}

export interface WishListItemCreateAttributes {
  name: string;
  purchase_location?: string;
  urgency?: Urgency;
}

export interface WishListItemUpdateAttributes {
  name?: string;
  purchase_location?: string;
  urgency?: Urgency;
}
