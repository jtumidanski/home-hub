export interface UserAttributes {
  email: string;
  displayName: string;
  givenName: string;
  familyName: string;
  avatarUrl: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  type: "users";
  attributes: UserAttributes;
}
