export interface UserAttributes {
  email: string;
  displayName: string;
  givenName: string;
  familyName: string;
  avatarUrl: string;
  providerAvatarUrl: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  type: "users";
  attributes: UserAttributes;
}

// --- Update attributes (F14) ---

export type UserUpdateAttributes = Partial<
  Pick<UserAttributes, "displayName" | "givenName" | "familyName" | "avatarUrl">
>;

// --- Helpers (F16) ---

export function getUserDisplayName(user: User): string {
  const { displayName, givenName, familyName } = user.attributes;
  if (displayName) {
    return displayName;
  }
  return [givenName, familyName].filter(Boolean).join(" ");
}
