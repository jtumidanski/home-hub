export interface ContextAttributes {
  resolvedTheme: "light" | "dark";
  resolvedRole: "admin" | "member" | "owner";
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

// --- Label maps (F15) ---

export const resolvedRoleLabelMap: Record<ContextAttributes["resolvedRole"], string> = {
  admin: "Admin",
  member: "Member",
  owner: "Owner",
};

export const resolvedThemeLabelMap: Record<ContextAttributes["resolvedTheme"], string> = {
  light: "Light",
  dark: "Dark",
};

// --- Helpers (F16) ---

export function isContextAdmin(ctx: AppContext): boolean {
  return ctx.attributes.resolvedRole === "admin";
}
