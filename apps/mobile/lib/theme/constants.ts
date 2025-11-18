// Theme constants and types

export type ThemeMode = 'system' | 'light' | 'dark';

// Preference key for theme in the backend
export const THEME_PREFERENCE_KEY = 'theme';

// Local storage key for theme
export const THEME_STORAGE_KEY = 'home-hub-theme';

// Default theme mode
export const DEFAULT_THEME_MODE: ThemeMode = 'system';

/**
 * Gets the current system theme preference
 */
export function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') {
    return 'light';
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

/**
 * Resolves a theme mode to an actual theme (light or dark)
 */
export function resolveTheme(mode: ThemeMode): 'light' | 'dark' {
  if (mode === 'system') {
    return getSystemTheme();
  }
  return mode;
}

/**
 * Applies a theme to the document
 */
export function applyTheme(theme: 'light' | 'dark'): void {
  if (typeof document === 'undefined') {
    return;
  }

  const root = document.documentElement;

  if (theme === 'dark') {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }
}
