export interface MembershipAttributes {
  role: "owner" | "admin" | "editor" | "viewer";
  isLastOwner?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Membership {
  id: string;
  type: "memberships";
  attributes: MembershipAttributes;
  relationships: {
    household: { data: { type: "households"; id: string } };
    user: { data: { type: "users"; id: string } };
  };
}

// --- Label maps ---

export const membershipRoleLabelMap: Record<MembershipAttributes["role"], string> = {
  owner: "Owner",
  admin: "Admin",
  editor: "Editor",
  viewer: "Viewer",
};

// --- Helpers ---

export function isMembershipPrivileged(membership: Membership): boolean {
  return membership.attributes.role === "owner" || membership.attributes.role === "admin";
}
