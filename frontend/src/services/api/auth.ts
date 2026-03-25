import { api } from "@/lib/api/client";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";
import type { User } from "@/types/models/user";

export interface AuthProvider {
  type: "auth-providers";
  id: string;
  attributes: { displayName: string };
}

class AuthService {
  getProviders() {
    return api.get<ApiListResponse<AuthProvider>>("/auth/providers");
  }

  getMe() {
    return api.get<ApiResponse<User>>("/users/me");
  }

  refreshToken() {
    return api.post("/auth/token/refresh");
  }

  logout() {
    return api.post("/auth/logout");
  }

  getLoginUrl(provider: string, redirect: string = "/app") {
    return `/api/v1/auth/login/${provider}?redirect=${encodeURIComponent(redirect)}`;
  }
}

export const authService = new AuthService();
