// Authentication API client

export interface User {
  id: string;
  email: string;
  displayName: string;
  provider?: 'google' | 'github';
  householdId?: string;
  preferences?: Record<string, string>;
  createdAt: string;
  updatedAt: string;
}

export interface MeResponse {
  user: User;
  roles: string[];
}

// JSON:API response structure
interface JsonApiResponse {
  data: {
    type: string;
    id: string;
    attributes: {
      email: string;
      displayName: string;
      provider: string;
      householdId?: string;
      roles: string[];
      preferences?: Record<string, string>;
      createdAt: string;
      updatedAt: string;
    };
  };
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
        'Content-Type': 'application/vnd.api+json',
        'Accept': 'application/vnd.api+json',
      },
    });

    if (response.status === 401) {
      // Not authenticated
      return null;
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.statusText}`);
    }

    const jsonApiData: JsonApiResponse = await response.json();

    // Transform JSON:API format to our internal format
    const user: User = {
      id: jsonApiData.data.id,
      email: jsonApiData.data.attributes.email,
      displayName: jsonApiData.data.attributes.displayName,
      provider: jsonApiData.data.attributes.provider as 'google' | 'github',
      householdId: jsonApiData.data.attributes.householdId,
      preferences: jsonApiData.data.attributes.preferences,
      createdAt: jsonApiData.data.attributes.createdAt,
      updatedAt: jsonApiData.data.attributes.updatedAt,
    };

    return {
      user,
      roles: jsonApiData.data.attributes.roles,
    };
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
  // This will clear the session cookie and redirect back to root landing page
  window.location.href = `/oauth2/${provider}/sign_out?rd=/`;
}
