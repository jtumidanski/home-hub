export interface MemberAttributes {
  displayName: string;
  role: string;
}

export interface Member {
  id: string;
  type: "members";
  attributes: MemberAttributes;
  relationships: {
    user: { data: { type: "users"; id: string } };
    household: { data: { type: "households"; id: string } };
  };
}
