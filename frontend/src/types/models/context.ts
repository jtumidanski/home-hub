export interface ContextAttributes {
  resolvedTheme: "light" | "dark";
  resolvedRole: string;
  canCreateHousehold: boolean;
}

export interface AppContext {
  id: "current";
  type: "contexts";
  attributes: ContextAttributes;
  relationships: {
    tenant: { data: { type: "tenants"; id: string } };
    activeHousehold?: { data: { type: "households"; id: string } };
    preference: { data: { type: "preferences"; id: string } };
    memberships: { data: Array<{ type: "memberships"; id: string }> };
  };
}
