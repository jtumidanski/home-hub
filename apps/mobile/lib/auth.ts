'use client';

import { useEffect, useState } from 'react';

export interface User {
  id: string;
  email: string;
  displayName: string;
  householdId?: string;
  roles?: string[];
}

export interface AuthError {
  message: string;
  code?: string;
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [roles, setRoles] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<AuthError | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const response = await fetch('/api/me', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/vnd.api+json',
            'Accept': 'application/vnd.api+json',
          },
        });

        if (response.status === 401) {
          // Not authenticated
          setUser(null);
          setRoles([]);
          setError(null);
          setLoading(false);
          return;
        }

        if (!response.ok) {
          throw new Error('Failed to fetch user');
        }

        const jsonApiData = await response.json();

        // Transform JSON:API format to our internal format
        const userData: User = {
          id: jsonApiData.data.id,
          email: jsonApiData.data.attributes.email,
          displayName: jsonApiData.data.attributes.displayName,
          householdId: jsonApiData.data.attributes.householdId,
          roles: jsonApiData.data.attributes.roles,
        };

        setUser(userData);
        setRoles(jsonApiData.data.attributes.roles || []);
        setError(null);
      } catch (err) {
        console.error('Auth error:', err);
        setError({
          message: err instanceof Error ? err.message : 'Authentication failed',
        });
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, []);

  return { user, roles, loading, error };
}
