// Authentication API client

export interface User {
  id: string;
  email: string;
  displayName: string;
  provider: 'google' | 'github';
  householdId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MeResponse {
  user: User;
  roles: string[];
}

/**
 * Fetches the current authenticated user's information
 * Returns null if not authenticated
 */
export async function fetchMe(): Promise<MeResponse | null> {
  try {
    const response = await fetch('/api/me', {
      method: 'GET',
      credentials: 'include', // Include cookies
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.status === 401) {
      // Not authenticated
      return null;
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.statusText}`);
    }

    const data: MeResponse = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching current user:', error);
    throw error;
  }
}

/**
 * Logs out the current user by redirecting to the OAuth provider's sign out endpoint
 */
export function logout(provider: 'google' | 'github' = 'google'): void {
  // Redirect to oauth2-proxy sign out endpoint
  // This will clear the session cookie and redirect back
  window.location.href = `/oauth2/${provider}/sign_out`;
}
