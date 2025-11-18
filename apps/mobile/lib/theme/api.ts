// Theme preference API client

import { ThemeMode } from './constants';

/**
 * Fetches the theme preference from the backend
 * Returns 'system' if no preference is set or on error
 */
export async function getThemePreference(): Promise<ThemeMode> {
  try {
    const response = await fetch('/api/users/me/preferences/theme', {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/vnd.api+json',
        'Accept': 'application/vnd.api+json',
      },
    });

    if (response.status === 404) {
      // Preference not set, return default
      return 'system';
    }

    if (!response.ok) {
      console.warn('Failed to fetch theme preference:', response.statusText);
      return 'system';
    }

    const data = await response.json();
    const theme = data.data?.attributes?.value as ThemeMode;

    // Validate the theme value
    if (theme === 'system' || theme === 'light' || theme === 'dark') {
      return theme;
    }

    console.warn('Invalid theme value from backend:', theme);
    return 'system';
  } catch (error) {
    console.error('Error fetching theme preference:', error);
    return 'system';
  }
}

/**
 * Sets the theme preference in the backend
 */
export async function setThemePreference(mode: ThemeMode): Promise<void> {
  try {
    const response = await fetch('/api/users/me/preferences/theme', {
      method: 'PATCH',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/vnd.api+json',
        'Accept': 'application/vnd.api+json',
      },
      body: JSON.stringify({
        data: {
          type: 'preferences',
          id: 'theme',
          attributes: {
            value: mode,
          },
        },
      }),
    });

    if (!response.ok) {
      throw new Error(`Failed to set theme preference: ${response.statusText}`);
    }
  } catch (error) {
    console.error('Error setting theme preference:', error);
    throw error;
  }
}
