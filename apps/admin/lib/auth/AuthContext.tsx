'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import { User, fetchMe } from './api';

interface AuthState {
  user: User | null;
  roles: string[];
  loading: boolean;
  error: Error | null;
}

interface AuthContextValue extends AuthState {
  refetch: () => Promise<void>;
  hasRole: (role: string) => boolean;
  hasAnyRole: (roles: string[]) => boolean;
  isAdmin: () => boolean;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export interface AuthProviderProps {
  children: React.ReactNode;
}

/**
 * AuthProvider fetches and provides the current authenticated user's information
 * to all child components via React Context.
 *
 * Usage:
 * ```tsx
 * <AuthProvider>
 *   <YourApp />
 * </AuthProvider>
 * ```
 */
export function AuthProvider({ children }: AuthProviderProps) {
  const [state, setState] = useState<AuthState>({
    user: null,
    roles: [],
    loading: true,
    error: null,
  });

  const fetchUser = async () => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const data = await fetchMe();

      if (data) {
        setState({
          user: data.user,
          roles: data.roles,
          loading: false,
          error: null,
        });
      } else {
        // Not authenticated
        setState({
          user: null,
          roles: [],
          loading: false,
          error: null,
        });
      }
    } catch (error) {
      console.error('Failed to fetch user:', error);
      setState({
        user: null,
        roles: [],
        loading: false,
        error: error instanceof Error ? error : new Error('Failed to fetch user'),
      });
    }
  };

  useEffect(() => {
    // Fetch user data on mount - this is a legitimate use case for setState in effect
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchUser();
  }, []);

  const hasRole = (role: string): boolean => {
    return state.roles.includes(role);
  };

  const hasAnyRole = (roles: string[]): boolean => {
    return roles.some(role => state.roles.includes(role));
  };

  const isAdmin = (): boolean => {
    return hasRole('admin');
  };

  const value: AuthContextValue = {
    ...state,
    refetch: fetchUser,
    hasRole,
    hasAnyRole,
    isAdmin,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

/**
 * Hook to access the auth context
 *
 * Usage:
 * ```tsx
 * function MyComponent() {
 *   const { user, roles, loading } = useAuth();
 *
 *   if (loading) return <div>Loading...</div>;
 *   if (!user) return <div>Not authenticated</div>;
 *
 *   return <div>Hello, {user.displayName}!</div>;
 * }
 * ```
 */
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);

  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return context;
}

/**
 * Hook to require authentication
 * Throws an error if not authenticated, useful for protected components
 *
 * Usage:
 * ```tsx
 * function ProtectedComponent() {
 *   const { user, roles } = useRequireAuth();
 *
 *   // user is guaranteed to be non-null here
 *   return <div>Hello, {user.displayName}!</div>;
 * }
 * ```
 */
export function useRequireAuth(): AuthContextValue {
  const context = useAuth();

  if (!context.user && !context.loading) {
    throw new Error('User must be authenticated to access this component');
  }

  return context;
}
