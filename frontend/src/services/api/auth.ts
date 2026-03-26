import { api, type RequestOptions } from "@/lib/api/client";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";
import type { User, UserUpdateAttributes } from "@/types/models/user";

export interface AuthProvider {
  type: "auth-providers";
  id: string;
  attributes: { slug: string; displayName: string };
}

class AuthService {
  getProviders() {
    return api.get<ApiListResponse<AuthProvider>>("/auth/providers");
  }

  getMe(options?: RequestOptions) {
    return api.get<ApiResponse<User>>("/users/me", options);
  }

  refreshToken() {
    return api.post("/auth/token/refresh");
  }

  logout() {
    return api.post("/auth/logout");
  }

  updateMe(attributes: UserUpdateAttributes) {
    return api.patch<ApiResponse<User>>("/users/me", {
      data: { type: "users", attributes },
    });
  }

  getUsersByIds(ids: string[]) {
    return api.get<ApiListResponse<User>>(
      `/users?filter[ids]=${ids.join(",")}`,
    );
  }

  getLoginUrl(provider: string, redirect: string = "/app") {
    return `/api/v1/auth/login/${provider}?redirect=${encodeURIComponent(redirect)}`;
  }
}

export const authService = new AuthService();
