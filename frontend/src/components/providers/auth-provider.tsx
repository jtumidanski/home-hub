import { createContext, useContext, type ReactNode } from "react";
import { useMe } from "@/lib/hooks/api/use-auth";
import { useAppContext } from "@/lib/hooks/api/use-context";
import type { User } from "@/types/models/user";
import type { AppContext } from "@/types/models/context";

interface AuthContextValue {
  user: User | null;
  appContext: AppContext | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  needsOnboarding: boolean;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const meQuery = useMe();
  const user = meQuery.data?.data ?? null;

  const contextQuery = useAppContext(!!user);
  const appContext = contextQuery.data?.data ?? null;

  const isLoading = meQuery.isLoading || (!!user && contextQuery.isLoading);
  const isAuthenticated = !!user;

  // Needs onboarding if authenticated but no tenant/household context
  const needsOnboarding = isAuthenticated && !contextQuery.isLoading && (
    !appContext?.relationships?.tenant?.data?.id ||
    !appContext?.relationships?.activeHousehold?.data?.id
  );

  return (
    <AuthContext.Provider value={{ user, appContext, isLoading, isAuthenticated, needsOnboarding }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
