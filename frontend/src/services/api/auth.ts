import { api } from "@/lib/api/client";
import type { JsonApiListResponse, JsonApiResponse } from "@/types/api/responses";
import type { User } from "@/types/models/user";

export interface AuthProvider {
  type: "auth-providers";
  id: string;
  attributes: { displayName: string };
}

class AuthService {
  getProviders() {
    return api.get<JsonApiListResponse<AuthProvider>>("/auth/providers");
  }

  getMe() {
    return api.get<JsonApiResponse<User>>("/users/me");
  }

  refreshToken() {
    return api.postNoContent("/auth/token/refresh");
  }

  logout() {
    return api.postNoContent("/auth/logout");
  }

  getLoginUrl(provider: string, redirect: string = "/app") {
    return `/api/v1/auth/login/${provider}?redirect=${encodeURIComponent(redirect)}`;
  }
}

export const authService = new AuthService();
