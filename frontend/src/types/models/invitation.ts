export interface InvitationAttributes {
  email: string;
  role: "admin" | "editor" | "viewer";
  status: "pending" | "accepted" | "declined" | "revoked";
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface Invitation {
  id: string;
  type: "invitations";
  attributes: InvitationAttributes;
  relationships: {
    household: { data: { type: "households"; id: string } };
    invitedBy: { data: { type: "users"; id: string } };
  };
}

export interface InvitationCreateAttributes {
  email: string;
  role?: "admin" | "editor" | "viewer";
}

// --- Label maps ---

export const invitationRoleLabelMap: Record<InvitationAttributes["role"], string> = {
  admin: "Admin",
  editor: "Editor",
  viewer: "Viewer",
};

export const invitationStatusLabelMap: Record<InvitationAttributes["status"], string> = {
  pending: "Pending",
  accepted: "Accepted",
  declined: "Declined",
  revoked: "Revoked",
};

// --- Helpers ---

export function isInvitationExpired(invitation: Invitation): boolean {
  return new Date(invitation.attributes.expiresAt) <= new Date();
}

export function isInvitationPending(invitation: Invitation): boolean {
  return invitation.attributes.status === "pending" && !isInvitationExpired(invitation);
}
