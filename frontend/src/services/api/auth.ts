import { api } from "@/lib/api/client";
import type { JsonApiResponse, JsonApiListResponse } from "@/types/api/responses";
import type { User } from "@/types/models/user";

interface AuthProvider {
  type: "auth-providers";
  id: string;
  attributes: { displayName: string };
}

export const authService = {
  getProviders: () =>
    api.get<JsonApiListResponse<AuthProvider>>("/auth/providers"),

  getMe: () =>
    api.get<JsonApiResponse<User>>("/users/me"),

  refreshToken: () =>
    api.postNoContent("/auth/token/refresh"),

  logout: () =>
    api.postNoContent("/auth/logout"),

  getLoginUrl: (provider: string, redirect: string = "/app") =>
    `/api/v1/auth/login/${provider}?redirect=${encodeURIComponent(redirect)}`,
};
